package main

import (
	"log"
	"math"
	"net/http"
	"time"

	"github.com/snowpackdata/cronos"
)

func (a *App) PortalDraftEntriesHandler(w http.ResponseWriter, r *http.Request) {
	// Get the account ID from middleware
	accountIDVal := r.Context().Value("account_id") // Use the correct context key
	accountID, ok := accountIDVal.(uint)
	if !ok || accountID == 0 {
		log.Printf("Error: PortalDraftInvoices - Unauthorized or invalid account_id in context: %v", accountIDVal)
		respondWithError(w, http.StatusUnauthorized, "Unauthorized: Valid Account ID not found in token claims.")
		return
	}
	// First we need the distinct list of projects for the account
	var projects []cronos.Project
	if err := a.cronosApp.DB.Where("account_id = ?", accountID).Distinct().Find(&projects).Error; err != nil {
		log.Printf("Error: PortalDraftEntries - Failed to retrieve projects: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve projects.")
		return
	}
	// Get the list of projects for the account
	var projectIDs []uint
	for _, project := range projects {
		projectIDs = append(projectIDs, project.ID)
	}
	// Retrieve all draft entries for the projects on this account
	var entries []cronos.Entry
	if err := a.cronosApp.DB.Preload("Project").Preload("Employee").Where("state = ? AND project_id IN (?)", cronos.EntryStateDraft, projectIDs).Order("project_id desc, start desc").Find(&entries).Error; err != nil {
		log.Printf("Error: PortalDraftEntries - Failed to retrieve draft entries: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve draft entries.")
		return
	}
	// Return the draft entries
	respondWithJSON(w, http.StatusOK, entries)
}

// ProjectBudgetStatus defines the structure for reporting budget status of a project.
// It includes overall project budget/usage and current period budget/usage.
type ProjectBudgetStatus struct {
	ProjectID          uint      `json:"project_id"`
	ProjectName        string    `json:"project_name"`
	BillingFrequency   string    `json:"billing_frequency"`
	ProjectActiveStart time.Time `json:"project_active_start"`
	ProjectActiveEnd   time.Time `json:"project_active_end"`

	// Budget figures defined on the project model (these are per period for periodic, or total for "project" type)
	BudgetHoursPerPeriodOrTotal   int `json:"budget_hours_per_period_or_total"`
	BudgetDollarsPerPeriodOrTotal int `json:"budget_dollars_per_period_or_total"` // Assumed to be in whole dollars
	BudgetCapHours                int `json:"budget_cap_hours"`                   // Total cap on hours for the project
	BudgetCapDollars              int `json:"budget_cap_dollars"`                 // Total cap on dollars for the project

	// Overall Project Budget & Usage (calculated for the entire project duration)
	CalculatedTotalProjectBudgetHours    float64 `json:"calculated_total_project_budget_hours"`
	CalculatedTotalProjectBudgetDollars  float64 `json:"calculated_total_project_budget_dollars"`
	TotalProjectTrackedHours             float64 `json:"total_project_tracked_hours"`
	TotalProjectTrackedDollars           float64 `json:"total_project_tracked_dollars"` // In dollars
	TotalProjectCompletionHoursPercent   float64 `json:"total_project_completion_hours_percent"`
	TotalProjectCompletionDollarsPercent float64 `json:"total_project_completion_dollars_percent"`

	// Current Period Budget & Usage
	CurrentPeriodStartDate                *time.Time `json:"current_period_start_date,omitempty"`
	CurrentPeriodEndDate                  *time.Time `json:"current_period_end_date,omitempty"`
	CurrentPeriodBudgetHours              float64    `json:"current_period_budget_hours"`
	CurrentPeriodBudgetDollars            float64    `json:"current_period_budget_dollars"` // In dollars
	CurrentPeriodTrackedHours             float64    `json:"current_period_tracked_hours"`
	CurrentPeriodTrackedDollars           float64    `json:"current_period_tracked_dollars"` // In dollars
	CurrentPeriodCompletionHoursPercent   float64    `json:"current_period_completion_hours_percent"`
	CurrentPeriodCompletionDollarsPercent float64    `json:"current_period_completion_dollars_percent"`
	IsProjectBasedBudget                  bool       `json:"is_project_based_budget"` // True if BillingFrequency is BILLING_TYPE_PROJECT
}

