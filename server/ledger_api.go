package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/snowpackdata/cronos"
)

// CombinedGeneralLedgerHandler returns combined ledger entries from Beancount + Journal DB
func (a *App) CombinedGeneralLedgerHandler(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters for date range
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")

	// Default to current year if not specified
	startDate := time.Date(time.Now().Year(), 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Now()

	if startDateStr != "" {
		parsed, err := time.Parse("2006-01-02", startDateStr)
		if err == nil {
			startDate = parsed
		}
	}

	if endDateStr != "" {
		parsed, err := time.Parse("2006-01-02", endDateStr)
		if err == nil {
			endDate = parsed
		}
	}

	// Get Beancount file path from environment variable or default
	beancountPath := os.Getenv("BEANCOUNT_FILE_PATH")
	if beancountPath == "" {
		beancountPath = "/Users/naterobinson/Projects/snowpack/finances/finances.beancount"
	}

	// Get combined ledger
	entries, err := a.cronosApp.GetCombinedGeneralLedger(beancountPath, startDate, endDate)
	if err != nil {
		log.Printf("Error: CombinedGeneralLedger - Failed to get ledger: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve general ledger")
		return
	}

	respondWithJSON(w, http.StatusOK, entries)
}

// ReconciliationReportHandler returns a reconciliation report comparing Beancount and Journal DB
func (a *App) ReconciliationReportHandler(w http.ResponseWriter, r *http.Request) {
	// Parse query parameter for as-of date
	asOfDateStr := r.URL.Query().Get("as_of_date")

	asOfDate := time.Now()
	if asOfDateStr != "" {
		parsed, err := time.Parse("2006-01-02", asOfDateStr)
		if err == nil {
			asOfDate = parsed
		}
	}

	// Get Beancount file path
	beancountPath := os.Getenv("BEANCOUNT_FILE_PATH")
	if beancountPath == "" {
		beancountPath = "/Users/naterobinson/Projects/snowpack/finances/finances.beancount"
	}

	// Generate reconciliation report
	report, err := a.cronosApp.GenerateReconciliationReport(beancountPath, asOfDate)
	if err != nil {
		log.Printf("Error: ReconciliationReport - Failed to generate report: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to generate reconciliation report")
		return
	}

	respondWithJSON(w, http.StatusOK, report)
}

// AccountSummaryHandler returns account balances grouped by account type
func (a *App) AccountSummaryHandler(w http.ResponseWriter, r *http.Request) {
	// Parse query parameter for as-of date
	asOfDateStr := r.URL.Query().Get("as_of_date")

	asOfDate := time.Now()
	if asOfDateStr != "" {
		parsed, err := time.Parse("2006-01-02", asOfDateStr)
		if err == nil {
			asOfDate = parsed
		}
	}

	// Get Beancount file path
	beancountPath := os.Getenv("BEANCOUNT_FILE_PATH")
	if beancountPath == "" {
		beancountPath = "/Users/naterobinson/Projects/snowpack/finances/finances.beancount"
	}

	// Get all entries up to date
	entries, err := a.cronosApp.GetCombinedGeneralLedger(beancountPath, time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), asOfDate)
	if err != nil {
		log.Printf("Error: AccountSummary - Failed to get ledger: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve account summary")
		return
	}

	// Calculate balances by account
	balances := calculateAccountBalances(entries)

	respondWithJSON(w, http.StatusOK, balances)
}

// AccountBalance represents the balance for a specific account
type AccountBalance struct {
	Account    string  `json:"account"`
	SubAccount string  `json:"sub_account,omitempty"`
	Balance    float64 `json:"balance"`
	Debits     float64 `json:"debits"`
	Credits    float64 `json:"credits"`
}

// calculateAccountBalances groups entries by account and calculates balances
func calculateAccountBalances(entries []cronos.LedgerEntry) map[string]AccountBalance {
	balances := make(map[string]AccountBalance)

	for _, entry := range entries {
		key := entry.Account
		if entry.SubAccount != "" {
			key = entry.Account + ":" + entry.SubAccount
		}

		balance, exists := balances[key]
		if !exists {
			balance = AccountBalance{
				Account:    entry.Account,
				SubAccount: entry.SubAccount,
			}
		}

		balance.Debits += entry.Debit
		balance.Credits += entry.Credit
		balance.Balance = balance.Debits - balance.Credits

		balances[key] = balance
	}

	return balances
}

// TrialBalanceHandler returns a trial balance report (sum of debits = sum of credits)
func (a *App) TrialBalanceHandler(w http.ResponseWriter, r *http.Request) {
	// Parse query parameter for as-of date
	asOfDateStr := r.URL.Query().Get("as_of_date")

	asOfDate := time.Now()
	if asOfDateStr != "" {
		parsed, err := time.Parse("2006-01-02", asOfDateStr)
		if err == nil {
			asOfDate = parsed
		}
	}

	// Get Beancount file path
	beancountPath := os.Getenv("BEANCOUNT_FILE_PATH")
	if beancountPath == "" {
		beancountPath = "/Users/naterobinson/Projects/snowpack/finances/finances.beancount"
	}

	// Get all entries
	entries, err := a.cronosApp.GetCombinedGeneralLedger(beancountPath, time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), asOfDate)
	if err != nil {
		log.Printf("Error: TrialBalance - Failed to get ledger: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve trial balance")
		return
	}

	// Calculate totals
	var totalDebits, totalCredits float64
	accountBalances := calculateAccountBalances(entries)

	for _, balance := range accountBalances {
		totalDebits += balance.Debits
		totalCredits += balance.Credits
	}

	// Convert to slice for JSON
	var balancesList []AccountBalance
	for _, balance := range accountBalances {
		balancesList = append(balancesList, balance)
	}

	response := map[string]interface{}{
		"as_of_date":       asOfDate,
		"total_debits":     totalDebits,
		"total_credits":    totalCredits,
		"balance":          totalDebits - totalCredits,
		"is_balanced":      (totalDebits-totalCredits) < 0.01 && (totalDebits-totalCredits) > -0.01,
		"account_balances": balancesList,
	}

	respondWithJSON(w, http.StatusOK, response)
}
