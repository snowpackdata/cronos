package cronos

import (
	"fmt"
	"log"
)

// CreateChartOfAccount creates a new GL account
func (a *App) CreateChartOfAccount(accountCode, accountName, accountType, description string, parentID *uint) (*ChartOfAccount, error) {
	// Validate account type
	validTypes := map[string]bool{
		"ASSET":     true,
		"LIABILITY": true,
		"EQUITY":    true,
		"REVENUE":   true,
		"EXPENSE":   true,
	}
	if !validTypes[accountType] {
		return nil, fmt.Errorf("invalid account type: %s", accountType)
	}

	account := ChartOfAccount{
		AccountCode:     accountCode,
		AccountName:     accountName,
		AccountType:     accountType,
		ParentID:        parentID,
		IsActive:        true,
		Description:     description,
		IsSystemDefined: false,
	}

	if err := a.DB.Create(&account).Error; err != nil {
		return nil, fmt.Errorf("failed to create account: %w", err)
	}

	log.Printf("Created new account: %s (%s)", accountCode, accountName)
	return &account, nil
}

// GetChartOfAccounts retrieves all or filtered accounts
func (a *App) GetChartOfAccounts(accountType string, activeOnly bool) ([]ChartOfAccount, error) {
	var accounts []ChartOfAccount

	query := a.DB

	if accountType != "" {
		query = query.Where("account_type = ?", accountType)
	}

	if activeOnly {
		query = query.Where("is_active = ?", true)
	}

	if err := query.Order("account_type ASC, account_code ASC").Find(&accounts).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve accounts: %w", err)
	}

	return accounts, nil
}

// UpdateChartOfAccount updates an existing account
func (a *App) UpdateChartOfAccount(accountCode string, updates map[string]interface{}) error {
	result := a.DB.Model(&ChartOfAccount{}).Where("account_code = ?", accountCode).Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("failed to update account: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("account not found: %s", accountCode)
	}

	log.Printf("Updated account: %s", accountCode)
	return nil
}

// DeactivateChartOfAccount soft-deletes an account
func (a *App) DeactivateChartOfAccount(accountCode string) error {
	return a.UpdateChartOfAccount(accountCode, map[string]interface{}{"is_active": false})
}

// CreateSubaccount creates a new subaccount
func (a *App) CreateSubaccount(code, name, accountCode, subaccountType string) (*Subaccount, error) {
	// Validate that the account exists
	var account ChartOfAccount
	if err := a.DB.Where("account_code = ?", accountCode).First(&account).Error; err != nil {
		return nil, fmt.Errorf("account code not found: %s", accountCode)
	}

	subaccount := Subaccount{
		Code:        code,
		Name:        name,
		AccountCode: accountCode,
		Type:        subaccountType,
		IsActive:    true,
	}

	if err := a.DB.Create(&subaccount).Error; err != nil {
		return nil, fmt.Errorf("failed to create subaccount: %w", err)
	}

	log.Printf("Created new subaccount: %s (%s) for account %s", code, name, accountCode)
	return &subaccount, nil
}

// GetSubaccounts retrieves all or filtered subaccounts
func (a *App) GetSubaccounts(accountCode, subaccountType string, activeOnly bool) ([]Subaccount, error) {
	var subaccounts []Subaccount

	query := a.DB

	if accountCode != "" {
		query = query.Where("account_code = ?", accountCode)
	}

	if subaccountType != "" {
		query = query.Where("type = ?", subaccountType)
	}

	if activeOnly {
		query = query.Where("is_active = ?", true)
	}

	if err := query.Order("account_code ASC, code ASC").Find(&subaccounts).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve subaccounts: %w", err)
	}

	return subaccounts, nil
}

// UpdateSubaccount updates an existing subaccount
func (a *App) UpdateSubaccount(code string, updates map[string]interface{}) error {
	result := a.DB.Model(&Subaccount{}).Where("code = ?", code).Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("failed to update subaccount: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("subaccount not found: %s", code)
	}

	log.Printf("Updated subaccount: %s", code)
	return nil
}

