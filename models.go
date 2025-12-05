package cronos

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"strings"
	"time"

	"cloud.google.com/go/storage"

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

type BillState string

func (s BillState) String() string {
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

type EmploymentStatus string

func (s EmploymentStatus) String() string {
	return string(s)
}

type CompensationType string

func (c CompensationType) String() string {
	return string(c)
}

type LineItemType string

func (l LineItemType) String() string {
	return string(l)
}

type ExpenseState string

func (e ExpenseState) String() string {
	return string(e)
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
	EntryStateRejected     EntryState = "ENTRY_STATE_REJECTED" // Rejected - staff NOT paid, shows red in UI
	EntryStateApproved     EntryState = "ENTRY_STATE_APPROVED"
	EntryStateExcluded     EntryState = "ENTRY_STATE_EXCLUDED" // Staff paid, but not billed to client
	EntryStateSent         EntryState = "ENTRY_STATE_SENT"
	EntryStatePaid         EntryState = "ENTRY_STATE_PAID"
	EntryStateVoid         EntryState = "ENTRY_STATE_VOID"

	InvoiceStateDraft    InvoiceState = "INVOICE_STATE_DRAFT"
	InvoiceStateApproved InvoiceState = "INVOICE_STATE_APPROVED"
	InvoiceStateSent     InvoiceState = "INVOICE_STATE_SENT"
	InvoiceStatePaid     InvoiceState = "INVOICE_STATE_PAID"
	InvoiceStateVoid     InvoiceState = "INVOICE_STATE_VOID"

	BillStateDraft    BillState = "BILL_STATE_DRAFT"
	BillStateAccepted BillState = "BILL_STATE_ACCEPTED"
	BillStatePaid     BillState = "BILL_STATE_PAID"
	BillStateVoid     BillState = "BILL_STATE_VOID"

	InvoiceTypeAR InvoiceType = "INVOICE_TYPE_ACCOUNTS_RECEIVABLE"
	InvoiceTypeAP InvoiceType = "INVOICE_TYPE_ACCOUNTS_PAYABLE"

	AdjustmentTypeCredit AdjustmentType = "ADJUSTMENT_TYPE_CREDIT"
	AdjustmentTypeFee    AdjustmentType = "ADJUSTMENT_TYPE_FEE"

	AdjustmentStateDraft    AdjustmentState = "ADJUSTMENT_STATE_DRAFT"
	AdjustmentStateApproved AdjustmentState = "ADJUSTMENT_STATE_APPROVED"
	AdjustmentStateSent     AdjustmentState = "ADJUSTMENT_STATE_SENT"
	AdjustmentStatePaid     AdjustmentState = "ADJUSTMENT_STATE_PAID"
	AdjustmentStateVoid     AdjustmentState = "ADJUSTMENT_STATE_VOID"

	// New accrual accounting structure
	// Assets
	AccountAccruedReceivables JournalAccountType = "ACCRUED_RECEIVABLES"
	AccountAccountsReceivable JournalAccountType = "ACCOUNTS_RECEIVABLE"
	AccountCash               JournalAccountType = "CASH"

	// Liabilities
	AccountAccruedPayroll         JournalAccountType = "ACCRUED_PAYROLL"
	AccountAccountsPayable        JournalAccountType = "ACCOUNTS_PAYABLE"
	AccountAccruedExpensesPayable JournalAccountType = "ACCRUED_EXPENSES_PAYABLE" // Contra account for unreconciled expenses

	// Revenue
	AccountRevenue           JournalAccountType = "REVENUE"
	AccountAdjustmentRevenue JournalAccountType = "ADJUSTMENT_REVENUE"

	// Contra-Revenue
	AccountCreditsIssued JournalAccountType = "CREDITS_ISSUED"
	AccountDiscounts     JournalAccountType = "DISCOUNTS"

	// Expenses
	AccountPayrollExpense    JournalAccountType = "PAYROLL_EXPENSE"
	AccountAdjustmentExpense JournalAccountType = "ADJUSTMENT_EXPENSE"

	// Operating Expenses (from Beancount)
	AccountOperatingExpensesSaaS          JournalAccountType = "OPERATING_EXPENSES_SAAS"
	AccountOperatingExpensesTravel        JournalAccountType = "OPERATING_EXPENSES_TRAVEL"
	AccountOperatingExpensesEquipment     JournalAccountType = "OPERATING_EXPENSES_EQUIPMENT"
	AccountOperatingExpensesFees          JournalAccountType = "OPERATING_EXPENSES_FEES"
	AccountOperatingExpensesLegal         JournalAccountType = "OPERATING_EXPENSES_LEGAL"
	AccountOperatingExpensesDiscretionary JournalAccountType = "OPERATING_EXPENSES_DISCRETIONARY"
	AccountOperatingExpensesTaxes         JournalAccountType = "OPERATING_EXPENSES_TAXES"
	AccountOperatingExpensesVendors       JournalAccountType = "OPERATING_EXPENSES_VENDORS"
	AccountOperatingExpensesOffice        JournalAccountType = "OPERATING_EXPENSES_OFFICE"
	AccountOwnerDistributions             JournalAccountType = "OWNER_DISTRIBUTIONS"

	// Additional Assets/Liabilities (from Beancount)
	AccountEquipment         JournalAccountType = "EQUIPMENT"
	AccountCreditCardPayable JournalAccountType = "CREDIT_CARD_PAYABLE"
	AccountEquityOwnership   JournalAccountType = "EQUITY_OWNERSHIP"
	AccountEquityPool        JournalAccountType = "EQUITY_POOL"
	AccountEquipmentExpense  JournalAccountType = "EQUIPMENT_EXPENSE"

	// Pass-through Expense Accounts
	AccountExpensePassThrough JournalAccountType = "EXPENSE_PASS_THROUGH"

	// Catch-all accounts
	AccountOtherAssets      JournalAccountType = "OTHER_ASSETS"
	AccountOtherLiabilities JournalAccountType = "OTHER_LIABILITIES"
	AccountOtherIncome      JournalAccountType = "OTHER_INCOME"
	AccountOtherExpenses    JournalAccountType = "OTHER_EXPENSES"
	AccountEquity           JournalAccountType = "EQUITY"
	AccountUnclassified     JournalAccountType = "UNCLASSIFIED"

	// AECommission rate constants
	// These rates are percentages (0.05 = 5%)
	AECommissionRateNewSmall = 0.08 // Projects under $10,000
	AECommissionRateNewLarge = 0.12 // Projects over $50,000

	// AECommissionRates for Existing Business
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

	EmploymentStatusActive     EmploymentStatus = "EMPLOYMENT_STATUS_ACTIVE"
	EmploymentStatusInactive   EmploymentStatus = "EMPLOYMENT_STATUS_INACTIVE"
	EmploymentStatusTerminated EmploymentStatus = "EMPLOYMENT_STATUS_TERMINATED"

	// Line item types
	LineItemTypeTimesheet  LineItemType = "LINE_ITEM_TYPE_TIMESHEET"
	LineItemTypeSalary     LineItemType = "LINE_ITEM_TYPE_SALARY"
	LineItemTypeCommission LineItemType = "LINE_ITEM_TYPE_COMMISSION"
	LineItemTypeAdjustment LineItemType = "LINE_ITEM_TYPE_ADJUSTMENT"
	LineItemTypeExpense    LineItemType = "LINE_ITEM_TYPE_EXPENSE"

	// Expense states
	ExpenseStateDraft     ExpenseState = "EXPENSE_STATE_DRAFT"
	ExpenseStateSubmitted ExpenseState = "EXPENSE_STATE_SUBMITTED"
	ExpenseStateApproved  ExpenseState = "EXPENSE_STATE_APPROVED"
	ExpenseStateRejected  ExpenseState = "EXPENSE_STATE_REJECTED"
	ExpenseStateInvoiced  ExpenseState = "EXPENSE_STATE_INVOICED"
	ExpenseStatePaid      ExpenseState = "EXPENSE_STATE_PAID"

	CompensationTypeFullyVariable    CompensationType = "COMPENSATION_TYPE_FULLY_VARIABLE"
	CompensationTypeSalaried         CompensationType = "COMPENSATION_TYPE_SALARIED"
	CompensationTypeBasePlusVariable CompensationType = "COMPENSATION_TYPE_BASE_PLUS_VARIABLE"
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
	UserID                  uint         `json:"user_id"`
	User                    User         `json:"user"`
	Title                   string       `json:"title"`
	FirstName               string       `json:"first_name"`
	LastName                string       `json:"last_name"`
	IsActive                bool         `json:"is_active"`
	EmploymentStatus        string       `json:"employment_status"` // "active", "inactive", "terminated"
	StartDate               time.Time    `json:"start_date"`
	EndDate                 time.Time    `json:"end_date"`
	HeadshotAssetID         *uint        `json:"headshot_asset_id"`
	HeadshotAsset           *Asset       `json:"headshot_asset" gorm:"foreignKey:HeadshotAssetID"`
	Entries                 []Entry      `json:"entries"`
	Commissions             []Commission `json:"commissions" gorm:"foreignKey:StaffID"`
	CapacityWeekly          int          `json:"capacity_weekly"`
	CompensationType        string       `json:"compensation_type"` // "fully-variable", "salaried", "base-plus-variable"
	SalaryAnnualized        int          `json:"salary_annualized"` // For salaried or base+variable compensation
	HasVariableInternalRate bool         `json:"is_variable_hourly"`
	HasFixedInternalRate    bool         `json:"is_fixed_hourly"`
	FixedHourlyRate         int          `json:"hourly_rate"`
	EntryPayEligibleState   string       `json:"entry_pay_eligible_state"`
}

// IsOwner returns true if the employee is an owner (has "partner" in their title)
// Owner distributions are not tax-deductible and should be tracked separately from payroll
func (e *Employee) IsOwner() bool {
	return strings.Contains(strings.ToLower(e.Title), "partner")
}

// RecurringEntry represents a template for auto-generating regular payroll entries
// Used for base salary, monthly bonuses, or other fixed compensation
type RecurringEntry struct {
	gorm.Model
	EmployeeID  uint       `json:"employee_id"`
	Employee    Employee   `json:"employee"`
	Type        string     `json:"type"`        // "base_salary", "bonus", "stipend"
	Description string     `json:"description"` // e.g., "Monthly Base Salary"
	Amount      int        `json:"amount"`      // Amount in cents (monthly)
	Frequency   string     `json:"frequency"`   // "monthly", "biweekly", "annual"
	StartDate   time.Time  `json:"start_date"`  // When to start generating
	EndDate     *time.Time `json:"end_date"`    // Optional end date
	IsActive    bool       `json:"is_active"`   // Can be toggled on/off

	// Generation tracking
	LastGeneratedDate *time.Time `json:"last_generated_date"` // Last time we created entries
	LastGeneratedFor  *time.Time `json:"last_generated_for"`  // Which period we generated for
}

// RecurringBillLineItem represents an auto-generated line item for a bill
// Created from RecurringEntry templates, separate from timesheet entries
type RecurringBillLineItem struct {
	gorm.Model
	BillID           uint           `json:"bill_id"`
	Bill             Bill           `json:"bill"`
	RecurringEntryID uint           `json:"recurring_entry_id"`
	RecurringEntry   RecurringEntry `json:"recurring_entry"`
	Description      string         `json:"description"`
	Amount           int            `json:"amount"` // Amount in cents
	PeriodStart      time.Time      `json:"period_start"`
	PeriodEnd        time.Time      `json:"period_end"`
	State            string         `json:"state"` // "pending", "approved", "paid"
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

// GoogleAuth stores OAuth tokens for Google Calendar integration
type GoogleAuth struct {
	gorm.Model
	UserID       uint      `json:"user_id" gorm:"uniqueIndex"`
	User         User      `json:"user"`
	AccessToken  string    `json:"-" gorm:"type:text"`
	RefreshToken string    `json:"-" gorm:"type:text"`
	TokenExpiry  time.Time `json:"token_expiry"`
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
	Assets                []Asset   `json:"assets"`
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
	Name                string               `json:"name"`
	Description         string               `json:"description"`
	AccountID           uint                 `json:"account_id"`
	Account             Account              `json:"account"`
	ActiveStart         time.Time            `json:"active_start"`
	ActiveEnd           time.Time            `json:"active_end"`
	BudgetHours         int                  `json:"budget_hours"`
	BudgetDollars       int                  `json:"budget_dollars"`
	BudgetCapHours      int                  `json:"budget_cap_hours"`
	BudgetCapDollars    int                  `json:"budget_cap_dollars"`
	Internal            bool                 `json:"internal"`
	BillingCodes        []BillingCode        `json:"billing_codes"`
	Entries             []Entry              `json:"entries"`
	Invoices            []Invoice            `json:"invoices"`
	BillingFrequency    string               `json:"billing_frequency"`
	ProjectType         string               `json:"project_type"`
	AEID                *uint                `json:"ae_id"`
	AE                  *Employee            `json:"ae"`
	SDRID               *uint                `json:"sdr_id"`
	SDR                 *Employee            `json:"sdr"`
	StaffingAssignments []StaffingAssignment `json:"staffing_assignments"`
	Assets              []Asset              `json:"assets"`
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
	ProjectID            uint               `json:"project_id"` // Can remove these, unnecessary with billing code
	Project              Project            `json:"project"`    // Can remove these, unnecessary with billing code
	Notes                string             `gorm:"type:varchar(2048)" json:"notes"`
	EmployeeID           uint               `json:"employee_id" gorm:"index:idx_employee_internal"`
	Employee             Employee           `json:"employee"`
	ImpersonateAsUserID  *uint              `json:"impersonate_as_user_id"`
	ImpersonateAsUser    *Employee          `json:"impersonate_as_user" gorm:"foreignKey:ImpersonateAsUserID"`
	BillingCodeID        uint               `json:"billing_code_id"`
	BillingCode          BillingCode        `json:"billing_code"`
	Start                time.Time          `json:"start"`
	End                  time.Time          `json:"end"`
	DurationMinutes      float64            `json:"duration_minutes"` // Auto-calculated: End - Start in minutes
	Internal             bool               `json:"internal" gorm:"index:idx_employee_internal"`
	IsMeeting            bool               `json:"is_meeting" gorm:"default:false"`
	Bill                 Bill               `json:"bill"`
	BillID               *uint              `json:"bill_id"`
	Invoice              Invoice            `json:"invoice"`
	InvoiceID            *uint              `json:"invoice_id"`
	StaffingAssignmentID *uint              `json:"staffing_assignment_id"`
	StaffingAssignment   StaffingAssignment `json:"staffing_assignment"`
	State                string             `json:"state"`
	Fee                  int                `json:"fee"`
}

func (e *Entry) BeforeSave(tx *gorm.DB) (err error) {
	// Auto-calculate duration in minutes
	e.DurationMinutes = e.End.Sub(e.Start).Minutes()

	// Recalculate the fee
	e.Fee = int(e.GetFee(tx) * 100)
	return nil
}

// Invoice is a record that is used to track the status of a billable invoice either as AR/AP.
// An invoice will have a collection of entries that are to be billed to a client as line items. While we use
// the term Invoice, these can mean either an invoice or bill in relationship to Snowpack.
type Invoice struct {
	gorm.Model
	Name             string            `json:"name"`
	AccountID        uint              `json:"account_id"`
	Account          Account           `json:"account"`
	ProjectID        *uint             `json:"project_id"`
	Project          Project           `json:"project"`
	PeriodStart      time.Time         `json:"period_start"`
	PeriodEnd        time.Time         `json:"period_end"`
	Entries          []Entry           `json:"entries"`
	Adjustments      []Adjustment      `json:"adjustments"`
	Expenses         []Expense         `json:"expenses"`
	LineItems        []InvoiceLineItem `json:"line_items"`
	AcceptedAt       time.Time         `json:"accepted_at"`
	SentAt           time.Time         `json:"sent_at"`
	DueAt            time.Time         `json:"due_at"`
	ClosedAt         time.Time         `json:"closed_at"`
	State            string            `json:"state"`
	Type             string            `json:"type"`
	TotalHours       float64           `json:"total_hours"`
	TotalFees        float64           `json:"total_fees"`
	TotalAdjustments float64           `json:"total_adjustments"`
	TotalExpenses    float64           `json:"total_expenses"`
	TotalAmount      float64           `json:"total_amount"`
	JournalID        *uint             `json:"journal_id"`
	GCSFile          string            `json:"file"`
}

// InvoiceLineItem represents a single line item on an invoice or bill
// For invoices: entries are rolled up by billing code
// For bills: separate lines for salary, commission, timesheet, adjustments
type InvoiceLineItem struct {
	gorm.Model
	InvoiceID   uint    `json:"invoice_id"`
	Type        string  `json:"type"` // LineItemType
	Description string  `json:"description"`
	Quantity    float64 `json:"quantity"` // Hours for timesheets
	Rate        float64 `json:"rate"`     // Hourly rate (for display, in dollars)
	Amount      int64   `json:"amount"`   // Total in cents

	// Source references for traceability
	BillingCodeID *uint        `json:"billing_code_id,omitempty"`
	BillingCode   *BillingCode `json:"billing_code,omitempty" gorm:"foreignKey:BillingCodeID"`
	EmployeeID    *uint        `json:"employee_id,omitempty"`
	Employee      *Employee    `json:"employee,omitempty" gorm:"foreignKey:EmployeeID"`
	EntryIDs      string       `json:"entry_ids,omitempty"` // JSON array of entry IDs
	AdjustmentID  *uint        `json:"adjustment_id,omitempty"`
	Adjustment    *Adjustment  `json:"adjustment,omitempty" gorm:"foreignKey:AdjustmentID"`
	CommissionID  *uint        `json:"commission_id,omitempty"`
	Commission    *Commission  `json:"commission,omitempty" gorm:"foreignKey:CommissionID"`
	ExpenseID     *uint        `json:"expense_id,omitempty"`
	Expense       *Expense     `json:"expense,omitempty" gorm:"foreignKey:ExpenseID"`
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
	Name                   string                  `json:"name"`
	State                  BillState               `json:"state"`
	EmployeeID             uint                    `json:"user_id"`
	Employee               Employee                `json:"user"`
	PeriodStart            time.Time               `json:"period_start"`
	PeriodEnd              time.Time               `json:"period_end"`
	Entries                []Entry                 `json:"entries"`
	Adjustments            []Adjustment            `json:"adjustments"`
	Commissions            []Commission            `json:"commissions" gorm:"foreignKey:BillID"`
	LineItems              []BillLineItem          `json:"line_items"`
	RecurringBillLineItems []RecurringBillLineItem `json:"recurring_bill_line_items" gorm:"foreignKey:BillID"`
	AcceptedAt             *time.Time              `json:"accepted_at"`
	ClosedAt               *time.Time              `json:"closed_at"`
	TotalHours             float64                 `json:"total_hours"`
	TotalFees              int                     `json:"total_fees"`
	TotalAdjustments       float64                 `json:"total_adjustments"`
	TotalCommissions       int                     `json:"total_commissions"`
	TotalAmount            int                     `json:"total_amount"`
	GCSFile                string                  `json:"file"`
}

// BillLineItem represents a single line item on a payroll bill
// Separate lines for: salary, commission, timesheet hours, adjustments
type BillLineItem struct {
	gorm.Model
	BillID      uint    `json:"bill_id"`
	Type        string  `json:"type"` // LineItemType
	Description string  `json:"description"`
	Quantity    float64 `json:"quantity"` // Hours for timesheets
	Rate        float64 `json:"rate"`     // Hourly rate (for display, in dollars)
	Amount      int64   `json:"amount"`   // Total in cents

	// Source references for traceability
	BillingCodeID *uint        `json:"billing_code_id,omitempty"`
	BillingCode   *BillingCode `json:"billing_code,omitempty" gorm:"foreignKey:BillingCodeID"`
	EntryIDs      string       `json:"entry_ids,omitempty"` // JSON array of entry IDs
	AdjustmentID  *uint        `json:"adjustment_id,omitempty"`
	Adjustment    *Adjustment  `json:"adjustment,omitempty" gorm:"foreignKey:AdjustmentID"`
	CommissionID  *uint        `json:"commission_id,omitempty"`
	Commission    *Commission  `json:"commission,omitempty" gorm:"foreignKey:CommissionID"`
}

// VerifyJournalBalance checks that all journal entries balance (total debits = total credits)
// Returns the net balance (should be 0) and an error if they don't balance
func (a *App) VerifyJournalBalance() (int64, error) {
	var totalDebits int64
	var totalCredits int64

	if err := a.DB.Raw("SELECT COALESCE(SUM(debit), 0) FROM journals").Scan(&totalDebits).Error; err != nil {
		return 0, fmt.Errorf("failed to sum debits: %w", err)
	}

	if err := a.DB.Raw("SELECT COALESCE(SUM(credit), 0) FROM journals").Scan(&totalCredits).Error; err != nil {
		return 0, fmt.Errorf("failed to sum credits: %w", err)
	}

	balance := totalDebits - totalCredits
	log.Printf("Journal Balance: Debits=$%.2f, Credits=$%.2f, Net=$%.2f",
		float64(totalDebits)/100, float64(totalCredits)/100, float64(balance)/100)

	if balance != 0 {
		return balance, fmt.Errorf("journal entries are unbalanced by $%.2f", float64(balance)/100)
	}

	return 0, nil
}

// Journal refers to a single entry in a journal, this is a single line item that is used to track
// the debits and credits for a specific account.
type Journal struct {
	gorm.Model
	Account                 string                `json:"account"`
	SubAccount              string                `json:"sub_account"`
	Invoice                 Invoice               `json:"invoice"`
	InvoiceID               *uint                 `json:"invoice_id"`
	Bill                    Bill                  `json:"bill"`
	BillID                  *uint                 `json:"bill_id"`
	RecurringBillLineItem   RecurringBillLineItem `json:"recurring_bill_line_item"`
	RecurringBillLineItemID *uint                 `json:"recurring_bill_line_item_id"`
	Memo                    string                `json:"memo"`
	Debit                   int64                 `json:"debit"`
	Credit                  int64                 `json:"credit"`
}

// OfflineJournal represents journal entries imported from external sources (e.g., Beancount)
type OfflineJournal struct {
	gorm.Model
	Date        time.Time `gorm:"index" json:"date"`
	Account     string    `gorm:"index" json:"account"`
	SubAccount  string    `json:"sub_account"`
	Description string    `json:"description"`
	Debit       int64     `json:"debit"`  // in cents
	Credit      int64     `json:"credit"` // in cents

	// Deduplication - SHA256 of date+account+subaccount+description+amounts
	ContentHash string `gorm:"uniqueIndex" json:"content_hash"`
	Source      string `gorm:"default:'beancount'" json:"source"`

	// Review workflow: pending_review, approved, duplicate, excluded, posted
	Status string `gorm:"default:'pending_review';index" json:"status"`

	// Audit trail
	ImportedAt time.Time  `json:"imported_at"`
	ReviewedAt *time.Time `json:"reviewed_at,omitempty"`
	ReviewedBy *uint      `json:"reviewed_by,omitempty"` // Staff ID
	Notes      string     `json:"notes,omitempty"`

	// Reconciliation - link to an expense if this is a payment for an internal expense
	ReconciledExpenseID *uint      `json:"reconciled_expense_id"`
	ReconciledAt        *time.Time `json:"reconciled_at"`
	ReconciledBy        *uint      `json:"reconciled_by"` // Staff ID who reconciled
	ReconciledExpense   *Expense   `json:"reconciled_expense" gorm:"foreignKey:ReconciledExpenseID"`
}

// CommitmentSegment represents a time period with a specific commitment level
type CommitmentSegment struct {
	StartDate  string `json:"start_date"` // Format: "2006-01-02"
	EndDate    string `json:"end_date"`   // Format: "2006-01-02"
	Commitment int    `json:"commitment"` // Weekly hours
}

// CommitmentSchedule represents variable commitment over time
type CommitmentSchedule struct {
	Segments []CommitmentSegment `json:"segments"`
}

type StaffingAssignment struct {
	// StaffingAssignment is a record of an employee's assignment to a project
	gorm.Model
	// This is a many-to-many relationship between employees and projects
	// An employee can be assigned to multiple projects, and a project can have multiple employees
	// assigned to it. This is a join table that links the two together.
	EmployeeID uint     `json:"employee_id"`
	Employee   Employee `json:"employee"`
	ProjectID  uint     `json:"project_id"`
	Project    Project  `json:"project"`

	// Legacy fields - kept for backward compatibility and as defaults
	Commitment int       `json:"commitment"` // Default/fallback weekly commitment
	StartDate  time.Time `json:"start_date"` // Overall assignment start
	EndDate    time.Time `json:"end_date"`   // Overall assignment end

	// New flexible scheduling - JSON field for variable commitments over time
	// If null/empty, falls back to simple Commitment for entire period
	CommitmentSchedule string `json:"commitment_schedule" gorm:"type:text"` // JSON-serialized CommitmentSchedule

	Entries []Entry `json:"entries"`
}

// GetCommitmentForWeek returns the commitment hours for a specific week
// Uses segments if available, otherwise falls back to simple Commitment field
func (sa *StaffingAssignment) GetCommitmentForWeek(weekStart time.Time) int {
	// Try to parse commitment schedule first
	if sa.CommitmentSchedule != "" {
		var schedule CommitmentSchedule
		if err := json.Unmarshal([]byte(sa.CommitmentSchedule), &schedule); err == nil {
			// Find which segment contains this week
			for _, segment := range schedule.Segments {
				segStart, _ := time.Parse("2006-01-02", segment.StartDate)
				segEnd, _ := time.Parse("2006-01-02", segment.EndDate)

				if (weekStart.Equal(segStart) || weekStart.After(segStart)) &&
					(weekStart.Equal(segEnd) || weekStart.Before(segEnd)) {
					return segment.Commitment
				}
			}
		}
	}

	// Fallback to simple commitment
	return sa.Commitment
}

// GetSegments returns the commitment segments, creating a simple one if schedule is empty
func (sa *StaffingAssignment) GetSegments() []CommitmentSegment {
	if sa.CommitmentSchedule != "" {
		var schedule CommitmentSchedule
		if err := json.Unmarshal([]byte(sa.CommitmentSchedule), &schedule); err == nil {
			return schedule.Segments
		}
	}

	// Return a simple single segment from legacy fields
	return []CommitmentSegment{
		{
			StartDate:  sa.StartDate.Format("2006-01-02"),
			EndDate:    sa.EndDate.Format("2006-01-02"),
			Commitment: sa.Commitment,
		},
	}
}

type Asset struct {
	// Asset is a record of an external asset -- these are typically saved to GCS
	// and are associated with a project. These can be images, documents, etc.
	gorm.Model
	// Can optionally be associated with a project or an account
	ProjectID *uint    `json:"project_id"`
	Project   *Project `json:"project"`
	AccountID *uint    `json:"account_id"`
	Account   *Account `json:"account"`
	AssetType string   `json:"asset_type"`
	Name      string   `json:"name"`
	Url       string   `json:"url"`
	IsPublic  bool     `json:"is_public"` // Whether the asset is publicly accessible

	// The following fields are used for GCS and are optional
	BucketName    *string    `json:"bucket_name"`               // GCS bucket name
	ContentType   *string    `json:"content_type"`              // MIME type of the asset
	Size          *int64     `json:"size"`                      // Size of the asset in bytes
	Checksum      *string    `json:"checksum"`                  // Checksum for data integrity
	UploadStatus  *string    `json:"upload_status"`             // Status of the upload process
	UploadedBy    *uint      `json:"uploaded_by"`               // ID of the user who uploaded the asset
	UploadedAt    *time.Time `json:"uploaded_at"`               // Timestamp of the upload
	ExpiresAt     *time.Time `json:"expires_at"`                // Expiration date for the asset
	Version       *int       `json:"version"`                   // Version number of the asset
	GCSObjectPath *string    `json:"gcs_object_path,omitempty"` // Actual GCS object path, e.g., assets/projects/1/file.txt
}

// ChartOfAccount represents a configurable GL account in the system
// This allows dynamic creation of accounts beyond the predefined constants
type ChartOfAccount struct {
	gorm.Model
	AccountCode     string `json:"account_code" gorm:"unique;not null"`                                          // e.g., "OPERATING_EXPENSES_SAAS"
	AccountName     string `json:"account_name"`                                                                 // e.g., "Operating Expenses - SaaS"
	AccountType     string `json:"account_type" gorm:"index:idx_account_type_active,priority:1"`                 // "ASSET", "LIABILITY", "EQUITY", "REVENUE", "EXPENSE"
	ParentID        *uint  `json:"parent_id"`                                                                    // For hierarchical accounts
	IsActive        bool   `json:"is_active" gorm:"default:true;index:idx_account_type_active,priority:2;index"` // Composite index with account_type + individual index
	Description     string `json:"description"`
	IsSystemDefined bool   `json:"is_system_defined" gorm:"default:false"` // True for predefined constants
}

// Subaccount represents a sub-ledger account (e.g., specific vendors, clients, employees)
type Subaccount struct {
	gorm.Model
	Code        string `json:"code" gorm:"not null;uniqueIndex:idx_subaccount_code_account"`                                        // e.g., "AWS", "VANTA_INC", "EMPLOYEE_123"
	Name        string `json:"name"`                                                                                                // e.g., "Amazon Web Services", "Vanta Inc"
	AccountCode string `json:"account_code" gorm:"uniqueIndex:idx_subaccount_code_account;index:idx_subaccount_filters,priority:1"` // Link to ChartOfAccount code, indexed for queries
	Type        string `json:"type" gorm:"index:idx_subaccount_filters,priority:2"`                                                 // "VENDOR", "CLIENT", "EMPLOYEE", "CUSTOM"
	IsActive    bool   `json:"is_active" gorm:"default:true;index:idx_subaccount_filters,priority:3;index"`                         // Composite index for account_code+type+is_active queries
}

// Expense represents a pass-through expense that will be billed to a client
type Expense struct {
	gorm.Model
	ProjectID       *uint     `json:"project_id"` // Nullable - internal expenses don't need a project
	Project         *Project  `json:"project" gorm:"foreignKey:ProjectID"`
	SubmitterID     uint      `json:"submitter_id"`
	Submitter       Employee  `json:"submitter"`
	ApproverID      *uint     `json:"approver_id"`
	Approver        *Employee `json:"approver" gorm:"foreignKey:ApproverID"`
	InvoiceID       *uint     `json:"invoice_id"`
	Invoice         *Invoice  `json:"invoice"`
	Amount          int       `json:"amount"` // Amount in cents
	Date            time.Time `json:"date"`
	Description     string    `json:"description" gorm:"type:varchar(2048)"`
	State           string    `json:"state"` // ExpenseState
	ReceiptID       *uint     `json:"receipt_id"`
	Receipt         *Asset    `json:"receipt" gorm:"foreignKey:ReceiptID"`
	RejectionReason string    `json:"rejection_reason,omitempty"`

	// New fields for flexible GL booking
	ExpenseAccountCode string `json:"expense_account_code"` // Which expense account to debit (e.g., "OPERATING_EXPENSES_SAAS")
	SubaccountCode     string `json:"subaccount_code"`      // Which subaccount to use (e.g., "AWS", vendor name)
	PaymentAccountCode string `json:"payment_account_code"` // Which account was used to pay (e.g., "CASH", "CREDIT_CARD_CHASE")

	// Category and Tags
	CategoryID uint            `json:"category_id"` // Required category
	Category   ExpenseCategory `json:"category"`
	Tags       []ExpenseTag    `json:"tags" gorm:"many2many:expense_tag_assignments;joinForeignKey:expense_id;joinReferences:expense_tag_id"`

	// Reconciliation - link to actual bank/CC transaction
	ReconciledOfflineJournalID *uint           `json:"reconciled_offline_journal_id"` // Link to the actual bank/CC transaction
	ReconciledAt               *time.Time      `json:"reconciled_at"`
	ReconciledBy               *uint           `json:"reconciled_by"` // Staff ID who reconciled
	ReconciledOfflineJournal   *OfflineJournal `json:"reconciled_offline_journal" gorm:"foreignKey:ReconciledOfflineJournalID"`
}

// ExpenseCategory represents a required categorization for expenses
type ExpenseCategory struct {
	gorm.Model
	Name          string `json:"name" gorm:"type:varchar(255);uniqueIndex"`
	Description   string `json:"description" gorm:"type:varchar(1024)"`
	GLAccountCode string `json:"gl_account_code" gorm:"type:varchar(100)"` // Maps to Chart of Accounts (e.g., OPERATING_EXPENSES_TRAVEL)
	Active        bool   `json:"active" gorm:"default:true"`
}

// ExpenseTag represents an optional tag for grouping expenses with budget tracking
type ExpenseTag struct {
	gorm.Model
	Name        string    `json:"name" gorm:"type:varchar(255);uniqueIndex"`
	Description string    `json:"description" gorm:"type:varchar(1024)"`
	Active      bool      `json:"active" gorm:"default:true"`
	Budget      *int      `json:"budget"` // Budget in cents, nullable
	Expenses    []Expense `json:"-" gorm:"many2many:expense_tag_assignments;joinForeignKey:expense_tag_id;joinReferences:expense_id"`
}

// ExpenseTagAssignment is the junction table for expense-to-tag many-to-many relationship
type ExpenseTagAssignment struct {
	ExpenseID uint      `gorm:"primaryKey;column:expense_id"`
	TagID     uint      `gorm:"primaryKey;column:expense_tag_id"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

type AssetType string

func (s AssetType) String() string {
	return string(s)
}

const (
	AssetTypePDF          AssetType = "application/pdf"
	AssetTypeDOCX         AssetType = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	AssetTypeXLSX         AssetType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	AssetTypeCSV          AssetType = "text/csv"
	AssetTypePNG          AssetType = "image/png"
	AssetTypeJPEG         AssetType = "image/jpeg"
	AssetTypeGoogleDoc    AssetType = "application/vnd.google-apps.document" // Matches frontend ASSET_TYPES
	AssetTypeGoogleSheet  AssetType = "application/vnd.google-apps.spreadsheet"
	AssetTypeGoogleSlides AssetType = "application/vnd.google-apps.presentation"
	AssetTypeExternalLink AssetType = "text/uri-list"
	AssetTypeGenericFile  AssetType = "file" // Matches frontend ASSET_TYPES
)

type AssetUploadStatus string

func (s AssetUploadStatus) String() string {
	return string(s)
}

const (
	AssetUploadStatusPending    AssetUploadStatus = "pending"
	AssetUploadStatusUploading  AssetUploadStatus = "uploading"
	AssetUploadStatusCompleted  AssetUploadStatus = "completed"
	AssetUploadStatusFailed     AssetUploadStatus = "failed"
	AssetUploadStatusProcessing AssetUploadStatus = "processing"
)

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

// GetEmployeeBillRate determines the appropriate hourly rate for billing an employee
// based on their rate configuration (fixed vs variable) and the billing code
func (a *App) GetEmployeeBillRate(employee *Employee, billingCodeID uint) float64 {
	if employee.HasFixedInternalRate {
		// Use the employee's fixed hourly rate (stored in cents)
		return float64(employee.FixedHourlyRate) / 100.0
	}

	// Use the billing code's internal rate (variable by project/billing code, or default)
	var billingCode BillingCode
	if err := a.DB.Preload("InternalRate").Where("id = ?", billingCodeID).First(&billingCode).Error; err != nil {
		log.Printf("Error loading billing code %d for rate calculation: %v", billingCodeID, err)
		return 0
	}
	return billingCode.InternalRate.Amount
}

// UpdateInvoiceTotals updates the totals for an invoice based on non-voided entries
// associated with the invoice. This saves us from having to recalculate the totals
func (a *App) UpdateInvoiceTotals(i *Invoice) {
	var totalHours float64
	var totalFeesInt int
	var totalAdjustments float64
	var totalExpensesInt int
	var entries []Entry
	a.DB.Where("invoice_id = ?", i.ID).Find(&entries)
	var adjustments []Adjustment
	a.DB.Where("invoice_id = ?", i.ID).Find(&adjustments)
	var expenses []Expense
	a.DB.Where("invoice_id = ? AND state = ?", i.ID, ExpenseStateInvoiced.String()).Find(&expenses)

	for _, entry := range entries {
		if entry.State != EntryStateVoid.String() {
			totalHours += entry.Duration().Hours()
			totalFeesInt += entry.Fee
		}
	}
	var multiplier float64
	for _, adjustment := range adjustments {
		if adjustment.State != AdjustmentStateVoid.String() {
			// Always use absolute value, then apply sign based on type
			absAmount := math.Abs(adjustment.Amount)
			if adjustment.Type == AdjustmentTypeCredit.String() {
				multiplier = -1.0
			} else {
				multiplier = 1.0
			}
			totalAdjustments += absAmount * multiplier
		}
	}
	for _, expense := range expenses {
		totalExpensesInt += expense.Amount
	}
	i.TotalHours = totalHours
	i.TotalFees = float64(totalFeesInt) / 100.0
	i.TotalAdjustments = totalAdjustments
	i.TotalExpenses = float64(totalExpensesInt) / 100.0
	i.TotalAmount = i.TotalFees + i.TotalAdjustments + i.TotalExpenses
	a.DB.Omit(clause.Associations).Save(&i)
}

// GenerateInvoiceLineItems creates line items for an invoice
// Entries are rolled up by billing code, adjustments are separate line items
func (a *App) GenerateInvoiceLineItems(invoice *Invoice) error {
	// Delete existing line items
	a.DB.Where("invoice_id = ?", invoice.ID).Delete(&InvoiceLineItem{})

	// Load entries with billing codes
	// Exclude VOID (reversed), REJECTED (never approved), and EXCLUDED (approved but not billed to client)
	var entries []Entry
	if err := a.DB.Preload("BillingCode").Preload("BillingCode.Rate").
		Where("invoice_id = ? AND state NOT IN ?", invoice.ID, []string{
			EntryStateVoid.String(),
			EntryStateRejected.String(),
			EntryStateExcluded.String(),
		}).
		Find(&entries).Error; err != nil {
		return fmt.Errorf("failed to load entries: %w", err)
	}

	// Group entries by billing code
	entriesByBillingCode := make(map[uint][]Entry)
	for _, entry := range entries {
		entriesByBillingCode[entry.BillingCodeID] = append(entriesByBillingCode[entry.BillingCodeID], entry)
	}

	// Create line items for each billing code
	for billingCodeID, bcEntries := range entriesByBillingCode {
		var totalHours float64
		var totalAmount int64
		var entryIDs []uint
		var billingCode BillingCode
		var rate float64

		for _, entry := range bcEntries {
			totalHours += entry.Duration().Hours()
			totalAmount += int64(entry.Fee)
			entryIDs = append(entryIDs, entry.ID)
			billingCode = entry.BillingCode
			if billingCode.Rate.Amount > 0 {
				rate = billingCode.Rate.Amount
			}
		}

		// Marshal entry IDs to JSON
		entryIDsJSON, _ := json.Marshal(entryIDs)

		description := fmt.Sprintf("%s - %.1f hours", billingCode.Name, totalHours)

		lineItem := InvoiceLineItem{
			InvoiceID:     invoice.ID,
			Type:          LineItemTypeTimesheet.String(),
			Description:   description,
			Quantity:      totalHours,
			Rate:          rate,
			Amount:        totalAmount,
			BillingCodeID: &billingCodeID,
			EntryIDs:      string(entryIDsJSON),
		}

		if err := a.DB.Create(&lineItem).Error; err != nil {
			return fmt.Errorf("failed to create timesheet line item: %w", err)
		}
	}

	// Create line items for adjustments
	var adjustments []Adjustment
	if err := a.DB.Where("invoice_id = ? AND state != ?", invoice.ID, AdjustmentStateVoid.String()).
		Find(&adjustments).Error; err != nil {
		return fmt.Errorf("failed to load adjustments: %w", err)
	}

	for _, adjustment := range adjustments {
		amount := int64(math.Abs(adjustment.Amount) * 100) // Convert to cents
		if adjustment.Type == AdjustmentTypeCredit.String() {
			amount = -amount // Credits are negative
		}

		adjType := "Fee"
		if adjustment.Type == AdjustmentTypeCredit.String() {
			adjType = "Credit"
		}
		description := fmt.Sprintf("%s: %s", adjType, adjustment.Notes)

		lineItem := InvoiceLineItem{
			InvoiceID:    invoice.ID,
			Type:         LineItemTypeAdjustment.String(),
			Description:  description,
			Quantity:     0,
			Rate:         0,
			Amount:       amount,
			AdjustmentID: &adjustment.ID,
		}

		if err := a.DB.Create(&lineItem).Error; err != nil {
			return fmt.Errorf("failed to create adjustment line item: %w", err)
		}
	}

	// Create line items for expenses
	var expenses []Expense
	if err := a.DB.Where("invoice_id = ? AND state = ?", invoice.ID, ExpenseStateInvoiced.String()).
		Find(&expenses).Error; err != nil {
		return fmt.Errorf("failed to load expenses: %w", err)
	}

	for _, expense := range expenses {
		description := fmt.Sprintf("Expense: %s", expense.Description)

		lineItem := InvoiceLineItem{
			InvoiceID:   invoice.ID,
			Type:        LineItemTypeExpense.String(),
			Description: description,
			Quantity:    0,
			Rate:        0,
			Amount:      int64(expense.Amount),
			ExpenseID:   &expense.ID,
		}

		if err := a.DB.Create(&lineItem).Error; err != nil {
			return fmt.Errorf("failed to create expense line item: %w", err)
		}
	}

	return nil
}

// GenerateBillLineItems creates line items for a bill
// Separate lines for: timesheet (grouped by billing code), commission, adjustments
func (a *App) GenerateBillLineItems(bill *Bill) error {
	// Delete existing line items
	a.DB.Where("bill_id = ?", bill.ID).Delete(&BillLineItem{})

	// Load employee
	var employee Employee
	if err := a.DB.Where("id = ?", bill.EmployeeID).First(&employee).Error; err != nil {
		return fmt.Errorf("failed to load employee: %w", err)
	}

	// Group timesheet entries by billing code
	// Exclude VOID (reversed) and REJECTED (never approved, not paid)
	// EXCLUDED entries are included (staff gets paid for excluded work, just not billed to client)
	var entries []Entry
	if err := a.DB.Preload("BillingCode").Preload("BillingCode.InternalRate").
		Where("bill_id = ? AND state NOT IN ?", bill.ID, []string{
			EntryStateVoid.String(),
			EntryStateRejected.String(),
		}).
		Find(&entries).Error; err != nil {
		return fmt.Errorf("failed to load entries: %w", err)
	}

	entriesByBillingCode := make(map[uint][]Entry)
	for _, entry := range entries {
		entriesByBillingCode[entry.BillingCodeID] = append(entriesByBillingCode[entry.BillingCodeID], entry)
	}

	// Create line items for each billing code
	for billingCodeID, bcEntries := range entriesByBillingCode {
		var totalHours float64
		var totalAmount int64
		var entryIDs []uint
		var billingCode BillingCode
		var rate float64
		var earliestDate, latestDate time.Time

		for i, entry := range bcEntries {
			hours := entry.Duration().Hours()
			totalHours += hours

			// Calculate internal cost
			internalRate := a.GetEmployeeBillRate(&employee, entry.BillingCodeID)
			totalAmount += int64(internalRate * hours * 100)

			entryIDs = append(entryIDs, entry.ID)
			billingCode = entry.BillingCode
			rate = internalRate

			// Track date range
			if i == 0 || entry.Start.Before(earliestDate) {
				earliestDate = entry.Start
			}
			if i == 0 || entry.Start.After(latestDate) {
				latestDate = entry.Start
			}
		}

		entryIDsJSON, _ := json.Marshal(entryIDs)

		// Include entry count and date range in description for better traceability
		description := fmt.Sprintf("%s - %.1f hours (%d entries: %s - %s)",
			billingCode.Name,
			totalHours,
			len(entryIDs),
			earliestDate.Format("01/02"),
			latestDate.Format("01/02"))

		lineItem := BillLineItem{
			BillID:        bill.ID,
			Type:          LineItemTypeTimesheet.String(),
			Description:   description,
			Quantity:      totalHours,
			Rate:          rate,
			Amount:        totalAmount,
			BillingCodeID: &billingCodeID,
			EntryIDs:      string(entryIDsJSON),
		}

		if err := a.DB.Create(&lineItem).Error; err != nil {
			return fmt.Errorf("failed to create timesheet line item: %w", err)
		}
	}

	// Create line items for commissions
	var commissions []Commission
	if err := a.DB.Where("bill_id = ?", bill.ID).Find(&commissions).Error; err != nil {
		return fmt.Errorf("failed to load commissions: %w", err)
	}

	for _, commission := range commissions {
		description := fmt.Sprintf("Commission - %s", commission.Role)
		if commission.ProjectName != "" {
			description = fmt.Sprintf("Commission - %s (%s)", commission.Role, commission.ProjectName)
		}

		lineItem := BillLineItem{
			BillID:       bill.ID,
			Type:         LineItemTypeCommission.String(),
			Description:  description,
			Quantity:     0,
			Rate:         0,
			Amount:       int64(commission.Amount),
			CommissionID: &commission.ID,
		}

		if err := a.DB.Create(&lineItem).Error; err != nil {
			return fmt.Errorf("failed to create commission line item: %w", err)
		}
	}

	// Create line items for adjustments
	var adjustments []Adjustment
	if err := a.DB.Where("bill_id = ? AND state != ?", bill.ID, AdjustmentStateVoid.String()).
		Find(&adjustments).Error; err != nil {
		return fmt.Errorf("failed to load adjustments: %w", err)
	}

	for _, adjustment := range adjustments {
		amount := int64(math.Abs(adjustment.Amount) * 100)
		if adjustment.Type == AdjustmentTypeCredit.String() {
			amount = -amount
		}

		adjType := "Fee"
		if adjustment.Type == AdjustmentTypeCredit.String() {
			adjType = "Credit"
		}
		description := fmt.Sprintf("%s: %s", adjType, adjustment.Notes)

		lineItem := BillLineItem{
			BillID:       bill.ID,
			Type:         LineItemTypeAdjustment.String(),
			Description:  description,
			Quantity:     0,
			Rate:         0,
			Amount:       amount,
			AdjustmentID: &adjustment.ID,
		}

		if err := a.DB.Create(&lineItem).Error; err != nil {
			return fmt.Errorf("failed to create adjustment line item: %w", err)
		}
	}

	return nil
}

// InvoiceLineItemDisplay is used for PDF generation and display purposes
type InvoiceLineItemDisplay struct {
	BillingCode    string  `json:"billing_code"`
	Project        string  `json:"project"`
	ProjectName    string  `json:"project_name"`
	Hours          float64 `json:"hours"`
	HoursFormatted string  `json:"hours_formatted"`
	Rate           float64 `json:"rate"`
	RateFormatted  string  `json:"rate_formatted"`
	Total          float64 `json:"total"`
}

// BillLineItemDisplay is used for PDF generation and display purposes
type BillLineItemDisplay struct {
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

func (a *App) GetInvoiceLineItems(i *Invoice) []InvoiceLineItemDisplay {
	// Load line items from database (created at invoice approval)
	var dbLineItems []InvoiceLineItem
	a.DB.Preload("BillingCode").Preload("BillingCode.Rate").
		Where("invoice_id = ? AND type = ?", i.ID, LineItemTypeTimesheet.String()).
		Find(&dbLineItems)

	// Convert database line items to display format
	var displayLineItems []InvoiceLineItemDisplay
	for _, lineItem := range dbLineItems {
		billingCodeCode := ""
		billingCodeName := ""
		if lineItem.BillingCode != nil {
			billingCodeCode = lineItem.BillingCode.Code
			billingCodeName = lineItem.BillingCode.Name
		}

		displayLineItems = append(displayLineItems, InvoiceLineItemDisplay{
			BillingCode:    billingCodeCode,
			Project:        billingCodeName,
			ProjectName:    "", // Not stored on line item, but not critical for display
			Hours:          lineItem.Quantity,
			HoursFormatted: fmt.Sprintf("%.2f", lineItem.Quantity),
			Rate:           lineItem.Rate,
			RateFormatted:  fmt.Sprintf("%.2f", lineItem.Rate),
			Total:          float64(lineItem.Amount) / 100.0, // Convert from cents
		})
	}

	return displayLineItems
}

func (a *App) GetInvoiceEntries(i *Invoice) []invoiceEntry {
	var invoiceEntries []invoiceEntry
	var entries []Entry
	a.DB.Preload("BillingCode").Preload("Employee").Preload("ImpersonateAsUser").Where("invoice_id = ? and state != ?", i.ID, EntryStateVoid.String()).Order("start asc").Find(&entries)
	for _, entry := range entries {
		staffName := entry.Employee.FirstName + " " + entry.Employee.LastName
		// Use impersonated user name if available
		if entry.ImpersonateAsUser != nil {
			staffName = entry.ImpersonateAsUser.FirstName + " " + entry.ImpersonateAsUser.LastName
		}

		invoiceEntries = append(invoiceEntries, invoiceEntry{
			dateString:     entry.Start.Format("01/02/2006"),
			billingCode:    entry.BillingCode.Code,
			staff:          staffName,
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
	EntryID             uint      `json:"entry_id"`
	ProjectID           uint      `json:"project_id"`
	BillingCodeID       uint      `json:"billing_code_id"`
	BillingCode         string    `json:"billing_code"`
	BillingCodeName     string    `json:"billing_code_name"`
	Start               time.Time `json:"start"`
	End                 time.Time `json:"end"`
	Notes               string    `json:"notes"`
	StartDate           string    `json:"start_date"`
	StartHour           int       `json:"start_hour"`
	StartMinute         int       `json:"start_minute"`
	EndDate             string    `json:"end_date"`
	EndHour             int       `json:"end_hour"`
	EndMinute           int       `json:"end_minute"`
	DurationHours       float64   `json:"duration_hours"`
	StartDayOfWeek      string    `json:"start_day_of_week"`
	StartIndex          float64   `json:"start_index"`
	State               string    `json:"state"`
	Fee                 float64   `json:"fee"`
	ImpersonateAsUserID *uint     `json:"impersonate_as_user_id,omitempty"`
	EmployeeName        string    `json:"employee_name,omitempty"`
	IsBeingImpersonated bool      `json:"is_being_impersonated,omitempty"`
	IsMeeting           bool      `json:"is_meeting"`
}

func (e *Entry) GetAPIEntry() ApiEntry {
	// Calculate duration and extract time components
	durationHours := float64(e.End.Sub(e.Start).Minutes()) / 60.0
	startHour := e.Start.In(time.UTC).Hour()
	startMinute := e.Start.Minute()
	endHour := e.End.In(time.UTC).Hour()
	endMinute := e.End.Minute()

	// Set default values for billing code information
	billingCode := ""
	billingCodeName := ""

	// Only attempt to access billing code information if it exists
	if e.BillingCode.ID != 0 {
		billingCode = e.BillingCode.Code
		billingCodeName = e.BillingCode.Name
	}

	return ApiEntry{
		EntryID:             e.ID,
		ProjectID:           e.ProjectID,
		BillingCodeID:       e.BillingCodeID,
		BillingCode:         billingCode,
		BillingCodeName:     billingCodeName,
		Start:               e.Start.In(time.UTC),
		End:                 e.End.In(time.UTC),
		Notes:               e.Notes,
		StartDate:           e.Start.In(time.UTC).Format("2006-01-02"),
		StartHour:           startHour,
		StartMinute:         startMinute,
		EndDate:             e.End.In(time.UTC).Format("2006-01-02"),
		EndHour:             endHour,
		EndMinute:           endMinute,
		DurationHours:       durationHours,
		StartDayOfWeek:      e.Start.In(time.UTC).Format("Monday"),
		StartIndex:          float64(e.Start.In(time.UTC).Hour()*60+e.Start.In(time.UTC).Minute()) / 60.0,
		State:               e.State,
		Fee:                 float64(e.Fee) / 100.0,
		ImpersonateAsUserID: e.ImpersonateAsUserID,
		EmployeeName:        e.Employee.FirstName + " " + e.Employee.LastName,
		IsMeeting:           e.IsMeeting,
	}
}

type DraftEntry struct {
	EntryID               uint    `json:"entry_id"`
	ProjectID             uint    `json:"project_id"`
	BillingCodeID         uint    `json:"billing_code_id"`
	BillingCode           string  `json:"billing_code"`
	Notes                 string  `json:"notes"`
	StartDate             string  `json:"start_date"`
	DurationHours         float64 `json:"duration_hours"`
	Fee                   float64 `json:"fee"`
	EmployeeName          string  `json:"user_name"`
	EmployeeRole          string  `json:"user_role"`
	ImpersonateAsUserID   *uint   `json:"impersonate_as_user_id"`
	ImpersonateAsUserName string  `json:"impersonate_as_user_name,omitempty"`
	IsImpersonated        bool    `json:"is_impersonated,omitempty"`
	CreatedByName         string  `json:"created_by_name,omitempty"`
	State                 string  `json:"state"`
}

func (a *App) GetDraftEntry(e *Entry) DraftEntry {
	var employeeName string
	var employeeRole string
	var impersonateAsUserName string
	var billingCodeCode string
	var createdByName string

	// Use preloaded employee data if available, otherwise query (for backward compatibility)
	if e.Employee.ID != 0 && e.Employee.FirstName != "" {
		employeeName = e.Employee.FirstName + " " + e.Employee.LastName
		if e.Employee.User.ID != 0 && e.Employee.User.Role != "" {
			employeeRole = e.Employee.User.Role
		} else {
			a.DB.Raw("SELECT role FROM users JOIN employees ON users.id = employees.user_id WHERE employees.id = ?", e.EmployeeID).Scan(&employeeRole)
		}
	} else {
		a.DB.Raw("SELECT concat(first_name, ' ', last_name) FROM employees WHERE id = ?", e.EmployeeID).Scan(&employeeName)
		a.DB.Raw("SELECT role FROM users JOIN employees ON users.id = employees.user_id WHERE employees.id = ?", e.EmployeeID).Scan(&employeeRole)
	}

	// Store creator's name for all entries
	createdByName = employeeName

	// Get impersonated user name if applicable
	if e.ImpersonateAsUserID != nil {
		if e.ImpersonateAsUser.ID != 0 && e.ImpersonateAsUser.FirstName != "" {
			impersonateAsUserName = e.ImpersonateAsUser.FirstName + " " + e.ImpersonateAsUser.LastName
		} else {
			a.DB.Raw("SELECT concat(first_name, ' ', last_name) FROM employees WHERE id = ?", *e.ImpersonateAsUserID).Scan(&impersonateAsUserName)
		}
		// When displaying an impersonated entry, show the impersonated user's name as the primary name
		employeeName = impersonateAsUserName
	}

	// Use preloaded billing code if available, otherwise query (for backward compatibility)
	if e.BillingCode.Code != "" {
		billingCodeCode = e.BillingCode.Code
	} else {
		a.DB.Raw("SELECT code FROM billing_codes WHERE id = ?", e.BillingCodeID).Scan(&billingCodeCode)
	}

	return DraftEntry{
		EntryID:               e.ID,
		ProjectID:             e.ProjectID,
		BillingCodeID:         e.BillingCodeID,
		BillingCode:           billingCodeCode,
		Notes:                 e.Notes,
		StartDate:             e.Start.Format("01/02/2006"),
		DurationHours:         e.Duration().Hours(),
		Fee:                   float64(e.Fee) / 100.0,
		EmployeeName:          employeeName,
		EmployeeRole:          employeeRole,
		ImpersonateAsUserID:   e.ImpersonateAsUserID,
		ImpersonateAsUserName: impersonateAsUserName,
		IsImpersonated:        e.ImpersonateAsUserID != nil,
		CreatedByName:         createdByName,
		State:                 e.State,
	}
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
	Expenses         []Expense    `json:"expenses"`
	Adjustments      []Adjustment `json:"adjustments"`
	TotalHours       float64      `json:"total_hours"`
	TotalFees        float64      `json:"total_fees"`
	TotalExpenses    float64      `json:"total_expenses"`
	TotalAdjustments float64      `json:"total_adjustments"`
	TotalAmount      float64      `json:"total_amount"`
	PeriodClosed     bool         `json:"period_closed"`
}

type AcceptedInvoice struct {
	InvoiceID      uint                     `json:"ID"`
	InvoiceName    string                   `json:"invoice_name"`
	AccountID      uint                     `json:"account_id"`
	AccountName    string                   `json:"account_name"`
	ProjectID      uint                     `json:"project_id"`
	ProjectName    string                   `json:"project_name"`
	PeriodStart    string                   `json:"period_start"`
	PeriodEnd      string                   `json:"period_end"`
	File           string                   `json:"file"`
	LineItemsCount int                      `json:"line_items_count"`
	TotalHours     float64                  `json:"total_hours"`
	TotalFees      float64                  `json:"total_fees"`
	State          string                   `json:"state"`
	SentAt         string                   `json:"sent_at"`
	DueAt          string                   `json:"due_at"`
	ClosedAt       string                   `json:"closed_at"`
	LineItems      []InvoiceLineItemDisplay `json:"line_items"`
}

func (a *App) GetDraftInvoice(i *Invoice) DraftInvoice {
	var accountName, projectName string
	var periodClosed bool

	// Use preloaded account name if available, otherwise query
	if i.Account.ID != 0 {
		accountName = i.Account.Name
	} else if i.AccountID != 0 {
		a.DB.Raw("SELECT name FROM accounts WHERE id = ?", i.AccountID).Scan(&accountName)
	}

	// Use preloaded project name if available, otherwise query
	if i.Project.ID != 0 {
		projectName = i.Project.Name
	} else if i.ProjectID != nil && *i.ProjectID != 0 {
		a.DB.Raw("SELECT name FROM projects WHERE id = ?", *i.ProjectID).Scan(&projectName)
	}

	periodClosed = i.ClosedAt.Before(time.Now())

	// Generate line items in bulk using preloaded data
	draftEntries := make([]DraftEntry, 0, len(i.Entries))
	for _, entry := range i.Entries {
		// Build draft entry directly using preloaded data
		employeeName := entry.Employee.FirstName + " " + entry.Employee.LastName
		employeeRole := entry.Employee.User.Role
		createdByName := employeeName
		var impersonateAsUserName string

		// Handle impersonation
		if entry.ImpersonateAsUserID != nil {
			impersonateAsUserName = entry.ImpersonateAsUser.FirstName + " " + entry.ImpersonateAsUser.LastName
			employeeName = impersonateAsUserName // Show impersonated user's name as primary
		}

		// Calculate duration
		durationHours := float64(entry.End.Sub(entry.Start).Minutes()) / 60.0

		draftEntry := DraftEntry{
			EntryID:               entry.ID,
			ProjectID:             entry.ProjectID,
			BillingCodeID:         entry.BillingCodeID,
			BillingCode:           entry.BillingCode.Code,
			Notes:                 entry.Notes,
			StartDate:             entry.Start.In(time.UTC).Format("01/02/2006"),
			DurationHours:         durationHours,
			State:                 entry.State,
			Fee:                   float64(entry.Fee) / 100.0,
			EmployeeName:          employeeName,
			EmployeeRole:          employeeRole,
			ImpersonateAsUserID:   entry.ImpersonateAsUserID,
			ImpersonateAsUserName: impersonateAsUserName,
			IsImpersonated:        entry.ImpersonateAsUserID != nil,
			CreatedByName:         createdByName,
		}

		draftEntries = append(draftEntries, draftEntry)
	}

	// Calculate totals
	var totalHours, totalFees, totalAmount float64
	var totalAdjustments float64

	// Sum entry fees and hours
	for _, entry := range draftEntries {
		if entry.State != EntryStateVoid.String() {
			totalHours += entry.DurationHours
			totalFees += entry.Fee
		}
	}

	// Use preloaded adjustments if available, otherwise query
	adjustments := i.Adjustments
	if len(adjustments) == 0 {
		a.DB.Where("invoice_id = ?", i.ID).Find(&adjustments)
	}

	// Sum adjustments
	for _, adjustment := range adjustments {
		if adjustment.State != AdjustmentStateVoid.String() {
			totalAdjustments += adjustment.Amount
		}
	}

	// Load approved expenses for this invoice with their associations
	var expenses []Expense
	a.DB.Where("invoice_id = ? AND state = ?", i.ID, ExpenseStateApproved).
		Preload("Project").
		Preload("Submitter").
		Preload("Receipt").
		Preload("Category").
		Preload("Tags").
		Find(&expenses)

	// Sum expenses
	var totalExpenses float64
	for _, expense := range expenses {
		totalExpenses += float64(expense.Amount) / 100.0
	}

	totalAmount = totalFees + totalExpenses + totalAdjustments

	return DraftInvoice{
		InvoiceID:        i.ID,
		InvoiceName:      i.Name,
		AccountID:        i.AccountID,
		AccountName:      accountName,
		ProjectID:        uintPtrToUint(i.ProjectID),
		ProjectName:      projectName,
		PeriodStart:      i.PeriodStart.Format("01/02/2006"),
		PeriodEnd:        i.PeriodEnd.Format("01/02/2006"),
		LineItems:        draftEntries,
		Expenses:         expenses,
		Adjustments:      adjustments,
		TotalHours:       totalHours,
		TotalFees:        totalFees,
		TotalExpenses:    totalExpenses,
		TotalAdjustments: totalAdjustments,
		TotalAmount:      totalAmount,
		PeriodClosed:     periodClosed,
	}
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
	log.Printf("Starting GenerateBills for invoice ID: %d", i.ID)

	userBillingCodeMap := make(map[uint]map[uint]float64)
	userEntryMap := make(map[uint][]Entry)
	entriesProcessed := 0
	entriesSkipped := 0
	entriesNotEligible := 0

	// Add up each of the fees for a user and billing code
	// and cache the value, but only for entries that don't already have a bill
	// and whose employee is eligible for pay at the current entry state
	for _, entry := range i.Entries {
		// Skip void entries and entries that already have a bill
		if entry.State == EntryStateVoid.String() {
			entriesSkipped++
			continue
		}

		if entry.BillID != nil && *entry.BillID > 0 {
			log.Printf("Skipping entry ID %d - already associated with bill ID %d", entry.ID, *entry.BillID)
			entriesSkipped++
			continue
		}

		// Load the employee to check their EntryPayEligibleState
		var employee Employee
		if err := a.DB.Where("id = ?", entry.EmployeeID).First(&employee).Error; err != nil {
			log.Printf("Error loading employee ID %d for entry ID %d: %v", entry.EmployeeID, entry.ID, err)
			entriesSkipped++
			continue
		}

		// Check if the employee is eligible for pay at the current entry state
		if employee.EntryPayEligibleState != "" && employee.EntryPayEligibleState != entry.State {
			log.Printf("Skipping entry ID %d - employee %s %s (ID: %d) not eligible for pay at state %s (eligible at: %s)",
				entry.ID, employee.FirstName, employee.LastName, employee.ID, entry.State, employee.EntryPayEligibleState)
			entriesNotEligible++
			continue
		}

		// Note: We always bill to the actual creator (EmployeeID), not the impersonated user
		// Initialize the billing code map for this user if it doesn't exist
		if _, ok := userBillingCodeMap[entry.EmployeeID]; !ok {
			userBillingCodeMap[entry.EmployeeID] = make(map[uint]float64)
			userEntryMap[entry.EmployeeID] = []Entry{}
		}

		// Store the employee data with the entry for later use
		entry.Employee = employee

		// Add the entry's minutes to the map
		userBillingCodeMap[entry.EmployeeID][entry.BillingCodeID] += entry.Duration().Minutes()

		// Keep track of this entry for later association with the bill
		userEntryMap[entry.EmployeeID] = append(userEntryMap[entry.EmployeeID], entry)
		entriesProcessed++
	}

	log.Printf("Processing %d entries, skipped %d entries with existing bill associations, %d entries not eligible for pay", entriesProcessed, entriesSkipped, entriesNotEligible)

	if entriesProcessed == 0 {
		log.Printf("No entries to process for invoice ID: %d, all entries already have bill associations", i.ID)
		return
	}

	// Now we need to iterate over the map and create or add to the bill for each user
	for user, billingCodes := range userBillingCodeMap {
		var userObj Employee
		a.DB.Where("id = ?", user).First(&userObj)
		log.Printf("Processing bill for employee: %s %s (ID: %d)", userObj.FirstName, userObj.LastName, user)

		var hours float64
		var fee int

		// Calculate the total hours and fee for this user
		for billingCode, loopMin := range billingCodes {
			hourAmount := loopMin / 60
			hours += hourAmount

			rate := a.GetEmployeeBillRate(&userObj, billingCode)
			floatFee := hourAmount * rate
			fee += int(floatFee * 100)

			log.Printf("  Billing code %d: %.2f hours at $%.2f/hr = $%.2f", billingCode, hourAmount, rate, floatFee)
		}

		// See if there is an existing bill for the user
		var bill Bill
		var err error
		bill, err = a.GetLatestBillIfExists(user)

		// If there is no bill, then create a new one for the user
		firstOfMonth := time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.UTC)
		lastOfMonth := firstOfMonth.AddDate(0, 1, -1)

		if err != nil && errors.Is(err, NoEligibleBill) {
			// Create a new bill for the user
			log.Printf("Creating new bill for employee ID: %d", user)
			bill = Bill{
				Name:        "Payroll " + userObj.FirstName + " " + userObj.LastName + " " + firstOfMonth.Format("01/02/2006") + " - " + lastOfMonth.Format("01/02/2006"),
				State:       BillStateDraft,
				EmployeeID:  user,
				PeriodStart: firstOfMonth,
				PeriodEnd:   lastOfMonth,
				TotalHours:  0,
				TotalFees:   0,
				TotalAmount: 0,
			}
			if err := a.DB.Create(&bill).Error; err != nil {
				log.Printf("Error creating bill: %v", err)
				continue
			}
			log.Printf("Created new bill ID: %d", bill.ID)
		} else {
			log.Printf("Using existing bill ID: %d", bill.ID)
		}

		// Update the entries to associate with the bill
		entriesUpdated := 0
		for _, entry := range userEntryMap[user] {
			entry.BillID = &bill.ID
			result := a.DB.Save(&entry)
			if result.Error != nil {
				log.Printf("Error associating entry ID %d with bill ID %d: %v", entry.ID, bill.ID, result.Error)
			} else {
				entriesUpdated++
			}
		}
		log.Printf("Associated %d entries with bill ID: %d", entriesUpdated, bill.ID)

		// Recalculate bill totals from all entries in the database
		a.RecalculateBillTotals(&bill)

		// Generate line items for the bill (salary, timesheet by billing code, commission, adjustments)
		log.Printf("Generating line items for bill ID: %d", bill.ID)
		if err := a.GenerateBillLineItems(&bill); err != nil {
			log.Printf("Warning: Failed to generate line items for bill %d: %v", bill.ID, err)
		}

		// Note: Journal entries are NOT booked here. They are booked at invoice approval (as accruals)
		// and then moved to AP when the invoice is sent/paid by calling BookBillAccrual explicitly.
		// This allows bills to be created without prematurely moving accruals to AP or double-booking expenses.

		// Note: SaveBillToGCS is not called here to avoid requiring GCS credentials
		// during testing. Callers should explicitly call SaveBillToGCS if needed.
	}

	log.Printf("Completed GenerateBills for invoice ID: %d", i.ID)
}

// RecalculateBillTotals recalculates the total hours, fees, and amounts for a bill
// based on all associated entries in the database
func (a *App) RecalculateBillTotals(bill *Bill) {
	log.Printf("Recalculating totals for bill ID: %d", bill.ID)

	// Reset totals
	bill.TotalHours = 0
	bill.TotalFees = 0

	// Get all non-void entries for this bill
	var entries []Entry
	if err := a.DB.Preload("BillingCode").Preload("BillingCode.InternalRate").
		Where("bill_id = ? AND state != ?", bill.ID, EntryStateVoid.String()).
		Find(&entries).Error; err != nil {
		log.Printf("Error loading entries for bill ID %d: %v", bill.ID, err)
		return
	}

	log.Printf("Found %d valid entries for bill ID: %d", len(entries), bill.ID)

	// Load the employee to get rate configuration
	var employee Employee
	if err := a.DB.Where("id = ?", bill.EmployeeID).First(&employee).Error; err != nil {
		log.Printf("Error loading employee for bill ID %d: %v", bill.ID, err)
		return
	}

	// Calculate totals from entries
	for _, entry := range entries {
		bill.TotalHours += entry.Duration().Hours()

		// Use the helper function to get the correct rate based on employee configuration
		hourAmount := entry.Duration().Hours()
		rate := a.GetEmployeeBillRate(&employee, entry.BillingCodeID)
		fee := int(hourAmount * rate * 100)
		bill.TotalFees += fee

		log.Printf("  Entry ID %d: %.2f hours at $%.2f/hr, fee: $%.2f", entry.ID, hourAmount, rate, float64(fee)/100)
	}

	// Get all non-void adjustments for this bill
	var adjustments []Adjustment
	if err := a.DB.Where("bill_id = ? AND state != ?", bill.ID, AdjustmentStateVoid.String()).Find(&adjustments).Error; err != nil {
		log.Printf("Error loading adjustments for bill ID %d: %v", bill.ID, err)
	}

	// Calculate total adjustments
	totalAdjustmentsAmount := 0
	for _, adjustment := range adjustments {
		multiplier := 1
		if adjustment.Type == AdjustmentTypeCredit.String() {
			multiplier = -1
		}

		adjustmentAmount := int(adjustment.Amount*100) * multiplier
		totalAdjustmentsAmount += adjustmentAmount

		log.Printf("  Adjustment ID %d: $%.2f", adjustment.ID, float64(adjustmentAmount)/100)
	}

	// Calculate total commissions
	var commissions []Commission
	bill.TotalCommissions = 0
	if err := a.DB.Where("bill_id = ?", bill.ID).Find(&commissions).Error; err != nil {
		log.Printf("Error loading commissions for bill ID %d: %v", bill.ID, err)
	} else {
		for _, commission := range commissions {
			bill.TotalCommissions += commission.Amount
			log.Printf("  Commission ID %d (%s): $%.2f", commission.ID, commission.Role, float64(commission.Amount)/100)
		}
	}

	// Calculate total recurring entries (base salary, etc.)
	var recurringLineItems []RecurringBillLineItem
	totalRecurringAmount := 0
	if err := a.DB.Where("bill_id = ? AND state != ?", bill.ID, "void").Find(&recurringLineItems).Error; err != nil {
		log.Printf("Error loading recurring line items for bill ID %d: %v", bill.ID, err)
	} else {
		for _, item := range recurringLineItems {
			totalRecurringAmount += item.Amount
			log.Printf("  Recurring: %s - $%.2f", item.Description, float64(item.Amount)/100)
		}
	}

	// Update the total amount (fees + commissions + adjustments + recurring)
	bill.TotalAmount = bill.TotalFees + bill.TotalCommissions + int(float64(totalAdjustmentsAmount)) + totalRecurringAmount
	bill.TotalHours = math.Round(bill.TotalHours*100) / 100

	log.Printf("Bill totals recalculated - Hours: %.2f, Fees: $%.2f, Recurring: $%.2f, Total: $%.2f",
		bill.TotalHours, float64(bill.TotalFees)/100, float64(totalRecurringAmount)/100, float64(bill.TotalAmount)/100)

	// Save the bill
	if err := a.DB.Save(&bill).Error; err != nil {
		log.Printf("Error saving bill with updated totals: %v", err)
	} else {
		log.Printf("Successfully saved bill ID %d with updated totals", bill.ID)
	}
}

// MarkBillPaid marks a bill as paid with the specified payment date
// paymentDate is the actual date the payment was made (can be backdated)
func (a *App) MarkBillPaid(b *Bill, paymentDate time.Time) {
	log.Printf("MarkBillPaid called for bill ID: %d, payment date: %s", b.ID, paymentDate.Format("2006-01-02"))

	// Recalculate bill totals first to ensure accurate values
	a.RecalculateBillTotals(b)

	// Batch approve any draft adjustments on this bill
	result := a.DB.Model(&Adjustment{}).Where("bill_id = ? AND state = ?", b.ID, AdjustmentStateDraft.String()).Update("state", AdjustmentStateApproved.String())
	if result.Error == nil && result.RowsAffected > 0 {
		log.Printf("Batch approved %d draft adjustments on bill %d", result.RowsAffected, b.ID)
	}

	// Now mark the bill as paid with the provided payment date
	b.State = BillStatePaid
	b.ClosedAt = &paymentDate

	// Save the updated bill
	if err := a.DB.Save(&b).Error; err != nil {
		log.Printf("Error saving bill as paid: %v", err)
		return
	}

	// Record cash payment and clear accounts payable (includes adjustments)
	log.Printf("Recording cash payment for bill ID: %d on date: %s", b.ID, paymentDate.Format("2006-01-02"))
	if err := a.RecordBillCashPayment(b, paymentDate); err != nil {
		log.Printf("Warning: Failed to record cash payment for bill %d: %v", b.ID, err)
	}

	log.Printf("Bill ID %d marked as paid with accurate totals", b.ID)
}

func (a *App) GetBillLineItems(b *Bill) []BillLineItemDisplay {
	// Load line items from database (created at bill generation)
	var dbLineItems []BillLineItem
	a.DB.Preload("BillingCode").
		Where("bill_id = ? AND type = ?", b.ID, LineItemTypeTimesheet.String()).
		Find(&dbLineItems)

	// Convert database line items to display format
	var displayLineItems []BillLineItemDisplay
	for _, lineItem := range dbLineItems {
		billingCodeName := ""
		billingCodeCode := ""
		if lineItem.BillingCode != nil {
			billingCodeName = lineItem.BillingCode.Name
			billingCodeCode = lineItem.BillingCode.Code
		}

		displayLineItems = append(displayLineItems, BillLineItemDisplay{
			BillingCode:     billingCodeName,
			BillingCodeCode: billingCodeCode,
			Hours:           lineItem.Quantity,
			HoursFormatted:  fmt.Sprintf("%.2f", lineItem.Quantity),
			Rate:            lineItem.Rate,
			RateFormatted:   fmt.Sprintf("%.2f", lineItem.Rate),
			Total:           float64(lineItem.Amount) / 100.0, // Convert from cents
		})
	}

	// Load recurring bill line items (e.g., base salary)
	var recurringLineItems []RecurringBillLineItem
	a.DB.Where("bill_id = ? AND state != ?", b.ID, "void").Find(&recurringLineItems)

	// Add recurring line items to display
	for _, recurringItem := range recurringLineItems {
		displayLineItems = append(displayLineItems, BillLineItemDisplay{
			BillingCode:     recurringItem.Description, // Use description for the "Description" column
			BillingCodeCode: "SALARY",                  // Code column
			Hours:           0,
			HoursFormatted:  "-",
			Rate:            0,
			RateFormatted:   "-",
			Total:           float64(recurringItem.Amount) / 100.0, // Convert from cents
		})
	}

	return displayLineItems
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

	// Calculate commission amount (in cents) - use math.Round to avoid truncation
	commissionAmount := int(math.Round(float64(totalProjectValue) * rate * 100))

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

// Helper function to convert a uint pointer to uint, with 0 as default
func uintPtrToUint(ptr *uint) uint {
	if ptr == nil {
		return 0
	}
	return *ptr
}

const signedURLExpiration = time.Hour * 24 * 7 // 7 days

// UploadObject uploads a file to GCS.
func (a *App) UploadObject(ctx context.Context, bucketName, objectName string, data io.Reader, contentType string) error {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to create GCS client")
	}
	defer client.Close()

	wc := client.Bucket(bucketName).Object(objectName).NewWriter(ctx)
	wc.ContentType = contentType

	if _, err := io.Copy(wc, data); err != nil {
		return errors.Wrap(err, "failed to write object to GCS")
	}

	if err := wc.Close(); err != nil {
		return errors.Wrap(err, "failed to close GCS writer")
	}

	return nil
}

// GetObjectURL retrieves the public URL of an object.
func (a *App) GetObjectURL(bucketName, objectName string) string {
	return fmt.Sprintf("https://storage.googleapis.com/%s/%s", bucketName, objectName)
}

// MakeObjectPublic makes an object publicly accessible.
func (a *App) MakeObjectPublic(ctx context.Context, bucketName, objectName string) error {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to create GCS client")
	}
	defer client.Close()

	acl := client.Bucket(bucketName).Object(objectName).ACL()
	if err := acl.Set(ctx, storage.AllUsers, storage.RoleReader); err != nil {
		return errors.Wrap(err, "failed to set public ACL")
	}

	return nil
}

// DownloadObject downloads an object from GCS.
func (a *App) DownloadObject(ctx context.Context, bucketName, objectName string) ([]byte, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create GCS client")
	}
	defer client.Close()

	rc, err := client.Bucket(bucketName).Object(objectName).NewReader(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open GCS object")
	}
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read GCS object")
	}

	return data, nil
}

// DeleteObject deletes an object from GCS.
func (a *App) DeleteObject(ctx context.Context, bucketName, objectName string) error {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to create GCS client")
	}
	defer client.Close()

	if err := client.Bucket(bucketName).Object(objectName).Delete(ctx); err != nil {
		return errors.Wrap(err, "failed to delete GCS object")
	}

	return nil
}

// ObjectExists checks if an object exists in GCS.
func (a *App) ObjectExists(ctx context.Context, bucketName, objectName string) (bool, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return false, errors.Wrap(err, "failed to create GCS client")
	}
	defer client.Close()

	_, err = client.Bucket(bucketName).Object(objectName).Attrs(ctx)
	if err == storage.ErrObjectNotExist {
		return false, nil
	} else if err != nil {
		return false, errors.Wrap(err, "failed to check GCS object existence")
	}

	return true, nil
}

// GenerateSignedURL generates a signed URL for accessing a private object.
// It now returns the generated URL, the expiration time of the URL, and any error.
// Uses IAM-based signing with service account credentials.
// Requires GOOGLE_APPLICATION_CREDENTIALS pointing to a service account key file.
// Falls back to public URLs if service account credentials are not available.
func (a *App) GenerateSignedURL(bucketName, objectName string) (string, time.Time, error) {
	ctx := context.Background()
	expiresTime := time.Now().Add(signedURLExpiration)

	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Printf("Failed to create storage client for signed URL: %v", err)
		return a.GetObjectURL(bucketName, objectName), expiresTime, err
	}
	defer client.Close()

	// Generate signed URL using IAM SignBytes
	opts := &storage.SignedURLOptions{
		Scheme:  storage.SigningSchemeV4,
		Method:  "GET",
		Expires: expiresTime,
	}

	url, err := client.Bucket(bucketName).SignedURL(objectName, opts)
	if err != nil {
		// If signing fails (e.g., using gcloud user credentials instead of service account),
		// fall back to public URL without logging as error
		log.Printf("Warning: Cannot generate signed URL for %s/%s: %v - falling back to public URL", bucketName, objectName, err)
		return a.GetObjectURL(bucketName, objectName), expiresTime, nil
	}

	log.Printf("Successfully generated signed URL for %s/%s (expires: %v)", bucketName, objectName, expiresTime)
	return url, expiresTime, nil
}

// RefreshAssetURLIfExpired checks if an asset's signed URL is expired and regenerates it if needed.
// Updates the database with the new URL and expiration time.
// Gracefully handles cases where signed URLs cannot be generated (e.g., missing service account credentials).
func (a *App) RefreshAssetURLIfExpired(asset *Asset) error {
	// Skip if asset doesn't have GCS storage info
	if asset.GCSObjectPath == nil || *asset.GCSObjectPath == "" || asset.BucketName == nil || *asset.BucketName == "" {
		return nil
	}

	// Check if URL is expired, missing, or is a public URL (needs to be converted to signed)
	now := time.Now()
	needsRefresh := false

	// Log current state for debugging
	urlPrefix := "unknown"
	if len(asset.Url) > 30 {
		urlPrefix = asset.Url[:30]
	} else if asset.Url != "" {
		urlPrefix = asset.Url
	}

	// Check if URL is a public URL (no query parameters) vs signed URL (has query params)
	isPublicURL := asset.Url != "" && !strings.Contains(asset.Url, "?")

	if asset.ExpiresAt == nil {
		needsRefresh = true
		log.Printf("Asset %d: no expiration (URL: %s...), needs refresh", asset.ID, urlPrefix)
	} else if asset.ExpiresAt.Before(now) {
		needsRefresh = true
		log.Printf("Asset %d: expired at %v (URL: %s...), needs refresh", asset.ID, asset.ExpiresAt, urlPrefix)
	} else if isPublicURL {
		needsRefresh = true
		log.Printf("Asset %d: converting public URL to signed URL (URL: %s...)", asset.ID, urlPrefix)
	} else {
		// Log that it doesn't need refresh
		log.Printf("Asset %d: still valid until %v (URL: %s...)", asset.ID, asset.ExpiresAt, urlPrefix)
	}

	if !needsRefresh {
		return nil
	}

	// Regenerate signed URL (or fall back to public URL if service account not available)
	log.Printf("Asset %d: generating signed URL for %s/%s", asset.ID, *asset.BucketName, *asset.GCSObjectPath)
	newURL, newExpiresAt, err := a.GenerateSignedURL(*asset.BucketName, *asset.GCSObjectPath)
	// GenerateSignedURL now returns nil error even if it falls back to public URL,
	// so we only check for actual errors (client creation failures, etc.)
	if err != nil {
		// Only log as warning, don't fail the entire request
		log.Printf("Warning: failed to refresh URL for asset %d: %v", asset.ID, err)
		return nil
	}

	log.Printf("Asset %d: new URL starts with: %s (length: %d)", asset.ID, newURL[:min(50, len(newURL))], len(newURL))

	// Update asset in memory
	asset.Url = newURL
	asset.ExpiresAt = &newExpiresAt

	// Save to database
	if err := a.DB.Model(asset).Updates(map[string]interface{}{
		"url":        newURL,
		"expires_at": newExpiresAt,
	}).Error; err != nil {
		log.Printf("Warning: failed to save refreshed URL for asset %d: %v", asset.ID, err)
		return nil
	}

	log.Printf("Asset %d: successfully saved new URL to database", asset.ID)
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// RefreshAssetsURLsIfExpired refreshes signed URLs for a slice of assets if expired.
func (a *App) RefreshAssetsURLsIfExpired(assets []Asset) error {
	if len(assets) == 0 {
		return nil
	}

	log.Printf("Checking %d assets for expired URLs...", len(assets))
	refreshed := 0

	for i := range assets {
		if err := a.RefreshAssetURLIfExpired(&assets[i]); err != nil {
			log.Printf("Warning: failed to refresh asset %d: %v", assets[i].ID, err)
			// Continue processing other assets instead of failing completely
		} else if assets[i].ExpiresAt != nil {
			refreshed++
		}
	}

	if refreshed > 0 {
		log.Printf("Refreshed %d/%d asset URLs", refreshed, len(assets))
	}

	return nil
}

// ApproveExpense approves an expense (either client or internal)
func (a *App) ApproveExpense(expenseID uint, approverID uint) error {
	var expense Expense
	if err := a.DB.Preload("Project").Preload("Category").First(&expense, expenseID).Error; err != nil {
		return fmt.Errorf("failed to load expense: %w", err)
	}

	if expense.State != ExpenseStateSubmitted.String() {
		return fmt.Errorf("expense must be in submitted state to approve")
	}

	expense.State = ExpenseStateApproved.String()
	expense.ApproverID = &approverID

	if err := a.DB.Save(&expense).Error; err != nil {
		return fmt.Errorf("failed to save approved expense: %w", err)
	}

	// Check if this is a client expense or internal expense
	if expense.ProjectID != nil {
		// CLIENT EXPENSE: Associate with invoice and book with revenue
		return a.approveClientExpense(&expense, approverID)
	} else {
		// INTERNAL EXPENSE: Book directly without invoice
		return a.approveInternalExpense(&expense, approverID)
	}
}

// approveClientExpense handles approval for client pass-through expenses
func (a *App) approveClientExpense(expense *Expense, approverID uint) error {
	// Get the project and account
	var project Project
	if err := a.DB.Preload("Account").First(&project, expense.ProjectID).Error; err != nil {
		return fmt.Errorf("failed to load project: %w", err)
	}

	// Find or create a draft invoice for this project
	var eligibleInvoices []Invoice
	if project.Account.ProjectsSingleInvoice {
		// Single invoice for all projects
		a.DB.Preload("Account").Where("account_id = ? AND type = ? AND state = ?",
			project.AccountID, InvoiceTypeAR.String(), InvoiceStateDraft.String()).
			Order("period_end desc").Find(&eligibleInvoices)
	} else {
		// Separate invoices per project
		a.DB.Preload("Account").Where("account_id = ? AND project_id = ? AND type = ? AND state = ?",
			project.AccountID, expense.ProjectID, InvoiceTypeAR.String(), InvoiceStateDraft.String()).
			Order("period_end desc").Find(&eligibleInvoices)
	}

	var invoice Invoice
	if len(eligibleInvoices) == 0 {
		// Create a new draft invoice
		var projectIDPtr *uint
		if !project.Account.ProjectsSingleInvoice {
			projectIDPtr = expense.ProjectID // Already a pointer
		}
		if err := a.CreateInvoice(project.AccountID, projectIDPtr, expense.Date); err != nil {
			return fmt.Errorf("failed to create draft invoice: %w", err)
		}

		// Query again for the new invoice
		if project.Account.ProjectsSingleInvoice {
			a.DB.Preload("Account").Where("account_id = ? AND type = ? AND state = ?",
				project.AccountID, InvoiceTypeAR.String(), InvoiceStateDraft.String()).
				Order("period_end desc").First(&invoice)
		} else {
			a.DB.Preload("Account").Where("account_id = ? AND project_id = ? AND type = ? AND state = ?",
				project.AccountID, expense.ProjectID, InvoiceTypeAR.String(), InvoiceStateDraft.String()).
				Order("period_end desc").First(&invoice)
		}
	} else {
		invoice = eligibleInvoices[0]
	}

	// Associate the expense with the invoice
	expense.InvoiceID = &invoice.ID
	if err := a.DB.Save(&expense).Error; err != nil {
		return fmt.Errorf("failed to associate expense with invoice: %w", err)
	}

	// Book the expense to the general ledger (with revenue for client reimbursement)
	if err := a.BookExpenseAccrual(expense, &invoice); err != nil {
		return fmt.Errorf("failed to book expense accrual: %w", err)
	}

	log.Printf("Approved client expense ID %d by approver ID %d and added to invoice ID %d", expense.ID, approverID, invoice.ID)
	return nil
}

// approveInternalExpense handles approval for internal company expenses (no client billing)
func (a *App) approveInternalExpense(expense *Expense, approverID uint) error {
	log.Printf("Approving internal expense ID %d", expense.ID)

	amountCents := int64(expense.Amount)

	// Determine expense category account
	expenseAccount := "OPERATING_EXPENSES_GENERAL" // Default
	if expense.Category.GLAccountCode != "" {
		expenseAccount = expense.Category.GLAccountCode
	} else if expense.ExpenseAccountCode != "" {
		expenseAccount = expense.ExpenseAccountCode
	}

	// Use category name as subaccount if no specific subaccount provided
	expenseSubAccount := expense.Category.Name
	if expense.SubaccountCode != "" {
		expenseSubAccount = expense.SubaccountCode
	}

	// Store the subaccount code for later reconciliation
	expense.SubaccountCode = expenseSubAccount
	if err := a.DB.Save(expense).Error; err != nil {
		return fmt.Errorf("failed to save expense subaccount: %w", err)
	}

	// Book DR: [Expense Category]
	expenseDR := Journal{
		Account:    expenseAccount,
		SubAccount: expenseSubAccount,
		Memo:       fmt.Sprintf("Internal expense: %s", expense.Description),
		Debit:      amountCents,
		Credit:     0,
	}
	if err := a.DB.Create(&expenseDR).Error; err != nil {
		return fmt.Errorf("failed to book internal expense debit: %w", err)
	}

	// Book CR: ACCRUED_EXPENSES_PAYABLE (contra account until reconciled)
	accrualCR := Journal{
		Account:    AccountAccruedExpensesPayable.String(),
		SubAccount: expenseSubAccount,
		Memo:       fmt.Sprintf("Internal expense accrual: %s", expense.Description),
		Debit:      0,
		Credit:     amountCents,
	}
	if err := a.DB.Create(&accrualCR).Error; err != nil {
		return fmt.Errorf("failed to book internal expense accrual credit: %w", err)
	}

	log.Printf("Booked internal expense ID %d: DR %s/%s, CR ACCRUED_EXPENSES_PAYABLE, amount=$%.2f",
		expense.ID, expenseAccount, expenseSubAccount, float64(amountCents)/100)
	return nil
}

// RejectExpense rejects a submitted expense
func (a *App) RejectExpense(expenseID uint, approverID uint, reason string) error {
	var expense Expense
	if err := a.DB.First(&expense, expenseID).Error; err != nil {
		return fmt.Errorf("failed to load expense: %w", err)
	}

	if expense.State != ExpenseStateSubmitted.String() {
		return fmt.Errorf("expense must be in submitted state to reject")
	}

	expense.State = ExpenseStateRejected.String()
	expense.ApproverID = &approverID
	expense.RejectionReason = reason

	if err := a.DB.Save(&expense).Error; err != nil {
		return fmt.Errorf("failed to save rejected expense: %w", err)
	}

	log.Printf("Rejected expense ID %d by approver ID %d: %s", expenseID, approverID, reason)
	return nil
}

// AddExpensesToInvoice associates approved expenses with an invoice
func (a *App) AddExpensesToInvoice(invoiceID uint, expenseIDs []uint) error {
	var invoice Invoice
	if err := a.DB.First(&invoice, invoiceID).Error; err != nil {
		return fmt.Errorf("failed to load invoice: %w", err)
	}

	if invoice.State != InvoiceStateDraft.String() {
		return fmt.Errorf("expenses can only be added to draft invoices")
	}

	// Load and validate expenses
	for _, expenseID := range expenseIDs {
		var expense Expense
		if err := a.DB.Preload("Project").First(&expense, expenseID).Error; err != nil {
			return fmt.Errorf("failed to load expense %d: %w", expenseID, err)
		}

		if expense.State != ExpenseStateApproved.String() {
			return fmt.Errorf("expense %d must be in approved state", expenseID)
		}

		// Verify expense project matches invoice project or account
		if invoice.ProjectID != nil {
			if expense.ProjectID == nil || *expense.ProjectID != *invoice.ProjectID {
				return fmt.Errorf("expense %d project does not match invoice project", expenseID)
			}
		} else {
			if expense.ProjectID != nil && expense.Project != nil && expense.Project.AccountID != invoice.AccountID {
				return fmt.Errorf("expense %d account does not match invoice account", expenseID)
			}
		}

		// Associate expense with invoice
		expense.InvoiceID = &invoiceID
		expense.State = ExpenseStateInvoiced.String()
		if err := a.DB.Save(&expense).Error; err != nil {
			return fmt.Errorf("failed to associate expense %d with invoice: %w", expenseID, err)
		}

		log.Printf("Added expense ID %d to invoice ID %d", expenseID, invoiceID)
	}

	// Update invoice totals
	a.UpdateInvoiceTotals(&invoice)

	return nil
}

// GetTagSpendSummary calculates total approved/invoiced spend and remaining budget for a tag
func (a *App) GetTagSpendSummary(tagID uint) (totalSpent int, budget *int, remaining *int, err error) {
	var tag ExpenseTag
	if err := a.DB.First(&tag, tagID).Error; err != nil {
		return 0, nil, nil, fmt.Errorf("failed to load tag: %w", err)
	}

	// Calculate total spent on approved and invoiced expenses with this tag
	var expenses []Expense
	if err := a.DB.
		Joins("JOIN expense_tag_assignments ON expense_tag_assignments.expense_id = expenses.id").
		Where("expense_tag_assignments.expense_tag_id = ?", tagID).
		Where("expenses.state IN ?", []string{ExpenseStateApproved.String(), ExpenseStateInvoiced.String()}).
		Find(&expenses).Error; err != nil {
		return 0, nil, nil, fmt.Errorf("failed to load expenses: %w", err)
	}

	totalSpent = 0
	for _, expense := range expenses {
		totalSpent += expense.Amount
	}

	if tag.Budget != nil {
		remainingBudget := *tag.Budget - totalSpent
		return totalSpent, tag.Budget, &remainingBudget, nil
	}

	return totalSpent, nil, nil, nil
}
