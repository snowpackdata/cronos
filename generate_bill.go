package cronos

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/jung-kurt/gofpdf"
)

// Set Constants for the invoice PDF

func (a *App) GenerateBillPDF(bill *Bill) []byte {
	// From the Invoice we need to generate the following:
	// 1. An overall invoice summary grouped at the billing code level
	// 2. A detailed invoice summary grouped at the individual entry level
	// 3. Associated Billable/Client Information
	// 4. Summary of Totals and Subtotals

	//1. Overall invoice summary grouped at the billing code level
	lineItems := a.GetBillLineItems(bill)
	//entryItems := a.GetInvoiceEntries(invoice)
	//adjustments := a.GetInvoiceAdjustments(invoice)

	// Load commissions for this bill
	var commissions []Commission
	a.DB.Where("bill_id = ?", bill.ID).Find(&commissions)

	var employee Employee
	a.DB.Preload("User").Where("id = ?", bill.EmployeeID).First(&employee)

	// Get the tenant's owner account for "From" information
	var ownerAccount Account
	a.DB.Preload("LogoAsset").Where("tenant_id = ? AND type = ?", bill.TenantID, AccountTypeInternal.String()).First(&ownerAccount)

	// Use owner account details or fall back to defaults
	fromName := defaultFromName
	fromAddress := defaultFromAddress
	fromContact := defaultContact
	var logoPath string // No default logo

	if ownerAccount.ID != 0 {
		if ownerAccount.LegalName != "" {
			fromName = ownerAccount.LegalName
		} else if ownerAccount.Name != "" {
			fromName = ownerAccount.Name
		}
		if ownerAccount.Address != "" {
			fromAddress = ownerAccount.Address
		}
		if ownerAccount.Email != "" {
			fromContact = ownerAccount.Email
		}
		// Check if custom logo exists
		if ownerAccount.LogoAsset != nil && ownerAccount.LogoAsset.GCSObjectPath != nil {
			// Download logo from GCS to temp file
			ctx := context.Background()
			storageClient := a.InitializeStorageClient(a.Project, *ownerAccount.LogoAsset.BucketName)
			bucket := storageClient.Bucket(*ownerAccount.LogoAsset.BucketName)

			rc, err := bucket.Object(*ownerAccount.LogoAsset.GCSObjectPath).NewReader(ctx)
			if err == nil {
				defer rc.Close()

				// Create temp file
				tmpFile, err := os.CreateTemp("", "logo-*"+filepath.Ext(*ownerAccount.LogoAsset.GCSObjectPath))
				if err == nil {
					defer os.Remove(tmpFile.Name())
					defer tmpFile.Close()

					// Copy logo to temp file
					if _, err := io.Copy(tmpFile, rc); err == nil {
						logoPath = tmpFile.Name()
					}
				}
			}
		}
	}

	InvoiceNumber := strconv.Itoa(time.Now().Year()) + "00" + strconv.Itoa(int(bill.ID))

	// Initialize the PDF document with set margins and add a page that we can work with
	pdf := gofpdf.New("P", "mm", "Letter", "")
	pdf.SetMargins(marginX, marginY, marginX)
	pdf.AddPage()
	pageW, _ := pdf.GetPageSize()
	safeAreaW := pageW - 2*marginX

	// Build the header - add logo only if custom logo exists
	if logoPath != "" {
		ext := strings.ToLower(filepath.Ext(logoPath))
		if ext == ".svg" {
			// Parse SVG file
			svgBasic, err := gofpdf.SVGBasicFileParse(logoPath)
			if err == nil {
				// Calculate scale to fit in 30x30 box
				scale := 30.0 / svgBasic.Wd
				if svgBasic.Ht*scale > 30.0 {
					scale = 30.0 / svgBasic.Ht
				}
				pdf.SVGBasicWrite(&svgBasic, scale)
			}
		} else {
			// Determine image type from file extension for raster images
			imageType := "PNG"
			if ext == ".jpg" || ext == ".jpeg" {
				imageType = "JPG"
			}
			pdf.ImageOptions(logoPath, 10, 0, 30, 30, false, gofpdf.ImageOptions{ImageType: imageType, ReadDpi: true}, 0, "")
		}
	}
	pdf.SetFont(defaultFont, "B", 16)
	_, lineHeight := pdf.GetFontSize()
	currentY := pdf.GetY() + lineHeight + gapY
	pdf.SetXY(marginX, currentY)
	pdf.Cell(headerWidth, headerHeight, fromName)

	leftY := pdf.GetY() + lineHeight + gapY

	// Build invoice word on right
	pdf.SetFont(defaultFont, "B", 14)
	_, lineHeight = pdf.GetFontSize()
	pdf.SetXY(80, currentY-lineHeight)
	pdf.MultiCell(120, 10, "Payroll Bill\n "+bill.Name, "0", "R", false)

	newY := leftY
	if (pdf.GetY() + gapY) > newY {
		newY = pdf.GetY() + gapY
	}

	newY += 10.0 // Add margin

	pdf.SetXY(marginX, newY)
	pdf.SetFont(defaultFont, "", 12)
	_, lineHeight = pdf.GetFontSize()
	lineBreak := lineHeight + float64(1)

	// Left hand info
	splittedFromAddress := breakAddress(fromAddress)
	for _, add := range splittedFromAddress {
		pdf.Cell(safeAreaW/2, lineHeight, add)
		pdf.Ln(lineBreak)
	}
	pdf.SetFontStyle("I")
	pdf.Cell(safeAreaW/2, lineHeight, fromContact)
	pdf.Ln(lineBreak)
	pdf.Ln(lineBreak)
	pdf.Ln(lineBreak)

	pdf.SetFontStyle("B")
	pdf.Cell(safeAreaW/2, lineHeight, "Payable To:")
	pdf.Line(marginX, pdf.GetY()+lineHeight, marginX+safeAreaW/2, pdf.GetY()+lineHeight)
	pdf.Ln(lineBreak)
	pdf.Cell(safeAreaW/2, lineHeight, employee.FirstName+" "+employee.LastName)
	pdf.SetFontStyle("")
	pdf.Ln(lineBreak)
	pdf.SetFontStyle("I")
	pdf.Cell(safeAreaW/2, lineHeight, fmt.Sprintf(employee.User.Email))

	endOfInvoiceDetailY := pdf.GetY() + lineHeight
	pdf.SetFontStyle("")

	// Right hand side info, invoice no & invoice date
	invoiceDetailW := float64(30)
	pdf.SetXY(safeAreaW/2+30, newY)
	pdf.Cell(invoiceDetailW, lineHeight, "Bill No:")
	pdf.Cell(invoiceDetailW, lineHeight, InvoiceNumber)
	pdf.Ln(lineBreak)
	pdf.SetX(safeAreaW/2 + 30)
	pdf.Cell(invoiceDetailW, lineHeight, "Issued Date:")
	pdf.Cell(invoiceDetailW, lineHeight, bill.CreatedAt.UTC().Format("01/02/2006"))
	pdf.Ln(lineBreak)
	pdf.SetX(safeAreaW/2 + 30)
	pdf.Cell(invoiceDetailW, lineHeight, "Due Date:")
	pdf.Cell(invoiceDetailW, lineHeight, bill.CreatedAt.AddDate(0, 1, 0).UTC().Format("01/02/2006"))
	pdf.Ln(lineBreak)

	// Draw the table
	pdf.SetFontSize(10.0)
	pdf.SetXY(marginX, endOfInvoiceDetailY+10.0)
	lineHt := 10.0
	const colNumber = 5
	header := [colNumber]string{"Code", "Description", "Hours", "Rate", "Total ($)"}
	colWidth := [colNumber]float64{40.0, 80.0, 25.0, 25.0, 30.0}

	// Headers
	pdf.SetFontStyle("B")
	pdf.SetFillColor(200, 200, 200)
	for colJ := 0; colJ < colNumber; colJ++ {
		pdf.CellFormat(colWidth[colJ], lineHt, header[colJ], "1", 0, "CM", true, 0, "")
	}

	pdf.Ln(-1)
	pdf.SetFillColor(255, 255, 255)

	// Table data
	pdf.SetFontStyle("")

	for rowJ := 0; rowJ < len(lineItems); rowJ++ {
		val := lineItems[rowJ]
		billingCode := val.BillingCodeCode
		description := val.BillingCode
		hours := val.HoursFormatted
		rate := val.RateFormatted
		total := fmt.Sprintf("%.2f", val.Total)

		pdf.CellFormat(colWidth[0], lineHt, billingCode, "1", 0, "LM", false, 0, "")
		pdf.CellFormat(colWidth[1], lineHt, description, "1", 0, "LM", false, 0, "")
		pdf.CellFormat(colWidth[2], lineHt, hours, "1", 0, "CM", false, 0, "")
		pdf.CellFormat(colWidth[3], lineHt, "$ "+rate, "1", 0, "LM", false, 0, "")
		pdf.CellFormat(colWidth[4], lineHt, "$ "+total, "1", 0, "RM", false, 0, "")
		pdf.Ln(-1)
	}

	// Add commissions to the bill if any exist
	if len(commissions) > 0 {
		for _, commission := range commissions {
			// Format commission amount
			commissionAmount := fmt.Sprintf("%.2f", float64(commission.Amount)/100)

			// Calculate commission rate as percentage
			commissionRate := a.CalculateCommissionRate(commission.Role, commission.ProjectType, 0) // Using 0 as deal size since we just want to display the rate
			rateFormatted := fmt.Sprintf("%.1f%%", commissionRate*100)

			// Determine project type display text (shorter version)
			projectTypeText := "New"
			if commission.ProjectType == ProjectTypeExisting.String() {
				projectTypeText = "Existing"
			}

			// Determine role display text (abbreviated)
			roleText := "AE"
			if commission.Role == CommissionRoleSDR.String() {
				roleText = "SDR"
			}

			// Create concise commission description with invoice reference
			// Format invoice number as "YYYYNNNN" where NNNN is the project ID (zero-padded to 4 digits)
			invoiceRef := fmt.Sprintf("#%d%04d", time.Now().Year(), commission.ProjectID)
			description := fmt.Sprintf("%s %s - %s (%s)", commission.ProjectName, invoiceRef, roleText, projectTypeText)

			pdf.CellFormat(colWidth[0], lineHt, "COMM", "1", 0, "LM", false, 0, "")
			pdf.CellFormat(colWidth[1], lineHt, description, "1", 0, "LM", false, 0, "")
			pdf.CellFormat(colWidth[2], lineHt, "-", "1", 0, "CM", false, 0, "")
			pdf.CellFormat(colWidth[3], lineHt, rateFormatted, "1", 0, "LM", false, 0, "")
			pdf.CellFormat(colWidth[4], lineHt, "$ "+commissionAmount, "1", 0, "RM", false, 0, "")
			pdf.Ln(-1)
		}
	}

	// Generate the total Rows

	// Calculate the subtotal
	pdf.SetFontStyle("B")
	leftIndent := 0.0
	for i := 0; i < 3; i++ {
		leftIndent += colWidth[i]
	}

	grandTotal := fmt.Sprintf("%.2f", float64(bill.TotalAmount)/100)
	pdf.SetX(marginX + leftIndent)
	pdf.CellFormat(colWidth[3], lineHt, "Total Due", "1", 0, "LM", false, 0, "")
	pdf.CellFormat(colWidth[4], lineHt, "$ "+grandTotal, "1", 0, "RM", false, 0, "")
	pdf.Ln(lineHt)

	pdf.SetFontStyle("")
	pdf.Ln(lineBreak)

	var buffer bytes.Buffer
	err := pdf.Output(&buffer)
	if err != nil {
		fmt.Println(err)
	}
	return buffer.Bytes()
}
