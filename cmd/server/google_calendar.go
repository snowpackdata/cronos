package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/snowpackdata/cronos"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

// CalendarEvent represents a simplified calendar event for API responses
type CalendarEvent struct {
	ID          string `json:"id"`
	Summary     string `json:"summary"`
	Description string `json:"description"`
	Start       string `json:"start"` // RFC3339 string to preserve timezone
	End         string `json:"end"`   // RFC3339 string to preserve timezone
}

// getGoogleOAuthConfig returns the OAuth2 configuration for Google Calendar
// If redirectURL is empty, it will use the environment variable
func getGoogleOAuthConfig(redirectURL string) *oauth2.Config {
	if redirectURL == "" {
		redirectURL = os.Getenv("GOOGLE_REDIRECT_URL")
	}
	return &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  redirectURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/calendar.readonly",
		},
		Endpoint: google.Endpoint,
	}
}

// getRedirectURLFromRequest constructs the redirect URL from the current request
// Strips subdomain to use the main domain for OAuth callbacks
func getRedirectURLFromRequest(r *http.Request) string {
	scheme := "https"
	if r.TLS == nil && r.Header.Get("X-Forwarded-Proto") != "https" {
		scheme = "http"
	}

	host := r.Host

	// Strip port if present to process domain
	hostWithoutPort := host
	port := ""
	if idx := strings.LastIndex(host, ":"); idx != -1 {
		hostWithoutPort = host[:idx]
		port = host[idx:]
	}

	// Strip subdomain if present (e.g., snowpack.localhost -> localhost)
	parts := strings.Split(hostWithoutPort, ".")
	if len(parts) > 1 {
		// If we have subdomain.domain or subdomain.domain.tld, strip the first part
		hostWithoutPort = strings.Join(parts[1:], ".")
	}

	host = hostWithoutPort + port

	return fmt.Sprintf("%s://%s/api/google/auth/callback", scheme, host)
}

// GoogleAuthURLHandler generates and returns the OAuth URL for user authorization
func (a *App) GoogleAuthURLHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Auto-detect redirect URL from request
	redirectURL := getRedirectURLFromRequest(r)
	config := getGoogleOAuthConfig(redirectURL)
	userID := r.Context().Value("user_id")

	// Generate a state token that includes the user ID for verification
	state := fmt.Sprintf("user_%v", userID)

	// Generate the authorization URL
	authURL := config.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"auth_url": authURL,
	})
}

// GoogleAuthCallbackHandler handles the OAuth callback from Google
func (a *App) GoogleAuthCallbackHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "No authorization code provided", http.StatusBadRequest)
		return
	}

	// Exchange the authorization code for tokens
	redirectURL := getRedirectURLFromRequest(r)
	config := getGoogleOAuthConfig(redirectURL)
	token, err := config.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, "Failed to exchange authorization code", http.StatusInternalServerError)
		return
	}

	// Extract user ID from state parameter
	state := r.URL.Query().Get("state")
	var userID uint
	_, err = fmt.Sscanf(state, "user_%d", &userID)
	if err != nil {
		http.Error(w, "Invalid state parameter", http.StatusBadRequest)
		return
	}

	// Store or update the tokens in the database
	var googleAuth cronos.GoogleAuth
	result := a.cronosApp.DB.Where("user_id = ?", userID).First(&googleAuth)

	if result.Error != nil {
		// Create new record
		googleAuth = cronos.GoogleAuth{
			UserID:       userID,
			AccessToken:  token.AccessToken,
			RefreshToken: token.RefreshToken,
			ExpiresAt:    token.Expiry,
		}
		a.cronosApp.DB.Create(&googleAuth)
	} else {
		// Update existing record
		googleAuth.AccessToken = token.AccessToken
		if token.RefreshToken != "" {
			googleAuth.RefreshToken = token.RefreshToken
		}
		googleAuth.ExpiresAt = token.Expiry
		a.cronosApp.DB.Save(&googleAuth)
	}

	// Redirect to a success page or close the popup
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `
		<html>
		<head><title>Authorization Successful</title></head>
		<body>
			<h1>Google Calendar Connected Successfully!</h1>
			<p>You can close this window and return to the application.</p>
			<script>
				// Notify parent window and close popup
				if (window.opener) {
					window.opener.postMessage({ type: 'google_auth_success' }, '*');
					window.close();
				}
			</script>
		</body>
		</html>
	`)
}

// GoogleAuthStatusHandler checks if the user has connected their Google Calendar
func (a *App) GoogleAuthStatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	userID := r.Context().Value("user_id")

	var googleAuth cronos.GoogleAuth
	result := a.cronosApp.DB.Where("user_id = ?", userID).First(&googleAuth)

	connected := result.Error == nil && googleAuth.ID != 0
	needsReauth := false

	if connected {
		// Check if token is expired and needs refresh
		needsReauth = time.Now().After(googleAuth.ExpiresAt)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{
		"connected":    connected,
		"needs_reauth": needsReauth,
	})
}

