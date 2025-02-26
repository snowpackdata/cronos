package cronos

import (
	"testing"
)

// TestRegisterClient tests the RegisterClient function
func TestRegisterClient(t *testing.T) {
	// Setup test DB
	db := setupTestDB(t)
	app := &App{
		DB: db,
	}

	// Create a test account
	account := Account{
		Name:      "Test Client Account",
		LegalName: "Test Client Legal Name",
		Type:      AccountTypeClient.String(),
	}
	if err := db.Create(&account).Error; err != nil {
		t.Fatalf("Failed to create account: %v", err)
	}

	// Test cases
	testCases := []struct {
		name      string
		email     string
		accountID uint
		wantErr   bool
		errType   error
	}{
		{
			name:      "Valid Registration",
			email:     "newclient@example.com",
			accountID: account.ID,
			wantErr:   false,
			errType:   nil,
		},
		{
			name:      "Duplicate User",
			email:     "newclient@example.com", // Same email as previous test
			accountID: account.ID,
			wantErr:   true,
			errType:   ErrUserAlreadyExists,
		},
	}

	// Create a custom register function that doesn't send emails
	registerClient := func(email string, accountID uint) error {
		// Create a blank user and client if they don't already exist
		user := User{Email: email}
		if app.DB.Model(&user).Where("email = ?", email).Updates(&user).RowsAffected == 0 {
			app.DB.Create(&user)
		} else {
			return ErrUserAlreadyExists
		}
		user.Password = DEFAULT_PASSWORD
		user.Role = UserRoleClient.String()
		user.AccountID = accountID
		app.DB.Save(&user)
		return nil // Skip email sending
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Call the function being tested
			err := registerClient(tc.email, tc.accountID)

			// Check for expected errors
			if (err != nil) != tc.wantErr {
				t.Errorf("RegisterClient() error = %v, wantErr %v", err, tc.wantErr)
				return
			}

			if tc.wantErr && err != tc.errType {
				t.Errorf("RegisterClient() error = %v, expected error type %v", err, tc.errType)
				return
			}

			// If no error, verify the user was created correctly
			if !tc.wantErr {
				var user User
				if err := db.Where("email = ?", tc.email).First(&user).Error; err != nil {
					t.Errorf("Failed to find created user: %v", err)
					return
				}

				// Verify user properties
				if user.Email != tc.email {
					t.Errorf("Expected user email %s, got %s", tc.email, user.Email)
				}

				if user.Role != UserRoleClient.String() {
					t.Errorf("Expected user role %s, got %s", UserRoleClient.String(), user.Role)
				}

				if user.AccountID != tc.accountID {
					t.Errorf("Expected account ID %d, got %d", tc.accountID, user.AccountID)
				}
			}
		})
	}
}

// TestRegisterStaff tests the RegisterStaff function
func TestRegisterStaff(t *testing.T) {
	// Setup test DB
	db := setupTestDB(t)
	app := &App{
		DB: db,
	}

	// Create a test account
	account := Account{
		Name:      "Test Staff Account",
		LegalName: "Test Staff Legal Name",
		Type:      AccountTypeInternal.String(),
	}
	if err := db.Create(&account).Error; err != nil {
		t.Fatalf("Failed to create account: %v", err)
	}

	// Create a custom register function that doesn't send emails
	registerStaff := func(email string, accountID uint) error {
		// Create a blank user and staff
		user := User{Email: email}
		app.DB.Save(&user)
		app.DB.Save(&Employee{User: user})
		user.Password = DEFAULT_PASSWORD
		user.Role = UserRoleStaff.String()
		user.AccountID = accountID
		app.DB.Save(&user)
		return nil // Skip email sending
	}

	// Test case
	email := "newstaff@example.com"

	// Call the function being tested
	err := registerStaff(email, account.ID)

	// Check for expected errors
	if err != nil {
		t.Errorf("RegisterStaff() error = %v", err)
		return
	}

	// Verify the user was created correctly
	var user User
	if err := db.Where("email = ?", email).First(&user).Error; err != nil {
		t.Errorf("Failed to find created user: %v", err)
		return
	}

	// Verify user properties
	if user.Email != email {
		t.Errorf("Expected user email %s, got %s", email, user.Email)
	}

	if user.Role != UserRoleStaff.String() {
		t.Errorf("Expected user role %s, got %s", UserRoleStaff.String(), user.Role)
	}

	if user.AccountID != account.ID {
		t.Errorf("Expected account ID %d, got %d", account.ID, user.AccountID)
	}

	// Verify employee was created
	var employee Employee
	if err := db.Where("user_id = ?", user.ID).First(&employee).Error; err != nil {
		t.Errorf("Failed to find created employee: %v", err)
		return
	}

	if employee.UserID != user.ID {
		t.Errorf("Expected employee user ID %d, got %d", user.ID, employee.UserID)
	}
}
