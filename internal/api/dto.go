package api

import "time"

// Household DTOs
type CreateHouseholdRequest struct {
	Name     string `json:"name"`
	Currency string `json:"currency"`
}

type HouseholdResponse struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Currency  string    `json:"currency"`
	OwnerID   int       `json:"owner_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Category DTOs
type CreateCategoryRequest struct {
	Name string `json:"name"`
}

type CategoryResponse struct {
	ID          int       `json:"id"`
	HouseholdID int      `json:"household_id"`
	Name        string    `json:"name"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Transaction DTOs
type CreateTransactionRequest struct {
	CategoryID  int    `json:"category_id"`
	Amount      string `json:"amount"`
	Description string `json:"description"`
	Date        string `json:"date"`
}

type TransactionResponse struct {
	ID          int       `json:"id"`
	HouseholdID int      `json:"household_id"`
	CategoryID  int       `json:"category_id"`
	Amount      string    `json:"amount"`
	Description string    `json:"description"`
	Date        string    `json:"date"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// RecurringExpense DTOs
type CreateRecurringExpenseRequest struct {
	CategoryID int     `json:"category_id"`
	Name       string  `json:"name"`
	Amount     string  `json:"amount"`
	Frequency  string  `json:"frequency"`
	StartDate  string  `json:"start_date"`
	EndDate    *string `json:"end_date,omitempty"`
}

type UpdateRecurringExpenseRequest struct {
	CategoryID int     `json:"category_id"`
	Name       string  `json:"name"`
	Amount     string  `json:"amount"`
	Frequency  string  `json:"frequency"`
	Active     bool    `json:"active"`
	StartDate  string  `json:"start_date"`
	EndDate    *string `json:"end_date,omitempty"`
}

type RecurringExpenseResponse struct {
	ID          int       `json:"id"`
	HouseholdID int      `json:"household_id"`
	CategoryID  int       `json:"category_id"`
	Name        string    `json:"name"`
	Amount      string    `json:"amount"`
	Frequency   string    `json:"frequency"`
	Active      bool      `json:"active"`
	StartDate   string    `json:"start_date"`
	EndDate     *string   `json:"end_date,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Summary DTOs
type SummaryResponse struct {
	Month             string                    `json:"month"`
	HouseholdID       int                       `json:"household_id"`
	TotalIncome       string                    `json:"total_income"`
	TotalExpenses     string                    `json:"total_expenses"`
	RecurringTotal    string                    `json:"recurring_total"`
	OneTimeTotal      string                    `json:"one_time_total"`
	CategoryBreakdown []CategorySummaryResponse `json:"category_breakdown"`
}

type CategorySummaryResponse struct {
	CategoryID   int    `json:"category_id"`
	CategoryName string `json:"category_name"`
	Recurring    string `json:"recurring"`
	OneTime      string `json:"one_time"`
	Total        string `json:"total"`
}
