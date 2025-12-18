package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/snowpackdata/cronos"
	"gorm.io/gorm"
)

// GetExpensesHandler returns expenses filtered by status and/or project
// This endpoint always returns only the current user's expenses
func (a *App) GetExpensesHandler(w http.ResponseWriter, r *http.Request) {
	tenant := MustGetTenant(r.Context())
	userIDVal := r.Context().Value("user_id")
	userID, ok := userIDVal.(uint)
	if !ok || userID == 0 {
		respondWithError(w, http.StatusUnauthorized, "User ID not found in context")
		return
	}

	// Get query parameters
	status := r.URL.Query().Get("status")
	projectIDStr := r.URL.Query().Get("project_id")

	// Find the employee record for this user (within tenant)
	var employee cronos.Employee
	if err := a.cronosApp.DB.Scopes(cronos.TenantScope(tenant.ID)).Where("user_id = ?", userID).First(&employee).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respondWithError(w, http.StatusNotFound, "Employee record not found for this user")
		} else {
			respondWithError(w, http.StatusInternalServerError, "Error finding employee record")
		}
		return
	}

	query := a.cronosApp.DB.Scopes(cronos.TenantScope(tenant.ID))

	// Always filter by current user's expenses only
	query = query.Where("submitter_id = ?", employee.ID)

	if status != "" {
		query = query.Where("state = ?", status)
	}

	if projectIDStr != "" {
		projectID, err := strconv.ParseUint(projectIDStr, 10, 64)
		if err == nil {
			query = query.Where("project_id = ?", uint(projectID))
		}
	}

	var expenses []cronos.Expense
	if err := query.
		Preload("Project").
		Preload("Submitter.HeadshotAsset").
		Preload("Category").
		Preload("Tags").
		Preload("Receipt").
		Order("date DESC").
		Find(&expenses).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error fetching expenses")
		return
	}

	respondWithJSON(w, http.StatusOK, expenses)
}

// GetExpensesForReviewHandler returns all expenses for admin review
// This endpoint is admin-only and shows all expenses across all users
func (a *App) GetExpensesForReviewHandler(w http.ResponseWriter, r *http.Request) {
	tenant := MustGetTenant(r.Context())

	// Check if user has admin role
	userRoleVal := r.Context().Value("user_role")
	userRole, _ := userRoleVal.(string)
	if userRole != "ADMIN" {
		respondWithError(w, http.StatusForbidden, "Admin access required")
		return
	}

	// Get query parameters
	status := r.URL.Query().Get("status")
	projectIDStr := r.URL.Query().Get("project_id")

	query := a.cronosApp.DB.Scopes(cronos.TenantScope(tenant.ID))

	if status != "" {
		query = query.Where("state = ?", status)
	}

	if projectIDStr != "" {
		projectID, err := strconv.ParseUint(projectIDStr, 10, 64)
		if err == nil {
			query = query.Where("project_id = ?", uint(projectID))
		}
	}

	var expenses []cronos.Expense
	if err := query.
		Preload("Project").
		Preload("Submitter.HeadshotAsset").
		Preload("Category").
		Preload("Tags").
		Preload("Receipt").
		Order("date DESC").
		Find(&expenses).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error fetching expenses")
		return
	}

	respondWithJSON(w, http.StatusOK, expenses)
}

