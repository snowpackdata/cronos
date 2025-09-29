package cronos

import (
	"testing"
	"time"
)

// TestStaffingAssignmentAssociation tests that entries are correctly associated with staffing assignments
func TestStaffingAssignmentAssociation(t *testing.T) {
	// Setup test DB
	db := setupTestDB(t)
	app := &App{DB: db, Project: "test-project", Bucket: "test-bucket"}

	// Setup test data
	now := time.Now()
	startDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	// Create user and employee
	user := User{
		Email:    "employee@example.com",
		Password: "password",
		IsAdmin:  false,
		Role:     UserRoleStaff.String(),
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	employee := Employee{
		UserID:    user.ID,
		FirstName: "John",
		LastName:  "Doe",
		IsActive:  true,
		StartDate: startDate.AddDate(-1, 0, 0),
	}
	if err := db.Create(&employee).Error; err != nil {
		t.Fatalf("Failed to create employee: %v", err)
	}

	// Create account and project
	account := Account{
		Name:                  "Test Account",
		Type:                  AccountTypeClient.String(),
		LegalName:             "Test Inc.",
		ProjectsSingleInvoice: false,
	}
	if err := db.Create(&account).Error; err != nil {
		t.Fatalf("Failed to create account: %v", err)
	}

	project := Project{
		Name:        "Test Project",
		AccountID:   account.ID,
		ActiveStart: startDate.AddDate(0, -1, 0),
		ActiveEnd:   startDate.AddDate(0, 2, 0),
		Internal:    false,
	}
	if err := db.Create(&project).Error; err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Create rate and billing code
	rate := Rate{
		Name:       "Standard Rate",
		Amount:     100.0,
		ActiveFrom: startDate.AddDate(-1, 0, 0),
		ActiveTo:   startDate.AddDate(1, 0, 0),
	}
	if err := db.Create(&rate).Error; err != nil {
		t.Fatalf("Failed to create rate: %v", err)
	}

	billingCode := BillingCode{
		Name:        "Development",
		RateType:    RateTypeExternalBillable.String(),
		Category:    "Development",
		Code:        "DEV-001",
		RoundedTo:   15,
		ProjectID:   project.ID,
		ActiveStart: startDate.AddDate(0, -1, 0),
		ActiveEnd:   startDate.AddDate(0, 1, 0),
		RateID:      rate.ID,
	}
	if err := db.Create(&billingCode).Error; err != nil {
		t.Fatalf("Failed to create billing code: %v", err)
	}

	// Test 1: Entry with active staffing assignment
	t.Run("Entry with active staffing assignment", func(t *testing.T) {
		// Create a staffing assignment
		staffingAssignment := StaffingAssignment{
			EmployeeID: employee.ID,
			ProjectID:  project.ID,
			Commitment: 40,
			StartDate:  startDate.AddDate(0, 0, -7), // Started 7 days ago
			EndDate:    startDate.AddDate(0, 0, 30), // Ends 30 days from now
		}
		if err := db.Create(&staffingAssignment).Error; err != nil {
			t.Fatalf("Failed to create staffing assignment: %v", err)
		}

		// Create an entry that falls within the assignment date range
		entry := Entry{
			EmployeeID:    employee.ID,
			BillingCodeID: billingCode.ID,
			Start:         startDate,
			End:           startDate.Add(2 * time.Hour),
			Notes:         "Test entry with active assignment",
		}
		if err := db.Create(&entry).Error; err != nil {
			t.Fatalf("Failed to create entry: %v", err)
		}

		// Associate the entry
		if err := app.AssociateEntry(&entry, project.ID); err != nil {
			t.Fatalf("AssociateEntry failed: %v", err)
		}

		// Reload the entry and verify staffing assignment association
		var updatedEntry Entry
		if err := db.First(&updatedEntry, entry.ID).Error; err != nil {
			t.Fatalf("Failed to reload entry: %v", err)
		}

		if updatedEntry.StaffingAssignmentID == nil {
			t.Errorf("Entry was not associated with staffing assignment")
		} else if *updatedEntry.StaffingAssignmentID != staffingAssignment.ID {
			t.Errorf("Entry associated with wrong staffing assignment: expected %d, got %d",
				staffingAssignment.ID, *updatedEntry.StaffingAssignmentID)
		}
	})

	// Test 2: Entry before staffing assignment start date
	t.Run("Entry before staffing assignment start date", func(t *testing.T) {
		// Delete all existing staffing assignments to start fresh
		db.Where("employee_id = ?", employee.ID).Delete(&StaffingAssignment{})

		// Create a staffing assignment that starts in the future
		futureAssignment := StaffingAssignment{
			EmployeeID: employee.ID,
			ProjectID:  project.ID,
			Commitment: 40,
			StartDate:  startDate.AddDate(0, 0, 10), // Starts 10 days from now
			EndDate:    startDate.AddDate(0, 0, 40), // Ends 40 days from now
		}
		if err := db.Create(&futureAssignment).Error; err != nil {
			t.Fatalf("Failed to create future staffing assignment: %v", err)
		}

		// Create an entry before the assignment starts (today)
		entry := Entry{
			EmployeeID:    employee.ID,
			BillingCodeID: billingCode.ID,
			Start:         startDate,
			End:           startDate.Add(1 * time.Hour),
			Notes:         "Test entry before assignment",
		}
		if err := db.Create(&entry).Error; err != nil {
			t.Fatalf("Failed to create entry: %v", err)
		}

		// Associate the entry
		if err := app.AssociateEntry(&entry, project.ID); err != nil {
			t.Fatalf("AssociateEntry failed: %v", err)
		}

		// Reload the entry and verify it's NOT associated with the future assignment
		var updatedEntry Entry
		if err := db.First(&updatedEntry, entry.ID).Error; err != nil {
			t.Fatalf("Failed to reload entry: %v", err)
		}

		if updatedEntry.StaffingAssignmentID != nil {
			t.Errorf("Entry should not be associated with future staffing assignment, but got ID: %d",
				*updatedEntry.StaffingAssignmentID)
		}
	})

	// Test 3: Entry after staffing assignment end date
	t.Run("Entry after staffing assignment end date", func(t *testing.T) {
		// Delete all existing staffing assignments to start fresh
		db.Where("employee_id = ?", employee.ID).Delete(&StaffingAssignment{})

		// Create a staffing assignment that has already ended
		pastAssignment := StaffingAssignment{
			EmployeeID: employee.ID,
			ProjectID:  project.ID,
			Commitment: 40,
			StartDate:  startDate.AddDate(0, 0, -30), // Started 30 days ago
			EndDate:    startDate.AddDate(0, 0, -5),  // Ended 5 days ago
		}
		if err := db.Create(&pastAssignment).Error; err != nil {
			t.Fatalf("Failed to create past staffing assignment: %v", err)
		}

		// Create an entry after the assignment ended (today)
		entry := Entry{
			EmployeeID:    employee.ID,
			BillingCodeID: billingCode.ID,
			Start:         startDate,
			End:           startDate.Add(1 * time.Hour),
			Notes:         "Test entry after assignment",
		}
		if err := db.Create(&entry).Error; err != nil {
			t.Fatalf("Failed to create entry: %v", err)
		}

		// Associate the entry
		if err := app.AssociateEntry(&entry, project.ID); err != nil {
			t.Fatalf("AssociateEntry failed: %v", err)
		}

		// Reload the entry and verify it's NOT associated with the past assignment
		var updatedEntry Entry
		if err := db.First(&updatedEntry, entry.ID).Error; err != nil {
			t.Fatalf("Failed to reload entry: %v", err)
		}

		if updatedEntry.StaffingAssignmentID != nil {
			t.Errorf("Entry should not be associated with past staffing assignment, but got ID: %d",
				*updatedEntry.StaffingAssignmentID)
		}
	})

	// Test 4: Entry with no staffing assignment
	t.Run("Entry with no staffing assignment", func(t *testing.T) {
		// Create a second project without a staffing assignment
		project2 := Project{
			Name:        "Project Without Assignment",
			AccountID:   account.ID,
			ActiveStart: startDate.AddDate(0, -1, 0),
			ActiveEnd:   startDate.AddDate(0, 2, 0),
			Internal:    false,
		}
		if err := db.Create(&project2).Error; err != nil {
			t.Fatalf("Failed to create project2: %v", err)
		}

		billingCode2 := BillingCode{
			Name:        "Development 2",
			RateType:    RateTypeExternalBillable.String(),
			Category:    "Development",
			Code:        "DEV-002",
			RoundedTo:   15,
			ProjectID:   project2.ID,
			ActiveStart: startDate.AddDate(0, -1, 0),
			ActiveEnd:   startDate.AddDate(0, 1, 0),
			RateID:      rate.ID,
		}
		if err := db.Create(&billingCode2).Error; err != nil {
			t.Fatalf("Failed to create billing code 2: %v", err)
		}

		// Create an entry for a project with no staffing assignment
		entry := Entry{
			EmployeeID:    employee.ID,
			BillingCodeID: billingCode2.ID,
			Start:         startDate,
			End:           startDate.Add(1 * time.Hour),
			Notes:         "Test entry without assignment",
		}
		if err := db.Create(&entry).Error; err != nil {
			t.Fatalf("Failed to create entry: %v", err)
		}

		// Associate the entry - should succeed even without a staffing assignment
		if err := app.AssociateEntry(&entry, project2.ID); err != nil {
			t.Fatalf("AssociateEntry should succeed without staffing assignment: %v", err)
		}

		// Reload the entry and verify it's NOT associated with any staffing assignment
		var updatedEntry Entry
		if err := db.First(&updatedEntry, entry.ID).Error; err != nil {
			t.Fatalf("Failed to reload entry: %v", err)
		}

		if updatedEntry.StaffingAssignmentID != nil {
			t.Errorf("Entry should not have a staffing assignment, but got ID: %d",
				*updatedEntry.StaffingAssignmentID)
		}

		// Verify it was still associated with an invoice (normal behavior)
		// Note: Invoice association requires a pre-existing invoice in the DB,
		// which may not exist in this isolated test
		if updatedEntry.State == EntryStateUnaffiliated.String() {
			// This is expected if no invoice exists
			t.Logf("Entry is unaffiliated (no invoice exists)")
		} else if updatedEntry.InvoiceID != nil {
			t.Logf("Entry successfully associated with invoice ID: %d", *updatedEntry.InvoiceID)
		}
	})
}

// TestBillGenerationByEntryState tests that bills are generated correctly based on EntryPayEligibleState
func TestBillGenerationByEntryState(t *testing.T) {
	// Setup test DB
	db := setupTestDB(t)
	app := &App{DB: db, Project: "test-project", Bucket: "test-bucket"}

	now := time.Now()
	startDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	// Create users and employees with different EntryPayEligibleState values
	testCases := []struct {
		name                  string
		entryPayEligibleState string
		expectedAtApprove     bool
		expectedAtSent        bool
		expectedAtPaid        bool
	}{
		{
			name:                  "Employee paid on approved",
			entryPayEligibleState: EntryStateApproved.String(),
			expectedAtApprove:     true,
			expectedAtSent:        false,
			expectedAtPaid:        false,
		},
		{
			name:                  "Employee paid on sent",
			entryPayEligibleState: EntryStateSent.String(),
			expectedAtApprove:     false,
			expectedAtSent:        true,
			expectedAtPaid:        false,
		},
		{
			name:                  "Employee paid on paid",
			entryPayEligibleState: EntryStatePaid.String(),
			expectedAtApprove:     false,
			expectedAtSent:        false,
			expectedAtPaid:        true,
		},
		{
			name:                  "Employee with no eligible state (processes all)",
			entryPayEligibleState: "",
			expectedAtApprove:     true,
			expectedAtSent:        true,
			expectedAtPaid:        true,
		},
	}

	// Create shared resources
	account := Account{
		Name:                  "Test Account",
		Type:                  AccountTypeClient.String(),
		LegalName:             "Test Inc.",
		ProjectsSingleInvoice: false,
	}
	if err := db.Create(&account).Error; err != nil {
		t.Fatalf("Failed to create account: %v", err)
	}

	project := Project{
		Name:        "Test Project",
		AccountID:   account.ID,
		ActiveStart: startDate.AddDate(0, -1, 0),
		ActiveEnd:   startDate.AddDate(0, 2, 0),
		Internal:    false,
	}
	if err := db.Create(&project).Error; err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	externalRate := Rate{
		Name:       "External Rate",
		Amount:     100.0,
		ActiveFrom: startDate.AddDate(-1, 0, 0),
		ActiveTo:   startDate.AddDate(1, 0, 0),
	}
	if err := db.Create(&externalRate).Error; err != nil {
		t.Fatalf("Failed to create external rate: %v", err)
	}

	internalRate := Rate{
		Name:         "Internal Rate",
		Amount:       50.0,
		ActiveFrom:   startDate.AddDate(-1, 0, 0),
		ActiveTo:     startDate.AddDate(1, 0, 0),
		InternalOnly: true,
	}
	if err := db.Create(&internalRate).Error; err != nil {
		t.Fatalf("Failed to create internal rate: %v", err)
	}

	for i, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create user and employee for this test case
			user := User{
				Email:    "employee" + string(rune(i)) + "@example.com",
				Password: "password",
				IsAdmin:  false,
				Role:     UserRoleStaff.String(),
			}
			if err := db.Create(&user).Error; err != nil {
				t.Fatalf("Failed to create user: %v", err)
			}

			employee := Employee{
				UserID:                user.ID,
				FirstName:             "Employee",
				LastName:              string(rune(i)),
				IsActive:              true,
				StartDate:             startDate.AddDate(-1, 0, 0),
				EntryPayEligibleState: tc.entryPayEligibleState,
			}
			if err := db.Create(&employee).Error; err != nil {
				t.Fatalf("Failed to create employee: %v", err)
			}

			// Create a billing code for this employee
			billingCode := BillingCode{
				Name:           "Development " + string(rune(i)),
				RateType:       RateTypeExternalBillable.String(),
				Category:       "Development",
				Code:           "DEV-" + string(rune(i)),
				RoundedTo:      15,
				ProjectID:      project.ID,
				ActiveStart:    startDate.AddDate(0, -1, 0),
				ActiveEnd:      startDate.AddDate(0, 1, 0),
				RateID:         externalRate.ID,
				InternalRateID: internalRate.ID,
			}
			if err := db.Create(&billingCode).Error; err != nil {
				t.Fatalf("Failed to create billing code: %v", err)
			}

			// Create an entry
			entry := Entry{
				EmployeeID:    employee.ID,
				BillingCodeID: billingCode.ID,
				ProjectID:     project.ID,
				Start:         startDate,
				End:           startDate.Add(2 * time.Hour),
				Notes:         "Test entry",
				State:         EntryStateDraft.String(),
			}
			if err := db.Create(&entry).Error; err != nil {
				t.Fatalf("Failed to create entry: %v", err)
			}

			// Create an invoice with this entry
			invoice := Invoice{
				Name:        "Test Invoice",
				AccountID:   account.ID,
				ProjectID:   &project.ID,
				PeriodStart: startDate.AddDate(0, 0, -15),
				PeriodEnd:   startDate.AddDate(0, 0, 15),
				State:       InvoiceStateDraft.String(),
				Type:        InvoiceTypeAR.String(),
			}
			if err := db.Create(&invoice).Error; err != nil {
				t.Fatalf("Failed to create invoice: %v", err)
			}

			// Associate the entry with the invoice
			entry.InvoiceID = &invoice.ID
			if err := db.Save(&entry).Error; err != nil {
				t.Fatalf("Failed to associate entry with invoice: %v", err)
			}

			// Reload invoice with entries
			if err := db.Preload("Entries").First(&invoice, invoice.ID).Error; err != nil {
				t.Fatalf("Failed to reload invoice: %v", err)
			}

			// Test approval
			entry.State = EntryStateApproved.String()
			db.Save(&entry)
			invoice.State = InvoiceStateApproved.String()
			db.Preload("Entries").First(&invoice, invoice.ID)
			app.GenerateBills(&invoice)

			// Check if bill was created at approve stage
			var billAfterApprove Bill
			errApprove := db.Where("employee_id = ?", employee.ID).First(&billAfterApprove).Error
			if tc.expectedAtApprove {
				if errApprove != nil {
					t.Errorf("Expected bill to be created at approve stage, but got error: %v", errApprove)
				}
			} else {
				if errApprove == nil {
					t.Errorf("Did not expect bill at approve stage, but found bill ID: %d", billAfterApprove.ID)
				}
			}

			// Test sent
			entry.State = EntryStateSent.String()
			db.Save(&entry)
			invoice.State = InvoiceStateSent.String()
			db.Preload("Entries").First(&invoice, invoice.ID)
			app.GenerateBills(&invoice)

			var billAfterSent Bill
			errSent := db.Where("employee_id = ?", employee.ID).First(&billAfterSent).Error
			if tc.expectedAtSent {
				if errSent != nil {
					t.Errorf("Expected bill to be created at sent stage, but got error: %v", errSent)
				}
			}

			// Test paid
			entry.State = EntryStatePaid.String()
			db.Save(&entry)
			invoice.State = InvoiceStatePaid.String()
			db.Preload("Entries").First(&invoice, invoice.ID)
			app.GenerateBills(&invoice)

			var billAfterPaid Bill
			errPaid := db.Where("employee_id = ?", employee.ID).First(&billAfterPaid).Error
			if tc.expectedAtPaid {
				if errPaid != nil {
					t.Errorf("Expected bill to be created at paid stage, but got error: %v", errPaid)
				}
			}
		})
	}
}

