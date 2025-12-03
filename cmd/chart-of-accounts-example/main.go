package main

import (
	"fmt"
	"log"
	"time"

	"github.com/snowpackdata/cronos"
)

// Example usage of the Chart of Accounts functionality
// This demonstrates the three main features:
// 1. Creating and managing accounts/subaccounts
// 2. Creating expenses with custom GL accounts
// 3. Importing and categorizing CSV transactions

func main() {
	// Initialize app (you'll need to provide actual DB connection)
	// app := cronos.NewApp(db, ...)
	
	fmt.Println("Chart of Accounts - Example Usage")
	fmt.Println("==================================")
	
	// Example 1: Seed system accounts (run once during setup)
	exampleSeedAccounts()
	
	// Example 2: Create custom accounts and subaccounts
	exampleCreateCustomAccounts()
	
	// Example 3: Create an expense with specific GL categorization
	exampleCreateCategorizedExpense()
	
	// Example 4: Import and categorize CSV transactions
	exampleCSVImport()
}

func exampleSeedAccounts() {
	fmt.Println("\n1. Seeding System Accounts")
	fmt.Println("---------------------------")
	
	// This creates all predefined system accounts
	// Run once during initial setup or migration
	
	/* 
	err := app.SeedSystemAccounts()
	if err != nil {
		log.Fatalf("Failed to seed: %v", err)
	}
	fmt.Println("Created all system accounts (CASH, REVENUE, PAYROLL_EXPENSE, etc.)")
	*/
	
	fmt.Println("Code example:")
	fmt.Println(`
	err := app.SeedSystemAccounts()
	if err != nil {
		log.Fatalf("Failed to seed: %v", err)
	}
	`)
}

func exampleCreateCustomAccounts() {
	fmt.Println("\n2. Creating Custom Accounts & Subaccounts")
	fmt.Println("------------------------------------------")
	
	// Create a custom expense category
	fmt.Println("Creating custom account: MARKETING_EXPENSES")
	
	/*
	account, err := app.CreateChartOfAccount(
		"MARKETING_EXPENSES",              // code
		"Marketing & Advertising",         // name
		"EXPENSE",                         // type
		"Marketing and advertising costs", // description
		nil,                              // no parent
	)
	if err != nil {
		log.Fatalf("Failed: %v", err)
	}
	fmt.Printf("Created: %s\n", account.AccountName)
	*/
	
	// Create subaccounts for different marketing channels
	fmt.Println("\nCreating subaccounts for marketing channels:")
	
	/*
	channels := []struct {
		Code string
		Name string
	}{
		{"GOOGLE_ADS", "Google Advertising"},
		{"FACEBOOK_ADS", "Facebook Advertising"},
		{"LINKEDIN_ADS", "LinkedIn Advertising"},
		{"CONTENT_MARKETING", "Content Marketing"},
	}
	
	for _, channel := range channels {
		sub, err := app.CreateSubaccount(
			channel.Code,
			channel.Name,
			"MARKETING_EXPENSES",
			"VENDOR",
		)
		if err != nil {
			log.Printf("Error creating %s: %v", channel.Code, err)
			continue
		}
		fmt.Printf("  - Created: %s (%s)\n", sub.Name, sub.Code)
	}
	*/
	
	fmt.Println("\nCode example:")
	fmt.Println(`
	// Create account
	account, err := app.CreateChartOfAccount(
		"MARKETING_EXPENSES",
		"Marketing & Advertising",
		"EXPENSE",
		"Marketing and advertising costs",
		nil,
	)
	
	// Create subaccounts
	app.CreateSubaccount("GOOGLE_ADS", "Google Advertising", "MARKETING_EXPENSES", "VENDOR")
	app.CreateSubaccount("FACEBOOK_ADS", "Facebook Advertising", "MARKETING_EXPENSES", "VENDOR")
	`)
}

