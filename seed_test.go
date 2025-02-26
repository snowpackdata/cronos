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
	if err := db.Where("email = ?", "nate@snowpack-data.io").First(&adminUser).Error; err != nil {
		t.Errorf("Admin user not found: %v", err)
	}

	if adminUser.Role != UserRoleAdmin.String() {
		t.Errorf("Expected admin role %s, got %s", UserRoleAdmin.String(), adminUser.Role)
	}

	// Verify staff user was created
	var staffUser User
	if err := db.Where("email = ?", "staff@snowpack-data.io").First(&staffUser).Error; err != nil {
		t.Errorf("Staff user not found: %v", err)
	}

	if staffUser.Role != UserRoleStaff.String() {
		t.Errorf("Expected staff role %s, got %s", UserRoleStaff.String(), staffUser.Role)
	}

	// Verify client user was created
	var clientUser User
	if err := db.Where("email = ?", "client@snowpack-data.io").First(&clientUser).Error; err != nil {
		t.Errorf("Client user not found: %v", err)
	}

	if clientUser.Role != UserRoleClient.String() {
		t.Errorf("Expected client role %s, got %s", UserRoleClient.String(), clientUser.Role)
	}

	// Verify employees were created
	var adminEmployee Employee
	if err := db.Where("user_id = ?", adminUser.ID).First(&adminEmployee).Error; err != nil {
		t.Errorf("Admin employee not found: %v", err)
	}

	var staffEmployee Employee
	if err := db.Where("user_id = ?", staffUser.ID).First(&staffEmployee).Error; err != nil {
		t.Errorf("Staff employee not found: %v", err)
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

	if account.Email != "info@snowpack-data.io" {
		t.Errorf("Expected email 'info@snowpack-data.io', got '%s'", account.Email)
	}

	if account.Website != "https://snowpack-data.io" {
		t.Errorf("Expected website 'https://snowpack-data.io', got '%s'", account.Website)
	}

	// Verify user account relationships
	if adminUser.AccountID != account.ID {
		t.Errorf("Admin user account ID %d doesn't match expected account ID %d", adminUser.AccountID, account.ID)
	}

	if staffUser.AccountID != account.ID {
		t.Errorf("Staff user account ID %d doesn't match expected account ID %d", staffUser.AccountID, account.ID)
	}
}
