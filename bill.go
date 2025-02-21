package cronos

import (
	"cloud.google.com/go/storage"
	"context"
	"crypto/rsa"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2/google"
	"log"
	"time"
)

// SaveBillToGCS saves the invoice to GCS
func (a *App) SaveBillToGCS(bill *Bill) error {
	ctx := context.Background()
	// Generate the invoice
	// The output must be stored as a list of bytes in-memory because of the readonly filesystem in GAE
	pdfBytes := a.GenerateBillPDF(bill)
	// Save the invoice to GCS
	client := a.InitializeStorageClient(a.Project, a.Bucket)

	// Create a bucket handle
	bucket := client.Bucket(a.Bucket)
	// Create a new object and write its contents to the bucket
	filename := bill.GetBillFilename() + ".pdf"
	objectName := "bills/" + filename
	writer := bucket.Object(objectName).NewWriter(ctx)
	if _, err := writer.Write(pdfBytes); err != nil {
		return err
	}
	writer.Close()

	// Set the object to be publicly accessible
	acl := bucket.Object(objectName).ACL()
	if err := acl.Set(ctx, storage.AllUsers, storage.RoleReader); err != nil {
		return err
	}
	attrsToUpdate := storage.ObjectAttrsToUpdate{
		CacheControl: "no-cache, max-age=0, must-revalidate",
	}
	if _, err := bucket.Object(objectName).Update(ctx, attrsToUpdate); err != nil {
		log.Fatalf("Failed to update object: %v", err)
	}
	// save the public invoice URL to the database
	bill.GCSFile = "https://storage.googleapis.com/" + a.Bucket + "/" + objectName
	a.DB.Save(&bill)
	return nil
}

// RegeneratePDF regenerates the PDF for a bill, we will call this when a bill is updated
func (a *App) RegeneratePDF(bill *Bill) error {
	err := a.SaveBillToGCS(bill)
	if err != nil {
		panic(err)
	}
	return nil
}
