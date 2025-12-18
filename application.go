package cronos

import (
	"context"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// App is used to initialize a database and hold our handler functions
type App struct {
	DB      *gorm.DB
	Project string
	Bucket  string
}

// InitializeSQLite allows us to initialize our application and connect to the local database
// This handler will hold on to our database operations throughout the lifetime of the application
func (a *App) InitializeSQLite() {
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second, // Slow SQL threshold
			LogLevel:                  logger.Info, // Log level
			IgnoreRecordNotFoundError: true,        // Ignore ErrRecordNotFound error for logger
			ParameterizedQueries:      false,       // Show params in SQL log for debugging
			Colorful:                  true,        // Disable color
		},
	)

	db, err := gorm.Open(sqlite.Open("cronos.db"), &gorm.Config{Logger: newLogger})
	if err != nil {
		panic("failed to connect database")
	}
	a.DB = db
	a.Bucket = os.Getenv("GCS_BUCKET")
	a.Project = os.Getenv("GCP_PROJECT")
	//a.SeedDatabase()

}

// InitializeLocal allows us to initialize our application and connect to the cloud database
func (a *App) InitializeLocal(user, password, connection, database string) {
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second, // Slow SQL threshold
			LogLevel:                  logger.Info, // Log level
			IgnoreRecordNotFoundError: true,        // Ignore ErrRecordNotFound error for logger
			ParameterizedQueries:      false,       // Show params in SQL log for debugging
			Colorful:                  true,        // Disable color
		},
	)
	// Check for custom port from environment, default to 5432 for Postgres
	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "5432"
	}
	dbURI := fmt.Sprintf("host=127.0.0.1 user=%s password=%s port=%s database=%s sslmode=disable TimeZone=UTC", user, password, port, database)
	db, err := gorm.Open(postgres.Open(dbURI), &gorm.Config{
		Logger:                                   newLogger,
		DisableForeignKeyConstraintWhenMigrating: true,
	})

	if err != nil {
		fmt.Println(err)
	}
	a.DB = db
	a.Bucket = os.Getenv("GCS_BUCKET")
	a.Project = os.Getenv("GCP_PROJECT")
}

// InitializeCloud allows us to initialize a connection to the cloud database
// while on google app engine
func (a *App) InitializeCloud(dbURI string) {
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second, // Slow SQL threshold
			LogLevel:                  logger.Info, // Log level
			IgnoreRecordNotFoundError: true,        // Ignore ErrRecordNotFound error for logger
			ParameterizedQueries:      false,       // Show params in SQL log for debugging
			Colorful:                  true,        // Disable color
		},
	)
	db, err := gorm.Open(postgres.Open(dbURI), &gorm.Config{
		Logger:                                   newLogger,
		DisableForeignKeyConstraintWhenMigrating: true})

	if err != nil {
		fmt.Println(err)
	}
	a.DB = db
	a.Bucket = os.Getenv("GCS_BUCKET")
	a.Project = os.Getenv("GCP_PROJECT")
}

func (a *App) InitializeStorageClient(projectID, bucketName string) *storage.Client {
	ctx := context.Background()
	storageClient, err := storage.NewClient(ctx)
	if err != nil {
		fmt.Println(err)
	}
	return storageClient
}

// MigrateTenants migrates only the Tenant and GoogleAuth tables (fast, targeted migration)
func (a *App) MigrateTenants() error {
	log.Println("Running targeted Tenant and GoogleAuth migration...")
	if err := a.DB.AutoMigrate(&Tenant{}, &GoogleAuth{}); err != nil {
		log.Printf("MigrateTenants error: %v", err)
		return err
	}
	log.Println("Tenant and GoogleAuth migration completed successfully")
	return nil
}

// MigrateModel
func (a *App) MigrateModel(model interface{}) {
	err := a.DB.AutoMigrate(model)
	if err != nil {
		log.Printf("AutoMigrate error for %T: %v", model, err)
	}
	log.Printf("AutoMigrate completed for %T", model)
}

// Calling the Migrate
func (a *App) Migrate() {
	// Migrate each model individually to handle constraint errors gracefully
	models := []interface{}{
		// Level 0: No foreign keys
		&Tenant{},
		&ChartOfAccount{},
		&ExpenseCategory{},
		&ExpenseTag{},

		// Level 1: Only references Tenant
		&Subaccount{},
		&Account{},
		&Client{},

		// Level 2: References Tenant + Account/Client/User
		&User{},
		&Asset{},
		&Employee{},
		&GoogleAuth{},
		&Commission{},
		&Rate{},
		&BillingCode{},

		// Level 3: References Employee (ae_id, sdr_id)
		&Project{},

		&Entry{},
		&Journal{},
		&OfflineJournal{},
		&Invoice{},
		&Bill{},
		&StaffingAssignment{},
		&Expense{},
		&RecurringEntry{},

		// Level 5: Junction tables and line items
		&InvoiceLineItem{},
		&BillLineItem{},
		&Adjustment{},
		&ExpenseTagAssignment{},
		&RecurringBillLineItem{},
	}

	for _, model := range models {
		err := a.DB.AutoMigrate(model)
		if err != nil {
			// Ignore "constraint does not exist" errors from old database
			if !strings.Contains(err.Error(), "does not exist") {
				log.Printf("AutoMigrate error for %T: %v", model, err)
			}
		}
	}
}

// GenerateSecureFilename generates a hash from the filename of an invoice
// to be used as a unique identifier in our URL
func GenerateSecureFilename(filename string) string {
	// We add a timestamp at generation so that we can ensure that the filename is unique
	// and clients cannot use their own filenames to back into other invoices
	currentTimeStampString := time.Now().String()
	filenameBytes := []byte(filename + currentTimeStampString)
	hasher := sha1.New()
	hasher.Write(filenameBytes)
	sha := base64.URLEncoding.EncodeToString(hasher.Sum(nil))
	return sha
}

// TenantScope returns a GORM scope that filters by tenant_id
func TenantScope(tenantID uint) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("tenant_id = ?", tenantID)
	}
}

// WithTenant creates a database session scoped to a tenant from context
func (a *App) WithTenant(ctx context.Context) *gorm.DB {
	tenantID := GetTenantIDFromContext(ctx)
	if tenantID == 0 {
		log.Printf("WARNING: No tenant in context, returning unscoped DB")
		return a.DB
	}
	return a.DB.Scopes(TenantScope(tenantID))
}

// GetTenantIDFromContext extracts tenant ID from context
func GetTenantIDFromContext(ctx context.Context) uint {
	type contextKey string
	const TenantContextKey contextKey = "tenant"

	tenant, ok := ctx.Value(TenantContextKey).(*Tenant)
	if !ok || tenant == nil {
		return 0
	}
	return tenant.ID
}
