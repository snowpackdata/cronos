package cronos

import (
	"fmt"
	_ "github.com/GoogleCloudPlatform/cloudsql-proxy/proxy/dialers/postgres"
	"gorm.io/driver/postgres"
	_ "gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"time"
)

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

	BillingFrequencyMonthly BillingFrequency = "BILLING_TYPE_MONTHLY"
	BillingFrequencyProject BillingFrequency = "BILLING_TYPE_PROJECT"

	EntryStateUnaffiliated EntryState = "ENTRY_STATE_UNAFFILIATED"
	EntryStateDraft        EntryState = "ENTRY_STATE_DRAFT"
	EntryStatePending      EntryState = "ENTRY_STATE_PENDING"
	EntryStateApproved     EntryState = "ENTRY_STATE_APPROVED"
	EntryStatePaid         EntryState = "ENTRY_STATE_PAID"
	EntryStateVoid         EntryState = "ENTRY_STATE_VOID"

	InvoiceStateDraft    InvoiceState = "INVOICE_STATE_DRAFT"
	InvoiceStatePending  InvoiceState = "INVOICE_STATE_PENDING"
	InvoiceStateApproved InvoiceState = "INVOICE_STATE_APPROVED"
	InvoiceStateSent     InvoiceState = "INVOICE_STATE_SENT"
	InvoiceStatePaid     InvoiceState = "INVOICE_STATE_PAID"
	InvoiceStateVoid     InvoiceState = "INVOICE_STATE_VOID"

	InvoiceTypeAR InvoiceType = "INVOICE_TYPE_ACCOUNTS_RECEIVABLE"
	InvoiceTypeAP InvoiceType = "INVOICE_TYPE_ACCOUNTS_PAYABLE"
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
	EmployeeID    uint        `json:"employee_id"`
	Employee      Employee    `json:"employee"`
	BillingCodeID uint        `json:"billing_code_id"`
	BillingCode   BillingCode `json:"billing_code"`
	Start         time.Time   `json:"start"`
	End           time.Time   `json:"end"`
	Internal      bool        `json:"internal"`
	LinkedEntryID *uint       `json:"linked_entry_id"`
	LinkedEntry   *Entry      `json:"-"`
	JournalID     uint        `json:"journal_id"`
	Journal       Journal     `json:"journal"`
	Invoice       Invoice     `json:"invoice"`
	InvoiceID     uint        `json:"invoice_id"`
	State         string      `json:"state"`
}

func (e *Entry) AfterDelete(tx *gorm.DB) (err error) {
	tx.Clauses(clause.Returning{}).Where("linked_entry_id = ? and delete_at is null", e.ID).Update("deleted_at", gorm.Expr("NOW()"))
	return
}

// Invoice is a record that is used to track the status of a billable invoice either as AR/AP.
// An invoice will have a collection of entries that are to be billed to a client as line items. While we use
// the term Invoice, these can mean either an invoice or bill in relationship to Snowpack.
type Invoice struct {
	gorm.Model
	Name        string    `json:"name"`
	ProjectID   uint      `json:"project_id"`
	Project     Project   `json:"project"`
	PeriodStart time.Time `json:"period_start"`
	PeriodEnd   time.Time `json:"period_end"`
	Entries     []Entry   `json:"entries"`
	SentAt      time.Time `json:"sent_at"`
	ClosedAt    time.Time `json:"closed_at"`
	State       string    `json:"state"`
	Type        string    `json:"type"`
}

// Journal is a preemptive object that we can use to split work across separate accounting Journals,
// this may refer specifically to internal vs external billable in the short term, however we may add
// additional journals in the future. This facilitates simple reporting and accounting at the ledger level.
type Journal struct {
	gorm.Model
	Name    string `json:"name"`
	Entries []*Entry
}

// OBJECT METHODS

// Duration finds the length of an Entry as a duration object
func (e *Entry) Duration() time.Duration {
	duration := e.End.Sub(e.Start)
	return duration
}

// Fee finds the applicable fee in USD for a particular entry rounded to the given minute
func (a *App) GetFee(e *Entry) float64 {
	var billingCode BillingCode
	a.DB.Preload("Rate").Where("id = ?", e.BillingCodeID).First(&billingCode)
	durationMinutes := e.Duration().Minutes()
	roundingFactor := float64(billingCode.RoundedTo) / HOUR
	hours := float64(durationMinutes) / HOUR
	roundedHours := float64(int(hours/roundingFactor)) * roundingFactor
	fee := roundedHours * billingCode.Rate.Amount
	return fee
}

// App is used to initialize a database and hold our handler functions
type App struct {
	DB *gorm.DB
}

// InitializeSQLite allows us to initialize our application and connect to the local database
// This handler will hold on to our database operations throughout the lifetime of the application
func (a *App) InitializeSQLite() {
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second, // Slow SQL threshold
			LogLevel:                  logger.Info, // Log level
			IgnoreRecordNotFoundError: true,        // Ignore ErrRecordNotFound error for logger
			ParameterizedQueries:      true,        // Don't include params in the SQL log
			Colorful:                  true,        // Disable color
		},
	)

	db, err := gorm.Open(sqlite.Open("cronos.db"), &gorm.Config{Logger: newLogger})
	if err != nil {
		panic("failed to connect database")
	}
	a.DB = db
}

