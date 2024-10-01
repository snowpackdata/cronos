package cronos

import (
	"cloud.google.com/go/storage"
	"context"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"time"
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
			ParameterizedQueries:      true,        // Don't include params in the SQL log
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
}

// InitializeLocal allows us to initialize our application and connect to the cloud database
func (a *App) InitializeLocal(user, password, connection, database string) {
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second, // Slow SQL threshold
			LogLevel:                  logger.Info, // Log level
			IgnoreRecordNotFoundError: true,        // Ignore ErrRecordNotFound error for logger
			ParameterizedQueries:      true,        // Don't include params in the SQL log
			Colorful:                  true,        // Disable color
		},
	)
	port := "3306"
	dbURI := fmt.Sprintf("host=127.0.0.1 user=%s password=%s port=%s database=%s sslmode=disable TimeZone=UTC", user, password, port, database)
	db, err := gorm.Open(postgres.Open(dbURI), &gorm.Config{Logger: newLogger})

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
			ParameterizedQueries:      true,        // Don't include params in the SQL log
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

// Calling the Migrate
func (a *App) Migrate() {
	// Migrate the schema
	_ = a.DB.AutoMigrate(&User{})
	_ = a.DB.AutoMigrate(&Employee{})
	_ = a.DB.AutoMigrate(&Client{})
	_ = a.DB.AutoMigrate(&Account{})
	_ = a.DB.AutoMigrate(&Rate{})
	_ = a.DB.AutoMigrate(&Project{})
	_ = a.DB.AutoMigrate(&Entry{})
	_ = a.DB.AutoMigrate(&BillingCode{})
	_ = a.DB.AutoMigrate(&Journal{})
	_ = a.DB.AutoMigrate(&Invoice{})
	_ = a.DB.AutoMigrate(&Adjustment{})
	_ = a.DB.AutoMigrate(&Survey{})
	_ = a.DB.AutoMigrate(&SurveyResponse{})
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
