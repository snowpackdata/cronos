package cronos

import (
	"bytes"
	"html/template"
	"log"
	"os"

	"github.com/pkg/errors"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type EmailType string

const (
	EmailTemplateFolder                   = "./templates/emails"
	EmailTypeRegisterClient     EmailType = "register_client.html"
	EmailTypeRegisterStaff      EmailType = "register_staff.html"
	EmailTypeResetPassword      EmailType = "reset_password.html"
	EmailTypeChangePassword     EmailType = "change_password.html"
	EmailTypeWelcome            EmailType = "welcome.html"
	EmailTypeSurveyConfirmation EmailType = "survey_confirmation.html"

	SNOWPACK_SENDER_ADDRESS string = "no-reply@snowpack-data.com"
)

func (s EmailType) String() string {
	return string(s)
}

type Email struct {
	// Email is a non-persistent object that is used to store the email
	// information for a user across the application
	SenderName       string `default:"Snowpack Data"`
	SenderEmail      string `default:"no-reply@snowpack-data.com"`
	RecipientName    string
	RecipientEmail   string
	Subject          string
	PlainTextContent string
	htmlFile         string
}

func (e *Email) HTMLContent() string {
	filename := EmailTemplateFolder + "/" + e.htmlFile
	t, err := template.New(e.htmlFile).ParseFiles(filename)
	if err != nil {
		log.Println("Error parsing template file")
		log.Println(err)
	}
	var tpl bytes.Buffer
	if err = t.ExecuteTemplate(&tpl, e.htmlFile, ""); err != nil {
		log.Println("Error executing template file")
		log.Println(err)
	}
	return tpl.String()

}

func (a *App) EmailFromAdmin(emailType EmailType, address string) error {
	var email Email
	switch emailType {
	case EmailTypeRegisterClient:
		email = Email{
			SenderName:     "Snowpack Data",
			SenderEmail:    SNOWPACK_SENDER_ADDRESS,
			RecipientEmail: address,
			Subject:        "Welcome to Snowpack Data",
			htmlFile:       emailType.String(),
		}
	case EmailTypeRegisterStaff:
		email = Email{
			SenderName:     "Snowpack Data",
			SenderEmail:    SNOWPACK_SENDER_ADDRESS,
			RecipientEmail: address,
			Subject:        "Welcome to Snowpack Data",
			htmlFile:       emailType.String(),
		}
	case EmailTypeSurveyConfirmation:
		email = Email{
			SenderName:     "Snowpack Data",
			SenderEmail:    SNOWPACK_SENDER_ADDRESS,
			RecipientEmail: address,
			Subject:        "We Received your Survey Response",
			htmlFile:       emailType.String(),
		}
	}
	from := mail.NewEmail(email.SenderName, email.SenderEmail)
	to := mail.NewEmail(email.RecipientName, email.RecipientEmail)
	message := mail.NewSingleEmail(from, email.Subject, to, email.PlainTextContent, email.HTMLContent())
	client := sendgrid.NewSendClient(os.Getenv("SENDGRID_API_KEY"))
	response, err := client.Send(message)
	if err != nil {
		return errors.Wrap(err, "error sending email")
	}
	log.Println(response.StatusCode)
	return nil
}

func (a *App) SendTextEmail(email Email) error {
	from := mail.NewEmail(email.SenderName, email.SenderEmail)
	to := mail.NewEmail(email.RecipientName, email.RecipientEmail)
	message := mail.NewSingleEmail(from, email.Subject, to, email.PlainTextContent, email.PlainTextContent)
	client := sendgrid.NewSendClient(os.Getenv("SENDGRID_API_KEY"))
	response, err := client.Send(message)
	if err != nil {
		return errors.Wrap(err, "error sending email")
	}
	log.Println(response.StatusCode)
	return nil
}
