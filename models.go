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
	RateTypeInternalNonBillable      RateType = "RATE_TYPE_INTERNAL_CLIENT_BILLABLE"
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
	LinkedEntryID uint        `json:"linked_entry_id"`
	LinkedEntry   *Entry      `json:"-"`
	JournalID     uint        `json:"journal_id"`
	Journal       Journal     `json:"journal"`
}

func (e *Entry) AfterDelete(tx *gorm.DB) (err error) {
	tx.Clauses(clause.Returning{}).Where("linked_entry_id = ? and delete_at is null", e.ID).Update("deleted_at", gorm.Expr("NOW()"))
	return
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
	duration := e.End.Sub(e.Start)
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
	dbURI := fmt.Sprintf("host=127.0.0.1 user=%s password=%s port=%s database=%s sslmode=disable", user, password, port, database)
	db, err := gorm.Open(postgres.Open(dbURI), &gorm.Config{Logger: newLogger})

	if err != nil {
		fmt.Println(err)
	}
	a.DB = db
}

// InitializeCloud allows us to initalize a connection to the cloud database
// while on google app engine
func (a *App) InitializeCloud(user, password, database, connection string) {
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
	socketPath := "/cloudsql/" + connection + "/.s.PGSQL.5432"
	dbURI := fmt.Sprintf("user=%s password=%s database=%s host=%s", user, password, database, socketPath)
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
		Start:          e.Start,
		End:            e.End,
		Notes:          e.Notes,
		StartDate:      e.Start.Format("2006-01-02"),
		StartHour:      e.Start.Hour(),
		StartMinute:    e.Start.Minute(),
		EndDate:        e.End.Format("2006-01-02"),
		EndHour:        e.End.Hour(),
		EndMinute:      e.End.Minute(),
		DurationHours:  e.Duration().Hours(),
		StartDayOfWeek: e.Start.Weekday().String(),
		StartIndex:     float64(e.Start.Hour()) + (float64(e.Start.Minute()) / 60.0),
	}
	return apiEntry
}
