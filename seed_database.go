package cronos

import "time"

const AccountsReceivable = "SNOWPACK_ACCOUNTS_RECEIVABLE"
const AccountsPayable = "SNOWPACK_ACCOUNTS_PAYABLE"
const CostOfGoodsSold = "SNOWPACK_COGS"

func (a *App) SeedDatabase() {
	journals := []Journal{
		Journal{Name: AccountsReceivable},
		Journal{Name: AccountsPayable},
		Journal{Name: CostOfGoodsSold},
	}
	_ = a.DB.Create(&journals)

	// Create Users
	users := []User{
		User{
			Email:    "nate@snowpack-data.com",
			IsAdmin:  true,
			Role:     UserRoleAdmin.String(),
			Password: DEFAULT_PASSWORD,
		},
		User{
			Email:    "kevin@snowpack-data.com",
			IsAdmin:  true,
			Role:     UserRoleStaff.String(),
			Password: DEFAULT_PASSWORD,
		},
		User{
			Email:    "david@snowpack-data.com",
			IsAdmin:  false,
			Role:     UserRoleStaff.String(),
			Password: DEFAULT_PASSWORD,
		},
	}
	_ = a.DB.Create(&users)

	employees := []Employee{
		Employee{
			User:      users[0],
			Title:     "Partner",
			FirstName: "Nate",
			LastName:  "Robinson",
			IsActive:  true,
			StartDate: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		Employee{
			User:      users[1],
			Title:     "Partner",
			FirstName: "Kevin",
			LastName:  "Koenitzer",
			IsActive:  true,
			StartDate: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		Employee{
			User:      users[2],
			Title:     "Partner",
			FirstName: "David",
			LastName:  "Shore",
			IsActive:  true,
			StartDate: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
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
		Clients:   users,
	}
	_ = a.DB.Create(&snowpack)
	a.DB.Save(&snowpack)

}
