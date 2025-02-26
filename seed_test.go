package cronos

import (
	"testing"
)

// TestSeedDatabase verifies that the SeedDatabase function creates
// the initial users, employees, and account records correctly
func TestSeedDatabase(t *testing.T) {
	// Setup test DB
	db := setupTestDB(t)
	app := &App{
		DB: db,
	}

	// Call the function being tested
	app.SeedDatabase()

	// Verify admin user was created
	var adminUser User
	if err := db.Where("email = ?", "nate@snowpack-data.com").First(&adminUser).Error; err != nil {
		t.Errorf("Admin user not found: %v", err)
	}

	if adminUser.Role != UserRoleAdmin.String() {
		t.Errorf("Expected admin role %s, got %s", UserRoleAdmin.String(), adminUser.Role)
	}

	// Verify staff users were created
	var kevinUser User
	if err := db.Where("email = ?", "kevin@snowpack-data.com").First(&kevinUser).Error; err != nil {
		t.Errorf("Kevin user not found: %v", err)
	}

	if kevinUser.Role != UserRoleStaff.String() {
		t.Errorf("Expected staff role %s, got %s", UserRoleStaff.String(), kevinUser.Role)
	}

	var davidUser User
	if err := db.Where("email = ?", "david@snowpack-data.com").First(&davidUser).Error; err != nil {
		t.Errorf("David user not found: %v", err)
	}

	if davidUser.Role != UserRoleStaff.String() {
		t.Errorf("Expected staff role %s, got %s", UserRoleStaff.String(), davidUser.Role)
	}

	// Verify employees were created
	var nateEmployee Employee
	if err := db.Where("user_id = ?", adminUser.ID).First(&nateEmployee).Error; err != nil {
		t.Errorf("Nate employee not found: %v", err)
	}

	var kevinEmployee Employee
	if err := db.Where("user_id = ?", kevinUser.ID).First(&kevinEmployee).Error; err != nil {
		t.Errorf("Kevin employee not found: %v", err)
	}

	var davidEmployee Employee
	if err := db.Where("user_id = ?", davidUser.ID).First(&davidEmployee).Error; err != nil {
		t.Errorf("David employee not found: %v", err)
	}

	// Verify account was created
	var account Account
	if err := db.Where("name = ?", "Snowpack Data").First(&account).Error; err != nil {
		t.Errorf("Snowpack Data account not found: %v", err)
	}

	if account.Type != AccountTypeInternal.String() {
		t.Errorf("Expected account type %s, got %s", AccountTypeInternal.String(), account.Type)
	}

	if account.LegalName != "Snowpack Data, LLC" {
		t.Errorf("Expected legal name 'Snowpack Data, LLC', got '%s'", account.LegalName)
	}

	if account.Email != "billing@snowpack-data.com" {
		t.Errorf("Expected email 'billing@snowpack-data.com', got '%s'", account.Email)
	}

	if account.Website != "https://snowpack-data.com" {
		t.Errorf("Expected website 'https://snowpack-data.com', got '%s'", account.Website)
	}
}
