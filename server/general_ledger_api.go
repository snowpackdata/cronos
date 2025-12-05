package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/snowpackdata/cronos"
)

// ReverseJournalEntryHandler creates a reversing entry for a journal entry
func (a *App) ReverseJournalEntryHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Reason          string `json:"reason"`
		CreateCorrected bool   `json:"create_corrected"`
		Corrected       *struct {
			Account    string  `json:"account"`
			SubAccount string  `json:"sub_account"`
			Memo       string  `json:"memo"`
			Debit      float64 `json:"debit"`  // In dollars
			Credit     float64 `json:"credit"` // In dollars
		} `json:"corrected,omitempty"`
	}

	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		a.logger.Printf("Failed to decode request body: %v", err)
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	a.logger.Printf("Reversing journal entry %d: reason=%s, create_corrected=%v", id, req.Reason, req.CreateCorrected)

	if req.Reason == "" {
		http.Error(w, "Reason is required", http.StatusBadRequest)
		return
	}

	var correctedEntry *cronos.Journal
	if req.CreateCorrected && req.Corrected != nil {
		a.logger.Printf("Creating corrected entry: account=%s, subaccount=%s, debit=%.2f, credit=%.2f", 
			req.Corrected.Account, req.Corrected.SubAccount, req.Corrected.Debit, req.Corrected.Credit)
		correctedEntry = &cronos.Journal{
			Account:    req.Corrected.Account,
			SubAccount: req.Corrected.SubAccount,
			Memo:       req.Corrected.Memo,
			Debit:      int64(req.Corrected.Debit * 100),  // Convert dollars to cents
			Credit:     int64(req.Corrected.Credit * 100), // Convert dollars to cents
		}
	}

	err = a.cronosApp.ReverseJournalEntry(uint(id), req.Reason, correctedEntry)
	if err != nil {
		a.logger.Printf("Failed to reverse journal entry %d: %v", id, err)
		http.Error(w, "Failed to reverse journal entry: "+err.Error(), http.StatusInternalServerError)
		return
	}

	a.logger.Printf("Successfully reversed journal entry %d", id)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Journal entry reversed successfully",
	})
}

