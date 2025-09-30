package cronos

import (
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const AccountsReceivable = "SNOWPACK_ACCOUNTS_RECEIVABLE"
const AccountsPayable = "SNOWPACK_ACCOUNTS_PAYABLE"
const CostOfGoodsSold = "SNOWPACK_COGS"

// Helper function to create a time pointer
func timePtr(t time.Time) *time.Time {
	return &t
}

// SeedDatabase populates the database with initial test data including users, employees,
// clients, billing codes, and timesheet entries with impersonation examples.
func (a *App) SeedDatabase() {

	// Delete all existing data except the existing development user
	a.DB.Exec("DELETE FROM adjustments")
	a.DB.Exec("DELETE FROM invoices")
	a.DB.Exec("DELETE FROM entries")
	a.DB.Exec("DELETE FROM billing_codes")
	a.DB.Exec("DELETE FROM rates")
	a.DB.Exec("DELETE FROM projects")
	a.DB.Exec("DELETE FROM accounts")
	a.DB.Exec("DELETE FROM employees WHERE user_id != 1")
	a.DB.Exec("DELETE FROM users WHERE id != 1")
	a.DB.Exec("DELETE FROM bills")

	// Get yesterday and tomorrow dates for testing
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	yesterday := today.AddDate(0, 0, -1)
	tomorrow := today.AddDate(0, 0, 1)

	// Get the existing development user with ID 1
	var devUser User
	var devEmployee Employee
	userExists := a.DB.Where("id = ?", 1).First(&devUser).RowsAffected > 0
	employeeExists := a.DB.Where("user_id = ?", 1).First(&devEmployee).RowsAffected > 0

	if !userExists {
		// If the user doesn't exist, create a default one (this shouldn't happen if the dev user was registered)
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("devpassword"), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("Error hashing password for dev user: %v", err)
			hashedPassword = []byte(DEFAULT_PASSWORD) // Fallback to plain text if hashing fails
		}
		devUser = User{
			Email:    "dev@example.com",
			IsAdmin:  true,
			Role:     UserRoleAdmin.String(),
			Password: string(hashedPassword),
		}
		a.DB.Create(&devUser)
		a.DB.Model(&devUser).Update("id", 1)
	} else {
		// User exists, but make sure password is hashed for dev environment
		if devUser.Password == DEFAULT_PASSWORD {
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte("devpassword"), bcrypt.DefaultCost)
			if err != nil {
				log.Printf("Error hashing password for existing dev user: %v", err)
			} else {
				devUser.Password = string(hashedPassword)
				a.DB.Save(&devUser)
				log.Printf("Updated existing dev user password to hashed version")
			}
		}
	}

	if !employeeExists {
		// If the employee doesn't exist, create a default one
		devEmployee = Employee{
			UserID:    1,
			Title:     "Development User",
			FirstName: "Dev",
			LastName:  "User",
			IsActive:  true,
			StartDate: yesterday,
		}
		a.DB.Create(&devEmployee)
	}

	// Create additional users for staff
	users := []User{
		{
			Email:    "nate@snowpack-data.com",
			IsAdmin:  true,
			Role:     UserRoleAdmin.String(),
			Password: DEFAULT_PASSWORD,
		},
		{
			Email:    "kevin@snowpack-data.com",
			IsAdmin:  true,
			Role:     UserRoleStaff.String(),
			Password: DEFAULT_PASSWORD,
		},
		{
			Email:    "david@snowpack-data.com",
			IsAdmin:  false,
			Role:     UserRoleStaff.String(),
			Password: DEFAULT_PASSWORD,
		},
		{
			Email:    "john@snowpack-data.com",
			IsAdmin:  false,
			Role:     UserRoleStaff.String(),
			Password: DEFAULT_PASSWORD,
		},
		{
			Email:    "jane@snowpack-data.com",
			IsAdmin:  false,
			Role:     UserRoleStaff.String(),
			Password: DEFAULT_PASSWORD,
		},
	}
	_ = a.DB.Create(&users)

	// Create additional employees (for the additional users)
	employees := []Employee{
		{
			User:      users[0],
			Title:     "Partner",
			FirstName: "Nate",
			LastName:  "Robinson",
			IsActive:  true,
			StartDate: yesterday,
		},
		{
			User:      users[1],
			Title:     "Partner",
			FirstName: "Kevin",
			LastName:  "Koenitzer",
			IsActive:  true,
			StartDate: yesterday,
		},
		{
			User:      users[2],
			Title:     "Partner",
			FirstName: "David",
			LastName:  "Shore",
			IsActive:  true,
			StartDate: yesterday,
		},
		{
			User:      users[3],
			Title:     "Senior Data Engineer",
			FirstName: "John",
			LastName:  "Doe",
			IsActive:  true,
			StartDate: yesterday,
		},
		{
			User:      users[4],
			Title:     "Data Scientist",
			FirstName: "Jane",
			LastName:  "Smith",
			IsActive:  true,
			StartDate: yesterday,
		},
	}
	_ = a.DB.Create(&employees)

	// Create initial Company for Snowpack with dev user as client
	snowpack := Account{
		Name:                  "Snowpack Data",
		Type:                  AccountTypeInternal.String(),
		LegalName:             "Snowpack Data, LLC",
		Email:                 "billing@snowpack-data.com",
		Website:               "https://snowpack-data.com",
		Clients:               []User{devUser},
		BillingFrequency:      BillingFrequencyMonthly.String(),
		ProjectsSingleInvoice: true,
	}
	_ = a.DB.Create(&snowpack)
	a.DB.Save(&snowpack)

	// Create client accounts for testing
	clientAccounts := []Account{
		{
			Name:                  "Acme Corporation",
			Type:                  AccountTypeClient.String(),
			LegalName:             "Acme Corp, Inc.",
			Email:                 "billing@acme.com",
			Website:               "https://acme.com",
			BillingFrequency:      BillingFrequencyMonthly.String(),
			ProjectsSingleInvoice: true,
		},
		{
			Name:                  "Tech Innovators",
			Type:                  AccountTypeClient.String(),
			LegalName:             "Tech Innovators LLC",
			Email:                 "accounts@techinnovators.com",
			Website:               "https://techinnovators.com",
			BillingFrequency:      BillingFrequencyBiweekly.String(),
			ProjectsSingleInvoice: false,
		},
		{
			Name:                  "Data Solutions",
			Type:                  AccountTypeClient.String(),
			LegalName:             "Data Solutions Inc.",
			Email:                 "finance@datasolutions.com",
			Website:               "https://datasolutions.com",
			BillingFrequency:      BillingFrequencyWeekly.String(),
			ProjectsSingleInvoice: true,
		},
	}
	_ = a.DB.Create(&clientAccounts)

	// Create projects for testing - using yesterday to tomorrow for active dates
	// Assign the dev user as Account Executive (AE) and Sales Development Representative (SDR)
	// for commission calculation purposes
	aeID := devEmployee.ID
	sdrID := employees[0].ID // Kevin as secondary SDR

	projects := []Project{
		{
			Name:             "Data Platform Development",
			AccountID:        clientAccounts[0].ID,
			ActiveStart:      yesterday,
			ActiveEnd:        tomorrow,
			BudgetHours:      500,
			BudgetDollars:    75000,
			BudgetCapHours:   2000,
			BudgetCapDollars: 300000,
			Internal:         false,
			BillingFrequency: BillingFrequencyMonthly.String(),
			ProjectType:      ProjectTypeNew.String(),
			AEID:             &aeID,
			SDRID:            &sdrID,
		},
		{
			Name:             "AI Implementation",
			AccountID:        clientAccounts[1].ID,
			ActiveStart:      yesterday,
			ActiveEnd:        tomorrow,
			BudgetHours:      300,
			BudgetDollars:    60000,
			BudgetCapHours:   1200,
			BudgetCapDollars: 240000,
			Internal:         false,
			BillingFrequency: BillingFrequencyBiweekly.String(),
			ProjectType:      ProjectTypeExisting.String(),
			AEID:             &aeID,
		},
		{
			Name:             "Data Strategy Consulting",
			AccountID:        clientAccounts[2].ID,
			ActiveStart:      yesterday,
			ActiveEnd:        tomorrow,
			BudgetHours:      200,
			BudgetDollars:    45000,
			BudgetCapHours:   800,
			BudgetCapDollars: 180000,
			Internal:         false,
			BillingFrequency: BillingFrequencyWeekly.String(),
			ProjectType:      ProjectTypeNew.String(),
			SDRID:            &aeID,
		},
		{
			Name:             "Internal Operations",
			AccountID:        snowpack.ID,
			ActiveStart:      yesterday,
			ActiveEnd:        tomorrow,
			BudgetHours:      100,
			BudgetDollars:    7500,
			BudgetCapHours:   400,
			BudgetCapDollars: 30000,
			Internal:         true,
			BillingFrequency: BillingFrequencyMonthly.String(),
		},
	}
	_ = a.DB.Create(&projects)

	// Create rates for billing codes
	rates := []Rate{
		{
			Name:         "Standard Rate",
			Amount:       150.00,
			ActiveFrom:   yesterday,
			ActiveTo:     tomorrow,
			InternalOnly: false,
		},
		{
			Name:         "Premium Rate",
			Amount:       200.00,
			ActiveFrom:   yesterday,
			ActiveTo:     tomorrow,
			InternalOnly: false,
		},
		{
			Name:         "Consulting Rate",
			Amount:       225.00,
			ActiveFrom:   yesterday,
			ActiveTo:     tomorrow,
			InternalOnly: false,
		},
		{
			Name:         "Internal Standard Rate",
			Amount:       75.00,
			ActiveFrom:   yesterday,
			ActiveTo:     tomorrow,
			InternalOnly: true,
		},
		{
			Name:         "Internal Premium Rate",
			Amount:       100.00,
			ActiveFrom:   yesterday,
			ActiveTo:     tomorrow,
			InternalOnly: true,
		},
		{
			Name:         "Internal Consulting Rate",
			Amount:       125.00,
			ActiveFrom:   yesterday,
			ActiveTo:     tomorrow,
			InternalOnly: true,
		},
	}
	_ = a.DB.Create(&rates)

	// Create billing codes for testing - using yesterday to tomorrow for active dates
	billingCodes := []BillingCode{
		{
			Name:           "Data Engineering",
			RateType:       RateTypeExternalBillable.String(),
			Category:       "Development",
			Code:           "DE-100",
			RoundedTo:      15,
			ProjectID:      projects[0].ID,
			ActiveStart:    yesterday,
			ActiveEnd:      tomorrow,
			RateID:         rates[0].ID,
			InternalRateID: rates[3].ID,
		},
		{
			Name:           "Data Analysis",
			RateType:       RateTypeExternalBillable.String(),
			Category:       "Analysis",
			Code:           "DA-101",
			RoundedTo:      15,
			ProjectID:      projects[0].ID,
			ActiveStart:    yesterday,
			ActiveEnd:      tomorrow,
			RateID:         rates[0].ID,
			InternalRateID: rates[3].ID,
		},
		{
			Name:           "Machine Learning",
			RateType:       RateTypeExternalBillable.String(),
			Category:       "AI",
			Code:           "ML-200",
			RoundedTo:      15,
			ProjectID:      projects[1].ID,
			ActiveStart:    yesterday,
			ActiveEnd:      tomorrow,
			RateID:         rates[1].ID,
			InternalRateID: rates[4].ID,
		},
		{
			Name:           "Data Strategy",
			RateType:       RateTypeExternalBillable.String(),
			Category:       "Consulting",
			Code:           "DS-300",
			RoundedTo:      15,
			ProjectID:      projects[2].ID,
			ActiveStart:    yesterday,
			ActiveEnd:      tomorrow,
			RateID:         rates[2].ID,
			InternalRateID: rates[5].ID,
		},
		{
			Name:           "Internal Operations",
			RateType:       RateTypeInternalProject.String(),
			Category:       "Operations",
			Code:           "INT-100",
			RoundedTo:      15,
			ProjectID:      projects[3].ID,
			ActiveStart:    yesterday,
			ActiveEnd:      tomorrow,
			RateID:         rates[3].ID,
			InternalRateID: rates[3].ID,
		},
	}
	_ = a.DB.Create(&billingCodes)

	// Create timesheet entries including impersonated entries
	// Helper function to create entry ID pointers
	createUintPtr := func(id uint) *uint {
		return &id
	}

	// Generate dates for the past two weeks
	twoWeeksAgo := today.AddDate(0, 0, -14)

	// Create a more varied set of entries spanning the past two weeks
	entries := []Entry{
		// Week 1 - Two weeks ago
		// Monday (Day 1)
		{
			ProjectID:     projects[0].ID, // Data Platform Development project
			EmployeeID:    devEmployee.ID,
			BillingCodeID: billingCodes[0].ID, // Data Engineering for Data Platform Development
			Start:         twoWeeksAgo.Add(9 * time.Hour),
			End:           twoWeeksAgo.Add(12 * time.Hour),
			Notes:         "Initial project planning and requirements gathering",
			State:         EntryStateDraft.String(),
		},
		{
			ProjectID:     projects[0].ID, // Data Platform Development project
			EmployeeID:    devEmployee.ID,
			BillingCodeID: billingCodes[0].ID, // Data Engineering
			Start:         twoWeeksAgo.Add(13 * time.Hour),
			End:           twoWeeksAgo.Add(17 * time.Hour),
			Notes:         "Architecture design for data pipeline",
			State:         EntryStateDraft.String(),
		},

		// Tuesday (Day 2)
		{
			ProjectID:     projects[0].ID, // Data Platform Development project
			EmployeeID:    devEmployee.ID,
			BillingCodeID: billingCodes[1].ID, // Data Analysis
			Start:         twoWeeksAgo.AddDate(0, 0, 1).Add(8 * time.Hour),
			End:           twoWeeksAgo.AddDate(0, 0, 1).Add(11 * time.Hour),
			Notes:         "Data analysis for quarterly report",
			State:         EntryStateDraft.String(),
		},
		{
			ProjectID:           projects[0].ID,                 // Data Platform Development project
			EmployeeID:          devEmployee.ID,                 // Dev user
			ImpersonateAsUserID: createUintPtr(employees[2].ID), // John
			BillingCodeID:       billingCodes[0].ID,             // Data Engineering
			Start:               twoWeeksAgo.AddDate(0, 0, 1).Add(13 * time.Hour),
			End:                 twoWeeksAgo.AddDate(0, 0, 1).Add(16 * time.Hour),
			Notes:               "Data ingestion framework development (impersonated)",
			State:               EntryStateDraft.String(),
		},

		// Wednesday (Day 3)
		{
			ProjectID:     projects[1].ID, // AI Implementation project
			EmployeeID:    devEmployee.ID,
			BillingCodeID: billingCodes[2].ID, // Machine Learning
			Start:         twoWeeksAgo.AddDate(0, 0, 2).Add(9 * time.Hour),
			End:           twoWeeksAgo.AddDate(0, 0, 2).Add(12 * time.Hour),
			Notes:         "Machine learning model selection and initial setup",
			State:         EntryStateDraft.String(),
		},
		{
			ProjectID:     projects[1].ID, // AI Implementation project
			EmployeeID:    devEmployee.ID,
			BillingCodeID: billingCodes[2].ID, // Machine Learning
			Start:         twoWeeksAgo.AddDate(0, 0, 2).Add(13 * time.Hour),
			End:           twoWeeksAgo.AddDate(0, 0, 2).Add(17 * time.Hour),
			Notes:         "Data preprocessing for ML training",
			State:         EntryStateDraft.String(),
		},

		// Thursday (Day 4)
		{
			ProjectID:           projects[2].ID,                 // Data Strategy Consulting project
			EmployeeID:          devEmployee.ID,                 // Dev user
			ImpersonateAsUserID: createUintPtr(employees[3].ID), // Jane
			BillingCodeID:       billingCodes[3].ID,             // Data Strategy
			Start:               twoWeeksAgo.AddDate(0, 0, 3).Add(10 * time.Hour),
			End:                 twoWeeksAgo.AddDate(0, 0, 3).Add(15 * time.Hour),
			Notes:               "Client consultation on data strategy roadmap (impersonated)",
			State:               EntryStateDraft.String(),
		},

		// Friday (Day 5)
		{
			ProjectID:     projects[3].ID, // Internal Operations project
			EmployeeID:    devEmployee.ID,
			BillingCodeID: billingCodes[4].ID, // Internal Operations
			Start:         twoWeeksAgo.AddDate(0, 0, 4).Add(9 * time.Hour),
			End:           twoWeeksAgo.AddDate(0, 0, 4).Add(12 * time.Hour),
			Notes:         "Team planning and sprint retrospective",
			State:         EntryStateDraft.String(),
			Internal:      true,
		},
		{
			ProjectID:     projects[3].ID, // Internal Operations project
			EmployeeID:    devEmployee.ID,
			BillingCodeID: billingCodes[4].ID, // Internal Operations
			Start:         twoWeeksAgo.AddDate(0, 0, 4).Add(13 * time.Hour),
			End:           twoWeeksAgo.AddDate(0, 0, 4).Add(17 * time.Hour),
			Notes:         "Documentation and knowledge sharing",
			State:         EntryStateDraft.String(),
			Internal:      true,
		},

		// Week 2 - Last week
		// Monday (Day 8)
		{
			ProjectID:     projects[0].ID, // Data Platform Development project
			EmployeeID:    devEmployee.ID,
			BillingCodeID: billingCodes[0].ID, // Data Engineering
			Start:         twoWeeksAgo.AddDate(0, 0, 7).Add(9 * time.Hour),
			End:           twoWeeksAgo.AddDate(0, 0, 7).Add(12 * time.Hour),
			Notes:         "Data pipeline implementation - batch processing module",
			State:         EntryStateDraft.String(),
		},
		{
			ProjectID:           projects[1].ID,                 // AI Implementation project
			EmployeeID:          devEmployee.ID,                 // Dev user
			ImpersonateAsUserID: createUintPtr(employees[1].ID), // David
			BillingCodeID:       billingCodes[2].ID,             // Machine Learning
			Start:               twoWeeksAgo.AddDate(0, 0, 7).Add(13 * time.Hour),
			End:                 twoWeeksAgo.AddDate(0, 0, 7).Add(16 * time.Hour),
			Notes:               "Machine learning model training session (impersonated)",
			State:               EntryStateDraft.String(),
		},

		// Tuesday (Day 9)
		{
			ProjectID:     projects[0].ID, // Data Platform Development project
			EmployeeID:    devEmployee.ID,
			BillingCodeID: billingCodes[1].ID, // Data Analysis
			Start:         twoWeeksAgo.AddDate(0, 0, 8).Add(8 * time.Hour),
			End:           twoWeeksAgo.AddDate(0, 0, 8).Add(12 * time.Hour),
			Notes:         "Data quality assessment and cleaning procedures",
			State:         EntryStateDraft.String(),
		},
		{
			ProjectID:     projects[0].ID, // Data Platform Development project
			EmployeeID:    devEmployee.ID,
			BillingCodeID: billingCodes[1].ID, // Data Analysis
			Start:         twoWeeksAgo.AddDate(0, 0, 8).Add(13 * time.Hour),
			End:           twoWeeksAgo.AddDate(0, 0, 8).Add(16 * time.Hour),
			Notes:         "Creating ETL workflows for scheduled reporting",
			State:         EntryStateDraft.String(),
		},

		// Wednesday (Day 10)
		{
			ProjectID:           projects[0].ID,                 // Data Platform Development project
			EmployeeID:          devEmployee.ID,                 // Dev user
			ImpersonateAsUserID: createUintPtr(employees[2].ID), // John
			BillingCodeID:       billingCodes[0].ID,             // Data Engineering
			Start:               twoWeeksAgo.AddDate(0, 0, 9).Add(9 * time.Hour),
			End:                 twoWeeksAgo.AddDate(0, 0, 9).Add(12 * time.Hour),
			Notes:               "Implementing data validation layers (impersonated)",
			State:               EntryStateDraft.String(),
		},
		{
			ProjectID:     projects[3].ID, // Internal Operations project
			EmployeeID:    devEmployee.ID,
			BillingCodeID: billingCodes[4].ID, // Internal Operations
			Start:         twoWeeksAgo.AddDate(0, 0, 9).Add(14 * time.Hour),
			End:           twoWeeksAgo.AddDate(0, 0, 9).Add(17 * time.Hour),
			Notes:         "Internal team training on new tools",
			State:         EntryStateDraft.String(),
			Internal:      true,
		},

		// Thursday (Day 11)
		{
			ProjectID:     projects[2].ID, // Data Strategy Consulting project
			EmployeeID:    devEmployee.ID,
			BillingCodeID: billingCodes[3].ID, // Data Strategy
			Start:         twoWeeksAgo.AddDate(0, 0, 10).Add(10 * time.Hour),
			End:           twoWeeksAgo.AddDate(0, 0, 10).Add(14 * time.Hour),
			Notes:         "Client meeting to discuss implementation strategy",
			State:         EntryStateDraft.String(),
		},
		{
			ProjectID:           projects[1].ID,                 // AI Implementation project
			EmployeeID:          devEmployee.ID,                 // Dev user
			ImpersonateAsUserID: createUintPtr(employees[3].ID), // Jane
			BillingCodeID:       billingCodes[2].ID,             // Machine Learning
			Start:               twoWeeksAgo.AddDate(0, 0, 10).Add(15 * time.Hour),
			End:                 twoWeeksAgo.AddDate(0, 0, 10).Add(18 * time.Hour),
			Notes:               "Feature engineering and model evaluation (impersonated)",
			State:               EntryStateDraft.String(),
		},

		// Friday (Day 12)
		{
			ProjectID:     projects[3].ID, // Internal Operations project
			EmployeeID:    devEmployee.ID,
			BillingCodeID: billingCodes[4].ID, // Internal Operations
			Start:         twoWeeksAgo.AddDate(0, 0, 11).Add(9 * time.Hour),
			End:           twoWeeksAgo.AddDate(0, 0, 11).Add(13 * time.Hour),
			Notes:         "Sprint planning and backlog grooming",
			State:         EntryStateDraft.String(),
			Internal:      true,
		},
		{
			ProjectID:           projects[2].ID,                 // Data Strategy Consulting project
			EmployeeID:          devEmployee.ID,                 // Dev user
			ImpersonateAsUserID: createUintPtr(employees[1].ID), // David
			BillingCodeID:       billingCodes[3].ID,             // Data Strategy
			Start:               twoWeeksAgo.AddDate(0, 0, 11).Add(14 * time.Hour),
			End:                 twoWeeksAgo.AddDate(0, 0, 11).Add(17 * time.Hour),
			Notes:               "Strategic roadmap development with client (impersonated)",
			State:               EntryStateDraft.String(),
		},

		// This week
		// Monday (Day 15) - Yesterday
		{
			ProjectID:     projects[0].ID, // Data Platform Development project
			EmployeeID:    devEmployee.ID,
			BillingCodeID: billingCodes[0].ID, // Data Engineering
			Start:         yesterday.Add(8 * time.Hour),
			End:           yesterday.Add(12 * time.Hour),
			Notes:         "Integration testing for data pipeline components",
			State:         EntryStateDraft.String(),
		},
		{
			ProjectID:     projects[0].ID, // Data Platform Development project
			EmployeeID:    devEmployee.ID,
			BillingCodeID: billingCodes[0].ID, // Data Engineering
			Start:         yesterday.Add(13 * time.Hour),
			End:           yesterday.Add(17 * time.Hour),
			Notes:         "Performance optimization for data processing",
			State:         EntryStateDraft.String(),
		},

		// Today
		{
			ProjectID:     projects[1].ID, // AI Implementation project
			EmployeeID:    devEmployee.ID,
			BillingCodeID: billingCodes[2].ID, // Machine Learning
			Start:         today.Add(9 * time.Hour),
			End:           today.Add(12 * time.Hour),
			Notes:         "ML model deployment preparation",
			State:         EntryStateDraft.String(),
		},
		{
			ProjectID:           projects[3].ID,                 // Internal Operations project
			EmployeeID:          devEmployee.ID,                 // Dev user
			ImpersonateAsUserID: createUintPtr(employees[3].ID), // Jane
			BillingCodeID:       billingCodes[4].ID,             // Internal Operations
			Start:               today.Add(13 * time.Hour),
			End:                 today.Add(16 * time.Hour),
			Notes:               "Documentation and process improvement (impersonated)",
			State:               EntryStateDraft.String(),
			Internal:            true,
		},
	}
	_ = a.DB.Create(&entries)

	// Create draft invoices for testing
	// First, set up some dates for different billing periods
	february2025Start := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)
	february2025End := time.Date(2025, 2, 28, 23, 59, 59, 0, time.UTC)

	march2025Start := time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC)
	march2025End := time.Date(2025, 3, 31, 23, 59, 59, 0, time.UTC)

	// Future project period (for project-based billing)
	futureProjectStart := time.Date(2025, 4, 1, 0, 0, 0, 0, time.UTC)
	futureProjectEnd := time.Date(2025, 6, 30, 23, 59, 59, 0, time.UTC)

	// 1. Create a draft invoice for Acme Corp (account with ProjectsSingleInvoice = true)
	// This account gets a single invoice for all projects
	acmeInvoice := Invoice{
		Name:        "Acme Corporation: 02.01.2025-02.28.2025",
		AccountID:   clientAccounts[0].ID,
		Account:     clientAccounts[0],
		PeriodStart: february2025Start,
		PeriodEnd:   february2025End,
		State:       InvoiceStateDraft.String(),
		Type:        InvoiceTypeAR.String(),
	}
	_ = a.DB.Create(&acmeInvoice)

	// Associate some draft entries with this invoice
	// For simplicity, we'll create new entries specifically for the invoice periods
	acmeEntries := []Entry{
		{
			ProjectID:     projects[0].ID,
			EmployeeID:    employees[0].ID,
			BillingCodeID: billingCodes[0].ID,
			Start:         february2025Start.Add(10 * time.Hour),
			End:           february2025Start.Add(14 * time.Hour),
			Notes:         "Acme data platform architecture review",
			State:         EntryStateDraft.String(),
			InvoiceID:     &acmeInvoice.ID,
		},
		{
			ProjectID:     projects[0].ID,
			EmployeeID:    employees[1].ID,
			BillingCodeID: billingCodes[1].ID,
			Start:         february2025Start.AddDate(0, 0, 2).Add(9 * time.Hour),
			End:           february2025Start.AddDate(0, 0, 2).Add(17 * time.Hour),
			Notes:         "Acme data analysis for quarterly planning",
			State:         EntryStateDraft.String(),
			InvoiceID:     &acmeInvoice.ID,
		},
	}
	_ = a.DB.Create(&acmeEntries)

	// Calculate and update the invoice totals
	var totalHours float64
	var totalFees float64
	for _, entry := range acmeEntries {
		// Load BillingCode and Rate for fee calculation
		var fullEntry Entry
		a.DB.Preload("BillingCode.Rate").Where("ID = ?", entry.ID).First(&fullEntry)

		duration := fullEntry.End.Sub(fullEntry.Start).Hours()
		fee := duration * fullEntry.BillingCode.Rate.Amount

		totalHours += duration
		totalFees += fee

		// Update the entry fee
		fullEntry.Fee = int(fee)
		a.DB.Save(&fullEntry)
	}

	acmeInvoice.TotalHours = totalHours
	acmeInvoice.TotalFees = totalFees
	acmeInvoice.TotalAmount = totalFees
	a.DB.Save(&acmeInvoice)

	// 2. Create separate project invoices for Tech Innovators (account with ProjectsSingleInvoice = false)
	// This account gets separate invoices per project
	techInvoice1 := Invoice{
		Name:        "Tech Innovators - AI Implementation: 02.01.2025-02.28.2025",
		AccountID:   clientAccounts[1].ID,
		Account:     clientAccounts[1],
		ProjectID:   &projects[1].ID,
		Project:     projects[1],
		PeriodStart: february2025Start,
		PeriodEnd:   february2025End,
		State:       InvoiceStateDraft.String(),
		Type:        InvoiceTypeAR.String(),
	}
	_ = a.DB.Create(&techInvoice1)

	// Create entries for this project invoice
	techEntries := []Entry{
		{
			ProjectID:     projects[1].ID,
			EmployeeID:    employees[2].ID,
			BillingCodeID: billingCodes[2].ID,
			Start:         february2025Start.AddDate(0, 0, 5).Add(8 * time.Hour),
			End:           february2025Start.AddDate(0, 0, 5).Add(12 * time.Hour),
			Notes:         "Tech Innovators ML algorithm implementation",
			State:         EntryStateDraft.String(),
			InvoiceID:     &techInvoice1.ID,
		},
		{
			ProjectID:     projects[1].ID,
			EmployeeID:    employees[3].ID,
			BillingCodeID: billingCodes[2].ID,
			Start:         february2025Start.AddDate(0, 0, 6).Add(13 * time.Hour),
			End:           february2025Start.AddDate(0, 0, 6).Add(18 * time.Hour),
			Notes:         "Tech Innovators model training and validation",
			State:         EntryStateDraft.String(),
			InvoiceID:     &techInvoice1.ID,
		},
	}
	_ = a.DB.Create(&techEntries)

	// Calculate totals
	totalHours = 0
	totalFees = 0
	for _, entry := range techEntries {
		var fullEntry Entry
		a.DB.Preload("BillingCode.Rate").Where("ID = ?", entry.ID).First(&fullEntry)

		duration := fullEntry.End.Sub(fullEntry.Start).Hours()
		fee := duration * fullEntry.BillingCode.Rate.Amount

		totalHours += duration
		totalFees += fee

		fullEntry.Fee = int(fee)
		a.DB.Save(&fullEntry)
	}

	techInvoice1.TotalHours = totalHours
	techInvoice1.TotalFees = totalFees
	techInvoice1.TotalAmount = totalFees
	a.DB.Save(&techInvoice1)

	// 3. Create a future period invoice for Data Solutions with a project-based billing frequency
	// First, update the project's billing frequency to project-based
	a.DB.Model(&projects[2]).Update("billing_frequency", BillingFrequencyProject.String())

	dataInvoice := Invoice{
		Name:        "Data Solutions - Data Strategy Consulting: 04.01.2025-06.30.2025",
		AccountID:   clientAccounts[2].ID,
		Account:     clientAccounts[2],
		ProjectID:   &projects[2].ID,
		Project:     projects[2],
		PeriodStart: futureProjectStart,
		PeriodEnd:   futureProjectEnd,
		State:       InvoiceStateDraft.String(),
		Type:        InvoiceTypeAR.String(),
	}
	_ = a.DB.Create(&dataInvoice)

	// Create entries for this project invoice
	dataEntries := []Entry{
		{
			ProjectID:     projects[2].ID,
			EmployeeID:    employees[0].ID,
			BillingCodeID: billingCodes[3].ID,
			Start:         futureProjectStart.AddDate(0, 0, 3).Add(9 * time.Hour),
			End:           futureProjectStart.AddDate(0, 0, 3).Add(17 * time.Hour),
			Notes:         "Data Solutions initial strategy workshop",
			State:         EntryStateDraft.String(),
			InvoiceID:     &dataInvoice.ID,
		},
		{
			ProjectID:     projects[2].ID,
			EmployeeID:    employees[1].ID,
			BillingCodeID: billingCodes[3].ID,
			Start:         futureProjectStart.AddDate(0, 0, 10).Add(10 * time.Hour),
			End:           futureProjectStart.AddDate(0, 0, 10).Add(16 * time.Hour),
			Notes:         "Data Solutions architecture planning",
			State:         EntryStateDraft.String(),
			InvoiceID:     &dataInvoice.ID,
		},
	}
	_ = a.DB.Create(&dataEntries)

	// Calculate totals
	totalHours = 0
	totalFees = 0
	for _, entry := range dataEntries {
		var fullEntry Entry
		a.DB.Preload("BillingCode.Rate").Where("ID = ?", entry.ID).First(&fullEntry)

		duration := fullEntry.End.Sub(fullEntry.Start).Hours()
		fee := duration * fullEntry.BillingCode.Rate.Amount

		totalHours += duration
		totalFees += fee

		fullEntry.Fee = int(fee)
		a.DB.Save(&fullEntry)
	}

	dataInvoice.TotalHours = totalHours
	dataInvoice.TotalFees = totalFees
	dataInvoice.TotalAmount = totalFees
	a.DB.Save(&dataInvoice)

	// 4. Create an invoice for March 2025 with Adjustments
	marchInvoice := Invoice{
		Name:        "Acme Corporation: 03.01.2025-03.31.2025",
		AccountID:   clientAccounts[0].ID,
		Account:     clientAccounts[0],
		PeriodStart: march2025Start,
		PeriodEnd:   march2025End,
		State:       InvoiceStateDraft.String(),
		Type:        InvoiceTypeAR.String(),
	}
	_ = a.DB.Create(&marchInvoice)

	// Create entries for March
	marchEntries := []Entry{
		{
			ProjectID:     projects[0].ID,
			EmployeeID:    employees[2].ID,
			BillingCodeID: billingCodes[0].ID,
			Start:         march2025Start.AddDate(0, 0, 5).Add(9 * time.Hour),
			End:           march2025Start.AddDate(0, 0, 5).Add(15 * time.Hour),
			Notes:         "Acme data pipeline implementation",
			State:         EntryStateDraft.String(),
			InvoiceID:     &marchInvoice.ID,
		},
	}
	_ = a.DB.Create(&marchEntries)

	// Calculate totals for entries
	totalHours = 0
	totalFees = 0
	for _, entry := range marchEntries {
		var fullEntry Entry
		a.DB.Preload("BillingCode.Rate").Where("ID = ?", entry.ID).First(&fullEntry)

		duration := fullEntry.End.Sub(fullEntry.Start).Hours()
		fee := duration * fullEntry.BillingCode.Rate.Amount

		totalHours += duration
		totalFees += fee

		fullEntry.Fee = int(fee)
		a.DB.Save(&fullEntry)
	}

	// Create adjustments for this invoice
	adjustments := []Adjustment{
		{
			InvoiceID: &marchInvoice.ID,
			Type:      AdjustmentTypeCredit.String(),
			State:     AdjustmentStateDraft.String(),
			Amount:    -250.00, // Credit (negative amount)
			Notes:     "Goodwill credit for exceeding estimated hours",
		},
		{
			InvoiceID: &marchInvoice.ID,
			Type:      AdjustmentTypeFee.String(),
			State:     AdjustmentStateDraft.String(),
			Amount:    100.00, // Additional fee (positive amount)
			Notes:     "Expedited delivery fee",
		},
	}
	_ = a.DB.Create(&adjustments)

	// Calculate total adjustments
	var totalAdjustments float64
	for _, adj := range adjustments {
		totalAdjustments += adj.Amount
	}

	// Update invoice totals including adjustments
	marchInvoice.TotalHours = totalHours
	marchInvoice.TotalFees = totalFees
	marchInvoice.TotalAdjustments = totalAdjustments
	marchInvoice.TotalAmount = totalFees + totalAdjustments
	a.DB.Save(&marchInvoice)

	// Create sample bills for Accounts Payable view
	// Define some periods
	previousMonth := time.Date(now.Year(), now.Month()-1, 1, 0, 0, 0, 0, time.UTC)
	previousMonthEnd := time.Date(now.Year(), now.Month(), 0, 23, 59, 59, 0, time.UTC)
	twoMonthsAgo := time.Date(now.Year(), now.Month()-2, 1, 0, 0, 0, 0, time.UTC)
	twoMonthsAgoEnd := time.Date(now.Year(), now.Month()-1, 0, 23, 59, 59, 0, time.UTC)

	// Add bills with different states
	bills := []Bill{
		{
			Name:        "Bill for John Smith - " + previousMonth.Format("01.02.2006") + " - " + previousMonthEnd.Format("01.02.2006"),
			EmployeeID:  employees[0].ID,
			Employee:    employees[0],
			PeriodStart: previousMonth,
			PeriodEnd:   previousMonthEnd,
			TotalHours:  42.5,
			TotalFees:   4250,
			TotalAmount: 4250,
			GCSFile:     "https://storage.googleapis.com/snowpack-cronos-testing/bills/Payroll_Nate_Robinson_02.01.2025_-_02.28.2025.pdf",
		},
		{
			Name:        "Bill for Jane Wilson - " + previousMonth.Format("01.02.2006") + " - " + previousMonthEnd.Format("01.02.2006"),
			EmployeeID:  employees[1].ID,
			Employee:    employees[1],
			PeriodStart: previousMonth,
			PeriodEnd:   previousMonthEnd,
			TotalHours:  36.0,
			TotalFees:   3600,
			TotalAmount: 3600,
			ClosedAt:    timePtr(time.Now().AddDate(0, 0, -3)),
		},
		{
			Name:        "Bill for Tom Johnson - " + twoMonthsAgo.Format("01.02.2006") + " - " + twoMonthsAgoEnd.Format("01.02.2006"),
			EmployeeID:  employees[2].ID,
			Employee:    employees[2],
			PeriodStart: twoMonthsAgo,
			PeriodEnd:   twoMonthsAgoEnd,
			TotalHours:  45.0,
			TotalFees:   4500,
			TotalAmount: 4500,
			ClosedAt:    timePtr(time.Now().AddDate(0, 0, -15)),
		},
		{
			Name:        "Bill for Jane Wilson - " + twoMonthsAgo.Format("01.02.2006") + " - " + twoMonthsAgoEnd.Format("01.02.2006"),
			EmployeeID:  employees[1].ID,
			Employee:    employees[1],
			PeriodStart: twoMonthsAgo,
			PeriodEnd:   twoMonthsAgoEnd,
			TotalHours:  40.0,
			TotalFees:   4000,
			TotalAmount: 4000,
		},
	}

	_ = a.DB.Create(&bills)

	// Create invoices in different states for Accounts Receivable view
	// Create an approved invoice
	approvedInvoice := Invoice{
		Name:        "Acme Technologies - Website Development: " + previousMonth.Format("01.02.2006") + "-" + previousMonthEnd.Format("01.02.2006"),
		AccountID:   clientAccounts[0].ID,
		Account:     clientAccounts[0],
		ProjectID:   &projects[0].ID,
		Project:     projects[0],
		PeriodStart: previousMonth,
		PeriodEnd:   previousMonthEnd,
		State:       InvoiceStateApproved.String(),
		Type:        InvoiceTypeAR.String(),
		TotalHours:  42.0,
		TotalFees:   6300.00,
		TotalAmount: 6300.00,
		AcceptedAt:  time.Now().AddDate(0, 0, -7),
		DueAt:       time.Now().AddDate(0, 0, 23), // Due in 23 days
		GCSFile:     "https://storage.googleapis.com/snowpack-cronos-testing/bills/Payroll_Nate_Robinson_02.01.2025_-_02.28.2025.pdf",
	}
	_ = a.DB.Create(&approvedInvoice)

	// Create a sent invoice
	sentInvoice := Invoice{
		Name:        "Tech Innovators - AI Implementation: " + previousMonth.Format("01.02.2006") + "-" + previousMonthEnd.Format("01.02.2006"),
		AccountID:   clientAccounts[1].ID,
		Account:     clientAccounts[1],
		ProjectID:   &projects[1].ID,
		Project:     projects[1],
		PeriodStart: previousMonth,
		PeriodEnd:   previousMonthEnd,
		State:       InvoiceStateSent.String(),
		Type:        InvoiceTypeAR.String(),
		TotalHours:  48.0,
		TotalFees:   7200.00,
		TotalAmount: 7200.00,
		AcceptedAt:  time.Now().AddDate(0, 0, -14),
		SentAt:      time.Now().AddDate(0, 0, -10),
		DueAt:       time.Now().AddDate(0, 0, 5), // Due in 5 days
		GCSFile:     "https://storage.googleapis.com/snowpack-cronos-testing/bills/Payroll_Nate_Robinson_02.01.2025_-_02.28.2025.pdf",
	}
	_ = a.DB.Create(&sentInvoice)

	// Create an overdue sent invoice
	overdueSentInvoice := Invoice{
		Name:        "Global Industries - Data Migration: " + twoMonthsAgo.Format("01.02.2006") + "-" + twoMonthsAgoEnd.Format("01.02.2006"),
		AccountID:   clientAccounts[2].ID,
		Account:     clientAccounts[2],
		ProjectID:   &projects[2].ID,
		Project:     projects[2],
		PeriodStart: twoMonthsAgo,
		PeriodEnd:   twoMonthsAgoEnd,
		State:       InvoiceStateSent.String(),
		Type:        InvoiceTypeAR.String(),
		TotalHours:  45.0,
		TotalFees:   6750.00,
		TotalAmount: 6750.00,
		AcceptedAt:  time.Now().AddDate(0, 0, -35),
		SentAt:      time.Now().AddDate(0, 0, -30),
		DueAt:       time.Now().AddDate(0, 0, -5), // 5 days overdue
	}
	_ = a.DB.Create(&overdueSentInvoice)

	// Create a paid invoice
	paidInvoice := Invoice{
		Name:        "Acme Corporation: " + twoMonthsAgo.Format("01.02.2006") + "-" + twoMonthsAgoEnd.Format("01.02.2006"),
		AccountID:   clientAccounts[0].ID,
		Account:     clientAccounts[0],
		PeriodStart: twoMonthsAgo,
		PeriodEnd:   twoMonthsAgoEnd,
		State:       InvoiceStatePaid.String(),
		Type:        InvoiceTypeAR.String(),
		TotalHours:  30.0,
		TotalFees:   4500.00,
		TotalAmount: 4500.00,
		AcceptedAt:  time.Now().AddDate(0, 0, -45),
		SentAt:      time.Now().AddDate(0, 0, -40),
		DueAt:       time.Now().AddDate(0, 0, -10),
		ClosedAt:    time.Now().AddDate(0, 0, -12),
	}
	_ = a.DB.Create(&paidInvoice)

	// Create a void invoice
	voidInvoice := Invoice{
		Name:        "Tech Innovators - Support Contract: " + twoMonthsAgo.Format("01.02.2006") + "-" + twoMonthsAgoEnd.Format("01.02.2006"),
		AccountID:   clientAccounts[1].ID,
		Account:     clientAccounts[1],
		ProjectID:   &projects[1].ID,
		Project:     projects[1],
		PeriodStart: twoMonthsAgo,
		PeriodEnd:   twoMonthsAgoEnd,
		State:       InvoiceStateVoid.String(),
		Type:        InvoiceTypeAR.String(),
		TotalHours:  28.0,
		TotalFees:   4200.00,
		TotalAmount: 4200.00,
		AcceptedAt:  time.Now().AddDate(0, 0, -45),
	}
	_ = a.DB.Create(&voidInvoice)
}
