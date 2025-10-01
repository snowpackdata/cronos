package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/snowpackdata/cronos"
)

// backfillStaffingAssignments finds entries without staffing_assignment_id and populates them
func backfillStaffingAssignments(app *cronos.App, dryRun bool) error {
	log.Println("Starting staffing assignment backfill...")

	// Get all entries without a staffing assignment
	var entries []cronos.Entry
	err := app.DB.Where("staffing_assignment_id IS NULL").
		Preload("Project").
		Find(&entries).Error
	if err != nil {
		return fmt.Errorf("failed to fetch entries: %w", err)
	}

	log.Printf("Found %d entries without staffing assignments", len(entries))

	stats := struct {
		totalProcessed int
		assigned       int
		noMatch        int
		ambiguous      int
		errors         int
	}{}

	// Process each entry
	for _, entry := range entries {
		stats.totalProcessed++

		// Find eligible staffing assignments for this entry
		var assignments []cronos.StaffingAssignment
		err := app.DB.Where("project_id = ? AND start_date <= ? AND end_date >= ?",
			entry.ProjectID,
			entry.Start,
			entry.Start).
			Find(&assignments).Error

		if err != nil {
			log.Printf("ERROR: Entry ID %d - failed to query assignments: %v", entry.ID, err)
			stats.errors++
			continue
		}

		if len(assignments) == 0 {
			log.Printf("NO MATCH: Entry ID %d (Project ID %d, Start: %s) - no eligible staffing assignment found",
				entry.ID, entry.ProjectID, entry.Start.Format("2006-01-02"))
			stats.noMatch++
			continue
		}

		// We have at least one match - select the assignment with the earliest start date
		var selectedAssignment *cronos.StaffingAssignment
		for i := range assignments {
			if selectedAssignment == nil || assignments[i].StartDate.Before(selectedAssignment.StartDate) {
				selectedAssignment = &assignments[i]
			}
		}

		if len(assignments) == 1 {
			log.Printf("MATCH: Entry ID %d -> Staffing Assignment ID %d (Project: %d, Period: %s to %s)",
				entry.ID,
				selectedAssignment.ID,
				selectedAssignment.ProjectID,
				selectedAssignment.StartDate.Format("2006-01-02"),
				selectedAssignment.EndDate.Format("2006-01-02"))
			stats.assigned++
		} else {
			// Multiple matches - using earliest start date to disambiguate
			log.Printf("AMBIGUOUS (using earliest): Entry ID %d has %d matching assignments, selected Assignment ID %d (earliest start: %s):",
				entry.ID, len(assignments), selectedAssignment.ID, selectedAssignment.StartDate.Format("2006-01-02"))
			for _, a := range assignments {
				marker := ""
				if a.ID == selectedAssignment.ID {
					marker = " <- SELECTED"
				}
				log.Printf("  - Assignment ID %d: %s to %s%s",
					a.ID,
					a.StartDate.Format("2006-01-02"),
					a.EndDate.Format("2006-01-02"),
					marker)
			}
			stats.ambiguous++
			stats.assigned++
		}

		// Assign the selected staffing assignment
		if !dryRun {
			entry.StaffingAssignmentID = &selectedAssignment.ID
			if err := app.DB.Save(&entry).Error; err != nil {
				log.Printf("ERROR: Failed to update entry ID %d: %v", entry.ID, err)
				stats.errors++
				continue
			}
		}
	}

	// Print summary
	log.Println("\n" + strings.Repeat("=", 60))
	log.Println("BACKFILL SUMMARY")
	log.Println(strings.Repeat("=", 60))
	log.Printf("Total entries processed:        %d", stats.totalProcessed)
	log.Printf("Successfully assigned:          %d", stats.assigned)
	log.Printf("  - Exact matches:              %d", stats.assigned-stats.ambiguous)
	log.Printf("  - Disambiguated (earliest):   %d", stats.ambiguous)
	log.Printf("No matching assignment:         %d", stats.noMatch)
	log.Printf("Errors:                         %d", stats.errors)
	log.Println(strings.Repeat("=", 60))

	if dryRun {
		log.Println("\nDRY RUN MODE - No changes were made to the database")
		log.Println("Run with -execute flag to apply changes")
	} else {
		log.Println("\nBackfill complete!")
	}

	return nil
}

func main() {
	dbPath := flag.String("db", "cronos.db", "Path to database file for SQLite")
	dbType := flag.String("type", "sqlite", "Database type: sqlite, postgres, or cloud")
	execute := flag.Bool("execute", false, "Execute the backfill (default is dry-run mode)")
	dbUser := flag.String("user", "", "Database user (for postgres)")
	dbPassword := flag.String("password", "", "Database password (for postgres)")
	dbName := flag.String("dbname", "", "Database name (for postgres)")
	dbURI := flag.String("uri", "", "Full database URI (for cloud/postgres)")
	flag.Parse()

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Initialize the cronos app
	app := &cronos.App{}

	switch *dbType {
	case "sqlite":
		log.Printf("Connecting to SQLite database: %s", *dbPath)
		app.InitializeSQLite()

	case "postgres":
		if *dbUser == "" || *dbPassword == "" || *dbName == "" {
			log.Fatal("For postgres, you must provide -user, -password, and -dbname flags")
		}
		log.Printf("Connecting to PostgreSQL database: %s", *dbName)
		app.InitializeLocal(*dbUser, *dbPassword, "", *dbName)

	case "cloud":
		if *dbURI == "" {
			log.Fatal("For cloud database, you must provide -uri flag")
		}
		log.Printf("Connecting to cloud database")
		app.InitializeCloud(*dbURI)

	default:
		log.Fatalf("Unknown database type: %s (must be sqlite, postgres, or cloud)", *dbType)
	}

	dryRun := !*execute
	if dryRun {
		log.Println("\n*** DRY RUN MODE - No changes will be made ***")
	} else {
		log.Println("\n*** EXECUTE MODE - Changes will be saved to database ***")
	}

	// Run the backfill
	if err := backfillStaffingAssignments(app, dryRun); err != nil {
		log.Fatalf("Backfill failed: %v", err)
	}
}
