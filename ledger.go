package cronos

import (
	"fmt"
	"log"
	"time"
)

// LedgerEntry represents a unified ledger entry from either Beancount or Journal DB
type LedgerEntry struct {
	Date        time.Time `json:"date"`
	Account     string    `json:"account"`
	SubAccount  string    `json:"sub_account"`
	Description string    `json:"description"`
	Debit       float64   `json:"debit"`  // In dollars
	Credit      float64   `json:"credit"` // In dollars
	Source      string    `json:"source"` // "beancount" or "journal_db"
	InvoiceID   *uint     `json:"invoice_id,omitempty"`
	BillID      *uint     `json:"bill_id,omitempty"`
	Tags        []string  `json:"tags,omitempty"`
}

// AccountMapping maps Beancount accounts to Journal account types
var BeancountToJournalAccountMap = map[string]string{
	// Assets
	"Assets:Checking:ChaseBusiness":        "CASH",
	"Assets:Equipment:Hardware":            "EQUIPMENT",
	"Assets:Ownership:AvailableEquityPool": "EQUITY_POOL",

	// Liabilities
	"Liabilities:CreditCard:ChaseCredit": "CREDIT_CARD_PAYABLE",

	// Income (mapped as contra to maintain debit/credit semantics)
	"Income:ClientBillables:*": "REVENUE",
	"Income:ACHVerification:*": "REVENUE",

	// Equity
	"Equity:Ownership:*":      "EQUITY_OWNERSHIP",
	"Equity:CompanyFormation": "EQUITY_OWNERSHIP",

	// Expenses
	"Expenses:Distributions:*":    "OWNER_DISTRIBUTIONS",
	"Expenses:Payroll:*":          "PAYROLL_EXPENSE",
	"Expenses:Equipment:Hardware": "EQUIPMENT_EXPENSE",
	"Expenses:Fees:*":             "OPERATING_EXPENSES_FEES",
	"Expenses:Legal:*":            "OPERATING_EXPENSES_LEGAL",
	"Expenses:SaaS:*":             "OPERATING_EXPENSES_SAAS",
	"Expenses:Travel:*":           "OPERATING_EXPENSES_TRAVEL",
	"Expenses:Discretionary:*":    "OPERATING_EXPENSES_DISCRETIONARY",
	"Expenses:Taxes:*":            "OPERATING_EXPENSES_TAXES",
	"Expenses:Vendors:*":          "OPERATING_EXPENSES_VENDORS",
	"Expenses:Office:*":           "OPERATING_EXPENSES_OFFICE",
}

// MapBeancountAccount maps a Beancount account name to a Journal account type
func MapBeancountAccount(beancountAccount string) string {
	// Try exact match first
	if mapped, ok := BeancountToJournalAccountMap[beancountAccount]; ok {
		return mapped
	}

	// Try wildcard patterns
	for pattern, mapped := range BeancountToJournalAccountMap {
		if len(pattern) > 0 && pattern[len(pattern)-1] == '*' {
			prefix := pattern[:len(pattern)-1]
			if len(beancountAccount) >= len(prefix) && beancountAccount[:len(prefix)] == prefix {
				return mapped
			}
		}
	}

	// Default: map to a generic account based on root type
	if len(beancountAccount) > 0 {
		switch {
		case beancountAccount[:6] == "Assets":
			return "OTHER_ASSETS"
		case beancountAccount[:11] == "Liabilities":
			return "OTHER_LIABILITIES"
		case beancountAccount[:6] == "Income":
			return "OTHER_INCOME"
		case beancountAccount[:6] == "Equity":
			return "EQUITY"
		case beancountAccount[:8] == "Expenses":
			return "OTHER_EXPENSES"
		}
	}

	return "UNCLASSIFIED"
}

// ExtractSubAccount extracts a sub-account identifier from Beancount account
// For example: "Income:ClientBillables:Vanta" -> "Vanta"
func ExtractSubAccount(beancountAccount string) string {
	parts := splitAccountName(beancountAccount)
	if len(parts) >= 3 {
		return parts[2]
	}
	if len(parts) >= 2 {
		return parts[1]
	}
	return ""
}

