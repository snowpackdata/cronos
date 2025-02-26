package cronos

import (
	"testing"
	"time"
)

// TestEntryDuration tests the Duration method of Entry
func TestEntryDuration(t *testing.T) {
	// Test cases
	testCases := []struct {
		name     string
		start    time.Time
		end      time.Time
		expected time.Duration
	}{
		{
			name:     "One Hour",
			start:    time.Date(2023, 1, 1, 9, 0, 0, 0, time.UTC),
			end:      time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC),
			expected: time.Hour,
		},
		{
			name:     "90 Minutes",
			start:    time.Date(2023, 1, 1, 9, 0, 0, 0, time.UTC),
			end:      time.Date(2023, 1, 1, 10, 30, 0, 0, time.UTC),
			expected: time.Hour + 30*time.Minute,
		},
		{
			name:     "Multiple Days",
			start:    time.Date(2023, 1, 1, 9, 0, 0, 0, time.UTC),
			end:      time.Date(2023, 1, 3, 9, 0, 0, 0, time.UTC),
			expected: 48 * time.Hour,
		},
		{
			name:     "Zero Duration",
			start:    time.Date(2023, 1, 1, 9, 0, 0, 0, time.UTC),
			end:      time.Date(2023, 1, 1, 9, 0, 0, 0, time.UTC),
			expected: 0,
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			entry := &Entry{
				Start: tc.start,
				End:   tc.end,
			}
			duration := entry.Duration()
			if duration != tc.expected {
				t.Errorf("Expected duration %v, got %v", tc.expected, duration)
			}
		})
	}
}

// TestEntryGetFee tests the GetFee method of Entry
func TestEntryGetFee(t *testing.T) {
	// Setup test DB
	db := setupTestDB(t)

	// Create a rate
	rate := Rate{
		Name:         "Test Rate",
		Amount:       100.0, // $100 per hour
		ActiveFrom:   time.Now().AddDate(-1, 0, 0),
		ActiveTo:     time.Now().AddDate(1, 0, 0),
		InternalOnly: false,
	}
	if err := db.Create(&rate).Error; err != nil {
		t.Fatalf("Failed to create rate: %v", err)
	}

	// Create an account
	account := Account{
		Name:      "Test Account",
		LegalName: "Test Legal Name",
		Type:      AccountTypeClient.String(),
	}
	if err := db.Create(&account).Error; err != nil {
		t.Fatalf("Failed to create account: %v", err)
	}

	// Create a project
	project := Project{
		Name:        "Test Project",
		AccountID:   account.ID,
		ActiveStart: time.Now().AddDate(-1, 0, 0),
		ActiveEnd:   time.Now().AddDate(1, 0, 0),
	}
	if err := db.Create(&project).Error; err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Create a billing code with 15-minute rounding
	billingCode := BillingCode{
		Name:        "Development",
		RateType:    RateTypeExternalBillable.String(),
		Category:    "Development",
		Code:        "DEV-001",
		RoundedTo:   15, // 15-minute rounding
		ProjectID:   project.ID,
		ActiveStart: time.Now().AddDate(-1, 0, 0),
		ActiveEnd:   time.Now().AddDate(1, 0, 0),
		RateID:      rate.ID,
	}
	if err := db.Create(&billingCode).Error; err != nil {
		t.Fatalf("Failed to create billing code: %v", err)
	}

	// Test cases
	testCases := []struct {
		name     string
		duration time.Duration
		expected float64
	}{
		{
			name:     "One Hour",
			duration: time.Hour,
			expected: 100.0, // 1 hour at $100/hour
		},
		{
			name:     "30 Minutes",
			duration: 30 * time.Minute,
			expected: 50.0, // 0.5 hours at $100/hour
		},
		{
			name:     "5 Minutes (Rounds to 15)",
			duration: 5 * time.Minute,
			expected: 0.0,
		},
		{
			name:     "16 Minutes (Rounds to 30)",
			duration: 16 * time.Minute,
			expected: 25.0,
		},
		{
			name:     "No Duration",
			duration: 0,
			expected: 0,
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			entry := &Entry{
				ProjectID:     project.ID,
				BillingCodeID: billingCode.ID,
				Start:         time.Now(),
				End:           time.Now().Add(tc.duration),
			}

			// Save the entry
			if err := db.Create(&entry).Error; err != nil {
				t.Fatalf("Failed to create entry: %v", err)
			}

			// Get the fee
			fee := entry.GetFee(db)
			if fee != tc.expected {
				t.Errorf("Expected fee %.2f, got %.2f", tc.expected, fee)
			}
		})
	}
}

