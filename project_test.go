package cronos

import (
	"testing"
	"time"
)

// TestCreateProject tests the creation of a new project
func TestCreateProject(t *testing.T) {
	// Setup test DB
	db := setupTestDB(t)

	// Create a test account
	account := Account{
		Name:      "Test Account",
		LegalName: "Test Legal Name",
		Type:      AccountTypeClient.String(),
		Email:     "test@example.com",
		Website:   "https://example.com",
	}
	if err := db.Create(&account).Error; err != nil {
		t.Fatalf("Failed to create account: %v", err)
	}

	// Create an account manager
	user := User{
		Email:    "manager@example.com",
		Password: DEFAULT_PASSWORD,
		Role:     UserRoleStaff.String(),
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	employee := Employee{
		UserID:    user.ID,
		FirstName: "Test",
		LastName:  "Manager",
		Title:     "Account Manager",
		IsActive:  true,
		StartDate: time.Now(),
	}
	if err := db.Create(&employee).Error; err != nil {
		t.Fatalf("Failed to create employee: %v", err)
	}

	// Test project creation
	now := time.Now()
	oneYearFromNow := now.AddDate(1, 0, 0)

	project := Project{
		Name:             "Test Project",
		AccountID:        account.ID,
		ActiveStart:      now,
		ActiveEnd:        oneYearFromNow,
		BudgetHours:      100,
		BudgetDollars:    10000,
		Internal:         false,
		BillingFrequency: "monthly",
		ProjectType:      ProjectTypeNew.String(),
		AEID:             &employee.ID,
	}

	// Save the project
	if err := db.Create(&project).Error; err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Retrieve the project
	var savedProject Project
	if err := db.Preload("Account").Preload("AE").First(&savedProject, project.ID).Error; err != nil {
		t.Fatalf("Failed to retrieve created project: %v", err)
	}

	// Verify project properties
	if savedProject.Name != project.Name {
		t.Errorf("Expected project name %s, got %s", project.Name, savedProject.Name)
	}

	if savedProject.AccountID != account.ID {
		t.Errorf("Expected account ID %d, got %d", account.ID, savedProject.AccountID)
	}

	if savedProject.BudgetHours != project.BudgetHours {
		t.Errorf("Expected budget hours %d, got %d", project.BudgetHours, savedProject.BudgetHours)
	}

	if savedProject.BudgetDollars != project.BudgetDollars {
		t.Errorf("Expected budget dollars %d, got %d", project.BudgetDollars, savedProject.BudgetDollars)
	}

	if savedProject.ProjectType != ProjectTypeNew.String() {
		t.Errorf("Expected project type %s, got %s", ProjectTypeNew.String(), savedProject.ProjectType)
	}

	if savedProject.AEID == nil || *savedProject.AEID != employee.ID {
		t.Errorf("Expected AE ID %d, got %v", employee.ID, savedProject.AEID)
	}

	// Verify the account relationship
	if savedProject.Account.ID != account.ID {
		t.Errorf("Expected project to be associated with account ID %d, got %d", account.ID, savedProject.Account.ID)
	}

	// Verify the AE relationship
	if savedProject.AE == nil || savedProject.AE.ID != employee.ID {
		t.Errorf("Expected project to be associated with AE ID %d, got %v", employee.ID, savedProject.AE)
	}
}

// TestUpdateProject tests updating an existing project
func TestUpdateProject(t *testing.T) {
	// Setup test DB
	db := setupTestDB(t)

	// Create a test account
	account := Account{
		Name:      "Test Account",
		LegalName: "Test Legal Name",
		Type:      AccountTypeClient.String(),
		Email:     "test@example.com",
		Website:   "https://example.com",
	}
	if err := db.Create(&account).Error; err != nil {
		t.Fatalf("Failed to create account: %v", err)
	}

	// Create an initial project
	now := time.Now()
	oneYearFromNow := now.AddDate(1, 0, 0)

	project := Project{
		Name:             "Initial Project Name",
		AccountID:        account.ID,
		ActiveStart:      now,
		ActiveEnd:        oneYearFromNow,
		BudgetHours:      100,
		BudgetDollars:    10000,
		Internal:         false,
		BillingFrequency: "monthly",
		ProjectType:      ProjectTypeNew.String(),
	}

	// Save the project
	if err := db.Create(&project).Error; err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Update the project
	project.Name = "Updated Project Name"
	project.BudgetHours = 200
	project.BudgetDollars = 20000
	project.BillingFrequency = "weekly"

	if err := db.Save(&project).Error; err != nil {
		t.Fatalf("Failed to update project: %v", err)
	}

	// Retrieve the updated project
	var updatedProject Project
	if err := db.First(&updatedProject, project.ID).Error; err != nil {
		t.Fatalf("Failed to retrieve updated project: %v", err)
	}

	// Verify updated properties
	if updatedProject.Name != "Updated Project Name" {
		t.Errorf("Expected updated name 'Updated Project Name', got '%s'", updatedProject.Name)
	}

	if updatedProject.BudgetHours != 200 {
		t.Errorf("Expected updated budget hours 200, got %d", updatedProject.BudgetHours)
	}

	if updatedProject.BudgetDollars != 20000 {
		t.Errorf("Expected updated budget dollars 20000, got %d", updatedProject.BudgetDollars)
	}

	if updatedProject.BillingFrequency != "weekly" {
		t.Errorf("Expected updated billing frequency 'weekly', got '%s'", updatedProject.BillingFrequency)
	}
}

// TestAddBillingCodeToProject tests adding a billing code to a project
func TestAddBillingCodeToProject(t *testing.T) {
	// Setup test DB
	db := setupTestDB(t)

	// Create a test account
	account := Account{
		Name:      "Test Account",
		LegalName: "Test Legal Name",
		Type:      AccountTypeClient.String(),
	}
	if err := db.Create(&account).Error; err != nil {
		t.Fatalf("Failed to create account: %v", err)
	}

	// Create a project
	now := time.Now()
	oneYearFromNow := now.AddDate(1, 0, 0)

	project := Project{
		Name:             "Test Project",
		AccountID:        account.ID,
		ActiveStart:      now,
		ActiveEnd:        oneYearFromNow,
		BudgetHours:      100,
		BudgetDollars:    10000,
		Internal:         false,
		BillingFrequency: "monthly",
		ProjectType:      ProjectTypeNew.String(),
	}

	// Save the project
	if err := db.Create(&project).Error; err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Create a billing code for the project
	billingCode := BillingCode{
		Name:        "Consulting",
		RateType:    "hourly",
		Category:    "consulting",
		Code:        "CONS-001",
		RoundedTo:   15,
		ProjectID:   project.ID,
		ActiveStart: now,
		ActiveEnd:   oneYearFromNow,
	}

	if err := db.Create(&billingCode).Error; err != nil {
		t.Fatalf("Failed to create billing code: %v", err)
	}

	// Retrieve the project with billing codes
	var projectWithBillingCodes Project
	if err := db.Preload("BillingCodes").First(&projectWithBillingCodes, project.ID).Error; err != nil {
		t.Fatalf("Failed to retrieve project with billing codes: %v", err)
	}

	// Verify the billing code was associated with the project
	if len(projectWithBillingCodes.BillingCodes) != 1 {
		t.Errorf("Expected 1 billing code, got %d", len(projectWithBillingCodes.BillingCodes))
	}

	if projectWithBillingCodes.BillingCodes[0].ID != billingCode.ID {
		t.Errorf("Expected billing code ID %d, got %d", billingCode.ID, projectWithBillingCodes.BillingCodes[0].ID)
	}

	if projectWithBillingCodes.BillingCodes[0].Name != "Consulting" {
		t.Errorf("Expected billing code name 'Consulting', got '%s'", projectWithBillingCodes.BillingCodes[0].Name)
	}

	if projectWithBillingCodes.BillingCodes[0].Code != "CONS-001" {
		t.Errorf("Expected billing code 'CONS-001', got '%s'", projectWithBillingCodes.BillingCodes[0].Code)
	}
}

// TestProjectWithHubspotDealID tests creating and updating a project with a HubSpot Deal ID
func TestProjectWithHubspotDealID(t *testing.T) {
	// Setup test DB
	db := setupTestDB(t)

	// Create a test account
	account := Account{
		Name:      "Test Account",
		LegalName: "Test Legal Name",
		Type:      AccountTypeClient.String(),
		Email:     "test@example.com",
		Website:   "https://example.com",
	}
	if err := db.Create(&account).Error; err != nil {
		t.Fatalf("Failed to create account: %v", err)
	}

	// Test 1: Create project with HubSpot Deal ID
	now := time.Now()
	oneYearFromNow := now.AddDate(1, 0, 0)
	hubspotDealID := uint(123456789)

	project := Project{
		Name:             "Test Project with HubSpot",
		AccountID:        account.ID,
		ActiveStart:      now,
		ActiveEnd:        oneYearFromNow,
		BudgetHours:      100,
		BudgetDollars:    10000,
		Internal:         false,
		BillingFrequency: "monthly",
		ProjectType:      ProjectTypeNew.String(),
		HubspotDealID:    &hubspotDealID,
	}

	// Save the project
	if err := db.Create(&project).Error; err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Retrieve the project
	var savedProject Project
	if err := db.First(&savedProject, project.ID).Error; err != nil {
		t.Fatalf("Failed to retrieve created project: %v", err)
	}

	// Verify HubSpot Deal ID was saved correctly
	if savedProject.HubspotDealID == nil {
		t.Errorf("Expected HubSpot Deal ID to be set, got nil")
	} else if *savedProject.HubspotDealID != hubspotDealID {
		t.Errorf("Expected HubSpot Deal ID %d, got %d", hubspotDealID, *savedProject.HubspotDealID)
	}

	// Test 2: Create project without HubSpot Deal ID (should be nil)
	projectWithoutHubspot := Project{
		Name:             "Test Project without HubSpot",
		AccountID:        account.ID,
		ActiveStart:      now,
		ActiveEnd:        oneYearFromNow,
		BudgetHours:      50,
		BudgetDollars:    5000,
		Internal:         false,
		BillingFrequency: "monthly",
		ProjectType:      ProjectTypeNew.String(),
	}

	if err := db.Create(&projectWithoutHubspot).Error; err != nil {
		t.Fatalf("Failed to create project without HubSpot: %v", err)
	}

	// Retrieve the project without HubSpot ID
	var savedProjectWithoutHubspot Project
	if err := db.First(&savedProjectWithoutHubspot, projectWithoutHubspot.ID).Error; err != nil {
		t.Fatalf("Failed to retrieve project without HubSpot: %v", err)
	}

	// Verify HubSpot Deal ID is nil
	if savedProjectWithoutHubspot.HubspotDealID != nil {
		t.Errorf("Expected HubSpot Deal ID to be nil, got %v", *savedProjectWithoutHubspot.HubspotDealID)
	}

	// Test 3: Update project to add HubSpot Deal ID
	newHubspotDealID := uint(987654321)
	savedProjectWithoutHubspot.HubspotDealID = &newHubspotDealID

	if err := db.Save(&savedProjectWithoutHubspot).Error; err != nil {
		t.Fatalf("Failed to update project with HubSpot Deal ID: %v", err)
	}

	// Retrieve the updated project
	var updatedProject Project
	if err := db.First(&updatedProject, savedProjectWithoutHubspot.ID).Error; err != nil {
		t.Fatalf("Failed to retrieve updated project: %v", err)
	}

	// Verify HubSpot Deal ID was updated correctly
	if updatedProject.HubspotDealID == nil {
		t.Errorf("Expected HubSpot Deal ID to be set after update, got nil")
	} else if *updatedProject.HubspotDealID != newHubspotDealID {
		t.Errorf("Expected HubSpot Deal ID %d after update, got %d", newHubspotDealID, *updatedProject.HubspotDealID)
	}
}
