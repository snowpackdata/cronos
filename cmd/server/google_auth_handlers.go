package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/snowpackdata/cronos"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// GoogleLoginHandler initiates the Google OAuth flow
func (a *App) GoogleLoginHandler(w http.ResponseWriter, r *http.Request) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		http.Error(w, "Failed to generate state", http.StatusInternalServerError)
		return
	}
	state := base64.URLEncoding.EncodeToString(b)

	// Check if this is for registration
	isRegistration := r.URL.Query().Get("registration") == "true"
	if isRegistration {
		state = "registration:" + state
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		MaxAge:   300,
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})

	config := getGoogleLoginOAuthConfig(r)
	url := config.AuthCodeURL(state, oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// GoogleLoginCallbackHandler handles OAuth callback for login
func (a *App) GoogleLoginCallbackHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("GoogleLoginCallback: Started - Host: %s, URL: %s", r.Host, r.URL.String())

	// Check if Google returned an error
	if errParam := r.URL.Query().Get("error"); errParam != "" {
		log.Printf("GoogleLoginCallback: Google OAuth error: %s", errParam)
		loginURL := fmt.Sprintf("http://%s/login?error=google_oauth_failed", r.Host)
		if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
			loginURL = fmt.Sprintf("https://%s/login?error=google_oauth_failed", r.Host)
		}
		http.Redirect(w, r, loginURL, http.StatusTemporaryRedirect)
		return
	}

	stateCookie, err := r.Cookie("oauth_state")
	if err != nil {
		log.Printf("GoogleLoginCallback: State cookie error: %v", err)
		loginURL := fmt.Sprintf("http://%s/login?error=invalid_state", r.Host)
		if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
			loginURL = fmt.Sprintf("https://%s/login?error=invalid_state", r.Host)
		}
		http.Redirect(w, r, loginURL, http.StatusTemporaryRedirect)
		return
	}

	if stateCookie.Value != r.URL.Query().Get("state") {
		log.Printf("GoogleLoginCallback: State mismatch - cookie: %s, query: %s",
			stateCookie.Value, r.URL.Query().Get("state"))
		loginURL := fmt.Sprintf("http://%s/login?error=invalid_state", r.Host)
		if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
			loginURL = fmt.Sprintf("https://%s/login?error=invalid_state", r.Host)
		}
		http.Redirect(w, r, loginURL, http.StatusTemporaryRedirect)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:   "oauth_state",
		Value:  "",
		MaxAge: -1,
		Path:   "/",
	})

	code := r.URL.Query().Get("code")
	config := getGoogleLoginOAuthConfig(r)
	token, err := config.Exchange(context.Background(), code)
	if err != nil {
		log.Printf("OAuth token exchange failed: %v", err)
		http.Error(w, "Failed to exchange token", http.StatusInternalServerError)
		return
	}

	client := config.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read user info", http.StatusInternalServerError)
		return
	}

	var userInfo struct {
		Email         string `json:"email"`
		VerifiedEmail bool   `json:"verified_email"`
		GivenName     string `json:"given_name"`
		FamilyName    string `json:"family_name"`
		Name          string `json:"name"`
	}
	if err := json.Unmarshal(body, &userInfo); err != nil {
		http.Error(w, "Failed to parse user info", http.StatusInternalServerError)
		return
	}

	if !userInfo.VerifiedEmail {
		http.Error(w, "Email not verified", http.StatusBadRequest)
		return
	}

	// Check if this is a registration flow
	isRegistration := strings.HasPrefix(stateCookie.Value, "registration:")
	if isRegistration {
		// Redirect to registration page with pre-filled data
		parts := strings.Split(userInfo.Email, "@")
		domain := ""
		if len(parts) == 2 {
			domain = parts[1]
		}

		regURL := fmt.Sprintf("http://%s/new-organization?email=%s&first_name=%s&last_name=%s&domain=%s",
			r.Host, userInfo.Email, userInfo.GivenName, userInfo.FamilyName, domain)
		if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
			regURL = fmt.Sprintf("https://%s/new-organization?email=%s&first_name=%s&last_name=%s&domain=%s",
				r.Host, userInfo.Email, userInfo.GivenName, userInfo.FamilyName, domain)
		}
		http.Redirect(w, r, regURL, http.StatusTemporaryRedirect)
		return
	}

	// Extract domain from email
	parts := strings.Split(userInfo.Email, "@")
	if len(parts) != 2 {
		http.Error(w, "Invalid email", http.StatusBadRequest)
		return
	}
	domain := parts[1]

	// Find tenant by domain
	var tenant cronos.Tenant
	if err := a.cronosApp.DB.Where("domain = ? AND status = ?", domain, "active").First(&tenant).Error; err != nil {
		log.Printf("GoogleLoginCallback: No tenant found for domain %s: %v", domain, err)

		// Redirect to login page with error message
		loginURL := fmt.Sprintf("http://%s/login?error=no_organization", r.Host)
		if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
			loginURL = fmt.Sprintf("https://%s/login?error=no_organization", r.Host)
		}
		http.Redirect(w, r, loginURL, http.StatusTemporaryRedirect)
		return
	}

	log.Printf("GoogleLoginCallback: Tenant found - ID: %d, Slug: %s, Domain: %s", tenant.ID, tenant.Slug, tenant.Domain)

	// Find user in tenant
	var user cronos.User
	if err := a.cronosApp.DB.Scopes(cronos.TenantScope(tenant.ID)).Where("email = ?", userInfo.Email).First(&user).Error; err != nil {
		log.Printf("GoogleLoginCallback: User not found for %s in tenant %s: %v", userInfo.Email, tenant.Slug, err)

		// Redirect to login page with error message
		loginURL := fmt.Sprintf("http://%s/login?error=user_not_found", r.Host)
		if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
			loginURL = fmt.Sprintf("https://%s/login?error=user_not_found", r.Host)
		}
		http.Redirect(w, r, loginURL, http.StatusTemporaryRedirect)
		return
	}

	log.Printf("GoogleLoginCallback: User found - ID: %d, Email: %s, Role: %s", user.ID, user.Email, user.Role)

	// Store OAuth tokens in User record for calendar API access
	user.GoogleAccessToken = token.AccessToken
	if token.RefreshToken != "" {
		user.GoogleRefreshToken = token.RefreshToken
	}
	user.GoogleTokenExpiry = &token.Expiry
	if err := a.cronosApp.DB.Save(&user).Error; err != nil {
		log.Printf("Failed to save Google tokens to user: %v", err)
	}

	// Update or create GoogleAuth record
	var googleAuth cronos.GoogleAuth
	result := a.cronosApp.DB.Scopes(cronos.TenantScope(tenant.ID)).Where("user_id = ?", user.ID).First(&googleAuth)
	if result.Error != nil {
		// Create new record
		googleAuth.TenantID = tenant.ID
		googleAuth.UserID = user.ID
		googleAuth.GoogleEmail = userInfo.Email
		googleAuth.AccessToken = token.AccessToken
		googleAuth.RefreshToken = token.RefreshToken
		googleAuth.ExpiresAt = token.Expiry
		a.cronosApp.DB.Create(&googleAuth)
	} else {
		// Update existing
		googleAuth.AccessToken = token.AccessToken
		if token.RefreshToken != "" {
			googleAuth.RefreshToken = token.RefreshToken
		}
		googleAuth.ExpiresAt = token.Expiry
		a.cronosApp.DB.Save(&googleAuth)
	}

	// Generate JWT using same function as regular login
	isStaff := user.Role == cronos.UserRoleStaff.String() || user.Role == cronos.UserRoleAdmin.String()
	issuer := "snowpackdata.com"
	if strings.Contains(r.Host, "localhost") || strings.Contains(r.Host, "127.0.0.1") {
		issuer = "localhost"
	}

	tokenString, err := generateTokenString(user, isStaff, user.AccountID, issuer, tenant.ID)
	if err != nil {
		log.Printf("GoogleLoginCallbackHandler: Token generation failed: %v", err)
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// Determine redirect path based on role
	redirectPath := "/portal/dashboard"
	if isStaff {
		redirectPath = "/admin/timesheet"
	}

	// Build tenant subdomain URL
	scheme := "http"
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}

	host := r.Host
	var targetHost string
	if strings.Contains(host, "localhost") || strings.Contains(host, "127.0.0.1") {
		// Local dev - use tenant.localhost
		portIdx := strings.Index(host, ":")
		port := ""
		if portIdx != -1 {
			port = host[portIdx:]
		}
		targetHost = fmt.Sprintf("%s.localhost%s", tenant.Slug, port)
	} else {
		// Production - use tenant.domain.com
		// Remove any existing subdomain and add tenant slug
		parts := strings.Split(host, ".")
		if len(parts) > 2 {
			// Has subdomain, replace it
			parts[0] = tenant.Slug
		} else {
			// No subdomain, prepend tenant slug
			parts = append([]string{tenant.Slug}, parts...)
		}
		targetHost = strings.Join(parts, ".")
	}

	redirectURL := fmt.Sprintf("%s://%s%s?token=%s", scheme, targetHost, redirectPath, tokenString)

	log.Printf("GoogleLoginCallback: Redirecting to: %s", redirectURL)
	log.Printf("GoogleLoginCallback: User: %s, Tenant: %s, IsStaff: %v", user.Email, tenant.Slug, isStaff)

	html := fmt.Sprintf(`<!DOCTYPE html>
<html><head><title>Redirecting...</title></head><body>
<script>
window.location.href = '%s';
</script>
<p>Redirecting...</p>
</body></html>`, redirectURL)

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func getGoogleLoginOAuthConfig(r *http.Request) *oauth2.Config {
	// Use fixed redirect URL from env var (same approach as calendar OAuth)
	// Google doesn't support wildcard redirect URIs, so we use a single fixed URL
	// The tenant is determined by email domain, not subdomain
	redirectURL := os.Getenv("GOOGLE_LOGIN_REDIRECT_URL")
	if redirectURL == "" {
		// Fallback for local dev
		redirectURL = "http://localhost:8080/auth/google/login/callback"
	}

	return &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),     // Same as calendar OAuth
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"), // Same as calendar OAuth
		RedirectURL:  redirectURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
			"https://www.googleapis.com/auth/calendar.readonly",
		},
		Endpoint: google.Endpoint,
	}
}
