package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

// OfflineJournalsListHandler lists offline journals with optional filters
func (a *App) OfflineJournalsListHandler(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")
	status := r.URL.Query().Get("status") // pending_review, approved, duplicate, excluded

	// Default to current month if not provided
	startDate := time.Now().AddDate(0, 0, -30)
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

	// Get offline journals
	journals, err := a.cronosApp.GetOfflineJournals(startDate, endDate, status)
	if err != nil {
		http.Error(w, "Failed to get offline journals: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(journals)
}

// UpdateOfflineJournalStatusHandler updates the status of a single offline journal
func (a *App) UpdateOfflineJournalStatusHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Status string `json:"status"` // approved, duplicate, excluded
		Notes  string `json:"notes"`
	}

	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate status
	validStatuses := map[string]bool{
		"approved":  true,
		"duplicate": true,
		"excluded":  true,
	}

	if !validStatuses[req.Status] {
		http.Error(w, "Invalid status", http.StatusBadRequest)
		return
	}

	// Get staff ID from context
	staffID := uint(1) // TODO: Get from JWT token

	err = a.cronosApp.UpdateOfflineJournalStatus(uint(id), req.Status, staffID, req.Notes)
	if err != nil {
		http.Error(w, "Failed to update status: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Status updated successfully"})
}

// BulkUpdateOfflineJournalStatusHandler updates multiple offline journals at once
func (a *App) BulkUpdateOfflineJournalStatusHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		IDs    []uint `json:"ids"`
		Status string `json:"status"` // approved, duplicate, excluded
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate status
	validStatuses := map[string]bool{
		"approved":  true,
		"duplicate": true,
		"excluded":  true,
	}

	if !validStatuses[req.Status] {
		http.Error(w, "Invalid status", http.StatusBadRequest)
		return
	}

	// Get staff ID from context
	staffID := uint(1) // TODO: Get from JWT token

	err = a.cronosApp.BulkUpdateOfflineJournalStatus(req.IDs, req.Status, staffID)
	if err != nil {
		http.Error(w, "Failed to update status: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Status updated successfully"})
}

// EditOfflineJournalHandler updates the details of an offline journal entry
func (a *App) EditOfflineJournalHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Account    string  `json:"account"`
		SubAccount string  `json:"sub_account"`
		Debit      float64 `json:"debit"`  // In dollars
		Credit     float64 `json:"credit"` // In dollars
	}

	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = a.cronosApp.EditOfflineJournal(uint(id), req.Account, req.SubAccount, req.Debit, req.Credit)
	if err != nil {
		http.Error(w, "Failed to update offline journal: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Offline journal updated successfully"})
}

// DeleteOfflineJournalHandler deletes an offline journal entry
func (a *App) DeleteOfflineJournalHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	err = a.cronosApp.DeleteOfflineJournal(uint(id))
	if err != nil {
		http.Error(w, "Failed to delete offline journal: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Offline journal deleted successfully"})
}

// PostOfflineJournalsToGLHandler posts approved offline journals to the main GL
func (a *App) PostOfflineJournalsToGLHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		IDs []uint `json:"ids"`
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.IDs) == 0 {
		http.Error(w, "No IDs provided", http.StatusBadRequest)
		return
	}

	err = a.cronosApp.PostOfflineJournalsToGL(req.IDs)
	if err != nil {
		http.Error(w, "Failed to post to GL: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Posted to General Ledger successfully",
		"count":   strconv.Itoa(len(req.IDs)),
	})
}
