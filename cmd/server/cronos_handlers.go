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
func generateTokenString(user cronos.User, isStaff bool, accountID uint, issuer string, tenantID uint) (string, error) {
	log.Printf("Generating token. UserID: %d, TenantID: %d, AccountID: %d, Email: %s, IsStaff: %v, Role: %s, Issuer: %s",
		user.ID, tenantID, accountID, user.Email, isStaff, user.Role, issuer)

	claims := Claims{ // This refers to main.Claims from middleware.go
		UserID:    user.ID,
		TenantID:  tenantID,
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

// TenantRegistrationLandingHandler serves the tenant registration page (hidden, non-public link)
func (a *App) TenantRegistrationLandingHandler(w http.ResponseWriter, req *http.Request) {
	tmpl, err := template.ParseFS(templates, "templates/tenant_registration.html")
	if err != nil {
		log.Printf("Error parsing tenant registration template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	err = tmpl.Execute(w, nil)
	if err != nil {
		log.Printf("Error executing tenant registration template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// RegisterTenant creates a new tenant organization and its first admin user
func (a *App) RegisterTenant(w http.ResponseWriter, req *http.Request) {
	// Parse form data
	tenantName := req.FormValue("tenant_name")
	tenantSlug := req.FormValue("tenant_slug")
	tenantDomain := req.FormValue("tenant_domain")
	adminEmail := req.FormValue("admin_email")
	adminPassword := req.FormValue("admin_password")
	adminFirstName := req.FormValue("admin_first_name")
	adminLastName := req.FormValue("admin_last_name")

	// Validate required fields (password is optional for Google OAuth users)
	if tenantName == "" || tenantSlug == "" || adminEmail == "" {
		log.Printf("RegisterTenant Error: Missing required fields (tenant_name=%s, tenant_slug=%s, admin_email=%s)", tenantName, tenantSlug, adminEmail)
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// If password is provided, validate minimum length
	if adminPassword != "" && len(adminPassword) < 8 {
		log.Printf("RegisterTenant Error: Password too short")
		http.Error(w, "Password must be at least 8 characters", http.StatusBadRequest)
		return
	}

	// Normalize slug to lowercase
	tenantSlug = strings.ToLower(strings.TrimSpace(tenantSlug))

	// Check if tenant slug already exists
	var existingTenant cronos.Tenant
	if err := a.cronosApp.DB.Where("slug = ?", tenantSlug).First(&existingTenant).Error; err == nil {
		log.Printf("RegisterTenant Error: Tenant slug already exists: %s", tenantSlug)
		http.Error(w, "Organization subdomain already exists", http.StatusConflict)
		return
	}

	// Check if domain already exists (if provided)
	if tenantDomain != "" {
		if err := a.cronosApp.DB.Where("domain = ?", tenantDomain).First(&existingTenant).Error; err == nil {
			log.Printf("RegisterTenant Error: Tenant domain already exists: %s", tenantDomain)
			http.Error(w, "Organization domain already exists", http.StatusConflict)
			return
		}
	}

	// Create GCS bucket for this tenant
	bucketName, err := a.cronosApp.CreateTenantBucket(tenantSlug)
	if err != nil {
		log.Printf("RegisterTenant Error: Failed to create bucket: %v", err)
		http.Error(w, "Failed to create storage bucket", http.StatusInternalServerError)
		return
	}

	// Create the tenant
	tenant := cronos.Tenant{
		Name:       tenantName,
		Slug:       tenantSlug,
		Domain:     tenantDomain,
		BucketName: bucketName,
		Plan:       "trial",
		Status:     "active",
	}

	if err := a.cronosApp.DB.Create(&tenant).Error; err != nil {
		log.Printf("RegisterTenant Error: Failed to create tenant: %v", err)
		http.Error(w, "Failed to create organization", http.StatusInternalServerError)
		return
	}

	log.Printf("RegisterTenant: Created tenant %s (ID: %d) with bucket %s", tenant.Name, tenant.ID, bucketName)

	// Hash the admin password (leave empty for Google OAuth users)
	var passwordHash string
	if adminPassword != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("RegisterTenant Error: Password hashing failed: %v", err)
			http.Error(w, "Error processing request", http.StatusInternalServerError)
			return
		}
		passwordHash = string(hashedPassword)
	}
	// If adminPassword is empty, passwordHash remains empty string
	// This indicates a Google-only user

	// Create owner account representing the tenant's company
	// This is the account that holds company details (legal name, address, etc.)
	ownerAccount := cronos.Account{
		Name:      tenantName,
		LegalName: tenantName, // Can be updated later with LLC, Inc., etc.
		Type:      cronos.AccountTypeInternal.String(),
		Email:     adminEmail, // Use admin email as initial contact
		TenantID:  tenant.ID,
	}
	if err := a.cronosApp.DB.Create(&ownerAccount).Error; err != nil {
		log.Printf("RegisterTenant Error: Failed to create owner account: %v", err)
		http.Error(w, "Failed to create owner account", http.StatusInternalServerError)
		return
	}

	log.Printf("RegisterTenant: Created owner account %s (ID: %d)", ownerAccount.Name, ownerAccount.ID)

	// Create the first admin user
	adminUser := cronos.User{
		Email:     adminEmail,
		Password:  passwordHash, // Empty string if Google-only user
		Role:      cronos.UserRoleAdmin.String(),
		TenantID:  tenant.ID,
		AccountID: ownerAccount.ID,
	}

	if err := a.cronosApp.DB.Create(&adminUser).Error; err != nil {
		log.Printf("RegisterTenant Error: Failed to create admin user: %v", err)
		http.Error(w, "Failed to create admin user", http.StatusInternalServerError)
		return
	}

	log.Printf("RegisterTenant: Created admin user %s (ID: %d)", adminUser.Email, adminUser.ID)

	// Create employee record for the admin
	employee := cronos.Employee{
		UserID:    adminUser.ID,
		TenantID:  tenant.ID,
		FirstName: adminFirstName,
		LastName:  adminLastName,
		StartDate: time.Now(),
	}

	if err := a.cronosApp.DB.Create(&employee).Error; err != nil {
		log.Printf("RegisterTenant Error: Failed to create employee record: %v", err)
		// Non-fatal, continue
	}

	// Generate JWT token for the new admin
	issuer := "snowpackdata.com"
	if host := req.Host; strings.Contains(host, "localhost") || strings.Contains(host, "127.0.0.1") {
		issuer = "localhost"
	}

	tokenString, err := generateTokenString(adminUser, true, ownerAccount.ID, issuer, tenant.ID)
	if err != nil {
		log.Printf("RegisterTenant Error: Failed to generate token: %v", err)
		http.Error(w, "Error generating authentication token", http.StatusInternalServerError)
		return
	}

	// Return success with tenant info and redirect URL
	response := map[string]interface{}{
		"success":     true,
		"tenant_slug": tenant.Slug,
		"token":       tokenString,
		"redirect":    "https://" + tenant.Slug + "." + strings.Replace(req.Host, "www.", "", 1) + "/admin/?token=" + tokenString,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// RegisterUser creates a new user in the database when accessed via POST request
func (a *App) RegisterUser(w http.ResponseWriter, req *http.Request) {
	// Extract tenant from subdomain
	slug := extractSubdomain(req.Host)
	var tenant cronos.Tenant
	if err := a.cronosApp.DB.Where("slug = ? AND status = ?", slug, "active").First(&tenant).Error; err != nil {
		log.Printf("RegisterUser Error: Tenant not found for slug '%s': %v", slug, err)
		http.Error(w, "Invalid tenant", http.StatusBadRequest)
		return
	}

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
		client := cronos.Client{UserID: uint(formUserID), TenantID: tenant.ID}
		if a.cronosApp.DB.Scopes(cronos.TenantScope(tenant.ID)).Where("user_id = ?", formUserID).First(&client).RowsAffected == 0 {
			a.cronosApp.DB.Create(&client)
		}
		client.FirstName = formFirstName
		client.LastName = formLastName
		a.cronosApp.DB.Save(&client)
	case cronos.UserRoleStaff.String(), cronos.UserRoleAdmin.String():
		employee := cronos.Employee{UserID: uint(formUserID), TenantID: tenant.ID}
		if a.cronosApp.DB.Scopes(cronos.TenantScope(tenant.ID)).Where("user_id = ?", formUserID).First(&employee).RowsAffected == 0 {
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
	if err := a.cronosApp.DB.Scopes(cronos.TenantScope(tenant.ID)).Where("id = ?", formUserID).First(&user).Error; err != nil {
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

	tokenString, err := generateTokenString(user, isStaff, accountID, issuer, tenant.ID)
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
	// Extract tenant from subdomain
	slug := extractSubdomain(req.Host)
	var tenant cronos.Tenant
	if err := a.cronosApp.DB.Where("slug = ? AND status = ?", slug, "active").First(&tenant).Error; err != nil {
		log.Printf("VerifyEmail Error: Tenant not found for slug '%s': %v", slug, err)
		http.Error(w, "Invalid tenant", http.StatusBadRequest)
		return
	}

	// Read email from the post request and check if the email exists as an account in
	// our database. If so send a 200
	// if not send a 300
	formEmail := req.FormValue("email")
	var user cronos.User
	if a.cronosApp.DB.Scopes(cronos.TenantScope(tenant.ID)).Where("email = ?", formEmail).First(&user).RowsAffected != 0 {
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
	formEmail := req.FormValue("email")
	formPassword := req.FormValue("password")

	// Extract domain from email to find tenant
	parts := strings.Split(formEmail, "@")
	if len(parts) != 2 {
		var resp = map[string]interface{}{"status": 400, "message": "Invalid email format"}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(resp)
		return
	}
	domain := parts[1]

	// Find tenant by domain
	var tenant cronos.Tenant
	if err := a.cronosApp.DB.Where("domain = ? AND status = ?", domain, "active").First(&tenant).Error; err != nil {
		log.Printf("VerifyLogin Error: No tenant found for domain '%s': %v", domain, err)
		var resp = map[string]interface{}{"status": 403, "message": "No organization found for your email domain. Please contact support."}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(resp)
		return
	}

	var user cronos.User

	if a.cronosApp.DB.Scopes(cronos.TenantScope(tenant.ID)).Where("email = ?", formEmail).First(&user).RowsAffected == 0 {
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

	tokenString, err := generateTokenString(user, isStaff, accountID, issuer, tenant.ID)
	if err != nil {
		log.Println("VerifyLogin Error: Token generation failed", err)
		http.Error(w, "Failed to generate authentication token", http.StatusInternalServerError)
		return
	}
	var resp = map[string]interface{}{
		"status":      200,
		"message":     "logged in",
		"token":       tokenString,
		"tenant_slug": tenant.Slug,
		"is_staff":    isStaff,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
	return
}

// AdminLandingHandler serves the admin page when accessed via GET request
func (a *App) AdminLandingHandler(w http.ResponseWriter, req *http.Request) {
	log.Printf("Serving admin landing page to %s", req.RemoteAddr)

	content, err := adminAssets.ReadFile("static/admin/index.html")
	if err != nil {
		log.Printf("Error reading admin index.html from embedded assets: %v", err)
		http.Error(w, "Admin interface not found.", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(content)
}

// PortalLandingHandler serves the portal page when accessed via GET request
func (a *App) PortalLandingHandler(w http.ResponseWriter, req *http.Request) {
	log.Printf("Serving portal landing page to %s", req.RemoteAddr)

	content, err := portalAssets.ReadFile("static/portal/index.html")
	if err != nil {
		log.Printf("Error reading portal index.html from embedded assets: %v", err)
		http.Error(w, "Portal interface not found.", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(content)
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

	// Extract tenant from subdomain
	slug := extractSubdomain(req.Host)
	var tenant cronos.Tenant
	if err := a.cronosApp.DB.Where("slug = ? AND status = ?", slug, "active").First(&tenant).Error; err != nil {
		log.Printf("RequestPasswordReset Error: Tenant not found for slug '%s': %v", slug, err)
		// Still return success to prevent tenant enumeration
	}

	// Check if user exists (within tenant)
	var user cronos.User
	if tenant.ID != 0 && a.cronosApp.DB.Scopes(cronos.TenantScope(tenant.ID)).Where("email = ?", email).First(&user).RowsAffected == 0 {
		// For security, don't reveal if the email exists or not
		// Return success message regardless
		log.Printf("Password reset requested for non-existent email: %s", email)
	} else if tenant.ID != 0 {
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

	// Extract tenant from subdomain
	slug := extractSubdomain(req.Host)
	var tenant cronos.Tenant
	if err := a.cronosApp.DB.Where("slug = ? AND status = ?", slug, "active").First(&tenant).Error; err != nil {
		log.Printf("ResetPassword Error: Tenant not found for slug '%s': %v", slug, err)
		http.Error(w, "Invalid tenant", http.StatusBadRequest)
		return
	}

	// Find the user in the database (within tenant)
	var user cronos.User
	if a.cronosApp.DB.Scopes(cronos.TenantScope(tenant.ID)).Where("email = ?", email).First(&user).RowsAffected == 0 {
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

	// Extract tenant from subdomain
	slug := extractSubdomain(req.Host)
	var tenant cronos.Tenant
	if err := a.cronosApp.DB.Where("slug = ? AND status = ?", slug, "active").First(&tenant).Error; err != nil {
		log.Printf("RefreshTokenHandler Error: Tenant not found for slug '%s': %v", slug, err)
		respondWithError(w, http.StatusUnauthorized, "Invalid tenant")
		return
	}

	// Get the user from the database to ensure they still exist and are active (within tenant)
	var user cronos.User
	if err := a.cronosApp.DB.Scopes(cronos.TenantScope(tenant.ID)).First(&user, tclaims.UserID).Error; err != nil {
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
	newTokenString, err := generateTokenString(user, isStaff, tclaims.AccountID, issuer, tclaims.TenantID)
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

// GetTenantHandler returns current tenant information for frontend
func (a *App) GetTenantHandler(w http.ResponseWriter, r *http.Request) {
	tenant := MustGetTenant(r.Context())

	response := map[string]interface{}{
		"id":       tenant.ID,
		"slug":     tenant.Slug,
		"name":     tenant.Name,
		"domain":   tenant.Domain,
		"plan":     tenant.Plan,
		"branding": tenant.Branding,
		"settings": tenant.Settings,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("GetTenantHandler Error: Failed to encode response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// UpdateTenantHandler updates tenant settings
func (a *App) UpdateTenantHandler(w http.ResponseWriter, r *http.Request) {
	tenant := MustGetTenant(r.Context())

	var updates struct {
		Name     *string `json:"name"`
		Slug     *string `json:"slug"`
		Domain   *string `json:"domain"`
		Settings *string `json:"settings"` // JSON string
		Branding *string `json:"branding"` // JSON string
	}

	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Update fields if provided
	if updates.Name != nil {
		tenant.Name = *updates.Name
	}
	if updates.Slug != nil {
		tenant.Slug = *updates.Slug
	}
	if updates.Domain != nil {
		tenant.Domain = *updates.Domain
	}
	if updates.Settings != nil {
		tenant.Settings = []byte(*updates.Settings)
	}
	if updates.Branding != nil {
		tenant.Branding = []byte(*updates.Branding)
	}

	if err := a.cronosApp.DB.Save(tenant).Error; err != nil {
		log.Printf("Error updating tenant: %v", err)
		http.Error(w, "Failed to update tenant", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"id":       tenant.ID,
		"slug":     tenant.Slug,
		"name":     tenant.Name,
		"domain":   tenant.Domain,
		"plan":     tenant.Plan,
		"branding": tenant.Branding,
		"settings": tenant.Settings,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
