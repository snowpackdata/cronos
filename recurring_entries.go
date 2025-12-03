package cronos

import (
	"fmt"
	"log"
	"strings"
	"time"
)

// GenerateRecurringEntriesForPeriod creates recurring bill line items for all active employees
// for a given period (typically the current month). This is idempotent - safe to call multiple times.
func (a *App) GenerateRecurringEntriesForPeriod(periodStart, periodEnd time.Time) error {
	log.Printf("Generating recurring entries for period %s to %s", periodStart.Format("2006-01-02"), periodEnd.Format("2006-01-02"))

	// Find all active recurring entries
	var recurringEntries []RecurringEntry
	if err := a.DB.
		Preload("Employee").
		Where("is_active = ?", true).
		Where("start_date <= ?", periodEnd).
		Where("end_date IS NULL OR end_date >= ?", periodStart).
		Find(&recurringEntries).Error; err != nil {
		return fmt.Errorf("failed to load recurring entries: %w", err)
	}

	log.Printf("Found %d active recurring entries to process", len(recurringEntries))

	generated := 0
	skipped := 0

	for _, entry := range recurringEntries {
		// Check if we've already generated for this period
		if entry.LastGeneratedFor != nil && !entry.LastGeneratedFor.Before(periodStart) {
			log.Printf("Skipping recurring entry %d (employee %d) - already generated for period %s (last generated: %s)", 
				entry.ID, entry.EmployeeID, periodStart.Format("2006-01-02"), entry.LastGeneratedFor.Format("2006-01-02"))
			skipped++
			continue
		}
		
		log.Printf("Processing recurring entry %d for employee %d (%s %s)", 
			entry.ID, entry.EmployeeID, entry.Employee.FirstName, entry.Employee.LastName)

		// Find or create a bill for this employee for this period
		var bill Bill
		if err := a.DB.Where("employee_id = ? AND period_start = ? AND period_end = ?",
			entry.EmployeeID, periodStart, periodEnd).First(&bill).Error; err != nil {
			// Bill doesn't exist, create it
			bill = Bill{
				Name:        fmt.Sprintf("%s - %s", entry.Employee.FirstName+" "+entry.Employee.LastName, periodStart.Format("Jan 2006")),
				State:       BillStateDraft,
				EmployeeID:  entry.EmployeeID,
				PeriodStart: periodStart,
				PeriodEnd:   periodEnd,
			}
			if err := a.DB.Create(&bill).Error; err != nil {
				log.Printf("Failed to create bill for employee %d: %v", entry.EmployeeID, err)
				continue
			}
			log.Printf("Created new bill %d for employee %d", bill.ID, entry.EmployeeID)
		}

		// Check if this recurring entry is already on the bill
		var existingLineItem RecurringBillLineItem
		err := a.DB.Where("bill_id = ? AND recurring_entry_id = ?", bill.ID, entry.ID).First(&existingLineItem).Error
		if err == nil {
			log.Printf("Recurring line item already exists for bill %d, entry %d - skipping", bill.ID, entry.ID)
			skipped++
			continue
		}

		// Create the recurring bill line item
		lineItem := RecurringBillLineItem{
			BillID:           bill.ID,
			RecurringEntryID: entry.ID,
			Description:      entry.Description,
			Amount:           entry.Amount,
			PeriodStart:      periodStart,
			PeriodEnd:        periodEnd,
			State:            "pending", // Will be approved when bill is approved
		}

		if err := a.DB.Create(&lineItem).Error; err != nil {
			log.Printf("Failed to create recurring line item: %v", err)
			continue
		}

		// Update last generated tracking
		now := time.Now()
		entry.LastGeneratedDate = &now
		entry.LastGeneratedFor = &periodStart
		if err := a.DB.Save(&entry).Error; err != nil {
			log.Printf("Failed to update recurring entry tracking: %v", err)
		}

		// Recalculate bill totals to include recurring entry
		a.RecalculateBillTotals(&bill)

		// Book payroll accrual for this recurring entry
		if err := a.BookRecurringEntryAccrual(&lineItem, &entry.Employee); err != nil {
			log.Printf("Failed to book accrual for recurring entry: %v", err)
		}

		log.Printf("Created recurring line item for employee %d: %s ($%.2f)", 
			entry.EmployeeID, entry.Description, float64(entry.Amount)/100)
		generated++
	}

	log.Printf("Recurring entry generation complete: %d created, %d skipped", generated, skipped)
	return nil
}

