package cronos

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"cloud.google.com/go/storage"
)

var ErrInvoiceOverlap = errors.New("new invoice overlaps with existing invoice")
var InvalidPriorState = errors.New("invalid prior state")
var ErrEntryDateOutOfRange = errors.New("entry date is out of range for project")

// CreateInvoice generates the draft invoice to be manually approved and sent
// Draft invoices are generated on the 1st of the month for the current month. Entries
// created during that month will be associated with the draft invoice. At the end of the
// month, the invoice will transition to the "pending" state until it's manually approved and
// sent to the client. This function serves the purpose of generating the draft invoice, and transitioning
// and previous draft invoices to the "pending" state.
func (a *App) CreateInvoice(accountID uint, projectID *uint, creationDate time.Time) error {
	// We need a timestamp to determine the start and end of the month
	startOfMonth := time.Date(creationDate.Year(), creationDate.Month(), 1, 0, 0, 0, 0, time.UTC)
	endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-time.Nanosecond)

	// Retrieve the account from the database
	var account Account
	a.DB.Where("ID = ?", accountID).First(&account)

	// next retrieve any past invoices for this account
	var invoices []Invoice
	if projectID != nil && !account.ProjectsSingleInvoice {
		// If separate invoices per project is enabled, query by both account and project
		a.DB.Order("period_end desc").Where("account_id = ? AND project_id = ? and state != ?", accountID, *projectID, InvoiceStateVoid).Find(&invoices)
	} else {
		// Otherwise query only by account
		a.DB.Order("period_end desc").Where("account_id = ? and state != ?", accountID, InvoiceStateVoid).Find(&invoices)
	}

	// Create new invoice
	var newARInvoice Invoice
	newARInvoice = Invoice{
		AccountID: accountID,
		Account:   account,
		State:     InvoiceStateDraft.String(),
		Type:      InvoiceTypeAR.String(),
	}

	// If we're creating a project-specific invoice
	if projectID != nil && !account.ProjectsSingleInvoice {
		var project Project
		a.DB.Where("ID = ?", *projectID).First(&project)
		newARInvoice.ProjectID = projectID
		newARInvoice.Project = project
	}

	// Set billing period based on account's billing frequency
	switch account.BillingFrequency {
	case BillingFrequencyProject.String():
		// For project-based billing, set the period to the project's start and end dates
		if projectID != nil {
			var project Project
			a.DB.Where("ID = ?", *projectID).First(&project)
			newARInvoice.PeriodStart = project.ActiveStart
			newARInvoice.PeriodEnd = project.ActiveEnd
			newARInvoice.Name = account.Name + " - " + project.Name + ": " + project.ActiveStart.Format("01.02.2006") + "-" + project.ActiveEnd.Format("01.02.2006")
		} else {
			// If no project specified for project-based billing on single invoice, use current month
			newARInvoice.PeriodStart = startOfMonth
			newARInvoice.PeriodEnd = endOfMonth
			newARInvoice.Name = account.Name + ": " + startOfMonth.Format("01.02.2006") + "-" + endOfMonth.Format("01.02.2006")
		}
	case BillingFrequencyMonthly.String():
		newARInvoice.PeriodStart = startOfMonth
		newARInvoice.PeriodEnd = endOfMonth
		if projectID != nil && !account.ProjectsSingleInvoice {
			var project Project
			a.DB.Where("ID = ?", *projectID).First(&project)
			newARInvoice.Name = account.Name + " - " + project.Name + ": " + startOfMonth.Format("01.02.2006") + "-" + endOfMonth.Format("01.02.2006")
		} else {
			newARInvoice.Name = account.Name + ": " + startOfMonth.Format("01.02.2006") + "-" + endOfMonth.Format("01.02.2006")
		}
	case BillingFrequencyBiweekly.String():
		// retrieve the ending date of the last invoice
		if len(invoices) > 0 {
			preCleanedStart := invoices[0].PeriodEnd.AddDate(0, 0, 1)
			newARInvoice.PeriodStart = time.Date(preCleanedStart.Year(), preCleanedStart.Month(), preCleanedStart.Day(), 0, 0, 0, 0, time.UTC)
		} else {
			// make the start date the beginning of this week
			weekday := creationDate.Weekday()
			preCleanedDate := creationDate.AddDate(0, 0, -int(weekday))
			newARInvoice.PeriodStart = time.Date(preCleanedDate.Year(), preCleanedDate.Month(), preCleanedDate.Day(), 0, 0, 0, 0, time.UTC)
		}
		newARInvoice.PeriodEnd = newARInvoice.PeriodStart.AddDate(0, 0, 13)
		if projectID != nil && !account.ProjectsSingleInvoice {
			var project Project
			a.DB.Where("ID = ?", *projectID).First(&project)
			newARInvoice.Name = account.Name + " - " + project.Name + ": " + newARInvoice.PeriodStart.Format("01.02.2006") + "-" + newARInvoice.PeriodEnd.Format("01.02.2006")
		} else {
			newARInvoice.Name = account.Name + ": " + newARInvoice.PeriodStart.Format("01.02.2006") + "-" + newARInvoice.PeriodEnd.Format("01.02.2006")
		}
	case BillingFrequencyWeekly.String():
		// retrieve the ending date of the last invoice
		if len(invoices) > 0 {
			preCleanedStart := invoices[0].PeriodEnd.AddDate(0, 0, 1)
			newARInvoice.PeriodStart = time.Date(preCleanedStart.Year(), preCleanedStart.Month(), preCleanedStart.Day(), 0, 0, 0, 0, time.UTC)
		} else {
			// make the start date the beginning of this week
			weekday := creationDate.Weekday()
			preCleanedDate := creationDate.AddDate(0, 0, -int(weekday))
			newARInvoice.PeriodStart = time.Date(preCleanedDate.Year(), preCleanedDate.Month(), preCleanedDate.Day(), 0, 0, 0, 0, time.UTC)
		}
		newARInvoice.PeriodEnd = newARInvoice.PeriodStart.AddDate(0, 0, 6)
		if projectID != nil && !account.ProjectsSingleInvoice {
			var project Project
			a.DB.Where("ID = ?", *projectID).First(&project)
			newARInvoice.Name = account.Name + " - " + project.Name + ": " + newARInvoice.PeriodStart.Format("01.02.2006") + "-" + newARInvoice.PeriodEnd.Format("01.02.2006")
		} else {
			newARInvoice.Name = account.Name + ": " + newARInvoice.PeriodStart.Format("01.02.2006") + "-" + newARInvoice.PeriodEnd.Format("01.02.2006")
		}
	}
	// Create the new invoice in the database
	a.DB.Create(&newARInvoice)
	return nil
}

