package cronos

import "github.com/pkg/errors"

var ErrUserAlreadyExists = errors.New("user already exists")

// RegisterClient performs the heavy lifting of registering a client before a customer
// ever interacts with the website. It will first create a blank User and Client object
// for the client before emailing them to complete the registration process.
func (a *App) RegisterClient(email string, accountId uint) error {
	// Create a blank user and client if they don't already exist
	user := User{Email: email}
	if a.DB.Model(&user).Where("email = ?", email).Updates(&user).RowsAffected == 0 {
		a.DB.Create(&user)
	} else {
		return ErrUserAlreadyExists
	}
	user.Password = DEFAULT_PASSWORD
	user.Role = UserRoleClient.String()
	user.AccountID = accountId
	a.DB.Save(&user)
	// Send the email to the user
	err := a.EmailFromAdmin(EmailTypeRegisterClient, email)
	return err
}

func (a *App) RegisterStaff(email string, accountId uint) error {
	// Create the user record first
	user := User{Email: email}
	a.DB.Save(&user)

	// Link the newly created user to an employee record using the user's ID
	employee := Employee{UserID: user.ID}
	a.DB.Create(&employee)

	user.Password = DEFAULT_PASSWORD
	user.Role = UserRoleStaff.String()
	user.AccountID = accountId
	a.DB.Save(&user)
	// Send the email to the user
	err := a.EmailFromAdmin(EmailTypeRegisterStaff, email)
	return err
}
