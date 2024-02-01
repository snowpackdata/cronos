package cronos

import (
	"cloud.google.com/go/storage"
	"context"
	"errors"
	"log"
	"time"
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
func (a *App) CreateInvoice(projectID uint, creationDate time.Time) error {
	// We need a timestamp to determine the start and end of the month
	startOfMonth := time.Date(creationDate.Year(), creationDate.Month(), 1, 0, 0, 0, 0, time.UTC)
	endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-time.Nanosecond)
	// Retrieve the project from the database
	var project Project
	a.DB.Where("ID = ?", projectID).First(&project)
	// next retrieve any currently pending invoices.
	var invoices []Invoice
	a.DB.Preload("Entries").Order("period_end desc").Where("project_id = ? AND state = ?", projectID, InvoiceStateDraft).Find(&invoices)
	// By default we'll create the AP invoice from the start of the current month to the end of the month
	var newARInvoice Invoice
	newARInvoice = Invoice{
		Name:        project.Name + ": " + startOfMonth.Format("01.02.2006") + "-" + endOfMonth.Format("01.02.2006"),
		ProjectID:   projectID,
		PeriodStart: startOfMonth,
		PeriodEnd:   endOfMonth,
		State:       InvoiceStateDraft.String(),
		Type:        InvoiceTypeAR.String(),
	}
	// if the projects are billed on a project basis, we'll update the start and end date to match the project
	if project.BillingFrequency == BillingFrequencyProject.String() {
		newARInvoice.PeriodStart = project.ActiveStart
		newARInvoice.PeriodEnd = project.ActiveEnd
	}
	// if there are currently other draft invoices, we'll confirm that there are no overlaps between the previously created
	// invoices and the new invoices. If there are overlaps we will generate an error and move on. Otherwise, we'll
	// move their state to pending along with associated entries.
	for _, invoice := range invoices {
		if invoice.PeriodEnd.After(newARInvoice.PeriodStart) && (invoice.State != InvoiceStateDraft.String() || invoice.State != InvoiceStateVoid.String()) {
			return ErrInvoiceOverlap
		}
	}
	// If there are no errors, we'll create the new invoices and save them to the database
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
	a.DB.Where("ID = ?", invoiceID).First(&invoice)
	if invoice.State != InvoiceStateApproved.String() {
		return InvalidPriorState
	}
	invoice.State = InvoiceStateSent.String()
	a.DB.Save(&invoice)
	return nil
}

// MarkInvoicePaid pays the invoice and transitions it to the "paid" state
func (a *App) MarkInvoicePaid(invoiceID uint) error {
	var invoice Invoice
	a.DB.Preload("Entries").Where("ID = ?", invoiceID).First(&invoice)
	if invoice.State != InvoiceStateSent.String() {
		return InvalidPriorState
	}
	invoice.State = InvoiceStatePaid.String()
	for _, entry := range invoice.Entries {
		entry.State = EntryStatePaid.String()
		a.DB.Save(&entry)
	}
	a.DB.Save(&invoice)
	return nil
}

// VoidInvoice cancels the invoice and transitions it to the "void" state along with any associated entries
func (a *App) VoidInvoice(invoiceID uint) error {
	var invoice Invoice
	a.DB.Preload("Entries").Where("ID = ?", invoiceID).First(&invoice)
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
	var project Project
	a.DB.Where("ID = ?", projectID).First(&project)
	// Ensure that the entry date is within the project active start and end dates
	if entry.Start.Before(project.ActiveStart) || entry.Start.After(project.ActiveEnd) {
		return ErrEntryDateOutOfRange
	}
	// Retrieve the appropriate invoice
	var eligibleInvoices []Invoice
	a.DB.Where("project_id = ? AND type = ? AND period_start <= ? AND period_end >= ? and state = ?", projectID, InvoiceTypeAR.String(), entry.Start, entry.End, InvoiceStateDraft.String()).Find(&eligibleInvoices)

	// If there are no eligible invoices, we'll create a new one
	if len(eligibleInvoices) == 0 {
		err := a.CreateInvoice(projectID, entry.Start)
		if err != nil {
			return err
		}
		a.DB.Where("project_id = ? AND type = ? AND period_start <= ? AND period_end >= ? and state = ?", projectID, InvoiceTypeAR.String(), entry.Start, entry.End, InvoiceStateDraft.String()).Find(&eligibleInvoices)
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
		entry.InvoiceID = invoice.ID
		entry.State = EntryStateDraft.String()
	}
	a.DB.Save(&entry)
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

// BackfillEntriesForProject backfills entries to the invoice they belong on
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