// CreateExpenseHandler creates a new expense with optional receipt upload
func (a *App) CreateExpenseHandler(w http.ResponseWriter, r *http.Request) {
	tenant := MustGetTenant(r.Context())
	userIDVal := r.Context().Value("user_id")
	userID, ok := userIDVal.(uint)
	if !ok || userID == 0 {
		respondWithError(w, http.StatusUnauthorized, "User ID not found in context")
		return
	}

	// Find the employee record for this user (within tenant)
	var employee cronos.Employee
	if err := a.cronosApp.DB.Scopes(cronos.TenantScope(tenant.ID)).Where("user_id = ?", userID).First(&employee).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respondWithError(w, http.StatusNotFound, "Employee record not found for this user")
		} else {
			respondWithError(w, http.StatusInternalServerError, "Error finding employee record")
		}
		return
	}

	// Parse multipart form
	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10MB limit
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("Could not parse form: %v", err))
		return
	}

	// Parse expense data
	// ProjectID is now optional (nullable) - for internal expenses
	projectIDStr := r.FormValue("project_id")
	var projectID *uint
	if projectIDStr != "" && projectIDStr != "null" {
		pid, err := strconv.ParseUint(projectIDStr, 10, 64)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid project ID")
			return
		}
		pidUint := uint(pid)
		projectID = &pidUint
	}

	amountStr := r.FormValue("amount")
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid amount")
		return
	}
	amountCents := int(amount * 100)

	dateStr := r.FormValue("date")
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid date format")
		return
	}

	description := r.FormValue("description")
	if description == "" {
		respondWithError(w, http.StatusBadRequest, "Description is required")
		return
	}

	// Parse category ID (required)
	categoryIDStr := r.FormValue("category_id")
	categoryID, err := strconv.ParseUint(categoryIDStr, 10, 64)
	if err != nil || categoryID == 0 {
		respondWithError(w, http.StatusBadRequest, "Valid category ID is required")
		return
	}

	// Parse tag IDs (optional, comma-separated)
	var tagIDs []uint
	tagIDsStr := r.FormValue("tag_ids")
	log.Printf("CreateExpense - Received tag_ids string: '%s'", tagIDsStr)
	if tagIDsStr != "" {
		tagIDsStrSlice := strings.Split(tagIDsStr, ",")
		for _, tidStr := range tagIDsStrSlice {
			tidStr = strings.TrimSpace(tidStr)
			if tidStr == "" {
				continue
			}
			tid, err := strconv.ParseUint(tidStr, 10, 64)
			if err != nil {
				respondWithError(w, http.StatusBadRequest, fmt.Sprintf("Invalid tag ID: %s", tidStr))
				return
			}
			tagIDs = append(tagIDs, uint(tid))
		}
	}
	log.Printf("CreateExpense - Parsed tag IDs: %v", tagIDs)

	// Payment account removed - will be determined during bank reconciliation

	// Handle receipt upload if provided
	var receiptID *uint
	file, header, err := r.FormFile("receipt")
	if err == nil {
		defer file.Close()

		// Read file content
		fileBytes, readErr := io.ReadAll(file)
		if readErr != nil {
			respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Error reading file: %v", readErr))
			return
		}

		contentType := http.DetectContentType(fileBytes)
		if header.Header.Get("Content-Type") != "" {
			contentType = header.Header.Get("Content-Type")
		}

		// Upload to GCS
		bucketName := a.cronosApp.Bucket
		if bucketName == "" {
			respondWithError(w, http.StatusInternalServerError, "GCS bucket not configured")
			return
		}

		// Generate UUID for secure filename
		newUUID, errUUID := uuid.NewRandom()
		if errUUID != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to generate unique filename")
			return
		}

		ext := filepath.Ext(header.Filename)
		// Use project ID if available, otherwise use "internal" for internal expenses
		projectFolder := "internal"
		if projectID != nil {
			projectFolder = fmt.Sprintf("%d", *projectID)
		}
		objectName := fmt.Sprintf("assets/expenses/%s/%s%s", projectFolder, newUUID.String(), ext)

		if errUpload := a.cronosApp.UploadObject(r.Context(), bucketName, objectName, bytes.NewReader(fileBytes), contentType); errUpload != nil {
			log.Printf("Failed to upload receipt: %v", errUpload)
			respondWithError(w, http.StatusInternalServerError, "Failed to upload receipt")
			return
		}

		// Keep files private - generate signed URLs on demand
		url := a.cronosApp.GetObjectURL(bucketName, objectName)
		var expiresAt *time.Time

		signedURL, expiresTime, signedURLErr := a.cronosApp.GenerateSignedURL(bucketName, objectName)
		if signedURLErr != nil {
			log.Printf("Failed to generate signed URL (using public URL instead): %v", signedURLErr)
			// Fallback to direct public URL
			url = a.cronosApp.GetObjectURL(bucketName, objectName)
		} else {
			url = signedURL
			expiresAt = &expiresTime
		}

		// Create asset record
		size := int64(len(fileBytes))
		uploadStatus := string(cronos.AssetUploadStatusCompleted)
		now := time.Now().UTC()

		asset := cronos.Asset{
			Name:          header.Filename,
			AssetType:     contentType,
			ContentType:   &contentType,
			Size:          &size,
			GCSObjectPath: &objectName,
			BucketName:    &bucketName,
			Url:           url,
			ExpiresAt:     expiresAt,
			UploadedBy:    &userID,
			UploadedAt:    &now,
			UploadStatus:  &uploadStatus,
		}

		asset.TenantID = tenant.ID
		if err := a.cronosApp.DB.Create(&asset).Error; err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to create asset record")
			return
		}

		receiptID = &asset.ID
	}

	// Create expense
	// Payment account removed - will be set during bank reconciliation
	expense := cronos.Expense{
		ProjectID:   projectID, // Now nullable for internal expenses
		SubmitterID: employee.ID,
		Amount:      amountCents,
		Date:        date,
		Description: description,
		State:       cronos.ExpenseStateDraft.String(),
		ReceiptID:   receiptID,
		CategoryID:  uint(categoryID),
	}

	expense.TenantID = tenant.ID
	if err := a.cronosApp.DB.Create(&expense).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create expense")
		return
	}

	// Assign tags (always call this, even if empty, to ensure consistency)
	log.Printf("CreateExpense - Assigning %d tags to expense ID %d", len(tagIDs), expense.ID)
	if err := a.cronosApp.AssignTagsToExpense(expense.ID, tagIDs); err != nil {
		log.Printf("Failed to assign tags to expense: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to assign tags to expense")
		return
	}

	// Reload with associations
	if err := a.cronosApp.DB.
		Preload("Project").
		Preload("Submitter.HeadshotAsset").
		Preload("Receipt").
		Preload("Category").
		Preload("Tags").
		First(&expense, expense.ID).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to reload expense")
		return
	}

	respondWithJSON(w, http.StatusCreated, expense)
}

