package main

import (
	jwt "github.com/dgrijalva/jwt-go"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
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

type RateType string

func (s RateType) String() string {
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
	RateTypeInternalNonBillable      RateType = "RATE_TYPE_INTERNAL_CLIENT_NON_BILLABLE"
	RateTypeInternalAdminBillable    RateType = "RATE_TYPE_INTERNAL_ADMINISTRATIVE"
	RateTypeInternalAdminNonBillable RateType = "RATE_TYPE_INTERNAL_ADMINISTRATIVE_NON_BILLABLE"
	RateTypeExternalBillable         RateType = "RATE_TYPE_EXTERNAL_CLIENT_BILLABLE"
	RateTypeExternalNonBillable      RateType = "RATE_TYPE_EXTERNAL_CLIENT_NON_BILLABLE"
	RateTypeInternalProject          RateType = "RATE_TYPE_INTERNAL_PROJECT"
)

type User struct {
	// User is the generic user object for anyone accessing the application
	gorm.Model
	Email     string `gorm:"unique" json:"email"`
	Password  string `json:"password"`
	IsAdmin   bool   `json:"is_admin"`
	Role      string `json:"role"`
	AccountID uint   `json:"account_id"`
}

// Token is a non-persistent object that is used to store the JWT token
type Token struct {
	UserID uint
	Email  string
	*jwt.StandardClaims
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
	Admin     User      `json:"admin"`
	AdminID   uint      `json:"admin_id"`
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
	Name          string        `json:"name"`
	AccountID     uint          `json:"account_id"`
	Account       Account       `json:"account"`
	ActiveStart   time.Time     `json:"active_start"`
	ActiveEnd     time.Time     `json:"active_end"`
	BudgetHours   int           `json:"budget_hours"`
	BudgetDollars int           `json:"budget_dollars"`
	Internal      bool          `json:"internal"`
	BillingCodes  []BillingCode `json:"billing_codes"`
	Entries       []Entry       `json:"entries"`
}

type BillingCode struct {
	gorm.Model
	Name        string    `json:"name"`
	RateType    string    `json:"type"`
	Category    string    `json:"category"`
	Code        string    `gorm:"unique" json:"code"`
	RoundedTo   int       `gorm:"default:15" json:"rounded_to"`
	ProjectID   uint      `json:"project"`
	ActiveStart time.Time `json:"active_start"`
	ActiveEnd   time.Time `json:"active_end"`
	Internal    bool      `json:"internal"`
	RateID      uint      `json:"rate_id"`
	Rate        Rate      `json:"rate"`
	Entries     []Entry   `json:"entries"`
}
type Entry struct {
	gorm.Model
	ProjectID     uint        `json:"project_id"`
	Project       Project     `json:"project"`
	EmployeeID    uint        `json:"employee_id"`
	Employee      Employee    `json:"employee"`
	BillingCodeID uint        `json:"billing_code_id"`
	BillingCode   BillingCode `json:"billing_code"`
	Start         time.Time   `json:"start_date"`
	End           time.Time   `json:"end_date"`
	Internal      bool        `json:"internal"`
	LinkedEntryID uint        `json:"linked_entry_id"`
	LinkedEntry   *Entry      `json:"linked_period"`
	JournalID     uint        `json:"journal_id"`
	Journal       Journal     `json:"journal"`
}

// Journal is a preemptive object that we can use to split work across separate accounting Journals,
// this may refer specifically to internal vs external billables in the short term, however we may add
// additional journals in the future. This facilitates simple reporting and accounting at the ledger level.
type Journal struct {
	gorm.Model
	Name    string `json:"name"`
	Entries []*Entry
}

// OBJECT METHODS

// Duration finds the length of an Entry as a duration object
func (e *Entry) Duration() time.Duration {
	duration := e.Start.Sub(e.End)
	return duration
}

// Fee finds the applicable fee in USD for a particular entry rounded to the given minute
func (e *Entry) Fee() float64 {
	durationMinutes := e.Duration().Minutes()
	roundingFactor := float64(e.BillingCode.RoundedTo) / HOUR
	hours := float64(durationMinutes) / HOUR
	roundedHours := float64(int(hours/roundingFactor)) * roundingFactor
	fee := roundedHours * e.BillingCode.Rate.Amount
	return fee
}

// App is used to initialize a database and hold our handler functions
type App struct {
	DB *gorm.DB
}

// Initialize allows us to initialize our application and connect to the database
// This handler will hold on to our database operations throughout the lifetime of the application
func (a *App) Initialize() {
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

// Calling the Migrate
func (a *App) migrate() {
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
}
