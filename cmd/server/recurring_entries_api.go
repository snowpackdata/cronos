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

// GenerateRecurringEntriesHandler triggers generation of recurring entries for the current month
// POST /api/admin/recurring-entries/generate
func (a *App) GenerateRecurringEntriesHandler(w http.ResponseWriter, r *http.Request) {
	tenant := MustGetTenant(r.Context())
	log.Printf("Manually triggered recurring entry generation")

	// Get current month period
	now := time.Now()
	periodStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

	// Count before generation (within tenant)
	var beforeCount int64
	a.cronosApp.DB.Scopes(cronos.TenantScope(tenant.ID)).Model(&cronos.RecurringBillLineItem{}).Count(&beforeCount)

	if err := a.cronosApp.GenerateRecurringEntriesForCurrentMonth(); err != nil {
		log.Printf("Failed to generate recurring entries: %v", err)
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to generate recurring entries: %v", err))
		return
	}

	// Count after generation (within tenant)
	var afterCount int64
	a.cronosApp.DB.Scopes(cronos.TenantScope(tenant.ID)).Model(&cronos.RecurringBillLineItem{}).Count(&afterCount)

	// Count GL entries (within tenant)
	var glCount int64
	a.cronosApp.DB.Scopes(cronos.TenantScope(tenant.ID)).Model(&cronos.Journal{}).Where("recurring_bill_line_item_id IS NOT NULL").Count(&glCount)

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message":           "Recurring entries generated successfully",
		"period":            periodStart.Format("2006-01-02"),
		"line_items_before": beforeCount,
		"line_items_after":  afterCount,
		"line_items_added":  afterCount - beforeCount,
		"gl_entries":        glCount,
	})
}

// ListRecurringEntriesHandler lists all recurring entries
// GET /api/admin/recurring-entries
func (a *App) ListRecurringEntriesHandler(w http.ResponseWriter, r *http.Request) {
	tenant := MustGetTenant(r.Context())
	var entries []cronos.RecurringEntry
	if err := a.cronosApp.DB.Scopes(cronos.TenantScope(tenant.ID)).Preload("Employee.HeadshotAsset").Order("employee_id, type").Find(&entries).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to load recurring entries")
		return
	}

	respondWithJSON(w, http.StatusOK, entries)
}

// CreateRecurringEntryHandler creates a new recurring entry
// POST /api/admin/recurring-entries
func (a *App) CreateRecurringEntryHandler(w http.ResponseWriter, r *http.Request) {
	tenant := MustGetTenant(r.Context())
	var req cronos.RecurringEntry
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate required fields
	if req.EmployeeID == 0 || req.Type == "" || req.Amount == 0 {
		respondWithError(w, http.StatusBadRequest, "employee_id, type, and amount are required")
		return
	}

	req.TenantID = tenant.ID
	if err := a.cronosApp.DB.Create(&req).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create recurring entry")
		return
	}

	respondWithJSON(w, http.StatusCreated, req)
}

// UpdateRecurringEntryHandler updates an existing recurring entry
// PUT /api/admin/recurring-entries/{id}
func (a *App) UpdateRecurringEntryHandler(w http.ResponseWriter, r *http.Request) {
	tenant := MustGetTenant(r.Context())
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	var entry cronos.RecurringEntry
	if err := a.cronosApp.DB.Scopes(cronos.TenantScope(tenant.ID)).First(&entry, id).Error; err != nil {
		respondWithError(w, http.StatusNotFound, "Recurring entry not found")
		return
	}

	var updates cronos.RecurringEntry
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Update fields
	if updates.Description != "" {
		entry.Description = updates.Description
	}
	if updates.Amount > 0 {
		entry.Amount = updates.Amount
	}
	if !updates.StartDate.IsZero() {
		entry.StartDate = updates.StartDate
	}
	if updates.EndDate != nil {
		entry.EndDate = updates.EndDate
	}
	entry.IsActive = updates.IsActive

	if err := a.cronosApp.DB.Save(&entry).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to update recurring entry")
		return
	}

	respondWithJSON(w, http.StatusOK, entry)
}

// DeleteRecurringEntryHandler deletes a recurring entry
// DELETE /api/admin/recurring-entries/{id}
func (a *App) DeleteRecurringEntryHandler(w http.ResponseWriter, r *http.Request) {
	tenant := MustGetTenant(r.Context())
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	if err := a.cronosApp.DB.Scopes(cronos.TenantScope(tenant.ID)).Delete(&cronos.RecurringEntry{}, id).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to delete recurring entry")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{
		"message": "Recurring entry deleted successfully",
	})
}

// SyncEmployeeRecurringEntriesHandler creates or updates recurring entries for all salaried employees
// POST /api/admin/recurring-entries/sync
func (a *App) SyncEmployeeRecurringEntriesHandler(w http.ResponseWriter, r *http.Request) {
	tenant := MustGetTenant(r.Context())
	log.Printf("Syncing recurring entries for all salaried employees")

	// Find all active salaried employees (support both formats) - within tenant
	var employees []cronos.Employee
	if err := a.cronosApp.DB.Scopes(cronos.TenantScope(tenant.ID)).Where("is_active = ? AND (compensation_type IN ? OR compensation_type IN ?)",
		true,
		[]string{"salaried", "SALARIED", "COMPENSATION_TYPE_SALARIED"},
		[]string{"base-plus-variable", "BASE-PLUS-VARIABLE", "COMPENSATION_TYPE_BASE_PLUS_VARIABLE"}).
		Find(&employees).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to load employees")
		return
	}

	created := 0
	updated := 0
	skipped := 0

	for _, employee := range employees {
		if err := a.cronosApp.CreateRecurringEntryForEmployee(employee.ID); err != nil {
			log.Printf("Failed to sync recurring entry for employee %d: %v", employee.ID, err)
			skipped++
		} else {
			// Check if we created or updated (within tenant)
			var entry cronos.RecurringEntry
			a.cronosApp.DB.Scopes(cronos.TenantScope(tenant.ID)).Where("employee_id = ? AND type = ? AND is_active = ?",
				employee.ID, "base_salary", true).First(&entry)

			if entry.CreatedAt.After(time.Now().Add(-1 * time.Minute)) {
				created++
			} else {
				updated++
			}
		}
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Sync complete",
		"created": created,
		"updated": updated,
		"skipped": skipped,
	})
}
