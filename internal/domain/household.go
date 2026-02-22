package domain

import "time"

type Household struct {
	ID        int
	Name      string
	Currency  string
	OwnerID   int
	CreatedAt time.Time
	UpdatedAt time.Time
}