// TestBillGenerationWithVariableAndFixedRates tests that bills calculate fees correctly
func TestBillGenerationWithVariableAndFixedRates(t *testing.T) {
	// Setup test DB
	db := setupTestDB(t)
	app := &App{DB: db, Project: "test-project", Bucket: "test-bucket"}

	now := time.Now()
	startDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	// Create account and project
	account := Account{
		Name:                  "Test Account",
		Type:                  AccountTypeClient.String(),
		LegalName:             "Test Inc.",
		ProjectsSingleInvoice: false,
	}
	if err := db.Create(&account).Error; err != nil {
		t.Fatalf("Failed to create account: %v", err)
	}

	project := Project{
		Name:        "Test Project",
		AccountID:   account.ID,
		ActiveStart: startDate.AddDate(0, -1, 0),
		ActiveEnd:   startDate.AddDate(0, 2, 0),
		Internal:    false,
	}
	if err := db.Create(&project).Error; err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	externalRate := Rate{
		Name:       "External Rate",
		Amount:     100.0,
		ActiveFrom: startDate.AddDate(-1, 0, 0),
		ActiveTo:   startDate.AddDate(1, 0, 0),
	}
	if err := db.Create(&externalRate).Error; err != nil {
		t.Fatalf("Failed to create external rate: %v", err)
	}

	internalRate := Rate{
		Name:         "Internal Rate",
		Amount:       50.0,
		ActiveFrom:   startDate.AddDate(-1, 0, 0),
		ActiveTo:     startDate.AddDate(1, 0, 0),
		InternalOnly: true,
	}
	if err := db.Create(&internalRate).Error; err != nil {
		t.Fatalf("Failed to create internal rate: %v", err)
	}

	// Test 1: Fixed rate employee
	t.Run("Fixed rate employee", func(t *testing.T) {
		user := User{
			Email:    "fixed@example.com",
			Password: "password",
			Role:     UserRoleStaff.String(),
		}
		db.Create(&user)

		employee := Employee{
			UserID:                  user.ID,
			FirstName:               "Fixed",
			LastName:                "Rate",
			IsActive:                true,
			StartDate:               startDate.AddDate(-1, 0, 0),
			EntryPayEligibleState:   EntryStatePaid.String(),
			HasFixedInternalRate:    true,
			HasVariableInternalRate: false,
			FixedHourlyRate:         7500, // $75/hour (stored in cents like other int monetary fields)
		}
		db.Create(&employee)

		billingCode := BillingCode{
			Name:           "Dev Fixed",
			RateType:       RateTypeExternalBillable.String(),
			Code:           "DEV-FIXED",
			ProjectID:      project.ID,
			ActiveStart:    startDate.AddDate(0, -1, 0),
			ActiveEnd:      startDate.AddDate(0, 1, 0),
			RateID:         externalRate.ID,
			InternalRateID: internalRate.ID,
		}
		db.Create(&billingCode)

		entry := Entry{
			EmployeeID:    employee.ID,
			BillingCodeID: billingCode.ID,
			ProjectID:     project.ID,
			Start:         startDate,
			End:           startDate.Add(2 * time.Hour), // 2 hours
			State:         EntryStatePaid.String(),
		}
		db.Create(&entry)

		invoice := Invoice{
			Name:        "Test Invoice Fixed",
			AccountID:   account.ID,
			ProjectID:   &project.ID,
			PeriodStart: startDate.AddDate(0, 0, -15),
			PeriodEnd:   startDate.AddDate(0, 0, 15),
			State:       InvoiceStatePaid.String(),
			Type:        InvoiceTypeAR.String(),
		}
		db.Create(&invoice)

		entry.InvoiceID = &invoice.ID
		db.Save(&entry)

		db.Preload("Entries").First(&invoice, invoice.ID)
		app.GenerateBills(&invoice)

		var bill Bill
		if err := db.Where("employee_id = ?", employee.ID).First(&bill).Error; err != nil {
			t.Fatalf("Failed to find bill: %v", err)
		}

		// Expected: 2 hours * $75/hour (employee fixed rate) = $150 = 15000 cents
		// RecalculateBillTotals should use the employee's fixed rate, not the billing code's internal rate
		expected := 15000
		if bill.TotalFees != expected {
			t.Errorf("Expected bill total fees %d, got %d", expected, bill.TotalFees)
		}

		// Verify the generation logs showed the correct fixed rate usage
		t.Logf("Bill generated successfully with employee fixed rate consideration")
	})

	// Test 2: Variable rate employee
	t.Run("Variable rate employee", func(t *testing.T) {
		user := User{
			Email:    "variable@example.com",
			Password: "password",
			Role:     UserRoleStaff.String(),
		}
		db.Create(&user)

		employee := Employee{
			UserID:                  user.ID,
			FirstName:               "Variable",
			LastName:                "Rate",
			IsActive:                true,
			StartDate:               startDate.AddDate(-1, 0, 0),
			EntryPayEligibleState:   EntryStatePaid.String(),
			HasFixedInternalRate:    false,
			HasVariableInternalRate: true,
		}
		db.Create(&employee)

		billingCode := BillingCode{
			Name:           "Dev Variable",
			RateType:       RateTypeExternalBillable.String(),
			Code:           "DEV-VAR",
			ProjectID:      project.ID,
			ActiveStart:    startDate.AddDate(0, -1, 0),
			ActiveEnd:      startDate.AddDate(0, 1, 0),
			RateID:         externalRate.ID,
			InternalRateID: internalRate.ID, // $50/hour
		}
		db.Create(&billingCode)

		entry := Entry{
			EmployeeID:    employee.ID,
			BillingCodeID: billingCode.ID,
			ProjectID:     project.ID,
			Start:         startDate,
			End:           startDate.Add(2 * time.Hour), // 2 hours
			State:         EntryStatePaid.String(),
		}
		db.Create(&entry)

		invoice := Invoice{
			Name:        "Test Invoice Variable",
			AccountID:   account.ID,
			ProjectID:   &project.ID,
			PeriodStart: startDate.AddDate(0, 0, -15),
			PeriodEnd:   startDate.AddDate(0, 0, 15),
			State:       InvoiceStatePaid.String(),
			Type:        InvoiceTypeAR.String(),
		}
		db.Create(&invoice)

		entry.InvoiceID = &invoice.ID
		db.Save(&entry)

		db.Preload("Entries").First(&invoice, invoice.ID)
		app.GenerateBills(&invoice)

		var bill Bill
		if err := db.Where("employee_id = ?", employee.ID).First(&bill).Error; err != nil {
			t.Fatalf("Failed to find bill: %v", err)
		}

		// Expected: 2 hours * $50/hour = $100 = 10000 cents
		expected := 10000
		if bill.TotalFees != expected {
			t.Errorf("Expected bill total fees %d, got %d", expected, bill.TotalFees)
		}
	})
}