// DeactivateSubaccount soft-deletes a subaccount
func (a *App) DeactivateSubaccount(code string) error {
	return a.UpdateSubaccount(code, map[string]interface{}{"is_active": false})
}

// SeedSystemAccounts creates the predefined system accounts in the database
// This should be called during initial setup or migration
func (a *App) SeedSystemAccounts() error {
	systemAccounts := []ChartOfAccount{
		// Assets
		{AccountCode: "ACCRUED_RECEIVABLES", AccountName: "Accrued Receivables", AccountType: "ASSET", IsSystemDefined: true, Description: "Work completed but not yet invoiced"},
		{AccountCode: "ACCOUNTS_RECEIVABLE", AccountName: "Accounts Receivable", AccountType: "ASSET", IsSystemDefined: true, Description: "Invoiced amounts owed by clients"},
		{AccountCode: "CASH", AccountName: "Cash", AccountType: "ASSET", IsSystemDefined: true, Description: "Cash in bank accounts"},
		{AccountCode: "EQUIPMENT", AccountName: "Equipment", AccountType: "ASSET", IsSystemDefined: true, Description: "Computer equipment and hardware"},
		{AccountCode: "EQUITY_POOL", AccountName: "Equity Pool", AccountType: "ASSET", IsSystemDefined: true, Description: "Available equity for distribution"},
		{AccountCode: "OTHER_ASSETS", AccountName: "Other Assets", AccountType: "ASSET", IsSystemDefined: true, Description: "Miscellaneous assets"},

		// Liabilities
		{AccountCode: "ACCRUED_PAYROLL", AccountName: "Accrued Payroll", AccountType: "LIABILITY", IsSystemDefined: true, Description: "Payroll owed but not yet paid"},
		{AccountCode: "ACCOUNTS_PAYABLE", AccountName: "Accounts Payable", AccountType: "LIABILITY", IsSystemDefined: true, Description: "Bills owed to employees and vendors"},
		{AccountCode: "ACCRUED_EXPENSES_PAYABLE", AccountName: "Accrued Expenses Payable", AccountType: "LIABILITY", IsSystemDefined: true, Description: "Expenses recorded but not yet reconciled with bank statements"},
		{AccountCode: "CREDIT_CARD_PAYABLE", AccountName: "Credit Card Payable", AccountType: "LIABILITY", IsSystemDefined: true, Description: "Credit card balances"},
		{AccountCode: "OTHER_LIABILITIES", AccountName: "Other Liabilities", AccountType: "LIABILITY", IsSystemDefined: true, Description: "Miscellaneous liabilities"},

		// Revenue
		{AccountCode: "REVENUE", AccountName: "Revenue", AccountType: "REVENUE", IsSystemDefined: true, Description: "Client billable revenue"},
		{AccountCode: "ADJUSTMENT_REVENUE", AccountName: "Adjustment Revenue", AccountType: "REVENUE", IsSystemDefined: true, Description: "Additional fees and charges"},
		{AccountCode: "OTHER_INCOME", AccountName: "Other Income", AccountType: "REVENUE", IsSystemDefined: true, Description: "Miscellaneous income"},

		// Contra-Revenue
		{AccountCode: "CREDITS_ISSUED", AccountName: "Credits Issued", AccountType: "REVENUE", IsSystemDefined: true, Description: "Credits and refunds to clients (contra-revenue)"},
		{AccountCode: "DISCOUNTS", AccountName: "Discounts", AccountType: "REVENUE", IsSystemDefined: true, Description: "Discounts given to clients (contra-revenue)"},

		// Expenses
		{AccountCode: "PAYROLL_EXPENSE", AccountName: "Payroll Expense", AccountType: "EXPENSE", IsSystemDefined: true, Description: "Employee compensation"},
		{AccountCode: "ADJUSTMENT_EXPENSE", AccountName: "Adjustment Expense", AccountType: "EXPENSE", IsSystemDefined: true, Description: "Adjustments to payroll"},
		{AccountCode: "EQUIPMENT_EXPENSE", AccountName: "Equipment Expense", AccountType: "EXPENSE", IsSystemDefined: true, Description: "Equipment depreciation and expense"},

		// Operating Expenses
		{AccountCode: "OPERATING_EXPENSES_SAAS", AccountName: "Operating Expenses - SaaS", AccountType: "EXPENSE", IsSystemDefined: true, Description: "Software and SaaS subscriptions"},
		{AccountCode: "OPERATING_EXPENSES_TRAVEL", AccountName: "Operating Expenses - Travel", AccountType: "EXPENSE", IsSystemDefined: true, Description: "Business travel expenses"},
		{AccountCode: "OPERATING_EXPENSES_EQUIPMENT", AccountName: "Operating Expenses - Equipment", AccountType: "EXPENSE", IsSystemDefined: true, Description: "Equipment purchases"},
		{AccountCode: "OPERATING_EXPENSES_FEES", AccountName: "Operating Expenses - Fees", AccountType: "EXPENSE", IsSystemDefined: true, Description: "Banking and transaction fees"},
		{AccountCode: "OPERATING_EXPENSES_LEGAL", AccountName: "Operating Expenses - Legal", AccountType: "EXPENSE", IsSystemDefined: true, Description: "Legal and professional fees"},
		{AccountCode: "OPERATING_EXPENSES_DISCRETIONARY", AccountName: "Operating Expenses - Discretionary", AccountType: "EXPENSE", IsSystemDefined: true, Description: "Discretionary business expenses"},
		{AccountCode: "OPERATING_EXPENSES_TAXES", AccountName: "Operating Expenses - Taxes", AccountType: "EXPENSE", IsSystemDefined: true, Description: "Business taxes and licenses"},
		{AccountCode: "OPERATING_EXPENSES_VENDORS", AccountName: "Operating Expenses - Vendors", AccountType: "EXPENSE", IsSystemDefined: true, Description: "Vendor and contractor payments"},
		{AccountCode: "OPERATING_EXPENSES_OFFICE", AccountName: "Operating Expenses - Office", AccountType: "EXPENSE", IsSystemDefined: true, Description: "Office supplies and expenses"},
		{AccountCode: "EXPENSE_PASS_THROUGH", AccountName: "Pass-Through Expenses", AccountType: "EXPENSE", IsSystemDefined: true, Description: "Client expenses to be reimbursed"},
		{AccountCode: "OTHER_EXPENSES", AccountName: "Other Expenses", AccountType: "EXPENSE", IsSystemDefined: true, Description: "Miscellaneous expenses"},

		// Equity
		{AccountCode: "EQUITY_OWNERSHIP", AccountName: "Equity - Ownership", AccountType: "EQUITY", IsSystemDefined: true, Description: "Owner equity"},
		{AccountCode: "OWNER_DISTRIBUTIONS", AccountName: "Owner Distributions", AccountType: "EQUITY", IsSystemDefined: true, Description: "Draws/distributions to owners (reduces equity)"},
		{AccountCode: "EQUITY", AccountName: "Equity", AccountType: "EQUITY", IsSystemDefined: true, Description: "General equity"},

		// Unclassified
		{AccountCode: "UNCLASSIFIED", AccountName: "Unclassified", AccountType: "EXPENSE", IsSystemDefined: true, Description: "Transactions needing categorization"},
	}

	created := 0
	skipped := 0

	for _, account := range systemAccounts {
		// Check if already exists
		var existing ChartOfAccount
		err := a.DB.Where("account_code = ?", account.AccountCode).First(&existing).Error
		if err == nil {
			skipped++
			continue
		}

		account.IsActive = true
		if err := a.DB.Create(&account).Error; err != nil {
			log.Printf("Error seeding account %s: %v", account.AccountCode, err)
			continue
		}
		created++
	}

	log.Printf("Seeded system accounts: %d created, %d skipped (already exist)", created, skipped)
	return nil
}

