package cronos

import (
	"fmt"
	"github.com/jung-kurt/gofpdf"
	"strconv"
	"strings"
	"time"
)

// Set Constants for the invoice PDF
const defaultFont = "Helvetica"
const defaultFromName = "Snowpack Data LLC"
const defaultFromAddress = "262 Dolores Street, San Francisco CA 94103"
const defaultContact = "billing@snowpack-data.io"
const headerWidth = 40.0
const headerHeight = 10.0
const marginX = 10.0
const marginY = 20.0
const gapY = 2.0

func (a *App) GenerateInvoicePDF(invoice *Invoice) string {
	// From the Invoice we need to generate the following:
	// 1. An overall invoice summary grouped at the billing code level
	// 2. A detailed invoice summary grouped at the individual entry level
	// 3. Associated Billable/Client Information
	// 4. Summary of Totals and Subtotals

	// 1. Overall invoice summary grouped at the billing code level
	lineItems := a.GetInvoiceLineItems(invoice)
	entryItems := a.GetInvoiceEntries(invoice)

	var project Project
	var account Account
	a.DB.Where("id = ?", invoice.ProjectID).First(&project)
	a.DB.Where("id = ?", project.AccountID).First(&account)

	InvoiceNumber := strconv.Itoa(time.Now().Year()) + "00" + strconv.Itoa(int(invoice.ID))

	// Initialize the PDF document with set margins and add a page that we can work with
	pdf := gofpdf.New("P", "mm", "Letter", "")
	pdf.SetMargins(marginX, marginY, marginX)
	pdf.AddPage()
	pageW, _ := pdf.GetPageSize()
	safeAreaW := pageW - 2*marginX

	// Build the header
	pdf.ImageOptions("./assets/img/graph-logo.png", 10, 10, 30, 15, false, gofpdf.ImageOptions{ImageType: "PNG", ReadDpi: true}, 0, "")
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
	colWidth := [colNumber]float64{50.0, 60.0, 25.0, 25.0, 40.0}

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
		pdf.CellFormat(colWidth[1], lineHt, description, "1", 0, "CM", true, 0, "")
		pdf.CellFormat(colWidth[2], lineHt, hours, "1", 0, "CM", true, 0, "")
		pdf.CellFormat(colWidth[3], lineHt, "$ "+rate, "1", 0, "CM", true, 0, "")
		pdf.CellFormat(colWidth[4], lineHt, "$ "+total, "1", 0, "RM", true, 0, "")
		pdf.Ln(-1)
	}

	// Calculate the subtotal
	pdf.SetFontStyle("B")
	leftIndent := 0.0
	for i := 0; i < 3; i++ {
		leftIndent += colWidth[i]
	}

	grandTotal := fmt.Sprintf("%.2f", invoice.TotalFees)
	pdf.SetX(marginX + leftIndent)
	pdf.CellFormat(colWidth[3], lineHt, "Invoice Total", "1", 0, "LM", true, 0, "")
	pdf.CellFormat(colWidth[4], lineHt, "$ "+grandTotal, "1", 0, "RM", true, 0, "")
	pdf.Ln(-1)

	pdf.SetFontStyle("")
	pdf.Ln(lineBreak)
	pdf.Cell(safeAreaW, lineHeight, "See second page for detailed timesheet entry breakdown.")

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
	filenameEscaped := invoice.GetInvoiceFilename()
	localPath := "./tmp/" + filenameEscaped + ".pdf"
	err := pdf.OutputFileAndClose(localPath)
	if err != nil {
		panic(err)
	}
	return localPath
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
