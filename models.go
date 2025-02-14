package cronos

import (
	"fmt"
	"github.com/pkg/errors"
	"strings"
	"time"

	_ "github.com/GoogleCloudPlatform/cloudsql-proxy/proxy/dialers/postgres"
	_ "gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var NoEligibleBill = errors.New("no eligible bill found")

// The following are all boilerplate methods and objects for
// standard interactions and comparisons with the database

type UserRole string

func (s UserRole) String() string {
	return string(s)
}

type AccountType string

func (s AccountType) String() string {
	return string(s)
}

type BillingFrequency string

func (s BillingFrequency) String() string {
	return string(s)
}

type RateType string

func (s RateType) String() string {
	return string(s)
}

type EntryState string

func (s EntryState) String() string {
	return string(s)
}

type InvoiceState string

func (s InvoiceState) String() string {
	return string(s)
}

type InvoiceType string

func (s InvoiceType) String() string {
	return string(s)
}

type AdjustmentType string

func (s AdjustmentType) String() string {
	return string(s)
}

type AdjustmentState string

func (s AdjustmentState) String() string {
	return string(s)
}

type JournalAccountType string

func (s JournalAccountType) String() string {
	return string(s)
}

const (
	HOUR                      = 60.0
	DEFAULT_PASSWORD          = "DEFAULT_PASSWORD"
	UserRoleAdmin    UserRole = "ADMIN"
	UserRoleStaff    UserRole = "STAFF"
	UserRoleClient   UserRole = "CLIENT"

	AccountTypeClient   AccountType = "ACCOUNT_TYPE_CLIENT"
	AccountTypeInternal AccountType = "ACCOUNT_TYPE_INTERNAL"

	RateTypeInternalBillable         RateType = "RATE_TYPE_INTERNAL_CLIENT_NON_BILLABLE"
	RateTypeInternalNonBillable      RateType = "RATE_TYPE_INTERNAL_CLIENT_BILLABLE"
	RateTypeInternalAdminBillable    RateType = "RATE_TYPE_INTERNAL_ADMINISTRATIVE"
	RateTypeInternalAdminNonBillable RateType = "RATE_TYPE_INTERNAL_ADMINISTRATIVE_NON_BILLABLE"
	RateTypeExternalBillable         RateType = "RATE_TYPE_EXTERNAL_CLIENT_BILLABLE"
	RateTypeExternalNonBillable      RateType = "RATE_TYPE_EXTERNAL_CLIENT_NON_BILLABLE"
	RateTypeInternalProject          RateType = "RATE_TYPE_INTERNAL_PROJECT"

	BillingFrequencyMonthly   BillingFrequency = "BILLING_TYPE_MONTHLY"
	BillingFrequencyProject   BillingFrequency = "BILLING_TYPE_PROJECT"
	BillingFrequencyBiweekly  BillingFrequency = "BILLING_TYPE_BIWEEKLY"
	BillingFrequencyWeekly    BillingFrequency = "BILLING_TYPE_WEEKLY"
	BillingFrequencyBiMonthly BillingFrequency = "BILLING_TYPE_BIMONTHLY"

	EntryStateUnaffiliated EntryState = "ENTRY_STATE_UNAFFILIATED"
	EntryStateDraft        EntryState = "ENTRY_STATE_DRAFT"
	EntryStateApproved     EntryState = "ENTRY_STATE_APPROVED"
	EntryStateSent         EntryState = "ENTRY_STATE_SENT"
	EntryStatePaid         EntryState = "ENTRY_STATE_PAID"
	EntryStateVoid         EntryState = "ENTRY_STATE_VOID"

	InvoiceStateDraft    InvoiceState = "INVOICE_STATE_DRAFT"
	InvoiceStateApproved InvoiceState = "INVOICE_STATE_APPROVED"
	InvoiceStateSent     InvoiceState = "INVOICE_STATE_SENT"
	InvoiceStatePaid     InvoiceState = "INVOICE_STATE_PAID"
	InvoiceStateVoid     InvoiceState = "INVOICE_STATE_VOID"

	InvoiceTypeAR InvoiceType = "INVOICE_TYPE_ACCOUNTS_RECEIVABLE"
	InvoiceTypeAP InvoiceType = "INVOICE_TYPE_ACCOUNTS_PAYABLE"

	AdjustmentTypeCredit AdjustmentType = "ADJUSTMENT_TYPE_CREDIT"
	AdjustmentTypeFee    AdjustmentType = "ADJUSTMENT_TYPE_FEE"

	AdjustmentStateDraft    AdjustmentState = "ADJUSTMENT_STATE_DRAFT"
	AdjustmentStateApproved AdjustmentState = "ADJUSTMENT_STATE_APPROVED"
	AdjustmentStateSent     AdjustmentState = "ADJUSTMENT_STATE_SENT"
	AdjustmentStatePaid     AdjustmentState = "ADJUSTMENT_STATE_PAID"
	AdjustmentStateVoid     AdjustmentState = "ADJUSTMENT_STATE_VOID"

	AccountARClientBillable JournalAccountType = "ACCOUNTS_RECEIVABLE_CLIENT_BILLABLE"
	AccountAPStaffPayroll   JournalAccountType = "ACCOUNTS_PAYABLE_STAFF_PAYROLL"
	AccountARIncome         JournalAccountType = "ACCOUNTS_RECEIVABLE_INCOME"
	AccountAPDiscount       JournalAccountType = "ACCOUNTS_PAYABLE_EXPENSE_DISCOUNT"
	AccountAPStaffBonus     JournalAccountType = "ACCOUNTS_PAYABLE_STAFF_BONUS"
	AccountARClientFee      JournalAccountType = "ACCOUNTS_RECEIVABLE_CLIENT_FEE"
)

type User struct {
	// User is the generic user object for anyone accessing the application
	gorm.Model
	Email     string `gorm:"unique" json:"email"`
	Password  string `json:"-"`
	IsAdmin   bool   `json:"is_admin"`
	Role      string `json:"role"`
	AccountID uint   `json:"account_id"`
}

type Employee struct {
	// Employee refers to internal information regarding an employee
	gorm.Model
	UserID    uint      `json:"user_id"`
	User      User      `json:"user"`
	Title     string    `json:"title"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	IsActive  bool      `json:"is_active"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	Entries   []Entry   `json:"entries"`
}

type Client struct {
	// Client refers to an external customer that may access the site to see time entries
	gorm.Model
	UserID    uint   `json:"user_id"`
	User      User   `json:"user"`
	Title     string `json:"title"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type Account struct {
	// Account is the specific customer account
	gorm.Model
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	LegalName string    `gorm:"unique" json:"legal_name"`
	Address   string    `json:"address"`
	Email     string    `json:"email"`
	Website   string    `json:"website"`
	Clients   []User    `json:"clients"`
	Projects  []Project `json:"projects"`
}

type Rate struct {
	// Rate stores all of available rates that can be added to individual projects
	gorm.Model
	Name         string    `json:"name"`
	Amount       float64   `json:"amount"`
	ActiveFrom   time.Time `json:"active_from"`
	ActiveTo     time.Time `json:"active_to"`
	InternalOnly bool      `json:"internal_only"`
	BillingCodes []BillingCode
}

type Project struct {
	// Project refers to a single unit of work with a customer
	// often with specific time period. A rate will have a specific billing code
	// associated with the project.
	gorm.Model
	Name             string        `json:"name"`
	AccountID        uint          `json:"account_id"`
	Account          Account       `json:"account"`
	ActiveStart      time.Time     `json:"active_start"`
	ActiveEnd        time.Time     `json:"active_end"`
	BudgetHours      int           `json:"budget_hours"`
	BudgetDollars    int           `json:"budget_dollars"`
	Internal         bool          `json:"internal"`
	BillingCodes     []BillingCode `json:"billing_codes"`
	Entries          []Entry       `json:"entries"`
	Invoices         []Invoice     `json:"invoices"`
	BillingFrequency string        `json:"billing_frequency"`
}

type BillingCode struct {
	gorm.Model
	Name           string    `json:"name"`
	RateType       string    `json:"type"`
	Category       string    `json:"category"`
	Code           string    `gorm:"unique" json:"code"`
	RoundedTo      int       `gorm:"default:15" json:"rounded_to"`
	ProjectID      uint      `json:"project"`
	ActiveStart    time.Time `json:"active_start"`
	ActiveEnd      time.Time `json:"active_end"`
	RateID         uint      `json:"rate_id"`
	Rate           Rate      `json:"rate"`
	InternalRateID uint      `json:"internal_rate_id"`
	InternalRate   Rate      `json:"internal_rate"`
	Entries        []Entry   `json:"entries"`
}
type Entry struct {
	gorm.Model
	ProjectID     uint        `json:"project_id"` // Can remove these, unnecessary with billing code
	Project       Project     `json:"project"`    // Can remove these, unnecessary with billing code
	Notes         string      `gorm:"type:varchar(2048)" json:"notes"`
	EmployeeID    uint        `json:"employee_id" gorm:"index:idx_employee_internal"`
	Employee      Employee    `json:"employee"`
	BillingCodeID uint        `json:"billing_code_id"`
	BillingCode   BillingCode `json:"billing_code"`
	Start         time.Time   `json:"start"`
	End           time.Time   `json:"end"`
	Internal      bool        `json:"internal" gorm:"index:idx_employee_internal"`
	Bill          Bill        `json:"bill"`
	BillID        *uint       `json:"bill_id"`
	Invoice       Invoice     `json:"invoice"`
	InvoiceID     *uint       `json:"invoice_id"`
	State         string      `json:"state"`
	Fee           int         `json:"fee"`
}

func (e *Entry) BeforeSave(tx *gorm.DB) (err error) {
	// recalculate the fee
	e.Fee = int(e.GetFee(tx) * 100)
	return nil
}

// Invoice is a record that is used to track the status of a billable invoice either as AR/AP.
// An invoice will have a collection of entries that are to be billed to a client as line items. While we use
// the term Invoice, these can mean either an invoice or bill in relationship to Snowpack.
type Invoice struct {
	gorm.Model
	Name             string       `json:"name"`
	ProjectID        uint         `json:"project_id"`
	Project          Project      `json:"project"`
	PeriodStart      time.Time    `json:"period_start"`
	PeriodEnd        time.Time    `json:"period_end"`
	Entries          []Entry      `json:"entries"`
	Adjustments      []Adjustment `json:"adjustments"`
	AcceptedAt       time.Time    `json:"accepted_at"`
	SentAt           time.Time    `json:"sent_at"`
	DueAt            time.Time    `json:"due_at"`
	ClosedAt         time.Time    `json:"closed_at"`
	State            string       `json:"state"`
	Type             string       `json:"type"`
	TotalHours       float64      `json:"total_hours"`
	TotalFees        float64      `json:"total_fees"`
	TotalAdjustments float64      `json:"total_adjustments"`
	TotalAmount      float64      `json:"total_amount"`
	JournalID        *uint        `json:"journal_id"`
	GCSFile          string       `json:"file"`
}

// Adjustment
// In the future I imagine that we will need to add an adjustment object to the invoice
// this will allow us to adjust the hours or the fee of an invoice in a single line item.
// This would allow us to add a credit or a discount to an invoice, or apply additional fees
// such as late fees or interest. This would be a separate object that would be added to the
// invoice as a line item.
type Adjustment struct {
	gorm.Model
	InvoiceID *uint   `json:"invoice_id"`
	Invoice   Invoice `json:"-"`
	Bill      Bill    `json:"-"`
	BillID    *uint   `json:"bill_id"`
	Type      string  `json:"type"`
	State     string  `json:"state"`
	Amount    float64 `json:"amount"`
	Notes     string  `json:"notes"`
}

// Bill
// This is a simple object that we can use to track the total hours and fees for an employee
// over a set period of time. While the entries may be tied to individual projects, the bills are directly
// linked to the employees.
type Bill struct {
	gorm.Model
	Name             string       `json:"name"`
	EmployeeID       uint         `json:"user_id"`
	Employee         Employee     `json:"user"`
	PeriodStart      time.Time    `json:"period_start"`
	PeriodEnd        time.Time    `json:"period_end"`
	Entries          []Entry      `json:"entries"`
	Adjustments      []Adjustment `json:"adjustments"`
	AcceptedAt       *time.Time   `json:"accepted_at"`
	ClosedAt         *time.Time   `json:"closed_at"`
	TotalHours       float64      `json:"total_hours"`
	TotalFees        int          `json:"total_fees"`
	TotalAdjustments float64      `json:"total_adjustments"`
	TotalAmount      int          `json:"total_amount"`
	GCSFile          string       `json:"file"`
}

// Journal refers to a single entry in a journal, this is a single line item that is used to track
// the debits and credits for a specific account.
type Journal struct {
	gorm.Model
	Account    string  `json:"account"`
	SubAccount string  `json:"sub_account"`
	Invoice    Invoice `json:"invoice"`
	InvoiceID  *uint   `json:"invoice_id"`
	Bill       Bill    `json:"bill"`
	BillID     *uint   `json:"bill_id"`
	Memo       string  `json:"memo"`
	Debit      int64   `json:"debit"`
	Credit     int64   `json:"credit"`
}

// WEBSITE SPECIFIC MODULES

// Survey is a simple object that we can use to track the responses to a survey
type Survey struct {
	gorm.Model
	SurveyType      string           `json:"survey_type"`
	UserEmail       string           `json:"user_email"`
	UserRole        string           `json:"user_role"`
	CompanyName     string           `json:"company_name"`
	Completed       bool             `json:"completed"`
	SurveyResponses []SurveyResponse `json:"survey_responses"`
}

// SurveyResponse is a response to a survey question. There may be any number of these
type SurveyResponse struct {
	gorm.Model
	SurveyID         uint   `json:"survey_id"`
	Survey           Survey `json:"survey"`
	Step             int    `json:"step"`
	Question         string `json:"question"`
	AnswerType       string `json:"answer_type"`
	StructuredAnswer string `json:"answer"`
	FreeformAnswer   string `json:"freeform_answer"`
}

// OBJECT METHODS

// Duration finds the length of an Entry as a duration object
func (e *Entry) Duration() time.Duration {
	duration := e.End.Sub(e.Start)
	return duration
}

// GetFee finds the applicable fee in USD for a particular entry rounded to the given minute
func (e *Entry) GetFee(tx *gorm.DB) float64 {
	var billingCode BillingCode
	var rate Rate
	tx.Where("id = ?", e.BillingCodeID).First(&billingCode)
	tx.Where("id = ?", billingCode.RateID).First(&rate)
	durationMinutes := e.Duration().Minutes()
	roundingFactor := float64(billingCode.RoundedTo) / HOUR
	hours := float64(durationMinutes) / HOUR
	roundedHours := float64(int(hours/roundingFactor)) * roundingFactor
	fee := roundedHours * rate.Amount
	return fee
}

// GetInternalFee gets the applicable fee in USD for an entry rounded to the minute
func (e *Entry) GetInternalFee(tx *gorm.DB) float64 {
	var billingCode BillingCode
	var rate Rate
	tx.Where("id = ?", e.BillingCodeID).First(&billingCode)
	tx.Where("id = ?", billingCode.InternalRateID).First(&rate)
	durationMinutes := e.Duration().Minutes()
	roundingFactor := float64(billingCode.RoundedTo) / HOUR
	hours := float64(durationMinutes) / HOUR
	roundedHours := float64(int(hours/roundingFactor)) * roundingFactor
	fee := roundedHours * rate.Amount
	return fee
}

// UpdateInvoiceTotals updates the totals for an invoice based on non-voided entries
// associated with the invoice. This saves us from having to recalculate the totals
func (a *App) UpdateInvoiceTotals(i *Invoice) {
	var totalHours float64
	var totalFeesInt int
	var totalAdjustments float64
	var entries []Entry
	a.DB.Where("invoice_id = ?", i.ID).Find(&entries)
	var adjustments []Adjustment
	a.DB.Where("invoice_id = ?", i.ID).Find(&adjustments)

	for _, entry := range entries {
		if entry.State != EntryStateVoid.String() {
			totalHours += entry.Duration().Hours()
			totalFeesInt += entry.Fee
		}
	}
	var multiplier float64
	for _, adjustment := range adjustments {
		if adjustment.Type == AdjustmentTypeCredit.String() {
			multiplier = -1.0
		} else {
			multiplier = 1.0
		}
		if adjustment.State != AdjustmentStateVoid.String() {
			totalAdjustments += adjustment.Amount * multiplier
		}
	}
	i.TotalHours = totalHours
	i.TotalFees = float64(totalFeesInt / 100)
	i.TotalAdjustments = totalAdjustments
	i.TotalAmount = i.TotalFees + i.TotalAdjustments
	a.DB.Omit(clause.Associations).Save(&i)
}

type InvoiceLineItem struct {
	BillingCode    string  `json:"billing_code"`
	Project        string  `json:"project"`
	Hours          float64 `json:"hours"`
	HoursFormatted string  `json:"hours_formatted"`
	Rate           float64 `json:"rate"`
	RateFormatted  string  `json:"rate_formatted"`
	Total          float64 `json:"total"`
}

type BillLineItem struct {
	BillingCode     string  `json:"billing_code"`
	BillingCodeCode string  `json:"billing_code_code"`
	Hours           float64 `json:"hours"`
	HoursFormatted  string  `json:"hours_formatted"`
	Rate            float64 `json:"rate"`
	RateFormatted   string  `json:"rate_formatted"`
	Total           float64 `json:"total"`
}
type invoiceEntry struct {
	dateString     string
	billingCode    string
	staff          string
	description    string
	hours          float64
	hoursFormatted string
}

func (a *App) GetInvoiceLineItems(i *Invoice) []InvoiceLineItem {
	var invoiceLineItems []InvoiceLineItem
	var entries []Entry
	a.DB.Preload("BillingCode").Preload("BillingCode.Rate").Where("invoice_id = ? AND state != ?", i.ID, EntryStateVoid.String()).Find(&entries)

	// Create a map to index line items
	billingCodeMap := make(map[string]InvoiceLineItem)

	// Populate the list of line items
	for _, entry := range entries {
		billingCode := entry.BillingCode.Code
		if lineItem, exists := billingCodeMap[billingCode]; exists {
			// Update the existing line item
			lineItem.Hours += entry.Duration().Hours()
			lineItem.Total = lineItem.Hours * lineItem.Rate
			billingCodeMap[billingCode] = lineItem
		} else {
			// Create a new line item
			billingCodeMap[billingCode] = InvoiceLineItem{
				BillingCode: entry.BillingCode.Code,
				Project:     entry.BillingCode.Name,
				Hours:       entry.Duration().Hours(),
				Rate:        entry.BillingCode.Rate.Amount,
				Total:       entry.Duration().Hours() * entry.BillingCode.Rate.Amount,
			}
		}
	}

	// Convert the map values to a slice
	for _, lineItem := range billingCodeMap {
		lineItem.HoursFormatted = fmt.Sprintf("%.2f", lineItem.Hours)
		lineItem.RateFormatted = fmt.Sprintf("%.2f", lineItem.Rate)
		invoiceLineItems = append(invoiceLineItems, lineItem)
	}

	for i, _ := range invoiceLineItems {
		invoiceLineItems[i].Total = invoiceLineItems[i].Hours * invoiceLineItems[i].Rate
		invoiceLineItems[i].HoursFormatted = fmt.Sprintf("%.2f", invoiceLineItems[i].Hours)
		invoiceLineItems[i].RateFormatted = fmt.Sprintf("%.2f", invoiceLineItems[i].Rate)
	}

	return invoiceLineItems
}

func (a *App) GetInvoiceEntries(i *Invoice) []invoiceEntry {
	var invoiceEntries []invoiceEntry
	var entries []Entry
	a.DB.Preload("BillingCode").Preload("Employee").Where("invoice_id = ? and state != ?", i.ID, EntryStateVoid.String()).Order("start asc").Find(&entries)
	for _, entry := range entries {
		invoiceEntries = append(invoiceEntries, invoiceEntry{
			dateString:     entry.Start.Format("01/02/2006"),
			billingCode:    entry.BillingCode.Code,
			staff:          entry.Employee.FirstName + " " + entry.Employee.LastName,
			description:    entry.Notes,
			hours:          entry.Duration().Hours(),
			hoursFormatted: fmt.Sprintf("%.2f", entry.Duration().Hours()),
		})
	}
	return invoiceEntries
}

func (a *App) GetInvoiceAdjustments(i *Invoice) []Adjustment {
	var adjustments []Adjustment
	a.DB.Where("invoice_id = ? and state != ?", i.ID, AdjustmentStateVoid.String()).Find(&adjustments)
	return adjustments
}

func (i *Invoice) GetInvoiceFilename() string {
	filename := strings.Replace(i.Name, " ", "_", -1)
	filename = strings.Replace(filename, ":", "", -1)
	filename = strings.Replace(filename, "/", ".", -1)
	filename = strings.Replace(filename, "(", "", -1)
	filename = strings.Replace(filename, ")", "", -1)
	filename = strings.Replace(filename, ",", "", -1)
	filename = strings.Replace(filename, "'", "", -1)
	return filename
}

func (b *Bill) GetBillFilename() string {
	filename := strings.Replace(b.Name, " ", "_", -1)
	filename = strings.Replace(filename, ":", "", -1)
	filename = strings.Replace(filename, "/", ".", -1)
	filename = strings.Replace(filename, "(", "", -1)
	filename = strings.Replace(filename, ")", "", -1)
	filename = strings.Replace(filename, ",", "", -1)
	filename = strings.Replace(filename, "'", "", -1)
	return filename
}

type ApiEntry struct {
	EntryID         uint      `json:"entry_id"`
	ProjectID       uint      `json:"project_id"`
	BillingCodeID   uint      `json:"billing_code_id"`
	BillingCode     string    `json:"billing_code"`
	BillingCodeName string    `json:"billing_code_name"`
	Start           time.Time `json:"start"`
	End             time.Time `json:"end"`
	Notes           string    `json:"notes"`
	StartDate       string    `json:"start_date"`
	StartHour       int       `json:"start_hour"`
	StartMinute     int       `json:"start_minute"`
	EndDate         string    `json:"end_date"`
	EndHour         int       `json:"end_hour"`
	EndMinute       int       `json:"end_minute"`
	DurationHours   float64   `json:"duration_hours"`
	StartDayOfWeek  string    `json:"start_day_of_week"`
	StartIndex      float64   `json:"start_index"`
	State           string    `json:"state"`
	Fee             float64   `json:"fee"`
}

func (e *Entry) GetAPIEntry() ApiEntry {
	apiEntry := ApiEntry{
		EntryID:         e.ID,
		ProjectID:       e.ProjectID,
		BillingCodeID:   e.BillingCodeID,
		BillingCode:     e.BillingCode.Code,
		BillingCodeName: e.BillingCode.Name,
		Start:           e.Start.In(time.UTC),
		End:             e.End.In(time.UTC),
		Notes:           e.Notes,
		StartDate:       e.Start.In(time.UTC).Format("2006-01-02"),
		StartHour:       e.Start.In(time.UTC).Hour(),
		StartMinute:     e.Start.Minute(),
		EndDate:         e.End.In(time.UTC).Format("2006-01-02"),
		EndHour:         e.End.In(time.UTC).Hour(),
		EndMinute:       e.End.Minute(),
		DurationHours:   e.Duration().Hours(),
		StartDayOfWeek:  e.Start.In(time.UTC).Weekday().String(),
		StartIndex:      float64(e.Start.In(time.UTC).Hour()) + (float64(e.Start.Minute()) / 60.0),
		State:           e.State,
		Fee:             float64(e.Fee) / 100.0,
	}
	return apiEntry
}

type DraftEntry struct {
	EntryID       uint    `json:"entry_id"`
	ProjectID     uint    `json:"project_id"`
	BillingCodeID uint    `json:"billing_code_id"`
	BillingCode   string  `json:"billing_code"`
	Notes         string  `json:"notes"`
	StartDate     string  `json:"start_date"`
	DurationHours float64 `json:"duration_hours"`
	Fee           float64 `json:"fee"`
	EmployeeName  string  `json:"user_name"`
	EmployeeRole  string  `json:"user_role"`
	State         string  `json:"state"`
}

func (a *App) GetDraftEntry(e *Entry) DraftEntry {
	var employee Employee
	a.DB.Where("id = ?", e.EmployeeID).First(&employee)
	var billingCode BillingCode
	a.DB.Where("id = ?", e.BillingCodeID).First(&billingCode)
	draftEntry := DraftEntry{
		EntryID:       e.ID,
		ProjectID:     e.ProjectID,
		BillingCodeID: e.BillingCodeID,
		BillingCode:   billingCode.Code,
		Notes:         e.Notes,
		StartDate:     e.Start.In(time.UTC).Format("January 2"),
		DurationHours: e.Duration().Hours(),
		Fee:           float64(e.Fee) / 100.0,
		EmployeeName:  employee.FirstName + " " + employee.LastName,
		EmployeeRole:  employee.Title,
		State:         e.State,
	}
	return draftEntry
}

type DraftInvoice struct {
	InvoiceID        uint         `json:"ID"`
	InvoiceName      string       `json:"invoice_name"`
	ProjectID        uint         `json:"project_id"`
	ProjectName      string       `json:"project_name"`
	PeriodStart      string       `json:"period_start"`
	PeriodEnd        string       `json:"period_end"`
	LineItems        []DraftEntry `json:"line_items"`
	Adjustments      []Adjustment `json:"adjustments"`
	TotalHours       float64      `json:"total_hours"`
	TotalFees        float64      `json:"total_fees"`
	TotalAdjustments float64      `json:"total_adjustments"`
	TotalAmount      float64      `json:"total_amount"`
	PeriodClosed     bool         `json:"period_closed"`
}

type AcceptedInvoice struct {
	InvoiceID      uint              `json:"ID"`
	InvoiceName    string            `json:"invoice_name"`
	ProjectID      uint              `json:"project_id"`
	ProjectName    string            `json:"project_name"`
	PeriodStart    string            `json:"period_start"`
	PeriodEnd      string            `json:"period_end"`
	File           string            `json:"file"`
	LineItemsCount int               `json:"line_items_count"`
	TotalHours     float64           `json:"total_hours"`
	TotalFees      float64           `json:"total_fees"`
	State          string            `json:"state"`
	SentAt         string            `json:"sent_at"`
	DueAt          string            `json:"due_at"`
	ClosedAt       string            `json:"closed_at"`
	LineItems      []InvoiceLineItem `json:"line_items"`
}

func (a *App) GetDraftInvoice(i *Invoice) DraftInvoice {
	a.UpdateInvoiceTotals(i)
	draftInvoice := DraftInvoice{
		InvoiceID:        i.ID,
		InvoiceName:      i.Name,
		ProjectID:        i.ProjectID,
		ProjectName:      i.Project.Name,
		PeriodStart:      i.PeriodStart.In(time.UTC).Format("01/02/2006"),
		PeriodEnd:        i.PeriodEnd.In(time.UTC).Format("01/02/2006"),
		TotalHours:       i.TotalHours,
		TotalFees:        i.TotalFees,
		TotalAdjustments: i.TotalAdjustments,
		TotalAmount:      i.TotalFees + i.TotalAdjustments,
	}
	for _, entry := range i.Entries {
		draftEntry := a.GetDraftEntry(&entry)
		draftInvoice.LineItems = append(draftInvoice.LineItems, draftEntry)
	}
	a.DB.Where("invoice_id = ?", i.ID).Find(&draftInvoice.Adjustments)
	var totalAdjustments float64
	for i, _ := range draftInvoice.Adjustments {
		multiplicationFactor := 1.0
		if draftInvoice.Adjustments[i].Type == AdjustmentTypeCredit.String() {
			multiplicationFactor = -1.0
		}
		if draftInvoice.Adjustments[i].State != AdjustmentStateVoid.String() {
			totalAdjustments += draftInvoice.Adjustments[i].Amount * multiplicationFactor
		}

	}
	draftInvoice.TotalAdjustments = totalAdjustments
	draftInvoice.TotalAmount = draftInvoice.TotalFees + draftInvoice.TotalAdjustments
	draftInvoice.PeriodClosed = i.PeriodEnd.In(time.UTC).Before(time.Now())
	return draftInvoice
}

func (a *App) GetAcceptedInvoice(i *Invoice) AcceptedInvoice {
	a.UpdateInvoiceTotals(i)
	acceptedInvoice := AcceptedInvoice{
		InvoiceID:      i.ID,
		InvoiceName:    i.Name,
		ProjectID:      i.ProjectID,
		ProjectName:    i.Project.Name,
		PeriodStart:    i.PeriodStart.In(time.UTC).Format("01/02/2006"),
		PeriodEnd:      i.PeriodEnd.In(time.UTC).Format("01/02/2006"),
		File:           i.GCSFile,
		LineItemsCount: len(i.Entries),
		State:          i.State,
		SentAt:         i.SentAt.In(time.UTC).Format("01/02/2006"),
		DueAt:          i.DueAt.In(time.UTC).Format("01/02/2006"),
		ClosedAt:       i.ClosedAt.In(time.UTC).Format("01/02/2006"),
		TotalHours:     i.TotalHours,
		TotalFees:      i.TotalFees,
	}
	acceptedInvoice.LineItems = a.GetInvoiceLineItems(i)
	return acceptedInvoice
}

// GetLatestBillIfExists returns the latest bill for a user if they have an active bill open
// otherwise it returns an error that indicates that they do not have an active bill
func (a *App) GetLatestBillIfExists(userID uint) (Bill, error) {
	var bill Bill
	// An active bill is one that has not been accepted, voided, or paid
	// and is active within the current month
	a.DB.Where("closed_at is null and employee_id = ?", userID).Order("period_end desc").First(&bill)
	if bill.Name == "" {
		return Bill{}, NoEligibleBill
	}
	return bill, nil
}

func (a *App) GenerateBills(i *Invoice) {
	userBillingCodeMap := make(map[uint]map[uint]float64)
	// Add up each of the fees for a user and billing code
	// and cache the value
	for _, entry := range i.Entries {
		if entry.State == EntryStateVoid.String() {
			continue
		}
		// first check if the user exists in the map, if not then add them
		if _, ok := userBillingCodeMap[entry.EmployeeID][entry.BillingCodeID]; !ok {
			// if the users and billing code do not exist, add them
			bc := make(map[uint]float64)
			bc[entry.BillingCodeID] = entry.Duration().Minutes()
			userBillingCodeMap[entry.EmployeeID] = bc
		} else {
			// otherwise add the fee to the existing fee
			userBillingCodeMap[entry.EmployeeID][entry.BillingCodeID] += entry.Duration().Minutes()
		}
	}
	// Now we need to iterate over the map and create or add to the bill for each user
	for user, billingCodes := range userBillingCodeMap {
		var userObj Employee
		a.DB.Where("id = ?", user).First(&userObj)
		var hours float64
		var fee int
		for billingCode, loopMin := range billingCodes {
			// sum up the hours for the billing code
			var billingCodeObj BillingCode
			a.DB.Preload("InternalRate").Where("id = ?", billingCode).First(&billingCodeObj)
			floatFee := (loopMin / 60) * billingCodeObj.InternalRate.Amount
			fee += int(floatFee * 100)
			hours += (loopMin / 60)
		}
		// See if there is an existing bill for the user
		var bill Bill
		var err error
		bill, err = a.GetLatestBillIfExists(user)
		// If there is no bill, then create a new one for the user for this month
		firstOfMonth := time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.UTC)
		lastOfMonth := firstOfMonth.AddDate(0, 1, -1)
		if err != nil && errors.Is(err, NoEligibleBill) {
			// Create a new bill for the user
			// First We need the user's name
			bill = Bill{
				Name:        "Payroll " + userObj.FirstName + " " + userObj.LastName + " " + firstOfMonth.Format("01/02/2006") + " - " + lastOfMonth.Format("01/02/2006"),
				EmployeeID:  user,
				PeriodStart: firstOfMonth,
				PeriodEnd:   lastOfMonth,
				TotalHours:  0,
				TotalFees:   0,
				TotalAmount: 0,
			}
			a.DB.Create(&bill)
		}
		// Add the fees to the existing bill
		bill.TotalHours += hours
		bill.TotalFees += fee
		bill.TotalAmount += fee
		a.DB.Save(&bill)

		// Update the entries to associate with the bill
		for _, entry := range i.Entries {
			if entry.EmployeeID == user {
				entry.BillID = &bill.ID
				a.DB.Save(&entry)
			}
		}
		// Save the bill
		err = a.SaveBillToGCS(&bill)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func (a *App) MarkBillPaid(b *Bill) {
	nowTime := time.Now()
	b.ClosedAt = &nowTime
	a.DB.Save(&b)

	// Get the User
	var userObj Employee
	a.DB.Where("id = ?", b.EmployeeID).First(&userObj)
	// Now add a journal entry to reflect the bill
	journal := Journal{
		Account:    AccountAPStaffPayroll.String(),
		SubAccount: userObj.FirstName + " " + userObj.LastName,
		Debit:      int64(b.TotalFees),
		Credit:     0,
		Memo:       b.Name,
		BillID:     &b.ID,
	}
	a.DB.Create(&journal)

}

func (a *App) AddJournalEntries(i *Invoice) {
	// First we need to add the entries for the total fee of the invoice
	var project Project
	a.DB.Preload("Account").Where("id = ?", i.ProjectID).First(&project)
	journal := Journal{
		Account:    AccountARClientBillable.String(),
		SubAccount: project.Account.LegalName,
		Debit:      0,
		Credit:     int64(i.TotalFees * 100),
		Memo:       i.Name,
		InvoiceID:  &i.ID,
	}
	a.DB.Create(&journal)
	// Associate the invoice
	i.JournalID = &journal.ID
	a.DB.Save(&i)
	// Now we need to add an entry for any adjustments
	if i.TotalAdjustments != 0 {
		adjustmentJournal := Journal{
			Account:    AccountARClientFee.String(),
			SubAccount: project.Account.LegalName,
			Debit:      0,
			Credit:     int64(i.TotalAdjustments * 100),
			Memo:       i.Name,
			InvoiceID:  &i.ID,
		}
		a.DB.Create(&adjustmentJournal)
	}
	return
}

func (a *App) GetBillLineItems(b *Bill) []BillLineItem {
	var billLineItems []BillLineItem
	var entries []Entry
	a.DB.Preload("BillingCode").Preload("BillingCode.InternalRate").Where("bill_id = ? AND state != ?", b.ID, EntryStateVoid.String()).Find(&entries)

	// Create a map to index line items
	billingCodeMap := make(map[string]BillLineItem)

	// Populate the list of line items
	for _, entry := range entries {
		billingCode := entry.BillingCode.Code
		if lineItem, exists := billingCodeMap[billingCode]; exists {
			// Update the existing line item
			lineItem.Hours += entry.Duration().Hours()
			lineItem.Total = lineItem.Hours * lineItem.Rate
			billingCodeMap[billingCode] = lineItem
		} else {
			// Create a new line item
			billingCodeMap[billingCode] = BillLineItem{
				BillingCode:     entry.BillingCode.Name,
				BillingCodeCode: entry.BillingCode.Code,
				Hours:           entry.Duration().Hours(),
				Rate:            entry.BillingCode.InternalRate.Amount,
				Total:           entry.Duration().Hours() * entry.BillingCode.InternalRate.Amount,
			}
		}
	}

	// Convert the map values to a slice
	for _, lineItem := range billingCodeMap {
		lineItem.HoursFormatted = fmt.Sprintf("%.2f", lineItem.Hours)
		lineItem.RateFormatted = fmt.Sprintf("%.2f", lineItem.Rate)
		billLineItems = append(billLineItems, lineItem)
	}

	for i, _ := range billLineItems {
		billLineItems[i].Total = billLineItems[i].Hours * billLineItems[i].Rate
		billLineItems[i].HoursFormatted = fmt.Sprintf("%.2f", billLineItems[i].Hours)
		billLineItems[i].RateFormatted = fmt.Sprintf("%.2f", billLineItems[i].Rate)
	}

	return billLineItems
}
