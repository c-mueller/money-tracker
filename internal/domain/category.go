package domain

import "time"

type Category struct {
	ID          int
	HouseholdID int
	Name        string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
