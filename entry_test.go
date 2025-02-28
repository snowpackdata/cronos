package cronos

import (
	"testing"
	"time"
)

// TestEntryDuration tests the Duration method of Entry
func TestEntryDuration(t *testing.T) {
	// Test cases
	testCases := []struct {
		name     string
		start    time.Time
		end      time.Time
		expected time.Duration
	}{
		{
			name:     "One Hour",
			start:    time.Date(2023, 1, 1, 9, 0, 0, 0, time.UTC),
			end:      time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC),
			expected: time.Hour,
		},
		{
			name:     "90 Minutes",
			start:    time.Date(2023, 1, 1, 9, 0, 0, 0, time.UTC),
			end:      time.Date(2023, 1, 1, 10, 30, 0, 0, time.UTC),
			expected: time.Hour + 30*time.Minute,
		},
		{
			name:     "Multiple Days",
			start:    time.Date(2023, 1, 1, 9, 0, 0, 0, time.UTC),
			end:      time.Date(2023, 1, 3, 9, 0, 0, 0, time.UTC),
			expected: 48 * time.Hour,
		},
		{
			name:     "Zero Duration",
			start:    time.Date(2023, 1, 1, 9, 0, 0, 0, time.UTC),
			end:      time.Date(2023, 1, 1, 9, 0, 0, 0, time.UTC),
			expected: 0,
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			entry := &Entry{
				Start: tc.start,
				End:   tc.end,
			}
			duration := entry.Duration()
			if duration != tc.expected {
				t.Errorf("Expected duration %v, got %v", tc.expected, duration)
			}
		})
	}
}

// TestEntryGetFee tests the GetFee method of Entry
func TestEntryGetFee(t *testing.T) {
	// Setup test DB
	db := setupTestDB(t)

	// Create a rate
	rate := Rate{
		Name:         "Test Rate",
		Amount:       100.0, // $100 per hour
		ActiveFrom:   time.Now().AddDate(-1, 0, 0),
		ActiveTo:     time.Now().AddDate(1, 0, 0),
		InternalOnly: false,
	}
	if err := db.Create(&rate).Error; err != nil {
		t.Fatalf("Failed to create rate: %v", err)
	}

	// Create an account
	account := Account{
		Name:      "Test Account",
		LegalName: "Test Legal Name",
		Type:      AccountTypeClient.String(),
	}
	if err := db.Create(&account).Error; err != nil {
		t.Fatalf("Failed to create account: %v", err)
	}

	// Create a project
	project := Project{
		Name:        "Test Project",
		AccountID:   account.ID,
		ActiveStart: time.Now().AddDate(-1, 0, 0),
		ActiveEnd:   time.Now().AddDate(1, 0, 0),
	}
	if err := db.Create(&project).Error; err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Create a billing code with 15-minute rounding
	billingCode := BillingCode{
		Name:        "Development",
		RateType:    RateTypeExternalBillable.String(),
		Category:    "Development",
		Code:        "DEV-001",
		RoundedTo:   15, // 15-minute rounding
		ProjectID:   project.ID,
		ActiveStart: time.Now().AddDate(-1, 0, 0),
		ActiveEnd:   time.Now().AddDate(1, 0, 0),
		RateID:      rate.ID,
	}
	if err := db.Create(&billingCode).Error; err != nil {
		t.Fatalf("Failed to create billing code: %v", err)
	}

	// Test cases
	testCases := []struct {
		name     string
		duration time.Duration
		expected float64
	}{
		{
			name:     "One Hour",
			duration: time.Hour,
			expected: 100.0, // 1 hour at $100/hour
		},
		{
			name:     "30 Minutes",
			duration: 30 * time.Minute,
			expected: 50.0, // 0.5 hours at $100/hour
		},
		{
			name:     "5 Minutes (Rounds to 15)",
			duration: 5 * time.Minute,
			expected: 0.0,
		},
		{
			name:     "16 Minutes (Rounds to 30)",
			duration: 16 * time.Minute,
			expected: 25.0,
		},
		{
			name:     "No Duration",
			duration: 0,
			expected: 0,
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			entry := &Entry{
				ProjectID:     project.ID,
				BillingCodeID: billingCode.ID,
				Start:         time.Now(),
				End:           time.Now().Add(tc.duration),
			}

			// Save the entry
			if err := db.Create(&entry).Error; err != nil {
				t.Fatalf("Failed to create entry: %v", err)
			}

			// Get the fee
			fee := entry.GetFee(db)
			if fee != tc.expected {
				t.Errorf("Expected fee %.2f, got %.2f", tc.expected, fee)
			}
		})
	}
}

