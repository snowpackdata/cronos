package cronos

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"html/template"
	"io"
	"log"
	"os"
	"regexp"
	"strings"

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

// generateInvoiceFilename creates a clean, descriptive filename for the invoice PDF
func generateInvoiceFilename(invoice *Invoice) string {
	// Get account name or use "Invoice" as default
	accountName := "Invoice"
	if invoice.Account.Name != "" {
		accountName = invoice.Account.Name
	} else if invoice.Project.Name != "" {
		accountName = invoice.Project.Name
	}
	
	// Sanitize account name: remove special characters, replace spaces with underscores
	reg := regexp.MustCompile(`[^a-zA-Z0-9-]+`)
	cleanAccountName := reg.ReplaceAllString(accountName, "_")
	cleanAccountName = strings.Trim(cleanAccountName, "_")
	
	// Format dates
	periodStart := invoice.PeriodStart.Format("2006-01-02")
	periodEnd := invoice.PeriodEnd.Format("2006-01-02")
	
	// Generate filename: invoice_123456_ClientName_2025-01-01_2025-01-31.pdf
	filename := fmt.Sprintf("invoice_%06d_%s_%s_%s.pdf", 
		invoice.ID, 
		cleanAccountName, 
		periodStart, 
		periodEnd)
	
	return filename
}

// SendInvoiceEmail sends an invoice email with HTML content, CC, and PDF attachment
func (a *App) SendInvoiceEmail(to, cc, subject, htmlBody string, pdfURL string, invoice *Invoice) error {
	from := mail.NewEmail("Snowpack Data", "billing@snowpack-data.com")
	toEmail := mail.NewEmail("", to)
	
	// Create the message
	message := mail.NewV3Mail()
	message.SetFrom(from)
	message.Subject = subject
	
	// Add personalization (to and cc)
	p := mail.NewPersonalization()
	p.AddTos(toEmail)
	
	// Add CC recipients if provided
	if cc != "" {
		// Split by comma for multiple CC addresses
		ccAddresses := bytes.Split([]byte(cc), []byte(","))
		for _, ccAddr := range ccAddresses {
			trimmedCC := bytes.TrimSpace(ccAddr)
			if len(trimmedCC) > 0 {
				p.AddCCs(mail.NewEmail("", string(trimmedCC)))
			}
		}
	}
	
	message.AddPersonalizations(p)
	
	// Add HTML content
	content := mail.NewContent("text/html", htmlBody)
	message.AddContent(content)
	
	// Download and attach the PDF from GCS
	if pdfURL != "" {
		// Parse the GCS URL to extract bucket and object path
		// URL format: https://storage.googleapis.com/bucket-name/path/to/file.pdf
		if bytes.Contains([]byte(pdfURL), []byte("storage.googleapis.com/")) {
			parts := bytes.SplitN([]byte(pdfURL), []byte("storage.googleapis.com/"), 2)
			if len(parts) == 2 {
				pathParts := bytes.SplitN(parts[1], []byte("/"), 2)
				if len(pathParts) == 2 {
					bucketName := string(pathParts[0])
					objectPath := string(pathParts[1])
					
					// Download the PDF from GCS
					ctx := context.Background()
					storageClient := a.InitializeStorageClient(a.Project, bucketName)
					bucket := storageClient.Bucket(bucketName)
					
					rc, err := bucket.Object(objectPath).NewReader(ctx)
					if err != nil {
						log.Printf("Error reading PDF from GCS: %v", err)
					} else {
						defer rc.Close()
						
						// Read the PDF bytes
						pdfBytes, err := io.ReadAll(rc)
						if err != nil {
							log.Printf("Error reading PDF bytes: %v", err)
						} else {
							// Base64 encode and attach
							encodedPDF := base64.StdEncoding.EncodeToString(pdfBytes)
							
							// Generate descriptive filename: invoice_123456_ClientName_2025-01-01_2025-01-31.pdf
							filename := generateInvoiceFilename(invoice)
							
							attachment := mail.NewAttachment()
							attachment.SetContent(encodedPDF)
							attachment.SetType("application/pdf")
							attachment.SetFilename(filename)
							attachment.SetDisposition("attachment")
							
							message.AddAttachment(attachment)
						}
					}
				}
			}
		}
	}
	
	// Send the email
	client := sendgrid.NewSendClient(os.Getenv("SENDGRID_API_KEY"))
	response, err := client.Send(message)
	if err != nil {
		return errors.Wrap(err, "error sending invoice email")
	}
	
	if response.StatusCode >= 400 {
		return errors.New("SendGrid returned error status: " + response.Body)
	}
	
	return nil
}