func exampleCreateCategorizedExpense() {
	fmt.Println("\n3. Creating Expense with GL Categorization")
	fmt.Println("------------------------------------------")
	
	fmt.Println("Creating an expense that will book to specific GL accounts:")
	
	/*
	expense := cronos.Expense{
		ProjectID:          projectID,
		SubmitterID:        employeeID,
		Amount:             250000, // $2,500 in cents
		Date:               time.Now(),
		Description:        "Google Ads Q1 Campaign",
		ExpenseAccountCode: "MARKETING_EXPENSES",  // Custom account we created
		SubaccountCode:     "GOOGLE_ADS",          // Specific vendor
		State:              cronos.ExpenseStateDraft.String(),
	}
	
	err := app.DB.Create(&expense).Error
	if err != nil {
		log.Fatalf("Failed: %v", err)
	}
	
	fmt.Printf("Created expense #%d\n", expense.ID)
	fmt.Printf("  Amount: $%.2f\n", float64(expense.Amount)/100)
	fmt.Printf("  GL Account: %s\n", expense.ExpenseAccountCode)
	fmt.Printf("  Subaccount: %s\n", expense.SubaccountCode)
	*/
	
	fmt.Println("\nWhen this expense is approved and invoiced, it will book as:")
	fmt.Println("  DR: MARKETING_EXPENSES (subaccount: GOOGLE_ADS) $2,500")
	fmt.Println("  CR: CASH (subaccount: ChaseBusiness) $2,500")
	
	fmt.Println("\nCode example:")
	fmt.Println(`
	expense := cronos.Expense{
		ProjectID:          projectID,
		SubmitterID:        employeeID,
		Amount:             250000,  // $2,500
		Date:               time.Now(),
		Description:        "Google Ads Q1 Campaign",
		ExpenseAccountCode: "MARKETING_EXPENSES",
		SubaccountCode:     "GOOGLE_ADS",
		State:              cronos.ExpenseStateDraft.String(),
	}
	
	app.DB.Create(&expense)
	`)
}