// TestEntryGetInternalFee tests the GetInternalFee method of Entry
func TestEntryGetInternalFee(t *testing.T) {
	// Setup test DB
	db := setupTestDB(t)

	// Create an external rate
	externalRate := Rate{
		Name:         "External Rate",
		Amount:       100.0, // $100 per hour
		ActiveFrom:   time.Now().AddDate(-1, 0, 0),
		ActiveTo:     time.Now().AddDate(1, 0, 0),
		InternalOnly: false,
	}
	if err := db.Create(&externalRate).Error; err != nil {
		t.Fatalf("Failed to create external rate: %v", err)
	}

	// Create an internal rate
	internalRate := Rate{
		Name:         "Internal Rate",
		Amount:       50.0, // $50 per hour
		ActiveFrom:   time.Now().AddDate(-1, 0, 0),
		ActiveTo:     time.Now().AddDate(1, 0, 0),
		InternalOnly: true,
	}
	if err := db.Create(&internalRate).Error; err != nil {
		t.Fatalf("Failed to create internal rate: %v", err)
	}

	// Create an account
	account := Account{
		Name:      "Test Account",
		LegalName: "Test Legal Name",
		Type:      AccountTypeClient.String(),
	}
	if err := db.Create(&account).Error; err != nil {
		t.Fatalf("Failed to create account: %v", err)
	}

	// Create a project
	project := Project{
		Name:        "Test Project",
		AccountID:   account.ID,
		ActiveStart: time.Now().AddDate(-1, 0, 0),
		ActiveEnd:   time.Now().AddDate(1, 0, 0),
	}
	if err := db.Create(&project).Error; err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Create a billing code with both external and internal rates
	billingCode := BillingCode{
		Name:           "Development",
		RateType:       RateTypeExternalBillable.String(),
		Category:       "Development",
		Code:           "DEV-001",
		RoundedTo:      15, // 15-minute rounding
		ProjectID:      project.ID,
		ActiveStart:    time.Now().AddDate(-1, 0, 0),
		ActiveEnd:      time.Now().AddDate(1, 0, 0),
		RateID:         externalRate.ID,
		InternalRateID: internalRate.ID,
	}
	if err := db.Create(&billingCode).Error; err != nil {
		t.Fatalf("Failed to create billing code: %v", err)
	}

	// Test cases
	testCases := []struct {
		name     string
		duration time.Duration
		expected float64
	}{
		{
			name:     "One Hour",
			duration: time.Hour,
			expected: 50.0, // 1 hour at $50/hour internal rate
		},
		{
			name:     "30 Minutes",
			duration: 30 * time.Minute,
			expected: 25.0, // 0.5 hours at $50/hour internal rate
		},
		{
			name:     "5 Minutes (Rounds to 15)",
			duration: 5 * time.Minute,
			expected: 0.0,
		},
		{
			name:     "16 Minutes (Rounds to 30)",
			duration: 16 * time.Minute,
			expected: 12.5,
		},
		{
			name:     "No Duration",
			duration: 0,
			expected: 0,
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			entry := &Entry{
				ProjectID:     project.ID,
				BillingCodeID: billingCode.ID,
				Start:         time.Now(),
				End:           time.Now().Add(tc.duration),
			}

			// Save the entry
			if err := db.Create(&entry).Error; err != nil {
				t.Fatalf("Failed to create entry: %v", err)
			}

			// Get the internal fee
			fee := entry.GetInternalFee(db)
			if fee != tc.expected {
				t.Errorf("Expected internal fee %.2f, got %.2f", tc.expected, fee)
			}
		})
	}
}