// ApproveInvoice approves the invoice and transitions it to the "approved" state
func (a *App) ApproveInvoice(invoiceID uint) error {
	var invoice Invoice
	a.DB.Preload("Entries").Where("ID = ?", invoiceID).First(&invoice)
	if invoice.State != InvoiceStateDraft.String() {
		return InvalidPriorState
	}
	invoice.State = InvoiceStateApproved.String()
	for _, entry := range invoice.Entries {
		entry.State = EntryStateApproved.String()
		a.DB.Save(&entry)
	}
	a.DB.Save(&invoice)
	return nil
}

// SendInvoice sends the invoice to the client and transitions it to the "sent" state
func (a *App) SendInvoice(invoiceID uint) error {
	var invoice Invoice
	a.DB.Preload("Account").Where("ID = ?", invoiceID).First(&invoice)
	if invoice.State != InvoiceStateApproved.String() {
		return InvalidPriorState
	}
	invoice.State = InvoiceStateSent.String()
	invoice.SentAt = time.Now()
	// Set the due date based on invoice date (e.g., net 30)
	invoice.DueAt = invoice.SentAt.AddDate(0, 0, 30) // Default to 30 days
	a.DB.Save(&invoice)
	return nil
}

// MarkInvoicePaid pays the invoice and transitions it to the "paid" state
func (a *App) MarkInvoicePaid(invoiceID uint) error {
	log.Printf("MarkInvoicePaid called for invoice ID: %d", invoiceID)

	var invoice Invoice
	a.DB.Preload("Entries").Preload("Project").Where("ID = ?", invoiceID).First(&invoice)

	log.Printf("Loaded invoice ID: %d, Name: %s", invoice.ID, invoice.Name)

	// Check if invoice was found
	if invoice.ID == 0 {
		err := fmt.Errorf("invoice with ID %d not found", invoiceID)
		log.Printf("Error: %v", err)
		return err
	}

	// Make sure we have the full project data for commission calculations
	if invoice.ProjectID != nil {
		log.Printf("Loading project details for project ID: %d", *invoice.ProjectID)
		a.DB.Preload("AE").Preload("SDR").Where("ID = ?", *invoice.ProjectID).First(&invoice.Project)
		log.Printf("Loaded project ID: %d, Name: %s", invoice.Project.ID, invoice.Project.Name)

		if invoice.Project.ProjectType != "" {
			log.Printf("Project type: %s", invoice.Project.ProjectType)
		} else {
			log.Printf("Warning: Project type not set for project ID: %d", invoice.Project.ID)
		}

		if invoice.Project.AEID != nil {
			log.Printf("AE ID: %d", *invoice.Project.AEID)
		}

		if invoice.Project.SDRID != nil {
			log.Printf("SDR ID: %d", *invoice.Project.SDRID)
		}
	} else {
		log.Printf("Warning: No project associated with invoice ID: %d - continuing anyway", invoice.ID)
	}

	if invoice.State != InvoiceStateSent.String() {
		log.Printf("Invalid prior state: %s, expected: %s", invoice.State, InvoiceStateSent.String())
		return InvalidPriorState
	}

	log.Printf("Marking invoice ID: %d as paid", invoice.ID)
	invoice.State = InvoiceStatePaid.String()
	invoice.ClosedAt = time.Now()

	// First save the invoice state to ensure it's properly updated
	// Use explicit db.Model().Update() to ensure only the state and closedAt are updated
	if err := a.DB.Model(&invoice).Updates(map[string]interface{}{
		"state":     invoice.State,
		"closed_at": invoice.ClosedAt,
	}).Error; err != nil {
		log.Printf("Error updating invoice state: %v", err)
		return err
	}
	log.Printf("Successfully updated invoice state to PAID in the database")

	log.Printf("Updating %d entries to paid state", len(invoice.Entries))
	entriesUpdated := 0
	for i := range invoice.Entries {
		entry := &invoice.Entries[i]
		entry.State = EntryStatePaid.String()
		result := a.DB.Save(entry)
		if result.Error != nil {
			log.Printf("Error updating entry ID %d: %v", entry.ID, result.Error)
		} else {
			entriesUpdated++
		}
	}
	log.Printf("Successfully updated %d/%d entries to paid state", entriesUpdated, len(invoice.Entries))

	// Save all invoice changes now
	result := a.DB.Save(&invoice)
	if result.Error != nil {
		log.Printf("Error saving invoice: %v", result.Error)
		return result.Error
	}
	log.Printf("Saved all invoice ID: %d changes", invoice.ID)

	// Verify invoice state was properly saved
	var verifyInvoice Invoice
	if err := a.DB.Where("ID = ?", invoice.ID).First(&verifyInvoice).Error; err != nil {
		log.Printf("Error verifying invoice state: %v", err)
	} else {
		log.Printf("Verified invoice state: %s", verifyInvoice.State)
		if verifyInvoice.State != InvoiceStatePaid.String() {
			log.Printf("WARNING: Invoice state not properly saved!")
		}
	}

	// Generate bills for the invoice
	log.Printf("Generating bills for invoice ID: %d", invoice.ID)
	a.GenerateBills(&invoice)

	// Add commissions to bills if applicable
	log.Printf("Adding commissions to bills for invoice ID: %d", invoice.ID)
	a.AddCommissionsToBills(&invoice)

	log.Printf("Successfully processed invoice ID: %d", invoice.ID)
	return nil
}