// UpdateExpenseHandler updates an expense
func (a *App) UpdateExpenseHandler(w http.ResponseWriter, r *http.Request) {
	tenant := MustGetTenant(r.Context())
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid expense ID")
		return
	}

	userIDVal := r.Context().Value("user_id")
	userID, ok := userIDVal.(uint)
	if !ok || userID == 0 {
		respondWithError(w, http.StatusUnauthorized, "User ID not found in context")
		return
	}

	// Find the employee record (within tenant)
	var employee cronos.Employee
	if err := a.cronosApp.DB.Scopes(cronos.TenantScope(tenant.ID)).Where("user_id = ?", userID).First(&employee).Error; err != nil {
		respondWithError(w, http.StatusNotFound, "Employee record not found")
		return
	}

	// Find the expense (within tenant)
	var expense cronos.Expense
	if err := a.cronosApp.DB.Scopes(cronos.TenantScope(tenant.ID)).First(&expense, uint(id)).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respondWithError(w, http.StatusNotFound, "Expense not found")
		} else {
			respondWithError(w, http.StatusInternalServerError, "Error finding expense")
		}
		return
	}

	// Only submitter can edit draft expenses
	if expense.SubmitterID != employee.ID {
		respondWithError(w, http.StatusForbidden, "You can only edit your own expenses")
		return
	}

	if expense.State != cronos.ExpenseStateDraft.String() {
		respondWithError(w, http.StatusForbidden, "Can only edit draft expenses")
		return
	}

	// Parse form data (try multipart first, fall back to regular form)
	contentType := r.Header.Get("Content-Type")
	log.Printf("UpdateExpense - Content-Type: %s", contentType)
	if strings.Contains(contentType, "multipart/form-data") {
		if err := r.ParseMultipartForm(32 << 20); err != nil { // 32 MB limit
			log.Printf("Failed to parse multipart form: %v", err)
			respondWithError(w, http.StatusBadRequest, "Failed to parse multipart form data")
			return
		}
	} else {
		if err := r.ParseForm(); err != nil {
			log.Printf("Failed to parse form: %v", err)
			respondWithError(w, http.StatusBadRequest, "Failed to parse form data")
			return
		}
	}

	log.Printf("UpdateExpense - Form values: project_id=%s, amount=%s, date=%s, description=%s, category_id=%s, tag_ids=%s",
		r.FormValue("project_id"), r.FormValue("amount"), r.FormValue("date"), r.FormValue("description"),
		r.FormValue("category_id"), r.FormValue("tag_ids"))

	// Update project ID (nullable for internal expenses)
	if projectIDStr := r.FormValue("project_id"); projectIDStr != "" && projectIDStr != "null" {
		projectID, errParse := strconv.ParseUint(projectIDStr, 10, 64)
		if errParse != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid project ID")
			return
		}
		pidUint := uint(projectID)
		expense.ProjectID = &pidUint
	} else {
		// Explicitly set to nil for internal expenses
		expense.ProjectID = nil
	}

	// Update amount (convert dollars to cents)
	if amountStr := r.FormValue("amount"); amountStr != "" {
		amountFloat, errParse := strconv.ParseFloat(amountStr, 64)
		if errParse != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid amount")
			return
		}
		expense.Amount = int(amountFloat * 100)
	}

	// Update date
	if dateStr := r.FormValue("date"); dateStr != "" {
		date, errParse := time.Parse("2006-01-02", dateStr)
		if errParse != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid date format")
			return
		}
		expense.Date = date
	}

	// Update description
	if description := r.FormValue("description"); description != "" {
		expense.Description = description
	}

	// Update category ID
	if categoryIDStr := r.FormValue("category_id"); categoryIDStr != "" {
		categoryID, errParse := strconv.ParseUint(categoryIDStr, 10, 64)
		if errParse != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid category ID")
			return
		}
		expense.CategoryID = uint(categoryID)
	} else {
		respondWithError(w, http.StatusBadRequest, "Category is required")
		return
	}

	// Payment account removed - will be set during bank reconciliation

	// Update tags (always process to allow clearing tags)
	var tagIDs []uint
	tagIDsStr := r.FormValue("tag_ids")
	log.Printf("UpdateExpense - Received tag_ids string: '%s'", tagIDsStr)
	if tagIDsStr != "" {
		tagIDsStrSlice := strings.Split(tagIDsStr, ",")
		for _, tidStr := range tagIDsStrSlice {
			tidStr = strings.TrimSpace(tidStr)
			if tidStr == "" {
				continue
			}
			tid, err := strconv.ParseUint(tidStr, 10, 64)
			if err != nil {
				respondWithError(w, http.StatusBadRequest, fmt.Sprintf("Invalid tag ID: %s", tidStr))
				return
			}
			tagIDs = append(tagIDs, uint(tid))
		}
	}
	log.Printf("UpdateExpense - Parsed tag IDs: %v", tagIDs)
	// Always assign tags (even if empty) to allow clearing
	log.Printf("UpdateExpense - Assigning %d tags to expense ID %d", len(tagIDs), expense.ID)
	if err := a.cronosApp.AssignTagsToExpense(expense.ID, tagIDs); err != nil {
		log.Printf("Failed to assign tags to expense: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to assign tags to expense")
		return
	}

	// Handle receipt upload if present (only for multipart requests)
	if strings.Contains(contentType, "multipart/form-data") {
		file, header, errFile := r.FormFile("receipt")
		if errFile == nil { // File is present
			defer file.Close()

			// Read file
			fileBytes, errRead := io.ReadAll(file)
			if errRead != nil {
				log.Printf("Failed to read receipt file: %v", errRead)
				respondWithError(w, http.StatusInternalServerError, "Failed to read receipt file")
				return
			}

			// Detect content type
			contentType := http.DetectContentType(fileBytes)

			// Upload to GCS
			bucketName := a.cronosApp.Bucket
			if bucketName == "" {
				log.Println("Bucket name not configured")
				respondWithError(w, http.StatusInternalServerError, "Storage not configured")
				return
			}

			// Generate UUID for secure filename
			newUUID, errUUID := uuid.NewRandom()
			if errUUID != nil {
				log.Printf("Failed to generate UUID: %v", errUUID)
				respondWithError(w, http.StatusInternalServerError, "Failed to generate filename")
				return
			}

			ext := filepath.Ext(header.Filename)
			// Use project ID if available, otherwise use "internal" for internal expenses
			projectFolder := "internal"
			if expense.ProjectID != nil {
				projectFolder = fmt.Sprintf("%d", *expense.ProjectID)
			}
			objectName := fmt.Sprintf("assets/expenses/%s/%s%s", projectFolder, newUUID.String(), ext)

			if errUpload := a.cronosApp.UploadObject(r.Context(), bucketName, objectName, bytes.NewReader(fileBytes), contentType); errUpload != nil {
				log.Printf("Failed to upload receipt: %v", errUpload)
				respondWithError(w, http.StatusInternalServerError, "Failed to upload receipt")
				return
			}

			// Generate URL
			url := a.cronosApp.GetObjectURL(bucketName, objectName)
			var expiresAt *time.Time

			signedURL, expiresTime, signedURLErr := a.cronosApp.GenerateSignedURL(bucketName, objectName)
			if signedURLErr != nil {
				log.Printf("Failed to generate signed URL (using public URL instead): %v", signedURLErr)
				url = a.cronosApp.GetObjectURL(bucketName, objectName)
				expiresAt = nil
			} else {
				url = signedURL
				expiresAt = &expiresTime
			}

			// Create or update asset record
			size := int64(len(fileBytes))
			uploadStatus := "completed"
			uploadedAt := time.Now()

			asset := cronos.Asset{
				ProjectID:     expense.ProjectID, // Already a pointer
				AssetType:     "receipt",
				Name:          header.Filename,
				Url:           url,
				IsPublic:      false,
				BucketName:    &bucketName,
				ContentType:   &contentType,
				Size:          &size,
				UploadStatus:  &uploadStatus,
				UploadedBy:    &userID,
				UploadedAt:    &uploadedAt,
				ExpiresAt:     expiresAt,
				GCSObjectPath: &objectName,
			}

			if errAsset := a.cronosApp.DB.Create(&asset).Error; errAsset != nil {
				log.Printf("Failed to create asset record: %v", errAsset)
				respondWithError(w, http.StatusInternalServerError, "Failed to save receipt metadata")
				return
			}

			expense.ReceiptID = &asset.ID
		} else if errFile != http.ErrMissingFile {
			log.Printf("Error accessing receipt file: %v", errFile)
			respondWithError(w, http.StatusBadRequest, "Error processing receipt file")
			return
		}
	}

	// Save updated expense
	if err := a.cronosApp.DB.Save(&expense).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to update expense")
		return
	}

	// Reload with associations
	if err := a.cronosApp.DB.
		Preload("Project").
		Preload("Submitter.HeadshotAsset").
		Preload("Receipt").
		Preload("Category").
		Preload("Tags").
		First(&expense, expense.ID).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to reload expense")
		return
	}

	respondWithJSON(w, http.StatusOK, expense)
}

