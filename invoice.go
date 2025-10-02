package cronos

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
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

// SaveBillPDFsForInvoice generates and saves PDFs for all bills associated with an invoice
func (a *App) SaveBillPDFsForInvoice(invoice *Invoice) {
	// Get all unique bill IDs from the invoice's entries
	billIDs := make(map[uint]bool)
	for _, entry := range invoice.Entries {
		if entry.BillID != nil && *entry.BillID > 0 {
			billIDs[*entry.BillID] = true
		}
	}

	// Save PDF for each bill
	for billID := range billIDs {
		var bill Bill
		if err := a.DB.Where("id = ?", billID).First(&bill).Error; err != nil {
			log.Printf("Error loading bill ID %d: %v", billID, err)
			continue
		}

		if err := a.SaveBillToGCS(&bill); err != nil {
			log.Printf("Error saving bill PDF for bill ID %d: %v", billID, err)
		} else {
			log.Printf("Successfully saved bill PDF for bill ID: %d", billID)
		}
	}
}

// ApproveInvoice approves the invoice and transitions it to the "approved" state
func (a *App) ApproveInvoice(invoiceID uint) error {
	log.Printf("ApproveInvoice called for invoice ID: %d", invoiceID)

	var invoice Invoice
	a.DB.Preload("Entries").Where("ID = ?", invoiceID).First(&invoice)
	if invoice.State != InvoiceStateDraft.String() {
		return InvalidPriorState
	}
	invoice.State = InvoiceStateApproved.String()

	log.Printf("Updating %d entries to approved state", len(invoice.Entries))
	for _, entry := range invoice.Entries {
		entry.State = EntryStateApproved.String()
		a.DB.Save(&entry)
	}

	// Also approve all draft adjustments on this invoice
	var adjustments []Adjustment
	if err := a.DB.Where("invoice_id = ? AND state = ?", invoiceID, AdjustmentStateDraft.String()).Find(&adjustments).Error; err == nil {
		log.Printf("Found %d draft adjustments to approve", len(adjustments))
		for _, adj := range adjustments {
			adj.State = AdjustmentStateApproved.String()
			a.DB.Save(&adj)
			log.Printf("Approved adjustment ID %d (type: %s, amount: $%.2f)", adj.ID, adj.Type, adj.Amount)
		}
	}

	// Update invoice totals to include adjustments
	a.UpdateInvoiceTotals(&invoice)

	a.DB.Save(&invoice)

	// Reload invoice with updated entries and account
	a.DB.Preload("Entries").Preload("Account").Where("ID = ?", invoiceID).First(&invoice)

	// Book accrual journal entries for approved work
	log.Printf("Booking accrual journal entries for invoice ID: %d", invoiceID)
	if err := a.BookInvoiceAccrual(&invoice); err != nil {
		log.Printf("Warning: Failed to book accrual for invoice %d: %v", invoiceID, err)
	}

	// Book adjustment accruals for all approved adjustments (only at approval - not at later states)
	log.Printf("Booking accrual journal entries for adjustments on invoice ID: %d", invoiceID)
	if err := a.DB.Where("invoice_id = ? AND state = ?", invoiceID, AdjustmentStateApproved.String()).Find(&adjustments).Error; err == nil {
		for _, adj := range adjustments {
			if err := a.BookAdjustmentAccrual(&adj); err != nil {
				log.Printf("Warning: Failed to book adjustment accrual for adjustment %d: %v", adj.ID, err)
			}
		}
	}

	// Generate bills for employees whose EntryPayEligibleState is ENTRY_STATE_APPROVED
	log.Printf("Generating bills for employees eligible at ENTRY_STATE_APPROVED")
	a.GenerateBills(&invoice)

	// Reload entries with bill associations and save PDFs
	a.DB.Preload("Entries").Where("ID = ?", invoiceID).First(&invoice)
	a.SaveBillPDFsForInvoice(&invoice)

	log.Printf("Successfully approved invoice ID: %d", invoiceID)
	return nil
}

