package cronos

import (
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"time"
)

// GenerateOfflineJournalHash creates a unique hash for deduplication
func GenerateOfflineJournalHash(date time.Time, account, subAccount, description string, debit, credit int64) string {
	data := fmt.Sprintf("%s|%s|%s|%s|%d|%d",
		date.Format("2006-01-02"),
		account,
		subAccount,
		description,
		debit,
		credit,
	)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// ImportBeancountToOfflineJournals parses a Beancount file and imports entries as OfflineJournals
// DEPRECATED: This function is deprecated now that all beancount data has been migrated.
// Use ImportCSVToOfflineJournals for new transaction imports.
func (a *App) ImportBeancountToOfflineJournals(beancountContent []byte) (int, int, error) {
	// Parse Beancount file
	ledger, err := ParseBeancountFromBytes(beancountContent)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse beancount: %w", err)
	}

	// Convert to ledger entries
	entries := ConvertBeancountToLedgerEntries(ledger)

	if len(entries) == 0 {
		return 0, 0, fmt.Errorf("no entries found in beancount file")
	}

	imported := 0
	skipped := 0
	failed := 0

	for _, entry := range entries {
		// Convert dollars to cents
		debitCents := int64(entry.Debit * 100)
		creditCents := int64(entry.Credit * 100)

		// Generate hash
		hash := GenerateOfflineJournalHash(
			entry.Date,
			entry.Account,
			entry.SubAccount,
			entry.Description,
			debitCents,
			creditCents,
		)

		// Check if already exists
		var existing OfflineJournal
		err := a.DB.Where("content_hash = ?", hash).First(&existing).Error
		if err == nil {
			// Already exists, skip
			skipped++
			continue
		}

		// Create new offline journal entry
		offline := OfflineJournal{
			Date:        entry.Date,
			Account:     entry.Account,
			SubAccount:  entry.SubAccount,
			Description: entry.Description,
			Debit:       debitCents,
			Credit:      creditCents,
			ContentHash: hash,
			Source:      "beancount",
			Status:      "pending_review",
			ImportedAt:  time.Now(),
		}

		err = a.DB.Create(&offline).Error
		if err != nil {
			log.Printf("Error importing offline journal (hash: %s): %v", hash[:8], err)
			failed++
			continue
		}

		imported++
	}

	log.Printf("Import complete: %d imported, %d skipped (duplicates), %d failed", imported, skipped, failed)

	// Don't return error if some entries were successfully imported
	if imported == 0 && failed > 0 {
		return 0, skipped, fmt.Errorf("failed to import any entries (%d failures)", failed)
	}

	return imported, skipped, nil
}

// GetOfflineJournals retrieves offline journals with optional filters
func (a *App) GetOfflineJournals(startDate, endDate time.Time, status string) ([]OfflineJournal, error) {
	var journals []OfflineJournal

	query := a.DB.Where("date >= ? AND date <= ?", startDate, endDate)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	err := query.Order("date ASC").Find(&journals).Error
	if err != nil {
		return nil, err
	}

	return journals, nil
}

// UpdateOfflineJournalStatus updates the status of an offline journal
func (a *App) UpdateOfflineJournalStatus(id uint, status string, staffID uint, notes string) error {
	now := time.Now()

	updates := map[string]interface{}{
		"status":      status,
		"reviewed_at": &now,
		"reviewed_by": staffID,
	}

	if notes != "" {
		updates["notes"] = notes
	}

	return a.DB.Model(&OfflineJournal{}).Where("id = ?", id).Updates(updates).Error
}

// BulkUpdateOfflineJournalStatus updates multiple offline journals
func (a *App) BulkUpdateOfflineJournalStatus(ids []uint, status string, staffID uint) error {
	now := time.Now()

	updates := map[string]interface{}{
		"status":      status,
		"reviewed_at": &now,
		"reviewed_by": staffID,
	}

	return a.DB.Model(&OfflineJournal{}).Where("id IN ?", ids).Updates(updates).Error
}

