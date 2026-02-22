package domain

import "time"

type Household struct {
	ID        int
	Name        string
	Description string
	Currency    string
	Icon        string
	OwnerID   int
	CreatedAt time.Time
	UpdatedAt time.Time
}