func (a *App) PortalProjectBudgetsHandler(w http.ResponseWriter, r *http.Request) {
	accountIDVal := r.Context().Value("account_id")
	accountID, ok := accountIDVal.(uint)
	if !ok || accountID == 0 {
		log.Printf("Error: PortalProjectBudgets - Unauthorized or invalid account_id in context: %v", accountIDVal)
		respondWithError(w, http.StatusUnauthorized, "Unauthorized: Valid Account ID not found in token claims.")
		return
	}

	var projects []cronos.Project
	if err := a.cronosApp.DB.Where("account_id = ?", accountID).Find(&projects).Error; err != nil {
		log.Printf("Error: PortalProjectBudgets - Failed to retrieve projects for account ID %d: %v", accountID, err)
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve projects.")
		return
	}

	var results []ProjectBudgetStatus
	now := time.Now()

	for _, project := range projects {
		status := ProjectBudgetStatus{
			ProjectID:                     project.ID,
			ProjectName:                   project.Name,
			BillingFrequency:              project.BillingFrequency,
			ProjectActiveStart:            project.ActiveStart,
			ProjectActiveEnd:              project.ActiveEnd,
			BudgetHoursPerPeriodOrTotal:   project.BudgetHours,
			BudgetDollarsPerPeriodOrTotal: project.BudgetDollars, // Assumed to be in whole dollars
			BudgetCapHours:                project.BudgetCapHours,
			BudgetCapDollars:              project.BudgetCapDollars,
			IsProjectBasedBudget:          project.BillingFrequency == cronos.BillingFrequencyProject.String(),
		}

		var entries []cronos.Entry
		if err := a.cronosApp.DB.Where("project_id = ? AND state != ? AND deleted_at IS NULL", project.ID, cronos.EntryStateVoid.String()).
			Find(&entries).Error; err != nil {
			log.Printf("Error fetching entries for project ID %d: %v", project.ID, err)
			results = append(results, status) // Append with mostly zero values
			continue
		}

		// Calculate Total Project Tracked Hours and Dollars
		for _, entry := range entries {
			if !entry.Start.IsZero() && !entry.End.IsZero() && entry.End.After(entry.Start) {
				duration := entry.End.Sub(entry.Start)
				status.TotalProjectTrackedHours += duration.Hours()
				status.TotalProjectTrackedDollars += float64(entry.Fee) / 100.0 // Fee is in cents
			}
		}

		// Calculate Total Project Budget Hours and Dollars
		if !project.ActiveStart.IsZero() && !project.ActiveEnd.IsZero() && project.ActiveEnd.After(project.ActiveStart) {
			durationDays := project.ActiveEnd.Sub(project.ActiveStart).Hours() / 24.0
			var numPeriods float64 = 1.0 // Default for project-based or if calculation below is not met

			// If BudgetCapHours and BudgetCapDollars are set, use them directly
			if project.BudgetCapHours > 0 {
				status.CalculatedTotalProjectBudgetHours = float64(project.BudgetCapHours)
			} else {
				// Otherwise, calculate based on billing frequency and periodic budget
				switch project.BillingFrequency {
				case cronos.BillingFrequencyProject.String():
					status.CalculatedTotalProjectBudgetHours = float64(project.BudgetHours)
				case cronos.BillingFrequencyMonthly.String():
					// Calculate number of months project is active, rounding up.
					// (Year2 - Year1)*12 + Month2 - Month1 + 1
					numMonths := float64((project.ActiveEnd.Year()-project.ActiveStart.Year())*12 + int(project.ActiveEnd.Month()) - int(project.ActiveStart.Month()) + 1)
					if project.ActiveEnd.Day() < project.ActiveStart.Day() && int(project.ActiveEnd.Month()) == int(project.ActiveStart.Month()) && project.ActiveEnd.Year() == project.ActiveStart.Year() {
						// If end day is before start day in the same month/year (e.g. start Jan 15, end Jan 10), effectively 0 full months by this simple calc.
						// This needs to be handled more carefully if precise month counting is critical vs. "active in month X".
						// For simplicity, if it spans at least one day into a month, that month counts.
						// The +1 above generally handles "active in this month"
					}
					numPeriods = numMonths
					status.CalculatedTotalProjectBudgetHours = float64(project.BudgetHours) * numPeriods
				case cronos.BillingFrequencyWeekly.String():
					if durationDays > 0 {
						numPeriods = math.Ceil(durationDays / 7.0)
					}
					status.CalculatedTotalProjectBudgetHours = float64(project.BudgetHours) * numPeriods
				case cronos.BillingFrequencyBiweekly.String():
					if durationDays > 0 {
						numPeriods = math.Ceil(durationDays / 14.0)
					}
					status.CalculatedTotalProjectBudgetHours = float64(project.BudgetHours) * numPeriods
				default: // Unknown or other types
					status.CalculatedTotalProjectBudgetHours = float64(project.BudgetHours) // Default to face value
				}
			}

			// If BudgetCapDollars is set, use it directly
			if project.BudgetCapDollars > 0 {
				status.CalculatedTotalProjectBudgetDollars = float64(project.BudgetCapDollars)
			} else {
				// Otherwise, calculate based on billing frequency and periodic budget
				switch project.BillingFrequency {
				case cronos.BillingFrequencyProject.String():
					status.CalculatedTotalProjectBudgetDollars = float64(project.BudgetDollars)
				case cronos.BillingFrequencyMonthly.String():
					numMonths := float64((project.ActiveEnd.Year()-project.ActiveStart.Year())*12 + int(project.ActiveEnd.Month()) - int(project.ActiveStart.Month()) + 1)
					status.CalculatedTotalProjectBudgetDollars = float64(project.BudgetDollars) * numMonths
				case cronos.BillingFrequencyWeekly.String():
					if durationDays > 0 {
						numPeriods = math.Ceil(durationDays / 7.0)
					}
					status.CalculatedTotalProjectBudgetDollars = float64(project.BudgetDollars) * numPeriods
				case cronos.BillingFrequencyBiweekly.String():
					if durationDays > 0 {
						numPeriods = math.Ceil(durationDays / 14.0)
					}
					status.CalculatedTotalProjectBudgetDollars = float64(project.BudgetDollars) * numPeriods
				default: // Unknown or other types
					status.CalculatedTotalProjectBudgetDollars = float64(project.BudgetDollars) // Default to face value
				}
			}
		} else { // If project dates are invalid, use face values
			// If BudgetCapHours is set, use it directly
			if project.BudgetCapHours > 0 {
				status.CalculatedTotalProjectBudgetHours = float64(project.BudgetCapHours)
			} else {
				status.CalculatedTotalProjectBudgetHours = float64(project.BudgetHours)
			}

			// If BudgetCapDollars is set, use it directly
			if project.BudgetCapDollars > 0 {
				status.CalculatedTotalProjectBudgetDollars = float64(project.BudgetCapDollars)
			} else {
				status.CalculatedTotalProjectBudgetDollars = float64(project.BudgetDollars)
			}
		}

		// Calculate Current Period Budget & Usage
		var currentPeriodStart, currentPeriodEnd time.Time
		periodDefined := true

		if status.IsProjectBasedBudget {
			currentPeriodStart = project.ActiveStart
			currentPeriodEnd = project.ActiveEnd
			status.CurrentPeriodBudgetHours = status.CalculatedTotalProjectBudgetHours
			status.CurrentPeriodBudgetDollars = status.CalculatedTotalProjectBudgetDollars
		} else {
			status.CurrentPeriodBudgetHours = float64(project.BudgetHours)
			status.CurrentPeriodBudgetDollars = float64(project.BudgetDollars)

			switch project.BillingFrequency {
			case cronos.BillingFrequencyWeekly.String():
				dayOfWeek := int(now.Weekday()) // Sunday = 0, ..., Saturday = 6
				currentPeriodStart = now.AddDate(0, 0, -dayOfWeek)
				currentPeriodStart = time.Date(currentPeriodStart.Year(), currentPeriodStart.Month(), currentPeriodStart.Day(), 0, 0, 0, 0, now.Location())
				currentPeriodEnd = currentPeriodStart.AddDate(0, 0, 7).Add(-time.Nanosecond) // End of Saturday
			case cronos.BillingFrequencyMonthly.String():
				currentPeriodStart = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
				currentPeriodEnd = currentPeriodStart.AddDate(0, 1, 0).Add(-time.Nanosecond) // End of the last day of the month
			case cronos.BillingFrequencyBiweekly.String():
				// Simplified bi-weekly: determine current week, then see if it's odd/even relative to an anchor (e.g. start of year)
				// This logic matches ProjectAnalyticsHandler for period start.
				_, weekNum := now.ISOWeek()                                        // Get ISO week number
				currentPeriodStartOfWeek := now.AddDate(0, 0, -int(now.Weekday())) // Start of current Sunday
				currentPeriodStartOfWeek = time.Date(currentPeriodStartOfWeek.Year(), currentPeriodStartOfWeek.Month(), currentPeriodStartOfWeek.Day(), 0, 0, 0, 0, now.Location())

				if weekNum%2 == 0 { // Even week - this is the second week of a bi-weekly period that started last week
					currentPeriodStart = currentPeriodStartOfWeek.AddDate(0, 0, -7)
				} else { // Odd week - this is the first week of a bi-weekly period that starts this week
					currentPeriodStart = currentPeriodStartOfWeek
				}
				currentPeriodEnd = currentPeriodStart.AddDate(0, 0, 14).Add(-time.Nanosecond) // Two weeks duration
			default:
				periodDefined = false
			}
			if periodDefined {
				status.CurrentPeriodStartDate = &currentPeriodStart
				status.CurrentPeriodEndDate = &currentPeriodEnd
			}
		}

		if periodDefined {
			for _, entry := range entries {
				if !entry.Start.IsZero() && !entry.Start.Before(currentPeriodStart) && entry.Start.Before(currentPeriodEnd) {
					if !entry.End.IsZero() && entry.End.After(entry.Start) {
						duration := entry.End.Sub(entry.Start)
						status.CurrentPeriodTrackedHours += duration.Hours()
						status.CurrentPeriodTrackedDollars += float64(entry.Fee) / 100.0 // Fee is in cents
					}
				}
			}
		}

		// Calculate Completion Percentages (handle division by zero)
		if status.CalculatedTotalProjectBudgetHours > 0 {
			status.TotalProjectCompletionHoursPercent = (status.TotalProjectTrackedHours / status.CalculatedTotalProjectBudgetHours) * 100
		}
		if status.CalculatedTotalProjectBudgetDollars > 0 {
			status.TotalProjectCompletionDollarsPercent = (status.TotalProjectTrackedDollars / status.CalculatedTotalProjectBudgetDollars) * 100
		}
		if status.CurrentPeriodBudgetHours > 0 {
			status.CurrentPeriodCompletionHoursPercent = (status.CurrentPeriodTrackedHours / status.CurrentPeriodBudgetHours) * 100
		}
		if status.CurrentPeriodBudgetDollars > 0 {
			status.CurrentPeriodCompletionDollarsPercent = (status.CurrentPeriodTrackedDollars / status.CurrentPeriodBudgetDollars) * 100
		}

		results = append(results, status)
	}
	respondWithJSON(w, http.StatusOK, results)
}