func splitAccountName(account string) []string {
	return splitOnColon(account)
}

func splitOnColon(s string) []string {
	var result []string
	var current string

	for _, char := range s {
		if char == ':' {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}

	if current != "" {
		result = append(result, current)
	}

	return result
}

// ConvertBeancountToLedgerEntries converts Beancount transactions to ledger entries
func ConvertBeancountToLedgerEntries(ledger *BeancountLedger) []LedgerEntry {
	var entries []LedgerEntry

	for _, tx := range ledger.Transactions {
		for _, posting := range tx.Postings {
			// Skip zero amounts
			if posting.Amount == 0 {
				continue
			}

			account := MapBeancountAccount(posting.Account)
			subAccount := ExtractSubAccount(posting.Account)

			// In Beancount, positive amounts are debits for Assets/Expenses
			// and credits for Liabilities/Equity/Income
			// We need to convert to standard debit/credit
			debit, credit := convertBeancountAmountToDebitCredit(posting.Account, posting.Amount)

			entry := LedgerEntry{
				Date:        tx.Date,
				Account:     account,
				SubAccount:  subAccount,
				Description: tx.Description,
				Debit:       debit,
				Credit:      credit,
				Source:      "beancount",
				Tags:        tx.Tags,
			}

			entries = append(entries, entry)
		}
	}

	return entries
}

// convertBeancountAmountToDebitCredit converts Beancount amount to debit/credit
func convertBeancountAmountToDebitCredit(account string, amount float64) (debit float64, credit float64) {
	// In Beancount, signs work differently for different account types:
	// Assets/Expenses: positive = debit (increase), negative = credit (decrease)
	// Income: positive = credit (increase), negative = debit (decrease)
	// Liabilities: negative = credit (increase), positive = debit (decrease)
	// Equity: similar to Liabilities

	// Check account type
	if len(account) >= 6 && account[:6] == "Assets" {
		// Assets: positive = debit, negative = credit
		if amount > 0 {
			return amount, 0
		}
		return 0, -amount
	} else if len(account) >= 8 && account[:8] == "Expenses" {
		// Expenses: positive = debit, negative = credit
		if amount > 0 {
			return amount, 0
		}
		return 0, -amount
	} else if len(account) >= 6 && account[:6] == "Income" {
		// Income: positive = credit, negative = debit
		if amount > 0 {
			return 0, amount
		}
		return -amount, 0
	} else if len(account) >= 11 && account[:11] == "Liabilities" {
		// Liabilities: negative = credit (increase), positive = debit (decrease)
		if amount > 0 {
			return amount, 0
		}
		return 0, -amount
	} else if len(account) >= 6 && account[:6] == "Equity" {
		// Equity: negative = credit (increase), positive = debit (decrease)
		if amount > 0 {
			return amount, 0
		}
		return 0, -amount
	}

	// Default (shouldn't reach here)
	log.Printf("Warning: unrecognized account type for %s", account)
	if amount > 0 {
		return amount, 0
	}
	return 0, -amount
}

// ConvertJournalToLedgerEntries converts Journal DB entries to ledger entries
func ConvertJournalToLedgerEntries(journals []Journal) []LedgerEntry {
	var entries []LedgerEntry

	for _, j := range journals {
		entry := LedgerEntry{
			Date:        j.CreatedAt,
			Account:     j.Account,
			SubAccount:  j.SubAccount,
			Description: j.Memo,
			Debit:       float64(j.Debit) / 100.0,  // Convert cents to dollars
			Credit:      float64(j.Credit) / 100.0, // Convert cents to dollars
			Source:      "journal_db",
			InvoiceID:   j.InvoiceID,
			BillID:      j.BillID,
		}

		entries = append(entries, entry)
	}

	return entries
}

// GetCombinedGeneralLedger retrieves and combines all ledger entries from both sources
func (a *App) GetCombinedGeneralLedger(beancountPath string, startDate, endDate time.Time) ([]LedgerEntry, error) {
	var allEntries []LedgerEntry

	// Load and parse Beancount file
	beancountLedger, err := ParseBeancountFile(beancountPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse beancount file: %w", err)
	}

	// Validate Beancount transactions
	validationErrors := beancountLedger.ValidateAll()
	if len(validationErrors) > 0 {
		log.Printf("Warning: %d Beancount validation errors found", len(validationErrors))
		for _, err := range validationErrors {
			log.Printf("  - %v", err)
		}
	}

	// Convert Beancount to ledger entries
	beancountEntries := ConvertBeancountToLedgerEntries(beancountLedger)

	// Filter by date range
	var filteredBeancount []LedgerEntry
	for _, entry := range beancountEntries {
		if (entry.Date.Equal(startDate) || entry.Date.After(startDate)) &&
			(entry.Date.Equal(endDate) || entry.Date.Before(endDate)) {
			filteredBeancount = append(filteredBeancount, entry)
		}
	}

	// Load Journal DB entries
	var journals []Journal
	err = a.DB.Where("created_at >= ? AND created_at <= ?", startDate, endDate).
		Order("created_at ASC").
		Find(&journals).Error
	if err != nil {
		return nil, fmt.Errorf("failed to load journal entries: %w", err)
	}

	// Convert Journal DB to ledger entries
	journalEntries := ConvertJournalToLedgerEntries(journals)

	// Remove duplicate Beancount entries (defer to Cronos/Journal DB)
	deduplicatedBeancount := removeDuplicateBeancountEntries(filteredBeancount, journalEntries)

	log.Printf("Beancount deduplication: %d entries -> %d entries (removed %d duplicates)",
		len(filteredBeancount), len(deduplicatedBeancount), len(filteredBeancount)-len(deduplicatedBeancount))

	// Combine Journal entries (priority) + non-duplicate Beancount entries
	allEntries = append(allEntries, journalEntries...)
	allEntries = append(allEntries, deduplicatedBeancount...)

	// Sort by date
	sortLedgerEntriesByDate(allEntries)

	return allEntries, nil
}

// removeDuplicateBeancountEntries filters out Beancount entries that duplicate Journal DB entries
// Strategy: Exclude ALL CASH transactions from Beancount (Journal DB already tracks cash via invoices/bills)
// Keep: Operating expenses, equipment, credit card, and other non-cash accounts
func removeDuplicateBeancountEntries(beancountEntries, journalEntries []LedgerEntry) []LedgerEntry {
	var filtered []LedgerEntry

	for _, bc := range beancountEntries {
		isDuplicate := false

		// Selectively exclude CASH transactions that duplicate Journal DB entries
		// Keep: Credit card payments, bank fees, expense payments (like tax prep)
		// Exclude: Client payments, payroll, owner transfers
		if bc.Account == "CASH" {
			// Always keep credit card payments
			isCreditCardPayment := containsKeywords(bc.Description, []string{"payment to chase card", "chase credit", "card payment", "cc payment"})

			// Always keep bank fees
			isBankFee := containsKeywords(bc.Description, []string{"ach pmnts initial fee", "low value fee", "standard ach", "rtp/same day"})

			// Exclude client payments (tracked via invoices)
			isClientPayment := containsKeywords(bc.Description, []string{"grid - grid retainer", "twillory parabola", "haberdash"}) ||
				(containsSubstring(bc.Description, "orig co name:") && bc.Debit > 0 && !containsKeywords(bc.Description, []string{"gusto", "wcg cpas"}))

			// Exclude payroll (tracked via bills/payroll expense)
			isPayroll := containsKeywords(bc.Description, []string{"basic online payroll payment", "payroll payment", "gusto payroll"})

			// Exclude owner transfers/distributions
			isOwnerTransfer := containsKeywords(bc.Description, []string{"online transfer to sav", "transfer to savings", "owner draw"})

			if isCreditCardPayment || isBankFee {
				// Keep these - important and not duplicated in Journal DB
				log.Printf("Keeping Beancount CASH: %s $%.2f (D) $%.2f (C) on %s - %s",
					bc.Description, bc.Debit, bc.Credit, bc.Date.Format("2006-01-02"),
					func() string {
						if isCreditCardPayment {
							return "Credit Card Payment"
						}
						return "Bank Fee"
					}())
			} else if isClientPayment {
				isDuplicate = true
				log.Printf("Excluding Beancount CASH: %s - Client payment tracked via invoice",
					bc.Description)
			} else if isPayroll {
				isDuplicate = true
				log.Printf("Excluding Beancount CASH: %s - Payroll tracked in Journal DB",
					bc.Description)
			} else if isOwnerTransfer {
				isDuplicate = true
				log.Printf("Excluding Beancount CASH: %s - Owner transfer tracked in Journal DB",
					bc.Description)
			} else {
				// For everything else, check for duplicates with Journal DB
				for _, jn := range journalEntries {
					if jn.Account != "CASH" {
						continue
					}

					// Check if amounts match within 1%
					bcAmount := bc.Debit - bc.Credit // Net amount (positive = receipt, negative = payment)
					jnAmount := jn.Debit - jn.Credit

					amountMatch := amountsMatchWithinTolerance(bcAmount, jnAmount, 0.01)
					if !amountMatch {
						continue
					}

					// Check if dates are within 7 days
					daysDiff := bc.Date.Sub(jn.Date).Hours() / 24
					dateMatch := daysDiff >= -7 && daysDiff <= 7

					if dateMatch && amountMatch {
						isDuplicate = true
						log.Printf("Duplicate CASH detected: Beancount %s $%.2f on %s matches Journal DB %s",
							bc.Description, bcAmount, bc.Date.Format("2006-01-02"), jn.Description)
						break
					}
				}

				// If not duplicate, keep it (e.g., tax prep payment, misc expenses)
				if !isDuplicate {
					log.Printf("Keeping Beancount CASH: %s $%.2f (D) $%.2f (C) on %s - Unique transaction",
						bc.Description, bc.Debit, bc.Credit, bc.Date.Format("2006-01-02"))
				}
			}
		}

		// Also exclude REVENUE from Beancount (tracked via invoices in Journal DB)
		if bc.Account == "REVENUE" || bc.Account == "ADJUSTMENT_REVENUE" {
			isDuplicate = true
			log.Printf("Excluding Beancount REVENUE entry: %s $%.2f on %s - Revenue tracked via invoices in Journal DB",
				bc.Description, bc.Credit, bc.Date.Format("2006-01-02"))
		}

		// Check for duplicate PAYROLL_EXPENSE (staff payments)
		// Only exclude if there's a matching entry in Journal DB
		if bc.Account == "PAYROLL_EXPENSE" {
			for _, jn := range journalEntries {
				// Journal DB records payroll as PAYROLL_EXPENSE or CASH
				if jn.Account != "PAYROLL_EXPENSE" && jn.Account != "CASH" {
					continue
				}

				// Payroll in Beancount could be debits (expense) or credits (cash out)
				bcAmount := bc.Debit
				if bc.Credit > 0 {
					bcAmount = bc.Credit
				}

				jnAmount := jn.Debit
				if jn.Credit > 0 {
					jnAmount = jn.Credit
				}

				if bcAmount == 0 || jnAmount == 0 {
					continue
				}

				amountMatch := amountsMatchWithinTolerance(bcAmount, jnAmount, 0.01)
				if !amountMatch {
					continue
				}

				// Check if dates are within 14 days (payroll might be delayed)
				daysDiff := bc.Date.Sub(jn.Date).Hours() / 24
				dateMatch := daysDiff >= -14 && daysDiff <= 14

				if dateMatch && amountMatch {
					isDuplicate = true
					log.Printf("Duplicate PAYROLL detected: Beancount %s $%.2f on %s matches Journal DB",
						bc.Description, bcAmount, bc.Date.Format("2006-01-02"))
					break
				}
			}
		}

		// Check for duplicate OWNER_DISTRIBUTIONS (payments to owners)
		if bc.Account == "OWNER_DISTRIBUTIONS" {
			for _, jn := range journalEntries {
				// Journal DB might record this as PAYROLL_EXPENSE or via bills
				if jn.Account != "PAYROLL_EXPENSE" && jn.Account != "CASH" {
					continue
				}

				// Owner distributions in Beancount are expenses (debits)
				// In Journal DB, they show as CASH credits (payments out)
				bcAmount := bc.Debit
				jnAmount := jn.Credit // Cash credit = payment out

				if bcAmount == 0 || jnAmount == 0 {
					continue
				}

				amountMatch := amountsMatchWithinTolerance(bcAmount, jnAmount, 0.01)
				if !amountMatch {
					continue
				}

				// Check if dates are within 14 days (distributions might be delayed)
				daysDiff := bc.Date.Sub(jn.Date).Hours() / 24
				dateMatch := daysDiff >= -14 && daysDiff <= 14

				if dateMatch && amountMatch {
					isDuplicate = true
					log.Printf("Duplicate DISTRIBUTION detected: Beancount %s $%.2f on %s matches Journal DB",
						bc.Description, bcAmount, bc.Date.Format("2006-01-02"))
					break
				}
			}
		}

		if !isDuplicate {
			filtered = append(filtered, bc)
		}
	}

	return filtered
}

// containsPayrollKeywords checks if a description contains payroll-related keywords
func containsPayrollKeywords(description string) bool {
	keywords := []string{
		"payroll", "salary", "wage", "transfer",
		"gusto", "payment to", "staff payment",
		"employee", "contractor",
	}
	return containsKeywords(description, keywords)
}

// containsKeywords checks if a description contains any of the provided keywords
func containsKeywords(description string, keywords []string) bool {
	descLower := ""
	for _, c := range description {
		if c >= 'A' && c <= 'Z' {
			descLower += string(c + 32)
		} else {
			descLower += string(c)
		}
	}

	for _, keyword := range keywords {
		if containsSubstring(descLower, keyword) {
			return true
		}
	}

	return false
}

// containsSubstring checks if s contains substr
func containsSubstring(s, substr string) bool {
	if len(substr) > len(s) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// amountsMatchWithinTolerance checks if two amounts match within a percentage tolerance
func amountsMatchWithinTolerance(amount1, amount2, tolerance float64) bool {
	if amount2 == 0 {
		return amount1 == 0
	}
	diff := (amount1 - amount2) / amount2
	if diff < 0 {
		diff = -diff
	}
	return diff <= tolerance
}

// sortLedgerEntriesByDate sorts ledger entries by date (oldest first)
func sortLedgerEntriesByDate(entries []LedgerEntry) {
	// Simple bubble sort - can be optimized if needed
	n := len(entries)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if entries[j].Date.After(entries[j+1].Date) {
				entries[j], entries[j+1] = entries[j+1], entries[j]
			}
		}
	}
}

