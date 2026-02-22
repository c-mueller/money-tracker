package domain

import (
	"context"
	"time"
)

type UserRepo interface {
	Create(ctx context.Context, user *User) (*User, error)
	GetByID(ctx context.Context, id int) (*User, error)
	GetBySubject(ctx context.Context, subject string) (*User, error)
	Update(ctx context.Context, user *User) (*User, error)
}

type HouseholdRepo interface {
	Create(ctx context.Context, household *Household) (*Household, error)
	GetByID(ctx context.Context, id int) (*Household, error)
	ListByOwner(ctx context.Context, ownerID int) ([]*Household, error)
	Update(ctx context.Context, household *Household) (*Household, error)
	Delete(ctx context.Context, id int) error
}

type CategoryRepo interface {
	Create(ctx context.Context, category *Category) (*Category, error)
	GetByID(ctx context.Context, id int) (*Category, error)
	ListByHousehold(ctx context.Context, householdID int) ([]*Category, error)
	Update(ctx context.Context, category *Category) (*Category, error)
	Delete(ctx context.Context, id int) error
}

type TransactionRepo interface {
	Create(ctx context.Context, tx *Transaction) (*Transaction, error)
	GetByID(ctx context.Context, id int) (*Transaction, error)
	ListByHouseholdAndMonth(ctx context.Context, householdID int, year int, month time.Month) ([]*Transaction, error)
	Delete(ctx context.Context, id int) error
}

type RecurringExpenseRepo interface {
	Create(ctx context.Context, expense *RecurringExpense) (*RecurringExpense, error)
	GetByID(ctx context.Context, id int) (*RecurringExpense, error)
	ListByHousehold(ctx context.Context, householdID int) ([]*RecurringExpense, error)
	ListActiveByHousehold(ctx context.Context, householdID int) ([]*RecurringExpense, error)
	Update(ctx context.Context, expense *RecurringExpense) (*RecurringExpense, error)
	Delete(ctx context.Context, id int) error
}

type APITokenRepo interface {
	Create(ctx context.Context, token *APIToken) (*APIToken, error)
	GetByHash(ctx context.Context, hash string) (*APIToken, error)
	ListByUser(ctx context.Context, userID int) ([]*APIToken, error)
	UpdateLastUsed(ctx context.Context, id int, t time.Time) error
	Delete(ctx context.Context, id int) error
}