// WeeklyHoursSummaryItem defines the structure for the weekly hours summary.
type WeeklyHoursSummaryItem struct {
	WeekStartDate string  `json:"week_start_date"`
	BilledHours   float64 `json:"billed_hours"`
	TargetHours   float64 `json:"target_hours"`
}

const numWeeksForSummary = 12

// PortalWeeklyHoursSummaryHandler serves weekly billed vs target hours.
func (a *App) PortalWeeklyHoursSummaryHandler(w http.ResponseWriter, r *http.Request) {
	accountIDVal := r.Context().Value("account_id")
	accountID, ok := accountIDVal.(uint)
	if !ok || accountID == 0 {
		log.Printf("Error: PortalWeeklyHoursSummary - Unauthorized or invalid account_id in context: %v", accountIDVal)
		respondWithError(w, http.StatusUnauthorized, "Unauthorized: Valid Account ID not found in token claims.")
		return
	}

	var projects []cronos.Project
	if err := a.cronosApp.DB.Where("account_id = ?", accountID).Find(&projects).Error; err != nil {
		log.Printf("Error: PortalWeeklyHoursSummary - Failed to retrieve projects for account ID %d: %v", accountID, err)
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve projects.")
		return
	}

	if len(projects) == 0 {
		respondWithJSON(w, http.StatusOK, []WeeklyHoursSummaryItem{}) // No projects, return empty summary
		return
	}

	var projectIDs []uint
	for _, p := range projects {
		projectIDs = append(projectIDs, p.ID)
	}

	var results []WeeklyHoursSummaryItem
	now := time.Now()

	// Determine the Monday of the current week
	weekday := now.Weekday()
	offset := int(time.Monday - weekday)
	if weekday == time.Sunday { // In Go, Sunday is 0, Monday is 1. If today is Sunday, Monday was 6 days ago.
		offset = -6
	} else if weekday == time.Saturday { // If Saturday, Monday was 5 days ago
		offset = -5
	} // For Mon-Fri, offset will be 0 to -4

	currentWeekMonday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, offset)

	for i := 0; i < numWeeksForSummary; i++ {
		// Iterate backwards from the current week
		weekStart := currentWeekMonday.AddDate(0, 0, -7*(numWeeksForSummary-1-i))
		weekEnd := weekStart.AddDate(0, 0, 7) // Exclusive end (start of next Monday)

		var totalBilledHoursThisWeek float64
		var entries []cronos.Entry
		if err := a.cronosApp.DB.Where(
			"project_id IN (?) AND state != ? AND deleted_at IS NULL AND start >= ? AND start < ?",
			projectIDs, cronos.EntryStateVoid.String(), weekStart, weekEnd,
		).Find(&entries).Error; err != nil {
			log.Printf("Error: PortalWeeklyHoursSummary - Failed to retrieve entries for week starting %s: %v", weekStart.Format("2006-01-02"), err)
			// Continue to next week or return error? For robustness, let's try to build partial data.
			// Or, one might choose to respondWithError here.
		}
		for _, entry := range entries {
			if !entry.Start.IsZero() && !entry.End.IsZero() && entry.End.After(entry.Start) {
				duration := entry.End.Sub(entry.Start)
				totalBilledHoursThisWeek += duration.Hours()
			}
		}

		var totalTargetHoursThisWeek float64
		var assignments []cronos.StaffingAssignment // Assuming cronos.StaffingAssignment exists
		// An assignment is active in a week if: assignment.StartDate <= weekEnd (exclusive end of our week i.e. start of next Monday)
		// AND assignment.EndDate >= weekStart (inclusive start of our week i.e. Monday 00:00)
		// Note: weekEnd for query should be the actual end of Sunday for assignments that might end on Sunday.
		actualWeekEnd := weekEnd.Add(-time.Nanosecond) // End of Sunday for precise overlap query

		if err := a.cronosApp.DB.Where(
			"project_id IN (?) AND start_date <= ? AND end_date >= ? AND deleted_at IS NULL",
			projectIDs, actualWeekEnd, weekStart,
		).Find(&assignments).Error; err != nil {
			log.Printf("Error: PortalWeeklyHoursSummary - Failed to retrieve staffing assignments for week starting %s: %v", weekStart.Format("2006-01-02"), err)
		}
		for _, assignment := range assignments {
			totalTargetHoursThisWeek += float64(assignment.Commitment)
		}

		results = append(results, WeeklyHoursSummaryItem{
			WeekStartDate: weekStart.Format("2006-01-02"),
			BilledHours:   totalBilledHoursThisWeek,
			TargetHours:   totalTargetHoursThisWeek,
		})
	}

	respondWithJSON(w, http.StatusOK, results)
}

