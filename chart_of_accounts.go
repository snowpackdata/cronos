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
		{AccountCode: "OWNER_DISTRIBUTIONS", AccountName: "Owner Distributions", AccountType: "EXPENSE", IsSystemDefined: true, Description: "Distributions to owners/partners"},
		{AccountCode: "EXPENSE_PASS_THROUGH", AccountName: "Pass-Through Expenses", AccountType: "EXPENSE", IsSystemDefined: true, Description: "Client expenses to be reimbursed"},
		{AccountCode: "OTHER_EXPENSES", AccountName: "Other Expenses", AccountType: "EXPENSE", IsSystemDefined: true, Description: "Miscellaneous expenses"},

		// Equity
		{AccountCode: "EQUITY_OWNERSHIP", AccountName: "Equity - Ownership", AccountType: "EQUITY", IsSystemDefined: true, Description: "Owner equity"},
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
