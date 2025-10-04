package cronos

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
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

