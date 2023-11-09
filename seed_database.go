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
	nateUser := User{
		Username: "nater",
		Email:    "nate@snowpack-data.com",
		IsAdmin:  true,
		Role:     UserRoleAdmin.String(),
	}
	nateEmployee := Employee{
		User:      nateUser,
		Title:     "Partner",
		FirstName: "Nate",
		LastName:  "Robinson",
	}
	kevinUser := User{
		Username: "kevinK",
		Email:    "kevin@snowpack-data.com",
		IsAdmin:  true,
		Role:     UserRoleStaff.String(),
	}
	kevinEmployee := Employee{
		User:      kevinUser,
		Title:     "Partner",
		FirstName: "Kevin",
		LastName:  "Koenitzer",
	}
	davidUser := User{
		Username: "kevinK",
		Email:    "kevin@snowpack-data.com",
		IsAdmin:  false,
		Role:     UserRoleStaff.String(),
	}
	davidEmployee := Employee{
		User:      davidUser,
		Title:     "Partner",
		FirstName: "David",
		LastName:  "Shore",
	}

	// Create Accounts for Nate & Kevin
	_ = a.DB.Create([]Employee{nateEmployee, kevinEmployee, davidEmployee})
	_ = a.DB.Create([]User{nateUser, kevinUser, davidUser})

	// Create initial Company for Snowpack
	snowpack := Account{
		Name:      "Snowpack Data",
		Type:      AccountTypeInternal.String(),
		LegalName: "Snowpack Data, LLC",
		Email:     "billing@snowpack-data.com",
		Website:   "https://snowpack-data.com",
		Admin:     nateUser,
		Clients:   []User{nateUser, kevinUser, davidUser},
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
