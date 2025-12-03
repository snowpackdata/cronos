package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/snowpackdata/cronos"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// backfillCommitmentSchedules converts existing simple staffing assignments
// to the new segment-based format for backward compatibility
func backfillCommitmentSchedules(db *gorm.DB, dryRun bool) error {
	var assignments []cronos.StaffingAssignment

	// Find all assignments that don't have a commitment schedule yet
	if err := db.Preload("Employee").Preload("Project").Where("commitment_schedule IS NULL OR commitment_schedule = ''").Find(&assignments).Error; err != nil {
		return fmt.Errorf("failed to fetch assignments: %w", err)
	}

	log.Printf("Found %d assignments to backfill", len(assignments))

	if dryRun {
		log.Println("===== DRY RUN MODE - NO CHANGES WILL BE MADE =====")
		log.Println("")
	}

	for i, assignment := range assignments {
		// Create a simple single-segment schedule from existing fields
		schedule := cronos.CommitmentSchedule{
			Segments: []cronos.CommitmentSegment{
				{
					StartDate:  assignment.StartDate.Format("2006-01-02"),
					EndDate:    assignment.EndDate.Format("2006-01-02"),
					Commitment: assignment.Commitment,
				},
			},
		}

		// Serialize to JSON
		scheduleJSON, err := json.Marshal(schedule)
		if err != nil {
			log.Printf("Warning: Failed to marshal schedule for assignment %d: %v", assignment.ID, err)
			continue
		}

		if dryRun {
			// Show what would be changed
			employeeName := "Unknown"
			if assignment.Employee.ID > 0 {
				employeeName = fmt.Sprintf("%s %s", assignment.Employee.FirstName, assignment.Employee.LastName)
			}
			projectName := "Unknown"
			if assignment.Project.ID > 0 {
				projectName = assignment.Project.Name
			}

			log.Printf("[DRY RUN] Would update assignment %d:", assignment.ID)
			log.Printf("  Employee: %s", employeeName)
			log.Printf("  Project: %s", projectName)
			log.Printf("  Current: %dh/week from %s to %s",
				assignment.Commitment,
				assignment.StartDate.Format("2006-01-02"),
				assignment.EndDate.Format("2006-01-02"))
			log.Printf("  New Schedule: %s", string(scheduleJSON))
			log.Println("")
		} else {
			// Update the assignment
			if err := db.Model(&assignment).Update("commitment_schedule", string(scheduleJSON)).Error; err != nil {
				log.Printf("Warning: Failed to update assignment %d: %v", assignment.ID, err)
				continue
			}

			if (i+1)%10 == 0 {
				log.Printf("Progress: %d/%d assignments backfilled", i+1, len(assignments))
			}
		}
	}

	if dryRun {
		log.Println("===== DRY RUN COMPLETE - NO CHANGES WERE MADE =====")
		log.Printf("Would have backfilled %d assignments", len(assignments))
	} else {
		log.Printf("Successfully backfilled %d assignments", len(assignments))
	}
	return nil
}

func main() {
	var db *gorm.DB
	var err error

	// Check for dry run mode
	dryRun := os.Getenv("DRY_RUN") == "true"

	// Get connection details from environment
	user := os.Getenv("CLOUD_SQL_USERNAME")
	password := os.Getenv("CLOUD_SQL_PASSWORD")
	dbHost := os.Getenv("CLOUD_SQL_CONNECTION_NAME")
	databaseName := os.Getenv("CLOUD_SQL_DATABASE_NAME")

	// Check if we're using Cloud SQL or local SQLite
	if user != "" && password != "" && databaseName != "" {
		// PostgreSQL connection (production Cloud SQL)
		if dbHost == "" {
			log.Fatal("Missing CLOUD_SQL_CONNECTION_NAME for Cloud SQL connection")
		}

		var dbURI string

		// Check if we're running on App Engine/GCE (unix socket) or via proxy
		if os.Getenv("GAE_ENV") != "" || os.Getenv("K_SERVICE") != "" {
			// Running in production App Engine or Cloud Run - use unix socket
			socketPath := "/cloudsql/" + dbHost
			dbURI = fmt.Sprintf("user=%s password=%s database=%s host=%s", user, password, databaseName, socketPath)
			log.Println("Connecting to Cloud SQL via unix socket (production)...")
		} else {
			// Running locally via cloud-sql-proxy
			port := os.Getenv("DB_PORT")
			if port == "" {
				port = "5432" // Default PostgreSQL port
			}
			dbURI = fmt.Sprintf("host=127.0.0.1 user=%s password=%s port=%s database=%s sslmode=disable TimeZone=UTC", user, password, port, databaseName)
			log.Println("Connecting to Cloud SQL via cloud-sql-proxy...")
		}

		db, err = gorm.Open(postgres.Open(dbURI), &gorm.Config{})

	} else {
		// Local SQLite mode (default)
		dbPath := os.Getenv("DB_PATH")
		if dbPath == "" {
			dbPath = "cronos.db"
		}

		log.Printf("Connecting to local SQLite database: %s", dbPath)
		db, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	}

	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("Connected to database successfully")

	// Run auto-migration to add the new field
	log.Println("Running auto-migration to add commitment_schedule field...")
	if err := db.AutoMigrate(&cronos.StaffingAssignment{}); err != nil {
		log.Fatalf("Failed to run migration: %v", err)
	}
	log.Println("Migration completed successfully")

	// Wait a moment for the migration to complete
	time.Sleep(time.Second)

	// Perform backfill
	log.Println("Starting backfill process...")
	if err := backfillCommitmentSchedules(db, dryRun); err != nil {
		log.Fatalf("Backfill failed: %v", err)
	}

	if dryRun {
		log.Println("")
		log.Println("To apply these changes, run again without DRY_RUN=true")
	} else {
		log.Println("Backfill completed successfully!")
	}
}