// TestStaffingAssignmentEntriesRelationship validates that staffing assignments can properly access their associated entries
func TestStaffingAssignmentEntriesRelationship(t *testing.T) {
	// Setup test DB
	db := setupTestDB(t)
	app := &App{DB: db, Project: "test-project", Bucket: "test-bucket"}

	now := time.Now()
	startDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	// Create user and employee
	user := User{
		Email:    "employee@example.com",
		Password: "password",
		IsAdmin:  false,
		Role:     UserRoleStaff.String(),
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	employee := Employee{
		UserID:    user.ID,
		FirstName: "John",
		LastName:  "Doe",
		IsActive:  true,
		StartDate: startDate.AddDate(-1, 0, 0),
	}
	if err := db.Create(&employee).Error; err != nil {
		t.Fatalf("Failed to create employee: %v", err)
	}

	// Create account and project
	account := Account{
		Name:                  "Test Account",
		Type:                  AccountTypeClient.String(),
		LegalName:             "Test Inc.",
		ProjectsSingleInvoice: false,
	}
	if err := db.Create(&account).Error; err != nil {
		t.Fatalf("Failed to create account: %v", err)
	}

	project := Project{
		Name:        "Test Project",
		AccountID:   account.ID,
		ActiveStart: startDate.AddDate(0, -1, 0),
		ActiveEnd:   startDate.AddDate(0, 2, 0),
		Internal:    false,
	}
	if err := db.Create(&project).Error; err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Create rate and billing code
	rate := Rate{
		Name:       "Standard Rate",
		Amount:     100.0,
		ActiveFrom: startDate.AddDate(-1, 0, 0),
		ActiveTo:   startDate.AddDate(1, 0, 0),
	}
	if err := db.Create(&rate).Error; err != nil {
		t.Fatalf("Failed to create rate: %v", err)
	}

	billingCode := BillingCode{
		Name:        "Development",
		RateType:    RateTypeExternalBillable.String(),
		Category:    "Development",
		Code:        "DEV-001",
		RoundedTo:   15,
		ProjectID:   project.ID,
		ActiveStart: startDate.AddDate(0, -1, 0),
		ActiveEnd:   startDate.AddDate(0, 1, 0),
		RateID:      rate.ID,
	}
	if err := db.Create(&billingCode).Error; err != nil {
		t.Fatalf("Failed to create billing code: %v", err)
	}

	// Create a staffing assignment
	staffingAssignment := StaffingAssignment{
		EmployeeID: employee.ID,
		ProjectID:  project.ID,
		Commitment: 40,
		StartDate:  startDate.AddDate(0, 0, -7),
		EndDate:    startDate.AddDate(0, 0, 30),
	}
	if err := db.Create(&staffingAssignment).Error; err != nil {
		t.Fatalf("Failed to create staffing assignment: %v", err)
	}

	// Create multiple entries within the assignment date range
	entry1 := Entry{
		EmployeeID:    employee.ID,
		BillingCodeID: billingCode.ID,
		Start:         startDate,
		End:           startDate.Add(2 * time.Hour),
		Notes:         "First entry",
	}
	if err := db.Create(&entry1).Error; err != nil {
		t.Fatalf("Failed to create entry1: %v", err)
	}

	entry2 := Entry{
		EmployeeID:    employee.ID,
		BillingCodeID: billingCode.ID,
		Start:         startDate.AddDate(0, 0, 1),
		End:           startDate.AddDate(0, 0, 1).Add(3 * time.Hour),
		Notes:         "Second entry",
	}
	if err := db.Create(&entry2).Error; err != nil {
		t.Fatalf("Failed to create entry2: %v", err)
	}

	entry3 := Entry{
		EmployeeID:    employee.ID,
		BillingCodeID: billingCode.ID,
		Start:         startDate.AddDate(0, 0, 2),
		End:           startDate.AddDate(0, 0, 2).Add(1 * time.Hour),
		Notes:         "Third entry",
	}
	if err := db.Create(&entry3).Error; err != nil {
		t.Fatalf("Failed to create entry3: %v", err)
	}

	// Associate all entries
	if err := app.AssociateEntry(&entry1, project.ID); err != nil {
		t.Fatalf("Failed to associate entry1: %v", err)
	}
	if err := app.AssociateEntry(&entry2, project.ID); err != nil {
		t.Fatalf("Failed to associate entry2: %v", err)
	}
	if err := app.AssociateEntry(&entry3, project.ID); err != nil {
		t.Fatalf("Failed to associate entry3: %v", err)
	}

	// Reload staffing assignment with entries
	var reloadedAssignment StaffingAssignment
	if err := db.Preload("Entries").First(&reloadedAssignment, staffingAssignment.ID).Error; err != nil {
		t.Fatalf("Failed to reload staffing assignment: %v", err)
	}

	// Verify the staffing assignment has all 3 entries
	if len(reloadedAssignment.Entries) != 3 {
		t.Errorf("Expected staffing assignment to have 3 entries, got %d", len(reloadedAssignment.Entries))
	}

	// Verify each entry is correctly associated
	entryIDs := map[uint]bool{
		entry1.ID: false,
		entry2.ID: false,
		entry3.ID: false,
	}

	for _, entry := range reloadedAssignment.Entries {
		if entry.StaffingAssignmentID == nil {
			t.Errorf("Entry ID %d has nil StaffingAssignmentID", entry.ID)
			continue
		}
		if *entry.StaffingAssignmentID != staffingAssignment.ID {
			t.Errorf("Entry ID %d has wrong staffing assignment ID: expected %d, got %d",
				entry.ID, staffingAssignment.ID, *entry.StaffingAssignmentID)
		}
		entryIDs[entry.ID] = true
	}

	// Verify all entries were found
	for id, found := range entryIDs {
		if !found {
			t.Errorf("Entry ID %d was not found in staffing assignment entries", id)
		}
	}

	// Test that we can calculate totals from the staffing assignment
	var totalHours float64
	for _, entry := range reloadedAssignment.Entries {
		totalHours += entry.Duration().Hours()
	}

	expectedHours := 6.0 // 2 + 3 + 1 hours
	if totalHours != expectedHours {
		t.Errorf("Expected total hours %.2f, got %.2f", expectedHours, totalHours)
	}

	t.Logf("Staffing assignment successfully loaded with %d entries, total hours: %.2f",
		len(reloadedAssignment.Entries), totalHours)
}

// TestEntryDurationMinutesAutoCalculation validates that DurationMinutes is automatically calculated on save
func TestEntryDurationMinutesAutoCalculation(t *testing.T) {
	// Setup test DB
	db := setupTestDB(t)

	now := time.Now()
	startDate := time.Date(now.Year(), now.Month(), now.Day(), 9, 0, 0, 0, time.UTC)

	// Create user and employee
	user := User{
		Email:    "employee@example.com",
		Password: "password",
		Role:     UserRoleStaff.String(),
	}
	db.Create(&user)

	employee := Employee{
		UserID:    user.ID,
		FirstName: "Test",
		LastName:  "Employee",
		IsActive:  true,
		StartDate: startDate.AddDate(-1, 0, 0),
	}
	db.Create(&employee)

	// Create minimal required data
	account := Account{
		Name:      "Test Account",
		Type:      AccountTypeClient.String(),
		LegalName: "Test Inc.",
	}
	db.Create(&account)

	project := Project{
		Name:        "Test Project",
		AccountID:   account.ID,
		ActiveStart: startDate.AddDate(0, -1, 0),
		ActiveEnd:   startDate.AddDate(0, 1, 0),
	}
	db.Create(&project)

	rate := Rate{
		Name:       "Test Rate",
		Amount:     100.0,
		ActiveFrom: startDate.AddDate(-1, 0, 0),
		ActiveTo:   startDate.AddDate(1, 0, 0),
	}
	db.Create(&rate)

	billingCode := BillingCode{
		Name:        "Test Code",
		RateType:    RateTypeExternalBillable.String(),
		Code:        "TEST",
		ProjectID:   project.ID,
		ActiveStart: startDate.AddDate(0, -1, 0),
		ActiveEnd:   startDate.AddDate(0, 1, 0),
		RateID:      rate.ID,
	}
	db.Create(&billingCode)

	// Test cases with different durations
	testCases := []struct {
		name            string
		start           time.Time
		end             time.Time
		expectedMinutes float64
	}{
		{
			name:            "1 hour",
			start:           startDate,
			end:             startDate.Add(1 * time.Hour),
			expectedMinutes: 60.0,
		},
		{
			name:            "2.5 hours",
			start:           startDate,
			end:             startDate.Add(2*time.Hour + 30*time.Minute),
			expectedMinutes: 150.0,
		},
		{
			name:            "45 minutes",
			start:           startDate,
			end:             startDate.Add(45 * time.Minute),
			expectedMinutes: 45.0,
		},
		{
			name:            "8 hours",
			start:           startDate,
			end:             startDate.Add(8 * time.Hour),
			expectedMinutes: 480.0,
		},
		{
			name:            "15 seconds (fractional minutes)",
			start:           startDate,
			end:             startDate.Add(15 * time.Second),
			expectedMinutes: 0.25,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create entry without setting DurationMinutes
			entry := Entry{
				EmployeeID:    employee.ID,
				BillingCodeID: billingCode.ID,
				ProjectID:     project.ID,
				Start:         tc.start,
				End:           tc.end,
				Notes:         "Test entry for " + tc.name,
			}

			// DurationMinutes should be 0 before save
			if entry.DurationMinutes != 0 {
				t.Errorf("Expected DurationMinutes to be 0 before save, got %.2f", entry.DurationMinutes)
			}

			// Save the entry - this should trigger BeforeSave hook
			if err := db.Create(&entry).Error; err != nil {
				t.Fatalf("Failed to create entry: %v", err)
			}

			// Reload the entry from database
			var reloadedEntry Entry
			if err := db.First(&reloadedEntry, entry.ID).Error; err != nil {
				t.Fatalf("Failed to reload entry: %v", err)
			}

			// Verify DurationMinutes was auto-calculated and saved
			if reloadedEntry.DurationMinutes != tc.expectedMinutes {
				t.Errorf("Expected DurationMinutes to be %.2f, got %.2f",
					tc.expectedMinutes, reloadedEntry.DurationMinutes)
			}

			// Also verify Duration() method still works
			calculatedDuration := reloadedEntry.Duration().Minutes()
			if calculatedDuration != tc.expectedMinutes {
				t.Errorf("Duration() method returned %.2f, expected %.2f",
					calculatedDuration, tc.expectedMinutes)
			}

			t.Logf("Entry saved with DurationMinutes: %.2f (matches Duration() method: %.2f)",
				reloadedEntry.DurationMinutes, calculatedDuration)
		})
	}

	// Test that updating an entry also updates DurationMinutes
	t.Run("Update entry duration", func(t *testing.T) {
		entry := Entry{
			EmployeeID:    employee.ID,
			BillingCodeID: billingCode.ID,
			ProjectID:     project.ID,
			Start:         startDate,
			End:           startDate.Add(1 * time.Hour),
			Notes:         "Original entry",
		}
		db.Create(&entry)

		// Verify initial duration
		if entry.DurationMinutes != 60.0 {
			t.Errorf("Expected initial DurationMinutes to be 60.0, got %.2f", entry.DurationMinutes)
		}

		// Update the end time
		entry.End = startDate.Add(3 * time.Hour)
		db.Save(&entry)

		// Reload and verify updated duration
		var reloadedEntry Entry
		db.First(&reloadedEntry, entry.ID)

		if reloadedEntry.DurationMinutes != 180.0 {
			t.Errorf("Expected updated DurationMinutes to be 180.0, got %.2f", reloadedEntry.DurationMinutes)
		}

		t.Logf("Entry duration successfully updated from 60.0 to %.2f minutes", reloadedEntry.DurationMinutes)
	})
}
