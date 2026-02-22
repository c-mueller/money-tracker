package domain

import "time"

type APIToken struct {
	ID        int
	UserID    int
	Name      string
	TokenHash string
	ExpiresAt *time.Time
	CreatedAt time.Time
	LastUsed  *time.Time
}
