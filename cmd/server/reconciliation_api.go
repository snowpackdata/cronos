package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/snowpackdata/cronos"
)

// SearchExpensesForReconciliationHandler searches for expenses that could match an offline journal transaction
// GET /api/reconciliation/expenses/search?query=google&amount=7873
func (a *App) SearchExpensesForReconciliationHandler(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	searchQuery := r.URL.Query().Get("query")
	dateStr := r.URL.Query().Get("date")
	amountStr := r.URL.Query().Get("amount")

	// Build flexible query - at least one filter must be provided
	if searchQuery == "" && amountStr == "" && dateStr == "" {
		respondWithError(w, http.StatusBadRequest, "At least one search parameter is required (query, amount, or date)")
		return
	}

	var expenses []cronos.Expense
	query := a.cronosApp.DB.
		Preload("Project").
		Preload("Submitter").
		Preload("Category").
		Preload("Tags").
		Preload("Receipt").
		Where("state = ?", cronos.ExpenseStateApproved.String()).
		Where("reconciled_offline_journal_id IS NULL") // Only unreconciled expenses

	// Filter by description if search query provided
	if searchQuery != "" {
		query = query.Where("LOWER(description) LIKE ?", "%"+searchQuery+"%")
	}

	// Filter by amount if provided (fuzzy match within ±$5)
	if amountStr != "" {
		amount, err := strconv.ParseInt(amountStr, 10, 64)
		if err == nil {
			fuzzyRange := int64(500) // ±$5 in cents
			query = query.Where("amount >= ? AND amount <= ?", amount-fuzzyRange, amount+fuzzyRange)
		}
	}

	// Filter by date range (±7 days) if date provided
	if dateStr != "" {
		transactionDate, err := time.Parse("2006-01-02", dateStr)
		if err == nil {
			startDate := transactionDate.AddDate(0, 0, -7)
			endDate := transactionDate.AddDate(0, 0, 7)
			query = query.Where("date >= ? AND date <= ?", startDate, endDate)
		}
	}

	if err := query.Order("date DESC").Limit(50).Find(&expenses).Error; err != nil {
		log.Printf("Error searching expenses for reconciliation: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Error searching expenses")
		return
	}

	log.Printf("Found %d matching expenses for reconciliation search (query=%s, amount=%s, date=%s)",
		len(expenses), searchQuery, amountStr, dateStr)

	respondWithJSON(w, http.StatusOK, expenses)
}

