package cronos

import "time"

const AccountsReceivable = "SNOWPACK_ACCOUNTS_RECEIVABLE"
const AccountsPayable = "SNOWPACK_ACCOUNTS_PAYABLE"
const CostOfGoodsSold = "SNOWPACK_COGS"

// SeedDatabase populates the database with initial test data including users, employees,
// clients, billing codes, and timesheet entries with impersonation examples.
func (a *App) SeedDatabase() {
	// Get yesterday and tomorrow dates for testing
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	yesterday := today.AddDate(0, 0, -1)
	tomorrow := today.AddDate(0, 0, 1)

	// Create Users
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

	// Create initial Company for Snowpack
	snowpack := Account{
		Name:                  "Snowpack Data",
		Type:                  AccountTypeInternal.String(),
		LegalName:             "Snowpack Data, LLC",
		Email:                 "billing@snowpack-data.com",
		Website:               "https://snowpack-data.com",
		Clients:               users,
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
	// Assign employees as Account Executives (AE) and Sales Development Representatives (SDR)
	// for commission calculation purposes
	aeID := employees[0].ID
	sdrID := employees[1].ID

	projects := []Project{
		{
			Name:             "Data Platform Development",
			AccountID:        clientAccounts[0].ID,
			ActiveStart:      yesterday,
			ActiveEnd:        tomorrow,
			BudgetHours:      500,
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
			Internal:         false,
			BillingFrequency: BillingFrequencyWeekly.String(),
			ProjectType:      ProjectTypeNew.String(),
			SDRID:            &sdrID,
		},
		{
			Name:             "Internal Operations",
			AccountID:        snowpack.ID,
			ActiveStart:      yesterday,
			ActiveEnd:        tomorrow,
			BudgetHours:      100,
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

	entries := []Entry{
		// Regular entries
		{
			EmployeeID:    employees[0].ID,
			BillingCodeID: billingCodes[0].ID,
			Start:         yesterday.Add(9 * time.Hour),
			End:           yesterday.Add(12 * time.Hour),
			Notes:         "Initial project planning",
			State:         EntryStateApproved.String(),
		},
		{
			EmployeeID:    employees[1].ID,
			BillingCodeID: billingCodes[1].ID,
			Start:         yesterday.Add(13 * time.Hour),
			End:           yesterday.Add(17 * time.Hour),
			Notes:         "Data analysis for quarterly report",
			State:         EntryStateApproved.String(),
		},
		{
			EmployeeID:    employees[2].ID,
			BillingCodeID: billingCodes[2].ID,
			Start:         today.Add(10 * time.Hour),
			End:           today.Add(15 * time.Hour),
			Notes:         "Machine learning model training",
			State:         EntryStateApproved.String(),
		},
		{
			EmployeeID:    employees[3].ID,
			BillingCodeID: billingCodes[4].ID,
			Start:         today.Add(9 * time.Hour),
			End:           today.Add(12 * time.Hour),
			Notes:         "Internal team operations",
			State:         EntryStateApproved.String(),
			Internal:      true,
		},

		// Impersonated entries - Nate impersonating John
		{
			EmployeeID:          employees[0].ID,                // Nate
			ImpersonateAsUserID: createUintPtr(employees[3].ID), // John
			BillingCodeID:       billingCodes[0].ID,
			Start:               yesterday.Add(9 * time.Hour),
			End:                 yesterday.Add(11 * time.Hour),
			Notes:               "Client data pipeline maintenance (impersonated)",
			State:               EntryStateApproved.String(),
		},

		// David impersonating Jane
		{
			EmployeeID:          employees[2].ID,                // David
			ImpersonateAsUserID: createUintPtr(employees[4].ID), // Jane
			BillingCodeID:       billingCodes[3].ID,
			Start:               today.Add(13 * time.Hour),
			End:                 today.Add(17 * time.Hour),
			Notes:               "Data strategy session with client (impersonated)",
			State:               EntryStateApproved.String(),
		},

		// Kevin impersonating David
		{
			EmployeeID:          employees[1].ID,                // Kevin
			ImpersonateAsUserID: createUintPtr(employees[2].ID), // David
			BillingCodeID:       billingCodes[2].ID,
			Start:               yesterday.Add(14 * time.Hour),
			End:                 yesterday.Add(18 * time.Hour),
			Notes:               "Machine learning algorithm optimization (impersonated)",
			State:               EntryStateApproved.String(),
		},

		// Internal impersonated entry
		{
			EmployeeID:          employees[0].ID,                // Nate
			ImpersonateAsUserID: createUintPtr(employees[4].ID), // Jane
			BillingCodeID:       billingCodes[4].ID,
			Start:               today.Add(14 * time.Hour),
			End:                 today.Add(16 * time.Hour),
			Notes:               "Internal operations review (impersonated)",
			State:               EntryStateApproved.String(),
			Internal:            true,
		},
	}
	_ = a.DB.Create(&entries)
}
