package cronos

import (
	"fmt"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	// Use in-memory SQLite for testing with a unique identifier to prevent conflicts
	dbName := fmt.Sprintf("file:memdb_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Create app instance and migrate schema
	app := &App{DB: db}
	app.Migrate()

	return db
}

// TestModelCreation tests that models can be created and retrieved from the database
func TestModelCreation(t *testing.T) {
	// Setup
	db := setupTestDB(t)

	// Create a user
	user := User{
		Email:    "test@example.com",
		Password: "password123",
		IsAdmin:  true,
		Role:     UserRoleAdmin.String(),
	}

	// Save the user
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Test user was created with ID
	if user.ID == 0 {
		t.Errorf("Expected user ID to be set, got %v", user.ID)
	}

	// Create an account
	account := Account{
		Name:                  "Test Account",
		LegalName:             "Test Legal Name",
		Type:                  AccountTypeClient.String(),
		Email:                 "account@example.com",
		Website:               "https://example.com",
		BillingFrequency:      BillingFrequencyMonthly.String(),
		BudgetHours:           100,
		BudgetDollars:         10000,
		ProjectsSingleInvoice: false,
	}

	// Save the account
	if err := db.Create(&account).Error; err != nil {
		t.Fatalf("Failed to create account: %v", err)
	}

	// Create an employee linked to the user
	employee := Employee{
		UserID:    user.ID,
		Title:     "Software Engineer",
		FirstName: "Test",
		LastName:  "User",
		IsActive:  true,
		StartDate: time.Now(),
	}

	// Save the employee
	if err := db.Create(&employee).Error; err != nil {
		t.Fatalf("Failed to create employee: %v", err)
	}

	// Create a rate
	rate := Rate{
		Name:         "Standard Rate",
		Amount:       125.0,
		ActiveFrom:   time.Now(),
		ActiveTo:     time.Now().AddDate(1, 0, 0), // 1 year from now
		InternalOnly: false,
	}

	// Save the rate
	if err := db.Create(&rate).Error; err != nil {
		t.Fatalf("Failed to create rate: %v", err)
	}

	// Create a project
	project := Project{
		Name:             "Test Project",
		AccountID:        account.ID,
		ActiveStart:      time.Now(),
		ActiveEnd:        time.Now().AddDate(0, 3, 0), // 3 months from now
		BudgetHours:      80,
		BudgetDollars:    8000,
		Internal:         false,
		BillingFrequency: BillingFrequencyMonthly.String(),
		ProjectType:      ProjectTypeNew.String(),
		AEID:             &employee.ID,
	}

	// Save the project
	if err := db.Create(&project).Error; err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Create a billing code
	billingCode := BillingCode{
		Name:        "Development",
		RateType:    RateTypeExternalBillable.String(),
		Category:    "Development",
		Code:        "DEV-001",
		RoundedTo:   15,
		ProjectID:   project.ID,
		ActiveStart: time.Now(),
		ActiveEnd:   time.Now().AddDate(0, 3, 0),
		RateID:      rate.ID,
	}

	// Save the billing code
	if err := db.Create(&billingCode).Error; err != nil {
		t.Fatalf("Failed to create billing code: %v", err)
	}

	// Create an entry
	entry := Entry{
		ProjectID:     project.ID,
		Notes:         "Working on feature X",
		EmployeeID:    employee.ID,
		BillingCodeID: billingCode.ID,
		Start:         time.Now().Add(-2 * time.Hour),
		End:           time.Now().Add(-1 * time.Hour),
		Internal:      false,
		State:         EntryStateApproved.String(),
	}

	// Save the entry
	if err := db.Create(&entry).Error; err != nil {
		t.Fatalf("Failed to create entry: %v", err)
	}

	// Test: Retrieve the entry with associations
	var retrievedEntry Entry
	if err := db.Preload("Project").Preload("BillingCode").Preload("Employee").First(&retrievedEntry, entry.ID).Error; err != nil {
		t.Fatalf("Failed to retrieve entry: %v", err)
	}

	// Verify the associations are loaded correctly
	if retrievedEntry.Project.ID != project.ID {
		t.Errorf("Expected project ID %v, got %v", project.ID, retrievedEntry.Project.ID)
	}

	if retrievedEntry.BillingCode.ID != billingCode.ID {
		t.Errorf("Expected billing code ID %v, got %v", billingCode.ID, retrievedEntry.BillingCode.ID)
	}

	if retrievedEntry.Employee.ID != employee.ID {
		t.Errorf("Expected employee ID %v, got %v", employee.ID, retrievedEntry.Employee.ID)
	}
}

// TestCalculateCommissionRate tests the CalculateCommissionRate function
func TestCalculateCommissionRate(t *testing.T) {
	// Setup
	db := setupTestDB(t)
	app := &App{DB: db}

	// Test cases
	testCases := []struct {
		name        string
		role        string
		projectType string
		dealSize    int
		expected    float64
	}{
		// AE commission rates
		{"AE New Small Deal", CommissionRoleAE.String(), ProjectTypeNew.String(), 50000, AECommissionRateNewSmall},
		{"AE New Large Deal", CommissionRoleAE.String(), ProjectTypeNew.String(), 150000, AECommissionRateNewLarge},
		{"AE Existing Small Deal", CommissionRoleAE.String(), ProjectTypeExisting.String(), 50000, AECommissionRateExistingSmall},
		{"AE Existing Large Deal", CommissionRoleAE.String(), ProjectTypeExisting.String(), 150000, AECommissionRateExistingLarge},

		// SDR commission rates
		{"SDR New Small Deal", CommissionRoleSDR.String(), ProjectTypeNew.String(), 50000, SDRCommissionRateNewSmall},
		{"SDR New Large Deal", CommissionRoleSDR.String(), ProjectTypeNew.String(), 150000, SDRCommissionRateNewLarge},
		{"SDR Existing Small Deal", CommissionRoleSDR.String(), ProjectTypeExisting.String(), 50000, SDRCommissionRateExistingSmall},
		{"SDR Existing Large Deal", CommissionRoleSDR.String(), ProjectTypeExisting.String(), 150000, SDRCommissionRateExistingLarge},

		// Edge cases
		{"AE Deal at threshold", CommissionRoleAE.String(), ProjectTypeNew.String(), DealSizeSmallThreshold, AECommissionRateNewLarge},
		{"SDR Deal at threshold", CommissionRoleSDR.String(), ProjectTypeNew.String(), DealSizeSmallThreshold, SDRCommissionRateNewLarge},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rate := app.CalculateCommissionRate(tc.role, tc.projectType, tc.dealSize)
			if rate != tc.expected {
				t.Errorf("Expected rate %v, got %v", tc.expected, rate)
			}
		})
	}
}

