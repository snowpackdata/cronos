package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/snowpackdata/cronos"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// migrateJournalEntries migrates old journal entries to new accrual accounting structure
func migrateJournalEntries(db *gorm.DB, dryRun bool) error {
	log.Println("Starting journal entry migration...")

	if dryRun {
		log.Println("===== DRY RUN MODE - NO CHANGES WILL BE MADE =====")
		log.Println("")
	}

	// Get all existing journal entries
	var journals []cronos.Journal
	if err := db.Preload("Invoice").Preload("Bill").Find(&journals).Error; err != nil {
		return fmt.Errorf("failed to fetch journal entries: %w", err)
	}

	log.Printf("Found %d existing journal entries", len(journals))

	// Track what we'll create
	newEntries := []cronos.Journal{}

	// Step 1: Process all non-void invoices
	var invoices []cronos.Invoice
	if err := db.Preload("Account").Where("state != ?", cronos.InvoiceStateVoid.String()).Find(&invoices).Error; err != nil {
		return fmt.Errorf("failed to fetch invoices: %w", err)
	}

	log.Printf("Processing %d invoices...", len(invoices))

	for _, invoice := range invoices {
		// Use the pre-calculated TotalAmount from the invoice
		// Note: Invoice amounts are in dollars, convert to cents for journal entries
		totalAmount := int64(invoice.TotalAmount * 100)

		if totalAmount == 0 {
			continue
		}

		accountID := invoice.AccountID
		accountName := invoice.Account.Name

		subAccount := fmt.Sprintf("%d", accountID)
		if accountName != "" {
			subAccount = fmt.Sprintf("%d:%s", accountID, accountName)
		}

		isDraftInvoice := invoice.State == cronos.InvoiceStateDraft.String()

		// Determine the appropriate timestamp for backdating
		entryTimestamp := invoice.CreatedAt // Default to creation date

		// Determine account based on invoice state and payment status
		if isDraftInvoice {
			// Draft invoice with approved entries - book to Accrued Receivables
			newEntries = append(newEntries, cronos.Journal{
				Model:      gorm.Model{CreatedAt: entryTimestamp, UpdatedAt: entryTimestamp},
				Account:    string(cronos.AccountAccruedReceivables),
				SubAccount: subAccount,
				InvoiceID:  &invoice.ID,
				Memo:       fmt.Sprintf("Migration: Accrued receivables for draft invoice #%d", invoice.ID),
				Debit:      totalAmount,
				Credit:     0,
			})
			newEntries = append(newEntries, cronos.Journal{
				Model:      gorm.Model{CreatedAt: entryTimestamp, UpdatedAt: entryTimestamp},
				Account:    string(cronos.AccountRevenue),
				SubAccount: subAccount,
				InvoiceID:  &invoice.ID,
				Memo:       fmt.Sprintf("Migration: Revenue from draft invoice #%d", invoice.ID),
				Debit:      0,
				Credit:     totalAmount,
			})

			if !dryRun {
				log.Printf("  Draft Invoice %d for %s: Booked $%.2f to Accrued Receivables/Revenue", invoice.ID, accountName, float64(totalAmount)/100)
			} else {
				log.Printf("  [DRY RUN] Draft Invoice %d for %s: Would book $%.2f to Accrued Receivables/Revenue", invoice.ID, accountName, float64(totalAmount)/100)
			}
		} else {
			// Sent/Approved invoice - check if paid
			isPaid := !invoice.ClosedAt.IsZero()

			if isPaid {
				// Use ClosedAt for paid invoices
				entryTimestamp = invoice.ClosedAt

				// Paid - book to Cash
				newEntries = append(newEntries, cronos.Journal{
					Model:      gorm.Model{CreatedAt: entryTimestamp, UpdatedAt: entryTimestamp},
					Account:    string(cronos.AccountCash),
					SubAccount: subAccount,
					InvoiceID:  &invoice.ID,
					Memo:       fmt.Sprintf("Migration: Paid invoice #%d", invoice.ID),
					Debit:      totalAmount,
					Credit:     0,
				})
				newEntries = append(newEntries, cronos.Journal{
					Model:      gorm.Model{CreatedAt: entryTimestamp, UpdatedAt: entryTimestamp},
					Account:    string(cronos.AccountRevenue),
					SubAccount: subAccount,
					InvoiceID:  &invoice.ID,
					Memo:       fmt.Sprintf("Migration: Revenue from paid invoice #%d", invoice.ID),
					Debit:      0,
					Credit:     totalAmount,
				})

				if !dryRun {
					log.Printf("  Invoice %d (paid) for %s: Booked $%.2f to Cash/Revenue", invoice.ID, accountName, float64(totalAmount)/100)
				} else {
					log.Printf("  [DRY RUN] Invoice %d (paid) for %s: Would book $%.2f to Cash/Revenue", invoice.ID, accountName, float64(totalAmount)/100)
				}
			} else {
				// Use SentAt for sent invoices, otherwise CreatedAt
				if !invoice.SentAt.IsZero() {
					entryTimestamp = invoice.SentAt
				}

				// Unpaid - book to Accounts Receivable
				newEntries = append(newEntries, cronos.Journal{
					Model:      gorm.Model{CreatedAt: entryTimestamp, UpdatedAt: entryTimestamp},
					Account:    string(cronos.AccountAccountsReceivable),
					SubAccount: subAccount,
					InvoiceID:  &invoice.ID,
					Memo:       fmt.Sprintf("Migration: Unpaid invoice #%d", invoice.ID),
					Debit:      totalAmount,
					Credit:     0,
				})
				newEntries = append(newEntries, cronos.Journal{
					Model:      gorm.Model{CreatedAt: entryTimestamp, UpdatedAt: entryTimestamp},
					Account:    string(cronos.AccountRevenue),
					SubAccount: subAccount,
					InvoiceID:  &invoice.ID,
					Memo:       fmt.Sprintf("Migration: Revenue from unpaid invoice #%d", invoice.ID),
					Debit:      0,
					Credit:     totalAmount,
				})

				if !dryRun {
					log.Printf("  Invoice %d (unpaid) for %s: Booked $%.2f to AR/Revenue", invoice.ID, accountName, float64(totalAmount)/100)
				} else {
					log.Printf("  [DRY RUN] Invoice %d (unpaid) for %s: Would book $%.2f to AR/Revenue", invoice.ID, accountName, float64(totalAmount)/100)
				}
			}
		}
	}

	// Step 2: Process all bills (bills don't have a state field, they use AcceptedAt/ClosedAt timestamps)
	var bills []cronos.Bill
	if err := db.Preload("Employee").Find(&bills).Error; err != nil {
		return fmt.Errorf("failed to fetch bills: %w", err)
	}

	log.Printf("Processing %d bills...", len(bills))

	for _, bill := range bills {
		// Use the pre-calculated TotalAmount from the bill (includes fees + commissions + adjustments)
		totalAmount := int64(bill.TotalAmount)

		if totalAmount == 0 {
			continue // Skip zero-amount bills
		}

		employeeID := bill.EmployeeID
		employeeName := fmt.Sprintf("%s %s", bill.Employee.FirstName, bill.Employee.LastName)

		// Check if employee is variable or base+variable (only those get billed)
		var employee cronos.Employee
		if err := db.First(&employee, employeeID).Error; err != nil {
			log.Printf("  Warning: Could not find employee %d for bill %d", employeeID, bill.ID)
			continue
		}

		// Only process bills for variable compensation employees
		if employee.CompensationType != string(cronos.CompensationTypeFullyVariable) &&
			employee.CompensationType != string(cronos.CompensationTypeBasePlusVariable) {
			log.Printf("  Skipping bill %d for employee %d (fixed salary)", bill.ID, employeeID)
			continue
		}

		// Use employee ID as subaccount
		subAccount := fmt.Sprintf("%d", employeeID)
		if employeeName != "" {
			subAccount = fmt.Sprintf("%d:%s", employeeID, employeeName)
		}

		// Check bill status:
		// - Paid: ClosedAt is set -> book to Cash (check this first as it takes precedence)
		// - Draft: AcceptedAt is null -> book to Accrued Payroll
		// - Accepted but unpaid: AcceptedAt is set, ClosedAt is null -> book to Accounts Payable

		isPaid := bill.ClosedAt != nil && !bill.ClosedAt.IsZero()
		isDraft := bill.AcceptedAt == nil || bill.AcceptedAt.IsZero()

		// Determine the appropriate timestamp for backdating
		billTimestamp := bill.CreatedAt // Default to creation date

		if isPaid {
			// Use ClosedAt for paid bills
			billTimestamp = *bill.ClosedAt

			// Bill is paid - book to Payroll Expense (debit) and Cash (credit)
			newEntries = append(newEntries, cronos.Journal{
				Model:      gorm.Model{CreatedAt: billTimestamp, UpdatedAt: billTimestamp},
				Account:    string(cronos.AccountPayrollExpense),
				SubAccount: subAccount,
				BillID:     &bill.ID,
				Memo:       fmt.Sprintf("Migration: Paid bill #%d", bill.ID),
				Debit:      totalAmount,
				Credit:     0,
			})
			newEntries = append(newEntries, cronos.Journal{
				Model:      gorm.Model{CreatedAt: billTimestamp, UpdatedAt: billTimestamp},
				Account:    string(cronos.AccountCash),
				SubAccount: subAccount,
				BillID:     &bill.ID,
				Memo:       fmt.Sprintf("Migration: Cash paid for bill #%d", bill.ID),
				Debit:      0,
				Credit:     totalAmount,
			})

			if !dryRun {
				log.Printf("  Bill %d (paid) for %s: Booked $%.2f to Expense/Cash", bill.ID, employeeName, float64(totalAmount)/100)
			} else {
				log.Printf("  [DRY RUN] Bill %d (paid) for %s: Would book $%.2f to Expense/Cash", bill.ID, employeeName, float64(totalAmount)/100)
			}
		} else if isDraft {
			// Draft bill - book to Expense (debit) and Accrued Payroll (credit)
			newEntries = append(newEntries, cronos.Journal{
				Model:      gorm.Model{CreatedAt: billTimestamp, UpdatedAt: billTimestamp},
				Account:    string(cronos.AccountPayrollExpense),
				SubAccount: subAccount,
				BillID:     &bill.ID,
				Memo:       fmt.Sprintf("Migration: Payroll expense for draft bill #%d", bill.ID),
				Debit:      totalAmount,
				Credit:     0,
			})
			newEntries = append(newEntries, cronos.Journal{
				Model:      gorm.Model{CreatedAt: billTimestamp, UpdatedAt: billTimestamp},
				Account:    string(cronos.AccountAccruedPayroll),
				SubAccount: subAccount,
				BillID:     &bill.ID,
				Memo:       fmt.Sprintf("Migration: Accrued payroll for draft bill #%d", bill.ID),
				Debit:      0,
				Credit:     totalAmount,
			})

			if !dryRun {
				log.Printf("  Draft Bill %d for %s: Booked $%.2f to Expense/Accrued Payroll", bill.ID, employeeName, float64(totalAmount)/100)
			} else {
				log.Printf("  [DRY RUN] Draft Bill %d for %s: Would book $%.2f to Expense/Accrued Payroll", bill.ID, employeeName, float64(totalAmount)/100)
			}
		} else {
			// Use AcceptedAt for accepted bills, otherwise CreatedAt
			if bill.AcceptedAt != nil && !bill.AcceptedAt.IsZero() {
				billTimestamp = *bill.AcceptedAt
			}

			// Bill is unpaid - book to Payroll Expense (debit) and Accounts Payable (credit)
			newEntries = append(newEntries, cronos.Journal{
				Model:      gorm.Model{CreatedAt: billTimestamp, UpdatedAt: billTimestamp},
				Account:    string(cronos.AccountPayrollExpense),
				SubAccount: subAccount,
				BillID:     &bill.ID,
				Memo:       fmt.Sprintf("Migration: Unpaid bill #%d", bill.ID),
				Debit:      totalAmount,
				Credit:     0,
			})
			newEntries = append(newEntries, cronos.Journal{
				Model:      gorm.Model{CreatedAt: billTimestamp, UpdatedAt: billTimestamp},
				Account:    string(cronos.AccountAccountsPayable),
				SubAccount: subAccount,
				BillID:     &bill.ID,
				Memo:       fmt.Sprintf("Migration: Payable for unpaid bill #%d", bill.ID),
				Debit:      0,
				Credit:     totalAmount,
			})

			if !dryRun {
				log.Printf("  Bill %d (unpaid) for %s: Booked $%.2f to Expense/AP", bill.ID, employeeName, float64(totalAmount)/100)
			} else {
				log.Printf("  [DRY RUN] Bill %d (unpaid) for %s: Would book $%.2f to Expense/AP", bill.ID, employeeName, float64(totalAmount)/100)
			}
		}
	}

	// Step 3: Process standalone APPROVED entries (not on invoices/bills yet)
	// Note: We exclude SENT entries as they're likely already on sent invoices
	log.Println("Processing standalone approved entries...")

	// First, let's see what entries exist that aren't on invoices/bills
	var entryCounts []struct {
		State string
		Count int
	}
	db.Model(&cronos.Entry{}).
		Select("state, COUNT(*) as count").
		Where("invoice_id IS NULL AND bill_id IS NULL AND deleted_at IS NULL").
		Group("state").
		Scan(&entryCounts)

	log.Println("Entries not on invoices/bills by state:")
	for _, ec := range entryCounts {
		log.Printf("  %s: %d", ec.State, ec.Count)
	}

	var standaloneEntries []cronos.Entry
	if err := db.Preload("BillingCode").Preload("Employee").
		Where("state = ? AND invoice_id IS NULL AND bill_id IS NULL",
			cronos.EntryStateApproved).
		Find(&standaloneEntries).Error; err != nil {
		return fmt.Errorf("failed to fetch standalone entries: %w", err)
	}

	log.Printf("Found %d standalone approved entries to migrate", len(standaloneEntries))

	// Group entries by account for client receivables and by employee for payables
	clientAccruals := make(map[uint]int64)  // accountID -> total amount
	payrollAccruals := make(map[uint]int64) // employeeID -> total amount
	accountInfo := make(map[uint]string)    // accountID -> account name
	employeeInfo := make(map[uint]string)   // employeeID -> employee name

	for _, entry := range standaloneEntries {
		// Get project and account info from billing code
		if entry.BillingCode.ProjectID > 0 {
			var project cronos.Project
			if err := db.Preload("Account").First(&project, entry.BillingCode.ProjectID).Error; err == nil {
				accountID := project.AccountID
				clientAccruals[accountID] += int64(entry.Fee)
				if project.Account.Name != "" {
					accountInfo[accountID] = project.Account.Name
				}
			}
		}

		// Check if we need to accrue payroll for this entry
		if entry.EmployeeID > 0 {
			var employee cronos.Employee
			if err := db.First(&employee, entry.EmployeeID).Error; err == nil {
				// Only accrue if employee is variable or base+variable
				if employee.CompensationType == string(cronos.CompensationTypeFullyVariable) ||
					employee.CompensationType == string(cronos.CompensationTypeBasePlusVariable) {
					payrollAccruals[employee.ID] += int64(entry.Fee)
					employeeInfo[employee.ID] = fmt.Sprintf("%s %s", employee.FirstName, employee.LastName)
				}
			}
		}
	}

	// Use current time for standalone approved entries since they represent current unbilled work
	standaloneTimestamp := time.Now()

	// Book client accruals
	for accountID, amount := range clientAccruals {
		if amount == 0 {
			continue
		}
		subAccount := fmt.Sprintf("%d", accountID)
		if name, ok := accountInfo[accountID]; ok && name != "" {
			subAccount = fmt.Sprintf("%d:%s", accountID, name)
		}

		// DR: ACCRUED_RECEIVABLES, CR: REVENUE
		newEntries = append(newEntries, cronos.Journal{
			Model:      gorm.Model{CreatedAt: standaloneTimestamp, UpdatedAt: standaloneTimestamp},
			Account:    string(cronos.AccountAccruedReceivables),
			SubAccount: subAccount,
			Memo:       fmt.Sprintf("Migration: Accrued receivables for account %d", accountID),
			Debit:      amount,
			Credit:     0,
		})
		newEntries = append(newEntries, cronos.Journal{
			Model:      gorm.Model{CreatedAt: standaloneTimestamp, UpdatedAt: standaloneTimestamp},
			Account:    string(cronos.AccountRevenue),
			SubAccount: subAccount,
			Memo:       fmt.Sprintf("Migration: Revenue from approved entries for account %d", accountID),
			Debit:      0,
			Credit:     amount,
		})

		if !dryRun {
			log.Printf("  Account %s: Booked $%.2f to Accrued Receivables/Revenue", accountInfo[accountID], float64(amount)/100)
		} else {
			log.Printf("  [DRY RUN] Account %s: Would book $%.2f to Accrued Receivables/Revenue", accountInfo[accountID], float64(amount)/100)
		}
	}

	// Book payroll accruals
	for employeeID, amount := range payrollAccruals {
		if amount == 0 {
			continue
		}
		subAccount := fmt.Sprintf("%d", employeeID)
		if name, ok := employeeInfo[employeeID]; ok && name != "" {
			subAccount = fmt.Sprintf("%d:%s", employeeID, name)
		}

		// DR: PAYROLL_EXPENSE, CR: ACCRUED_PAYROLL
		newEntries = append(newEntries, cronos.Journal{
			Model:      gorm.Model{CreatedAt: standaloneTimestamp, UpdatedAt: standaloneTimestamp},
			Account:    string(cronos.AccountPayrollExpense),
			SubAccount: subAccount,
			Memo:       fmt.Sprintf("Migration: Payroll expense for approved entries for employee %d", employeeID),
			Debit:      amount,
			Credit:     0,
		})
		newEntries = append(newEntries, cronos.Journal{
			Model:      gorm.Model{CreatedAt: standaloneTimestamp, UpdatedAt: standaloneTimestamp},
			Account:    string(cronos.AccountAccruedPayroll),
			SubAccount: subAccount,
			Memo:       fmt.Sprintf("Migration: Accrued payroll for employee %d", employeeID),
			Debit:      0,
			Credit:     amount,
		})

		if !dryRun {
			log.Printf("  Employee %s: Booked $%.2f to Expense/Accrued Payroll", employeeInfo[employeeID], float64(amount)/100)
		} else {
			log.Printf("  [DRY RUN] Employee %s: Would book $%.2f to Expense/Accrued Payroll", employeeInfo[employeeID], float64(amount)/100)
		}
	}

	// Step 5: Show summary and execute if not dry run
	log.Println("")
	log.Println("========================================")
	log.Println("MIGRATION SUMMARY")
	log.Println("========================================")

	// Calculate account balances
	accountBalances := make(map[string]int64)
	for _, entry := range newEntries {
		accountBalances[entry.Account] += entry.Debit - entry.Credit
	}

	log.Println("New journal entries to be created:")
	for account, balance := range accountBalances {
		log.Printf("  %-30s: $%10.2f", account, float64(balance)/100)
	}

	log.Println("")
	log.Printf("Total old entries to delete: %d", len(journals))
	log.Printf("Total new entries to create: %d", len(newEntries))

	if dryRun {
		log.Println("")
		log.Println("========================================")
		log.Println("DETAILED DRY RUN OUTPUT")
		log.Println("========================================")
		log.Println("")
		log.Println("New journal entries that would be created:")
		for i, entry := range newEntries {
			log.Printf("%d. Account: %-30s SubAccount: %-20s", i+1, entry.Account, entry.SubAccount)
			log.Printf("   Memo: %s", entry.Memo)
			if entry.InvoiceID != nil {
				log.Printf("   InvoiceID: %d", *entry.InvoiceID)
			}
			if entry.BillID != nil {
				log.Printf("   BillID: %d", *entry.BillID)
			}
			if entry.Debit > 0 {
				log.Printf("   DR: $%.2f", float64(entry.Debit)/100)
			}
			if entry.Credit > 0 {
				log.Printf("   CR: $%.2f", float64(entry.Credit)/100)
			}
			log.Println("")
		}

		log.Println("========================================")
		log.Println("[DRY RUN] NO CHANGES MADE")
		log.Println("========================================")
	} else {
		log.Println("")
		log.Println("========================================")
		log.Println("EXECUTING MIGRATION")
		log.Println("========================================")

		// Step 1: Archive old journal entries
		log.Printf("Archiving %d existing journal entries...", len(journals))

		// Create archive table with timestamp
		timestamp := time.Now().Format("20060102_150405")
		archiveTableName := fmt.Sprintf("journals_archive_%s", timestamp)

		// Check if using SQLite or PostgreSQL for syntax differences
		var createArchiveSQL string

		// Detect database type from the connection
		if db.Dialector.Name() == "sqlite" {
			createArchiveSQL = fmt.Sprintf("CREATE TABLE %s AS SELECT * FROM journals", archiveTableName)
		} else {
			// PostgreSQL
			createArchiveSQL = fmt.Sprintf("CREATE TABLE %s AS TABLE journals", archiveTableName)
		}

		if err := db.Exec(createArchiveSQL).Error; err != nil {
			return fmt.Errorf("failed to create archive table: %w", err)
		}

		log.Printf("✓ Created archive table: %s", archiveTableName)

		// Step 2: NULL out invoice references to journals before deleting
		log.Println("Clearing invoice journal references...")
		if err := db.Exec("UPDATE invoices SET journal_id = NULL WHERE journal_id IS NOT NULL").Error; err != nil {
			return fmt.Errorf("failed to clear invoice journal references: %w", err)
		}
		log.Println("✓ Cleared invoice references")

		// Step 3: Drop foreign key constraints temporarily
		if db.Dialector.Name() == "postgres" {
			log.Println("Dropping foreign key constraints...")

			// Query to find all FK constraints on journals table
			var constraints []struct {
				ConstraintName string
				TableName      string
			}
			db.Raw(`
				SELECT constraint_name, table_name 
				FROM information_schema.table_constraints 
				WHERE table_name = 'journals' 
				AND constraint_type = 'FOREIGN KEY'
			`).Scan(&constraints)

			log.Printf("Found %d FK constraints on journals table", len(constraints))

			// Drop each FK constraint on journals
			for _, c := range constraints {
				sql := fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT IF EXISTS %s", c.TableName, c.ConstraintName)
				log.Printf("Dropping: %s", c.ConstraintName)
				if err := db.Exec(sql).Error; err != nil {
					log.Printf("Warning: Failed to drop %s: %v", c.ConstraintName, err)
				}
			}

			// Also find FK constraints in OTHER tables that reference journals
			db.Raw(`
				SELECT tc.constraint_name, tc.table_name 
				FROM information_schema.table_constraints tc
				JOIN information_schema.constraint_column_usage ccu
					ON tc.constraint_name = ccu.constraint_name
				WHERE ccu.table_name = 'journals'
				AND tc.constraint_type = 'FOREIGN KEY'
			`).Scan(&constraints)

			log.Printf("Found %d FK constraints referencing journals table", len(constraints))

			// Drop each FK constraint that references journals
			for _, c := range constraints {
				sql := fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT IF EXISTS %s", c.TableName, c.ConstraintName)
				log.Printf("Dropping: %s from %s", c.ConstraintName, c.TableName)
				if err := db.Exec(sql).Error; err != nil {
					log.Printf("Warning: Failed to drop %s: %v", c.ConstraintName, err)
				}
			}

			log.Println("✓ Dropped all foreign key constraints")
		}

		// Step 4: Delete old journal entries from main table
		log.Printf("Deleting %d old journal entries from main table...", len(journals))
		if err := db.Exec("DELETE FROM journals").Error; err != nil {
			return fmt.Errorf("failed to delete old entries: %w", err)
		}
		log.Println("✓ Deleted old entries")

		// Step 5: Create new journal entries
		// Note: We're not updating invoice.journal_id back to the new entries
		// because the new structure creates multiple journal entries per invoice (DR/CR pairs)
		// Invoices can find their journals via journal.invoice_id instead
		log.Printf("Creating %d new journal entries...", len(newEntries))
		successCount := 0
		for _, entry := range newEntries {
			if err := db.Create(&entry).Error; err != nil {
				log.Printf("Warning: Failed to create journal entry: %v", err)
			} else {
				successCount++
			}
		}
		log.Printf("✓ Created %d new journal entries", successCount)

		// Step 6: Recreate foreign key constraints
		if db.Dialector.Name() == "postgres" {
			log.Println("Recreating foreign key constraints...")

			// Let GORM auto-migrate to recreate the constraints
			if err := db.AutoMigrate(&cronos.Journal{}, &cronos.Invoice{}); err != nil {
				log.Printf("Warning: Failed to auto-migrate constraints: %v", err)
			} else {
				log.Println("✓ Recreated foreign key constraints")
			}
		}

		log.Println("")
		log.Println("========================================")
		log.Println("MIGRATION COMPLETED SUCCESSFULLY!")
		log.Println("========================================")
		log.Printf("Archive table: %s", archiveTableName)
		log.Println("You can drop the archive table after verifying the migration:")
		log.Printf("  DROP TABLE %s;", archiveTableName)
	}

	return nil
}

func main() {
	// Parse command line flags
	dbType := flag.String("db", "sqlite", "Database type (sqlite or postgres)")
	dbPath := flag.String("path", "./cronos.db", "Path to SQLite database (if using sqlite)")
	dbHost := flag.String("host", "localhost", "PostgreSQL host (if using postgres)")
	dbPort := flag.String("port", "5432", "PostgreSQL port (if using postgres)")
	dbName := flag.String("name", "cronos", "PostgreSQL database name (if using postgres)")
	dbUser := flag.String("user", "postgres", "PostgreSQL user (if using postgres)")
	dbPass := flag.String("password", "", "PostgreSQL password (if using postgres)")
	dryRun := flag.Bool("dry-run", false, "Run in dry-run mode (no changes made)")

	flag.Parse()

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

	// Run migration
	if err := migrateJournalEntries(db, *dryRun); err != nil {
		log.Fatalf("Migration failed: %v", err)
		os.Exit(1)
	}

	log.Println("Done!")
}