// WeeklyUtilization represents actual hours worked in a specific week
type WeeklyUtilization struct {
	WeekStart   string  `json:"week_start"`
	ActualHours float64 `json:"actual_hours"`
	Commitment  int     `json:"commitment"`
	Utilization float64 `json:"utilization"` // Percentage (0-100+)
}

// AssignmentWithUtilization combines staffing assignment with weekly utilization data
type AssignmentWithUtilization struct {
	cronos.StaffingAssignment
	Segments          []cronos.CommitmentSegment   `json:"segments"` // Parsed commitment segments
	WeeklyUtilization map[string]WeeklyUtilization `json:"weekly_utilization"`
}

// CapacityDataHandler fetches staffing assignments for capacity management view with utilization data
func (a *App) CapacityDataHandler(w http.ResponseWriter, r *http.Request) {
	var assignments []cronos.StaffingAssignment

	// Fetch all staffing assignments with employee and project preloaded (but NOT entries)
	if err := a.cronosApp.DB.
		Preload("Employee.HeadshotAsset").
		Preload("Project").
		Preload("Project.Account").
		Find(&assignments).Error; err != nil {
		log.Printf("Error fetching staffing assignments: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch capacity data")
		return
	}

	// Fetch aggregated hours by assignment and week using a single SQL query
	// This is MUCH faster than loading all entries into memory
	type WeeklyHours struct {
		StaffingAssignmentID uint
		WeekStart            time.Time
		TotalMinutes         float64
	}

	var weeklyHours []WeeklyHours
	// Use raw SQL for optimal performance - aggregate hours by week for each assignment
	// Note: We calculate week start as Sunday to match frontend logic
	err := a.cronosApp.DB.Raw(`
		SELECT 
			staffing_assignment_id,
			DATE_TRUNC('week', start AT TIME ZONE 'UTC') - INTERVAL '1 day' as week_start,
			SUM(duration_minutes) as total_minutes
		FROM entries
		WHERE staffing_assignment_id IS NOT NULL
		  AND deleted_at IS NULL
		GROUP BY staffing_assignment_id, week_start
		ORDER BY staffing_assignment_id, week_start
	`).Scan(&weeklyHours).Error

	if err != nil {
		log.Printf("Error fetching weekly hours: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch utilization data")
		return
	}

	// Build a map of assignment_id -> week_start -> total_hours for fast lookup
	hoursMap := make(map[uint]map[string]float64)
	for _, wh := range weeklyHours {
		if _, exists := hoursMap[wh.StaffingAssignmentID]; !exists {
			hoursMap[wh.StaffingAssignmentID] = make(map[string]float64)
		}
		weekStartStr := wh.WeekStart.Format("2006-01-02")
		hoursMap[wh.StaffingAssignmentID][weekStartStr] = wh.TotalMinutes / 60.0
	}

	// Build response with utilization data
	response := make([]AssignmentWithUtilization, len(assignments))
	for i, assignment := range assignments {
		weeklyUtil := make(map[string]WeeklyUtilization)

		// Initialize utilization for all weeks in the assignment period with commitments
		assignmentStart := assignment.StartDate
		assignmentEnd := assignment.EndDate

		// Iterate through all weeks in the assignment range, normalized to UTC
		currentWeek := assignmentStart.UTC()
		currentWeek = time.Date(currentWeek.Year(), currentWeek.Month(), currentWeek.Day(), 0, 0, 0, 0, time.UTC)
		currentWeek = currentWeek.AddDate(0, 0, -int(currentWeek.Weekday())) // Start of week (Sunday)

		assignmentEndUTC := assignmentEnd.UTC()
		assignmentEndUTC = time.Date(assignmentEndUTC.Year(), assignmentEndUTC.Month(), assignmentEndUTC.Day(), 0, 0, 0, 0, time.UTC)

		for currentWeek.Before(assignmentEndUTC) || currentWeek.Equal(assignmentEndUTC) {
			weekStartStr := currentWeek.Format("2006-01-02")
			weekCommitment := assignment.GetCommitmentForWeek(currentWeek)

			// Only initialize if there's a commitment for this week
			if weekCommitment > 0 {
				// Get actual hours from the precomputed map
				actualHours := 0.0
				if assignmentHours, exists := hoursMap[assignment.ID]; exists {
					if hours, hasHours := assignmentHours[weekStartStr]; hasHours {
						actualHours = hours
					}
				}

				utilization := 0.0
				if weekCommitment > 0 {
					utilization = (actualHours / float64(weekCommitment)) * 100
				}

				weeklyUtil[weekStartStr] = WeeklyUtilization{
					WeekStart:   weekStartStr,
					Commitment:  weekCommitment,
					ActualHours: actualHours,
					Utilization: utilization,
				}
			}

			// Move to next week
			currentWeek = currentWeek.AddDate(0, 0, 7)
		}

		response[i] = AssignmentWithUtilization{
			StaffingAssignment: assignment,
			Segments:           assignment.GetSegments(),
			WeeklyUtilization:  weeklyUtil,
		}
	}

	respondWithJSON(w, http.StatusOK, response)
}

