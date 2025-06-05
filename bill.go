package cronos

import (
	"cloud.google.com/go/storage"
	"context"
	"fmt"
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
	if err := writer.Close(); err != nil {
		return err
	}

	// Objects remain private; update cache settings
	attrsToUpdate := storage.ObjectAttrsToUpdate{
		CacheControl: "no-cache, max-age=0, must-revalidate",
	}
	if _, err := bucket.Object(objectName).Update(ctx, attrsToUpdate); err != nil {
		return fmt.Errorf("failed to update object: %w", err)
	}
	// save the invoice URL to the database
	bill.GCSFile = "https://storage.googleapis.com/" + a.Bucket + "/" + objectName
	a.DB.Save(&bill)
	return nil
}

// RegeneratePDF regenerates the PDF for a bill, we will call this when a bill is updated
func (a *App) RegeneratePDF(bill *Bill) error {
	if err := a.SaveBillToGCS(bill); err != nil {
		return err
	}
	return nil
}