// ReconciliationReport represents differences between Beancount and Journal DB
type ReconciliationReport struct {
	CashBalanceBeancount float64              `json:"cash_balance_beancount"`
	CashBalanceJournalDB float64              `json:"cash_balance_journal_db"`
	Difference           float64              `json:"difference"`
	PotentialDuplicates  []PotentialDuplicate `json:"potential_duplicates,omitempty"`
	AsOfDate             time.Time            `json:"as_of_date"`
}

// PotentialDuplicate represents a transaction that might exist in both systems
type PotentialDuplicate struct {
	BeancountEntry LedgerEntry `json:"beancount_entry"`
	JournalEntry   LedgerEntry `json:"journal_entry"`
	Confidence     string      `json:"confidence"` // "high", "medium", "low"
}

// GenerateReconciliationReport compares Beancount and Journal DB for discrepancies
func (a *App) GenerateReconciliationReport(beancountPath string, asOfDate time.Time) (*ReconciliationReport, error) {
	// Parse Beancount ledger
	beancountLedger, err := ParseBeancountFile(beancountPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse beancount: %w", err)
	}

	// Calculate Beancount cash balance
	beancountCash := calculateBeancountCashBalance(beancountLedger, asOfDate)

	// Calculate Journal DB cash balance
	var journalCash int64
	err = a.DB.Raw(`
		SELECT SUM(debit) - SUM(credit)
		FROM journals
		WHERE account = ? AND created_at <= ?
	`, AccountCash.String(), asOfDate).Scan(&journalCash).Error
	if err != nil {
		return nil, fmt.Errorf("failed to calculate journal cash: %w", err)
	}
	journalCashDollars := float64(journalCash) / 100.0

	// Find potential duplicates
	duplicates := findPotentialDuplicates(beancountLedger, a, asOfDate)

	report := &ReconciliationReport{
		CashBalanceBeancount: beancountCash,
		CashBalanceJournalDB: journalCashDollars,
		Difference:           beancountCash - journalCashDollars,
		PotentialDuplicates:  duplicates,
		AsOfDate:             asOfDate,
	}

	return report, nil
}