// TestEntryGetInternalFee tests the GetInternalFee method of Entry
func TestEntryGetInternalFee(t *testing.T) {
	// Setup test DB
	db := setupTestDB(t)

	// Create an external rate
	externalRate := Rate{
		Name:         "External Rate",
		Amount:       100.0, // $100 per hour
		ActiveFrom:   time.Now().AddDate(-1, 0, 0),
		ActiveTo:     time.Now().AddDate(1, 0, 0),
		InternalOnly: false,
	}
	if err := db.Create(&externalRate).Error; err != nil {
		t.Fatalf("Failed to create external rate: %v", err)
	}

	// Create an internal rate
	internalRate := Rate{
		Name:         "Internal Rate",
		Amount:       50.0, // $50 per hour
		ActiveFrom:   time.Now().AddDate(-1, 0, 0),
		ActiveTo:     time.Now().AddDate(1, 0, 0),
		InternalOnly: true,
	}
	if err := db.Create(&internalRate).Error; err != nil {
		t.Fatalf("Failed to create internal rate: %v", err)
	}

	// Create an account
	account := Account{
		Name:      "Test Account",
		LegalName: "Test Legal Name",
		Type:      AccountTypeClient.String(),
	}
	if err := db.Create(&account).Error; err != nil {
		t.Fatalf("Failed to create account: %v", err)
	}

	// Create a project
	project := Project{
		Name:        "Test Project",
		AccountID:   account.ID,
		ActiveStart: time.Now().AddDate(-1, 0, 0),
		ActiveEnd:   time.Now().AddDate(1, 0, 0),
	}
	if err := db.Create(&project).Error; err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Create a billing code with both external and internal rates
	billingCode := BillingCode{
		Name:           "Development",
		RateType:       RateTypeExternalBillable.String(),
		Category:       "Development",
		Code:           "DEV-001",
		RoundedTo:      15, // 15-minute rounding
		ProjectID:      project.ID,
		ActiveStart:    time.Now().AddDate(-1, 0, 0),
		ActiveEnd:      time.Now().AddDate(1, 0, 0),
		RateID:         externalRate.ID,
		InternalRateID: internalRate.ID,
	}
	if err := db.Create(&billingCode).Error; err != nil {
		t.Fatalf("Failed to create billing code: %v", err)
	}

	// Test cases
	testCases := []struct {
		name     string
		duration time.Duration
		expected float64
	}{
		{
			name:     "One Hour",
			duration: time.Hour,
			expected: 50.0, // 1 hour at $50/hour internal rate
		},
		{
			name:     "30 Minutes",
			duration: 30 * time.Minute,
			expected: 25.0, // 0.5 hours at $50/hour internal rate
		},
		{
			name:     "5 Minutes (Rounds to 15)",
			duration: 5 * time.Minute,
			expected: 0.0,
		},
		{
			name:     "16 Minutes (Rounds to 30)",
			duration: 16 * time.Minute,
			expected: 12.5,
		},
		{
			name:     "No Duration",
			duration: 0,
			expected: 0,
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			entry := &Entry{
				ProjectID:     project.ID,
				BillingCodeID: billingCode.ID,
				Start:         time.Now(),
				End:           time.Now().Add(tc.duration),
			}

			// Save the entry
			if err := db.Create(&entry).Error; err != nil {
				t.Fatalf("Failed to create entry: %v", err)
			}

			// Get the internal fee
			fee := entry.GetInternalFee(db)
			if fee != tc.expected {
				t.Errorf("Expected internal fee %.2f, got %.2f", tc.expected, fee)
			}
		})
	}
}
