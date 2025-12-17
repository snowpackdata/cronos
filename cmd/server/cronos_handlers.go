package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/snowpackdata/cronos"
	"golang.org/x/crypto/bcrypt"
)

var JWTSecret = func() string {
	// Read from environment variable with a fallback to a default
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		// Fallback to a default secret for development
		jwtSecret = "default_development_jwt_secret_for_snowpack_data"
	}

	return jwtSecret
}()

// generateTokenString creates a new JWT for a given user.
func generateTokenString(user cronos.User, isStaff bool, accountID uint, issuer string) (string, error) {
	log.Printf("Generating token. UserID: %d, AccountID: %d, Email: %s, IsStaff: %v, Role: %s, Issuer: %s",
		user.ID, accountID, user.Email, isStaff, user.Role, issuer)

	claims := Claims{ // This refers to main.Claims from middleware.go
		UserID:    user.ID,
		AccountID: accountID,
		Email:     user.Email,
		IsStaff:   isStaff,
		Role:      user.Role,
		RegisteredClaims: &jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 720)), // 30 days
			Issuer:    issuer,
		},
	}
	token := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), claims)

	secretPrefix := JWTSecret
	if len(JWTSecret) > 5 {
		secretPrefix = JWTSecret[:5]
	}
	log.Printf("generateTokenString - Using JWT secret with prefix: %s...", secretPrefix)

	return token.SignedString([]byte(JWTSecret))
}

// RegistrationLandingHandler serves the registration page when accessed via GET request
func (a *App) RegistrationLandingHandler(w http.ResponseWriter, req *http.Request) {
	data, err := templates.ReadFile("templates/registration.html")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(data)
}