// InitializeLocal allows us to initialize our application and connect to the cloud database
func (a *App) InitializeLocal(user, password, connection, database string) {
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second, // Slow SQL threshold
			LogLevel:                  logger.Info, // Log level
			IgnoreRecordNotFoundError: true,        // Ignore ErrRecordNotFound error for logger
			ParameterizedQueries:      true,        // Don't include params in the SQL log
			Colorful:                  true,        // Disable color
		},
	)
	port := "3306"
	dbURI := fmt.Sprintf("host=127.0.0.1 user=%s password=%s port=%s database=%s sslmode=disable TimeZone=UTC", user, password, port, database)
	db, err := gorm.Open(postgres.Open(dbURI), &gorm.Config{Logger: newLogger})

	if err != nil {
		fmt.Println(err)
	}
	a.DB = db
}

// InitializeCloud allows us to initialize a connection to the cloud database
// while on google app engine
func (a *App) InitializeCloud(dbURI string) {
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second, // Slow SQL threshold
			LogLevel:                  logger.Info, // Log level
			IgnoreRecordNotFoundError: true,        // Ignore ErrRecordNotFound error for logger
			ParameterizedQueries:      true,        // Don't include params in the SQL log
			Colorful:                  true,        // Disable color
		},
	)
	db, err := gorm.Open(postgres.Open(dbURI), &gorm.Config{
		Logger:                                   newLogger,
		DisableForeignKeyConstraintWhenMigrating: true})

	if err != nil {
		fmt.Println(err)
	}
	a.DB = db
}

// Calling the Migrate
func (a *App) Migrate() {
	// Migrate the schema
	_ = a.DB.AutoMigrate(&User{})
	_ = a.DB.AutoMigrate(&Employee{})
	_ = a.DB.AutoMigrate(&Client{})
	_ = a.DB.AutoMigrate(&Account{})
	_ = a.DB.AutoMigrate(&Rate{})
	_ = a.DB.AutoMigrate(&Project{})
	_ = a.DB.AutoMigrate(&Entry{})
	_ = a.DB.AutoMigrate(&BillingCode{})
	_ = a.DB.AutoMigrate(&Journal{})
	_ = a.DB.AutoMigrate(&Invoice{})
}

type ApiEntry struct {
	EntryID        uint      `json:"entry_id"`
	ProjectID      uint      `json:"project_id"`
	BillingCodeID  uint      `json:"billing_code_id"`
	BillingCode    string    `json:"billing_code"`
	Start          time.Time `json:"start"`
	End            time.Time `json:"end"`
	Notes          string    `json:"notes"`
	StartDate      string    `json:"start_date"`
	StartHour      int       `json:"start_hour"`
	StartMinute    int       `json:"start_minute"`
	EndDate        string    `json:"end_date"`
	EndHour        int       `json:"end_hour"`
	EndMinute      int       `json:"end_minute"`
	DurationHours  float64   `json:"duration_hours"`
	StartDayOfWeek string    `json:"start_day_of_week"`
	StartIndex     float64   `json:"start_index"`
}

func (e *Entry) GetAPIEntry() ApiEntry {
	apiEntry := ApiEntry{
		EntryID:        e.ID,
		ProjectID:      e.ProjectID,
		BillingCodeID:  e.BillingCodeID,
		BillingCode:    e.BillingCode.Code,
		Start:          e.Start.In(time.UTC),
		End:            e.End.In(time.UTC),
		Notes:          e.Notes,
		StartDate:      e.Start.In(time.UTC).Format("2006-01-02"),
		StartHour:      e.Start.In(time.UTC).Hour(),
		StartMinute:    e.Start.Minute(),
		EndDate:        e.End.In(time.UTC).Format("2006-01-02"),
		EndHour:        e.End.In(time.UTC).Hour(),
		EndMinute:      e.End.Minute(),
		DurationHours:  e.Duration().Hours(),
		StartDayOfWeek: e.Start.In(time.UTC).Weekday().String(),
		StartIndex:     float64(e.Start.In(time.UTC).Hour()) + (float64(e.Start.Minute()) / 60.0),
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
		Fee:           a.GetFee(e),
		EmployeeName:  employee.FirstName + " " + employee.LastName,
		EmployeeRole:  employee.Title,
	}
	return draftEntry
}

type DraftInvoice struct {
	InvoiceID   uint         `json:"invoice_id"`
	InvoiceName string       `json:"invoice_name"`
	ProjectID   uint         `json:"project_id"`
	ProjectName string       `json:"project_name"`
	PeriodStart string       `json:"period_start"`
	PeriodEnd   string       `json:"period_end"`
	LineItems   []DraftEntry `json:"line_items"`
	TotalHours  float64      `json:"total_hours"`
	TotalFees   float64      `json:"total_fees"`
}

func (a *App) GetDraftInvoice(i *Invoice) DraftInvoice {
	draftInvoice := DraftInvoice{
		InvoiceID:   i.ID,
		InvoiceName: i.Name,
		ProjectID:   i.ProjectID,
		ProjectName: i.Project.Name,
		PeriodStart: i.PeriodStart.In(time.UTC).Format("01/02/2006"),
		PeriodEnd:   i.PeriodEnd.In(time.UTC).Format("01/02/2006"),
	}
	hourCounter := 0.0
	feeCounter := 0.0
	for _, entry := range i.Entries {
		draftEntry := a.GetDraftEntry(&entry)
		hourCounter += draftEntry.DurationHours
		feeCounter += draftEntry.Fee
		draftInvoice.LineItems = append(draftInvoice.LineItems, draftEntry)
	}
	draftInvoice.TotalHours = hourCounter
	draftInvoice.TotalFees = feeCounter
	return draftInvoice
}