// VoidInvoice cancels the invoice and transitions it to the "void" state along with any associated entries
func (a *App) VoidInvoice(invoiceID uint) error {
	var invoice Invoice
	a.DB.Preload("Entries").Preload("Account").Where("ID = ?", invoiceID).First(&invoice)
	// Invoice can be voided at any point
	invoice.State = InvoiceStateVoid.String()
	for _, entry := range invoice.Entries {
		entry.State = EntryStateVoid.String()
		a.DB.Save(&entry)
	}
	a.DB.Save(&invoice)
	return nil
}

// AssociateEntry associates an entry with the proper invoice. This function is called just after an entry is created
// and associates it to the appropriate invoice based on the entry date and the AP/AR state.
func (a *App) AssociateEntry(entry *Entry, projectID uint) error {
	if entry.Internal == true {
		return nil
	}

	// Retrieve the project
	var project Project
	a.DB.Where("ID = ?", projectID).First(&project)

	// Ensure that the entry date is within the project active start and end dates
	if entry.Start.Before(project.ActiveStart) || entry.Start.After(project.ActiveEnd) {
		return ErrEntryDateOutOfRange
	}

	// Get the account associated with the project
	var account Account
	a.DB.Where("ID = ?", project.AccountID).First(&account)

	// Retrieve the appropriate invoice
	var eligibleInvoices []Invoice

	// Check if this account uses separate invoices per project
	if account.ProjectsSingleInvoice {
		// Single invoice for all projects - look for account-level invoice
		a.DB.Order("period_end desc").Where("account_id = ? AND type = ? AND period_start <= ? AND period_end >= ? and state = ?",
			project.AccountID, InvoiceTypeAR.String(), entry.Start, entry.End, InvoiceStateDraft.String()).Find(&eligibleInvoices)
	} else {
		// Separate invoices per project - look for project-specific invoice
		a.DB.Order("period_end desc").Where("account_id = ? AND project_id = ? AND type = ? AND period_start <= ? AND period_end >= ? and state = ?",
			project.AccountID, projectID, InvoiceTypeAR.String(), entry.Start, entry.End, InvoiceStateDraft.String()).Find(&eligibleInvoices)
	}

	// If there are no eligible invoices, we'll create a new one
	if len(eligibleInvoices) == 0 {
		var projectIDPtr *uint
		if !account.ProjectsSingleInvoice {
			projectIDPtr = &projectID
		}
		err := a.CreateInvoice(project.AccountID, projectIDPtr, entry.Start)
		if err != nil {
			return err
		}

		// Query again for the new invoice
		if account.ProjectsSingleInvoice {
			a.DB.Order("period_end desc").Where("account_id = ? AND type = ? AND period_start <= ? AND period_end >= ? and state = ?",
				project.AccountID, InvoiceTypeAR.String(), entry.Start, entry.End, InvoiceStateDraft.String()).Find(&eligibleInvoices)
		} else {
			a.DB.Order("period_end desc").Where("account_id = ? AND project_id = ? AND type = ? AND period_start <= ? AND period_end >= ? and state = ?",
				project.AccountID, projectID, InvoiceTypeAR.String(), entry.Start, entry.End, InvoiceStateDraft.String()).Find(&eligibleInvoices)
		}
	}

	// We now need to provide a way to select the appropriate invoice if there are multiple. We'll do this via waterfall method.
	// If there is a pending invoice we'll add it to that, allowing us to edit invoices before they are sent. Otherwise, we'll
	// add it to the draft invoice. We are assuming that we cannot add entries to invoices that have already been approved or sent.
	var invoice Invoice
	for _, eligibleInvoice := range eligibleInvoices {
		if eligibleInvoice.State == InvoiceStateDraft.String() {
			invoice = eligibleInvoice
			break
		}
	}

	if invoice.ID == 0 {
		entry.State = EntryStateUnaffiliated.String()
	} else {
		entry.InvoiceID = &invoice.ID
		entry.State = EntryStateDraft.String()
	}

	a.DB.Save(&entry)

	// Update invoice totals after associating the entry
	if invoice.ID != 0 {
		a.UpdateInvoiceTotals(&invoice)
	}

	return nil
}