// GetAccountForCode retrieves account details by code
func (a *App) GetAccountForCode(accountCode string) (*ChartOfAccount, error) {
	var account ChartOfAccount
	if err := a.DB.Where("account_code = ?", accountCode).First(&account).Error; err != nil {
		return nil, fmt.Errorf("account not found: %s", accountCode)
	}
	return &account, nil
}

// GetSubaccountForCode retrieves subaccount details by code
func (a *App) GetSubaccountForCode(code string) (*Subaccount, error) {
	var subaccount Subaccount
	if err := a.DB.Where("code = ?", code).First(&subaccount).Error; err != nil {
		return nil, fmt.Errorf("subaccount not found: %s", code)
	}
	return &subaccount, nil
}

// MarkEmployeesAsOwners marks specific employees as owners based on their IDs
// This is a helper migration to set the IsOwner flag for existing employees
func (a *App) MarkEmployeesAsOwners(employeeIDs []uint) error {
	if len(employeeIDs) == 0 {
		log.Println("No employee IDs provided for owner marking")
		return nil
	}

	result := a.DB.Model(&Employee{}).
		Where("id IN ?", employeeIDs).
		Update("is_owner", true)

	if result.Error != nil {
		return fmt.Errorf("failed to mark employees as owners: %w", result.Error)
	}

	log.Printf("Marked %d employees as owners (IDs: %v)", result.RowsAffected, employeeIDs)
	return nil
}