func exampleCSVImport() {
	fmt.Println("\n4. CSV Import & Categorization")
	fmt.Println("-------------------------------")
	
	// Example CSV content
	csvExample := `Date,Description,Amount
01/15/2024,"AWS Invoice","-1234.56"
01/16/2024,"Client Payment","5000.00"
01/17/2024,"AMERICAN AIRLINES","-456.78"
01/18/2024,"GOOGLE ADS","-2500.00"
01/19/2024,"Office Depot","-123.45"`
	
	fmt.Println("Sample CSV file:")
	fmt.Println(csvExample)
	
	fmt.Println("\nImporting CSV:")
	
	/*
	csvContent := []byte(csvExample)
	
	// Import creates 2 UNCLASSIFIED entries per transaction
	imported, skipped, err := app.ImportCSVToOfflineJournals(
		csvContent,
		0,               // date column
		1,               // description column
		2,               // amount column
		true,            // has header
		"01/02/2006",    // date format
	)
	
	if err != nil {
		log.Fatalf("Import failed: %v", err)
	}
	
	fmt.Printf("Imported %d transactions (%d journal entries), %d skipped (duplicates)\n", 
		imported, imported*2, skipped)
	*/
	
	fmt.Println("\nReview and categorize by specifying FROM and TO accounts:")
	
	/*
	// Get transactions grouped by date+description
	transactions, err := app.GetOfflineJournalTransactions(
		time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC),
		"pending_review",
	)
	
	// Categorize each transaction by specifying both FROM and TO accounts
	for _, txEntries := range transactions {
		if len(txEntries) < 2 {
			continue
		}
		
		desc := txEntries[0].Description
		date := txEntries[0].Date
		
		if strings.Contains(desc, "AWS") {
			// Expense: FROM Operating Expenses (debit) TO Cash (credit)
			app.CategorizeCSVTransaction(
				date, desc,
				"OPERATING_EXPENSES_SAAS", "AWS",  // FROM (debit)
				"CASH", "ChaseBusiness",            // TO (credit)
			)
		} else if strings.Contains(desc, "Client Payment") {
			// Revenue: FROM Cash (debit) TO Revenue (credit)
			app.CategorizeCSVTransaction(
				date, desc,
				"CASH", "ChaseBusiness",            // FROM (debit)
				"REVENUE", "Client XYZ",            // TO (credit)
			)
		} else if strings.Contains(desc, "AMERICAN AIRLINES") {
			// Travel: FROM Travel Expenses (debit) TO Cash (credit)
			app.CategorizeCSVTransaction(
				date, desc,
				"OPERATING_EXPENSES_TRAVEL", "Airfare",
				"CASH", "ChaseBusiness",
			)
		} else if strings.Contains(desc, "GOOGLE ADS") {
			// Marketing: FROM Marketing (debit) TO Cash (credit)
			app.CategorizeCSVTransaction(
				date, desc,
				"MARKETING_EXPENSES", "GOOGLE_ADS",
				"CASH", "ChaseBusiness",
			)
		}
	}
	*/
	
	fmt.Println("\nApprove and book to GL:")
	
	/*
	// Approve complete transactions (both sides categorized)
	for _, txEntries := range transactions {
		if len(txEntries) < 2 {
			continue
		}
		
		// Check if fully categorized
		allCategorized := true
		for _, entry := range txEntries {
			if entry.Account == "UNCLASSIFIED" {
				allCategorized = false
				break
			}
		}
		
		if allCategorized {
			desc := txEntries[0].Description
			date := txEntries[0].Date
			
			booked, err := app.ApproveTransactionPair(date, desc, staffID)
			if err != nil {
				log.Printf("Error: %v", err)
			} else {
				fmt.Printf("Booked: %s (%d entries)\n", desc, booked)
			}
		}
	}
	*/
	
	fmt.Println("\nCode example:")
	fmt.Println(`
	// Import (creates 2 unclassified entries per transaction)
	imported, skipped, err := app.ImportCSVToOfflineJournals(
		csvContent,
		0, 1, 2,         // column indices
		true,            // has header
		"01/02/2006",    // date format
	)
	
	// Get transactions grouped by date+description
	transactions, _ := app.GetOfflineJournalTransactions(startDate, endDate, "pending_review")
	
	// Categorize each transaction with FROM and TO accounts
	for _, txEntries := range transactions {
		desc := txEntries[0].Description
		date := txEntries[0].Date
		
		if strings.Contains(desc, "AWS") {
			// Expense: FROM expense account TO cash
			app.CategorizeCSVTransaction(
				date, desc,
				"OPERATING_EXPENSES_SAAS", "AWS",  // FROM (debit)
				"CASH", "ChaseBusiness",            // TO (credit)
			)
		}
	}
	
	// Approve complete transactions
	for _, txEntries := range transactions {
		desc := txEntries[0].Description
		date := txEntries[0].Date
		booked, _ := app.ApproveTransactionPair(date, desc, staffID)
	}
	`)
}

func exampleQueryAndReporting() {
	fmt.Println("\n5. Querying Accounts & Subaccounts")
	fmt.Println("----------------------------------")
	
	fmt.Println("Get all expense accounts:")
	
	/*
	expenses, err := app.GetChartOfAccounts("EXPENSE", true)
	if err != nil {
		log.Fatalf("Failed: %v", err)
	}
	
	fmt.Printf("Found %d expense accounts:\n", len(expenses))
	for _, account := range expenses {
		fmt.Printf("  - %s: %s\n", account.AccountCode, account.AccountName)
	}
	*/
	
	fmt.Println("\nGet subaccounts for a specific account:")
	
	/*
	subs, err := app.GetSubaccounts("OPERATING_EXPENSES_SAAS", "", true)
	if err != nil {
		log.Fatalf("Failed: %v", err)
	}
	
	fmt.Printf("SaaS subaccounts:\n")
	for _, sub := range subs {
		fmt.Printf("  - %s: %s\n", sub.Code, sub.Name)
	}
	*/
	
	fmt.Println("\nCode example:")
	fmt.Println(`
	// Get all expense accounts
	expenses, err := app.GetChartOfAccounts("EXPENSE", true)
	for _, account := range expenses {
		fmt.Printf("%s: %s\n", account.AccountCode, account.AccountName)
	}
	
	// Get subaccounts
	subs, err := app.GetSubaccounts("OPERATING_EXPENSES_SAAS", "", true)
	for _, sub := range subs {
		fmt.Printf("%s: %s\n", sub.Code, sub.Name)
	}
	`)
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

