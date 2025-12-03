package main

import (
	"log"
	"os"

	"github.com/snowpackdata/cronos"
)

// MigrateChartOfAccountsFromJournal creates ChartOfAccount and Subaccount records
// from existing Journal entries in the database
func main() {
	log.Println("Starting Chart of Accounts migration from existing journal entries...")

	// Initialize the cronos app based on environment
	var app cronos.App
	
	if os.Getenv("ENVIRONMENT") == "production" {
		// Production mode
		dbURI := os.Getenv("CLOUD_SQL_CONNECTION_NAME") + "/" + os.Getenv("CLOUD_SQL_DATABASE_NAME") + "?user=" + os.Getenv("CLOUD_SQL_USERNAME") + "&password=" + os.Getenv("CLOUD_SQL_PASSWORD")
		log.Println("Connecting to production database...")
		app.InitializeCloud(dbURI)
	} else {
		// Development/local mode
		user := os.Getenv("CLOUD_SQL_USERNAME")
		password := os.Getenv("CLOUD_SQL_PASSWORD") // Can be empty for local postgres
		dbHost := os.Getenv("CLOUD_SQL_CONNECTION_NAME")
		databaseName := os.Getenv("CLOUD_SQL_DATABASE_NAME")
		
		if user == "" || dbHost == "" || databaseName == "" {
			log.Fatal("Missing required environment variables: CLOUD_SQL_USERNAME, CLOUD_SQL_CONNECTION_NAME, CLOUD_SQL_DATABASE_NAME")
		}
		
		log.Printf("Connecting to local/development database (host=%s, db=%s, user=%s)...", dbHost, databaseName, user)
		app.InitializeLocal(user, password, dbHost, databaseName)
	}

	log.Println("Database connected successfully")

	// Get all unique accounts from journal
	var accounts []string
	if err := app.DB.Raw(`
		SELECT DISTINCT account 
		FROM journals 
		WHERE account IS NOT NULL AND account != '' 
		ORDER BY account
	`).Scan(&accounts).Error; err != nil {
		log.Fatalf("Failed to fetch unique accounts: %v", err)
	}

	log.Printf("Found %d unique accounts in journal", len(accounts))

	// Create ChartOfAccount for each unique account
	createdAccounts := 0
	skippedAccounts := 0
	for _, accountCode := range accounts {
		// Check if it already exists
		var existing cronos.ChartOfAccount
		err := app.DB.Where("account_code = ?", accountCode).First(&existing).Error
		if err == nil {
			log.Printf("Account %s already exists, skipping", accountCode)
			skippedAccounts++
			continue
		}

		// Generate a name from the account code
		accountName := generateAccountName(accountCode)

		account := cronos.ChartOfAccount{
			AccountCode:     accountCode,
			AccountName:     accountName,
			AccountType:     "ASSET", // Default, will need manual review
			IsActive:        true,
			IsSystemDefined: false, // These are imported, not system accounts
		}

		if err := app.DB.Create(&account).Error; err != nil {
			log.Printf("Failed to create account %s: %v", accountCode, err)
			continue
		}

		log.Printf("Created account: %s (%s)", accountCode, accountName)
		createdAccounts++
	}

	log.Printf("Created %d new accounts, skipped %d existing accounts", createdAccounts, skippedAccounts)

	// Get all unique subaccounts per account from journal
	type AccountSubaccount struct {
		Account    string
		SubAccount string
	}

	var subaccounts []AccountSubaccount
	if err := app.DB.Raw(`
		SELECT DISTINCT account, sub_account 
		FROM journals 
		WHERE sub_account IS NOT NULL AND sub_account != '' 
		ORDER BY account, sub_account
	`).Scan(&subaccounts).Error; err != nil {
		log.Fatalf("Failed to fetch unique subaccounts: %v", err)
	}

	log.Printf("Found %d unique subaccount combinations in journal", len(subaccounts))

	// Create Subaccount for each unique subaccount
	createdSubaccounts := 0
	skippedSubaccounts := 0
	for _, sa := range subaccounts {
		// Check if the parent account exists
		var account cronos.ChartOfAccount
		if err := app.DB.Where("account_code = ?", sa.Account).First(&account).Error; err != nil {
			log.Printf("Account %s not found for subaccount %s, skipping", sa.Account, sa.SubAccount)
			continue
		}

		// Check if subaccount already exists
		var existing cronos.Subaccount
		err := app.DB.Where("account_code = ? AND code = ?", sa.Account, sa.SubAccount).First(&existing).Error
		if err == nil {
			log.Printf("Subaccount %s:%s already exists, skipping", sa.Account, sa.SubAccount)
			skippedSubaccounts++
			continue
		}

		subaccount := cronos.Subaccount{
			Code:        sa.SubAccount,
			Name:        sa.SubAccount, // Use code as name, can be updated manually
			AccountCode: sa.Account,
			Type:        "CUSTOM", // Default type
			IsActive:    true,
		}

		if err := app.DB.Create(&subaccount).Error; err != nil {
			log.Printf("Failed to create subaccount %s:%s: %v", sa.Account, sa.SubAccount, err)
			continue
		}

		log.Printf("Created subaccount: %s:%s", sa.Account, sa.SubAccount)
		createdSubaccounts++
	}

	log.Printf("Created %d new subaccounts, skipped %d existing subaccounts", createdSubaccounts, skippedSubaccounts)
	log.Println("Chart of Accounts migration complete!")
}

// generateAccountName creates a human-readable name from an account code
func generateAccountName(code string) string {
	// Simple name generation: replace underscores with spaces and title case
	// e.g., "ACCRUED_RECEIVABLES" -> "Accrued Receivables"
	name := ""
	words := []rune{}
	
	for i, r := range code {
		if r == '_' {
			if len(words) > 0 {
				if name != "" {
					name += " "
				}
				name += titleCase(string(words))
				words = []rune{}
			}
		} else {
			if i == 0 || code[i-1] == '_' {
				words = append(words, r) // Keep first letter as-is (uppercase)
			} else {
				words = append(words, toLower(r))
			}
		}
	}
	
	// Add remaining words
	if len(words) > 0 {
		if name != "" {
			name += " "
		}
		name += titleCase(string(words))
	}
	
	return name
}

func titleCase(s string) string {
	if len(s) == 0 {
		return s
	}
	runes := []rune(s)
	runes[0] = toUpper(runes[0])
	return string(runes)
}

func toUpper(r rune) rune {
	if r >= 'a' && r <= 'z' {
		return r - 32
	}
	return r
}

func toLower(r rune) rune {
	if r >= 'A' && r <= 'Z' {
		return r + 32
	}
	return r
}