// CapacityDetailHandler fetches detailed time entries for a specific assignment and week
func (a *App) CapacityDetailHandler(w http.ResponseWriter, r *http.Request) {
	assignmentID := r.URL.Query().Get("assignment_id")
	weekStart := r.URL.Query().Get("week_start")

	if assignmentID == "" || weekStart == "" {
		respondWithError(w, http.StatusBadRequest, "assignment_id and week_start are required")
		return
	}

	// Parse the week start date
	weekStartDate, err := time.Parse("2006-01-02", weekStart)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid week_start format (use YYYY-MM-DD)")
		return
	}

	// Calculate week end (7 days later)
	weekEndDate := weekStartDate.AddDate(0, 0, 7)

	// Fetch entries for this assignment within the week
	var entries []cronos.Entry
	if err := a.cronosApp.DB.
		Where("staffing_assignment_id = ? AND start >= ? AND start < ? AND deleted_at IS NULL", assignmentID, weekStartDate, weekEndDate).
		Order("start ASC").
		Find(&entries).Error; err != nil {
		log.Printf("Error fetching entries for assignment %s, week %s: %v", assignmentID, weekStart, err)
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch entries")
		return
	}

	// Convert entries to a lightweight response format
	type EntryDetail struct {
		ID              uint      `json:"id"`
		Start           time.Time `json:"start"`
		Notes           string    `json:"notes"`
		DurationMinutes float64   `json:"duration_minutes"`
	}

	response := make([]EntryDetail, len(entries))
	for i, entry := range entries {
		response[i] = EntryDetail{
			ID:              entry.ID,
			Start:           entry.Start,
			Notes:           entry.Notes,
			DurationMinutes: entry.DurationMinutes,
		}
	}

	respondWithJSON(w, http.StatusOK, response)
}