// LoginLandingHandler serves the registration page when accessed via GET request
func (a *App) LoginLandingHandler(w http.ResponseWriter, req *http.Request) {
	tmpl, err := template.ParseFS(templates, "templates/login.html")
	if err != nil {
		log.Printf("Error parsing login template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	err = tmpl.Execute(w, a.GitHash)
	if err != nil {
		log.Printf("Error executing login template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// RegisterUser creates a new user in the database when accessed via POST request
func (a *App) RegisterUser(w http.ResponseWriter, req *http.Request) {
	// Read email and password from the post request
	formRole := req.FormValue("role")
	formUserID, err := strconv.ParseUint(req.FormValue("user_id"), 10, 32)
	if err != nil {
		log.Println("RegisterUser Error: Invalid user_id format", err)
		http.Error(w, "Invalid user_id format", http.StatusBadRequest)
		return
	}
	formFirstName := req.FormValue("first_name")
	formLastName := req.FormValue("last_name")
	formPassword := req.FormValue("password")

	// Create a new client object and fill in the fields
	isStaff := false
	switch formRole {
	case cronos.UserRoleClient.String():
		client := cronos.Client{UserID: uint(formUserID)}
		if a.cronosApp.DB.Where("user_id = ?", formUserID).First(&client).RowsAffected == 0 {
			a.cronosApp.DB.Create(&client)
		}
		client.FirstName = formFirstName
		client.LastName = formLastName
		a.cronosApp.DB.Save(&client)
	case cronos.UserRoleStaff.String(), cronos.UserRoleAdmin.String():
		employee := cronos.Employee{UserID: uint(formUserID)}
		if a.cronosApp.DB.Where("user_id = ?", formUserID).First(&employee).RowsAffected == 0 {
			a.cronosApp.DB.Create(&employee)
		}
		employee.FirstName = formFirstName
		employee.LastName = formLastName
		employee.StartDate = time.Now()
		isStaff = true
		a.cronosApp.DB.Save(&employee)
	default:
		log.Println("RegisterUser Error: Invalid role specified", formRole)
		http.Error(w, "Invalid role specified", http.StatusBadRequest)
		return
	}
	var user cronos.User
	if err := a.cronosApp.DB.Where("id = ?", formUserID).First(&user).Error; err != nil {
		log.Println("RegisterUser Error: User not found after profile creation", err)
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(formPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Println("RegisterUser Error: Password hashing failed", err)
		http.Error(w, "Error processing request", http.StatusInternalServerError)
		return
	}

	user.Password = string(hashedPassword)
	if err := a.cronosApp.DB.Save(&user).Error; err != nil {
		log.Println("RegisterUser Error: Failed to save user with new password", err)
		http.Error(w, "Error saving user data", http.StatusInternalServerError)
		return
	}

	// Determine issuer based on the host
	issuer := "snowpackdata.com"
	if host := req.Host; strings.Contains(host, "localhost") || strings.Contains(host, "127.0.0.1") {
		issuer = "localhost"
	}

	// Log the issuer for debugging
	log.Printf("Token issuer for registration: %s", issuer)

	// Ensure AccountID is correctly obtained. user.AccountID might be a pointer or direct value.
	var accountID uint
	if user.AccountID != 0 { // Assuming user.AccountID is uint and 0 is its zero/unassigned value
		accountID = user.AccountID
	} else {
		log.Println("RegisterUser Warning: User does not have an AccountID associated (AccountID is 0).")
	}

	tokenString, err := generateTokenString(user, isStaff, accountID, issuer)
	if err != nil {
		log.Println("RegisterUser Error: Token generation failed", err)
		http.Error(w, "Failed to generate authentication token", http.StatusInternalServerError)
		return
	}

	resp := map[string]interface{}{"token": tokenString}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) // Typically 201 for successful registration
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Println("RegisterUser Error: Failed to encode response", err)
	}
}

func (a *App) VerifyEmail(w http.ResponseWriter, req *http.Request) {
	// Read email from the post request and check if the email exists as an account in
	// our database. If so send a 200
	// if not send a 300
	formEmail := req.FormValue("email")
	var user cronos.User
	if a.cronosApp.DB.Where("email = ?", formEmail).First(&user).RowsAffected != 0 {
		response := map[string]interface{}{
			"user_id": user.ID,
			"role":    user.Role,
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Println(err)
		}
		return
	} else {
		http.Error(w, "Email not found", http.StatusNotFound)
	}
}

func (a *App) VerifyLogin(w http.ResponseWriter, req *http.Request) {
	// Verify login checks a customers hashed password against the database to determine if
	// they are verified. If they are, it generates a new JWT token and returns it to the
	// customer.

	formEmail := req.FormValue("email")
	formPassword := req.FormValue("password")

	var user cronos.User

	if a.cronosApp.DB.Where("email = ?", formEmail).First(&user).RowsAffected == 0 {
		var resp = map[string]interface{}{"status": 403, "message": "Invalid login credentials. Please try again"}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(resp)
		return
	}

	// validate password
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(formPassword))
	if err != nil {
		var resp = map[string]interface{}{"status": 403, "message": "Invalid login credentials. Please try again"}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(resp)
		return
	}

	isStaff := false
	if user.Role == cronos.UserRoleStaff.String() || user.Role == cronos.UserRoleAdmin.String() {
		isStaff = true
	}

	// Determine issuer based on the host
	issuer := "snowpackdata.com"
	if host := req.Host; strings.Contains(host, "localhost") || strings.Contains(host, "127.0.0.1") {
		issuer = "localhost"
	}

	// Log the issuer for debugging
	log.Printf("Token issuer for login: %s", issuer)

	// Ensure AccountID is correctly obtained. user.AccountID might be a pointer or direct value.
	var accountID uint
	if user.AccountID != 0 { // Assuming user.AccountID is uint and 0 is its zero/unassigned value
		accountID = user.AccountID
	} else {
		log.Println("VerifyLogin Warning: User does not have an AccountID associated (AccountID is 0).")
	}

	tokenString, err := generateTokenString(user, isStaff, accountID, issuer)
	if err != nil {
		log.Println("VerifyLogin Error: Token generation failed", err)
		http.Error(w, "Failed to generate authentication token", http.StatusInternalServerError)
		return
	}
	var resp = map[string]interface{}{"status": 200, "message": "logged in"}
	resp["token"] = tokenString //Store the token in the response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
	return
}

// AdminLandingHandler serves the admin page when accessed via GET request
func (a *App) AdminLandingHandler(w http.ResponseWriter, req *http.Request) {
	log.Printf("Serving admin landing page to %s", req.RemoteAddr)

	// Check if the file exists first
	filePath := "./static/admin/index.html"
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Printf("Warning: Admin index file not found at %s", filePath)
		// Try fallback locations
		alternativePaths := []string{
			"./website/static/admin/index.html",
			"./admin/dist/index.html",
			"./website/admin/dist/index.html",
		}

		for _, altPath := range alternativePaths {
			if _, err := os.Stat(altPath); err == nil {
				log.Printf("Found admin index at alternative path: %s", altPath)
				filePath = altPath
				break
			}
		}
	}

	// If file still not found, return 404
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Printf("Error: Admin index.html still not found after checking alternatives.")
		http.Error(w, "Admin interface not found.", http.StatusNotFound)
		return
	}

	http.ServeFile(w, req, filePath)
}

// PortalLandingHandler serves the portal page when accessed via GET request
func (a *App) PortalLandingHandler(w http.ResponseWriter, req *http.Request) {
	log.Printf("Serving portal landing page to %s", req.RemoteAddr)

	// TODO: Consider serving from portalAssets embed.FS like other assets for consistency
	filePath := "./static/portal/index.html"
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Printf("Warning: Portal index file not found at %s", filePath)
		// Add fallbacks if necessary, e.g., for development builds
		alternativePaths := []string{
			"./portal/dist/index.html",
			"./website/portal/dist/index.html",
		}

		for _, altPath := range alternativePaths {
			if _, err := os.Stat(altPath); err == nil {
				log.Printf("Found portal index at alternative path: %s", altPath)
				filePath = altPath
				break
			}
		}
	}

	// If file still not found, return 404
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Printf("Error: Portal index.html still not found after checking alternatives.")
		http.Error(w, "Portal interface not found.", http.StatusNotFound)
		return
	}

	http.ServeFile(w, req, filePath)
}