// GetCombinedJournals returns both regular journals and approved offline journals
func (a *App) GetCombinedJournals(startDate, endDate time.Time) ([]Journal, error) {
	var combinedJournals []Journal

	// Get regular journals
	var journals []Journal
	err := a.DB.Where("created_at >= ? AND created_at <= ?", startDate, endDate).
		Order("created_at ASC").
		Find(&journals).Error
	if err != nil {
		return nil, fmt.Errorf("failed to load journals: %w", err)
	}

	// Get approved offline journals
	var offlineJournals []OfflineJournal
	err = a.DB.Where("date >= ? AND date <= ? AND status = ?", startDate, endDate, "approved").
		Order("date ASC").
		Find(&offlineJournals).Error
	if err != nil {
		return nil, fmt.Errorf("failed to load offline journals: %w", err)
	}

	// Convert offline journals to Journal format
	for _, offline := range offlineJournals {
		journal := Journal{
			Account:    offline.Account,
			SubAccount: offline.SubAccount,
			Debit:      offline.Debit,
			Credit:     offline.Credit,
			Memo:       offline.Description,
		}
		journal.CreatedAt = offline.Date
		combinedJournals = append(combinedJournals, journal)
	}

	// Add regular journals
	combinedJournals = append(combinedJournals, journals...)

	return combinedJournals, nil
}

// CSVTransaction represents a row from a CSV import (bank or credit card statement)
type CSVTransaction struct {
	Date        time.Time
	Description string
	Amount      float64   // Positive for credits/deposits, negative for debits/expenses
	Balance     *float64  // Optional running balance
	Reference   string    // Check number, transaction ID, etc.
}

// convertDateFormat converts user-friendly date format strings to Go's format strings
func convertDateFormat(userFormat string) string {
	// Replace common user-friendly patterns with Go's reference time format
	format := userFormat
	format = strings.ReplaceAll(format, "YYYY", "2006")
	format = strings.ReplaceAll(format, "yyyy", "2006")
	format = strings.ReplaceAll(format, "YY", "06")
	format = strings.ReplaceAll(format, "yy", "06")
	format = strings.ReplaceAll(format, "MM", "01")
	format = strings.ReplaceAll(format, "M", "1")
	format = strings.ReplaceAll(format, "DD", "02")
	format = strings.ReplaceAll(format, "dd", "02")
	format = strings.ReplaceAll(format, "D", "2")
	format = strings.ReplaceAll(format, "d", "2")
	return format
}

// ParseCSVTransactions parses a CSV file with flexible column mapping
// Supports common CSV formats from banks and credit cards
func ParseCSVTransactions(csvContent []byte, dateCol, descCol, amountCol int, hasHeader bool, dateFormat string) ([]CSVTransaction, error) {
	reader := csv.NewReader(strings.NewReader(string(csvContent)))
	
	var transactions []CSVTransaction
	lineNum := 0

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error reading CSV at line %d: %w", lineNum, err)
		}

		lineNum++

		// Skip header row
		if hasHeader && lineNum == 1 {
			continue
		}

		// Validate we have enough columns
		maxCol := dateCol
		if descCol > maxCol {
			maxCol = descCol
		}
		if amountCol > maxCol {
			maxCol = amountCol
		}

		if len(record) <= maxCol {
			log.Printf("Skipping line %d: insufficient columns (got %d, need %d)", lineNum, len(record), maxCol+1)
			continue
		}

	// Parse date (try common formats if dateFormat not specified)
	var txDate time.Time
	if dateFormat != "" {
		// Convert user-friendly format strings to Go format strings
		goFormat := convertDateFormat(dateFormat)
		txDate, err = time.Parse(goFormat, strings.TrimSpace(record[dateCol]))
	} else {
			// Try common formats
			dateStr := strings.TrimSpace(record[dateCol])
			formats := []string{
				"2006-01-02",
				"01/02/2006",
				"1/2/2006",
				"2006/01/02",
				"Jan 2, 2006",
				"January 2, 2006",
			}
			for _, format := range formats {
				txDate, err = time.Parse(format, dateStr)
				if err == nil {
					break
				}
			}
		}

		if err != nil {
			log.Printf("Skipping line %d: invalid date format '%s': %v", lineNum, record[dateCol], err)
			continue
		}

		// Parse amount (handle various formats: $1,234.56, -123.45, (123.45) for negatives, etc.)
		amountStr := strings.TrimSpace(record[amountCol])
		// Remove currency symbols and commas
		amountStr = strings.ReplaceAll(amountStr, "$", "")
		amountStr = strings.ReplaceAll(amountStr, ",", "")
		amountStr = strings.ReplaceAll(amountStr, " ", "")

		// Handle parentheses as negative
		isNegative := false
		if strings.HasPrefix(amountStr, "(") && strings.HasSuffix(amountStr, ")") {
			isNegative = true
			amountStr = strings.Trim(amountStr, "()")
		}

		amount, err := strconv.ParseFloat(amountStr, 64)
		if err != nil {
			log.Printf("Skipping line %d: invalid amount '%s': %v", lineNum, record[amountCol], err)
			continue
		}

		if isNegative {
			amount = -amount
		}

		tx := CSVTransaction{
			Date:        txDate,
			Description: strings.TrimSpace(record[descCol]),
			Amount:      amount,
		}

		transactions = append(transactions, tx)
	}

	log.Printf("Parsed %d transactions from CSV (%d lines total)", len(transactions), lineNum)
	return transactions, nil
}

