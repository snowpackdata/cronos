package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/snowpackdata/cronos"
)

// GetExpenseCategoriesHandler returns all expense categories
func (a *App) GetExpenseCategoriesHandler(w http.ResponseWriter, r *http.Request) {
	activeOnly := r.URL.Query().Get("active_only") == "true"

	categories, err := a.cronosApp.GetExpenseCategories(activeOnly)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get expense categories")
		return
	}

	respondWithJSON(w, http.StatusOK, categories)
}

// CreateExpenseCategoryHandler creates a new expense category
func (a *App) CreateExpenseCategoryHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if req.Name == "" {
		respondWithError(w, http.StatusBadRequest, "Category name is required")
		return
	}

	category, err := a.cronosApp.CreateExpenseCategory(req.Name, req.Description)
	if err != nil {
		log.Printf("Failed to create expense category: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to create expense category")
		return
	}

	respondWithJSON(w, http.StatusCreated, category)
}

// UpdateExpenseCategoryHandler updates an existing expense category
func (a *App) UpdateExpenseCategoryHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid category ID")
		return
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Active      bool   `json:"active"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if req.Name == "" {
		respondWithError(w, http.StatusBadRequest, "Category name is required")
		return
	}

	if err := a.cronosApp.UpdateExpenseCategory(uint(id), req.Name, req.Description, req.Active); err != nil {
		log.Printf("Failed to update expense category: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to update expense category")
		return
	}

	// Fetch updated category to return
	category, err := a.cronosApp.GetExpenseCategory(uint(id))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch updated category")
		return
	}

	respondWithJSON(w, http.StatusOK, category)
}

// DeleteExpenseCategoryHandler deletes an expense category
func (a *App) DeleteExpenseCategoryHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid category ID")
		return
	}

	if err := a.cronosApp.DeleteExpenseCategory(uint(id)); err != nil {
		log.Printf("Failed to delete expense category: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to delete expense category")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Category deleted successfully"})
}

// GetExpenseTagsHandler returns all expense tags with spend summaries
func (a *App) GetExpenseTagsHandler(w http.ResponseWriter, r *http.Request) {
	activeOnly := r.URL.Query().Get("active_only") == "true"

	tags, err := a.cronosApp.GetExpenseTags(activeOnly)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get expense tags")
		return
	}

	// Enrich tags with spend summaries
	type TagWithSpend struct {
		cronos.ExpenseTag
		TotalSpent       int  `json:"total_spent"`
		RemainingBudget  *int `json:"remaining_budget"`
		BudgetPercentage *int `json:"budget_percentage"` // Percentage of budget used
	}

	enrichedTags := make([]TagWithSpend, len(tags))
	for i, tag := range tags {
		totalSpent, budget, remaining, err := a.cronosApp.GetTagSpendSummary(tag.ID)
		if err != nil {
			log.Printf("Failed to get spend summary for tag %d: %v", tag.ID, err)
			totalSpent = 0
		}

		var budgetPercentage *int
		if budget != nil && *budget > 0 {
			pct := int((float64(totalSpent) / float64(*budget)) * 100)
			budgetPercentage = &pct
		}

		enrichedTags[i] = TagWithSpend{
			ExpenseTag:       tag,
			TotalSpent:       totalSpent,
			RemainingBudget:  remaining,
			BudgetPercentage: budgetPercentage,
		}
	}

	respondWithJSON(w, http.StatusOK, enrichedTags)
}

// CreateExpenseTagHandler creates a new expense tag
func (a *App) CreateExpenseTagHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Active      bool   `json:"active"`
		Budget      *int   `json:"budget"` // Budget in cents
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if req.Name == "" {
		respondWithError(w, http.StatusBadRequest, "Tag name is required")
		return
	}

	tag, err := a.cronosApp.CreateExpenseTag(req.Name, req.Description, req.Active, req.Budget)
	if err != nil {
		log.Printf("Failed to create expense tag: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to create expense tag")
		return
	}

	respondWithJSON(w, http.StatusCreated, tag)
}

// UpdateExpenseTagHandler updates an existing expense tag
func (a *App) UpdateExpenseTagHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid tag ID")
		return
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Active      bool   `json:"active"`
		Budget      *int   `json:"budget"` // Budget in cents
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if req.Name == "" {
		respondWithError(w, http.StatusBadRequest, "Tag name is required")
		return
	}

	if err := a.cronosApp.UpdateExpenseTag(uint(id), req.Name, req.Description, req.Active, req.Budget); err != nil {
		log.Printf("Failed to update expense tag: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to update expense tag")
		return
	}

	// Fetch updated tag to return
	tag, err := a.cronosApp.GetExpenseTag(uint(id))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch updated tag")
		return
	}

	respondWithJSON(w, http.StatusOK, tag)
}

// DeleteExpenseTagHandler deletes an expense tag
func (a *App) DeleteExpenseTagHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid tag ID")
		return
	}

	if err := a.cronosApp.DeleteExpenseTag(uint(id)); err != nil {
		log.Printf("Failed to delete expense tag: %v", err)
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to delete expense tag: %v", err))
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Tag deleted successfully"})
}