// calculateBeancountCashBalance calculates cash balance from Beancount up to a date
func calculateBeancountCashBalance(ledger *BeancountLedger, asOfDate time.Time) float64 {
	var balance float64

	for _, tx := range ledger.Transactions {
		if tx.Date.After(asOfDate) {
			continue
		}

		for _, posting := range tx.Postings {
			if posting.Account == "Assets:Checking:ChaseBusiness" {
				balance += posting.Amount
			}
		}
	}

	return balance
}

// findPotentialDuplicates identifies transactions that might exist in both systems
func findPotentialDuplicates(beancountLedger *BeancountLedger, app *App, asOfDate time.Time) []PotentialDuplicate {
	var duplicates []PotentialDuplicate

	// Load all Journal entries up to date
	var journals []Journal
	app.DB.Where("created_at <= ?", asOfDate).
		Preload("Invoice").
		Preload("Bill").
		Find(&journals)

	journalEntries := ConvertJournalToLedgerEntries(journals)
	beancountEntries := ConvertBeancountToLedgerEntries(beancountLedger)

	// Filter beancount entries by date
	var filteredBeancount []LedgerEntry
	for _, entry := range beancountEntries {
		if entry.Date.Before(asOfDate) || entry.Date.Equal(asOfDate) {
			filteredBeancount = append(filteredBeancount, entry)
		}
	}

	// Compare entries - look for same date, similar amounts, and cash/revenue accounts
	for _, bc := range filteredBeancount {
		// Only look for cash receipts (potential client payments)
		if bc.Account != "CASH" {
			continue
		}
		if bc.Debit == 0 {
			continue
		}

		for _, jn := range journalEntries {
			// Look for matching cash debits in journal
			if jn.Account != "CASH" {
				continue
			}
			if jn.Debit == 0 {
				continue
			}

			// Check if dates are within 3 days
			daysDiff := bc.Date.Sub(jn.Date).Hours() / 24
			if daysDiff < -3 || daysDiff > 3 {
				continue
			}

			// Check if amounts match (within 1%)
			amountDiff := (bc.Debit - jn.Debit) / jn.Debit
			if amountDiff < -0.01 || amountDiff > 0.01 {
				continue
			}

			// Potential duplicate found
			confidence := "medium"
			if bc.Date.Equal(jn.Date) && bc.Debit == jn.Debit {
				confidence = "high"
			}

			duplicates = append(duplicates, PotentialDuplicate{
				BeancountEntry: bc,
				JournalEntry:   jn,
				Confidence:     confidence,
			})
		}
	}

	return duplicates
}
