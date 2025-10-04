package cronos

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// BeancountTransaction represents a parsed transaction from a Beancount file
type BeancountTransaction struct {
	Date        time.Time
	Flag        string // * or !
	Payee       string
	Description string
	Tags        []string
	Postings    []BeancountPosting
	LineNumber  int
}

// BeancountPosting represents a single posting within a transaction
type BeancountPosting struct {
	Account  string
	Amount   float64 // Amount in base currency units (e.g., dollars)
	Currency string
}

// BeancountAccount represents an account opening directive
type BeancountAccount struct {
	Date       time.Time
	Name       string
	LineNumber int
}

// BeancountBalance represents a balance assertion
type BeancountBalance struct {
	Date       time.Time
	Account    string
	Amount     float64
	Currency   string
	LineNumber int
}

// BeancountLedger represents the parsed contents of a Beancount file
type BeancountLedger struct {
	Transactions []BeancountTransaction
	Accounts     []BeancountAccount
	Balances     []BeancountBalance
	FilePath     string
}

// ParseBeancountFile parses a Beancount file and returns the ledger
func ParseBeancountFile(filepath string) (*BeancountLedger, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open beancount file: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	ledger, err := ParseBeancountFromBytes(data)
	if err != nil {
		return nil, err
	}
	ledger.FilePath = filepath
	return ledger, nil
}

// ParseBeancountFromBytes parses Beancount data from a byte slice
func ParseBeancountFromBytes(data []byte) (*BeancountLedger, error) {
	ledger := &BeancountLedger{
		Transactions: []BeancountTransaction{},
		Accounts:     []BeancountAccount{},
		Balances:     []BeancountBalance{},
	}

	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	lineNumber := 0
	var currentTransaction *BeancountTransaction

	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)

		// Skip empty lines and comments
		if trimmedLine == "" || strings.HasPrefix(trimmedLine, ";") || strings.HasPrefix(trimmedLine, "#") {
			continue
		}

		// Handle transaction start (date + flag + description)
		if isTransactionStart(trimmedLine) {
			// Save previous transaction if exists
			if currentTransaction != nil {
				ledger.Transactions = append(ledger.Transactions, *currentTransaction)
			}

			// Parse new transaction
			tx, err := parseTransactionHeader(trimmedLine, lineNumber)
			if err != nil {
				return nil, fmt.Errorf("line %d: %w", lineNumber, err)
			}
			currentTransaction = tx
			continue
		}

		// Handle account opening
		if strings.HasPrefix(trimmedLine, "open ") {
			account, err := parseAccountOpen(trimmedLine, lineNumber)
			if err != nil {
				return nil, fmt.Errorf("line %d: %w", lineNumber, err)
			}
			ledger.Accounts = append(ledger.Accounts, account)
			continue
		}

		// Handle balance assertion
		if strings.HasPrefix(trimmedLine, "balance ") {
			balance, err := parseBalance(trimmedLine, lineNumber)
			if err != nil {
				return nil, fmt.Errorf("line %d: %w", lineNumber, err)
			}
			ledger.Balances = append(ledger.Balances, balance)
			continue
		}

		// Handle posting (indented line under a transaction)
		if currentTransaction != nil && (strings.HasPrefix(line, "  ") || strings.HasPrefix(line, "\t")) {
			posting, err := parsePosting(trimmedLine)
			if err != nil {
				return nil, fmt.Errorf("line %d: %w", lineNumber, err)
			}
			if posting != nil {
				currentTransaction.Postings = append(currentTransaction.Postings, *posting)
			}
			continue
		}

		// Handle option or commodity directives (skip for now)
		if strings.HasPrefix(trimmedLine, "option ") ||
			strings.HasPrefix(trimmedLine, "commodity ") ||
			strings.HasPrefix(trimmedLine, "close ") {
			continue
		}
	}

	// Save last transaction
	if currentTransaction != nil {
		ledger.Transactions = append(ledger.Transactions, *currentTransaction)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return ledger, nil
}

// isTransactionStart checks if a line starts a transaction (has date + flag)
func isTransactionStart(line string) bool {
	// Match: YYYY-MM-DD * "..." or YYYY-MM-DD txn "..."
	// Handle both spaces and tabs
	datePattern := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}[\s\t]+[\*!]`)
	return datePattern.MatchString(line)
}

// parseTransactionHeader parses the first line of a transaction
func parseTransactionHeader(line string, lineNum int) (*BeancountTransaction, error) {
	// Pattern: YYYY-MM-DD FLAG "PAYEE" "DESCRIPTION" or YYYY-MM-DD FLAG "DESCRIPTION"
	// Normalize tabs to spaces
	line = strings.ReplaceAll(line, "\t", " ")
	parts := strings.Fields(line) // Fields handles multiple spaces
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid transaction header")
	}

	date, err := time.Parse("2006-01-02", parts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid date: %w", err)
	}

	flag := parts[1]
	// Rest is everything after the flag
	rest := ""
	if len(parts) > 2 {
		rest = strings.Join(parts[2:], " ")
	}

	// Extract strings between quotes and tags
	description := ""
	tags := []string{}

	// Find all quoted strings
	quotePattern := regexp.MustCompile(`"([^"]*)"`)
	matches := quotePattern.FindAllStringSubmatch(rest, -1)
	if len(matches) > 0 {
		description = matches[0][1]
	}

	// Find tags (words starting with #)
	tagPattern := regexp.MustCompile(`#(\w+)`)
	tagMatches := tagPattern.FindAllStringSubmatch(rest, -1)
	for _, match := range tagMatches {
		tags = append(tags, match[1])
	}

	return &BeancountTransaction{
		Date:        date,
		Flag:        flag,
		Description: description,
		Tags:        tags,
		Postings:    []BeancountPosting{},
		LineNumber:  lineNum,
	}, nil
}

