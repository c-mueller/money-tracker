package domain

import (
	"regexp"
	"time"

	"github.com/shopspring/decimal"
)

var (
	currencyRegex = regexp.MustCompile(`^[A-Z]{3}$`)
	emailRegex    = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	monthRegex    = regexp.MustCompile(`^\d{4}-(0[1-9]|1[0-2])$`)
	maxAmount     = decimal.NewFromFloat(999999999.99)
)

func ValidateCurrency(currency string) error {
	if !currencyRegex.MatchString(currency) {
		return NewValidationError("currency", "must be a 3-letter uppercase ISO code")
	}
	return nil
}

func ValidateEmail(email string) error {
	if !emailRegex.MatchString(email) {
		return NewValidationError("email", "invalid email format")
	}
	return nil
}

func ValidateAmount(amount Money) error {
	if amount.IsZero() {
		return NewValidationError("amount", "must not be zero")
	}
	if amount.Abs().GreaterThan(maxAmount) {
		return NewValidationError("amount", "exceeds maximum (999999999.99)")
	}
	return nil
}

func ValidateDateRange(start, end time.Time) error {
	if !end.IsZero() && end.Before(start) {
		return NewValidationError("end_date", "must be after start_date")
	}
	return nil
}

func ValidateMonth(month string) error {
	if !monthRegex.MatchString(month) {
		return NewValidationError("month", "must be in YYYY-MM format")
	}
	return nil
}

func ValidateHouseholdName(name string) error {
	if len(name) == 0 || len(name) > 100 {
		return NewValidationError("name", "must be 1-100 characters")
	}
	return nil
}

func ValidateCategoryName(name string) error {
	if len(name) == 0 || len(name) > 50 {
		return NewValidationError("name", "must be 1-50 characters")
	}
	return nil
}

func ValidateDescription(desc string) error {
	if len(desc) > 500 {
		return NewValidationError("description", "must be at most 500 characters")
	}
	return nil
}