// LegacyAdminLandingHandler serves the legacy admin page when accessed via GET request
func (a *App) LegacyAdminLandingHandler(w http.ResponseWriter, req *http.Request) {
	data, err := templates.ReadFile("templates/admin.html")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(data)
}

// RequestPasswordReset handles the password reset request and sends reset instructions
func (a *App) RequestPasswordReset(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	email := req.FormValue("email")
	if email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}

	// Check if user exists
	var user cronos.User
	if a.cronosApp.DB.Where("email = ?", email).First(&user).RowsAffected == 0 {
		// For security, don't reveal if the email exists or not
		// Return success message regardless
		log.Printf("Password reset requested for non-existent email: %s", email)
	} else {
		log.Printf("Password reset requested for: %s", email)
		// TODO: Generate reset token and send email
		// For now, just log that a reset was requested
	}

	// Always return success to prevent email enumeration
	response := map[string]interface{}{
		"status":  200,
		"message": "If an account exists with this email, password reset instructions have been sent.",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// ResetPassword resets a user's password to a desired value
// This endpoint is meant for emergency local use only and requires a secret passphrase
func (a *App) ResetPassword(w http.ResponseWriter, req *http.Request) {
	// Check if the request is using the proper method
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract email, secretPhrase, and newPassword from request
	email := req.FormValue("email")
	secretPhrase := req.FormValue("secret")
	newPassword := req.FormValue("password")

	// Validate the secret passphrase
	if secretPhrase != "Snowpack1!" {
		log.Printf("Password reset attempted with invalid secret for email: %s", email)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check if the email was provided
	if email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}

	// Check if the new password was provided
	if newPassword == "" {
		http.Error(w, "New password is required", http.StatusBadRequest)
		return
	}

	// Find the user in the database
	var user cronos.User
	if a.cronosApp.DB.Where("email = ?", email).First(&user).RowsAffected == 0 {
		log.Printf("Password reset attempted for non-existent user: %s", email)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Hash the new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error hashing password for %s: %v", email, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Update the user's password in the database
	user.Password = string(hashedPassword)
	if err := a.cronosApp.DB.Save(&user).Error; err != nil {
		log.Printf("Error saving updated password for %s: %v", email, err)
		http.Error(w, "Failed to update password", http.StatusInternalServerError)
		return
	}

	// Log the successful password reset
	log.Printf("Password successfully reset for user: %s", email)

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  200,
		"message": "Password has been successfully reset",
		"email":   email,
	})
}

// RefreshTokenHandler allows users to get a new token if their current token is still valid
func (a *App) RefreshTokenHandler(w http.ResponseWriter, req *http.Request) {
	// Get the current token from the request header
	tokenString := req.Header.Get("x-access-token")
	tokenString = strings.TrimSpace(tokenString)

	if tokenString == "" {
		respondWithError(w, http.StatusBadRequest, "No token provided")
		return
	}

	// Parse and validate the current token
	tclaims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, tclaims, func(token *jwt.Token) (interface{}, error) {
		return []byte(JWTSecret), nil
	})

	if err != nil || !token.Valid || !IsValidIssuer(tclaims.RegisteredClaims.Issuer) {
		log.Printf("RefreshTokenHandler: Invalid token - err: %v, valid: %v, issuer: %s", err, token.Valid, tclaims.RegisteredClaims.Issuer)
		respondWithError(w, http.StatusUnauthorized, "Invalid or expired token")
		return
	}

	// Check if token is close to expiring (within 7 days)
	expirationTime := tclaims.RegisteredClaims.ExpiresAt
	if expirationTime == nil {
		respondWithError(w, http.StatusBadRequest, "Token has no expiration time")
		return
	}

	// If token expires in more than 7 days, no need to refresh
	sevenDaysFromNow := time.Now().Add(7 * 24 * time.Hour)
	if expirationTime.Time.After(sevenDaysFromNow) {
		respondWithError(w, http.StatusBadRequest, "Token is not close to expiring")
		return
	}

	// Get the user from the database to ensure they still exist and are active
	var user cronos.User
	if err := a.cronosApp.DB.First(&user, tclaims.UserID).Error; err != nil {
		respondWithError(w, http.StatusUnauthorized, "User not found")
		return
	}

	// Determine if user is staff
	isStaff := false
	if user.Role == cronos.UserRoleStaff.String() || user.Role == cronos.UserRoleAdmin.String() {
		isStaff = true
	}

	// Determine issuer based on the host
	issuer := "snowpackdata.com"
	if host := req.Host; strings.Contains(host, "localhost") || strings.Contains(host, "127.0.0.1") {
		issuer = "localhost"
	}

	// Generate a new token
	newTokenString, err := generateTokenString(user, isStaff, tclaims.AccountID, issuer)
	if err != nil {
		log.Printf("RefreshTokenHandler Error: Failed to generate new token for user %d: %v", user.ID, err)
		respondWithError(w, http.StatusInternalServerError, "Failed to generate new token")
		return
	}

	// Return the new token
	response := map[string]interface{}{
		"status":  200,
		"message": "Token refreshed successfully",
		"token":   newTokenString,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("RefreshTokenHandler Error: Failed to encode response: %v", err)
	}
}
