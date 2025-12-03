package cronos

import (
	"fmt"
	"log"
)

// CreateExpenseCategory creates a new expense category
func (a *App) CreateExpenseCategory(name, description string) (*ExpenseCategory, error) {
	category := ExpenseCategory{
		Name:        name,
		Description: description,
		Active:      true,
	}

	if err := a.DB.Create(&category).Error; err != nil {
		return nil, fmt.Errorf("failed to create expense category: %w", err)
	}

	log.Printf("Created expense category: %s (ID: %d)", category.Name, category.ID)
	return &category, nil
}

// GetExpenseCategories retrieves all expense categories
func (a *App) GetExpenseCategories(activeOnly bool) ([]ExpenseCategory, error) {
	var categories []ExpenseCategory
	query := a.DB

	if activeOnly {
		query = query.Where("active = ?", true)
	}

	if err := query.Order("name ASC").Find(&categories).Error; err != nil {
		return nil, fmt.Errorf("failed to get expense categories: %w", err)
	}

	return categories, nil
}

// GetExpenseCategory retrieves a single expense category by ID
func (a *App) GetExpenseCategory(id uint) (*ExpenseCategory, error) {
	var category ExpenseCategory
	if err := a.DB.First(&category, id).Error; err != nil {
		return nil, fmt.Errorf("failed to get expense category: %w", err)
	}
	return &category, nil
}

// UpdateExpenseCategory updates an existing expense category
func (a *App) UpdateExpenseCategory(id uint, name, description string, active bool) error {
	var category ExpenseCategory
	if err := a.DB.First(&category, id).Error; err != nil {
		return fmt.Errorf("failed to find expense category: %w", err)
	}

	category.Name = name
	category.Description = description
	category.Active = active

	if err := a.DB.Save(&category).Error; err != nil {
		return fmt.Errorf("failed to update expense category: %w", err)
	}

	log.Printf("Updated expense category ID %d: %s", id, name)
	return nil
}

// DeleteExpenseCategory soft deletes an expense category
func (a *App) DeleteExpenseCategory(id uint) error {
	if err := a.DB.Delete(&ExpenseCategory{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete expense category: %w", err)
	}

	log.Printf("Deleted expense category ID %d", id)
	return nil
}

// CreateExpenseTag creates a new expense tag
func (a *App) CreateExpenseTag(name, description string, active bool, budget *int) (*ExpenseTag, error) {
	tag := ExpenseTag{
		Name:        name,
		Description: description,
		Active:      active,
		Budget:      budget,
	}

	if err := a.DB.Create(&tag).Error; err != nil {
		return nil, fmt.Errorf("failed to create expense tag: %w", err)
	}

	log.Printf("Created expense tag: %s (ID: %d)", tag.Name, tag.ID)
	return &tag, nil
}

// GetExpenseTags retrieves all expense tags with optional spend summaries
func (a *App) GetExpenseTags(activeOnly bool) ([]ExpenseTag, error) {
	var tags []ExpenseTag
	query := a.DB

	if activeOnly {
		query = query.Where("active = ?", true)
	}

	if err := query.Order("name ASC").Find(&tags).Error; err != nil {
		return nil, fmt.Errorf("failed to get expense tags: %w", err)
	}

	return tags, nil
}

// GetExpenseTag retrieves a single expense tag by ID
func (a *App) GetExpenseTag(id uint) (*ExpenseTag, error) {
	var tag ExpenseTag
	if err := a.DB.First(&tag, id).Error; err != nil {
		return nil, fmt.Errorf("failed to get expense tag: %w", err)
	}
	return &tag, nil
}

// UpdateExpenseTag updates an existing expense tag
func (a *App) UpdateExpenseTag(id uint, name, description string, active bool, budget *int) error {
	var tag ExpenseTag
	if err := a.DB.First(&tag, id).Error; err != nil {
		return fmt.Errorf("failed to find expense tag: %w", err)
	}

	tag.Name = name
	tag.Description = description
	tag.Active = active
	tag.Budget = budget

	if err := a.DB.Save(&tag).Error; err != nil {
		return fmt.Errorf("failed to update expense tag: %w", err)
	}

	log.Printf("Updated expense tag ID %d: %s", id, name)
	return nil
}

// DeleteExpenseTag soft deletes an expense tag
func (a *App) DeleteExpenseTag(id uint) error {
	// Check if tag is in use
	var count int64
	if err := a.DB.Model(&ExpenseTagAssignment{}).Where("tag_id = ?", id).Count(&count).Error; err != nil {
		return fmt.Errorf("failed to check tag usage: %w", err)
	}

	if count > 0 {
		return fmt.Errorf("cannot delete tag that is in use by %d expense(s)", count)
	}

	if err := a.DB.Delete(&ExpenseTag{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete expense tag: %w", err)
	}

	log.Printf("Deleted expense tag ID %d", id)
	return nil
}

// AssignTagsToExpense assigns multiple tags to an expense
func (a *App) AssignTagsToExpense(expenseID uint, tagIDs []uint) error {
	// Remove existing tag assignments
	if err := a.DB.Where("expense_id = ?", expenseID).Delete(&ExpenseTagAssignment{}).Error; err != nil {
		return fmt.Errorf("failed to remove existing tag assignments: %w", err)
	}

	// Add new tag assignments
	for _, tagID := range tagIDs {
		assignment := ExpenseTagAssignment{
			ExpenseID: expenseID,
			TagID:     tagID,
		}
		if err := a.DB.Create(&assignment).Error; err != nil {
			return fmt.Errorf("failed to assign tag %d to expense: %w", tagID, err)
		}
	}

	log.Printf("Assigned %d tags to expense ID %d", len(tagIDs), expenseID)
	return nil
}

// GetExpensesByTag retrieves all expenses with a specific tag
func (a *App) GetExpensesByTag(tagID uint) ([]Expense, error) {
	var expenses []Expense
	if err := a.DB.
		Joins("JOIN expense_tag_assignments ON expense_tag_assignments.expense_id = expenses.id").
		Where("expense_tag_assignments.expense_tag_id = ?", tagID).
		Preload("Project").
		Preload("Submitter").
		Preload("Category").
		Preload("Tags").
		Find(&expenses).Error; err != nil {
		return nil, fmt.Errorf("failed to get expenses by tag: %w", err)
	}

	return expenses, nil
}