// TestCalculateCommissionAmount tests the CalculateCommissionAmount function
func TestCalculateCommissionAmount(t *testing.T) {
	// Setup
	db := setupTestDB(t)
	app := &App{DB: db}

	// Create a project for testing
	newProject := &Project{
		Name:        "New Test Project",
		ProjectType: ProjectTypeNew.String(),
	}

	existingProject := &Project{
		Name:        "Existing Test Project",
		ProjectType: ProjectTypeExisting.String(),
	}

	// Test cases
	testCases := []struct {
		name         string
		project      *Project
		role         string
		invoiceTotal float64
		expected     int
	}{
		{"AE New Project Small", newProject, CommissionRoleAE.String(), 50000, int(50000 * AECommissionRateNewSmall * 100)},
		{"AE New Project Large", newProject, CommissionRoleAE.String(), 150000, int(150000 * AECommissionRateNewLarge * 100)},
		{"SDR Existing Project Small", existingProject, CommissionRoleSDR.String(), 50000, int(50000 * SDRCommissionRateExistingSmall * 100)},
		{"SDR Existing Project Large", existingProject, CommissionRoleSDR.String(), 150000, int(150000 * SDRCommissionRateExistingLarge * 100)},

		// Edge case - zero invoice
		{"Zero Invoice", newProject, CommissionRoleAE.String(), 0, 0},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			amount := app.CalculateCommissionAmount(tc.project, tc.role, tc.invoiceTotal)
			if amount != tc.expected {
				t.Errorf("Expected amount %v, got %v", tc.expected, amount)
			}
		})
	}
}

// TestTimeCalculationFunctions tests the various time calculation helper functions
func TestTimeCalculationFunctions(t *testing.T) {
	// Test cases for calculateMonths
	monthTestCases := []struct {
		name     string
		start    time.Time
		end      time.Time
		expected int
	}{
		{"One Month", time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2023, 2, 1, 0, 0, 0, 0, time.UTC), 1},
		{"Two Months", time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2023, 3, 1, 0, 0, 0, 0, time.UTC), 2},
		{"Partial Month", time.Date(2023, 1, 15, 0, 0, 0, 0, time.UTC), time.Date(2023, 2, 10, 0, 0, 0, 0, time.UTC), 1},
		{"Crossing Year", time.Date(2022, 12, 1, 0, 0, 0, 0, time.UTC), time.Date(2023, 2, 1, 0, 0, 0, 0, time.UTC), 2},
		{"Same Day", time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC), 1}, // Minimum of 1
	}

	for _, tc := range monthTestCases {
		t.Run(tc.name, func(t *testing.T) {
			result := calculateMonths(tc.start, tc.end)
			if result != tc.expected {
				t.Errorf("Expected %d months, got %d", tc.expected, result)
			}
		})
	}

	// Test cases for calculateWeeks
	weekTestCases := []struct {
		name     string
		start    time.Time
		end      time.Time
		expected int
	}{
		{"One Week", time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2023, 1, 8, 0, 0, 0, 0, time.UTC), 1},
		{"Two Weeks", time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2023, 1, 15, 0, 0, 0, 0, time.UTC), 2},
		{"Partial Week", time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2023, 1, 4, 0, 0, 0, 0, time.UTC), 1},
		{"Less Than One Day", time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC), 1}, // Minimum of 1
	}

	for _, tc := range weekTestCases {
		t.Run(tc.name, func(t *testing.T) {
			result := calculateWeeks(tc.start, tc.end)
			if result != tc.expected {
				t.Errorf("Expected %d weeks, got %d", tc.expected, result)
			}
		})
	}

	// Test cases for calculateBiWeeks
	biWeekTestCases := []struct {
		name     string
		start    time.Time
		end      time.Time
		expected int
	}{
		{"One BiWeek", time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2023, 1, 15, 0, 0, 0, 0, time.UTC), 1},
		{"Two BiWeeks", time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2023, 1, 29, 0, 0, 0, 0, time.UTC), 2},
		{"Partial BiWeek", time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2023, 1, 10, 0, 0, 0, 0, time.UTC), 1},
		{"Less Than One Day", time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC), 1}, // Minimum of 1
	}

	for _, tc := range biWeekTestCases {
		t.Run(tc.name, func(t *testing.T) {
			result := calculateBiWeeks(tc.start, tc.end)
			if result != tc.expected {
				t.Errorf("Expected %d bi-weeks, got %d", tc.expected, result)
			}
		})
	}
}