// ReclassifyOwnerPayrollToDistributions reclassifies historical PAYROLL_EXPENSE journal entries
// for employees marked as owners to OWNER_DISTRIBUTIONS (equity)
// This migration updates existing journal entries to reflect proper accounting treatment
func (a *App) ReclassifyOwnerPayrollToDistributions() error {
	log.Println("Running migration: Reclassify historical owner payroll expenses to equity distributions")

	// Get all employees marked as owners
	var owners []Employee
	if err := a.DB.Where("is_owner = ?", true).Find(&owners).Error; err != nil {
		return fmt.Errorf("failed to find owner employees: %w", err)
	}

	if len(owners) == 0 {
		log.Println("No owners found - skipping payroll reclassification")
		return nil
	}

	ownerIDs := make([]uint, len(owners))
	for i, owner := range owners {
		ownerIDs[i] = owner.ID
		log.Printf("  Owner %d: %s %s (ID: %d)", i+1, owner.FirstName, owner.LastName, owner.ID)
	}

	log.Printf("Found %d owners to reclassify", len(owners))

	// Find all journal entries with PAYROLL_EXPENSE account that are linked to bills for these owners
	// We need to join through bills to get the employee_id
	var journalsToUpdate []Journal
	err := a.DB.
		Joins("INNER JOIN bills ON bills.id = journals.bill_id").
		Where("journals.account = ? AND bills.employee_id IN ?", "PAYROLL_EXPENSE", ownerIDs).
		Find(&journalsToUpdate).Error

	if err != nil {
		return fmt.Errorf("failed to find owner payroll journal entries: %w", err)
	}

	if len(journalsToUpdate) == 0 {
		log.Println("No historical payroll expense entries found for owners - nothing to reclassify")
		return nil
	}

	log.Printf("Found %d journal entries to reclassify from PAYROLL_EXPENSE to OWNER_DISTRIBUTIONS", len(journalsToUpdate))

	// Update all matching journal entries
	journalIDs := make([]uint, len(journalsToUpdate))
	for i, j := range journalsToUpdate {
		journalIDs[i] = j.ID
	}

	result := a.DB.Model(&Journal{}).
		Where("id IN ?", journalIDs).
		Update("account", "OWNER_DISTRIBUTIONS")

	if result.Error != nil {
		return fmt.Errorf("failed to update journal entries: %w", result.Error)
	}

	log.Printf("✓ Successfully reclassified %d journal entries from PAYROLL_EXPENSE to OWNER_DISTRIBUTIONS", result.RowsAffected)

	// Also check for matching ACCRUED_PAYROLL entries (the credit side of payroll accrual)
	// These should be paired with the payroll expense entries
	var accruedPayrollCount int64
	a.DB.Model(&Journal{}).
		Joins("INNER JOIN bills ON bills.id = journals.bill_id").
		Where("journals.account = ? AND bills.employee_id IN ?", "ACCRUED_PAYROLL", ownerIDs).
		Count(&accruedPayrollCount)

	if accruedPayrollCount > 0 {
		log.Printf("Note: Found %d ACCRUED_PAYROLL entries for owners (these remain as-is for AP tracking)", accruedPayrollCount)
	}

	return nil
}

