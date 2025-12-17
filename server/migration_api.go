package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/snowpackdata/cronos"
)

// MigrateChartOfAccountsHandler migrates existing GL accounts from journals to chart_of_accounts
func (a *App) MigrateChartOfAccountsHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Starting Chart of Accounts migration from existing journal entries...")

	// Get all unique accounts from both journals and offline_journals
	var accounts []string
	if err := a.cronosApp.DB.Raw(`
		SELECT DISTINCT account 
		FROM (
			SELECT account FROM journals WHERE account IS NOT NULL AND account != ''
			UNION
			SELECT account FROM offline_journals WHERE account IS NOT NULL AND account != ''
		) AS combined_accounts
		ORDER BY account
	`).Scan(&accounts).Error; err != nil {
		log.Printf("Failed to fetch unique accounts: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch unique accounts")
		return
	}

	log.Printf("Found %d unique accounts from journals and offline_journals", len(accounts))

	// Create ChartOfAccount for each unique account
	createdAccounts := 0
	skippedAccounts := 0
	for _, accountCode := range accounts {
		// Check if it already exists
		var existing cronos.ChartOfAccount
		err := a.cronosApp.DB.Where("account_code = ?", accountCode).First(&existing).Error
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

		if err := a.cronosApp.DB.Create(&account).Error; err != nil {
			log.Printf("Failed to create account %s: %v", accountCode, err)
			continue
		}

		log.Printf("Created account: %s (%s)", accountCode, accountName)
		createdAccounts++
	}

	log.Printf("Created %d new accounts, skipped %d existing accounts", createdAccounts, skippedAccounts)

	// Get all unique subaccounts per account from both journals and offline_journals
	type AccountSubaccount struct {
		Account    string
		SubAccount string
	}

	var subaccounts []AccountSubaccount
	if err := a.cronosApp.DB.Raw(`
		SELECT DISTINCT account, sub_account 
		FROM (
			SELECT account, sub_account FROM journals WHERE sub_account IS NOT NULL AND sub_account != ''
			UNION
			SELECT account, sub_account FROM offline_journals WHERE sub_account IS NOT NULL AND sub_account != ''
		) AS combined_subaccounts
		ORDER BY account, sub_account
	`).Scan(&subaccounts).Error; err != nil {
		log.Printf("Failed to fetch unique subaccounts: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch unique subaccounts")
		return
	}

	log.Printf("Found %d unique subaccount combinations from journals and offline_journals", len(subaccounts))

	// Create Subaccount for each unique subaccount
	createdSubaccounts := 0
	skippedSubaccounts := 0
	for _, sa := range subaccounts {
		// Check if the parent account exists
		var account cronos.ChartOfAccount
		if err := a.cronosApp.DB.Where("account_code = ?", sa.Account).First(&account).Error; err != nil {
			log.Printf("Account %s not found for subaccount %s, skipping", sa.Account, sa.SubAccount)
			continue
		}

		// Code should be the full value from journal (e.g., "37:Grid")
		// Name should be just the readable part (e.g., "Grid")
		var subCode, subName string
		subCode = sa.SubAccount // Keep the full code "37:Grid"
		if strings.Contains(sa.SubAccount, ":") {
			parts := strings.SplitN(sa.SubAccount, ":", 2)
			subName = parts[1] // Just the name "Grid"
		} else {
			// No colon, use as both code and name
			subName = sa.SubAccount
		}

		// Check if subaccount already exists
		var existing cronos.Subaccount
		err := a.cronosApp.DB.Where("account_code = ? AND code = ?", sa.Account, subCode).First(&existing).Error
		if err == nil {
			log.Printf("Subaccount %s:%s already exists, skipping", sa.Account, subCode)
			skippedSubaccounts++
			continue
		}

		subaccount := cronos.Subaccount{
			Code:        subCode,
			Name:        subName,
			AccountCode: sa.Account,
			Type:        "CUSTOM", // Default type
			IsActive:    true,
		}

		if err := a.cronosApp.DB.Create(&subaccount).Error; err != nil {
			log.Printf("Failed to create subaccount %s:%s: %v", sa.Account, subCode, err)
			continue
		}

		log.Printf("Created subaccount: code='%s' name='%s' account='%s'", subCode, subName, sa.Account)
		createdSubaccounts++
	}

	log.Printf("Created %d new subaccounts, skipped %d existing subaccounts", createdSubaccounts, skippedSubaccounts)
	
	// Now create subaccounts for all existing staff members
	log.Println("Creating subaccounts for existing staff members...")
	var employees []cronos.Employee
	if err := a.cronosApp.DB.Find(&employees).Error; err != nil {
		log.Printf("Failed to fetch employees: %v", err)
	} else {
		log.Printf("Found %d employees total", len(employees))
		staffAccountCodes := []string{"PAYROLL_EXPENSE", "ACCRUED_PAYROLL", "ACCOUNTS_PAYABLE"}
		for _, emp := range employees {
			empName := fmt.Sprintf("%s %s", emp.FirstName, emp.LastName)
			empCode := fmt.Sprintf("%d:%s", emp.ID, empName) // e.g., "1:Nate Robinson"
			log.Printf("Processing employee: %s (code='%s')", empName, empCode)
			
			for _, accountCode := range staffAccountCodes {
				var existing cronos.Subaccount
				if err := a.cronosApp.DB.Where("code = ? AND account_code = ?", empCode, accountCode).First(&existing).Error; err == nil {
					// Already exists
					log.Printf("  %s subaccount already exists for %s", accountCode, empName)
					continue
				}
				
				subaccount := cronos.Subaccount{
					Code:        empCode,
					Name:        empName,
					AccountCode: accountCode,
					Type:        "EMPLOYEE",
					IsActive:    true,
				}
				
				if err := a.cronosApp.DB.Create(&subaccount).Error; err != nil {
					log.Printf("  Failed to create %s subaccount for %s: %v", accountCode, empName, err)
				} else {
					log.Printf("  Created %s subaccount for %s", accountCode, empName)
					createdSubaccounts++
				}
			}
		}
	}
	
	// Create subaccounts for all existing client accounts
	log.Println("Creating subaccounts for existing client accounts...")
	var clientAccounts []cronos.Account
	if err := a.cronosApp.DB.Find(&clientAccounts).Error; err != nil {
		log.Printf("Failed to fetch accounts: %v", err)
	} else {
		clientAccountCodes := []string{"REVENUE", "ACCOUNTS_RECEIVABLE", "ACCRUED_RECEIVABLES"}
		for _, acc := range clientAccounts {
			accName := acc.Name
			accCode := fmt.Sprintf("%d:%s", acc.ID, accName) // e.g., "37:Grid"
			
			for _, accountCode := range clientAccountCodes {
				var existing cronos.Subaccount
				if err := a.cronosApp.DB.Where("code = ? AND account_code = ?", accCode, accountCode).First(&existing).Error; err == nil {
					// Already exists
					continue
				}
				
				subaccount := cronos.Subaccount{
					Code:        accCode,
					Name:        accName,
					AccountCode: accountCode,
					Type:        "CLIENT",
					IsActive:    true,
				}
				
				if err := a.cronosApp.DB.Create(&subaccount).Error; err != nil {
					log.Printf("Failed to create client subaccount %s:%s: %v", accountCode, accCode, err)
				} else {
					log.Printf("Created client subaccount: %s:%s (%s)", accountCode, accCode, accName)
					createdSubaccounts++
				}
			}
		}
	}
	
	log.Println("Chart of Accounts migration complete!")

	// Return summary
	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success":              true,
		"message":              "Chart of Accounts migration complete",
		"accounts_created":     createdAccounts,
		"accounts_skipped":     skippedAccounts,
		"subaccounts_created":  createdSubaccounts,
		"subaccounts_skipped":  skippedSubaccounts,
		"total_accounts":       len(accounts),
		"total_subaccounts":    len(subaccounts),
	})
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
				words = append(words, toLowerRune(r))
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
	runes[0] = toUpperRune(runes[0])
	return string(runes)
}

