package main

import (
	"context"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path"
	"strings"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/snowpackdata/cronos"
)

//go:embed static/admin
var adminAssets embed.FS

//go:embed static/portal
var portalAssets embed.FS

//go:embed templates
var templates embed.FS

//go:embed assets
var publicAssets embed.FS

//go:embed branding
var brandingAssets embed.FS

// App holds our information for accessing cronos application and methods across modules
type App struct {
	cronosApp *cronos.App
	logger    *log.Logger
	GitHash   string
	DevToken  string // JWT token for development environment
}

// createFileServer creates a file server for embedded assets with proper MIME types
func createFileServer(embeddedFS embed.FS, fsRoot string) http.Handler {
	subFS, err := fs.Sub(embeddedFS, fsRoot)
	if err != nil {
		log.Fatal("Failed to create sub-filesystem:", err)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Serving embedded file: %s from root %s", r.URL.Path, fsRoot)

		ext := path.Ext(r.URL.Path)
		switch ext {
		case ".css":
			w.Header().Set("Content-Type", "text/css; charset=utf-8")
		case ".js":
			w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
		case ".svg":
			w.Header().Set("Content-Type", "image/svg+xml")
		case ".png":
			w.Header().Set("Content-Type", "image/png")
		case ".jpg", ".jpeg":
			w.Header().Set("Content-Type", "image/jpeg")
		case ".woff":
			w.Header().Set("Content-Type", "font/woff")
		case ".woff2":
			w.Header().Set("Content-Type", "font/woff2")
		case ".ttf":
			w.Header().Set("Content-Type", "font/ttf")
		}

		http.FileServer(http.FS(subFS)).ServeHTTP(w, r)
	})
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	environment := os.Getenv("ENVIRONMENT")
	if environment == "" {
		environment = "local"
	}

	log.Printf("Starting Cronos server in %s mode", strings.ToUpper(environment))

	var wait time.Duration

	// Initialize the cronos application
	user := os.Getenv("CLOUD_SQL_USERNAME")
	password := os.Getenv("CLOUD_SQL_PASSWORD")
	dbHost := os.Getenv("CLOUD_SQL_CONNECTION_NAME")
	databaseName := os.Getenv("CLOUD_SQL_DATABASE_NAME")
	gitHash := os.Getenv("GIT_HASH")
	if gitHash == "" {
		gitHash = "dev"
	}

	socketPath := "/cloudsql/" + dbHost
	cronosApp := cronos.App{}
	dbURI := fmt.Sprintf("user=%s password=%s database=%s host=%s", user, password, databaseName, socketPath)
	fmt.Println(dbURI)

	// Establish database connection based on environment
	if os.Getenv("ENVIRONMENT") == "production" {
		log.Println("Initializing in PRODUCTION mode")
		cronosApp.InitializeCloud(dbURI)
	} else if os.Getenv("ENVIRONMENT") == "development" {
		log.Println("Initializing in DEVELOPMENT mode")
		dbPort := os.Getenv("DB_PORT")
		if dbPort == "" {
			dbPort = "5432"
		}
		log.Printf("Connecting to Cloud SQL via proxy at localhost:%s", dbPort)
		log.Printf("Database: %s, User: %s", databaseName, user)
		cronosApp.InitializeLocal(user, password, dbHost, databaseName)
		// cronosApp.MigrateModel(&cronos.Account{}) // Fast targeted migration for Tenant model only
	} else {
		log.Println("Initializing in LOCAL mode with SQLite")
		cronosApp.InitializeSQLite()
		cronosApp.Migrate()
		cronosApp.SeedDatabase()
	}

	a := &App{
		cronosApp: &cronosApp,
		logger:    log.New(os.Stdout, "http: ", log.LstdFlags),
		GitHash:   gitHash,
	}

	// Log credentials for local development
	env := os.Getenv("ENVIRONMENT")
	if env == "local" {
		log.Println("--- LOCAL DEVELOPMENT CREDENTIALS ---")
		log.Println("Dev User Email:    dev@example.com")
		log.Println("Dev User Password: devpassword")
		log.Println("Sample Client User Email:    client@example.com")
		log.Println("Sample Client User Password: password")
		log.Println("---------------------------------------")
	}

	r := mux.NewRouter()

	// Add middleware
	r.Use(a.AppContextMiddleware)
	r.Use(ParseTokenAndSetUserContext)

	// Static file servers for embedded admin and portal assets
	adminStatic := r.PathPrefix("/admin/assets/").Subrouter()
	adminStatic.PathPrefix("/").Handler(http.StripPrefix("/admin/assets/", createFileServer(adminAssets, "static/admin/assets")))

	portalStatic := r.PathPrefix("/portal/assets/").Subrouter()
	portalStatic.PathPrefix("/").Handler(http.StripPrefix("/portal/assets/", createFileServer(portalAssets, "static/portal/assets")))

	// Static file server for public assets (CSS, images, JS for landing pages)
	publicStatic := r.PathPrefix("/assets/").Subrouter()
	publicStatic.PathPrefix("/").Handler(http.StripPrefix("/assets/", createFileServer(publicAssets, "assets")))

	// Static file server for branding assets (logos)
	brandingStatic := r.PathPrefix("/branding/").Subrouter()
	brandingStatic.PathPrefix("/").Handler(http.StripPrefix("/branding/", createFileServer(brandingAssets, "branding")))

	// Fallback handler for admin assets
	r.PathPrefix("/admin/assets/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestedPath := r.URL.Path
		log.Printf("Direct file request: %s", requestedPath)

		if strings.HasSuffix(requestedPath, ".css") {
			w.Header().Set("Content-Type", "text/css; charset=utf-8")

			if strings.Contains(requestedPath, "index-") && strings.HasSuffix(requestedPath, ".css") {
				entries, err := adminAssets.ReadDir("static/admin/assets")
				if err != nil {
					http.Error(w, "Failed to read directory", http.StatusInternalServerError)
					return
				}

				for _, entry := range entries {
					if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".css") && strings.HasPrefix(entry.Name(), "index-") {
						log.Printf("Found substitute CSS file: %s for requested: %s", entry.Name(), requestedPath)

						cssFile, err := adminAssets.Open("static/admin/assets/" + entry.Name())
						if err != nil {
							continue
						}
						defer cssFile.Close()

						stat, err := cssFile.Stat()
						if err != nil {
							continue
						}

						http.ServeContent(w, r, stat.Name(), stat.ModTime(), cssFile.(io.ReadSeeker))
						return
					}
				}
			}
		}

		path := strings.TrimPrefix(requestedPath, "/admin/assets/")
		file, err := adminAssets.Open("static/admin/assets/" + path)
		if err != nil {
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}
		defer file.Close()

		stat, err := file.Stat()
		if err != nil {
			http.Error(w, "Failed to get file info", http.StatusInternalServerError)
			return
		}

		if stat.IsDir() {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}

		http.ServeContent(w, r, stat.Name(), stat.ModTime(), file.(io.ReadSeeker))
	})

	// Admin API routes - require tenant, valid token AND IsStaff == true
	adminApi := r.PathPrefix("/api").Subrouter()
	adminApi.Use(TenantMiddleware(&cronosApp)) // Apply tenant middleware first
	adminApi.Use(JwtVerify)
	adminApi.Use(RequireStaff)

	// Tenant information endpoint
	adminApi.HandleFunc("/tenant", a.GetTenantHandler).Methods("GET")
	adminApi.HandleFunc("/tenant", a.UpdateTenantHandler).Methods("PUT")

	// Invoice routes
	adminApi.HandleFunc("/invoices/draft", a.DraftInvoiceListHandler).Methods("GET")
	adminApi.HandleFunc("/invoices/accepted", a.InvoiceListHandler).Methods("GET")
	adminApi.HandleFunc("/invoices/{id:[0-9]+}/{state:(?:approve)|(?:send)|(?:paid)|(?:void)|(?:regenerate_pdf)}", a.InvoiceStateHandler).Methods("POST")
	adminApi.HandleFunc("/invoices/{id:[0-9]+}/send_email", a.SendInvoiceEmailHandler).Methods("POST")

	// Project routes
	adminApi.HandleFunc("/projects", a.ProjectsListHandler).Methods("GET")
	adminApi.HandleFunc("/projects/{id:[0-9]+}", a.ProjectHandler).Methods("GET", "PUT", "POST", "DELETE")
	adminApi.HandleFunc("/projects/{id:[0-9]+}/analytics", a.ProjectAnalyticsHandler).Methods("GET")
	adminApi.HandleFunc("/projects/{id:[0-9]+}/backfill", a.BackfillProjectInvoicesHandler).Methods("POST")
	adminApi.HandleFunc("/projects/{id:[0-9]+}/assets", a.ProjectAssetsCreateHandler).Methods("POST")
	adminApi.HandleFunc("/projects/{id:[0-9]+}/assets/{assetID}", a.ProjectAssetDeleteHandler).Methods("DELETE")
	adminApi.HandleFunc("/projects/{id:[0-9]+}/billing_codes", a.ProjectBillingCodesListHandler).Methods("GET")

	// Entry routes
	adminApi.HandleFunc("/entries", a.EntriesListHandler).Methods("GET")
	adminApi.HandleFunc("/entries/{id:[0-9]+}", a.EntryHandler).Methods("GET", "PUT", "POST", "DELETE")
	adminApi.HandleFunc("/entries/state/{id:[0-9]+}/{state:(?:void)|(?:draft)|(?:approve)|(?:reject)|(?:exclude)}", a.EntryStateHandler).Methods("POST")

	// Staff routes
	adminApi.HandleFunc("/staff", a.StaffListHandler).Methods("GET")
	adminApi.HandleFunc("/staff/{id:[0-9]+}", a.StaffHandler).Methods("GET", "PUT", "POST", "DELETE")

	// Account routes
	adminApi.HandleFunc("/accounts", a.AccountsListHandler).Methods("GET")
	adminApi.HandleFunc("/accounts/{id:[0-9]+}", a.AccountHandler).Methods("GET", "PUT", "POST", "DELETE")
	adminApi.HandleFunc("/accounts/{id:[0-9]+}/invite/{user_id:[0-9]+}", a.InviteUserHandler).Methods("POST")
	adminApi.HandleFunc("/accounts/{id:[0-9]+}/assets", a.AccountAssetsCreateHandler).Methods("POST")

	// Rate routes
	adminApi.HandleFunc("/rates", a.RatesListHandler).Methods("GET")
	adminApi.HandleFunc("/rates/{id:[0-9]+}", a.RateHandler).Methods("GET", "PUT", "POST", "DELETE")

	// Billing code routes
	adminApi.HandleFunc("/billing_codes", a.BillingCodesListHandler).Methods("GET")
	adminApi.HandleFunc("/billing_codes/{id:[0-9]+}", a.BillingCodeHandler).Methods("GET", "PUT", "POST", "DELETE")
	adminApi.HandleFunc("/active_billing_codes", a.ActiveBillingCodesListHandler).Methods("GET")

	// Adjustment routes
	adminApi.HandleFunc("/adjustments/{id:[0-9]+}", a.AdjustmentHandler).Methods("GET", "PUT", "POST", "DELETE")
	adminApi.HandleFunc("/adjustments/state/{id:[0-9]+}/{state:(?:void)|(?:draft)|(?:approve)}", a.AdjustmentStateHandler).Methods("POST")

	// Capacity routes
	adminApi.HandleFunc("/capacity", a.CapacityDataHandler).Methods("GET")
	adminApi.HandleFunc("/capacity/detail", a.CapacityDetailHandler).Methods("GET")

	// Bill routes
	adminApi.HandleFunc("/bills", a.BillListHandler).Methods("GET")
	adminApi.HandleFunc("/bills/{id:[0-9]+}", a.BillHandler).Methods("GET")
	adminApi.HandleFunc("/bills/{id:[0-9]+}/regenerate", a.RegenerateBillHandler).Methods("POST")
	adminApi.HandleFunc("/bills/{id:[0-9]+}/{state:(?:accept)|(?:paid)|(?:void)}", a.BillStateHandler).Methods("POST")

	// Project assignment routes
	adminApi.HandleFunc("/project_assignments/{id:[0-9]+}", a.ProjectAssignmentHandler).Methods("GET", "PUT", "POST", "DELETE")

	// Asset routes
	adminApi.HandleFunc("/assets/{id:[0-9]+}/refresh-url", a.RefreshAssetURLHandler).Methods("POST")
	adminApi.HandleFunc("/assets/{id:[0-9]+}/download", a.AssetDownloadHandler).Methods("GET")

	// Journal routes
	adminApi.HandleFunc("/cronos/journals", a.JournalsListHandler).Methods("GET")
	adminApi.HandleFunc("/cronos/journals/manual", a.ManualJournalEntryHandler).Methods("POST")
	adminApi.HandleFunc("/cronos/accounts/balances", a.AccountBalancesHandler).Methods("GET")
	adminApi.HandleFunc("/cronos/ledger/combined", a.CombinedGeneralLedgerHandler).Methods("GET")
	adminApi.HandleFunc("/cronos/ledger/reconciliation", a.ReconciliationReportHandler).Methods("GET")
	adminApi.HandleFunc("/cronos/ledger/account-summary", a.AccountSummaryHandler).Methods("GET")
	adminApi.HandleFunc("/cronos/ledger/trial-balance", a.TrialBalanceHandler).Methods("GET")

	// Chart of Accounts routes
	adminApi.HandleFunc("/cronos/chart-of-accounts", a.ListChartOfAccountsHandler).Methods("GET")
	adminApi.HandleFunc("/cronos/chart-of-accounts", a.CreateChartOfAccountHandler).Methods("POST")
	adminApi.HandleFunc("/cronos/chart-of-accounts/{code}", a.UpdateChartOfAccountHandler).Methods("PUT")
	adminApi.HandleFunc("/cronos/chart-of-accounts/{code}", a.DeactivateChartOfAccountHandler).Methods("DELETE")
	adminApi.HandleFunc("/cronos/chart-of-accounts/seed", a.SeedSystemAccountsHandler).Methods("POST")

	// Subaccounts routes
	adminApi.HandleFunc("/cronos/subaccounts", a.ListSubaccountsHandler).Methods("GET")
	adminApi.HandleFunc("/cronos/subaccounts", a.CreateSubaccountHandler).Methods("POST")
	adminApi.HandleFunc("/cronos/subaccounts/{code}", a.UpdateSubaccountHandler).Methods("PUT")
	adminApi.HandleFunc("/cronos/subaccounts/{code}", a.DeactivateSubaccountHandler).Methods("DELETE")

	// General Ledger Adjustments
	adminApi.HandleFunc("/cronos/journals/{id:[0-9]+}/reverse", a.ReverseJournalEntryHandler).Methods("POST")

	// Offline Journals (CSV import)
	adminApi.HandleFunc("/cronos/offline-journals/upload-csv", a.UploadCSVHandler).Methods("POST")
	adminApi.HandleFunc("/cronos/offline-journals/transactions", a.GetOfflineJournalTransactionsHandler).Methods("GET")
	adminApi.HandleFunc("/cronos/offline-journals/categorize", a.CategorizeCSVTransactionHandler).Methods("POST")
	adminApi.HandleFunc("/cronos/offline-journals/approve-transaction", a.ApproveTransactionPairHandler).Methods("POST")
	adminApi.HandleFunc("/cronos/offline-journals/suggest-categorization", a.GetSuggestedCategorizationsHandler).Methods("GET")
	adminApi.HandleFunc("/cronos/offline-journals", a.OfflineJournalsListHandler).Methods("GET")
	adminApi.HandleFunc("/cronos/offline-journals/{id:[0-9]+}", a.UpdateOfflineJournalStatusHandler).Methods("PUT")
	adminApi.HandleFunc("/cronos/offline-journals/{id:[0-9]+}/edit", a.EditOfflineJournalHandler).Methods("PUT")
	adminApi.HandleFunc("/cronos/offline-journals/{id:[0-9]+}", a.DeleteOfflineJournalHandler).Methods("DELETE")
	adminApi.HandleFunc("/cronos/offline-journals/post-to-gl", a.PostOfflineJournalsToGLHandler).Methods("POST")
	adminApi.HandleFunc("/cronos/offline-journals/bulk-update", a.BulkUpdateOfflineJournalStatusHandler).Methods("POST")

	// Expenses routes
	adminApi.HandleFunc("/expenses", a.GetExpensesHandler).Methods("GET")
	adminApi.HandleFunc("/expenses/review", a.GetExpensesForReviewHandler).Methods("GET")
	adminApi.HandleFunc("/expenses", a.CreateExpenseHandler).Methods("POST")
	adminApi.HandleFunc("/expenses/{id:[0-9]+}", a.UpdateExpenseHandler).Methods("PUT")
	adminApi.HandleFunc("/expenses/{id:[0-9]+}", a.DeleteExpenseHandler).Methods("DELETE")
	adminApi.HandleFunc("/expenses/{id:[0-9]+}/submit", a.SubmitExpenseHandler).Methods("POST")
	adminApi.HandleFunc("/expenses/{id:[0-9]+}/approve", a.ApproveExpenseHandler).Methods("POST")
	adminApi.HandleFunc("/expenses/{id:[0-9]+}/reject", a.RejectExpenseHandler).Methods("POST")
	adminApi.HandleFunc("/expenses/receipts/{assetId:[0-9]+}/refresh-url", a.RefreshExpenseReceiptURLHandler).Methods("POST")

	// Expense Reconciliation routes
	adminApi.HandleFunc("/reconciliation/expenses/search", a.SearchExpensesForReconciliationHandler).Methods("GET")
	adminApi.HandleFunc("/reconciliation/expenses/{id:[0-9]+}/reconcile", a.ReconcileExpenseWithOfflineJournalHandler).Methods("POST")
	adminApi.HandleFunc("/reconciliation/offline-journals/{id:[0-9]+}/unreconcile", a.UnreconcileTransactionHandler).Methods("POST")

	// Recurring Entries routes
	adminApi.HandleFunc("/admin/recurring-entries", a.ListRecurringEntriesHandler).Methods("GET")
	adminApi.HandleFunc("/admin/recurring-entries", a.CreateRecurringEntryHandler).Methods("POST")
	adminApi.HandleFunc("/admin/recurring-entries/{id:[0-9]+}", a.UpdateRecurringEntryHandler).Methods("PUT")
	adminApi.HandleFunc("/admin/recurring-entries/{id:[0-9]+}", a.DeleteRecurringEntryHandler).Methods("DELETE")
	adminApi.HandleFunc("/admin/recurring-entries/generate", a.GenerateRecurringEntriesHandler).Methods("POST")
	adminApi.HandleFunc("/admin/recurring-entries/sync", a.SyncEmployeeRecurringEntriesHandler).Methods("POST")

	// Expense Categories routes
	adminApi.HandleFunc("/expense-categories", a.GetExpenseCategoriesHandler).Methods("GET")
	adminApi.HandleFunc("/expense-categories", a.CreateExpenseCategoryHandler).Methods("POST")
	adminApi.HandleFunc("/expense-categories/{id:[0-9]+}", a.UpdateExpenseCategoryHandler).Methods("PUT")
	adminApi.HandleFunc("/expense-categories/{id:[0-9]+}", a.DeleteExpenseCategoryHandler).Methods("DELETE")

	// Expense Tags routes
	adminApi.HandleFunc("/expense-tags", a.GetExpenseTagsHandler).Methods("GET")
	adminApi.HandleFunc("/expense-tags", a.CreateExpenseTagHandler).Methods("POST")
	adminApi.HandleFunc("/expense-tags/{id:[0-9]+}", a.UpdateExpenseTagHandler).Methods("PUT")
	adminApi.HandleFunc("/expense-tags/{id:[0-9]+}", a.DeleteExpenseTagHandler).Methods("DELETE")

	// Google Calendar Integration routes
	adminApi.HandleFunc("/google/auth/url", a.GoogleAuthURLHandler).Methods("POST")
	adminApi.HandleFunc("/google/auth/status", a.GoogleAuthStatusHandler).Methods("GET")
	adminApi.HandleFunc("/google/auth/disconnect", a.GoogleAuthDisconnectHandler).Methods("DELETE")
	adminApi.HandleFunc("/google/calendar/events", a.GoogleCalendarEventsHandler).Methods("GET")

	// Portal API Routes (scoped to client's account)
	portalApi := r.PathPrefix("/api/portal").Subrouter()
	portalApi.Use(JwtVerify)

	portalApi.HandleFunc("/invoices/draft", a.PortalDraftInvoiceListHandler).Methods("GET")
	portalApi.HandleFunc("/invoices/accepted", a.PortalInvoiceListHandler).Methods("GET")
	portalApi.HandleFunc("/projects", a.PortalProjectsListHandler).Methods("GET")
	portalApi.HandleFunc("/draft_entries", a.PortalDraftEntriesHandler).Methods("GET")
	portalApi.HandleFunc("/project_budgets", a.PortalProjectBudgetsHandler).Methods("GET")
	portalApi.HandleFunc("/weekly_hours_summary", a.PortalWeeklyHoursSummaryHandler).Methods("GET")
	portalApi.HandleFunc("/capacity", a.PortalCapacityDataHandler).Methods("GET")
	portalApi.HandleFunc("/account-details", a.PortalAccountDetailsHandler).Methods("GET")
	portalApi.HandleFunc("/assets/{assetId:[0-9]+}/refresh-url", a.PortalRefreshAssetURLHandler).Methods("POST")
	portalApi.HandleFunc("/assets/{id:[0-9]+}/download", a.AssetDownloadHandler).Methods("GET")

	// Public landing and error pages
	r.HandleFunc("/", a.CronosLandingHandler).Methods("GET")
	r.HandleFunc("/400", a.BadRequestHandler).Methods("GET")
	r.HandleFunc("/404", a.NotFoundHandler).Methods("GET")

	// SPA Entry Points
	r.HandleFunc("/admin/{any:.*}", a.AdminLandingHandler).Methods("GET")
	r.HandleFunc("/portal/{any:.*}", a.PortalLandingHandler).Methods("GET")

	// Login/Registration endpoints
	r.HandleFunc("/login", a.LoginLandingHandler).Methods("GET")
	r.HandleFunc("/verify_login", a.VerifyLogin).Methods("POST")
	r.HandleFunc("/register", a.RegistrationLandingHandler).Methods("GET")
	r.HandleFunc("/register_user", a.RegisterUser).Methods("POST")
	r.HandleFunc("/verify_email", a.VerifyEmail).Methods("POST")

	// Tenant registration (hidden link, not publicly advertised)
	r.HandleFunc("/new-organization", a.TenantRegistrationLandingHandler).Methods("GET")
	r.HandleFunc("/register_tenant", a.RegisterTenant).Methods("POST")

	// Google OAuth login endpoints (separate from calendar OAuth)
	r.HandleFunc("/auth/google/login", a.GoogleLoginHandler).Methods("GET")
	r.HandleFunc("/auth/google/login/callback", a.GoogleLoginCallbackHandler).Methods("GET")

	// Password reset endpoints
	r.HandleFunc("/password-reset", a.PasswordResetLandingHandler).Methods("GET")
	r.HandleFunc("/request_password_reset", a.RequestPasswordReset).Methods("POST")
	r.HandleFunc("/reset_password", a.ResetPassword).Methods("POST")

	// Token refresh endpoint
	r.HandleFunc("/api/refresh_token", a.RefreshTokenHandler).Methods("POST")

	// Google OAuth callback
	r.HandleFunc("/api/google/auth/callback", a.GoogleAuthCallbackHandler).Methods("GET")

	// Migration endpoints - TEMPORARY unauthenticated for development
	r.HandleFunc("/api/migrate/chart-of-accounts", a.MigrateChartOfAccountsHandler).Methods("POST")
	r.HandleFunc("/api/migrate/cleanup-subaccounts", a.CleanupSubaccountsHandler).Methods("POST")

	// Logging for web server
	f, _ := os.Create("/var/log/golang/golang-server.log")
	defer func() {
		_ = f.Close()
	}()

	// CORS middleware
	corsMiddleware := handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization", "x-access-token", "*"}),
		handlers.AllowCredentials(),
		handlers.MaxAge(86400),
	)

	logger := handlers.CombinedLoggingHandler(os.Stdout, r)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	srv := &http.Server{
		Addr:         ":" + port,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      corsMiddleware(logger),
	}

	go func() {
		log.Printf("Cronos Server Running on %q\n", srv.Addr)
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	<-c

	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()
	srv.Shutdown(ctx)
	log.Println("shutting down")
	os.Exit(0)
}