// SaveInvoiceToGCS saves the invoice to GCS
func (a *App) SaveInvoiceToGCS(invoice *Invoice) error {
	ctx := context.Background()
	// Generate the invoice
	// The output must be stored as a list of bytes in-memory becasue of the readonly filesystem in GAE
	pdfBytes := a.GenerateInvoicePDF(invoice)
	// Save the invoice to GCS
	client := a.InitializeStorageClient(a.Project, a.Bucket)

	// Create a bucket handle
	bucket := client.Bucket(a.Bucket)
	// Create a new object and write its contents to the bucket
	filename := GenerateSecureFilename(invoice.GetInvoiceFilename()) + ".pdf"
	objectName := "invoices/" + filename
	writer := bucket.Object(objectName).NewWriter(ctx)
	if _, err := writer.Write(pdfBytes); err != nil {
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}

	// keep object private and store link

	// save the public invoice URL to the database
	invoice.GCSFile = "https://storage.googleapis.com/" + a.Bucket + "/" + objectName
	a.DB.Save(&invoice)
	return nil
}

// BackfillEntriesForProject backfills entries to the invoice they belong on for a specific project
func (a *App) BackfillEntriesForProject(projectID string) {
	var entries []Entry
	a.DB.Where("project_id = ?", projectID).Find(&entries)
	for i, _ := range entries {
		err := a.AssociateEntry(&entries[i], entries[i].ProjectID)
		if err != nil {
			log.Println(err)
		}
	}
	return
}