// ImportCSVToOfflineJournals imports CSV transactions as offline journals for review
// This creates UNCLASSIFIED entries that need to be categorized with from/to accounts
func (a *App) ImportCSVToOfflineJournals(csvContent []byte, dateCol, descCol, amountCol int, hasHeader bool, dateFormat string) (int, int, error) {
	// Parse CSV
	transactions, err := ParseCSVTransactions(csvContent, dateCol, descCol, amountCol, hasHeader, dateFormat)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse CSV: %w", err)
	}

	if len(transactions) == 0 {
		return 0, 0, fmt.Errorf("no valid transactions found in CSV")
	}

	imported := 0
	skipped := 0

	for _, tx := range transactions {
		// Each CSV transaction creates TWO unclassified journal entries
		// Sign doesn't matter - we always create one debit and one credit
		// User will categorize by assigning which accounts they represent
		
		// Always use absolute value - sign is irrelevant for double-entry
		amountCents := int64(tx.Amount * 100)
		if amountCents < 0 {
			amountCents = -amountCents
		}

		// Normalize date to midnight for consistent comparison
		normalizedDate := time.Date(tx.Date.Year(), tx.Date.Month(), tx.Date.Day(), 0, 0, 0, 0, time.UTC)

		// Create debit entry (needs account assignment)
		debitEntry := OfflineJournal{
			Date:        normalizedDate,
			Account:     "UNCLASSIFIED",
			SubAccount:  "DEBIT - Assign Account",
			Description: tx.Description,
			Debit:       amountCents,
			Credit:      0,
			Source:      "csv_import",
			Status:      "pending_review",
			ImportedAt:  time.Now(),
		}

		// Create credit entry (needs account assignment)
		creditEntry := OfflineJournal{
			Date:        normalizedDate,
			Account:     "UNCLASSIFIED",
			SubAccount:  "CREDIT - Assign Account",
			Description: tx.Description,
			Debit:       0,
			Credit:      amountCents,
			Source:      "csv_import",
			Status:      "pending_review",
			ImportedAt:  time.Now(),
		}

		// Generate hashes for both entries
		debitEntry.ContentHash = GenerateOfflineJournalHash(
			debitEntry.Date,
			debitEntry.Account,
			debitEntry.SubAccount,
			debitEntry.Description,
			debitEntry.Debit,
			debitEntry.Credit,
		)

		creditEntry.ContentHash = GenerateOfflineJournalHash(
			creditEntry.Date,
			creditEntry.Account,
			creditEntry.SubAccount,
			creditEntry.Description,
			creditEntry.Debit,
			creditEntry.Credit,
		)

		// Check for duplicates
		var existingDebit, existingCredit OfflineJournal
		debitExists := a.DB.Where("content_hash = ?", debitEntry.ContentHash).First(&existingDebit).Error == nil
		creditExists := a.DB.Where("content_hash = ?", creditEntry.ContentHash).First(&existingCredit).Error == nil

		if debitExists && creditExists {
			skipped++
			continue
		}

		// Create both entries (as a pair)
		if !debitExists {
			if err := a.DB.Create(&debitEntry).Error; err != nil {
				log.Printf("Error importing debit entry (hash: %s): %v", debitEntry.ContentHash[:8], err)
				continue
			}
		}

		if !creditExists {
			if err := a.DB.Create(&creditEntry).Error; err != nil {
				log.Printf("Error importing credit entry (hash: %s): %v", creditEntry.ContentHash[:8], err)
				// Try to delete the debit entry if credit fails (keep them paired)
				if !debitExists {
					a.DB.Delete(&debitEntry)
				}
				continue
			}
		}

		imported++
	}

	log.Printf("CSV import complete: %d transactions imported (%d journal entries), %d skipped (duplicates)", 
		imported, imported*2, skipped)
	return imported, skipped, nil
}