// TestValidateExpenseDate_FutureDate tests that future dates are allowed
func TestValidateExpenseDate_FutureDate(t *testing.T) {
	// Future dates should be allowed (for prepaid/scheduled expenses)
	futureDate := time.Now().AddDate(0, 0, 7) // 7 days in the future
	err := ValidateExpenseDate(futureDate)
	if err != nil {
		t.Errorf("expected nil error for future date, got: %v", err)
	}
}

// TestValidateExpenseDate_TooOld tests that expenses older than 1 year are rejected
func TestValidateExpenseDate_TooOld(t *testing.T) {
	// Should reject expense more than 1 year old
	oldDate := time.Now().AddDate(-2, 0, 0) // 2 years ago
	err := ValidateExpenseDate(oldDate)
	if err == nil {
		t.Error("expected error for date too far in past, got nil")
	}
}

// TestValidateExpenseDate_ValidDate tests that valid historical dates are accepted
func TestValidateExpenseDate_ValidDate(t *testing.T) {
	// Should accept valid historical date
	validDate := time.Now().AddDate(0, -6, 0) // 6 months ago
	err := ValidateExpenseDate(validDate)
	if err != nil {
		t.Errorf("expected nil error for valid date, got: %v", err)
	}
}

// TestValidateExpenseDate_EdgeCase tests the boundary at exactly 1 year ago
func TestValidateExpenseDate_EdgeCase(t *testing.T) {
	// Date exactly at the 1 year boundary should be accepted
	oneYearAgo := time.Now().AddDate(-1, 0, 0)
	err := ValidateExpenseDate(oneYearAgo)
	if err != nil {
		t.Errorf("expected nil error for date exactly 1 year ago, got: %v", err)
	}

	// Date just over 1 year ago should be rejected
	justOverOneYearAgo := time.Now().AddDate(-1, 0, -1)
	err = ValidateExpenseDate(justOverOneYearAgo)
	if err == nil {
		t.Error("expected error for date just over 1 year ago, got nil")
	}
}

// TestApproveExpense_RejectsOldDate tests that expense approval fails for old-dated expense
func TestApproveExpense_RejectsOldDate(t *testing.T) {
	// Setup test database
	db := setupTestDB(t)
	app := &App{DB: db}

	// Create required expense category
	category := ExpenseCategory{
		Name:          "Test Category",
		Description:   "Test Description",
		GLAccountCode: "OPERATING_EXPENSES_GENERAL",
		Active:        true,
	}
	if err := db.Create(&category).Error; err != nil {
		t.Fatalf("Failed to create expense category: %v", err)
	}

	// Create a user and employee for the expense submitter
	user := User{
		Email:    "submitter@example.com",
		Password: "password123",
		Role:     UserRoleStaff.String(),
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	employee := Employee{
		UserID:    user.ID,
		FirstName: "Test",
		LastName:  "Submitter",
		IsActive:  true,
		StartDate: time.Now().AddDate(-3, 0, 0),
	}
	if err := db.Create(&employee).Error; err != nil {
		t.Fatalf("Failed to create employee: %v", err)
	}

	// Create an internal expense with a date more than 1 year ago
	oldDate := time.Now().AddDate(-2, 0, 0) // 2 years ago
	expense := Expense{
		SubmitterID:  employee.ID,
		Amount:       10000, // $100.00
		Date:         oldDate,
		Description:  "Old expense for testing",
		State:        ExpenseStateSubmitted.String(),
		CategoryID:   category.ID,
		IsReimbursable: false,
	}
	if err := db.Create(&expense).Error; err != nil {
		t.Fatalf("Failed to create expense: %v", err)
	}

	// Attempt to approve - should fail due to old date
	err := app.ApproveExpense(expense.ID, employee.ID)
	if err == nil {
		t.Error("expected error when approving expense with old date, got nil")
	}

	// Verify the error message mentions the date issue
	if err != nil && !containsString(err.Error(), "too old") {
		t.Errorf("expected error about old date, got: %v", err)
	}
}

// containsString is a helper to check if a string contains a substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStringHelper(s, substr))
}

func containsStringHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