// SubmitExpenseHandler submits an expense for approval
func (a *App) SubmitExpenseHandler(w http.ResponseWriter, r *http.Request) {
	tenant := MustGetTenant(r.Context())
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid expense ID")
		return
	}

	userIDVal := r.Context().Value("user_id")
	userID, ok := userIDVal.(uint)
	if !ok || userID == 0 {
		respondWithError(w, http.StatusUnauthorized, "User ID not found in context")
		return
	}

	// Find the employee record (within tenant)
	var employee cronos.Employee
	if err := a.cronosApp.DB.Scopes(cronos.TenantScope(tenant.ID)).Where("user_id = ?", userID).First(&employee).Error; err != nil {
		respondWithError(w, http.StatusNotFound, "Employee record not found")
		return
	}

	// Find the expense (within tenant)
	var expense cronos.Expense
	if err := a.cronosApp.DB.Scopes(cronos.TenantScope(tenant.ID)).First(&expense, uint(id)).Error; err != nil {
		respondWithError(w, http.StatusNotFound, "Expense not found")
		return
	}

	// Verify ownership
	if expense.SubmitterID != employee.ID {
		respondWithError(w, http.StatusForbidden, "You can only submit your own expenses")
		return
	}

	if expense.State != cronos.ExpenseStateDraft.String() {
		respondWithError(w, http.StatusBadRequest, "Can only submit draft expenses")
		return
	}

	// Update state
	expense.State = cronos.ExpenseStateSubmitted.String()
	if err := a.cronosApp.DB.Save(&expense).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to submit expense")
		return
	}

	// Reload with associations (within tenant)
	if err := a.cronosApp.DB.Scopes(cronos.TenantScope(tenant.ID)).
		Preload("Project").
		Preload("Submitter.HeadshotAsset").
		Preload("Receipt").
		Preload("Category").
		Preload("Tags").
		First(&expense, expense.ID).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to reload expense")
		return
	}

	respondWithJSON(w, http.StatusOK, expense)
}

