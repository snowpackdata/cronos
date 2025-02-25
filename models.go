package cronos

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/pkg/errors"

	"log"

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

type ProjectType string

func (s ProjectType) String() string {
	return string(s)
}

type CommissionRole string

func (s CommissionRole) String() string {
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

	ProjectTypeNew      ProjectType = "PROJECT_TYPE_NEW"
	ProjectTypeExisting ProjectType = "PROJECT_TYPE_EXISTING"

	CommissionRoleAE  CommissionRole = "COMMISSION_ROLE_AE"
	CommissionRoleSDR CommissionRole = "COMMISSION_ROLE_SDR"

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

	// Commission rate constants
	// These rates are percentages (0.05 = 5%)
	AECommissionRateNewSmall = 0.08 // Projects under $10,000
	AECommissionRateNewLarge = 0.12 // Projects over $50,000

	// AE Commission Rates for Existing Business
	AECommissionRateExistingSmall = 0.032 // Projects under $10,000
	AECommissionRateExistingLarge = 0.072 // Projects over $50,000

	// SDR Commission Rates for New Business
	SDRCommissionRateNewSmall = 0.02 // Projects under $10,000
	SDRCommissionRateNewLarge = 0.03 // Projects over $50,000

	// SDR Commission Rates for Existing Business
	SDRCommissionRateExistingSmall = 0.008 // Projects under $10,000
	SDRCommissionRateExistingLarge = 0.018 // Projects over $50,000

	// Deal size thresholds (in dollars)
	DealSizeSmallThreshold = 100000
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
	Name                  string    `json:"name"`
	Type                  string    `json:"type"`
	LegalName             string    `gorm:"unique" json:"legal_name"`
	Address               string    `json:"address"`
	Email                 string    `json:"email"`
	Website               string    `json:"website"`
	Clients               []User    `json:"clients"`
	Projects              []Project `json:"projects"`
	Invoices              []Invoice `json:"invoices"`
	BillingFrequency      string    `json:"billing_frequency"`
	BudgetHours           int       `json:"budget_hours"`
	BudgetDollars         int       `json:"budget_dollars"`
	ProjectsSingleInvoice bool      `json:"projects_single_invoice"`
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
	ProjectType      string        `json:"project_type"`
	AEID             *uint         `json:"ae_id"`
	AE               *Employee     `json:"ae"`
	SDRID            *uint         `json:"sdr_id"`
	SDR              *Employee     `json:"sdr"`
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
	AccountID        uint         `json:"account_id"`
	Account          Account      `json:"account"`
	ProjectID        *uint        `json:"project_id"`
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

// Commission represents a commission payment to a staff member
type Commission struct {
	gorm.Model
	StaffID     uint     `json:"staff_id"`
	Staff       Employee `json:"staff"`
	Role        string   `json:"role"`
	Amount      int      `json:"amount"`
	BillID      uint     `json:"bill_id"`
	Bill        Bill     `json:"bill" gorm:"foreignKey:BillID"`
	ProjectID   uint     `json:"project_id"`
	ProjectName string   `json:"project_name"`
	ProjectType string   `json:"project_type"`
	Paid        bool     `json:"paid"`
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
	Commissions      []Commission `json:"commissions" gorm:"foreignKey:BillID"`
	AcceptedAt       *time.Time   `json:"accepted_at"`
	ClosedAt         *time.Time   `json:"closed_at"`
	TotalHours       float64      `json:"total_hours"`
	TotalFees        int          `json:"total_fees"`
	TotalAdjustments float64      `json:"total_adjustments"`
	TotalCommissions int          `json:"total_commissions"`
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
	ProjectName    string  `json:"project_name"`
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
	a.DB.Preload("BillingCode").Preload("BillingCode.Rate").Preload("Project").Where("invoice_id = ? AND state != ?", i.ID, EntryStateVoid.String()).Find(&entries)

	// Create a map to index line items - use a composite key of project ID + billing code
	billingCodeMap := make(map[string]InvoiceLineItem)

	// Populate the list of line items
	for _, entry := range entries {
		// Create a unique key that includes both project and billing code
		mapKey := fmt.Sprintf("%d-%s", entry.ProjectID, entry.BillingCode.Code)

		if lineItem, exists := billingCodeMap[mapKey]; exists {
			// Update the existing line item
			lineItem.Hours += entry.Duration().Hours()
			lineItem.Total = lineItem.Hours * lineItem.Rate
			billingCodeMap[mapKey] = lineItem
		} else {
			// Create a new line item
			billingCodeMap[mapKey] = InvoiceLineItem{
				BillingCode: entry.BillingCode.Code,
				Project:     entry.BillingCode.Name,
				ProjectName: entry.Project.Name,
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
	AccountID        uint         `json:"account_id"`
	AccountName      string       `json:"account_name"`
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
	AccountID      uint              `json:"account_id"`
	AccountName    string            `json:"account_name"`
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

	// Load the account and project if not already loaded
	if i.Account.ID == 0 {
		a.DB.Where("id = ?", i.AccountID).First(&i.Account)
	}

	draftInvoice := DraftInvoice{
		InvoiceID:        i.ID,
		InvoiceName:      i.Name,
		AccountID:        i.AccountID,
		AccountName:      i.Account.Name,
		PeriodStart:      i.PeriodStart.In(time.UTC).Format("01/02/2006"),
		PeriodEnd:        i.PeriodEnd.In(time.UTC).Format("01/02/2006"),
		TotalHours:       i.TotalHours,
		TotalFees:        i.TotalFees,
		TotalAdjustments: i.TotalAdjustments,
		TotalAmount:      i.TotalFees + i.TotalAdjustments,
	}

	// Handle project information if available
	if i.ProjectID != nil {
		if i.Project.ID == 0 {
			a.DB.Where("id = ?", *i.ProjectID).First(&i.Project)
		}
		draftInvoice.ProjectID = *i.ProjectID
		draftInvoice.ProjectName = i.Project.Name
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
	// Round to the nearest cent
	draftInvoice.TotalHours = math.Round(draftInvoice.TotalHours*100) / 100
	draftInvoice.TotalFees = math.Round(draftInvoice.TotalFees*100) / 100
	draftInvoice.TotalAdjustments = math.Round(draftInvoice.TotalAdjustments*100) / 100
	draftInvoice.TotalAmount = math.Round(draftInvoice.TotalAmount*100) / 100
	draftInvoice.PeriodClosed = i.PeriodEnd.In(time.UTC).Before(time.Now())
	return draftInvoice
}

func (a *App) GetAcceptedInvoice(i *Invoice) AcceptedInvoice {
	// Load the account and project if not already loaded
	if i.Account.ID == 0 {
		a.DB.Where("id = ?", i.AccountID).First(&i.Account)
	}

	acceptedInvoice := AcceptedInvoice{
		InvoiceID:   i.ID,
		InvoiceName: i.Name,
		AccountID:   i.AccountID,
		AccountName: i.Account.Name,
		PeriodStart: i.PeriodStart.In(time.UTC).Format("01/02/2006"),
		PeriodEnd:   i.PeriodEnd.In(time.UTC).Format("01/02/2006"),
		File:        i.GCSFile,
		TotalHours:  i.TotalHours,
		TotalFees:   i.TotalFees,
		State:       i.State,
		SentAt:      i.SentAt.Format("01/02/2006"),
		DueAt:       i.DueAt.Format("01/02/2006"),
		ClosedAt:    i.ClosedAt.Format("01/02/2006"),
	}

	// Handle project information if available
	if i.ProjectID != nil {
		if i.Project.ID == 0 {
			a.DB.Where("id = ?", *i.ProjectID).First(&i.Project)
		}
		acceptedInvoice.ProjectID = *i.ProjectID
		acceptedInvoice.ProjectName = i.Project.Name
	}

	acceptedInvoice.LineItems = a.GetInvoiceLineItems(i)
	acceptedInvoice.LineItemsCount = len(acceptedInvoice.LineItems)
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
		bill.TotalHours = math.Round(bill.TotalHours*100) / 100
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
	var account Account

	if i.ProjectID != nil && *i.ProjectID != 0 {
		a.DB.Preload("Account").Where("id = ?", *i.ProjectID).First(&project)
		account = project.Account
	} else {
		// If no project is associated, load the account directly
		a.DB.Where("id = ?", i.AccountID).First(&account)
	}

	journal := Journal{
		Account:    AccountARClientBillable.String(),
		SubAccount: account.LegalName,
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
			SubAccount: account.LegalName,
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

// CalculateCommissionRate determines the appropriate commission rate based on role, project type, and deal size
func (a *App) CalculateCommissionRate(role string, projectType string, dealSize int) float64 {
	isNew := projectType == ProjectTypeNew.String()
	isAE := role == CommissionRoleAE.String()

	log.Printf("Commission rate calculation - Role: %s, Project Type: %s, Deal Size: $%d",
		role, projectType, dealSize)

	// Determine deal size category
	var sizeCategory string
	if dealSize < DealSizeSmallThreshold {
		sizeCategory = "Small"
	} else {
		sizeCategory = "Large"
	}

	log.Printf("Deal size category: %s", sizeCategory)

	var rate float64
	// Return appropriate rate based on role, project type, and deal size
	if isAE {
		if isNew {
			if sizeCategory == "Small" {
				rate = AECommissionRateNewSmall
			} else {
				rate = AECommissionRateNewLarge
			}
		} else {
			if sizeCategory == "Small" {
				rate = AECommissionRateExistingSmall
			} else {
				rate = AECommissionRateExistingLarge
			}
		}
	} else { // SDR
		if isNew {
			if sizeCategory == "Small" {
				rate = SDRCommissionRateNewSmall
			} else {
				rate = SDRCommissionRateNewLarge
			}
		} else {
			if sizeCategory == "Small" {
				rate = SDRCommissionRateExistingSmall
			} else {
				rate = SDRCommissionRateExistingLarge
			}
		}
	}

	log.Printf("Selected commission rate: %.2f%%", rate*100)
	return rate
}

// CalculateCommissionAmount calculates the commission amount based on the project and role
func (a *App) CalculateCommissionAmount(project *Project, role string, invoiceTotal float64) int {
	// Calculate total project value based on invoice amount
	totalProjectValue := int(invoiceTotal)

	log.Printf("Commission calculation for invoice amount: $%d", totalProjectValue)

	// Get the commission rate
	rate := a.CalculateCommissionRate(role, project.ProjectType, totalProjectValue)

	// Log the commission calculation details
	log.Printf("Commission calculation - Role: %s, Project Type: %s, Invoice Total: $%d, Rate: %.2f%%",
		role, project.ProjectType, totalProjectValue, rate*100)

	// Calculate commission amount (in cents)
	commissionAmount := int(float64(totalProjectValue) * rate * 100)

	log.Printf("Calculated commission amount: $%.2f", float64(commissionAmount)/100)

	return commissionAmount
}

// calculateMonths calculates the number of whole months between two dates
func calculateMonths(start, end time.Time) int {
	months := (end.Year()-start.Year())*12 + int(end.Month()-start.Month())
	// Round up if there are days remaining
	if end.Day() > start.Day() {
		months++
	}
	if months < 1 {
		return 1 // Minimum of 1 month
	}
	return months
}

// calculateWeeks calculates the number of whole weeks between two dates
func calculateWeeks(start, end time.Time) int {
	days := int(end.Sub(start).Hours() / 24)
	weeks := (days + 6) / 7 // Round up to nearest week
	if weeks < 1 {
		return 1 // Minimum of 1 week
	}
	return weeks
}

// calculateBiWeeks calculates the number of bi-weekly periods between two dates
func calculateBiWeeks(start, end time.Time) int {
	days := int(end.Sub(start).Hours() / 24)
	biweeks := (days + 13) / 14 // Round up to nearest bi-week (14 days)
	if biweeks < 1 {
		return 1 // Minimum of 1 bi-week
	}
	return biweeks
}