// SendInvoice sends the invoice to the client and transitions it to the "sent" state
func (a *App) SendInvoice(invoiceID uint) error {
	log.Printf("SendInvoice called for invoice ID: %d", invoiceID)

	var invoice Invoice
	a.DB.Preload("Entries").Preload("Account").Where("ID = ?", invoiceID).First(&invoice)
	if invoice.State != InvoiceStateApproved.String() {
		return InvalidPriorState
	}
	invoice.State = InvoiceStateSent.String()
	invoice.SentAt = time.Now()
	// Set the due date based on invoice date (e.g., net 30)
	invoice.DueAt = invoice.SentAt.AddDate(0, 0, 30) // Default to 30 days

	// Update entry states to sent
	log.Printf("Updating %d entries to sent state", len(invoice.Entries))
	for _, entry := range invoice.Entries {
		entry.State = EntryStateSent.String()
		a.DB.Save(&entry)
	}

	// Also approve and process any draft adjustments added after initial approval
	var adjustments []Adjustment
	if err := a.DB.Where("invoice_id = ? AND state = ?", invoiceID, AdjustmentStateDraft.String()).Find(&adjustments).Error; err == nil {
		if len(adjustments) > 0 {
			log.Printf("Found %d draft adjustments to approve and process", len(adjustments))
			for _, adj := range adjustments {
				adj.State = AdjustmentStateApproved.String()
				a.DB.Save(&adj)
				log.Printf("Approved adjustment ID %d (type: %s, amount: $%.2f)", adj.ID, adj.Type, adj.Amount)
			}
		}
	}

	// Update invoice totals to include all adjustments
	a.UpdateInvoiceTotals(&invoice)

	a.DB.Save(&invoice)

	// Move from accrued receivables to formal accounts receivable (includes adjustments)
	log.Printf("Moving accrued receivables to AR for invoice ID: %d", invoiceID)
	if err := a.MoveInvoiceToAccountsReceivable(&invoice); err != nil {
		log.Printf("Warning: Failed to move to AR for invoice %d: %v", invoiceID, err)
	}

	// Reload invoice with updated entries before generating bills
	a.DB.Preload("Entries").Where("ID = ?", invoiceID).First(&invoice)

	// Generate bills for employees whose EntryPayEligibleState is ENTRY_STATE_SENT
	log.Printf("Generating bills for employees eligible at ENTRY_STATE_SENT")
	a.GenerateBills(&invoice)

	// Move accrued payroll to AP for all bills associated with this invoice
	log.Printf("Moving accrued payroll to AP for invoice ID: %d", invoiceID)
	var bills []Bill
	if err := a.DB.Preload("Entries").Preload("Entries.Employee").
		Joins("INNER JOIN entries ON entries.bill_id = bills.id").
		Where("entries.invoice_id = ?", invoiceID).
		Group("bills.id").
		Find(&bills).Error; err == nil {
		for _, bill := range bills {
			if err := a.BookBillAccrual(&bill); err != nil {
				log.Printf("Warning: Failed to move accruals to AP for bill %d: %v", bill.ID, err)
			}
		}
	}

	// Reload entries with bill associations and save PDFs
	a.DB.Preload("Entries").Where("ID = ?", invoiceID).First(&invoice)
	a.SaveBillPDFsForInvoice(&invoice)

	log.Printf("Successfully sent invoice ID: %d", invoiceID)
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

	// Also approve and process any draft adjustments added after sending
	var adjustments []Adjustment
	if err := a.DB.Where("invoice_id = ? AND state = ?", invoiceID, AdjustmentStateDraft.String()).Find(&adjustments).Error; err == nil {
		if len(adjustments) > 0 {
			log.Printf("Found %d draft adjustments to approve and process", len(adjustments))
			for _, adj := range adjustments {
				adj.State = AdjustmentStateApproved.String()
				a.DB.Save(&adj)
				log.Printf("Approved adjustment ID %d (type: %s, amount: $%.2f)", adj.ID, adj.Type, adj.Amount)
			}
		}
	}

	// Update invoice totals to include all adjustments
	a.UpdateInvoiceTotals(&invoice)

	// Reload invoice with updated entries before generating bills
	a.DB.Preload("Entries").Where("ID = ?", invoiceID).First(&invoice)

	// Generate bills for employees whose EntryPayEligibleState is ENTRY_STATE_PAID
	// Note: Employees with other eligible states (APPROVED, SENT) have already had bills generated
	log.Printf("Generating bills for employees eligible at ENTRY_STATE_PAID")
	a.GenerateBills(&invoice)

	// Move accrued payroll to AP for all bills associated with this invoice
	log.Printf("Moving accrued payroll to AP for invoice ID: %d", invoiceID)
	var bills []Bill
	if err := a.DB.Preload("Entries").Preload("Entries.Employee").
		Joins("INNER JOIN entries ON entries.bill_id = bills.id").
		Where("entries.invoice_id = ?", invoiceID).
		Group("bills.id").
		Find(&bills).Error; err == nil {
		for _, bill := range bills {
			if err := a.BookBillAccrual(&bill); err != nil {
				log.Printf("Warning: Failed to move accruals to AP for bill %d: %v", bill.ID, err)
			}
		}
	}

	// Add commissions to bills if applicable (commissions book their own journal entries)
	log.Printf("Adding commissions to bills for invoice ID: %d", invoice.ID)
	a.AddCommissionsToBills(&invoice)

	// Reload entries with bill associations and save PDFs for all bills
	// This ensures bills without commissions also get PDFs, and regenerates PDFs for bills with commissions
	a.DB.Preload("Entries").Where("ID = ?", invoiceID).First(&invoice)
	a.SaveBillPDFsForInvoice(&invoice)

	// Record cash receipt and clear accounts receivable (includes adjustments)
	log.Printf("Recording cash payment for invoice ID: %d", invoiceID)
	if err := a.RecordInvoiceCashPayment(&invoice); err != nil {
		log.Printf("Warning: Failed to record cash payment for invoice %d: %v", invoiceID, err)
	}

	log.Printf("Successfully processed invoice ID: %d", invoice.ID)
	return nil
}

// VoidInvoice cancels the invoice and transitions it to the "void" state along with any associated entries
// This also creates reversing journal entries to undo any accruals, AR, or cash entries
func (a *App) VoidInvoice(invoiceID uint) error {
	log.Printf("VoidInvoice called for invoice ID: %d", invoiceID)

	var invoice Invoice
	a.DB.Preload("Entries").Preload("Account").Where("ID = ?", invoiceID).First(&invoice)

	// Reverse all journal entries for this invoice
	log.Printf("Reversing journal entries for invoice ID: %d", invoiceID)
	if err := a.ReverseInvoiceJournalEntries(&invoice); err != nil {
		log.Printf("Warning: Failed to reverse journal entries for invoice %d: %v", invoiceID, err)
	}

	// Invoice can be voided at any point
	invoice.State = InvoiceStateVoid.String()
	for _, entry := range invoice.Entries {
		entry.State = EntryStateVoid.String()
		a.DB.Save(&entry)
	}
	a.DB.Save(&invoice)

	log.Printf("Successfully voided invoice ID: %d", invoiceID)
	return nil
}