// ApproveExpenseHandler approves an expense (admin only)
func (a *App) ApproveExpenseHandler(w http.ResponseWriter, r *http.Request) {
	tenant := MustGetTenant(r.Context())
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid expense ID")
		return
	}

	userIDVal := r.Context().Value("user_id")
	userID, ok := userIDVal.(uint)
	if !ok || userID == 0 {
		respondWithError(w, http.StatusUnauthorized, "User ID not found in context")
		return
	}

	// Find the employee record (within tenant)
	var employee cronos.Employee
	if err := a.cronosApp.DB.Scopes(cronos.TenantScope(tenant.ID)).Where("user_id = ?", userID).First(&employee).Error; err != nil {
		respondWithError(w, http.StatusNotFound, "Employee record not found")
		return
	}

	// Use the ApproveExpense function which handles invoice creation and GL booking
	if err := a.cronosApp.ApproveExpense(uint(id), employee.ID); err != nil {
		log.Printf("Failed to approve expense: %v", err)
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to approve expense: %v", err))
		return
	}

	// Reload the expense with all associations to return to frontend (within tenant)
	var expense cronos.Expense
	if err := a.cronosApp.DB.Scopes(cronos.TenantScope(tenant.ID)).
		Preload("Project").
		Preload("Submitter").
		Preload("Approver").
		Preload("Receipt").
		Preload("Category").
		Preload("Tags").
		First(&expense, uint(id)).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to reload expense")
		return
	}

	respondWithJSON(w, http.StatusOK, expense)
}