// parsePosting parses a posting line
func parsePosting(line string) (*BeancountPosting, error) {
	// Pattern: ACCOUNT [AMOUNT CURRENCY]
	// Amount and currency are optional (can be inferred)

	trimmed := strings.TrimSpace(line)
	parts := strings.Fields(trimmed)

	if len(parts) == 0 {
		return nil, nil
	}

	account := parts[0]

	// If no amount specified, return posting with zero amount
	if len(parts) == 1 {
		return &BeancountPosting{
			Account:  account,
			Amount:   0,
			Currency: "USD",
		}, nil
	}

	// Parse amount - could be formats like: -123.45, 123, 1,234.56
	amountStr := parts[1]
	// Remove commas from amount
	amountStr = strings.ReplaceAll(amountStr, ",", "")

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid amount '%s': %w", parts[1], err)
	}

	currency := "USD"
	if len(parts) > 2 {
		currency = parts[2]
	}

	return &BeancountPosting{
		Account:  account,
		Amount:   amount,
		Currency: currency,
	}, nil
}

// parseAccountOpen parses an account opening directive
func parseAccountOpen(line string, lineNum int) (BeancountAccount, error) {
	// Pattern: YYYY-MM-DD open ACCOUNT [CURRENCIES...]
	parts := strings.Fields(line)
	if len(parts) < 3 {
		return BeancountAccount{}, fmt.Errorf("invalid account open directive")
	}

	date, err := time.Parse("2006-01-02", parts[0])
	if err != nil {
		return BeancountAccount{}, fmt.Errorf("invalid date: %w", err)
	}

	return BeancountAccount{
		Date:       date,
		Name:       parts[2],
		LineNumber: lineNum,
	}, nil
}

// parseBalance parses a balance assertion
func parseBalance(line string, lineNum int) (BeancountBalance, error) {
	// Pattern: YYYY-MM-DD balance ACCOUNT AMOUNT CURRENCY
	parts := strings.Fields(line)
	if len(parts) < 5 {
		return BeancountBalance{}, fmt.Errorf("invalid balance assertion")
	}

	date, err := time.Parse("2006-01-02", parts[0])
	if err != nil {
		return BeancountBalance{}, fmt.Errorf("invalid date: %w", err)
	}

	account := parts[2]

	// Remove commas from amount
	amountStr := strings.ReplaceAll(parts[3], ",", "")
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		return BeancountBalance{}, fmt.Errorf("invalid amount: %w", err)
	}

	currency := parts[4]

	return BeancountBalance{
		Date:       date,
		Account:    account,
		Amount:     amount,
		Currency:   currency,
		LineNumber: lineNum,
	}, nil
}

// BalancePostings adds implicit amounts to postings with zero amounts
// In Beancount, one posting can omit the amount and it's inferred
func (tx *BeancountTransaction) BalancePostings() error {
	var total float64
	var zeroIndex = -1

	for i, posting := range tx.Postings {
		if posting.Amount == 0 {
			if zeroIndex >= 0 {
				return fmt.Errorf("multiple postings with inferred amounts")
			}
			zeroIndex = i
		} else {
			total += posting.Amount
		}
	}

	// If we have exactly one zero posting, infer its amount
	if zeroIndex >= 0 {
		tx.Postings[zeroIndex].Amount = -total
	}

	return nil
}

// Validate checks that a transaction balances (sum of postings = 0)
func (tx *BeancountTransaction) Validate() error {
	if err := tx.BalancePostings(); err != nil {
		return err
	}

	var total float64
	for _, posting := range tx.Postings {
		total += posting.Amount
	}

	// Allow for small floating point errors (less than 1 cent)
	if total > 0.01 || total < -0.01 {
		return fmt.Errorf("transaction does not balance: sum=%.2f, line=%d", total, tx.LineNumber)
	}

	return nil
}

// ValidateAll validates all transactions in the ledger
func (ledger *BeancountLedger) ValidateAll() []error {
	var errors []error
	for _, tx := range ledger.Transactions {
		if err := tx.Validate(); err != nil {
			errors = append(errors, err)
		}
	}
	return errors
}
