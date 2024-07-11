package cronos

import (
	"fmt"
	"time"
)

const ProjectBudgetQuery = `SELECT sum(extract(HOUR from (\"end\" - start)) + extract(MIN from (\"end\" - start)) / 60) as duration from entries where start > date_trunc('day', %s) and \"end\" < date_trunc('day', %s) and state not in ('ENTRY_STATE_VOID', 'ENTRY_STATE_UNAFFILIATED') and deleted_at is null;`
const MonthlyBudgetQuery = `SELECT sum(extract(HOUR from (\"end\" - start)) + extract(MIN from (\"end\" - start)) / 60) as duration from entries where start => date_trunc('month', %s) \"end\" < (date_trunc('MONTH', current_date) + interval '1 month') and state not in ('ENTRY_STATE_VOID', 'ENTRY_STATE_UNAFFILIATED') and deleted_at is null;`

// WORKERS

// StartValidateEntryWorkers kicks off a number of workers to validate entries as they are generated
func (a *App) StartValidateEntryWorkers(validateEntryChan <-chan Entry, workerCount int) {
	for i := 0; i < workerCount; i++ {
		go func(id int, taskChan <-chan Entry) {
			for entry := range taskChan {
				a.validateEntry(&entry)
			}
		}(i, validateEntryChan)
	}
}

// TASK FUNCTIONS

// validateEntry checks that the hours entered are within the project and billing code budget
func (a *App) validateEntry(entry *Entry) {
	// Find the BillingCode to understand the budget type
	var billingCode BillingCode
	a.DB.Where("ID = ?", entry.BillingCodeID).First(&billingCode)
	var billablesInPeriod float64
	switch billingCode.BudgetPeriod {
	case BudgetPeriodProject.String():
		// Perform a database query to get the project's budget
		query := fmt.Sprintf(ProjectBudgetQuery, billingCode.ActiveStart.Format(time.RFC3339), billingCode.ActiveEnd.Format(time.RFC3339))
		rows, err := a.DB.Raw(query).Rows()
		if err != nil {
			fmt.Println(err)
		}
		for rows.Next() {
			var duration float64
			err = rows.Scan(&duration)
			if err != nil {
				fmt.Println(err)
			}
			billablesInPeriod = duration
			fmt.Println(duration)
		}
		err = rows.Close()
		if err != nil {
			fmt.Println(err)
		}
	case BudgetPeriodMonthly.String():
		// Perform a database query to get the project's budget
		query := fmt.Sprintf(MonthlyBudgetQuery, entry.Start.Format(time.RFC3339), entry.Start.Format(time.RFC3339))
		rows, err := a.DB.Raw(query).Rows()
		if err != nil {
			fmt.Println(err)
		}
		for rows.Next() {
			var duration float64
			err = rows.Scan(&duration)
			if err != nil {
				fmt.Println(err)
			}
			billablesInPeriod = duration
			fmt.Println(duration)
		}
		err = rows.Close()
		if err != nil {
			fmt.Println(err)
		}
	}
	// Determine if the entry is within budget
	if int(billablesInPeriod) > billingCode.PeriodHours {
		entry.State = EntryStateApprovalRequired.String()
		// Send an email to the project manager and cc the staff
		// Get the User first
		var employee Employee
		a.DB.Where("ID = ?", entry.EmployeeID).First(&employee)
		var user User
		a.DB.Where("ID = ?", employee.UserID).First(&user)

		email := &Email{
			SenderName:       "Snowpack-Data",
			SenderEmail:      "accounts@snowpack-data.io",
			RecipientEmail:   user.Email,
			RecipientName:    employee.FirstName + " " + employee.LastName,
			PlainTextContent: fmt.Sprintf("Your recent entry is over budget for billing code %s and requires approval", billingCode.Code),
			Subject:          fmt.Sprintf("%s Over budget", billingCode.Name),
		}
		err := a.EmailUsersAlert(email)
		if err != nil {
			fmt.Println(err)
		}

	}
	a.DB.Save(&entry)
}