// GoogleAuthDisconnectHandler revokes Google Calendar access
func (a *App) GoogleAuthDisconnectHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	userID := r.Context().Value("user_id")

	// Delete the GoogleAuth record
	a.cronosApp.DB.Where("user_id = ?", userID).Delete(&cronos.GoogleAuth{})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "disconnected",
	})
}

// getCalendarService creates an authenticated Google Calendar service
func (a *App) getCalendarService(userID interface{}) (*calendar.Service, error) {
	// First, try to get tokens from User record (from Google Login)
	var user cronos.User
	if err := a.cronosApp.DB.First(&user, userID).Error; err == nil && user.GoogleAccessToken != "" {
		// Use tokens from User record
		config := getGoogleOAuthConfig("")
		token := &oauth2.Token{
			AccessToken:  user.GoogleAccessToken,
			RefreshToken: user.GoogleRefreshToken,
			Expiry:       *user.GoogleTokenExpiry,
		}

		// Create token source that handles automatic refresh
		tokenSource := config.TokenSource(context.Background(), token)

		// Get potentially refreshed token
		newToken, err := tokenSource.Token()
		if err != nil {
			return nil, fmt.Errorf("failed to refresh token: %v", err)
		}

		// Update token in User record if it was refreshed
		if newToken.AccessToken != token.AccessToken {
			user.GoogleAccessToken = newToken.AccessToken
			if newToken.RefreshToken != "" {
				user.GoogleRefreshToken = newToken.RefreshToken
			}
			user.GoogleTokenExpiry = &newToken.Expiry
			a.cronosApp.DB.Save(&user)
		}

		// Create calendar service
		ctx := context.Background()
		service, err := calendar.NewService(ctx, option.WithTokenSource(tokenSource))
		if err != nil {
			return nil, fmt.Errorf("failed to create calendar service: %v", err)
		}

		return service, nil
	}

	// Fall back to GoogleAuth table (legacy calendar-specific OAuth)
	var googleAuth cronos.GoogleAuth
	result := a.cronosApp.DB.Where("user_id = ?", userID).First(&googleAuth)
	if result.Error != nil {
		return nil, fmt.Errorf("user has not connected Google Calendar")
	}

	// Redirect URL doesn't matter for token refresh
	config := getGoogleOAuthConfig("")
	token := &oauth2.Token{
		AccessToken:  googleAuth.AccessToken,
		RefreshToken: googleAuth.RefreshToken,
		Expiry:       googleAuth.ExpiresAt,
	}

	// Create token source that handles automatic refresh
	tokenSource := config.TokenSource(context.Background(), token)

	// Get potentially refreshed token
	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %v", err)
	}

	// Update token in database if it was refreshed
	if newToken.AccessToken != token.AccessToken {
		googleAuth.AccessToken = newToken.AccessToken
		if newToken.RefreshToken != "" {
			googleAuth.RefreshToken = newToken.RefreshToken
		}
		googleAuth.ExpiresAt = newToken.Expiry
		a.cronosApp.DB.Save(&googleAuth)
	}

	// Create calendar service
	ctx := context.Background()
	service, err := calendar.NewService(ctx, option.WithTokenSource(tokenSource))
	if err != nil {
		return nil, fmt.Errorf("failed to create calendar service: %v", err)
	}

	return service, nil
}

// GoogleCalendarEventsHandler fetches calendar events for a date range
func (a *App) GoogleCalendarEventsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	userID := r.Context().Value("user_id")

	// Parse query parameters for date range
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")

	if startDateStr == "" || endDateStr == "" {
		http.Error(w, "start_date and end_date are required", http.StatusBadRequest)
		return
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		http.Error(w, "Invalid start_date format. Use YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		http.Error(w, "Invalid end_date format. Use YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	// Set time to start/end of day
	startDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, startDate.Location())
	endDate = time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 999999999, endDate.Location())

	// Get calendar service
	service, err := a.getCalendarService(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// Fetch events from primary calendar
	events, err := service.Events.List("primary").
		TimeMin(startDate.Format(time.RFC3339)).
		TimeMax(endDate.Format(time.RFC3339)).
		SingleEvents(true).
		OrderBy("startTime").
		Do()

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to fetch calendar events: %v", err), http.StatusInternalServerError)
		return
	}

	// Convert to our CalendarEvent format
	calendarEvents := make([]CalendarEvent, 0)
	for _, event := range events.Items {
		// Skip all-day events and events without start/end times
		if event.Start.DateTime == "" || event.End.DateTime == "" {
			continue
		}

		// Use the RFC3339 strings directly to preserve timezone information
		calendarEvents = append(calendarEvents, CalendarEvent{
			ID:          event.Id,
			Summary:     event.Summary,
			Description: event.Description,
			Start:       event.Start.DateTime,
			End:         event.End.DateTime,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(calendarEvents)
}
