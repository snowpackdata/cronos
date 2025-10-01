package cronos

import (
	"context"
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

type StaffingAssignment struct {
	// StaffingAssignment is a record of an employee's assignment to a project
	gorm.Model
	// This is a many-to-many relationship between employees and projects
	// An employee can be assigned to multiple projects, and a project can have multiple employees
	// assigned to it. This is a join table that links the two together.
	// The commitment is the weekly commitment of the employee to the project
	EmployeeID uint      `json:"employee_id"`
	Employee   Employee  `json:"employee"`
	ProjectID  uint      `json:"project_id"`
	Project    Project   `json:"project"`
	Commitment int       `json:"commitment"`
	StartDate  time.Time `json:"start_date"`
	EndDate    time.Time `json:"end_date"`
	Entries    []Entry   `json:"entries"`
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
	totalAmount = totalFees + totalAdjustments

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
		Adjustments:      adjustments,
		TotalHours:       totalHours,
		TotalFees:        totalFees,
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

	// Update the total amount
	bill.TotalAmount = bill.TotalFees + bill.TotalCommissions + int(float64(totalAdjustmentsAmount))
	bill.TotalHours = math.Round(bill.TotalHours*100) / 100

	log.Printf("Bill totals recalculated - Hours: %.2f, Fees: $%.2f, Total: $%.2f",
		bill.TotalHours, float64(bill.TotalFees)/100, float64(bill.TotalAmount)/100)

	// Save the bill
	if err := a.DB.Save(&bill).Error; err != nil {
		log.Printf("Error saving bill with updated totals: %v", err)
	} else {
		log.Printf("Successfully saved bill ID %d with updated totals", bill.ID)
	}
}

func (a *App) MarkBillPaid(b *Bill) {
	// Recalculate bill totals first to ensure accurate values
	a.RecalculateBillTotals(b)

	// Now mark the bill as paid
	nowTime := time.Now()
	b.ClosedAt = &nowTime

	// Save the updated bill
	if err := a.DB.Save(&b).Error; err != nil {
		log.Printf("Error saving bill as paid: %v", err)
		return
	}
	log.Printf("Bill ID %d marked as paid with accurate totals", b.ID)

	// Get the User
	var userObj Employee
	a.DB.Where("id = ?", b.EmployeeID).First(&userObj)
	// Now add a journal entry to reflect the bill
	journal := Journal{
		Account:    AccountAPStaffPayroll.String(),
		SubAccount: userObj.FirstName + " " + userObj.LastName,
		Debit:      int64(b.TotalAmount), // Use the total amount which includes fees, commissions, and adjustments
		Credit:     0,
		Memo:       b.Name,
		BillID:     &b.ID,
	}
	if err := a.DB.Create(&journal).Error; err != nil {
		log.Printf("Error creating journal entry for bill: %v", err)
	} else {
		log.Printf("Created journal entry for bill ID %d with amount $%.2f", b.ID, float64(b.TotalAmount)/100)
	}
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
func (a *App) GenerateSignedURL(bucketName, objectName string) (string, time.Time, error) {
	client, err := storage.NewClient(context.Background())
	if err != nil {
		return "", time.Time{}, errors.Wrap(err, "failed to create GCS client")
	}
	defer client.Close()

	expiresTime := time.Now().Add(signedURLExpiration)
	url, err := client.Bucket(bucketName).SignedURL(objectName, &storage.SignedURLOptions{
		Method:  "GET",
		Expires: expiresTime,
	})
	if err != nil {
		return "", time.Time{}, errors.Wrap(err, "failed to generate signed URL")
	}

	return url, expiresTime, nil
}
