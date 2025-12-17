package main

import (
	"errors" // For gorm.ErrRecordNotFound comparison
	"log"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/snowpackdata/cronos"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// registerDevUser creates a test admin user and a test client user.
// It is intended for local development and testing.
// Assumes:
// - JWTSecret is a global variable (e.g., in cronos_handlers.go)
// - App struct is defined (e.g., in main.go)
// - Claims struct is defined (e.g., in middleware.go)
// - min helper function is defined (e.g., in cronos_handlers.go)
func (a *App) registerDevUser() string {
	// --- Create/Update Admin Dev User (ID 1) ---
	var adminUser cronos.User
	var adminErr error

	adminErr = a.cronosApp.DB.Where("id = ?", 1).First(&adminUser).Error

	if adminErr != nil && !errors.Is(adminErr, gorm.ErrRecordNotFound) {
		log.Printf("Error fetching admin user ID 1: %v", adminErr)
		return ""
	}

	if errors.Is(adminErr, gorm.ErrRecordNotFound) {
		log.Println("Creating development admin user with ID 1")
		adminUser = cronos.User{
			Email: "dev@example.com",
			Role:  cronos.UserRoleAdmin.String(),
		}
		// Try to create, then assign ID 1 if possible.
		if createErr := a.cronosApp.DB.Create(&adminUser).Error; createErr != nil {
			log.Printf("Error creating admin user dev@example.com: %v. Attempting to fetch by email.", createErr)
			if fetchErr := a.cronosApp.DB.Where("email = ?", "dev@example.com").First(&adminUser).Error; fetchErr != nil {
				log.Printf("Failed to fetch admin user dev@example.com by email: %v", fetchErr)
				return ""
			}
		}
		// Ensure ID is 1 if created or fetched and ID is different
		if adminUser.ID != 1 {
			originalID := adminUser.ID
			var tempUser cronos.User
			// Check if ID 1 is taken by another user
			if err := a.cronosApp.DB.Where("id = ?", 1).First(&tempUser).Error; err == nil && tempUser.Email != adminUser.Email {
				log.Printf("Admin user ID 1 is already taken by %s. Using current ID %d for dev@example.com.", tempUser.Email, originalID)
				// adminUser.ID remains originalID
			} else {
				// ID 1 is free or belongs to this user, or another error occurred (which implies ID 1 might be assignable)
				if updateErr := a.cronosApp.DB.Model(&adminUser).Update("id", 1).Error; updateErr != nil {
					log.Printf("Failed to update admin user ID from %d to 1: %v. Using ID %d.", originalID, updateErr, originalID)
				} else {
					adminUser.ID = 1
					log.Printf("Admin user dev@example.com set to ID 1 (was %d).", originalID)
				}
			}
		}
	} else {
		log.Println("Development admin user with ID 1 already exists.")
		adminUser.ID = 1
		if adminUser.Role != cronos.UserRoleAdmin.String() {
			adminUser.Role = cronos.UserRoleAdmin.String()
			// Password will be reset next
		}
	}

	adminPassword := "devpassword"
	hashedAdminPassword, errHashAdmin := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
	if errHashAdmin != nil {
		log.Println("Error generating password hash for admin user:", errHashAdmin)
	} else if adminUser.ID != 0 { // Only save if adminUser has an ID
		adminUser.Password = string(hashedAdminPassword)
		if saveErr := a.cronosApp.DB.Save(&adminUser).Error; saveErr != nil {
			log.Printf("Error saving admin user (ID %d) password: %v", adminUser.ID, saveErr)
		}
	}

	var employee cronos.Employee
	if adminUser.ID != 0 { // Proceed only if adminUser has a valid ID
		errEmployee := a.cronosApp.DB.Where("user_id = ?", adminUser.ID).First(&employee).Error
		if errEmployee != nil && !errors.Is(errEmployee, gorm.ErrRecordNotFound) {
			log.Printf("Error fetching employee record for admin user ID %d: %v", adminUser.ID, errEmployee)
		} else if errors.Is(errEmployee, gorm.ErrRecordNotFound) {
			log.Println("Creating employee record for development admin user ID:", adminUser.ID)
			employee = cronos.Employee{UserID: adminUser.ID, FirstName: "Dev", LastName: "Admin", StartDate: time.Now()}
			a.cronosApp.DB.Create(&employee)
		} else { // Record found
			log.Println("Updating existing employee record for development admin user ID:", adminUser.ID)
			employee.FirstName = "Dev"
			employee.LastName = "Admin"
			a.cronosApp.DB.Save(&employee)
		}
	}

	adminTokenString := ""
	if adminUser.ID != 0 {
		adminClaims := Claims{
			UserID: adminUser.ID, Email: adminUser.Email, IsStaff: true, Role: adminUser.Role,
			RegisteredClaims: &jwt.RegisteredClaims{
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 720)),
				Issuer:    "localhost",
			},
		}
		adminToken := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), adminClaims)
		var errSignAdmin error
		adminTokenString, errSignAdmin = adminToken.SignedString([]byte(JWTSecret))
		if errSignAdmin != nil {
			log.Println("Error signing JWT token for admin:", errSignAdmin)
			adminTokenString = ""
		} else {
			// Ensure adminTokenString is not shorter than 10 before slicing
			logTokenChars := 10
			if len(adminTokenString) < logTokenChars {
				logTokenChars = len(adminTokenString)
			}
			log.Printf("Created development JWT token for admin user_id=%d (first %d chars: %s...)", adminUser.ID, logTokenChars, adminTokenString[:logTokenChars])
		}
	}

	// --- Create/Update Test Client User ---
	var finalTestAccountID uint
	desiredTestAccountID := uint(1)
	testAccountName := "Test Client Account"
	var testAccount cronos.Account

	if err := a.cronosApp.DB.Where("id = ?", desiredTestAccountID).First(&testAccount).Error; err == nil {
		finalTestAccountID = testAccount.ID
		log.Printf("Test account '%s' (ID %d) already exists.", testAccount.Name, finalTestAccountID)
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		var existingByName cronos.Account
		if errFindByName := a.cronosApp.DB.Where("name = ?", testAccountName).First(&existingByName).Error; errFindByName == nil {
			testAccount = existingByName
			finalTestAccountID = existingByName.ID
			log.Printf("Found existing test account by name: '%s' (ID %d). Using this account.", testAccount.Name, finalTestAccountID)
		} else {
			log.Printf("Creating new test account: '%s'", testAccountName)
			// Ensure LegalName is unique if there's a constraint
			testAccount = cronos.Account{Name: testAccountName, Type: cronos.AccountTypeClient.String(), LegalName: testAccountName + "-" + time.Now().Format("20060102150405")}
			if createErr := a.cronosApp.DB.Create(&testAccount).Error; createErr != nil {
				log.Printf("Error creating test account '%s': %v", testAccountName, createErr)
				return adminTokenString
			}
			finalTestAccountID = testAccount.ID
			log.Printf("New test account '%s' created with ID %d", testAccount.Name, finalTestAccountID)
		}
	} else {
		log.Printf("Error fetching account with ID %d: %v", desiredTestAccountID, err)
		return adminTokenString
	}
	if finalTestAccountID == 0 {
		log.Printf("Error: Could not obtain valid ID for test account.")
		return adminTokenString
	}

	var clientUser cronos.User
	desiredClientUserID := uint(2)
	clientUserEmail := "client@example.com"
	clientUserPassword := "clientpassword"

	if err := a.cronosApp.DB.Where("id = ?", desiredClientUserID).First(&clientUser).Error; err == nil {
		log.Printf("Test client user ID %d (%s) already exists.", clientUser.ID, clientUser.Email)
		if clientUser.Role != cronos.UserRoleClient.String() || clientUser.AccountID != finalTestAccountID {
			clientUser.Role = cronos.UserRoleClient.String()
			clientUser.AccountID = finalTestAccountID
			a.cronosApp.DB.Save(&clientUser)
			log.Printf("Updated existing client user ID %d to Role: Client, AccountID: %d", clientUser.ID, finalTestAccountID)
		}
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		// Try to create with preferred ID first
		clientUserAttempt := cronos.User{Model: gorm.Model{ID: desiredClientUserID}, Email: clientUserEmail, Role: cronos.UserRoleClient.String(), AccountID: finalTestAccountID}
		if createErr := a.cronosApp.DB.Create(&clientUserAttempt).Error; createErr == nil {
			clientUser = clientUserAttempt // Use the successfully created user
			log.Printf("Test client user %s created with preferred ID %d", clientUser.Email, clientUser.ID)
		} else {
			log.Printf("Failed to create client user %s with preferred ID %d (error: %v). Fetching by email or creating with GORM-assigned ID.", clientUserEmail, desiredClientUserID, createErr)
			// Check if it exists by email (could be created by another process or unique constraint on email)
			if fetchByEmailErr := a.cronosApp.DB.Where("email = ?", clientUserEmail).First(&clientUser).Error; fetchByEmailErr == nil {
				log.Printf("Found client user by email: %s (ID %d). Ensuring AccountID is correct.", clientUser.Email, clientUser.ID)
				if clientUser.AccountID != finalTestAccountID {
					clientUser.AccountID = finalTestAccountID
					a.cronosApp.DB.Save(&clientUser)
				}
			} else { // Truly does not exist by email, create with GORM-assigned ID
				clientUser = cronos.User{Email: clientUserEmail, Role: cronos.UserRoleClient.String(), AccountID: finalTestAccountID}
				if createErr2 := a.cronosApp.DB.Create(&clientUser).Error; createErr2 != nil {
					log.Printf("Error creating test client user %s (even with GORM-assigned ID): %v", clientUserEmail, createErr2)
					return adminTokenString
				}
				log.Printf("Test client user %s created with GORM-assigned ID %d", clientUser.Email, clientUser.ID)
			}
		}
	} else {
		log.Printf("Error fetching client user ID %d: %v", desiredClientUserID, err)
		return adminTokenString
	}
	if clientUser.ID == 0 {
		log.Printf("Error: clientUser ID is 0 for %s", clientUserEmail)
		return adminTokenString
	}

	clientHashedPassword, errHashClient := bcrypt.GenerateFromPassword([]byte(clientUserPassword), bcrypt.DefaultCost)
	if errHashClient != nil {
		log.Println("Error generating password hash for client user:", errHashClient)
	} else {
		clientUser.Password = string(clientHashedPassword)
		if err := a.cronosApp.DB.Save(&clientUser).Error; err != nil {
			log.Printf("Error saving password for client user ID %d: %v", clientUser.ID, err)
		} else {
			log.Printf("Set/Reset password for test client user ID %d (%s)", clientUser.ID, clientUser.Email)
		}
	}

	var testClientProfile cronos.Client
	if err := a.cronosApp.DB.Where("user_id = ?", clientUser.ID).First(&testClientProfile).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf("Error fetching client profile for UserID %d: %v", clientUser.ID, err)
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		testClientProfile = cronos.Client{UserID: clientUser.ID, FirstName: "Test", LastName: "ClientUser", Title: "Client"}
		if err := a.cronosApp.DB.Create(&testClientProfile).Error; err != nil {
			log.Printf("Error creating client profile for UserID %d: %v", clientUser.ID, err)
		}
		log.Printf("Created client profile for test client user ID %d", clientUser.ID)
	} else { // Profile found, update it
		testClientProfile.FirstName = "Test"
		testClientProfile.LastName = "ClientUser"
		testClientProfile.Title = "Client"
		if err := a.cronosApp.DB.Save(&testClientProfile).Error; err != nil {
			log.Printf("Error updating client profile for UserID %d: %v", clientUser.ID, err)
		}
		log.Printf("Updated client profile for test client user ID %d", clientUser.ID)
	}
	log.Printf("Test Client User Setup: Email=%s, Password=%s, UserID=%d, AccountID=%d, ClientProfileID=%d", clientUserEmail, clientUserPassword, clientUser.ID, clientUser.AccountID, testClientProfile.ID)

	return adminTokenString
}
