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
const defaultFont = "Helvetica"
const defaultFromName = "Snowpack Data LLC"
const defaultFromAddress = "2261 Market Street STE 22279, San Francisco CA 94114"
const defaultContact = "billing@snowpack-data.io"
const headerWidth = 40.0
const headerHeight = 10.0
const marginX = 10.0
const marginY = 20.0
const gapY = 2.0

func (a *App) GenerateInvoicePDF(invoice *Invoice) []byte {
	// From the Invoice we need to generate the following:
	// 1. An overall invoice summary grouped at the billing code level
	// 2. A detailed invoice summary grouped at the individual entry level
	// 3. Associated Billable/Client Information
	// 4. Summary of Totals and Subtotals

	// 1. Overall invoice summary grouped at the billing code level
	lineItems := a.GetInvoiceLineItems(invoice)
	entryItems := a.GetInvoiceEntries(invoice)
	adjustments := a.GetInvoiceAdjustments(invoice)
	
	// Load expenses with receipts for this invoice
	var expenses []Expense
	a.DB.Where("invoice_id = ?", invoice.ID).Preload("Receipt").Preload("Submitter").Find(&expenses)

	var project Project
	var account Account

	// Get account information - invoice now has direct AccountID
	a.DB.Where("id = ?", invoice.AccountID).First(&account)

	// If this is a project-specific invoice, get the project information
	if invoice.ProjectID != nil && *invoice.ProjectID != 0 {
		a.DB.Where("id = ?", *invoice.ProjectID).First(&project)
	}

	InvoiceNumber := strconv.Itoa(time.Now().Year()) + "00" + strconv.Itoa(int(invoice.ID))

	// Initialize the PDF document with set margins and add a page that we can work with
	pdf := gofpdf.New("P", "mm", "Letter", "")
	pdf.SetMargins(marginX, marginY, marginX)
	pdf.AddPage()
	pageW, _ := pdf.GetPageSize()
	safeAreaW := pageW - 2*marginX

	// Build the header
	pdf.ImageOptions("./branding/logo/logo-large-light.png", 10, 0, 30, 30, false, gofpdf.ImageOptions{ImageType: "PNG", ReadDpi: true}, 0, "")
	pdf.SetFont(defaultFont, "B", 16)
	_, lineHeight := pdf.GetFontSize()
	currentY := pdf.GetY() + lineHeight + gapY
	pdf.SetXY(marginX, currentY)
	pdf.Cell(headerWidth, headerHeight, defaultFromName)

	leftY := pdf.GetY() + lineHeight + gapY

	// Build invoice word on right
	pdf.SetFont(defaultFont, "B", 16)
	_, lineHeight = pdf.GetFontSize()
	pdf.SetXY(80, currentY-lineHeight)
	pdf.MultiCell(120, 10, "INVOICE\n "+invoice.Name, "0", "R", false)

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
	splittedFromAddress := breakAddress(defaultFromAddress)
	for _, add := range splittedFromAddress {
		pdf.Cell(safeAreaW/2, lineHeight, add)
		pdf.Ln(lineBreak)
	}
	pdf.SetFontStyle("I")
	pdf.Cell(safeAreaW/2, lineHeight, defaultContact)
	pdf.Ln(lineBreak)
	pdf.Ln(lineBreak)
	pdf.Ln(lineBreak)

	pdf.SetFontStyle("B")
	pdf.Cell(safeAreaW/2, lineHeight, "Bill To:")
	pdf.Line(marginX, pdf.GetY()+lineHeight, marginX+safeAreaW/2, pdf.GetY()+lineHeight)
	pdf.Ln(lineBreak)
	pdf.Cell(safeAreaW/2, lineHeight, account.LegalName)
	pdf.SetFontStyle("")
	pdf.Ln(lineBreak)
	splittedToAddress := breakAddress(account.Address)
	for _, add := range splittedToAddress {
		pdf.Cell(safeAreaW/2, lineHeight, add)
		pdf.Ln(lineBreak)
	}
	pdf.SetFontStyle("I")
	pdf.Cell(safeAreaW/2, lineHeight, fmt.Sprintf(account.Email))

	endOfInvoiceDetailY := pdf.GetY() + lineHeight
	pdf.SetFontStyle("")

	// Right hand side info, invoice no & invoice date
	invoiceDetailW := float64(30)
	pdf.SetXY(safeAreaW/2+30, newY)
	pdf.Cell(invoiceDetailW, lineHeight, "Invoice No:")
	pdf.Cell(invoiceDetailW, lineHeight, InvoiceNumber)
	pdf.Ln(lineBreak)
	pdf.SetX(safeAreaW/2 + 30)
	pdf.Cell(invoiceDetailW, lineHeight, "Issued Date:")
	pdf.Cell(invoiceDetailW, lineHeight, invoice.SentAt.UTC().Format("01/02/2006"))
	pdf.Ln(lineBreak)
	pdf.SetX(safeAreaW/2 + 30)
	pdf.Cell(invoiceDetailW, lineHeight, "Due Date:")
	pdf.Cell(invoiceDetailW, lineHeight, invoice.DueAt.UTC().Format("01/02/2006"))
	pdf.Ln(lineBreak)

	// Draw the table
	pdf.SetFontSize(10.0)
	pdf.SetXY(marginX, endOfInvoiceDetailY+10.0)
	lineHt := 10.0
	const colNumber = 5
	header := [colNumber]string{"Billing Code", "Project", "Hours", "Rate ($)", "Total ($)"}
	colWidth := [colNumber]float64{40.0, 70.0, 25.0, 25.0, 40.0}

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
		billingCode := val.BillingCode
		description := val.Project
		hours := val.HoursFormatted
		rate := val.RateFormatted
		total := fmt.Sprintf("%.2f", val.Hours*val.Rate)

		pdf.CellFormat(colWidth[0], lineHt, billingCode, "1", 0, "CM", true, 0, "")
		pdf.CellFormat(colWidth[1], lineHt, description, "1", 0, "LM", true, 0, "")
		pdf.CellFormat(colWidth[2], lineHt, hours, "1", 0, "CM", true, 0, "")
		pdf.CellFormat(colWidth[3], lineHt, "$ "+rate, "1", 0, "CM", true, 0, "")
		pdf.CellFormat(colWidth[4], lineHt, "$ "+total, "1", 0, "RM", true, 0, "")
		pdf.Ln(-1)
	}
	// add adjustments if they exist, text should be italic
	if len(adjustments) > 0 {
		for rowJ := 0; rowJ < len(adjustments); rowJ++ {
			pdf.SetFontStyle("I")
			val := adjustments[rowJ]
			var adjustmentType string
			var adjustmentMultiplier float64
			if val.Type == AdjustmentTypeCredit.String() {
				adjustmentType = "CREDIT"
				adjustmentMultiplier = -1.0
			} else {
				adjustmentType = "FEE"
				adjustmentMultiplier = 1.0
			}
			description := val.Notes
			amount := fmt.Sprintf("%.2f", val.Amount*adjustmentMultiplier)
			pdf.CellFormat(colWidth[0], lineHt, adjustmentType, "1", 0, "CM", true, 0, "")
			pdf.CellFormat(colWidth[1], lineHt, description, "1", 0, "LM", true, 0, "")
			pdf.CellFormat(colWidth[2], lineHt, "", "1", 0, "CM", true, 0, "")
			pdf.CellFormat(colWidth[3], lineHt, "", "1", 0, "LM", true, 0, "")
			pdf.CellFormat(colWidth[4], lineHt, "$ "+amount, "1", 0, "RM", true, 0, "")
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

	totalFees := fmt.Sprintf("%.2f", invoice.TotalFees)
	pdf.SetX(marginX + leftIndent)
	pdf.CellFormat(colWidth[3], lineHt, "Fees", "1", 0, "LM", true, 0, "")
	pdf.CellFormat(colWidth[4], lineHt, "$ "+totalFees, "1", 0, "RM", true, 0, "")
	pdf.Ln(lineHt)

	totalAdjustments := fmt.Sprintf("%.2f", invoice.TotalAdjustments)
	pdf.SetX(marginX + leftIndent)
	pdf.CellFormat(colWidth[3], lineHt, "Adjustments", "1", 0, "LM", true, 0, "")
	pdf.CellFormat(colWidth[4], lineHt, "$ "+totalAdjustments, "1", 0, "RM", true, 0, "")
	pdf.Ln(lineHt)
	
	// Add expenses total if there are any
	if len(expenses) > 0 {
		totalExpenses := fmt.Sprintf("%.2f", invoice.TotalExpenses)
		pdf.SetX(marginX + leftIndent)
		pdf.CellFormat(colWidth[3], lineHt, "Expenses", "1", 0, "LM", true, 0, "")
		pdf.CellFormat(colWidth[4], lineHt, "$ "+totalExpenses, "1", 0, "RM", true, 0, "")
		pdf.Ln(lineHt)
	}

	grandTotal := fmt.Sprintf("%.2f", invoice.TotalAmount)
	pdf.SetX(marginX + leftIndent)
	pdf.CellFormat(colWidth[3], lineHt, "Total Due", "1", 0, "LM", true, 0, "")
	pdf.CellFormat(colWidth[4], lineHt, "$ "+grandTotal, "1", 0, "RM", true, 0, "")
	pdf.Ln(lineHt)

	pdf.SetFontStyle("")
	pdf.Ln(lineBreak)

	// Update footer message based on whether there are expenses
	if len(expenses) > 0 {
		pdf.Cell(safeAreaW, lineHeight, "See following pages for detailed timesheet entry breakdown and expense receipts.")
	} else {
		pdf.Cell(safeAreaW, lineHeight, "See second page for detailed timesheet entry breakdown.")
	}

	// Add a second page for individual entries
	pdf.AddPage()
	pdf.SetFontStyle("B")
	pageW, _ = pdf.GetPageSize()
	safeAreaW = pageW - 2*marginX
	pdf.SetXY(marginX, marginY)

	// Draw the table

	const entryColNumber = 5
	const entryFontSize = 8.0
	entryHeader := [entryColNumber]string{"Date", "Billing Code", "Staff", "Description", "Hours"}
	entryColWidth := [colNumber]float64{25.0, 25.0, 25.0, 100.0, 25.0}

	// Headers
	pdf.SetFontStyle("B")
	pdf.SetFontSize(entryFontSize)
	pdf.SetFillColor(200, 200, 200)
	for colJ := 0; colJ < colNumber; colJ++ {
		pdf.CellFormat(entryColWidth[colJ], lineHt, entryHeader[colJ], "1", 0, "CM", true, 0, "")
	}
	pdf.Ln(-1)
	pdf.SetFillColor(255, 255, 255)

	for rowJ := 0; rowJ < len(entryItems); rowJ++ {
		pdf.SetFontStyle("")
		val := entryItems[rowJ]
		dateString := val.dateString
		billingCode := val.billingCode
		staffName := val.staff
		description := val.description
		hours := val.hoursFormatted

		// Calculate the maximum height needed for the multiline cell
		maxWidth := entryColWidth[3]
		maxHeight := getMaxHeight(pdf, description, maxWidth)
		pdf.CellFormat(entryColWidth[0], maxHeight, dateString, "1", 0, "CM", true, 0, "")
		pdf.CellFormat(entryColWidth[1], maxHeight, billingCode, "1", 0, "CM", true, 0, "")
		pdf.CellFormat(entryColWidth[2], maxHeight, staffName, "1", 0, "CM", true, 0, "")
		// Check if the description needs to be split
		if pdf.GetStringWidth(description) > maxWidth {
			pdf.MultiCell(maxWidth, entryFontSize, description, "1", "LM", true)
			pdf.SetXY(marginX+entryColWidth[0]+entryColWidth[1]+entryColWidth[2]+entryColWidth[3], pdf.GetY()-maxHeight)
		} else {
			pdf.CellFormat(maxWidth, maxHeight, description, "1", 0, "LM", true, 0, "")
		}
		// Set position for the next row to account for the lack of ln option in multicell
		pdf.CellFormat(entryColWidth[4], maxHeight, hours, "1", 0, "CM", true, 0, "")
		pdf.Ln(-1)
	}
	
	// Add expense details page if there are expenses
	if len(expenses) > 0 {
		pdf.AddPage()
		pdf.SetFontStyle("B")
		pdf.SetFontSize(14.0)
		pdf.SetXY(marginX, marginY)
		pdf.Cell(safeAreaW, lineHeight, "Expense Details")
		pdf.Ln(lineHeight + gapY)
		
		pdf.SetFontSize(10.0)
		pdf.SetFontStyle("")
		
		for i, expense := range expenses {
			// Wrap each expense in error recovery to prevent crashes
			func() {
				defer func() {
					if r := recover(); r != nil {
						// Log the error but continue processing
						fmt.Printf("Error processing expense %d: %v\n", expense.ID, r)
					}
				}()
				
				// Add some spacing between expenses
				if pdf.GetY() > 230 {
					pdf.AddPage()
					pdf.SetXY(marginX, marginY)
				}
				
				startY := pdf.GetY()
				
				// Expense details section (left side)
				pdf.SetFontStyle("B")
				pdf.SetX(marginX)
				pdf.Cell(80, 6, fmt.Sprintf("Expense: %s", expense.Description))
				pdf.Ln(6)
				
				pdf.SetFontStyle("")
				pdf.SetX(marginX)
				pdf.Cell(40, 5, "Date:")
				pdf.Cell(40, 5, expense.Date.Format("01/02/2006"))
				pdf.Ln(5)
				
				pdf.SetX(marginX)
				pdf.Cell(40, 5, "Amount:")
				pdf.Cell(40, 5, fmt.Sprintf("$%.2f", float64(expense.Amount)/100.0))
				pdf.Ln(5)
				
				pdf.SetX(marginX)
				pdf.Cell(40, 5, "Submitted by:")
				pdf.Cell(40, 5, fmt.Sprintf("%s %s", expense.Submitter.FirstName, expense.Submitter.LastName))
				pdf.Ln(5)
			
			// Add receipt image if available
			if expense.Receipt != nil && expense.Receipt.Url != "" {
				pdf.SetX(marginX)
				pdf.SetFontStyle("I")
				pdf.Cell(80, 5, "Receipt:")
				pdf.Ln(8)
				
				receiptURL := expense.Receipt.Url
				
				// Determine image type from content type or URL
				imageType := "PNG"
				fileExt := ".png"
				isPDF := false
				
				if expense.Receipt.ContentType != nil {
					contentType := *expense.Receipt.ContentType
					if strings.Contains(contentType, "jpeg") || strings.Contains(contentType, "jpg") {
						imageType = "JPG"
						fileExt = ".jpg"
					} else if strings.Contains(contentType, "png") {
						imageType = "PNG"
						fileExt = ".png"
					} else if strings.Contains(contentType, "pdf") {
						isPDF = true
					}
				} else if strings.Contains(strings.ToLower(receiptURL), ".jpg") || strings.Contains(strings.ToLower(receiptURL), ".jpeg") {
					imageType = "JPG"
					fileExt = ".jpg"
				} else if strings.Contains(strings.ToLower(receiptURL), ".pdf") {
					isPDF = true
				}
				
				if isPDF {
					// Skip PDFs for now, just show link
					pdf.SetX(marginX)
					pdf.SetFontStyle("")
					pdf.Cell(80, 5, fmt.Sprintf("Receipt (PDF): %s", receiptURL))
					pdf.Ln(5)
				} else if expense.Receipt.BucketName != nil && expense.Receipt.GCSObjectPath != nil {
					// Download from GCS bucket using the app's storage client
					tempDir := os.TempDir()
					tempFile := filepath.Join(tempDir, fmt.Sprintf("receipt_%d%s", expense.ID, fileExt))
					
					// Download the image from GCS
					ctx := context.Background()
					client := a.InitializeStorageClient(a.Project, *expense.Receipt.BucketName)
					bucket := client.Bucket(*expense.Receipt.BucketName)
					obj := bucket.Object(*expense.Receipt.GCSObjectPath)
					
					reader, err := obj.NewReader(ctx)
					if err == nil {
						defer reader.Close()
						
						// Create the temporary file
						out, err := os.Create(tempFile)
						if err == nil {
							// Copy the image data
							_, err = io.Copy(out, reader)
							out.Close()
							
							if err == nil {
								// Add the image to the PDF
								imageX := marginX
								imageY := pdf.GetY()
								maxImageWidth := 90.0
								
								imageOpts := gofpdf.ImageOptions{
									ImageType: imageType,
									ReadDpi:   true,
								}
								
								// Check if we need a new page
								if imageY > 200 {
									pdf.AddPage()
									imageY = marginY
									pdf.SetXY(marginX, imageY)
								}
								
								// Add the image
								pdf.ImageOptions(tempFile, imageX, imageY, maxImageWidth, 0, false, imageOpts, 0, "")
								
								// Move Y position past the image
								pdf.SetY(imageY + 80) // Approximate image height
							}
							
							// Clean up temp file
							os.Remove(tempFile)
						}
					}
					
					// If any step failed, show a note
					if err != nil {
						pdf.SetX(marginX)
						pdf.SetFontStyle("")
						pdf.Cell(80, 5, "Receipt image unavailable")
						pdf.Ln(5)
					}
				} else {
					// No GCS path available, just note that receipt exists
					pdf.SetX(marginX)
					pdf.SetFontStyle("")
					pdf.Cell(80, 5, "Receipt on file")
					pdf.Ln(5)
				}
			}
				
				pdf.Ln(10) // Space before next expense
				
				// Draw a separator line
				pdf.Line(marginX, pdf.GetY(), marginX+safeAreaW, pdf.GetY())
				pdf.Ln(5)
				
				_ = startY // Suppress unused variable warning if not needed
				_ = i // Suppress unused variable warning
			}() // Close the defer func
		}
	}
	
	var buffer bytes.Buffer
	err := pdf.Output(&buffer)
	if err != nil {
		fmt.Println(err)
	}
	return buffer.Bytes()
}

func breakAddress(input string) []string {
	var address []string
	const limit = 10
	splitted := strings.Split(input, ",")
	prevAddress := ""
	for _, add := range splitted {
		if len(add) < 10 {
			prevAddress = add
			continue
		}
		currentAdd := strings.TrimSpace(add)
		if prevAddress != "" {
			currentAdd = prevAddress + ", " + currentAdd
		}
		address = append(address, currentAdd)
		prevAddress = ""
	}

	return address
}

// getMaxHeight calculates the maximum height needed for a multiline cell
func getMaxHeight(pdf *gofpdf.Fpdf, text string, maxWidth float64) float64 {
	fontSize, _ := pdf.GetFontSize()
	if pdf.GetStringWidth(text) <= maxWidth {
		return fontSize
	}

	lines := pdf.SplitLines([]byte(text), maxWidth)
	return float64(len(lines)) * fontSize
}