// UpdateOfflineJournalAccounts updates the account and subaccount for an offline journal entry
// This is used during the review process to categorize transactions
func (a *App) UpdateOfflineJournalAccounts(id uint, account, subAccount string) error {
	var offline OfflineJournal
	if err := a.DB.First(&offline, id).Error; err != nil {
		return fmt.Errorf("offline journal not found: %w", err)
	}

	// Update account and subaccount
	offline.Account = account
	offline.SubAccount = subAccount

	// Regenerate hash with new account info
	offline.ContentHash = GenerateOfflineJournalHash(
		offline.Date,
		offline.Account,
		offline.SubAccount,
		offline.Description,
		offline.Debit,
		offline.Credit,
	)

	if err := a.DB.Save(&offline).Error; err != nil {
		return fmt.Errorf("failed to update offline journal: %w", err)
	}

	log.Printf("Updated offline journal %d: account=%s, subaccount=%s", id, account, subAccount)
	return nil
}

// BulkUpdateOfflineJournalAccounts updates accounts for multiple transactions
func (a *App) BulkUpdateOfflineJournalAccounts(ids []uint, account, subAccount string) error {
	for _, id := range ids {
		if err := a.UpdateOfflineJournalAccounts(id, account, subAccount); err != nil {
			log.Printf("Error updating offline journal %d: %v", id, err)
			return err
		}
	}
	return nil
}

// GetOfflineJournalTransactions retrieves offline journals grouped by transaction
// Returns a map of description+date -> list of journal entries (typically 2 per transaction)
func (a *App) GetOfflineJournalTransactions(startDate, endDate time.Time, status string) (map[string][]OfflineJournal, error) {
	journals, err := a.GetOfflineJournals(startDate, endDate, status)
	if err != nil {
		return nil, err
	}

	// Group by date + description (transactions that belong together)
	transactions := make(map[string][]OfflineJournal)
	for _, journal := range journals {
		key := fmt.Sprintf("%s|%s", journal.Date.Format("2006-01-02"), journal.Description)
		transactions[key] = append(transactions[key], journal)
	}

	return transactions, nil
}

// CategorizeCSVTransaction categorizes a CSV transaction by specifying from and to accounts
// This updates both sides of the double-entry (debit and credit)
func (a *App) CategorizeCSVTransaction(date time.Time, description string, 
	fromAccount, fromSubAccount, toAccount, toSubAccount string) error {
	
	// Find all offline journals for this transaction (should be 2: debit and credit)
	// Use a 48-hour range to handle timezone differences between stored data and search query
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC).Add(-24 * time.Hour)
	endOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC).Add(48 * time.Hour)
	
	log.Printf("Searching for transactions: date range [%v to %v), description=%q, account=UNCLASSIFIED, status=pending_review", 
		startOfDay, endOfDay, description)
	
	var journals []OfflineJournal
	err := a.DB.Where("date >= ? AND date < ? AND description = ? AND account = ? AND status = ?",
		startOfDay, endOfDay, description, "UNCLASSIFIED", "pending_review").
		Order("debit DESC").  // Debit entry first
		Find(&journals).Error
	if err != nil {
		return fmt.Errorf("failed to find transactions: %w", err)
	}

	log.Printf("Found %d journal entries matching all filters", len(journals))
	for i, j := range journals {
		log.Printf("  Entry %d: ID=%d, Date=%v, Debit=%d, Credit=%d, Status=%s", i+1, j.ID, j.Date, j.Debit, j.Credit, j.Status)
	}

	if len(journals) != 2 {
		return fmt.Errorf("expected 2 unclassified entries, found %d for transaction: %s", 
			len(journals), description)
	}

	// First entry (debit) = FROM account
	debitEntry := journals[0]
	if debitEntry.Debit == 0 {
		return fmt.Errorf("first entry should be debit but has no debit amount")
	}

	debitEntry.Account = fromAccount
	debitEntry.SubAccount = fromSubAccount
	debitEntry.ContentHash = GenerateOfflineJournalHash(
		debitEntry.Date,
		debitEntry.Account,
		debitEntry.SubAccount,
		debitEntry.Description,
		debitEntry.Debit,
		debitEntry.Credit,
	)

	if err := a.DB.Save(&debitEntry).Error; err != nil {
		return fmt.Errorf("failed to update FROM account: %w", err)
	}

	// Second entry (credit) = TO account
	creditEntry := journals[1]
	if creditEntry.Credit == 0 {
		return fmt.Errorf("second entry should be credit but has no credit amount")
	}

	creditEntry.Account = toAccount
	creditEntry.SubAccount = toSubAccount
	creditEntry.ContentHash = GenerateOfflineJournalHash(
		creditEntry.Date,
		creditEntry.Account,
		creditEntry.SubAccount,
		creditEntry.Description,
		creditEntry.Debit,
		creditEntry.Credit,
	)

	if err := a.DB.Save(&creditEntry).Error; err != nil {
		return fmt.Errorf("failed to update TO account: %w", err)
	}

	log.Printf("Categorized transaction '%s': FROM %s/%s TO %s/%s", 
		description, fromAccount, fromSubAccount, toAccount, toSubAccount)
	return nil
}

