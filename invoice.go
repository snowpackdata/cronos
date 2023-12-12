package cronos

import (
	"errors"
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
	endOfMonth := startOfMonth.AddDate(0, 1, 0)
	// Retrieve the project from the database
	var project Project
	a.DB.Where("ID = ?", projectID).First(&project)
	// next retrieve any currently pending invoices.
	var invoices []Invoice
	a.DB.Preload("Entries").Order("period_end desc").Where("project_id = ? AND state = ?", projectID, InvoiceStateDraft).Find(&invoices)
	// By default we'll create the AP and AR invoices from the start of the current month to the end of the month
	var newARInvoice Invoice
	var newAPInvoice Invoice
	newARInvoice = Invoice{
		Name:        project.Name + " : " + startOfMonth.Format("01/02/2006") + " - " + endOfMonth.Format("01/02/2006"),
		ProjectID:   projectID,
		PeriodStart: startOfMonth,
		PeriodEnd:   endOfMonth,
		State:       InvoiceStateDraft.String(),
		Type:        InvoiceTypeAR.String(),
	}
	newAPInvoice = Invoice{
		Name:        project.Name + " : " + startOfMonth.Format("01/02/2006") + " - " + endOfMonth.Format("01/02/2006"),
		ProjectID:   projectID,
		PeriodStart: startOfMonth,
		PeriodEnd:   endOfMonth,
		State:       InvoiceStateDraft.String(),
		Type:        InvoiceTypeAR.String(),
	}
	// if there are no currently pending invoices then we'll update the start date to match the beginning of the project
	// and then move on, this allows us to backdate entries if they are added after the first of the month
	if len(invoices) == 0 {
		newARInvoice.PeriodStart = project.ActiveStart
	}
	// if the projects are billed on a project basis, we'll update the end date to match the end of the project
	if project.BillingFrequency == BillingFrequencyProject.String() {
		newARInvoice.PeriodEnd = project.ActiveEnd
		newAPInvoice.PeriodEnd = project.ActiveEnd
	}
	// if there are currently pending invoices, we'll confirm that there are no overlaps between the previously created
	// invoices and the new invoices. If there are overlaps, we'll generate an error and move on. Otherwise, we'll
	// move their state to pending along with associated entries.
	for _, invoice := range invoices {
		if invoice.PeriodEnd.After(newARInvoice.PeriodStart) {
			return ErrInvoiceOverlap
		}
		invoice.State = InvoiceStatePending.String()
		for _, entry := range invoice.Entries {
			entry.State = EntryStatePending.String()
			a.DB.Save(&entry)
		}
		a.DB.Save(&invoice)
	}
	// If there are no errors, we'll create the new invoices and save them to the database
	a.DB.Create(&newARInvoice)
	a.DB.Create(&newAPInvoice)
	return nil
}

// ApproveInvoice approves the invoice and transitions it to the "approved" state
func (a *App) ApproveInvoice(invoiceID uint) error {
	var invoice Invoice
	a.DB.Preload("Entries").Where("ID = ?", invoiceID).First(&invoice)
	if invoice.State != InvoiceStatePending.String() {
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
	var project Project
	a.DB.Where("ID = ?", projectID).First(&project)
	// Ensure that the entry date is within the project active start and end dates
	if entry.Start.Before(project.ActiveStart) || entry.Start.After(project.ActiveEnd) {
		return ErrEntryDateOutOfRange
	}
	// Retrieve the appropriate invoice
	var eligibleInvoices []Invoice
	var invoiceType string
	if entry.Internal {
		invoiceType = InvoiceTypeAP.String()
	} else {
		invoiceType = InvoiceTypeAR.String()
	}
	a.DB.Where("project_id = ? AND type = ? AND period_start <= ? AND period_end >= ?", projectID, invoiceType, entry.Start, entry.Start).Find(&eligibleInvoices)

	// If there are no eligible invoices, we'll create a new one
	if len(eligibleInvoices) == 0 {
		err := a.CreateInvoice(projectID, entry.Start)
		if err != nil {
			return err
		}
		a.DB.Where("project_id = ? AND type = ? AND period_start <= ? AND period_end >= ?", projectID, invoiceType, entry.Start, entry.Start).Find(&eligibleInvoices)
	}
	// We now need to provide a way to select the appropriate invoice if there are multiple. We'll do this via waterfall method.
	// If there is a pending invoice we'll add it to that, allowing us to edit invoices before they are sent. Otherwise, we'll
	// add it to the draft invoice. We are assuming that we cannot add entries to invoices that have already been approved or sent.
	var invoice Invoice
	for _, eligibleInvoice := range eligibleInvoices {
		if eligibleInvoice.State == InvoiceStatePending.String() {
			invoice = eligibleInvoice
			break
		}
		if eligibleInvoice.State == InvoiceStateDraft.String() {
			invoice = eligibleInvoice
		}
	}
	if invoice.ID == 0 {
		entry.State = EntryStateUnaffiliated.String()
	} else {
		entry.InvoiceID = invoice.ID
		entry.State = EntryStateDraft.String()
		return nil
	}
	return nil
}
