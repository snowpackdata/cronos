package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/snowpackdata/cronos"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	// Parse command-line flags
	dbType := flag.String("db", "sqlite", "Database type (sqlite or postgres)")
	dbPath := flag.String("path", "./cronos.db", "Path to SQLite database (if using sqlite)")
	dbHost := flag.String("host", "localhost", "PostgreSQL host (if using postgres)")
	dbPort := flag.String("port", "5432", "PostgreSQL port (if using postgres)")
	dbName := flag.String("name", "cronos", "PostgreSQL database name (if using postgres)")
	dbUser := flag.String("user", "postgres", "PostgreSQL user (if using postgres)")
	dbPass := flag.String("password", "", "PostgreSQL password (if using postgres)")
	dryRun := flag.Bool("dry-run", false, "Preview changes without committing to database")
	flag.Parse()

	log.Println("==============================================")
	log.Println("Line Items Backfill Script")
	log.Println("==============================================")
	log.Println("This script generates line items for all existing invoices and bills")
	if *dryRun {
		log.Println("DRY RUN MODE: No database changes will be made")
	}
	log.Println()

	// Connect to database
	var db *gorm.DB
	var err error

	if *dbType == "sqlite" {
		log.Printf("Connecting to SQLite database at %s...", *dbPath)
		db, err = gorm.Open(sqlite.Open(*dbPath), &gorm.Config{})
	} else {
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
			*dbHost, *dbUser, *dbPass, *dbName, *dbPort)
		log.Printf("Connecting to PostgreSQL database %s at %s:%s...", *dbName, *dbHost, *dbPort)
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	}

	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Create app instance
	app := &cronos.App{DB: db}

	// Run migrations to ensure line item tables exist
	log.Println("Running migrations to ensure line item tables exist...")
	if err := db.AutoMigrate(&cronos.InvoiceLineItem{}, &cronos.BillLineItem{}); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	log.Println("Migrations completed successfully")

	// Counters
	var invoicesProcessed, invoicesSkipped int
	var billsProcessed, billsSkipped int

	// ========================================
	// Step 1: Process Invoices
	// ========================================
	log.Println("\n=== STEP 1: Processing Invoices ===")

	var invoices []cronos.Invoice
	if err := app.DB.Preload("Entries").Preload("Adjustments").Preload("Account").
		Where("type = ? AND state != ?", cronos.InvoiceTypeAR.String(), cronos.InvoiceStateVoid.String()).
		Find(&invoices).Error; err != nil {
		log.Fatalf("Failed to load invoices: %v", err)
	}

	log.Printf("Found %d non-void invoices to process\n", len(invoices))

	for _, invoice := range invoices {
		// Check if line items already exist
		var existingCount int64
		app.DB.Model(&cronos.InvoiceLineItem{}).Where("invoice_id = ?", invoice.ID).Count(&existingCount)

		if existingCount > 0 {
			log.Printf("  Skipping invoice #%d - %d line items already exist", invoice.ID, existingCount)
			invoicesSkipped++
			continue
		}

		log.Printf("  Processing invoice #%d (%s) - %d entries, %d adjustments",
			invoice.ID, invoice.Name, len(invoice.Entries), len(invoice.Adjustments))

		if !*dryRun {
			if err := app.GenerateInvoiceLineItems(&invoice); err != nil {
				log.Printf("    ERROR: Failed to generate line items: %v", err)
				invoicesSkipped++
				continue
			}

			// Count created line items
			var lineItemCount int64
			app.DB.Model(&cronos.InvoiceLineItem{}).Where("invoice_id = ?", invoice.ID).Count(&lineItemCount)
			log.Printf("    Created %d line items", lineItemCount)
		} else {
			log.Printf("    [DRY RUN] Would generate line items for invoice #%d", invoice.ID)
		}

		invoicesProcessed++
	}

	log.Printf("\nInvoices Summary: %d processed, %d skipped\n", invoicesProcessed, invoicesSkipped)

	// ========================================
	// Step 2: Process Bills
	// ========================================
	log.Println("\n=== STEP 2: Processing Bills ===")

	var bills []cronos.Bill
	if err := app.DB.Preload("Entries").Preload("Commissions").Preload("Adjustments").Preload("Employee").
		Find(&bills).Error; err != nil {
		log.Fatalf("Failed to load bills: %v", err)
	}

	log.Printf("Found %d bills to process\n", len(bills))

	for _, bill := range bills {
		// Check if line items already exist
		var existingCount int64
		app.DB.Model(&cronos.BillLineItem{}).Where("bill_id = ?", bill.ID).Count(&existingCount)

		if existingCount > 0 {
			log.Printf("  Skipping bill #%d - %d line items already exist", bill.ID, existingCount)
			billsSkipped++
			continue
		}

		log.Printf("  Processing bill #%d (Employee: %s %s) - %d entries, %d commissions, %d adjustments",
			bill.ID, bill.Employee.FirstName, bill.Employee.LastName,
			len(bill.Entries), len(bill.Commissions), len(bill.Adjustments))

		if !*dryRun {
			if err := app.GenerateBillLineItems(&bill); err != nil {
				log.Printf("    ERROR: Failed to generate line items: %v", err)
				billsSkipped++
				continue
			}

			// Count created line items
			var lineItemCount int64
			app.DB.Model(&cronos.BillLineItem{}).Where("bill_id = ?", bill.ID).Count(&lineItemCount)
			log.Printf("    Created %d line items", lineItemCount)
		} else {
			log.Printf("    [DRY RUN] Would generate line items for bill #%d", bill.ID)
		}

		billsProcessed++
	}

	log.Printf("\nBills Summary: %d processed, %d skipped\n", billsProcessed, billsSkipped)

	// ========================================
	// Summary
	// ========================================
	log.Println("\n==============================================")
	log.Println("Line Items Backfill Summary")
	log.Println("==============================================")
	log.Printf("Invoices: %d processed, %d skipped", invoicesProcessed, invoicesSkipped)
	log.Printf("Bills:    %d processed, %d skipped", billsProcessed, billsSkipped)

	if *dryRun {
		log.Println("\n*** DRY RUN COMPLETED - No changes were made ***")
	} else {
		log.Println("\n*** BACKFILL COMPLETED SUCCESSFULLY ***")

		// Verification
		var totalInvoiceLineItems, totalBillLineItems int64
		app.DB.Model(&cronos.InvoiceLineItem{}).Count(&totalInvoiceLineItems)
		app.DB.Model(&cronos.BillLineItem{}).Count(&totalBillLineItems)

		log.Printf("\nVerification:")
		log.Printf("  Total invoice line items in database: %d", totalInvoiceLineItems)
		log.Printf("  Total bill line items in database: %d", totalBillLineItems)
	}

	log.Println("==============================================")

	if *dryRun {
		fmt.Println("\nTo apply changes, run again without -dry-run flag")
	}
}