// AssociateEntry associates an entry with the proper invoice. This function is called just after an entry is created
// and associates it to the appropriate invoice based on the entry date and the AP/AR state.
func (a *App) AssociateEntry(entry *Entry, projectID uint) error {
	if entry.Internal {
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

	// Find and associate the appropriate StaffingAssignment
	// Look for a staffing assignment that matches:
	// 1. The entry's employee
	// 2. The entry's project
	// 3. The entry date falls within the assignment's date range
	var staffingAssignment StaffingAssignment
	err := a.DB.Where("employee_id = ? AND project_id = ? AND start_date <= ? AND end_date >= ?",
		entry.EmployeeID, projectID, entry.Start, entry.Start).First(&staffingAssignment).Error

	if err == nil {
		// Found a matching staffing assignment
		entry.StaffingAssignmentID = &staffingAssignment.ID
		log.Printf("Associated entry ID %d with staffing assignment ID %d (Employee: %d, Project: %d)",
			entry.ID, staffingAssignment.ID, entry.EmployeeID, projectID)
	} else {
		// No matching staffing assignment found - this is okay, not all entries require assignments
		log.Printf("No staffing assignment found for entry ID %d (Employee: %d, Project: %d, Date: %s)",
			entry.ID, entry.EmployeeID, projectID, entry.Start.Format("2006-01-02"))
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
	writer.Close()

	// Set the object to be publicly accessible
	acl := bucket.Object(objectName).ACL()
	if err := acl.Set(ctx, storage.AllUsers, storage.RoleReader); err != nil {
		return err
	}

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
		projectInvoiceTotal = float64(invoice.TotalFees) / 100.0 // Convert cents to dollars
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

	// Book the commission as a payroll expense journal entry
	// Get employee details for subaccount
	var employee Employee
	if err := a.DB.Where("id = ?", staffID).First(&employee).Error; err != nil {
		log.Printf("Error loading employee for journal entry: %v", err)
		return
	}

	subAccount := fmt.Sprintf("%d:%s %s", employee.ID, employee.FirstName, employee.LastName)

	// DR: PAYROLL_EXPENSE
	commissionExpense := Journal{
		Account:    AccountPayrollExpense.String(),
		SubAccount: subAccount,
		BillID:     &bill.ID,
		Memo:       fmt.Sprintf("Commission expense for %s on project %s", role, project.Name),
		Debit:      int64(commissionAmount),
		Credit:     0,
	}
	if err := a.DB.Create(&commissionExpense).Error; err != nil {
		log.Printf("Warning: Failed to book commission expense: %v", err)
	}

	// CR: ACCOUNTS_PAYABLE
	commissionAP := Journal{
		Account:    AccountAccountsPayable.String(),
		SubAccount: subAccount,
		BillID:     &bill.ID,
		Memo:       fmt.Sprintf("Commission payable for %s on project %s", role, project.Name),
		Debit:      0,
		Credit:     int64(commissionAmount),
	}
	if err := a.DB.Create(&commissionAP).Error; err != nil {
		log.Printf("Warning: Failed to book commission AP: %v", err)
	}

	log.Printf("Booked commission journal entries for $%.2f", float64(commissionAmount)/100)

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

// BookInvoiceAccrual books accrual journal entries when an invoice is approved
// Books both revenue and expense sides:
// - DR: ACCRUED_RECEIVABLES, CR: REVENUE (for client billable amount)
// - DR: PAYROLL_EXPENSE, CR: ACCRUED_PAYROLL (for staff payable amount)
func (a *App) BookInvoiceAccrual(invoice *Invoice) error {
	log.Printf("BookInvoiceAccrual: Processing invoice ID %d", invoice.ID)

	// Calculate revenue total from approved entries
	var revenueAmount int64 = 0
	for _, entry := range invoice.Entries {
		if entry.State == EntryStateApproved.String() {
			revenueAmount += int64(entry.Fee)
		}
	}

	// Include adjustments in revenue
	var adjustments []Adjustment
	if err := a.DB.Where("invoice_id = ? AND state != ?", invoice.ID, AdjustmentStateVoid).Find(&adjustments).Error; err == nil {
		for _, adj := range adjustments {
			revenueAmount += int64(adj.Amount * 100)
		}
	}

	// Create subaccount identifier for revenue
	subAccount := fmt.Sprintf("%d:%s", invoice.AccountID, invoice.Account.Name)

	// Book revenue side if there's revenue to recognize
	if revenueAmount > 0 {
		// Book: DR ACCRUED_RECEIVABLES
		accrualDR := Journal{
			Account:    AccountAccruedReceivables.String(),
			SubAccount: subAccount,
			InvoiceID:  &invoice.ID,
			Memo:       fmt.Sprintf("Accrued receivables for approved work on invoice #%d", invoice.ID),
			Debit:      revenueAmount,
			Credit:     0,
		}
		if err := a.DB.Create(&accrualDR).Error; err != nil {
			return fmt.Errorf("failed to book accrued receivables DR: %w", err)
		}

		// Book: CR REVENUE
		revenueCR := Journal{
			Account:    AccountRevenue.String(),
			SubAccount: subAccount,
			InvoiceID:  &invoice.ID,
			Memo:       fmt.Sprintf("Revenue recognized for approved work on invoice #%d", invoice.ID),
			Debit:      0,
			Credit:     revenueAmount,
		}
		if err := a.DB.Create(&revenueCR).Error; err != nil {
			return fmt.Errorf("failed to book revenue CR: %w", err)
		}

		log.Printf("Booked revenue accrual for invoice ID %d: $%.2f", invoice.ID, float64(revenueAmount)/100)
	}

	// Now book payroll expense accruals for ALL approved entries
	// Group by employee to create one journal entry per employee
	employeeExpenses := make(map[uint]int64)
	employeeNames := make(map[uint]string)
	employeeCache := make(map[uint]*Employee)

	for _, entry := range invoice.Entries {
		if entry.State != EntryStateApproved.String() {
			continue
		}

		// Load employee (with caching to avoid duplicate queries)
		employee, exists := employeeCache[entry.EmployeeID]
		if !exists {
			var emp Employee
			if err := a.DB.First(&emp, entry.EmployeeID).Error; err != nil {
				log.Printf("Warning: Could not find employee %d for entry %d", entry.EmployeeID, entry.ID)
				continue
			}
			employee = &emp
			employeeCache[entry.EmployeeID] = employee
		}

		// Only accrue for variable or base+variable compensation
		if employee.CompensationType == string(CompensationTypeFullyVariable) ||
			employee.CompensationType == string(CompensationTypeBasePlusVariable) {

			if _, exists := employeeExpenses[employee.ID]; !exists {
				employeeExpenses[employee.ID] = 0
				employeeNames[employee.ID] = fmt.Sprintf("%s %s", employee.FirstName, employee.LastName)
			}

			// Calculate the internal/employee cost using the same logic as bills
			hourAmount := entry.Duration().Hours()
			rate := a.GetEmployeeBillRate(employee, entry.BillingCodeID)
			internalFee := int64(hourAmount * rate * 100)

			employeeExpenses[employee.ID] += internalFee

			log.Printf("  Entry ID %d for employee %d: %.2f hours at $%.2f/hr = $%.2f",
				entry.ID, employee.ID, hourAmount, rate, float64(internalFee)/100)
		}
	}

	// Book payroll expense accruals for each employee
	for employeeID, amount := range employeeExpenses {
		if amount == 0 {
			continue
		}

		employeeSubAccount := fmt.Sprintf("%d:%s", employeeID, employeeNames[employeeID])

		// Book: DR PAYROLL_EXPENSE
		expenseDR := Journal{
			Account:    AccountPayrollExpense.String(),
			SubAccount: employeeSubAccount,
			InvoiceID:  &invoice.ID,
			Memo:       fmt.Sprintf("Payroll expense accrued for approved work on invoice #%d", invoice.ID),
			Debit:      amount,
			Credit:     0,
		}
		if err := a.DB.Create(&expenseDR).Error; err != nil {
			log.Printf("Warning: Failed to book payroll expense DR for employee %d: %v", employeeID, err)
			continue
		}

		// Book: CR ACCRUED_PAYROLL
		payrollCR := Journal{
			Account:    AccountAccruedPayroll.String(),
			SubAccount: employeeSubAccount,
			InvoiceID:  &invoice.ID,
			Memo:       fmt.Sprintf("Accrued payroll for approved work on invoice #%d", invoice.ID),
			Debit:      0,
			Credit:     amount,
		}
		if err := a.DB.Create(&payrollCR).Error; err != nil {
			log.Printf("Warning: Failed to book accrued payroll CR for employee %d: %v", employeeID, err)
			continue
		}

		log.Printf("Booked payroll expense accrual for employee %d on invoice ID %d: $%.2f",
			employeeID, invoice.ID, float64(amount)/100)
	}

	log.Printf("Successfully booked all accruals for invoice ID %d", invoice.ID)
	return nil
}

// BookBillAccrual handles journal entries when a bill is created
// Since accruals are already booked at invoice approval, this moves them to AP
// Commissions are booked as new expenses since they don't exist at invoice approval time
func (a *App) BookBillAccrual(bill *Bill) error {
	log.Printf("BookBillAccrual: Processing bill ID %d", bill.ID)

	// Load employee once
	var employee Employee
	if err := a.DB.First(&employee, bill.EmployeeID).Error; err != nil {
		log.Printf("Error loading employee %d for bill %d", bill.EmployeeID, bill.ID)
		return fmt.Errorf("failed to load employee: %w", err)
	}

	employeeID := employee.ID
	employeeName := fmt.Sprintf("%s %s", employee.FirstName, employee.LastName)
	subAccount := fmt.Sprintf("%d:%s", employeeID, employeeName)

	// Only process variable or base+variable compensation
	if employee.CompensationType != string(CompensationTypeFullyVariable) &&
		employee.CompensationType != string(CompensationTypeBasePlusVariable) {
		log.Printf("Employee %d has fixed salary, no AP booking needed", employee.ID)
		return nil
	}

	// Calculate internal cost for entries (this was already accrued at invoice approval)
	var entriesAmount int64 = 0
	for _, entry := range bill.Entries {
		if entry.State == EntryStateVoid.String() {
			continue
		}

		hourAmount := entry.Duration().Hours()
		rate := a.GetEmployeeBillRate(&employee, entry.BillingCodeID)
		internalFee := int64(hourAmount * rate * 100)

		entriesAmount += internalFee
	}

	// Calculate adjustments separately
	var adjustmentsAmount int64 = 0
	var adjustments []Adjustment
	if err := a.DB.Where("bill_id = ? AND state != ?", bill.ID, AdjustmentStateVoid).Find(&adjustments).Error; err == nil {
		for _, adj := range adjustments {
			adjustmentsAmount += int64(adj.Amount * 100)
		}
	}

	// Note: Commissions are NOT handled here - they're booked directly when created via AddCommissionsToBills
	// Total AP will be entries + adjustments (commissions are booked separately)
	totalAPAmount := entriesAmount + adjustmentsAmount

	if totalAPAmount == 0 {
		log.Printf("No variable entries to process for bill ID %d", bill.ID)
		return nil
	}

	log.Printf("Bill breakdown - Entries: $%.2f, Adjustments: $%.2f, Total AP: $%.2f",
		float64(entriesAmount)/100, float64(adjustmentsAmount)/100, float64(totalAPAmount)/100)

	// Check if we've already moved accruals to AP for this bill
	var existingAPEntries []Journal
	a.DB.Where("bill_id = ? AND account = ?", bill.ID, AccountAccountsPayable.String()).Find(&existingAPEntries)

	if len(existingAPEntries) > 0 {
		log.Printf("Accruals already moved to AP for bill ID %d, skipping", bill.ID)
		return nil
	}

	// Check if accruals exist (from invoice approval)
	var existingAccruals []Journal
	a.DB.Where("sub_account = ? AND account = ? AND credit > 0",
		subAccount, AccountAccruedPayroll.String()).Find(&existingAccruals)

	if len(existingAccruals) > 0 {
		// Accruals exist for entries and adjustments, reverse them
		log.Printf("Moving accrued payroll to AP for bill ID %d (accrued: $%.2f)", bill.ID, float64(totalAPAmount)/100)

		// Reverse the accrued payroll (debit to contra it)
		reverseAccrual := Journal{
			Account:    AccountAccruedPayroll.String(),
			SubAccount: subAccount,
			BillID:     &bill.ID,
			Memo:       fmt.Sprintf("Move accrued payroll to AP for bill #%d", bill.ID),
			Debit:      totalAPAmount,
			Credit:     0,
		}
		if err := a.DB.Create(&reverseAccrual).Error; err != nil {
			log.Printf("Warning: Failed to reverse accrued payroll: %v", err)
		}

		// Book formal Accounts Payable for entries and adjustments
		// (Commissions will have their own AP entries created when they're added)
		apEntry := Journal{
			Account:    AccountAccountsPayable.String(),
			SubAccount: subAccount,
			BillID:     &bill.ID,
			Memo:       fmt.Sprintf("Accounts payable for bill #%d", bill.ID),
			Debit:      0,
			Credit:     totalAPAmount,
		}
		if err := a.DB.Create(&apEntry).Error; err != nil {
			return fmt.Errorf("failed to book accounts payable: %w", err)
		}

		log.Printf("Successfully moved accruals to AP for bill ID %d: $%.2f", bill.ID, float64(totalAPAmount)/100)
	} else {
		// No existing accruals (shouldn't happen if invoice was approved, but handle it)
		log.Printf("No existing accruals found for bill ID %d, booking new expense", bill.ID)

		// Book all as new payroll expense
		expenseDR := Journal{
			Account:    AccountPayrollExpense.String(),
			SubAccount: subAccount,
			BillID:     &bill.ID,
			Memo:       fmt.Sprintf("Payroll expense for bill #%d", bill.ID),
			Debit:      totalAPAmount,
			Credit:     0,
		}
		if err := a.DB.Create(&expenseDR).Error; err != nil {
			return fmt.Errorf("failed to book payroll expense DR: %w", err)
		}

		// Book CR ACCOUNTS_PAYABLE
		apEntry := Journal{
			Account:    AccountAccountsPayable.String(),
			SubAccount: subAccount,
			BillID:     &bill.ID,
			Memo:       fmt.Sprintf("Accounts payable for bill #%d", bill.ID),
			Debit:      0,
			Credit:     totalAPAmount,
		}
		if err := a.DB.Create(&apEntry).Error; err != nil {
			return fmt.Errorf("failed to book accounts payable CR: %w", err)
		}

		log.Printf("Successfully booked expense and AP for bill ID %d: $%.2f", bill.ID, float64(totalAPAmount)/100)
	}

	return nil
}

// MoveInvoiceToAccountsReceivable moves accrued receivables to formal AR when invoice is sent
// Reverses: DR ACCRUED_RECEIVABLES, CR: (contra)
// Books: DR ACCOUNTS_RECEIVABLE
func (a *App) MoveInvoiceToAccountsReceivable(invoice *Invoice) error {
	log.Printf("MoveInvoiceToAccountsReceivable: Processing invoice ID %d", invoice.ID)

	// Find existing accrued receivables journal entries for this invoice
	var accrualEntries []Journal
	if err := a.DB.Where("invoice_id = ? AND account = ?", invoice.ID, AccountAccruedReceivables.String()).Find(&accrualEntries).Error; err != nil {
		return fmt.Errorf("failed to find accrual entries: %w", err)
	}

	if len(accrualEntries) == 0 {
		log.Printf("No accrual entries found for invoice ID %d - skipping AR conversion", invoice.ID)
		return nil
	}

	// Calculate NET total from accrual entries (debits - credits)
	// This includes both the main invoice entries AND adjustments (fees/credits)
	var totalDebits int64 = 0
	var totalCredits int64 = 0
	var subAccount string
	for _, entry := range accrualEntries {
		totalDebits += entry.Debit
		totalCredits += entry.Credit
		subAccount = entry.SubAccount
	}

	netAmount := totalDebits - totalCredits
	if netAmount == 0 {
		return nil
	}

	// Reverse the accrued receivables (net amount)
	reverseAccrual := Journal{
		Account:    AccountAccruedReceivables.String(),
		SubAccount: subAccount,
		InvoiceID:  &invoice.ID,
		Memo:       fmt.Sprintf("Move accrued receivables to AR for sent invoice #%d", invoice.ID),
		Debit:      totalCredits, // Reverse the credits
		Credit:     totalDebits,  // Reverse the debits
	}
	if err := a.DB.Create(&reverseAccrual).Error; err != nil {
		return fmt.Errorf("failed to reverse accrued receivables: %w", err)
	}

	// Book formal Accounts Receivable (net amount)
	arEntry := Journal{
		Account:    AccountAccountsReceivable.String(),
		SubAccount: subAccount,
		InvoiceID:  &invoice.ID,
		Memo:       fmt.Sprintf("Accounts receivable for sent invoice #%d", invoice.ID),
		Debit:      netAmount,
		Credit:     0,
	}
	if err := a.DB.Create(&arEntry).Error; err != nil {
		return fmt.Errorf("failed to book accounts receivable: %w", err)
	}

	log.Printf("Successfully moved invoice ID %d to AR: $%.2f (net of adjustments)", invoice.ID, float64(netAmount)/100)
	return nil
}

// RecordInvoiceCashPayment records cash receipt and clears AR when invoice is paid
// Reverses: DR ACCOUNTS_RECEIVABLE, CR: (contra)
// Books: DR CASH
func (a *App) RecordInvoiceCashPayment(invoice *Invoice) error {
	log.Printf("RecordInvoiceCashPayment: Processing invoice ID %d", invoice.ID)

	// Find existing AR entries for this invoice
	var arEntries []Journal
	if err := a.DB.Where("invoice_id = ? AND account = ?", invoice.ID, AccountAccountsReceivable.String()).Find(&arEntries).Error; err != nil {
		return fmt.Errorf("failed to find AR entries: %w", err)
	}

	if len(arEntries) == 0 {
		log.Printf("No AR entries found for invoice ID %d - skipping cash recording", invoice.ID)
		return nil
	}

	// Calculate total from AR entries
	var totalAmount int64 = 0
	var subAccount string
	for _, entry := range arEntries {
		totalAmount += entry.Debit - entry.Credit
	}

	if totalAmount == 0 {
		return nil
	}

	subAccount = arEntries[0].SubAccount

	// Clear the accounts receivable (credit to contra it)
	clearAR := Journal{
		Account:    AccountAccountsReceivable.String(),
		SubAccount: subAccount,
		InvoiceID:  &invoice.ID,
		Memo:       fmt.Sprintf("Clear AR for paid invoice #%d", invoice.ID),
		Debit:      0,
		Credit:     totalAmount,
	}
	if err := a.DB.Create(&clearAR).Error; err != nil {
		return fmt.Errorf("failed to clear accounts receivable: %w", err)
	}

	// Record cash receipt
	cashEntry := Journal{
		Account:    AccountCash.String(),
		SubAccount: subAccount,
		InvoiceID:  &invoice.ID,
		Memo:       fmt.Sprintf("Cash received for invoice #%d", invoice.ID),
		Debit:      totalAmount,
		Credit:     0,
	}
	if err := a.DB.Create(&cashEntry).Error; err != nil {
		return fmt.Errorf("failed to record cash receipt: %w", err)
	}

	log.Printf("Successfully recorded cash payment for invoice ID %d: $%.2f", invoice.ID, float64(totalAmount)/100)
	return nil
}

// MoveBillToAccountsPayable moves accrued payroll to formal AP when bill is accepted
// Reverses: CR ACCRUED_PAYROLL, DR: (contra)
// Books: CR ACCOUNTS_PAYABLE
func (a *App) MoveBillToAccountsPayable(bill *Bill) error {
	log.Printf("MoveBillToAccountsPayable: Processing bill ID %d", bill.ID)

	// Find existing accrued payroll entries for this bill
	var accrualEntries []Journal
	if err := a.DB.Where("bill_id = ? AND account = ?", bill.ID, AccountAccruedPayroll.String()).Find(&accrualEntries).Error; err != nil {
		return fmt.Errorf("failed to find accrual entries: %w", err)
	}

	if len(accrualEntries) == 0 {
		log.Printf("No accrual entries found for bill ID %d - skipping AP conversion", bill.ID)
		return nil
	}

	// Calculate total from accrual entries
	var totalAmount int64 = 0
	var subAccount string
	for _, entry := range accrualEntries {
		totalAmount += entry.Credit
		subAccount = entry.SubAccount
	}

	if totalAmount == 0 {
		return nil
	}

	// Reverse the accrued payroll (debit to contra it)
	reverseAccrual := Journal{
		Account:    AccountAccruedPayroll.String(),
		SubAccount: subAccount,
		BillID:     &bill.ID,
		Memo:       fmt.Sprintf("Move accrued payroll to AP for accepted bill #%d", bill.ID),
		Debit:      totalAmount,
		Credit:     0,
	}
	if err := a.DB.Create(&reverseAccrual).Error; err != nil {
		return fmt.Errorf("failed to reverse accrued payroll: %w", err)
	}

	// Book formal Accounts Payable
	apEntry := Journal{
		Account:    AccountAccountsPayable.String(),
		SubAccount: subAccount,
		BillID:     &bill.ID,
		Memo:       fmt.Sprintf("Accounts payable for accepted bill #%d", bill.ID),
		Debit:      0,
		Credit:     totalAmount,
	}
	if err := a.DB.Create(&apEntry).Error; err != nil {
		return fmt.Errorf("failed to book accounts payable: %w", err)
	}

	log.Printf("Successfully moved bill ID %d to AP: $%.2f", bill.ID, float64(totalAmount)/100)
	return nil
}

// RecordBillCashPayment records cash payment and clears AP when bill is paid
// Reverses: CR ACCOUNTS_PAYABLE, DR: (contra)
// Books: CR CASH
func (a *App) RecordBillCashPayment(bill *Bill) error {
	log.Printf("RecordBillCashPayment: Processing bill ID %d", bill.ID)

	// Find existing AP entries for this bill
	var apEntries []Journal
	if err := a.DB.Where("bill_id = ? AND account = ?", bill.ID, AccountAccountsPayable.String()).Find(&apEntries).Error; err != nil {
		return fmt.Errorf("failed to find AP entries: %w", err)
	}

	if len(apEntries) == 0 {
		log.Printf("No AP entries found for bill ID %d - skipping cash recording", bill.ID)
		return nil
	}

	// Calculate total from AP entries
	var totalAmount int64 = 0
	var subAccount string
	for _, entry := range apEntries {
		totalAmount += entry.Credit - entry.Debit
	}

	if totalAmount == 0 {
		return nil
	}

	subAccount = apEntries[0].SubAccount

	// Clear the accounts payable (debit to contra it)
	clearAP := Journal{
		Account:    AccountAccountsPayable.String(),
		SubAccount: subAccount,
		BillID:     &bill.ID,
		Memo:       fmt.Sprintf("Clear AP for paid bill #%d", bill.ID),
		Debit:      totalAmount,
		Credit:     0,
	}
	if err := a.DB.Create(&clearAP).Error; err != nil {
		return fmt.Errorf("failed to clear accounts payable: %w", err)
	}

	// Record cash payment
	cashEntry := Journal{
		Account:    AccountCash.String(),
		SubAccount: subAccount,
		BillID:     &bill.ID,
		Memo:       fmt.Sprintf("Cash paid for bill #%d", bill.ID),
		Debit:      0,
		Credit:     totalAmount,
	}
	if err := a.DB.Create(&cashEntry).Error; err != nil {
		return fmt.Errorf("failed to record cash payment: %w", err)
	}

	log.Printf("Successfully recorded cash payment for bill ID %d: $%.2f", bill.ID, float64(totalAmount)/100)
	return nil
}

// ReverseInvoiceJournalEntries creates reversing entries for all journal entries associated with an invoice
// This is called when an invoice is voided to undo all accounting effects
func (a *App) ReverseInvoiceJournalEntries(invoice *Invoice) error {
	log.Printf("ReverseInvoiceJournalEntries: Processing invoice ID %d", invoice.ID)

	// Find all journal entries for this invoice
	var journalEntries []Journal
	if err := a.DB.Where("invoice_id = ?", invoice.ID).Find(&journalEntries).Error; err != nil {
		return fmt.Errorf("failed to find journal entries: %w", err)
	}

	if len(journalEntries) == 0 {
		log.Printf("No journal entries found for invoice ID %d", invoice.ID)
		return nil
	}

	log.Printf("Found %d journal entries to reverse for invoice ID %d", len(journalEntries), invoice.ID)

	// Create reversing entries (swap debit and credit)
	for _, entry := range journalEntries {
		reversingEntry := Journal{
			Account:    entry.Account,
			SubAccount: entry.SubAccount,
			InvoiceID:  &invoice.ID,
			Memo:       fmt.Sprintf("VOID: Reverse %s", entry.Memo),
			Debit:      entry.Credit, // Swap debit and credit
			Credit:     entry.Debit,
		}
		if err := a.DB.Create(&reversingEntry).Error; err != nil {
			log.Printf("Warning: Failed to create reversing entry for journal ID %d: %v", entry.ID, err)
		}
	}

	log.Printf("Successfully reversed %d journal entries for invoice ID %d", len(journalEntries), invoice.ID)
	return nil
}

// ReverseBillJournalEntries creates reversing entries for all journal entries associated with a bill
// This is called when a bill is voided to undo all accounting effects
func (a *App) ReverseBillJournalEntries(bill *Bill) error {
	log.Printf("ReverseBillJournalEntries: Processing bill ID %d", bill.ID)

	// Find all journal entries for this bill
	var journalEntries []Journal
	if err := a.DB.Where("bill_id = ?", bill.ID).Find(&journalEntries).Error; err != nil {
		return fmt.Errorf("failed to find journal entries: %w", err)
	}

	if len(journalEntries) == 0 {
		log.Printf("No journal entries found for bill ID %d", bill.ID)
		return nil
	}

	log.Printf("Found %d journal entries to reverse for bill ID %d", len(journalEntries), bill.ID)

	// Create reversing entries (swap debit and credit)
	for _, entry := range journalEntries {
		reversingEntry := Journal{
			Account:    entry.Account,
			SubAccount: entry.SubAccount,
			BillID:     &bill.ID,
			Memo:       fmt.Sprintf("VOID: Reverse %s", entry.Memo),
			Debit:      entry.Credit, // Swap debit and credit
			Credit:     entry.Debit,
		}
		if err := a.DB.Create(&reversingEntry).Error; err != nil {
			log.Printf("Warning: Failed to create reversing entry for journal ID %d: %v", entry.ID, err)
		}
	}

	log.Printf("Successfully reversed %d journal entries for bill ID %d", len(journalEntries), bill.ID)
	return nil
}

// BookAdjustmentAccrual books the initial accrual for an adjustment when it's approved
// Called only once at approval, then the adjustment follows the same transitions as the invoice/bill
func (a *App) BookAdjustmentAccrual(adjustment *Adjustment) error {
	log.Printf("BookAdjustmentAccrual: Processing adjustment ID %d", adjustment.ID)

	if adjustment.Amount == 0 {
		log.Printf("Adjustment amount is zero, skipping journal entry")
		return nil
	}

	// Always use absolute value - the type field determines if it's a credit or fee
	amountCents := int64(math.Abs(adjustment.Amount) * 100)

	// Handle invoice adjustments
	if adjustment.InvoiceID != nil {
		var invoice Invoice
		if err := a.DB.Preload("Account").First(&invoice, *adjustment.InvoiceID).Error; err != nil {
			return fmt.Errorf("failed to load invoice: %w", err)
		}

		subAccount := fmt.Sprintf("%d:%s", invoice.AccountID, invoice.Account.Name)
		isCredit := adjustment.Type == AdjustmentTypeCredit.String()

		if isCredit {
			// Credit reduces what we expect to receive
			// CR: ACCRUED_RECEIVABLES (reduce asset)
			// DR: CREDITS_ISSUED (contra-revenue)
			a.DB.Create(&Journal{
				Account:    string(AccountAccruedReceivables),
				SubAccount: subAccount,
				InvoiceID:  adjustment.InvoiceID,
				Memo:       fmt.Sprintf("Adjustment: Credit issued - %s", adjustment.Notes),
				Debit:      0,
				Credit:     amountCents,
			})
			a.DB.Create(&Journal{
				Account:    string(AccountCreditsIssued),
				SubAccount: subAccount,
				InvoiceID:  adjustment.InvoiceID,
				Memo:       fmt.Sprintf("Adjustment: Credit issued - %s", adjustment.Notes),
				Debit:      amountCents,
				Credit:     0,
			})
		} else {
			// Fee increases what we expect to receive
			// DR: ACCRUED_RECEIVABLES (increase asset)
			// CR: ADJUSTMENT_REVENUE
			a.DB.Create(&Journal{
				Account:    string(AccountAccruedReceivables),
				SubAccount: subAccount,
				InvoiceID:  adjustment.InvoiceID,
				Memo:       fmt.Sprintf("Adjustment: Fee added - %s", adjustment.Notes),
				Debit:      amountCents,
				Credit:     0,
			})
			a.DB.Create(&Journal{
				Account:    string(AccountAdjustmentRevenue),
				SubAccount: subAccount,
				InvoiceID:  adjustment.InvoiceID,
				Memo:       fmt.Sprintf("Adjustment: Fee added - %s", adjustment.Notes),
				Debit:      0,
				Credit:     amountCents,
			})
		}

		log.Printf("Recorded invoice adjustment accrual for adjustment ID %d: $%.2f", adjustment.ID, adjustment.Amount)
		return nil
	}

	// Handle bill adjustments (not typically used, but supported)
	if adjustment.BillID != nil {
		var bill Bill
		if err := a.DB.Preload("Employee").First(&bill, *adjustment.BillID).Error; err != nil {
			return fmt.Errorf("failed to load bill: %w", err)
		}

		subAccount := fmt.Sprintf("%d:%s %s", bill.EmployeeID, bill.Employee.FirstName, bill.Employee.LastName)

		// DR: ADJUSTMENT_EXPENSE, CR: ACCRUED_PAYROLL
		a.DB.Create(&Journal{
			Account:    string(AccountAdjustmentExpense),
			SubAccount: subAccount,
			BillID:     adjustment.BillID,
			Memo:       fmt.Sprintf("Adjustment: Expense addition - %s", adjustment.Notes),
			Debit:      amountCents,
			Credit:     0,
		})
		a.DB.Create(&Journal{
			Account:    string(AccountAccruedPayroll),
			SubAccount: subAccount,
			BillID:     adjustment.BillID,
			Memo:       fmt.Sprintf("Adjustment: Accrued payroll - %s", adjustment.Notes),
			Debit:      0,
			Credit:     amountCents,
		})

		log.Printf("Recorded bill adjustment accrual for adjustment ID %d: $%.2f", adjustment.ID, adjustment.Amount)
		return nil
	}

	return fmt.Errorf("adjustment must have either invoice_id or bill_id")
}

// RecordAdjustmentJournal creates a journal entry for an invoice or bill adjustment (credit/discount/fee)
// This handles adjustments based on the parent invoice/bill state
func (a *App) RecordAdjustmentJournal(adjustment *Adjustment) error {
	log.Printf("RecordAdjustmentJournal: Processing adjustment ID %d", adjustment.ID)

	if adjustment.Amount == 0 {
		log.Printf("Adjustment amount is zero, skipping journal entry")
		return nil
	}

	amountCents := int64(adjustment.Amount * 100)

	// Handle invoice adjustments
	if adjustment.InvoiceID != nil {
		var invoice Invoice
		if err := a.DB.Preload("Account").First(&invoice, *adjustment.InvoiceID).Error; err != nil {
			return fmt.Errorf("failed to load invoice: %w", err)
		}

		subAccount := fmt.Sprintf("%d:%s", invoice.AccountID, invoice.Account.Name)

		// Determine if this is a credit (reduces revenue) or fee (increases revenue)
		isCredit := adjustment.Type == AdjustmentTypeCredit.String()
		var revenueAccount JournalAccountType
		if isCredit {
			revenueAccount = AccountCreditsIssued // Contra-revenue
		} else {
			revenueAccount = AccountAdjustmentRevenue
		}

		// Book based on invoice state
		switch invoice.State {
		case InvoiceStateDraft.String():
			// Draft: Don't book yet, will be included when invoice is approved
			log.Printf("Invoice is draft, adjustment will be booked on approval")
			return nil

		case InvoiceStateApproved.String():
			// Approved: Book to accrued receivables
			if isCredit {
				// Credit reduces what we expect to receive
				// CR: ACCRUED_RECEIVABLES (reduce asset)
				// DR: CREDITS_ISSUED (contra-revenue)
				a.DB.Create(&Journal{
					Account:    string(AccountAccruedReceivables),
					SubAccount: subAccount,
					InvoiceID:  adjustment.InvoiceID,
					Memo:       fmt.Sprintf("Adjustment: Credit issued - %s", adjustment.Notes),
					Debit:      0,
					Credit:     amountCents,
				})
				a.DB.Create(&Journal{
					Account:    string(revenueAccount),
					SubAccount: subAccount,
					InvoiceID:  adjustment.InvoiceID,
					Memo:       fmt.Sprintf("Adjustment: Credit issued - %s", adjustment.Notes),
					Debit:      amountCents,
					Credit:     0,
				})
			} else {
				// Fee increases what we expect to receive
				// DR: ACCRUED_RECEIVABLES (increase asset)
				// CR: ADJUSTMENT_REVENUE
				a.DB.Create(&Journal{
					Account:    string(AccountAccruedReceivables),
					SubAccount: subAccount,
					InvoiceID:  adjustment.InvoiceID,
					Memo:       fmt.Sprintf("Adjustment: Fee added - %s", adjustment.Notes),
					Debit:      amountCents,
					Credit:     0,
				})
				a.DB.Create(&Journal{
					Account:    string(revenueAccount),
					SubAccount: subAccount,
					InvoiceID:  adjustment.InvoiceID,
					Memo:       fmt.Sprintf("Adjustment: Fee added - %s", adjustment.Notes),
					Debit:      0,
					Credit:     amountCents,
				})
			}

		case InvoiceStateSent.String():
			// Sent: Book to accounts receivable
			if isCredit {
				// CR: ACCOUNTS_RECEIVABLE (reduce asset)
				// DR: CREDITS_ISSUED
				a.DB.Create(&Journal{
					Account:    string(AccountAccountsReceivable),
					SubAccount: subAccount,
					InvoiceID:  adjustment.InvoiceID,
					Memo:       fmt.Sprintf("Adjustment: Credit issued - %s", adjustment.Notes),
					Debit:      0,
					Credit:     amountCents,
				})
				a.DB.Create(&Journal{
					Account:    string(revenueAccount),
					SubAccount: subAccount,
					InvoiceID:  adjustment.InvoiceID,
					Memo:       fmt.Sprintf("Adjustment: Credit issued - %s", adjustment.Notes),
					Debit:      amountCents,
					Credit:     0,
				})
			} else {
				// DR: ACCOUNTS_RECEIVABLE (increase asset)
				// CR: ADJUSTMENT_REVENUE
				a.DB.Create(&Journal{
					Account:    string(AccountAccountsReceivable),
					SubAccount: subAccount,
					InvoiceID:  adjustment.InvoiceID,
					Memo:       fmt.Sprintf("Adjustment: Fee added - %s", adjustment.Notes),
					Debit:      amountCents,
					Credit:     0,
				})
				a.DB.Create(&Journal{
					Account:    string(revenueAccount),
					SubAccount: subAccount,
					InvoiceID:  adjustment.InvoiceID,
					Memo:       fmt.Sprintf("Adjustment: Fee added - %s", adjustment.Notes),
					Debit:      0,
					Credit:     amountCents,
				})
			}

		case InvoiceStatePaid.String():
			// Paid: This is complex - the adjustment affects both revenue and cash
			// We need to record it as a separate transaction
			if isCredit {
				// We're giving back cash (refund) and reducing revenue
				// DR: CREDITS_ISSUED, CR: CASH
				a.DB.Create(&Journal{
					Account:    string(revenueAccount),
					SubAccount: subAccount,
					InvoiceID:  adjustment.InvoiceID,
					Memo:       fmt.Sprintf("Adjustment: Credit issued on paid invoice - %s", adjustment.Notes),
					Debit:      amountCents,
					Credit:     0,
				})
				a.DB.Create(&Journal{
					Account:    string(AccountCash),
					SubAccount: subAccount,
					InvoiceID:  adjustment.InvoiceID,
					Memo:       fmt.Sprintf("Adjustment: Refund for credit - %s", adjustment.Notes),
					Debit:      0,
					Credit:     amountCents,
				})
			} else {
				// We're receiving additional cash and increasing revenue
				// DR: CASH, CR: ADJUSTMENT_REVENUE
				a.DB.Create(&Journal{
					Account:    string(AccountCash),
					SubAccount: subAccount,
					InvoiceID:  adjustment.InvoiceID,
					Memo:       fmt.Sprintf("Adjustment: Additional payment for fee - %s", adjustment.Notes),
					Debit:      amountCents,
					Credit:     0,
				})
				a.DB.Create(&Journal{
					Account:    string(revenueAccount),
					SubAccount: subAccount,
					InvoiceID:  adjustment.InvoiceID,
					Memo:       fmt.Sprintf("Adjustment: Fee added to paid invoice - %s", adjustment.Notes),
					Debit:      0,
					Credit:     amountCents,
				})
			}
		}

		log.Printf("Recorded invoice adjustment journal for invoice ID %d (state: %s): $%.2f",
			invoice.ID, invoice.State, adjustment.Amount)
		return nil
	}

	// Handle bill adjustments
	if adjustment.BillID != nil {
		var bill Bill
		if err := a.DB.Preload("Employee").First(&bill, *adjustment.BillID).Error; err != nil {
			return fmt.Errorf("failed to load bill: %w", err)
		}

		subAccount := fmt.Sprintf("%d:%s %s", bill.EmployeeID, bill.Employee.FirstName, bill.Employee.LastName)

		// Determine bill state (draft, accepted, paid)
		isDraft := bill.AcceptedAt == nil || bill.AcceptedAt.IsZero()
		isPaid := bill.ClosedAt != nil && !bill.ClosedAt.IsZero()

		if isDraft {
			// Draft: Don't book yet
			log.Printf("Bill is draft, adjustment will be booked when bill is accepted")
			return nil
		} else if isPaid {
			// Paid: Book as additional expense and cash payment
			// DR: ADJUSTMENT_EXPENSE, CR: CASH
			a.DB.Create(&Journal{
				Account:    string(AccountAdjustmentExpense),
				SubAccount: subAccount,
				BillID:     adjustment.BillID,
				Memo:       fmt.Sprintf("Adjustment: Expense on paid bill - %s", adjustment.Notes),
				Debit:      amountCents,
				Credit:     0,
			})
			a.DB.Create(&Journal{
				Account:    string(AccountCash),
				SubAccount: subAccount,
				BillID:     adjustment.BillID,
				Memo:       fmt.Sprintf("Adjustment: Additional payment - %s", adjustment.Notes),
				Debit:      0,
				Credit:     amountCents,
			})
		} else {
			// Accepted but unpaid: Book to accounts payable
			// DR: ADJUSTMENT_EXPENSE, CR: ACCOUNTS_PAYABLE
			a.DB.Create(&Journal{
				Account:    string(AccountAdjustmentExpense),
				SubAccount: subAccount,
				BillID:     adjustment.BillID,
				Memo:       fmt.Sprintf("Adjustment: Expense addition - %s", adjustment.Notes),
				Debit:      amountCents,
				Credit:     0,
			})
			a.DB.Create(&Journal{
				Account:    string(AccountAccountsPayable),
				SubAccount: subAccount,
				BillID:     adjustment.BillID,
				Memo:       fmt.Sprintf("Adjustment: AP addition - %s", adjustment.Notes),
				Debit:      0,
				Credit:     amountCents,
			})
		}

		log.Printf("Recorded bill adjustment journal for bill ID %d: $%.2f", bill.ID, adjustment.Amount)
		return nil
	}

	return fmt.Errorf("adjustment must have either invoice_id or bill_id")
}