// RejectExpenseHandler rejects an expense (admin only)
func (a *App) RejectExpenseHandler(w http.ResponseWriter, r *http.Request) {
	tenant := MustGetTenant(r.Context())
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid expense ID")
		return
	}

	userIDVal := r.Context().Value("user_id")
	userID, ok := userIDVal.(uint)
	if !ok || userID == 0 {
		respondWithError(w, http.StatusUnauthorized, "User ID not found in context")
		return
	}

	// Find the employee record
	var employee cronos.Employee
	if err := a.cronosApp.DB.Where("user_id = ?", userID).First(&employee).Error; err != nil {
		respondWithError(w, http.StatusNotFound, "Employee record not found")
		return
	}

	// Parse rejection reason
	var reqBody struct {
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if reqBody.Reason == "" {
		respondWithError(w, http.StatusBadRequest, "Rejection reason is required")
		return
	}

	// Use the RejectExpense function
	if err := a.cronosApp.RejectExpense(uint(id), employee.ID, reqBody.Reason); err != nil {
		log.Printf("Failed to reject expense: %v", err)
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to reject expense: %v", err))
		return
	}

	// Reload the expense with all associations to return to frontend (within tenant)
	var expense cronos.Expense
	if err := a.cronosApp.DB.Scopes(cronos.TenantScope(tenant.ID)).
		Preload("Project").
		Preload("Submitter").
		Preload("Approver").
		Preload("Receipt").
		Preload("Category").
		Preload("Tags").
		First(&expense, uint(id)).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to reload expense")
		return
	}

	respondWithJSON(w, http.StatusOK, expense)
}