// ApproveAndBookOfflineJournals approves offline journals and books them to the main GL
// This approves matching pairs of debit/credit entries together
func (a *App) ApproveAndBookOfflineJournals(ids []uint, staffID uint) (int, error) {
	booked := 0

	for _, id := range ids {
		var offline OfflineJournal
		if err := a.DB.First(&offline, id).Error; err != nil {
			log.Printf("Error loading offline journal %d: %v", id, err)
			continue
		}

		// Mark as approved
		now := time.Now()
		offline.Status = "approved"
		offline.ReviewedAt = &now
		offline.ReviewedBy = &staffID

		if err := a.DB.Save(&offline).Error; err != nil {
			log.Printf("Error approving offline journal %d: %v", id, err)
			continue
		}

		// Create corresponding Journal entry
		journal := Journal{
			Account:    offline.Account,
			SubAccount: offline.SubAccount,
			Memo:       offline.Description,
			Debit:      offline.Debit,
			Credit:     offline.Credit,
		}
		journal.CreatedAt = offline.Date // Backdate to transaction date

		if err := a.DB.Create(&journal).Error; err != nil {
			log.Printf("Error booking offline journal %d to GL: %v", id, err)
			// Revert approval
			offline.Status = "pending_review"
			offline.ReviewedAt = nil
			offline.ReviewedBy = nil
			a.DB.Save(&offline)
			continue
		}

		booked++
		log.Printf("Booked offline journal %d to GL as journal entry %d", id, journal.ID)
	}

	return booked, nil
}

// ApproveTransactionPair approves both sides of a transaction (debit and credit)
// Finds all offline journals with matching date+description and approves them together
func (a *App) ApproveTransactionPair(date time.Time, description string, staffID uint) (int, error) {
	// Find all journal entries for this transaction
	var journals []OfflineJournal
	err := a.DB.Where("date = ? AND description = ? AND status = ?",
		date, description, "pending_review").Find(&journals).Error
	if err != nil {
		return 0, fmt.Errorf("failed to find transaction: %w", err)
	}

	if len(journals) == 0 {
		return 0, fmt.Errorf("no pending transactions found")
	}

	// Check that none are still unclassified
	for _, j := range journals {
		if j.Account == "UNCLASSIFIED" {
			return 0, fmt.Errorf("transaction has unclassified entries, please categorize first")
		}
	}

	// Collect IDs
	var ids []uint
	for _, j := range journals {
		ids = append(ids, j.ID)
	}

	// Approve all entries
	booked, err := a.ApproveAndBookOfflineJournals(ids, staffID)
	if err != nil {
		return 0, err
	}

	log.Printf("Approved and booked complete transaction: %s (%d entries)", description, booked)
	return booked, nil
}

