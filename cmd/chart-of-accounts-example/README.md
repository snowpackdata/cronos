# Chart of Accounts Example

This example demonstrates how to use the new Chart of Accounts features in Cronos.

## Features Demonstrated

1. **Seeding System Accounts** - Initialize the predefined GL accounts
2. **Creating Custom Accounts** - Add new expense categories beyond the defaults
3. **Managing Subaccounts** - Create subaccounts for vendors, clients, etc.
4. **Categorized Expenses** - Create expenses that book to specific GL accounts
5. **CSV Import** - Import bank/credit card statements
6. **Transaction Categorization** - Review and categorize imported transactions
7. **Approval & Booking** - Approve categorized transactions and book to GL

## Running the Example

This is a code reference example. To use it in your application:

```bash
# Copy the patterns into your application code
# The example uses commented code to show the API usage
```

## Quick Start Guide

### 1. Initial Setup (One Time)

```go
// Seed all system-defined accounts
err := app.SeedSystemAccounts()
if err != nil {
    log.Fatalf("Failed to seed accounts: %v", err)
}
```

### 2. Create Custom Categories

```go
// Create a new expense category
account, err := app.CreateChartOfAccount(
    "MARKETING_EXPENSES",
    "Marketing & Advertising",
    "EXPENSE",
    "Marketing and advertising costs",
    nil,
)

// Add subaccounts for tracking
app.CreateSubaccount("GOOGLE_ADS", "Google Advertising", "MARKETING_EXPENSES", "VENDOR")
app.CreateSubaccount("FACEBOOK_ADS", "Facebook Advertising", "MARKETING_EXPENSES", "VENDOR")
```

### 3. Create Categorized Expense

```go
expense := cronos.Expense{
    ProjectID:          projectID,
    SubmitterID:        employeeID,
    Amount:             250000, // $2,500
    Date:               time.Now(),
    Description:        "Google Ads campaign",
    ExpenseAccountCode: "MARKETING_EXPENSES",
    SubaccountCode:     "GOOGLE_ADS",
    State:              cronos.ExpenseStateDraft.String(),
}
```

### 4. Import CSV Transactions

```go
// Read CSV file
csvContent, _ := ioutil.ReadFile("statement.csv")

// Import to offline journals for review
// Creates 2 UNCLASSIFIED entries per transaction (FROM and TO)
imported, skipped, err := app.ImportCSVToOfflineJournals(
    csvContent,
    0, 1, 2,         // date, description, amount columns
    true,            // has header row
    "01/02/2006",    // date format
)
// Each transaction creates a debit entry and a credit entry
```

### 5. Categorize Transactions

```go
// Get pending transactions grouped by date+description
transactions, _ := app.GetOfflineJournalTransactions(startDate, endDate, "pending_review")

// Categorize each transaction by specifying FROM and TO accounts
for _, txEntries := range transactions {
    desc := txEntries[0].Description
    date := txEntries[0].Date
    
    if strings.Contains(desc, "AWS") {
        // Expense: FROM Operating Expenses (debit) TO Cash (credit)
        app.CategorizeCSVTransaction(
            date, desc,
            "OPERATING_EXPENSES_SAAS", "AWS",    // FROM (debit)
            "CASH", "ChaseBusiness",              // TO (credit)
        )
    } else if strings.Contains(desc, "Client Payment") {
        // Revenue: FROM Cash (debit) TO Revenue (credit)
        app.CategorizeCSVTransaction(
            date, desc,
            "CASH", "ChaseBusiness",              // FROM (debit)
            "REVENUE", "Client XYZ",              // TO (credit)
        )
    }
}
```

### 6. Approve & Book

```go
// Approve complete transactions (both sides must be categorized)
for _, txEntries := range transactions {
    desc := txEntries[0].Description
    date := txEntries[0].Date
    
    // Check if both sides are categorized
    allCategorized := true
    for _, entry := range txEntries {
        if entry.Account == "UNCLASSIFIED" {
            allCategorized = false
            break
        }
    }
    
    if allCategorized {
        booked, err := app.ApproveTransactionPair(date, desc, staffID)
        log.Printf("Booked transaction: %s (%d entries)", desc, booked)
    }
}
```

## Account Types

- `ASSET` - Assets (cash, receivables, equipment)
- `LIABILITY` - Liabilities (payables, credit cards)
- `EQUITY` - Equity accounts
- `REVENUE` - Revenue and income
- `EXPENSE` - Expenses and costs

## Subaccount Types

- `VENDOR` - Vendor/supplier accounts
- `CLIENT` - Client/customer accounts
- `EMPLOYEE` - Employee accounts
- `CUSTOM` - Custom categorization

## CSV Format Support

The CSV parser is flexible and supports:

### Date Formats (auto-detected)
- ISO: `2006-01-02`
- US: `01/02/2006` or `1/2/2006`
- Slash: `2006/01/02`
- Named: `Jan 2, 2006` or `January 2, 2006`

### Amount Formats
- Plain: `1234.56`
- Currency: `$1,234.56`
- Negative: `-123.45`
- Parentheses: `(123.45)` (treated as negative)

### Common Bank CSV Formats

**Chase Bank:**
```csv
Date,Description,Amount,Balance
01/15/2024,AWS SERVICES,-1234.56,5432.10
```

**American Express:**
```csv
Date,Description,Amount
01/15/2024,AMAZON WEB SERVICES,(1234.56)
```

**Generic:**
```csv
Transaction Date,Merchant,Debit,Credit
2024-01-15,AWS Inc,1234.56,
```

## Testing

To test in your application:

1. Create test database
2. Run `SeedSystemAccounts()`
3. Create a few custom accounts
4. Import a sample CSV
5. Categorize and approve

See `CHART_OF_ACCOUNTS.md` in the main cronos directory for full API documentation.

## Migration

If you're adding this to an existing Cronos installation:

```go
// In your migration code
db.AutoMigrate(&cronos.ChartOfAccount{}, &cronos.Subaccount{})

// Seed system accounts
app := &cronos.App{DB: db, ...}
err := app.SeedSystemAccounts()
if err != nil {
    log.Fatalf("Failed to seed: %v", err)
}

// Update Expense model (already has new fields)
db.AutoMigrate(&cronos.Expense{})
```

## Next Steps

After understanding the example:

1. Build UI for account/subaccount management
2. Add CSV upload endpoint
3. Create transaction categorization interface
4. Add approval workflow UI
5. Build reporting by account/subaccount
6. Consider auto-categorization rules