// BackfillEntriesForAccount backfills all entries for all projects in an account
func (a *App) BackfillEntriesForAccount(accountID string) {
	var projects []Project
	a.DB.Where("account_id = ?", accountID).Find(&projects)

	for _, project := range projects {
		a.BackfillEntriesForProject(fmt.Sprintf("%d", project.ID))
	}
	return
}

// AddCommissionsToBills adds commission entries to bills for eligible staff members
func (a *App) AddCommissionsToBills(invoice *Invoice) {
	log.Printf("Starting AddCommissionsToBills for invoice ID: %d", invoice.ID)

	// If the invoice has a specific project ID, process it directly
	if invoice.ProjectID != nil {
		log.Printf("Processing single project commission for invoice ID: %d, Project ID: %d", invoice.ID, *invoice.ProjectID)
		a.processProjectCommission(invoice, *invoice.ProjectID)
		log.Printf("Completed single project commission processing for invoice ID: %d", invoice.ID)
		return
	}

	// For account-level invoices with no direct project ID, extract projects from entries
	log.Printf("No direct project ID for invoice ID: %d, processing multi-project invoice", invoice.ID)

	// Ensure entries are loaded
	if len(invoice.Entries) == 0 {
		log.Printf("Loading entries for invoice ID: %d", invoice.ID)
		if err := a.DB.Where("invoice_id = ?", invoice.ID).Find(&invoice.Entries).Error; err != nil {
			log.Printf("Error loading entries for invoice ID %d: %v", invoice.ID, err)
			return
		}
	}

	if len(invoice.Entries) == 0 {
		log.Printf("No entries found for invoice ID: %d, skipping commission processing", invoice.ID)
		return
	}

	// Group entries by project
	projectIDs := make(map[uint]bool)
	for _, entry := range invoice.Entries {
		if entry.ProjectID != 0 {
			projectIDs[entry.ProjectID] = true
		}
	}

	if len(projectIDs) == 0 {
		log.Printf("No projects identified from entries for invoice ID: %d", invoice.ID)
		return
	}

	log.Printf("Found %d distinct projects in invoice ID: %d", len(projectIDs), invoice.ID)

	// Process each project
	for projectID := range projectIDs {
		log.Printf("Processing project ID: %d from multi-project invoice ID: %d", projectID, invoice.ID)
		a.processProjectCommission(invoice, projectID)
	}

	log.Printf("Completed multi-project commission processing for invoice ID: %d", invoice.ID)
}

