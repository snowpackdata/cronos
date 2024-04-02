package cronos

import (
	"bytes"
	"fmt"
	"github.com/jung-kurt/gofpdf"
	"strconv"
	"time"
)

// Set Constants for the invoice PDF

func (a *App) GenerateBillPDF(bill *Bill) []byte {
	// From the Invoice we need to generate the following:
	// 1. An overall invoice summary grouped at the billing code level
	// 2. A detailed invoice summary grouped at the individual entry level
	// 3. Associated Billable/Client Information
	// 4. Summary of Totals and Subtotals

	// 1. Overall invoice summary grouped at the billing code level
	//lineItems := a.GetInvoiceLineItems(invoice)
	//entryItems := a.GetInvoiceEntries(invoice)
	//adjustments := a.GetInvoiceAdjustments(invoice)

	var employee Employee
	a.DB.Preload("User").Where("id = ?", bill.EmployeeID).First(&employee)

	InvoiceNumber := strconv.Itoa(time.Now().Year()) + "00" + strconv.Itoa(int(bill.ID))

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
	pdf.MultiCell(120, 10, "INTERNAL INVOICE\n "+bill.Name, "0", "R", false)

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
	pdf.Cell(invoiceDetailW, lineHeight, "Invoice No:")
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
	const colNumber = 4
	header := [colNumber]string{"Name", "Item", "Hours", "Total ($)"}
	colWidth := [colNumber]float64{50.0, 100.0, 25.0, 25.0}

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

	for rowJ := 0; rowJ < 1; rowJ++ {
		name := employee.FirstName + " " + employee.LastName
		item := bill.Name
		hours := fmt.Sprintf("%.2f", bill.TotalHours)
		total := fmt.Sprintf("%.2f", float64(bill.TotalAmount)/100)

		pdf.CellFormat(colWidth[0], lineHt, name, "1", 0, "LM", true, 0, "")
		pdf.CellFormat(colWidth[1], lineHt, item, "1", 0, "LM", true, 0, "")
		pdf.CellFormat(colWidth[2], lineHt, hours, "1", 0, "RM", true, 0, "")
		pdf.CellFormat(colWidth[3], lineHt, "$ "+total, "1", 0, "RM", true, 0, "")
		pdf.Ln(-1)
	}

	// Generate the total Rows

	// Calculate the subtotal
	pdf.SetFontStyle("B")
	leftIndent := 0.0
	for i := 0; i < 2; i++ {
		leftIndent += colWidth[i]
	}

	totalFees := fmt.Sprintf("%.2f", float64(bill.TotalFees)/100)
	pdf.SetX(marginX + leftIndent)
	pdf.CellFormat(colWidth[2], lineHt, "Fees", "1", 0, "LM", true, 0, "")
	pdf.CellFormat(colWidth[3], lineHt, "$ "+totalFees, "1", 0, "RM", true, 0, "")
	pdf.Ln(lineHt)

	grandTotal := fmt.Sprintf("%.2f", float64(bill.TotalAmount)/100)
	pdf.SetX(marginX + leftIndent)
	pdf.CellFormat(colWidth[2], lineHt, "Total Due", "1", 0, "LM", true, 0, "")
	pdf.CellFormat(colWidth[3], lineHt, "$ "+grandTotal, "1", 0, "RM", true, 0, "")
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
