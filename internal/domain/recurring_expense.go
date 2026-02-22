package domain

import "time"

type RecurringExpense struct {
	ID          int
	HouseholdID int
	CategoryID  int
	Name        string
	Amount      Money
	Frequency   Frequency
	Active      bool
	StartDate   time.Time
	EndDate     *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