// TestAssociateEntry tests that entries (both regular and impersonated) are correctly
// associated with invoices under various conditions
func TestAssociateEntry(t *testing.T) {
	// Setup test DB
	db := setupTestDB(t)
	app := &App{DB: db}

	// Setup test data
	now := time.Now()
	startDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	// Create users and employees
	user1 := User{
		Email:    "user1@example.com",
		Password: "password",
		IsAdmin:  true,
		Role:     UserRoleStaff.String(),
	}
	if err := db.Create(&user1).Error; err != nil {
		t.Fatalf("Failed to create user1: %v", err)
	}

	user2 := User{
		Email:    "user2@example.com",
		Password: "password",
		IsAdmin:  false,
		Role:     UserRoleStaff.String(),
	}
	if err := db.Create(&user2).Error; err != nil {
		t.Fatalf("Failed to create user2: %v", err)
	}

	employee1 := Employee{
		UserID:    user1.ID,
		FirstName: "John",
		LastName:  "Doe",
		IsActive:  true,
		StartDate: startDate.AddDate(-1, 0, 0),
	}
	if err := db.Create(&employee1).Error; err != nil {
		t.Fatalf("Failed to create employee1: %v", err)
	}

	employee2 := Employee{
		UserID:    user2.ID,
		FirstName: "Jane",
		LastName:  "Smith",
		IsActive:  true,
		StartDate: startDate.AddDate(-1, 0, 0),
	}
	if err := db.Create(&employee2).Error; err != nil {
		t.Fatalf("Failed to create employee2: %v", err)
	}

	// Create rates
	rate := Rate{
		Name:       "Standard Rate",
		Amount:     100.0,
		ActiveFrom: startDate.AddDate(-1, 0, 0),
		ActiveTo:   startDate.AddDate(1, 0, 0),
	}
	if err := db.Create(&rate).Error; err != nil {
		t.Fatalf("Failed to create rate: %v", err)
	}

	// Create accounts with different invoice settings
	accountSingleInvoice := Account{
		Name:                  "SingleInvoiceAccount",
		Type:                  AccountTypeClient.String(),
		LegalName:             "SingleInvoice Inc.",
		ProjectsSingleInvoice: true,
	}
	if err := db.Create(&accountSingleInvoice).Error; err != nil {
		t.Fatalf("Failed to create accountSingleInvoice: %v", err)
	}

	accountMultiInvoice := Account{
		Name:                  "MultiInvoiceAccount",
		Type:                  AccountTypeClient.String(),
		LegalName:             "MultiInvoice Inc.",
		ProjectsSingleInvoice: false,
	}
	if err := db.Create(&accountMultiInvoice).Error; err != nil {
		t.Fatalf("Failed to create accountMultiInvoice: %v", err)
	}

	// Create projects with different date ranges
	projectActive := Project{
		Name:        "Active Project",
		AccountID:   accountSingleInvoice.ID,
		ActiveStart: startDate.AddDate(0, -1, 0),
		ActiveEnd:   startDate.AddDate(0, 1, 0),
		Internal:    false,
	}
	if err := db.Create(&projectActive).Error; err != nil {
		t.Fatalf("Failed to create projectActive: %v", err)
	}

	projectMultiInvoice := Project{
		Name:        "Multi-Invoice Project",
		AccountID:   accountMultiInvoice.ID,
		ActiveStart: startDate.AddDate(0, -1, 0),
		ActiveEnd:   startDate.AddDate(0, 1, 0),
		Internal:    false,
	}
	if err := db.Create(&projectMultiInvoice).Error; err != nil {
		t.Fatalf("Failed to create projectMultiInvoice: %v", err)
	}

	projectInactive := Project{
		Name:        "Inactive Project",
		AccountID:   accountSingleInvoice.ID,
		ActiveStart: startDate.AddDate(0, -3, 0),
		ActiveEnd:   startDate.AddDate(0, -2, 0), // Ended in the past
		Internal:    false,
	}
	if err := db.Create(&projectInactive).Error; err != nil {
		t.Fatalf("Failed to create projectInactive: %v", err)
	}

	projectInternal := Project{
		Name:        "Internal Project",
		AccountID:   accountSingleInvoice.ID,
		ActiveStart: startDate.AddDate(0, -1, 0),
		ActiveEnd:   startDate.AddDate(0, 1, 0),
		Internal:    true,
	}
	if err := db.Create(&projectInternal).Error; err != nil {
		t.Fatalf("Failed to create projectInternal: %v", err)
	}

	// Create billing codes
	billingCodeActive := BillingCode{
		Name:        "Active Billing Code",
		RateType:    RateTypeExternalBillable.String(),
		Category:    "Development",
		Code:        "DEV-001",
		RoundedTo:   15,
		ProjectID:   projectActive.ID,
		ActiveStart: startDate.AddDate(0, -1, 0),
		ActiveEnd:   startDate.AddDate(0, 1, 0),
		RateID:      rate.ID,
	}
	if err := db.Create(&billingCodeActive).Error; err != nil {
		t.Fatalf("Failed to create billingCodeActive: %v", err)
	}

	billingCodeMultiInvoice := BillingCode{
		Name:        "Multi-Invoice Billing Code",
		RateType:    RateTypeExternalBillable.String(),
		Category:    "Development",
		Code:        "DEV-002",
		RoundedTo:   15,
		ProjectID:   projectMultiInvoice.ID,
		ActiveStart: startDate.AddDate(0, -1, 0),
		ActiveEnd:   startDate.AddDate(0, 1, 0),
		RateID:      rate.ID,
	}
	if err := db.Create(&billingCodeMultiInvoice).Error; err != nil {
		t.Fatalf("Failed to create billingCodeMultiInvoice: %v", err)
	}

	billingCodeInactive := BillingCode{
		Name:        "Inactive Billing Code",
		RateType:    RateTypeExternalBillable.String(),
		Category:    "Development",
		Code:        "DEV-003",
		RoundedTo:   15,
		ProjectID:   projectInactive.ID,
		ActiveStart: startDate.AddDate(0, -3, 0),
		ActiveEnd:   startDate.AddDate(0, -2, 0),
		RateID:      rate.ID,
	}
	if err := db.Create(&billingCodeInactive).Error; err != nil {
		t.Fatalf("Failed to create billingCodeInactive: %v", err)
	}

	billingCodeInternal := BillingCode{
		Name:        "Internal Billing Code",
		RateType:    RateTypeInternalProject.String(),
		Category:    "Internal",
		Code:        "INT-001",
		RoundedTo:   15,
		ProjectID:   projectInternal.ID,
		ActiveStart: startDate.AddDate(0, -1, 0),
		ActiveEnd:   startDate.AddDate(0, 1, 0),
		RateID:      rate.ID,
	}
	if err := db.Create(&billingCodeInternal).Error; err != nil {
		t.Fatalf("Failed to create billingCodeInternal: %v", err)
	}

	// Set up the mock bucket and project for the app
	app.Project = "test-project"
	app.Bucket = "test-bucket"

	// Helper function to validate invoice association
	validateInvoiceAssociation := func(t *testing.T, entry *Entry, projectID uint, shouldBeAssociated bool) {
		// First clear any existing invoice associations
		entry.InvoiceID = nil
		db.Save(entry)

		// Call the function to test
		err := app.AssociateEntry(entry, projectID)

		// If we're testing a case that should fail
		if !shouldBeAssociated {
			if err == nil {
				t.Errorf("Expected AssociateEntry to fail for entry in inactive project, but it succeeded")
			}
			return
		}

		// For cases that should succeed
		if err != nil {
			t.Errorf("AssociateEntry failed: %v", err)
			return
		}

		// Reload the entry to check if it was associated
		var updatedEntry Entry
		if err := db.First(&updatedEntry, entry.ID).Error; err != nil {
			t.Fatalf("Failed to reload entry: %v", err)
		}

		// For internal entries, they shouldn't be associated
		if entry.Internal {
			if updatedEntry.InvoiceID != nil {
				t.Errorf("Internal entry should not be associated with an invoice")
			}
			return
		}

		// For regular entries, they should be associated
		if updatedEntry.InvoiceID == nil {
			t.Errorf("Entry was not associated with an invoice")
			return
		}

		// Load the invoice to verify it's correct
		var invoice Invoice
		if err := db.First(&invoice, *updatedEntry.InvoiceID).Error; err != nil {
			t.Fatalf("Failed to load invoice: %v", err)
		}

		// Verify the invoice is correctly created for the right account/project
		project := Project{}
		if err := db.First(&project, projectID).Error; err != nil {
			t.Fatalf("Failed to load project: %v", err)
		}

		if invoice.AccountID != project.AccountID {
			t.Errorf("Invoice associated with wrong account: expected %d, got %d", project.AccountID, invoice.AccountID)
		}

		// For multi-invoice accounts, check project association
		account := Account{}
		if err := db.First(&account, project.AccountID).Error; err != nil {
			t.Fatalf("Failed to load account: %v", err)
		}

		if !account.ProjectsSingleInvoice {
			if invoice.ProjectID == nil || *invoice.ProjectID != projectID {
				t.Errorf("Invoice not correctly associated with project")
			}
		}
	}

	// Create initial invoices to ensure they are available in the database
	// This is needed to make sure the test environment has the necessary data
	activeInvoice := Invoice{
		Name:        "Active Invoice",
		AccountID:   accountSingleInvoice.ID,
		PeriodStart: startDate.AddDate(0, 0, -15),
		PeriodEnd:   startDate.AddDate(0, 0, 15),
		State:       InvoiceStateDraft.String(),
		Type:        InvoiceTypeAR.String(),
	}
	if err := db.Create(&activeInvoice).Error; err != nil {
		t.Fatalf("Failed to create active invoice: %v", err)
	}

	// Test 1: Entry is created in viable billing code and project window -> Success
	t.Run("Entry in viable window", func(t *testing.T) {
		entry := Entry{
			EmployeeID:    employee1.ID,
			BillingCodeID: billingCodeActive.ID,
			Start:         startDate,
			End:           startDate.Add(1 * time.Hour),
			Notes:         "Test entry in viable window",
		}
		if err := db.Create(&entry).Error; err != nil {
			t.Fatalf("Failed to create entry: %v", err)
		}

		validateInvoiceAssociation(t, &entry, projectActive.ID, true)
	})

	// Test 2: Entry is created outside of viable window -> Failure
	t.Run("Entry outside viable window", func(t *testing.T) {
		entry := Entry{
			EmployeeID:    employee1.ID,
			BillingCodeID: billingCodeInactive.ID,
			Start:         startDate,
			End:           startDate.Add(1 * time.Hour),
			Notes:         "Test entry outside viable window",
		}
		if err := db.Create(&entry).Error; err != nil {
			t.Fatalf("Failed to create entry: %v", err)
		}

		validateInvoiceAssociation(t, &entry, projectInactive.ID, false)
	})

	// Test 3: Existing Invoice exists -> Add it to the correct invoice
	t.Run("Entry with existing invoice", func(t *testing.T) {
		// Now create an entry that should be associated with this invoice
		entry := Entry{
			EmployeeID:    employee1.ID,
			BillingCodeID: billingCodeActive.ID,
			Start:         startDate,
			End:           startDate.Add(1 * time.Hour),
			Notes:         "Test entry with existing invoice",
		}
		if err := db.Create(&entry).Error; err != nil {
			t.Fatalf("Failed to create entry: %v", err)
		}

		validateInvoiceAssociation(t, &entry, projectActive.ID, true)

		// Reload the entry to verify it was associated with the existing invoice
		var updatedEntry Entry
		if err := db.First(&updatedEntry, entry.ID).Error; err != nil {
			t.Fatalf("Failed to reload entry: %v", err)
		}

		if updatedEntry.InvoiceID == nil || *updatedEntry.InvoiceID != activeInvoice.ID {
			t.Errorf("Entry was not associated with the existing invoice")
		}
	})

	// Test 4: No invoice exists -> Create the correct invoice
	t.Run("Entry with no existing invoice", func(t *testing.T) {
		// Delete all existing invoices for this account to ensure we start clean
		db.Where("account_id = ?", accountMultiInvoice.ID).Delete(&Invoice{})

		// Create a draft multi-invoice explicitly to ensure one exists
		multiInvoice := Invoice{
			Name:        "Multi Invoice",
			AccountID:   accountMultiInvoice.ID,
			ProjectID:   &projectMultiInvoice.ID,
			PeriodStart: startDate.AddDate(0, 0, -15),
			PeriodEnd:   startDate.AddDate(0, 0, 15),
			State:       InvoiceStateDraft.String(),
			Type:        InvoiceTypeAR.String(),
		}
		if err := db.Create(&multiInvoice).Error; err != nil {
			t.Fatalf("Failed to create multi invoice: %v", err)
		}

		// Create an entry that should be associated with this invoice
		entry := Entry{
			EmployeeID:    employee1.ID,
			BillingCodeID: billingCodeMultiInvoice.ID,
			Start:         startDate,
			End:           startDate.Add(1 * time.Hour),
			Notes:         "Test entry with existing multi-invoice",
		}
		if err := db.Create(&entry).Error; err != nil {
			t.Fatalf("Failed to create entry: %v", err)
		}

		validateInvoiceAssociation(t, &entry, projectMultiInvoice.ID, true)

		// Reload the entry to verify it was associated with the multi invoice
		var updatedEntry Entry
		if err := db.First(&updatedEntry, entry.ID).Error; err != nil {
			t.Fatalf("Failed to reload entry: %v", err)
		}

		if updatedEntry.InvoiceID == nil || *updatedEntry.InvoiceID != multiInvoice.ID {
			t.Errorf("Entry was not associated with the multi invoice")
		}
	})

	// Test 5: Impersonated Entry - should work the same as normal entry
	t.Run("Impersonated entry in viable window", func(t *testing.T) {
		// Create an entry with impersonation
		impersonateID := employee2.ID
		entry := Entry{
			EmployeeID:          employee1.ID,   // Employee1 creates entry
			ImpersonateAsUserID: &impersonateID, // As Employee2
			BillingCodeID:       billingCodeActive.ID,
			Start:               startDate,
			End:                 startDate.Add(1 * time.Hour),
			Notes:               "Test impersonated entry",
		}
		if err := db.Create(&entry).Error; err != nil {
			t.Fatalf("Failed to create entry: %v", err)
		}

		validateInvoiceAssociation(t, &entry, projectActive.ID, true)

		// Retrieve the updated entry to verify it maintained impersonation data
		var updatedEntry Entry
		if err := db.Preload("ImpersonateAsUser").First(&updatedEntry, entry.ID).Error; err != nil {
			t.Fatalf("Failed to reload entry: %v", err)
		}

		if updatedEntry.ImpersonateAsUserID == nil || *updatedEntry.ImpersonateAsUserID != employee2.ID {
			t.Errorf("Impersonation data was lost during invoice association")
		}
	})

	// Test 6: Internal Entry - should not be associated with an invoice
	t.Run("Internal entry not associated with invoice", func(t *testing.T) {
		entry := Entry{
			EmployeeID:    employee1.ID,
			BillingCodeID: billingCodeInternal.ID,
			Start:         startDate,
			End:           startDate.Add(1 * time.Hour),
			Notes:         "Test internal entry",
			Internal:      true,
		}
		if err := db.Create(&entry).Error; err != nil {
			t.Fatalf("Failed to create entry: %v", err)
		}

		validateInvoiceAssociation(t, &entry, projectInternal.ID, true) // Should succeed but not associate

		// Verify the entry wasn't associated with an invoice
		var updatedEntry Entry
		if err := db.First(&updatedEntry, entry.ID).Error; err != nil {
			t.Fatalf("Failed to reload entry: %v", err)
		}

		if updatedEntry.InvoiceID != nil {
			t.Errorf("Internal entry should not be associated with an invoice")
		}
	})

	// Test 7: Impersonated Internal Entry - should also not be associated
	t.Run("Impersonated internal entry not associated with invoice", func(t *testing.T) {
		impersonateID := employee2.ID
		entry := Entry{
			EmployeeID:          employee1.ID,
			ImpersonateAsUserID: &impersonateID,
			BillingCodeID:       billingCodeInternal.ID,
			Start:               startDate,
			End:                 startDate.Add(1 * time.Hour),
			Notes:               "Test impersonated internal entry",
			Internal:            true,
		}
		if err := db.Create(&entry).Error; err != nil {
			t.Fatalf("Failed to create entry: %v", err)
		}

		validateInvoiceAssociation(t, &entry, projectInternal.ID, true) // Should succeed but not associate

		// Verify the entry wasn't associated with an invoice
		var updatedEntry Entry
		if err := db.First(&updatedEntry, entry.ID).Error; err != nil {
			t.Fatalf("Failed to reload entry: %v", err)
		}

		if updatedEntry.InvoiceID != nil {
			t.Errorf("Impersonated internal entry should not be associated with an invoice")
		}

		// Verify impersonation data was preserved
		if updatedEntry.ImpersonateAsUserID == nil || *updatedEntry.ImpersonateAsUserID != employee2.ID {
			t.Errorf("Impersonation data was lost for internal entry")
		}
	})
}
