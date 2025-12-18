package cronos

import "github.com/pkg/errors"

var ErrUserAlreadyExists = errors.New("user already exists")

// RegisterClient performs the heavy lifting of registering a client before a customer
// ever interacts with the website. It will first create a blank User and Client object
// for the client before emailing them to complete the registration process.
func (a *App) RegisterClient(email string, accountId uint, adminName string, orgName string, tenantSlug string) error {
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
	
	// Prepare template data for email
	templateData := map[string]interface{}{
		"AdminName":   adminName,
		"OrgName":     orgName,
		"TenantSlug":  tenantSlug,
		"RegistrationURL": "https://" + tenantSlug + ".cronosplatform.com/register",
	}
	
	// Send the email to the user
	err := a.EmailFromAdmin(EmailTypeRegisterClient, email, templateData)
	return err
}

func (a *App) RegisterStaff(email string, accountId uint, adminName string, orgName string, tenantSlug string) error {
	// Create a blank user and client if they don't already exist
	user := User{Email: email}
	a.DB.Save(&user)
	a.DB.Save(&Employee{User: user})
	user.Password = DEFAULT_PASSWORD
	user.Role = UserRoleStaff.String()
	user.AccountID = accountId
	a.DB.Save(&user)
	
	// Prepare template data for email
	templateData := map[string]interface{}{
		"AdminName":   adminName,
		"OrgName":     orgName,
		"TenantSlug":  tenantSlug,
		"RegistrationURL": "https://" + tenantSlug + ".cronosplatform.com/register",
	}
	
	// Send the email to the user
	err := a.EmailFromAdmin(EmailTypeRegisterStaff, email, templateData)
	return err
}
