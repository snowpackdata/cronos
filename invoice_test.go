package cronos

import (
	"testing"
	"time"
)

// TestUpdateInvoiceTotals tests the UpdateInvoiceTotals function
func TestUpdateInvoiceTotals(t *testing.T) {
	// Setup test DB
	db := setupTestDB(t)
	app := &App{DB: db}

	// Create account
	account := Account{
		Name:      "Invoice Test Account",
		LegalName: "Invoice Test Legal Name",
		Type:      AccountTypeClient.String(),
	}
	if err := db.Create(&account).Error; err != nil {
		t.Fatalf("Failed to create account: %v", err)
	}

	// Create project
	project := Project{
		Name:        "Invoice Test Project",
		AccountID:   account.ID,
		ActiveStart: time.Now().AddDate(-1, 0, 0),
		ActiveEnd:   time.Now().AddDate(1, 0, 0),
	}
	if err := db.Create(&project).Error; err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Create user & employee
	user := User{
		Email:    "invoice-test@example.com",
		Password: "password123",
		Role:     UserRoleStaff.String(),
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	employee := Employee{
		UserID:    user.ID,
		FirstName: "Invoice",
		LastName:  "Tester",
		IsActive:  true,
		StartDate: time.Now().AddDate(-1, 0, 0),
	}
	if err := db.Create(&employee).Error; err != nil {
		t.Fatalf("Failed to create employee: %v", err)
	}

	// Create rate
	rate := Rate{
		Name:         "Invoice Test Rate",
		Amount:       150.0,
		ActiveFrom:   time.Now().AddDate(-1, 0, 0),
		ActiveTo:     time.Now().AddDate(1, 0, 0),
		InternalOnly: false,
	}
	if err := db.Create(&rate).Error; err != nil {
		t.Fatalf("Failed to create rate: %v", err)
	}

	// Create billing code
	billingCode := BillingCode{
		Name:        "Invoice Test Billing Code",
		RateType:    RateTypeExternalBillable.String(),
		Category:    "Development",
		Code:        "INV-001",
		RoundedTo:   15,
		ProjectID:   project.ID,
		ActiveStart: time.Now().AddDate(-1, 0, 0),
		ActiveEnd:   time.Now().AddDate(1, 0, 0),
		RateID:      rate.ID,
	}
	if err := db.Create(&billingCode).Error; err != nil {
		t.Fatalf("Failed to create billing code: %v", err)
	}

	// Create an invoice
	periodStart := time.Now().AddDate(0, -1, 0) // Last month
	periodEnd := time.Now()

	invoice := Invoice{
		Name:        "INV-TEST-001",
		AccountID:   account.ID,
		ProjectID:   &project.ID,
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
		State:       InvoiceStateDraft.String(),
		Type:        InvoiceTypeAR.String(),
	}
	if err := db.Create(&invoice).Error; err != nil {
		t.Fatalf("Failed to create invoice: %v", err)
	}

	// Create entries associated with the invoice
	entries := []Entry{
		{
			ProjectID:     project.ID,
			Notes:         "Invoice test entry 1",
			EmployeeID:    employee.ID,
			BillingCodeID: billingCode.ID,
			Start:         periodStart.Add(24 * time.Hour),
			End:           periodStart.Add(24*time.Hour + 2*time.Hour), // 2 hours
			Internal:      false,
			InvoiceID:     &invoice.ID,
			State:         EntryStateApproved.String(),
		},
		{
			ProjectID:     project.ID,
			Notes:         "Invoice test entry 2",
			EmployeeID:    employee.ID,
			BillingCodeID: billingCode.ID,
			Start:         periodStart.Add(48 * time.Hour),
			End:           periodStart.Add(48*time.Hour + 3*time.Hour), // 3 hours
			Internal:      false,
			InvoiceID:     &invoice.ID,
			State:         EntryStateApproved.String(),
		},
	}

	for i := range entries {
		if err := db.Create(&entries[i]).Error; err != nil {
			t.Fatalf("Failed to create entry %d: %v", i, err)
		}
	}

	// Create adjustments - now totaling $250 (a fee of $100, a credit of $-50, a fee of $100, and an additional fee of $100)
	adjustments := []Adjustment{
		{
			InvoiceID: &invoice.ID,
			Type:      AdjustmentTypeFee.String(),
			State:     AdjustmentStateApproved.String(),
			Amount:    100.0, // $100 additional fee
			Notes:     "Additional services fee",
		},
		{
			InvoiceID: &invoice.ID,
			Type:      AdjustmentTypeCredit.String(),
			State:     AdjustmentStateApproved.String(),
			Amount:    -50.0, // $50 discount
			Notes:     "Loyalty discount",
		},
		{
			InvoiceID: &invoice.ID,
			Type:      AdjustmentTypeFee.String(),
			State:     AdjustmentStateApproved.String(),
			Amount:    100.0, // Another $100 fee
			Notes:     "Extra fee",
		},
		{
			InvoiceID: &invoice.ID,
			Type:      AdjustmentTypeFee.String(),
			State:     AdjustmentStateApproved.String(),
			Amount:    100.0, // One more $100 fee
			Notes:     "Service fee",
		},
	}

	for i := range adjustments {
		if err := db.Create(&adjustments[i]).Error; err != nil {
			t.Fatalf("Failed to create adjustment %d: %v", i, err)
		}
	}

	// Update the invoice totals
	app.UpdateInvoiceTotals(&invoice)

	// Verify totals
	// Hours: 2 + 3 = 5 hours
	// Fees: 5 hours * $150/hour = $750
	// Adjustments: $100 - $50 + $100 + $100 = $250
	// Total: $750 + $250 = $1000

	if invoice.TotalHours != 5.0 {
		t.Errorf("Expected total hours to be 5.0, got %.2f", invoice.TotalHours)
	}

	if invoice.TotalFees != 750.0 {
		t.Errorf("Expected total fees to be 750.0, got %.2f", invoice.TotalFees)
	}

	if invoice.TotalAdjustments != 250.0 {
		t.Errorf("Expected total adjustments to be 250.0, got %.2f", invoice.TotalAdjustments)
	}

	if invoice.TotalAmount != 1000.0 {
		t.Errorf("Expected total amount to be 1000.0, got %.2f", invoice.TotalAmount)
	}
}

// TestGetInvoiceLineItems tests the GetInvoiceLineItems function
func TestGetInvoiceLineItems(t *testing.T) {
	// Setup test DB
	db := setupTestDB(t)
	app := &App{DB: db}

	// Create account
	account := Account{
		Name:      "Line Item Test Account",
		LegalName: "Line Item Test Legal Name",
		Type:      AccountTypeClient.String(),
	}
	if err := db.Create(&account).Error; err != nil {
		t.Fatalf("Failed to create account: %v", err)
	}

	// Create project
	project := Project{
		Name:        "Line Item Test Project",
		AccountID:   account.ID,
		ActiveStart: time.Now().AddDate(-1, 0, 0),
		ActiveEnd:   time.Now().AddDate(1, 0, 0),
	}
	if err := db.Create(&project).Error; err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Create user & employee
	user := User{
		Email:    "lineitem-test@example.com",
		Password: "password123",
		Role:     UserRoleStaff.String(),
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	employee := Employee{
		UserID:    user.ID,
		FirstName: "LineItem",
		LastName:  "Tester",
		IsActive:  true,
		StartDate: time.Now().AddDate(-1, 0, 0),
	}
	if err := db.Create(&employee).Error; err != nil {
		t.Fatalf("Failed to create employee: %v", err)
	}

	// Create two different rates
	rate1 := Rate{
		Name:         "Rate 1",
		Amount:       100.0,
		ActiveFrom:   time.Now().AddDate(-1, 0, 0),
		ActiveTo:     time.Now().AddDate(1, 0, 0),
		InternalOnly: false,
	}
	if err := db.Create(&rate1).Error; err != nil {
		t.Fatalf("Failed to create rate1: %v", err)
	}

	rate2 := Rate{
		Name:         "Rate 2",
		Amount:       150.0,
		ActiveFrom:   time.Now().AddDate(-1, 0, 0),
		ActiveTo:     time.Now().AddDate(1, 0, 0),
		InternalOnly: false,
	}
	if err := db.Create(&rate2).Error; err != nil {
		t.Fatalf("Failed to create rate2: %v", err)
	}

	// Create two billing codes
	billingCode1 := BillingCode{
		Name:        "Development",
		RateType:    RateTypeExternalBillable.String(),
		Category:    "Development",
		Code:        "DEV-001",
		RoundedTo:   15,
		ProjectID:   project.ID,
		ActiveStart: time.Now().AddDate(-1, 0, 0),
		ActiveEnd:   time.Now().AddDate(1, 0, 0),
		RateID:      rate1.ID,
	}
	if err := db.Create(&billingCode1).Error; err != nil {
		t.Fatalf("Failed to create billing code1: %v", err)
	}

	billingCode2 := BillingCode{
		Name:        "Design",
		RateType:    RateTypeExternalBillable.String(),
		Category:    "Design",
		Code:        "DES-001",
		RoundedTo:   15,
		ProjectID:   project.ID,
		ActiveStart: time.Now().AddDate(-1, 0, 0),
		ActiveEnd:   time.Now().AddDate(1, 0, 0),
		RateID:      rate2.ID,
	}
	if err := db.Create(&billingCode2).Error; err != nil {
		t.Fatalf("Failed to create billing code2: %v", err)
	}

	// Create an invoice
	periodStart := time.Now().AddDate(0, -1, 0) // Last month
	periodEnd := time.Now()

	invoice := Invoice{
		Name:        "INV-LINEITEM-001",
		AccountID:   account.ID,
		ProjectID:   &project.ID,
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
		State:       InvoiceStateDraft.String(),
		Type:        InvoiceTypeAR.String(),
	}
	if err := db.Create(&invoice).Error; err != nil {
		t.Fatalf("Failed to create invoice: %v", err)
	}

	// Create entries with different billing codes for the same project
	entries := []Entry{
		{
			ProjectID:     project.ID,
			Notes:         "Development work",
			EmployeeID:    employee.ID,
			BillingCodeID: billingCode1.ID,
			Start:         periodStart.Add(24 * time.Hour),
			End:           periodStart.Add(24*time.Hour + 2*time.Hour), // 2 hours
			Internal:      false,
			InvoiceID:     &invoice.ID,
			State:         EntryStateApproved.String(),
		},
		{
			ProjectID:     project.ID,
			Notes:         "More development",
			EmployeeID:    employee.ID,
			BillingCodeID: billingCode1.ID,
			Start:         periodStart.Add(48 * time.Hour),
			End:           periodStart.Add(48*time.Hour + 1*time.Hour), // 1 hour
			Internal:      false,
			InvoiceID:     &invoice.ID,
			State:         EntryStateApproved.String(),
		},
		{
			ProjectID:     project.ID,
			Notes:         "Design work",
			EmployeeID:    employee.ID,
			BillingCodeID: billingCode2.ID,
			Start:         periodStart.Add(72 * time.Hour),
			End:           periodStart.Add(72*time.Hour + 3*time.Hour), // 3 hours
			Internal:      false,
			InvoiceID:     &invoice.ID,
			State:         EntryStateApproved.String(),
		},
	}

	for i := range entries {
		if err := db.Create(&entries[i]).Error; err != nil {
			t.Fatalf("Failed to create entry %d: %v", i, err)
		}
	}

	// Get line items
	lineItems := app.GetInvoiceLineItems(&invoice)

	// Verify that we get two line items (one for each billing code)
	if len(lineItems) != 2 {
		t.Fatalf("Expected 2 line items, got %d", len(lineItems))
	}

	// Check the contents of the line items
	for _, item := range lineItems {
		if item.BillingCode == billingCode1.Code {
			// Development line item: 3 hours total at $100/hour
			if item.Hours != 3.0 {
				t.Errorf("Expected 3.0 hours for development, got %.2f", item.Hours)
			}
			if item.Rate != 100.0 {
				t.Errorf("Expected rate of 100.0 for development, got %.2f", item.Rate)
			}
			if item.Total != 300.0 {
				t.Errorf("Expected total of 300.0 for development, got %.2f", item.Total)
			}
		} else if item.BillingCode == billingCode2.Code {
			// Design line item: 3 hours at $150/hour
			if item.Hours != 3.0 {
				t.Errorf("Expected 3.0 hours for design, got %.2f", item.Hours)
			}
			if item.Rate != 150.0 {
				t.Errorf("Expected rate of 150.0 for design, got %.2f", item.Rate)
			}
			if item.Total != 450.0 {
				t.Errorf("Expected total of 450.0 for design, got %.2f", item.Total)
			}
		} else {
			t.Errorf("Unexpected billing code: %s", item.BillingCode)
		}
	}
}
