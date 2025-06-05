package cronos

import (
	"github.com/pkg/errors"
)

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
	hashed, err := hashPassword(DEFAULT_PASSWORD)
	if err != nil {
		return err
	}
	user.Password = hashed
	user.Role = UserRoleClient.String()
	user.AccountID = accountId
	a.DB.Save(&user)
	// Send the email to the user
	sendErr := a.EmailFromAdmin(EmailTypeRegisterClient, email)
	return sendErr
}

func (a *App) RegisterStaff(email string, accountId uint) error {
	// Create a blank user and client if they don't already exist
	user := User{Email: email}
	a.DB.Save(&user)
	a.DB.Save(&Employee{User: user})
	hashed, err := hashPassword(DEFAULT_PASSWORD)
	if err != nil {
		return err
	}
	user.Password = hashed
	user.Role = UserRoleStaff.String()
	user.AccountID = accountId
	a.DB.Save(&user)
	// Send the email to the user
	sendErr := a.EmailFromAdmin(EmailTypeRegisterStaff, email)
	return sendErr
}
