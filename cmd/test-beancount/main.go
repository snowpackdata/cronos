package main

import (
	"fmt"
	"log"
	"time"

	"github.com/snowpackdata/cronos"
)

func main() {
	fmt.Println("Beancount Parser Test")
	fmt.Println("=====================\n")

	// Parse the Beancount file
	beancountPath := "/Users/naterobinson/Projects/snowpack/finances/finances.beancount"
	fmt.Printf("Parsing: %s\n\n", beancountPath)

	ledger, err := cronos.ParseBeancountFile(beancountPath)
	if err != nil {
		log.Fatalf("Failed to parse Beancount file: %v", err)
	}

	// Print summary statistics
	fmt.Printf("Summary:\n")
	fmt.Printf("  Total Transactions: %d\n", len(ledger.Transactions))
	fmt.Printf("  Total Accounts: %d\n", len(ledger.Accounts))
	fmt.Printf("  Total Balance Assertions: %d\n\n", len(ledger.Balances))

	// Validate all transactions
	fmt.Println("Validating transactions...")
	errors := ledger.ValidateAll()
	if len(errors) > 0 {
		fmt.Printf("\nValidation Errors Found: %d\n", len(errors))
		for i, err := range errors {
			if i < 10 { // Show first 10 errors
				fmt.Printf("  - %v\n", err)
			}
		}
		if len(errors) > 10 {
			fmt.Printf("  ... and %d more errors\n", len(errors)-10)
		}
	} else {
		fmt.Println("✓ All transactions validated successfully!")
	}

	// Show sample transactions
	fmt.Println("\nSample Transactions (first 5):")
	for i, tx := range ledger.Transactions {
		if i >= 5 {
			break
		}
		fmt.Printf("\n%d. %s - %s\n", i+1, tx.Date.Format("2006-01-02"), tx.Description)
		for _, posting := range tx.Postings {
			fmt.Printf("   %-45s %10.2f %s\n", posting.Account, posting.Amount, posting.Currency)
		}
	}

	// Convert to ledger entries
	fmt.Println("\n\nConverting to Ledger Entries...")
	entries := cronos.ConvertBeancountToLedgerEntries(ledger)
	fmt.Printf("Total Ledger Entries: %d\n", len(entries))

	// Show sample converted entries
	fmt.Println("\nSample Converted Entries (first 5):")
	for i, entry := range entries {
		if i >= 5 {
			break
		}
		fmt.Printf("%d. %s | %s:%s | DR:%.2f CR:%.2f | %s\n",
			i+1,
			entry.Date.Format("2006-01-02"),
			entry.Account,
			entry.SubAccount,
			entry.Debit,
			entry.Credit,
			entry.Description,
		)
	}

	// Calculate cash balance
	fmt.Println("\n\nCalculating Cash Balance...")
	var cashBalance float64
	for _, tx := range ledger.Transactions {
		for _, posting := range tx.Postings {
			if posting.Account == "Assets:Checking:ChaseBusiness" {
				cashBalance += posting.Amount
			}
		}
	}
	fmt.Printf("Current Cash Balance (from Beancount): $%.2f\n", cashBalance)

	// Check latest balance assertion
	if len(ledger.Balances) > 0 {
		latestBalance := ledger.Balances[len(ledger.Balances)-1]
		fmt.Printf("\nLatest Balance Assertion:\n")
		fmt.Printf("  Date: %s\n", latestBalance.Date.Format("2006-01-02"))
		fmt.Printf("  Account: %s\n", latestBalance.Account)
		fmt.Printf("  Expected: $%.2f\n", latestBalance.Amount)

		// Calculate actual balance up to that date
		var actualBalance float64
		for _, tx := range ledger.Transactions {
			if tx.Date.After(latestBalance.Date) {
				continue
			}
			for _, posting := range tx.Postings {
				if posting.Account == latestBalance.Account {
					actualBalance += posting.Amount
				}
			}
		}
		fmt.Printf("  Actual: $%.2f\n", actualBalance)
		diff := actualBalance - latestBalance.Amount
		if diff > 0.01 || diff < -0.01 {
			fmt.Printf("  ⚠️  Mismatch: $%.2f\n", diff)
		} else {
			fmt.Printf("  ✓ Balance assertion passed!\n")
		}
	}

	// Account mapping test
	fmt.Println("\n\nAccount Mapping Examples:")
	testAccounts := []string{
		"Assets:Checking:ChaseBusiness",
		"Income:ClientBillables:Vanta",
		"Expenses:SaaS:GoogleCloud",
		"Liabilities:CreditCard:ChaseCredit",
		"Equity:Ownership:NateRobinson",
	}
	for _, account := range testAccounts {
		mapped := cronos.MapBeancountAccount(account)
		subAccount := cronos.ExtractSubAccount(account)
		fmt.Printf("  %-40s → %-30s (sub: %s)\n", account, mapped, subAccount)
	}

	// Time range analysis
	fmt.Println("\n\nTransaction Date Range:")
	if len(ledger.Transactions) > 0 {
		firstTx := ledger.Transactions[0]
		lastTx := ledger.Transactions[len(ledger.Transactions)-1]
		fmt.Printf("  First: %s\n", firstTx.Date.Format("2006-01-02"))
		fmt.Printf("  Last: %s\n", lastTx.Date.Format("2006-01-02"))

		// Count by year
		yearCounts := make(map[int]int)
		for _, tx := range ledger.Transactions {
			yearCounts[tx.Date.Year()]++
		}
		fmt.Println("\n  Transactions by Year:")
		for year := 2023; year <= time.Now().Year(); year++ {
			if count, ok := yearCounts[year]; ok {
				fmt.Printf("    %d: %d transactions\n", year, count)
			}
		}
	}

	fmt.Println("\n✓ Test completed successfully!")
}