func toUpperRune(r rune) rune {
	if r >= 'a' && r <= 'z' {
		return r - 32
	}
	return r
}

func toLowerRune(r rune) rune {
	if r >= 'A' && r <= 'Z' {
		return r + 32
	}
	return r
}

// CleanupSubaccountsHandler fixes subaccounts where code contains ":" and should be split
func (a *App) CleanupSubaccountsHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Starting subaccount cleanup...")
	
	// Find all subaccounts where code contains ":"
	var badSubaccounts []cronos.Subaccount
	if err := a.cronosApp.DB.Where("code LIKE ?", "%:%").Find(&badSubaccounts).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch subaccounts: "+err.Error())
		return
	}
	
	log.Printf("Found %d subaccounts with ':' in code field", len(badSubaccounts))
	
	fixed := 0
	deleted := 0
	
	for _, sub := range badSubaccounts {
		// Code is correct as-is (e.g., "37:Grid")
		// But if name also contains "37:Grid", fix it to just "Grid"
		// This handles the case where name was incorrectly set to the full code
		
		if sub.Name == sub.Code && strings.Contains(sub.Code, ":") {
			parts := strings.SplitN(sub.Code, ":", 2)
			if len(parts) != 2 {
				log.Printf("Skipping %s - unexpected format", sub.Code)
				continue
			}
			
			properName := parts[1]  // "Grid"
			
			log.Printf("Processing subaccount ID=%d, code='%s' (keeping), name='%s' -> '%s'", sub.ID, sub.Code, sub.Name, properName)
			
			// Just update the name, keep code as-is
			if err := a.cronosApp.DB.Model(&sub).Updates(map[string]interface{}{
				"name": properName,
			}).Error; err != nil {
				log.Printf("Failed to fix subaccount ID=%d: %v", sub.ID, err)
			} else {
				log.Printf("Fixed subaccount: account='%s' code='%s' name='%s'", sub.AccountCode, sub.Code, properName)
				fixed++
			}
		} else {
			log.Printf("Subaccount ID=%d looks fine, skipping (code='%s', name='%s')", sub.ID, sub.Code, sub.Name)
		}
	}
	
	log.Printf("Cleanup complete: %d fixed, %d duplicates deleted", fixed, deleted)
	
	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Subaccount cleanup complete",
		"fixed":   fixed,
		"deleted": deleted,
	})
}