// CreateRecurringEntryForEmployee creates a recurring entry for an employee's base salary
// This should be called when an employee is hired or their salary is changed
func (a *App) CreateRecurringEntryForEmployee(employeeID uint) error {
	var employee Employee
	if err := a.DB.First(&employee, employeeID).Error; err != nil {
		return fmt.Errorf("failed to load employee: %w", err)
	}

	// Only create recurring entries for salaried or base+variable employees
	// Support both formats: "salaried"/"base-plus-variable" and "COMPENSATION_TYPE_*"
	compType := strings.ToUpper(employee.CompensationType)
	isSalaried := compType == "SALARIED" || compType == "COMPENSATION_TYPE_SALARIED"
	isBaseVariable := compType == "BASE-PLUS-VARIABLE" || compType == "COMPENSATION_TYPE_BASE_PLUS_VARIABLE"
	
	if !isSalaried && !isBaseVariable {
		log.Printf("Employee %d is %s - no recurring entry needed", employeeID, employee.CompensationType)
		return nil
	}

	// Check if recurring entry already exists
	var existing RecurringEntry
	err := a.DB.Where("employee_id = ? AND type = ? AND is_active = ?", 
		employeeID, "base_salary", true).First(&existing).Error
	
	if err == nil {
		// Update existing entry if salary changed
		monthlySalary := employee.SalaryAnnualized / 12
		if existing.Amount != monthlySalary {
			log.Printf("Updating recurring entry %d for employee %d: $%d -> $%d", 
				existing.ID, employeeID, existing.Amount, monthlySalary)
			existing.Amount = monthlySalary
			return a.DB.Save(&existing).Error
		}
		log.Printf("Recurring entry already exists for employee %d", employeeID)
		return nil
	}

	// Create new recurring entry
	monthlySalary := employee.SalaryAnnualized / 12 // Convert annual to monthly (in cents)
	
	entry := RecurringEntry{
		EmployeeID:  employeeID,
		Type:        "base_salary",
		Description: "Monthly Base Salary",
		Amount:      monthlySalary,
		Frequency:   "monthly",
		StartDate:   employee.StartDate,
		IsActive:    true,
	}

	if err := a.DB.Create(&entry).Error; err != nil {
		return fmt.Errorf("failed to create recurring entry: %w", err)
	}

	log.Printf("Created recurring entry for employee %d: $%.2f/month", employeeID, float64(monthlySalary)/100)
	return nil
}

// GenerateRecurringEntriesForCurrentMonth is a convenience function for manual/cron triggering
func (a *App) GenerateRecurringEntriesForCurrentMonth() error {
	now := time.Now()
	// Get first and last day of current month
	periodStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	periodEnd := periodStart.AddDate(0, 1, -1) // Last day of month
	
	return a.GenerateRecurringEntriesForPeriod(periodStart, periodEnd)
}

// BookRecurringEntryAccrual books payroll expense and accrued payroll for a recurring entry
// DR: Payroll Expense, CR: Accrued Payroll
func (a *App) BookRecurringEntryAccrual(lineItem *RecurringBillLineItem, employee *Employee) error {
	employeeName := fmt.Sprintf("%s %s", employee.FirstName, employee.LastName)
	subAccount := fmt.Sprintf("%d:%s", employee.ID, employeeName)

	// Check if we've already booked this recurring entry
	var existing []Journal
	a.DB.Where("recurring_bill_line_item_id = ?", lineItem.ID).Find(&existing)
	if len(existing) > 0 {
		log.Printf("Accrual already booked for recurring line item %d, skipping", lineItem.ID)
		return nil
	}

	// DR: Payroll Expense
	debitEntry := Journal{
		Account:                   AccountPayrollExpense.String(),
		SubAccount:                subAccount,
		RecurringBillLineItemID:   &lineItem.ID,
		BillID:                    &lineItem.BillID,
		Memo:                      fmt.Sprintf("Recurring payroll: %s", lineItem.Description),
		Debit:                     int64(lineItem.Amount),
		Credit:                    0,
	}
	if err := a.DB.Create(&debitEntry).Error; err != nil {
		return fmt.Errorf("failed to book payroll expense: %w", err)
	}

	// CR: Accrued Payroll
	creditEntry := Journal{
		Account:                   AccountAccruedPayroll.String(),
		SubAccount:                subAccount,
		RecurringBillLineItemID:   &lineItem.ID,
		BillID:                    &lineItem.BillID,
		Memo:                      fmt.Sprintf("Recurring payroll: %s", lineItem.Description),
		Debit:                     0,
		Credit:                    int64(lineItem.Amount),
	}
	if err := a.DB.Create(&creditEntry).Error; err != nil {
		return fmt.Errorf("failed to book accrued payroll: %w", err)
	}

	// Update line item state
	lineItem.State = "approved"
	if err := a.DB.Save(lineItem).Error; err != nil {
		return fmt.Errorf("failed to update line item state: %w", err)
	}

	log.Printf("Booked recurring entry accrual: $%.2f for %s", float64(lineItem.Amount)/100, employeeName)
	return nil
}

