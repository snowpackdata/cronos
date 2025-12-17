package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

// ListChartOfAccountsHandler lists chart of accounts with optional filters
func (a *App) ListChartOfAccountsHandler(w http.ResponseWriter, r *http.Request) {
	accountType := r.URL.Query().Get("account_type") // ASSET, LIABILITY, EQUITY, REVENUE, EXPENSE
	activeOnly := r.URL.Query().Get("active_only") == "true"

	accounts, err := a.cronosApp.GetChartOfAccounts(accountType, activeOnly)
	if err != nil {
		http.Error(w, "Failed to get accounts: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(accounts)
}

// CreateChartOfAccountHandler creates a new chart of account
func (a *App) CreateChartOfAccountHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		AccountCode string `json:"account_code"`
		AccountName string `json:"account_name"`
		AccountType string `json:"account_type"`
		Description string `json:"description"`
		ParentID    *uint  `json:"parent_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.AccountCode == "" || req.AccountName == "" || req.AccountType == "" {
		http.Error(w, "Missing required fields: account_code, account_name, account_type", http.StatusBadRequest)
		return
	}

	account, err := a.cronosApp.CreateChartOfAccount(
		req.AccountCode,
		req.AccountName,
		req.AccountType,
		req.Description,
		req.ParentID,
	)
	if err != nil {
		http.Error(w, "Failed to create account: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(account)
}

// UpdateChartOfAccountHandler updates an existing chart of account
func (a *App) UpdateChartOfAccountHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	accountCode := vars["code"]

	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := a.cronosApp.UpdateChartOfAccount(accountCode, req); err != nil {
		http.Error(w, "Failed to update account: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Account updated successfully"})
}

// DeactivateChartOfAccountHandler deactivates a chart of account
func (a *App) DeactivateChartOfAccountHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	accountCode := vars["code"]

	if err := a.cronosApp.DeactivateChartOfAccount(accountCode); err != nil {
		http.Error(w, "Failed to deactivate account: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Account deactivated successfully"})
}

// SeedSystemAccountsHandler seeds the system-defined accounts (one-time operation)
func (a *App) SeedSystemAccountsHandler(w http.ResponseWriter, r *http.Request) {
	if err := a.cronosApp.SeedSystemAccounts(); err != nil {
		http.Error(w, "Failed to seed accounts: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "System accounts seeded successfully"})
}

// ListSubaccountsHandler lists subaccounts with optional filters
func (a *App) ListSubaccountsHandler(w http.ResponseWriter, r *http.Request) {
	accountCode := r.URL.Query().Get("account_code")
	subaccountType := r.URL.Query().Get("type") // VENDOR, CLIENT, EMPLOYEE, CUSTOM
	activeOnly := r.URL.Query().Get("active_only") == "true"

	subaccounts, err := a.cronosApp.GetSubaccounts(accountCode, subaccountType, activeOnly)
	if err != nil {
		http.Error(w, "Failed to get subaccounts: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(subaccounts)
}

// CreateSubaccountHandler creates a new subaccount
func (a *App) CreateSubaccountHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Code        string `json:"code"`
		Name        string `json:"name"`
		AccountCode string `json:"account_code"`
		Type        string `json:"type"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Code == "" || req.Name == "" || req.AccountCode == "" || req.Type == "" {
		http.Error(w, "Missing required fields: code, name, account_code, type", http.StatusBadRequest)
		return
	}

	subaccount, err := a.cronosApp.CreateSubaccount(req.Code, req.Name, req.AccountCode, req.Type)
	if err != nil {
		http.Error(w, "Failed to create subaccount: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(subaccount)
}

// UpdateSubaccountHandler updates an existing subaccount
func (a *App) UpdateSubaccountHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	code := vars["code"]

	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := a.cronosApp.UpdateSubaccount(code, req); err != nil {
		http.Error(w, "Failed to update subaccount: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Subaccount updated successfully"})
}

// DeactivateSubaccountHandler deactivates a subaccount
func (a *App) DeactivateSubaccountHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	code := vars["code"]

	if err := a.cronosApp.DeactivateSubaccount(code); err != nil {
		http.Error(w, "Failed to deactivate subaccount: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Subaccount deactivated successfully"})
}

// UploadCSVHandler handles CSV file upload for transaction import
func (a *App) UploadCSVHandler(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form (max 10MB)
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Get the file from the form
	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to get file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Extract file name for tracking
	sourceFileName := fileHeader.Filename
	if sourceFileName == "" {
		sourceFileName = "unknown.csv"
	}

	// Get column mappings from form
	dateCol := 0
	descCol := 1
	amountCol := 2
	hasHeader := true
	dateFormat := ""

	if r.FormValue("date_col") != "" {
		if n, err := json.Number(r.FormValue("date_col")).Int64(); err == nil {
			dateCol = int(n)
		}
	}
	if r.FormValue("desc_col") != "" {
		if n, err := json.Number(r.FormValue("desc_col")).Int64(); err == nil {
			descCol = int(n)
		}
	}
	if r.FormValue("amount_col") != "" {
		if n, err := json.Number(r.FormValue("amount_col")).Int64(); err == nil {
			amountCol = int(n)
		}
	}
	if r.FormValue("has_header") == "false" {
		hasHeader = false
	}
	if r.FormValue("date_format") != "" {
		dateFormat = r.FormValue("date_format")
	}

	// Read file contents
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read file", http.StatusInternalServerError)
		return
	}

	// Import CSV to offline journals
	imported, skipped, err := a.cronosApp.ImportCSVToOfflineJournals(
		fileBytes,
		dateCol,
		descCol,
		amountCol,
		hasHeader,
		dateFormat,
		sourceFileName,
	)
	if err != nil {
		http.Error(w, "Failed to import: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("CSV import completed: %d imported, %d skipped", imported, skipped)

	response := map[string]interface{}{
		"imported": imported,
		"skipped":  skipped,
		"message":  "CSV import successful",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetOfflineJournalTransactionsHandler returns offline journals grouped by transaction
func (a *App) GetOfflineJournalTransactionsHandler(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")
	status := r.URL.Query().Get("status")

	// Parse dates
	startDate, endDate, err := parseDateRange(startDateStr, endDateStr)
	if err != nil {
		http.Error(w, "Invalid date format: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Get transactions grouped
	transactions, err := a.cronosApp.GetOfflineJournalTransactions(startDate, endDate, status)
	if err != nil {
		http.Error(w, "Failed to get transactions: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(transactions)
}

// CategorizeCSVTransactionHandler categorizes a transaction with FROM and TO accounts
func (a *App) CategorizeCSVTransactionHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Date               string `json:"date"`
		Description        string `json:"description"`
		FromAccount        string `json:"from_account"`
		FromSubAccount     string `json:"from_subaccount"`
		ToAccount          string `json:"to_account"`
		ToSubAccount       string `json:"to_subaccount"`
		TransactionGroupID string `json:"transaction_group_id"` // Optional: for precise matching
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Parse date
	date, err := parseDate(req.Date)
	if err != nil {
		http.Error(w, "Invalid date format", http.StatusBadRequest)
		return
	}

	// Categorize transaction (with optional transaction group ID)
	err = a.cronosApp.CategorizeCSVTransaction(
		date,
		req.Description,
		req.FromAccount,
		req.FromSubAccount,
		req.ToAccount,
		req.ToSubAccount,
		req.TransactionGroupID,
	)
	if err != nil {
		http.Error(w, "Failed to categorize transaction: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Transaction categorized successfully"})
}

// ApproveTransactionPairHandler approves both sides of a transaction
func (a *App) ApproveTransactionPairHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Date               string `json:"date"`
		Description        string `json:"description"`
		TransactionGroupID string `json:"transaction_group_id"` // Optional: for precise matching
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Parse date
	date, err := parseDate(req.Date)
	if err != nil {
		http.Error(w, "Invalid date format", http.StatusBadRequest)
		return
	}

	// Get staff ID from context (TODO: from JWT)
	staffID := uint(1)

	// Approve transaction (with optional transaction group ID)
	booked, err := a.cronosApp.ApproveTransactionPair(date, req.Description, staffID, req.TransactionGroupID)
	if err != nil {
		http.Error(w, "Failed to approve transaction: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Transaction approved and booked successfully",
		"booked":  booked,
	})
}

// GetSuggestedCategorizationsHandler returns suggested categorizations based on fuzzy matching
func (a *App) GetSuggestedCategorizationsHandler(w http.ResponseWriter, r *http.Request) {
	description := r.URL.Query().Get("description")
	if description == "" {
		http.Error(w, "Description is required", http.StatusBadRequest)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 5 // Default to 5 suggestions
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	suggestions, err := a.cronosApp.GetSuggestedCategorizations(description, limit)
	if err != nil {
		http.Error(w, "Failed to get suggestions: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(suggestions)
}

// Helper functions

// parseDateRange parses start and end date strings, with defaults
func parseDateRange(startDateStr, endDateStr string) (time.Time, time.Time, error) {
	// Default to current month if not provided
	startDate := time.Now().AddDate(0, 0, -30)
	endDate := time.Now()

	if startDateStr != "" {
		parsed, err := time.Parse("2006-01-02", startDateStr)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
		startDate = parsed
	}

	if endDateStr != "" {
		parsed, err := time.Parse("2006-01-02", endDateStr)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
		endDate = parsed
	}

	return startDate, endDate, nil
}

// parseDate parses a date string in YYYY-MM-DD format
func parseDate(dateStr string) (time.Time, error) {
	return time.Parse("2006-01-02", dateStr)
}
