package domain

type MonthlySummary struct {
	Month             string // YYYY-MM
	HouseholdID       int
	TotalIncome       Money
	TotalExpenses     Money
	RecurringTotal    Money
	OneTimeTotal      Money
	CategoryBreakdown []CategorySummary
}

type CategorySummary struct {
	CategoryID   int
	CategoryName string
	Recurring    Money
	OneTime      Money
	Total        Money
}