// EditOfflineJournal updates the account details of an offline journal entry
func (a *App) EditOfflineJournal(id uint, account, subAccount string, debit, credit float64) error {
	var journal OfflineJournal
	if err := a.DB.First(&journal, id).Error; err != nil {
		return fmt.Errorf("failed to find offline journal: %w", err)
	}

	// Update fields
	journal.Account = account
	journal.SubAccount = subAccount
	journal.Debit = int64(debit * 100)   // Convert dollars to cents
	journal.Credit = int64(credit * 100) // Convert dollars to cents

	// Regenerate content hash with new values
	journal.ContentHash = GenerateOfflineJournalHash(
		journal.Date,
		journal.Account,
		journal.SubAccount,
		journal.Description,
		journal.Debit,
		journal.Credit,
	)

	if err := a.DB.Save(&journal).Error; err != nil {
		return fmt.Errorf("failed to update offline journal: %w", err)
	}

	log.Printf("Updated offline journal %d: account=%s, subaccount=%s", id, account, subAccount)
	return nil
}

// DeleteOfflineJournal deletes an offline journal entry
func (a *App) DeleteOfflineJournal(id uint) error {
	result := a.DB.Delete(&OfflineJournal{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete offline journal: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("offline journal not found")
	}

	log.Printf("Deleted offline journal %d", id)
	return nil
}

// PostOfflineJournalsToGL posts approved offline journals to the main General Ledger
// This moves entries from the staging area (offline_journals) to the main journal table
func (a *App) PostOfflineJournalsToGL(ids []uint) error {
	var journals []OfflineJournal
	err := a.DB.Where("id IN ?", ids).Find(&journals).Error
	if err != nil {
		return fmt.Errorf("failed to find offline journals: %w", err)
	}

	// Validate all are approved and categorized
	for _, j := range journals {
		if j.Status != "approved" {
			return fmt.Errorf("offline journal %d is not approved (status: %s)", j.ID, j.Status)
		}
		if j.Account == "UNCLASSIFIED" {
			return fmt.Errorf("offline journal %d is not categorized", j.ID)
		}
	}

	// Create Journal entries for each
	for _, offline := range journals {
		journal := Journal{
			Account:    offline.Account,
			SubAccount: offline.SubAccount,
			Memo:       offline.Description,
			Debit:      offline.Debit,
			Credit:     offline.Credit,
		}
		// Backdate to original transaction date
		journal.CreatedAt = offline.Date

		if err := a.DB.Create(&journal).Error; err != nil {
			return fmt.Errorf("failed to create journal entry from offline journal %d: %w", offline.ID, err)
		}

		log.Printf("Posted offline journal %d to GL as journal entry %d", offline.ID, journal.ID)
	}

	// Mark offline journals as posted
	err = a.DB.Model(&OfflineJournal{}).Where("id IN ?", ids).Update("status", "posted").Error
	if err != nil {
		return fmt.Errorf("failed to update offline journal status: %w", err)
	}

	log.Printf("Posted %d offline journals to General Ledger", len(ids))
	return nil
}

// ReverseJournalEntry creates a reversing entry for a given journal entry
// This is used to correct mistakes in the main General Ledger
func (a *App) ReverseJournalEntry(journalID uint, reason string, correctedEntry *Journal) error {
	var original Journal
	if err := a.DB.First(&original, journalID).Error; err != nil {
		return fmt.Errorf("failed to find journal entry: %w", err)
	}

	// Create reversing entry (swap debit and credit)
	reversing := Journal{
		Account:    original.Account,
		SubAccount: original.SubAccount,
		Memo:       fmt.Sprintf("REVERSAL: %s (Original: %s)", reason, original.Memo),
		Debit:      original.Credit, // Swap
		Credit:     original.Debit,  // Swap
	}

	if err := a.DB.Create(&reversing).Error; err != nil {
		return fmt.Errorf("failed to create reversing entry: %w", err)
	}

	log.Printf("Created reversing entry %d for journal %d", reversing.ID, journalID)

	// If a corrected entry is provided, create it
	if correctedEntry != nil {
		correctedEntry.Memo = fmt.Sprintf("CORRECTION: %s (Original entry: #%d)", correctedEntry.Memo, journalID)
		if err := a.DB.Create(correctedEntry).Error; err != nil {
			return fmt.Errorf("failed to create corrected entry: %w", err)
		}
		log.Printf("Created corrected entry %d", correctedEntry.ID)
	}

	return nil
}