// processProjectCommission processes commissions for a specific project
func (a *App) processProjectCommission(invoice *Invoice, projectID uint) {
	// Load the full project with AE and SDR relationships
	var project Project
	result := a.DB.Preload("AE").Preload("SDR").Where("ID = ?", projectID).First(&project)
	if result.Error != nil {
		log.Printf("Error loading project ID %d with relationships: %v", projectID, result.Error)
		return
	}

	if project.ID == 0 {
		log.Printf("Project ID %d not found in database", projectID)
		return
	}

	log.Printf("Processing commission for project ID: %d, Name: %s", project.ID, project.Name)

	// Skip if project type is not set
	if project.ProjectType == "" {
		log.Printf("Skipping commission: Project type not set for project ID: %d", project.ID)
		return
	}

	// Skip if neither AE nor SDR is assigned
	if project.AEID == nil && project.SDRID == nil {
		log.Printf("Skipping commission: No AE or SDR assigned to project ID: %d", project.ID)
		return
	}

	// Calculate the invoice amount for this specific project
	var projectInvoiceTotal float64 = 0

	// If it's a direct project invoice, use the invoice total
	if invoice.ProjectID != nil && *invoice.ProjectID == projectID {
		projectInvoiceTotal = invoice.TotalFees
		log.Printf("Using direct invoice total for project %d: $%.2f", projectID, projectInvoiceTotal)
	} else {
		// Otherwise, sum the fees from entries for this specific project
		for _, entry := range invoice.Entries {
			if entry.ProjectID == projectID && entry.State != EntryStateVoid.String() {
				projectInvoiceTotal += float64(entry.Fee) / 100.0
			}
		}
		log.Printf("Calculated invoice total for project %d from entries: $%.2f", projectID, projectInvoiceTotal)
	}

	// Skip if invoice amount is zero
	if projectInvoiceTotal <= 0 {
		log.Printf("Skipping commission: Invoice amount is zero for project ID: %d", project.ID)
		return
	}

	log.Printf("Project type: %s, Invoice amount: $%.2f", project.ProjectType, projectInvoiceTotal)

	// Process AE commission if applicable
	if project.AEID != nil && project.AE != nil {
		log.Printf("Processing AE commission for %s %s (ID: %d)", project.AE.FirstName, project.AE.LastName, *project.AEID)
		a.processCommission(&project, CommissionRoleAE.String(), *project.AEID, project.AE.FirstName+" "+project.AE.LastName, projectInvoiceTotal)
	} else if project.AEID != nil {
		log.Printf("Warning: AE ID %d is set but AE data not loaded", *project.AEID)
		var employee Employee
		if err := a.DB.Where("ID = ?", *project.AEID).First(&employee).Error; err != nil {
			log.Printf("Error loading AE: %v", err)
		} else {
			log.Printf("Processing AE commission for %s %s (ID: %d)", employee.FirstName, employee.LastName, *project.AEID)
			a.processCommission(&project, CommissionRoleAE.String(), *project.AEID, employee.FirstName+" "+employee.LastName, projectInvoiceTotal)
		}
	}

	// Process SDR commission if applicable
	if project.SDRID != nil && project.SDR != nil {
		log.Printf("Processing SDR commission for %s %s (ID: %d)", project.SDR.FirstName, project.SDR.LastName, *project.SDRID)
		a.processCommission(&project, CommissionRoleSDR.String(), *project.SDRID, project.SDR.FirstName+" "+project.SDR.LastName, projectInvoiceTotal)
	} else if project.SDRID != nil {
		log.Printf("Warning: SDR ID %d is set but SDR data not loaded", *project.SDRID)
		var employee Employee
		if err := a.DB.Where("ID = ?", *project.SDRID).First(&employee).Error; err != nil {
			log.Printf("Error loading SDR: %v", err)
		} else {
			log.Printf("Processing SDR commission for %s %s (ID: %d)", employee.FirstName, employee.LastName, *project.SDRID)
			a.processCommission(&project, CommissionRoleSDR.String(), *project.SDRID, employee.FirstName+" "+employee.LastName, projectInvoiceTotal)
		}
	}
}