// ReconcileExpenseWithOfflineJournalHandler links an expense with an offline journal transaction
// POST /api/reconciliation/expenses/{id}/reconcile
// Body: { "offline_journal_id": 123 }
func (a *App) ReconcileExpenseWithOfflineJournalHandler(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userIDVal := r.Context().Value("user_id")
	userID, ok := userIDVal.(uint)
	if !ok || userID == 0 {
		respondWithError(w, http.StatusUnauthorized, "User ID not found in context")
		return
	}

	// Find the employee record
	var employee cronos.Employee
	if err := a.cronosApp.DB.Where("user_id = ?", userID).First(&employee).Error; err != nil {
		respondWithError(w, http.StatusUnauthorized, "Employee record not found")
		return
	}

	// Get expense ID from URL
	vars := mux.Vars(r)
	expenseID, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid expense ID")
		return
	}

	// Parse request body
	var reqBody struct {
		OfflineJournalID uint `json:"offline_journal_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Fetch expense
	var expense cronos.Expense
	if err := a.cronosApp.DB.First(&expense, expenseID).Error; err != nil {
		respondWithError(w, http.StatusNotFound, "Expense not found")
		return
	}

	// Verify expense is approved and not already reconciled
	if expense.State != cronos.ExpenseStateApproved.String() && expense.State != cronos.ExpenseStateInvoiced.String() {
		respondWithError(w, http.StatusBadRequest, "Expense must be approved to reconcile")
		return
	}
	if expense.ReconciledOfflineJournalID != nil {
		respondWithError(w, http.StatusBadRequest, "Expense is already reconciled")
		return
	}

	// Fetch offline journal
	var offlineJournal cronos.OfflineJournal
	if err := a.cronosApp.DB.First(&offlineJournal, reqBody.OfflineJournalID).Error; err != nil {
		respondWithError(w, http.StatusNotFound, "Offline journal entry not found")
		return
	}

	// Verify offline journal is not already reconciled
	if offlineJournal.ReconciledExpenseID != nil {
		respondWithError(w, http.StatusBadRequest, "Transaction is already reconciled")
		return
	}

	// Verify amounts match (one should be debit, one credit, and they should equal)
	expenseAmount := int64(expense.Amount)
	journalAmount := offlineJournal.Debit
	if journalAmount == 0 {
		journalAmount = offlineJournal.Credit
	}

	if expenseAmount != journalAmount {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("Amounts do not match: expense=%d, transaction=%d", expenseAmount, journalAmount))
		return
	}

	// Payment account is determined from the bank transaction (offlineJournal.Account)
	// No need to verify against expense.PaymentAccountCode since we don't collect it anymore

	// Find the matching DR side of the transaction (expense account side)
	// It should have the same date, opposite debit/credit, and matching amount
	var matchingDREntry cronos.OfflineJournal
	var hasDRSide bool

	// The offlineJournal we received is the CR side (payment account)
	// Look for the DR side (expense account) that was created during categorization
	if offlineJournal.Credit > 0 {
		// This is a credit entry, look for matching debit entry
		err := a.cronosApp.DB.Where("date = ? AND debit = ? AND credit = 0 AND status = ?",
			offlineJournal.Date, offlineJournal.Credit, "approved").
			First(&matchingDREntry).Error

		if err == nil {
			hasDRSide = true
			log.Printf("Found matching DR side: ID %d, Account: %s", matchingDREntry.ID, matchingDREntry.Account)
		} else {
			log.Printf("No matching DR side found (this is OK for reconciliation): %v", err)
		}
	}

	// Perform reconciliation
	now := time.Now()

	expense.ReconciledOfflineJournalID = &reqBody.OfflineJournalID
	expense.ReconciledAt = &now
	expense.ReconciledBy = &employee.ID
	expense.PaymentAccountCode = offlineJournal.Account // Set payment account from bank transaction

	offlineJournal.ReconciledExpenseID = &expense.ID
	offlineJournal.ReconciledAt = &now
	offlineJournal.ReconciledBy = &employee.ID
	offlineJournal.Status = "posted" // Mark as posted (clearing entries will be created below)

	// Save expense
	if err := a.cronosApp.DB.Save(&expense).Error; err != nil {
		log.Printf("Failed to save expense reconciliation: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to reconcile expense")
		return
	}

	// Save CR side (payment account)
	if err := a.cronosApp.DB.Save(&offlineJournal).Error; err != nil {
		log.Printf("Failed to save offline journal reconciliation: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to reconcile transaction")
		return
	}

	// If we found the DR side, mark it as posted too (but don't post it to GL again - expense already posted)
	if hasDRSide {
		matchingDREntry.Status = "posted"
		matchingDREntry.ReconciledExpenseID = &expense.ID
		matchingDREntry.ReconciledAt = &now
		matchingDREntry.ReconciledBy = &employee.ID

		if err := a.cronosApp.DB.Save(&matchingDREntry).Error; err != nil {
			log.Printf("Failed to mark DR side as posted: %v", err)
			// Don't fail the whole reconciliation for this
		} else {
			log.Printf("Marked matching DR side as posted: ID %d", matchingDREntry.ID)
		}
	}

	// Book the clearing entry to move from ACCRUED_EXPENSES_PAYABLE to actual payment account
	// This posts directly to the main GL (not offline journal)
	//
	// DR: ACCRUED_EXPENSES_PAYABLE (clear the contra account)
	// CR: [Actual Payment Account] (e.g., CREDIT_CARD_CHASE)
	//
	// IMPORTANT: Use the SAME subaccount that was used when the expense was originally approved
	// (stored in expense.SubaccountCode) so the entries properly zero out
	//
	// Note: The DR expense side from the offline journal is NOT posted to GL because
	// the expense was already posted when it was approved. We just mark it as "posted"
	// in the offline journal for record-keeping.
	clearingDR := cronos.Journal{
		Account:    "ACCRUED_EXPENSES_PAYABLE",
		SubAccount: expense.SubaccountCode, // Use the subaccount from original expense booking
		Memo:       fmt.Sprintf("Cleared expense payment via reconciliation: %s (tx date: %s)", expense.Description, offlineJournal.Date.Format("2006-01-02")),
		Debit:      journalAmount,
		Credit:     0,
	}
	if err := a.cronosApp.DB.Create(&clearingDR).Error; err != nil {
		log.Printf("Failed to book clearing DR: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to book clearing entry")
		return
	}

	clearingCR := cronos.Journal{
		Account:    offlineJournal.Account, // The actual payment account (e.g., CREDIT_CARD_CHASE)
		SubAccount: expense.SubaccountCode, // Use same subaccount for consistency
		Memo:       fmt.Sprintf("Cleared expense payment via reconciliation: %s (tx date: %s)", expense.Description, offlineJournal.Date.Format("2006-01-02")),
		Debit:      0,
		Credit:     journalAmount,
	}
	if err := a.cronosApp.DB.Create(&clearingCR).Error; err != nil {
		log.Printf("Failed to book clearing CR: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to book clearing entry")
		return
	}

	message := fmt.Sprintf("Reconciled expense %d with offline journal %d, booked clearing entry, and marked both sides as posted",
		expense.ID, offlineJournal.ID)
	if hasDRSide {
		message = fmt.Sprintf("%s (DR ID: %d, CR ID: %d)", message, matchingDREntry.ID, offlineJournal.ID)
	}
	log.Printf("%s by employee %d", message, employee.ID)

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message":         "Reconciliation complete - clearing entries posted to GL and offline entries marked as posted",
		"expense":         expense,
		"offline_journal": offlineJournal,
		"dr_side_found":   hasDRSide,
	})
}

// UnreconcileTransactionHandler removes reconciliation link
// POST /api/reconciliation/offline-journals/{id}/unreconcile
func (a *App) UnreconcileTransactionHandler(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userIDVal := r.Context().Value("user_id")
	userID, ok := userIDVal.(uint)
	if !ok || userID == 0 {
		respondWithError(w, http.StatusUnauthorized, "User ID not found in context")
		return
	}

	// Get offline journal ID from URL
	vars := mux.Vars(r)
	journalID, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid journal ID")
		return
	}

	// Fetch offline journal with reconciled expense
	var offlineJournal cronos.OfflineJournal
	if err := a.cronosApp.DB.Preload("ReconciledExpense").First(&offlineJournal, journalID).Error; err != nil {
		respondWithError(w, http.StatusNotFound, "Offline journal entry not found")
		return
	}

	// Verify it's reconciled
	if offlineJournal.ReconciledExpenseID == nil {
		respondWithError(w, http.StatusBadRequest, "Transaction is not reconciled")
		return
	}

	// Clear reconciliation on both sides
	expenseID := *offlineJournal.ReconciledExpenseID

	// Update expense
	if err := a.cronosApp.DB.Model(&cronos.Expense{}).
		Where("id = ?", expenseID).
		Updates(map[string]interface{}{
			"reconciled_offline_journal_id": nil,
			"reconciled_at":                 nil,
			"reconciled_by":                 nil,
		}).Error; err != nil {
		log.Printf("Failed to clear expense reconciliation: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to unreconcile")
		return
	}

	// Update offline journal
	offlineJournal.ReconciledExpenseID = nil
	offlineJournal.ReconciledAt = nil
	offlineJournal.ReconciledBy = nil
	offlineJournal.Status = "pending_review" // Reset to pending review

	if err := a.cronosApp.DB.Save(&offlineJournal).Error; err != nil {
		log.Printf("Failed to clear offline journal reconciliation: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to unreconcile")
		return
	}

	log.Printf("Unreconciled offline journal %d from expense %d by user %d", journalID, expenseID, userID)

	respondWithJSON(w, http.StatusOK, map[string]string{
		"message": "Unreconciled successfully",
	})
}
