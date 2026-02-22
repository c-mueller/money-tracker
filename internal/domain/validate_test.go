package domain

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestValidateCurrency(t *testing.T) {
	tests := []struct {
		name     string
		currency string
		wantErr  bool
	}{
		{"valid EUR", "EUR", false},
		{"valid USD", "USD", false},
		{"valid CHF", "CHF", false},
		{"lowercase", "eur", true},
		{"mixed case", "Eur", true},
		{"two letters", "EU", true},
		{"four letters", "EURO", true},
		{"empty", "", true},
		{"numbers", "123", true},
		{"with number", "EU1", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCurrency(tt.currency)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCurrency(%q) error = %v, wantErr %v", tt.currency, err, tt.wantErr)
			}
			if err != nil && !errors.Is(err, ErrValidation) {
				t.Errorf("expected ErrValidation, got %v", err)
			}
		})
	}
}

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{"valid", "user@example.com", false},
		{"valid with dots", "first.last@example.com", false},
		{"valid with plus", "user+tag@example.com", false},
		{"empty", "", true},
		{"no at sign", "userexample.com", true},
		{"no domain", "user@", true},
		{"no user", "@example.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEmail(tt.email)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateEmail(%q) error = %v, wantErr %v", tt.email, err, tt.wantErr)
			}
		})
	}
}

func TestValidateAmount(t *testing.T) {
	tests := []struct {
		name    string
		amount  string
		wantErr bool
	}{
		{"positive", "100.50", false},
		{"negative", "-50.00", false},
		{"small", "0.01", false},
		{"large valid", "999999999.99", false},
		{"zero", "0", true},
		{"exceeds max", "1000000000.00", true},
		{"exceeds max negative", "-1000000000.00", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			amount, err := NewMoney(tt.amount)
			if err != nil {
				t.Fatalf("NewMoney(%q) failed: %v", tt.amount, err)
			}
			err = ValidateAmount(amount)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAmount(%s) error = %v, wantErr %v", tt.amount, err, tt.wantErr)
			}
		})
	}
}

func TestValidateDateRange(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name    string
		start   time.Time
		end     time.Time
		wantErr bool
	}{
		{"end after start", now, now.Add(time.Hour), false},
		{"zero end", now, time.Time{}, false},
		{"end before start", now, now.Add(-time.Hour), true},
		{"same time", now, now, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDateRange(tt.start, tt.end)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDateRange() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateMonth(t *testing.T) {
	tests := []struct {
		name    string
		month   string
		wantErr bool
	}{
		{"valid january", "2026-01", false},
		{"valid december", "2026-12", false},
		{"invalid month 13", "2026-13", true},
		{"invalid month 00", "2026-00", true},
		{"invalid format", "2026/01", true},
		{"empty", "", true},
		{"just year", "2026", true},
		{"full date", "2026-01-01", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMonth(tt.month)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateMonth(%q) error = %v, wantErr %v", tt.month, err, tt.wantErr)
			}
		})
	}
}

func TestValidateHouseholdName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid", "My Household", false},
		{"single char", "A", false},
		{"max length", strings.Repeat("a", 100), false},
		{"empty", "", true},
		{"too long", strings.Repeat("a", 101), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateHouseholdName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateHouseholdName(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestValidateCategoryName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid", "Groceries", false},
		{"single char", "A", false},
		{"max length", strings.Repeat("a", 50), false},
		{"empty", "", true},
		{"too long", strings.Repeat("a", 51), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCategoryName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCategoryName(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestValidateDescription(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid", "Some description", false},
		{"empty", "", false},
		{"max length", strings.Repeat("a", 500), false},
		{"too long", strings.Repeat("a", 501), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDescription(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDescription(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}