// MigrateOwnerDistributionsToEquity migrates the OWNER_DISTRIBUTIONS account from EXPENSE to EQUITY
// This is a one-time migration for issue #67
// Also updates existing journal entries to reflect the new classification
func (a *App) MigrateOwnerDistributionsToEquity() error {
	log.Println("Running migration: Update OWNER_DISTRIBUTIONS account type from EXPENSE to EQUITY")

	// Step 1: Update the Chart of Accounts entry
	result := a.DB.Model(&ChartOfAccount{}).
		Where("account_code = ? AND account_type = ?", "OWNER_DISTRIBUTIONS", "EXPENSE").
		Updates(map[string]interface{}{
			"account_type": "EQUITY",
			"description":  "Draws/distributions to owners (reduces equity)",
		})

	if result.Error != nil {
		return fmt.Errorf("failed to migrate OWNER_DISTRIBUTIONS chart of accounts: %w", result.Error)
	}

	if result.RowsAffected > 0 {
		log.Printf("Migrated OWNER_DISTRIBUTIONS account type from EXPENSE to EQUITY (%d chart of accounts rows affected)", result.RowsAffected)
	} else {
		log.Println("OWNER_DISTRIBUTIONS account already migrated or does not exist in chart of accounts")
	}

	// Step 2: Update existing Journal entries that use OWNER_DISTRIBUTIONS
	// This allows us to see the impact on the general ledger
	var journalCount int64
	err := a.DB.Model(&Journal{}).
		Where("account = ?", "OWNER_DISTRIBUTIONS").
		Count(&journalCount).Error

	if err != nil {
		log.Printf("Warning: Failed to count OWNER_DISTRIBUTIONS journal entries: %v", err)
	} else if journalCount > 0 {
		log.Printf("Found %d existing journal entries with OWNER_DISTRIBUTIONS account - these are now classified as EQUITY", journalCount)
		// Note: We don't need to UPDATE the journal entries themselves
		// The classification comes from the Chart of Accounts lookup
		// Just logging the count so we know what's being reclassified
	}

	// Step 3: Update existing OfflineJournal entries if any
	var offlineCount int64
	err = a.DB.Model(&OfflineJournal{}).
		Where("account = ?", "OWNER_DISTRIBUTIONS").
		Count(&offlineCount).Error

	if err != nil {
		log.Printf("Warning: Failed to count OWNER_DISTRIBUTIONS offline journal entries: %v", err)
	} else if offlineCount > 0 {
		log.Printf("Found %d existing offline journal entries with OWNER_DISTRIBUTIONS account - these are now classified as EQUITY", offlineCount)
	}

	totalAffected := journalCount + offlineCount
	if totalAffected > 0 {
		log.Printf("Migration complete: %d total journal entries are now classified under EQUITY instead of EXPENSE", totalAffected)
	}

	return nil
}
