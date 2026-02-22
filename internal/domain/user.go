package domain

import "time"

type User struct {
	ID        int
	Email     string
	Name      string
	Subject   string // OIDC subject
	CreatedAt time.Time
	UpdatedAt time.Time
}
