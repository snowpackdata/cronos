package cronos

import (
	"gorm.io/gorm"
	"time"
)

type User struct {
	// User is the generic user object for anyone accessing the application
	gorm.Model
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	IsAdmin  bool   `json:"is_admin"`
	Role     string `json:"role"`
}

type Employee struct {
	// Employee refers to internal information regarding an employee
	gorm.Model
	User      User   `json:"user"`
	ID        uint   `json:"id"`
	Title     string `json:"title"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type Client struct {
	// Client refers to an external customer that may access the site to see time entries
	gorm.Model
	User      User      `json:"user"`
	ID        uint      `json:"id"`
	Title     string    `json:"title"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Accounts  []Account `gorm:"many2many:user_accounts;"`
}

type Account struct {
	// Account is the specific customer account
	gorm.Model
	ID           uint          `json:"id"`
	Name         string        `json:"name"`
	LegalName    string        `json:"legal_name"`
	Email        string        `json:"email"`
	Website      string        `json:"website"`
	Admin        Client        `json:"admin"`
	BillingCodes []BillingCode `json:"codes"`
	Projects     []Project     `json:"projects"`
	Clients      []*Client     `gorm:"many2many:user_accounts;"`
}

type Rate struct {
	// Rate stores all of available rates that can be added to individual projects
	gorm.Model
	ID         uint      `json:"id"`
	Name       string    `json:"name"`
	Type       string    `json:"type"`
	Amount     float64   `json:"amount"`
	ActiveFrom time.Time `json:"active_from"`
	ActiveTo   time.Time `json:"active_to"`
	Internal   bool      `json:"internal"`
}

type Project struct {
	// Project refers to a single unit of work with a customer
	// often with specific time period. A rate will have a specific billing code
	// associated with the project.
	gorm.Model
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Account     Account   `json:"account"`
	ActiveStart time.Time `json:"active_start"`
	ActiveEnd   time.Time `json:"active_end"`
	Budget      int       `json:"budget"`
	Internal    bool      `json:"internal"`
}

type BillingCode struct {
	gorm.Model
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Category    string    `json:"category"`
	Code        string    `json:"code"`
	Project     Project   `json:"project"`
	ActiveStart time.Time `json:"active_start"`
	ActiveEnd   time.Time `json:"active_end"`
	Internal    bool      `json:"internal"`
	Rate        Rate      `json:"rate"`
}
type Entry struct {
	gorm.Model
	ID          uint        `json:"id"`
	Project     Project     `json:"project"`
	Employee    Employee    `json:"employee"`
	BillingCode BillingCode `json:"billing_code"`
	Start       time.Time   `json:"start_date"`
	End         time.Time   `json:"end_date"`
	Internal    bool        `json:"internal"`
	LinkedEntry *Entry      `json:"linked_period"`
	Journal     Journal     `json:"journal"`
}

// Journal is a preemptive object that we can use to split work across separate accounting Journals,
// this may refer specifically to internal vs external billables in the short term, however we may add
// additional journals in the future. This facilitates simple reporting and accounting at the ledger level.
type Journal struct {
	gorm.Model
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Entries []*Entry
}

// OBJECT METHODS

// Duration finds the length of an Entry as a duration object
func (e *Entry) Duration() time.Duration {
	duration := e.Start.Sub(e.End)
	return duration
}

// Fee finds the applicable fee in USD for a particular entry rounded to 15 minutes
func (e *Entry) Fee() float64 {
	durationMinutes := e.Duration().Minutes()
	hours := float64(durationMinutes) / 60.0
	roundedHours := float64(int(hours/0.25)) * 0.25
	fee := roundedHours * e.BillingCode.Rate.Amount
	return fee
}