// DeleteExpenseHandler deletes a draft expense
func (a *App) DeleteExpenseHandler(w http.ResponseWriter, r *http.Request) {
	tenant := MustGetTenant(r.Context())
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid expense ID")
		return
	}

	userIDVal := r.Context().Value("user_id")
	userID, ok := userIDVal.(uint)
	if !ok || userID == 0 {
		respondWithError(w, http.StatusUnauthorized, "User ID not found in context")
		return
	}

	// Find the employee record (within tenant)
	var employee cronos.Employee
	if err := a.cronosApp.DB.Scopes(cronos.TenantScope(tenant.ID)).Where("user_id = ?", userID).First(&employee).Error; err != nil {
		respondWithError(w, http.StatusNotFound, "Employee record not found")
		return
	}

	// Find the expense (within tenant)
	var expense cronos.Expense
	if err := a.cronosApp.DB.Scopes(cronos.TenantScope(tenant.ID)).First(&expense, uint(id)).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respondWithError(w, http.StatusNotFound, "Expense not found")
		} else {
			respondWithError(w, http.StatusInternalServerError, "Error finding expense")
		}
		return
	}

	// Only submitter can delete their own draft expenses
	if expense.SubmitterID != employee.ID {
		respondWithError(w, http.StatusForbidden, "You can only delete your own expenses")
		return
	}

	if expense.State != cronos.ExpenseStateDraft.String() {
		respondWithError(w, http.StatusForbidden, "Can only delete draft expenses")
		return
	}

	// Delete the expense (soft delete via GORM)
	if err := a.cronosApp.DB.Delete(&expense).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to delete expense")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Expense deleted"})
}

// RefreshExpenseReceiptURLHandler refreshes the signed URL for an expense receipt
func (a *App) RefreshExpenseReceiptURLHandler(w http.ResponseWriter, r *http.Request) {
	tenant := MustGetTenant(r.Context())
	vars := mux.Vars(r)
	assetIDStr := vars["assetId"]
	assetID, err := strconv.ParseUint(assetIDStr, 10, 64)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid asset ID")
		return
	}

	var asset cronos.Asset
	if err := a.cronosApp.DB.Scopes(cronos.TenantScope(tenant.ID)).First(&asset, uint(assetID)).Error; err != nil {
		respondWithError(w, http.StatusNotFound, "Asset not found")
		return
	}

	if asset.BucketName == nil || asset.GCSObjectPath == nil {
		respondWithError(w, http.StatusBadRequest, "Asset is not stored in GCS")
		return
	}

	signedURL, expiresTime, signedURLErr := a.cronosApp.GenerateSignedURL(*asset.BucketName, *asset.GCSObjectPath)
	if signedURLErr != nil {
		log.Printf("Failed to generate signed URL for asset %d: %v", assetID, signedURLErr)
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to generate signed URL: %v", signedURLErr))
		return
	}

	asset.Url = signedURL
	asset.ExpiresAt = &expiresTime
	if err := a.cronosApp.DB.Save(&asset).Error; err != nil {
		log.Printf("Failed to update asset URL: %v", err)
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"url":        signedURL,
		"expires_at": expiresTime,
	})
}
