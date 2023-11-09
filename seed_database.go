package main

import "time"

const ACCOUNTS_RECEIVABLE = "SNOWPACK_ACCOUNTS_RECEIVABLE"
const ACCOUNTS_PAYABLE = "SNOWPACK_ACCOUNTS_PAYABLE"

func (a *App) SeedDatabase() {
	journals := []Journal{
		Journal{Name: ACCOUNTS_RECEIVABLE},
		Journal{Name: ACCOUNTS_PAYABLE},
	}
	_ = a.DB.Create(&journals)

	// Create Users
	users := []User{
		User{
			Username: "nater",
			Email:    "nate@snowpack-data.com",
			IsAdmin:  true,
			Role:     UserRoleAdmin.String(),
		},
		User{
			Username: "kevink",
			Email:    "kevin@snowpack-data.com",
			IsAdmin:  true,
			Role:     UserRoleStaff.String(),
		},
		User{
			Username: "davids",
			Email:    "kevin@snowpack-data.com",
			IsAdmin:  false,
			Role:     UserRoleStaff.String(),
		},
	}
	_ = a.DB.Create(&users)

	employees := []Employee{
		Employee{
			User:      users[0],
			Title:     "Partner",
			FirstName: "Nate",
			LastName:  "Robinson",
		},
		Employee{
			User:      users[1],
			Title:     "Partner",
			FirstName: "Kevin",
			LastName:  "Koenitzer",
		},
		Employee{
			User:      users[2],
			Title:     "Partner",
			FirstName: "David",
			LastName:  "Shore",
		},
	}
	_ = a.DB.Create(&employees)

	// Create initial Company for Snowpack
	snowpack := Account{
		Name:      "Snowpack Data",
		Type:      AccountTypeInternal.String(),
		LegalName: "Snowpack Data, LLC",
		Email:     "billing@snowpack-data.com",
		Website:   "https://snowpack-data.com",
		Admin:     users[0],
		Clients:   users,
	}
	_ = a.DB.Create(&snowpack)

	// Create initial set of internal rates
	snowpackRates := []Rate{
		Rate{
			Name:         "Partner Standard",
			Amount:       250.00,
			ActiveFrom:   time.Now(),
			ActiveTo:     time.Now().AddDate(2, 0, 0),
			InternalOnly: false,
		},
		Rate{
			Name:         "Partner Discounted",
			Amount:       225.00,
			ActiveFrom:   time.Now(),
			ActiveTo:     time.Now().AddDate(2, 0, 0),
			InternalOnly: false,
		},
		Rate{
			Name:         "Staff Standard",
			Amount:       175.00,
			ActiveFrom:   time.Now(),
			ActiveTo:     time.Now().AddDate(2, 0, 0),
			InternalOnly: false,
		},
		Rate{
			Name:         "Staff Discounted",
			Amount:       125.00,
			ActiveFrom:   time.Now(),
			ActiveTo:     time.Now().AddDate(2, 0, 0),
			InternalOnly: false,
		},
	}
	_ = a.DB.Create(&snowpackRates)
	// Create initial set of internal billing codes
	snowpackBillingCodes := []BillingCode{
		BillingCode{
			Name:        "Admin Non-Billable - Partner",
			RateType:    RateTypeInternalAdminNonBillable.String(),
			Category:    "Administrative",
			Code:        "ADMIN_0000",
			RoundedTo:   15,
			ActiveStart: time.Now(),
			ActiveEnd:   time.Now().AddDate(2, 0, 0),
			Internal:    true,
			Rate:        snowpackRates[0],
		},
		BillingCode{
			Name:        "Admin Non-Billable - Staff",
			RateType:    RateTypeInternalAdminNonBillable.String(),
			Category:    "Administrative",
			Code:        "ADMIN_0001",
			RoundedTo:   15,
			ActiveStart: time.Now(),
			ActiveEnd:   time.Now().AddDate(2, 0, 0),
			Internal:    true,
			Rate:        snowpackRates[2],
		},
		BillingCode{
			Name:        "Project Work - Partner",
			RateType:    RateTypeInternalProject.String(),
			Category:    "Project Work",
			Code:        "PROJ_0000",
			RoundedTo:   15,
			ActiveStart: time.Now(),
			ActiveEnd:   time.Now().AddDate(2, 0, 0),
			Internal:    true,
			Rate:        snowpackRates[0],
		},
		BillingCode{
			Name:        "Project Work - Staff",
			RateType:    RateTypeInternalProject.String(),
			Category:    "Project Work",
			Code:        "PROJ_0001",
			RoundedTo:   15,
			ActiveStart: time.Now(),
			ActiveEnd:   time.Now().AddDate(2, 0, 0),
			Internal:    true,
			Rate:        snowpackRates[2],
		},
	}
	_ = a.DB.Create(&snowpackBillingCodes)

	// Create initial set of internal projects
	snowpackProjects := []Project{
		Project{
			Name:          "Snowpack Website",
			Account:       snowpack,
			ActiveStart:   time.Now(),
			ActiveEnd:     time.Now().AddDate(2, 0, 0),
			BudgetHours:   0,
			BudgetDollars: 0,
			Internal:      true,
			BillingCodes:  []BillingCode{snowpackBillingCodes[2], snowpackBillingCodes[3]},
		},
		Project{
			Name:          "Project Cronos",
			Account:       snowpack,
			ActiveStart:   time.Now(),
			ActiveEnd:     time.Now().AddDate(2, 0, 0),
			BudgetHours:   0,
			BudgetDollars: 10000,
			Internal:      true,
			BillingCodes:  []BillingCode{snowpackBillingCodes[2], snowpackBillingCodes[3]},
		},
		Project{
			Name:          "Snowpack Admin",
			Account:       snowpack,
			ActiveStart:   time.Now(),
			ActiveEnd:     time.Now().AddDate(2, 0, 0),
			BudgetHours:   0,
			BudgetDollars: 0,
			Internal:      true,
			BillingCodes:  []BillingCode{snowpackBillingCodes[0], snowpackBillingCodes[1]},
		},
	}
	_ = a.DB.Create(&snowpackProjects)
	snowpack.Projects = snowpackProjects
	a.DB.Save(&snowpack)

}