// processCommission creates a commission entry and adds it to the staff member's bill
func (a *App) processCommission(project *Project, role string, staffID uint, staffName string, invoiceTotal float64) {
	log.Printf("Starting processCommission for %s (ID: %d), role: %s", staffName, staffID, role)

	// Calculate commission amount using the invoice total
	commissionAmount := a.CalculateCommissionAmount(project, role, invoiceTotal)
	log.Printf("Calculated commission amount for %s: $%.2f", staffName, float64(commissionAmount)/100)

	// Skip if commission amount is zero
	if commissionAmount <= 0 {
		log.Printf("Skipping commission for %s: Amount is zero", staffName)
		return
	}

	// Get the latest bill for the staff member
	bill, err := a.GetLatestBillIfExists(staffID)
	if err != nil {
		log.Printf("Error getting latest bill for staff ID %d: %v", staffID, err)
		if !errors.Is(err, NoEligibleBill) {
			return
		}
	}

	// If no bill exists, create a new one
	firstOfMonth := time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.UTC)
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1)

	if err != nil && errors.Is(err, NoEligibleBill) {
		// Get the staff member
		var employee Employee
		if err := a.DB.Where("id = ?", staffID).First(&employee).Error; err != nil {
			log.Printf("Error loading employee ID %d: %v", staffID, err)
			return
		}
		log.Printf("Creating new bill for %s %s (ID: %d)", employee.FirstName, employee.LastName, staffID)

		// Create a new bill
		bill = Bill{
			Name:        "Payroll " + employee.FirstName + " " + employee.LastName + " " + firstOfMonth.Format("01/02/2006") + " - " + lastOfMonth.Format("01/02/2006"),
			EmployeeID:  staffID,
			PeriodStart: firstOfMonth,
			PeriodEnd:   lastOfMonth,
			TotalHours:  0,
			TotalFees:   0,
			TotalAmount: 0,
		}
		if err := a.DB.Create(&bill).Error; err != nil {
			log.Printf("Error creating new bill: %v", err)
			return
		}
		log.Printf("Created new bill ID: %d for staff ID: %d", bill.ID, staffID)
	} else {
		log.Printf("Using existing bill ID: %d for staff ID: %d", bill.ID, staffID)
	}

	// Check if a commission for this project and role already exists for this bill
	var existingCommission Commission
	if err := a.DB.Where("bill_id = ? AND project_id = ? AND role = ?", bill.ID, project.ID, role).First(&existingCommission).Error; err == nil && existingCommission.ID > 0 {
		log.Printf("Commission already exists (ID: %d) for project %s, role %s on bill %d", existingCommission.ID, project.Name, role, bill.ID)
		return
	}

	// Create the commission entry
	commission := Commission{
		StaffID:     staffID,
		Role:        role,
		Amount:      commissionAmount,
		BillID:      bill.ID,
		ProjectID:   project.ID,
		ProjectName: project.Name,
		ProjectType: project.ProjectType,
		Paid:        false,
	}

	// Save the commission
	if err := a.DB.Create(&commission).Error; err != nil {
		log.Printf("Error creating commission: %v", err)
		return
	}
	log.Printf("Created commission ID: %d for staff ID: %d, amount: $%.2f", commission.ID, staffID, float64(commissionAmount)/100)

	// Update the bill totals
	bill.TotalCommissions += commissionAmount
	bill.TotalAmount += commissionAmount
	if err := a.DB.Save(&bill).Error; err != nil {
		log.Printf("Error updating bill: %v", err)
		return
	}
	log.Printf("Updated bill ID: %d, new total commissions: $%.2f, new total amount: $%.2f",
		bill.ID, float64(bill.TotalCommissions)/100, float64(bill.TotalAmount)/100)

	// Verify commission was properly saved
	var verifyCommission Commission
	if err := a.DB.Where("ID = ?", commission.ID).First(&verifyCommission).Error; err != nil || verifyCommission.ID == 0 {
		log.Printf("Error verifying commission ID %d: %v", commission.ID, err)
	} else {
		log.Printf("Verified commission ID %d was properly saved", commission.ID)
	}

	// Regenerate the bill PDF
	err = a.SaveBillToGCS(&bill)
	if err != nil {
		log.Printf("Error saving bill to GCS: %v", err)
	} else {
		log.Printf("Successfully saved bill PDF to GCS for bill ID: %d", bill.ID)
	}

	log.Printf("Completed commission processing for %s", staffName)
}