// PortalCapacityDataHandler fetches capacity data filtered by the client's account
func (a *App) PortalCapacityDataHandler(w http.ResponseWriter, r *http.Request) {
	// Get account ID from context (set by JwtVerify middleware)
	accountIDVal := r.Context().Value("account_id")
	accountID, ok := accountIDVal.(uint)
	if !ok || accountID == 0 {
		log.Printf("Error: PortalCapacityDataHandler - Unauthorized or invalid account_id in context: %v", accountIDVal)
		respondWithError(w, http.StatusUnauthorized, "Unauthorized: Valid Account ID not found in token claims.")
		return
	}

	var assignments []cronos.StaffingAssignment

	// Fetch staffing assignments for projects belonging to this account only (but NOT entries)
	if err := a.cronosApp.DB.
		Joins("JOIN projects ON projects.id = staffing_assignments.project_id").
		Where("projects.account_id = ?", accountID).
		Preload("Employee").
		Preload("Project").
		Preload("Project.Account").
		Find(&assignments).Error; err != nil {
		log.Printf("Error fetching portal capacity data for account %d: %v", accountID, err)
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch capacity data")
		return
	}

	// Fetch aggregated hours by assignment and week using a single SQL query for this account
	type WeeklyHours struct {
		StaffingAssignmentID uint
		WeekStart            time.Time
		TotalMinutes         float64
	}

	var weeklyHours []WeeklyHours
	// Use raw SQL for optimal performance - aggregate hours by week for assignments in this account
	// Note: We calculate week start as Sunday to match frontend logic
	err := a.cronosApp.DB.Raw(`
		SELECT 
			e.staffing_assignment_id,
			DATE_TRUNC('week', e.start AT TIME ZONE 'UTC') - INTERVAL '1 day' as week_start,
			SUM(e.duration_minutes) as total_minutes
		FROM entries e
		INNER JOIN staffing_assignments sa ON sa.id = e.staffing_assignment_id
		INNER JOIN projects p ON p.id = sa.project_id
		WHERE e.staffing_assignment_id IS NOT NULL
		  AND e.deleted_at IS NULL
		  AND sa.deleted_at IS NULL
		  AND p.account_id = ?
		GROUP BY e.staffing_assignment_id, week_start
		ORDER BY e.staffing_assignment_id, week_start
	`, accountID).Scan(&weeklyHours).Error

	if err != nil {
		log.Printf("Error fetching weekly hours for account %d: %v", accountID, err)
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch utilization data")
		return
	}

	// Build a map of assignment_id -> week_start -> total_hours for fast lookup
	hoursMap := make(map[uint]map[string]float64)
	for _, wh := range weeklyHours {
		if _, exists := hoursMap[wh.StaffingAssignmentID]; !exists {
			hoursMap[wh.StaffingAssignmentID] = make(map[string]float64)
		}
		weekStartStr := wh.WeekStart.Format("2006-01-02")
		hoursMap[wh.StaffingAssignmentID][weekStartStr] = wh.TotalMinutes / 60.0
	}

	// Build response with utilization data
	response := make([]AssignmentWithUtilization, len(assignments))
	for i, assignment := range assignments {
		weeklyUtil := make(map[string]WeeklyUtilization)

		// Initialize utilization for all weeks in the assignment period with commitments
		assignmentStart := assignment.StartDate
		assignmentEnd := assignment.EndDate

		// Iterate through all weeks in the assignment range, normalized to UTC
		currentWeek := assignmentStart.UTC()
		// Normalize to start of day at midnight UTC
		currentWeek = time.Date(currentWeek.Year(), currentWeek.Month(), currentWeek.Day(), 0, 0, 0, 0, time.UTC)
		currentWeek = currentWeek.AddDate(0, 0, -int(currentWeek.Weekday())) // Start of week (Sunday)

		assignmentEndUTC := assignmentEnd.UTC()
		assignmentEndUTC = time.Date(assignmentEndUTC.Year(), assignmentEndUTC.Month(), assignmentEndUTC.Day(), 0, 0, 0, 0, time.UTC)

		for currentWeek.Before(assignmentEndUTC) || currentWeek.Equal(assignmentEndUTC) {
			weekStartStr := currentWeek.Format("2006-01-02")
			weekCommitment := assignment.GetCommitmentForWeek(currentWeek)

			// Only initialize if there's a commitment for this week
			if weekCommitment > 0 {
				// Get actual hours from the precomputed map
				actualHours := 0.0
				if assignmentHours, exists := hoursMap[assignment.ID]; exists {
					if hours, hasHours := assignmentHours[weekStartStr]; hasHours {
						actualHours = hours
					}
				}

				utilization := 0.0
				if weekCommitment > 0 {
					utilization = (actualHours / float64(weekCommitment)) * 100
				}

				weeklyUtil[weekStartStr] = WeeklyUtilization{
					WeekStart:   weekStartStr,
					Commitment:  weekCommitment,
					ActualHours: actualHours,
					Utilization: utilization,
				}
			}

			// Move to next week
			currentWeek = currentWeek.AddDate(0, 0, 7)
		}

		response[i] = AssignmentWithUtilization{
			StaffingAssignment: assignment,
			Segments:           assignment.GetSegments(),
			WeeklyUtilization:  weeklyUtil,
		}
	}

	respondWithJSON(w, http.StatusOK, response)
}
