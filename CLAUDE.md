# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build and Test Commands

```bash
# Run all tests
go test -v ./...

# Run a specific test
go test -v ./... -run "TestName"

# Run project-related tests
go test -v ./... -run "TestCreateProject|TestUpdateProject|TestAddBillingCodeToProject"

# Run user registration tests
go test -v ./... -run "TestRegisterClient|TestRegisterStaff"
```

## Architecture Overview

Cronos is an internal timekeeping and billing system built as a Go library. It provides AP/AR functionality for tracking accounts, users, billing codes, and time entries. The core logic is in this repository; the web API/handlers live in a separate internal repository.

### Core Pattern: App Struct

All business logic operates through the `App` struct which holds the GORM database connection:

```go
type App struct {
    DB      *gorm.DB
    Project string  // GCP project
    Bucket  string  // GCS bucket for PDFs
}
```

Initialize with `InitializeSQLite()` for local dev or `InitializeLocal()`/`InitializeCloud()` for Postgres.

### Key Domain Models (models.go)

- **User/Employee/Client**: User authentication and employee/client profiles
- **Account**: Customer accounts with billing settings (frequency, budget)
- **Project**: Work units with budgets, date ranges, and sales attribution (AE/SDR)
- **BillingCode**: Links projects to rates; entries reference billing codes
- **Entry**: Time entries with start/end times, employee, optional impersonation
- **Invoice**: AR invoices to clients (tracks state: Draft → Approved → Sent → Paid)
- **Bill**: AP bills to employees for their work

### Accounting System

**Double-entry journal system** in `models.go`:
- `Journal`: Real-time entries created from invoice/bill state transitions
- `OfflineJournal`: Historical entries imported from Beancount files or CSV

**Ledger unification** (`ledger.go`): Merges Journal and OfflineJournal entries for reporting with account mapping between Beancount format and internal account types.

### PDF Generation

`generate_invoice.go` and `generate_bill.go` use gofpdf to create PDF invoices/bills, stored in GCS.

### Impersonation Feature

Entries support impersonation where one employee can create entries displayed as another on client invoices, while internal billing credits the actual creator. Check `ImpersonateAsUserID` on Entry.

### State Machines

Entries: `Unaffiliated → Draft → Approved/Rejected/Excluded → Sent → Paid/Void`
Invoices: `Draft → Approved → Sent → Paid/Void`
Bills: `Draft → Accepted → Paid/Void`

### Commission Calculation

Projects have `ProjectType` (New/Existing) affecting commission rates for AE and SDR roles. See constants in `models.go` (e.g., `AECommissionRateNewLarge`, `SDRCommissionRateExistingSmall`).

### Expenses System

`expense_categories_tags.go`: Expense tracking with categories, tags, and state workflow (Draft → Submitted → Approved → Invoiced → Paid).

### Recurring Entries

`recurring_entries.go`: Templates for auto-generating regular payroll entries (base salary, bonuses, stipends).
