package domain

import "time"

type Transaction struct {
	ID          int
	HouseholdID int
	CategoryID  int
	Amount      Money
	Description string
	Date        time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
