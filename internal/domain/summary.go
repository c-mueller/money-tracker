package domain

import "time"

type MonthlySummary struct {
	Month             string // YYYY-MM
	HouseholdID       int
	TotalIncome       Money
	TotalExpenses     Money
	RecurringTotal    Money
	RecurringIncome   Money
	RecurringExpenses Money
	OneTimeTotal      Money
	OneTimeIncome     Money
	OneTimeExpenses   Money
	MonthlyTotal      Money
	CategoryBreakdown []CategorySummary
	RecurringGroups   []RecurringFrequencyGroup
}

type CategorySummary struct {
	CategoryID   int
	CategoryName string
	Recurring    Money
	OneTime      Money
	Total        Money
}

type RecurringFrequencyGroup struct {
	Frequency Frequency
	Total     Money
	Entries   []RecurringEntry
}

type RecurringEntry struct {
	Name          string
	Amount        Money
	Frequency     Frequency
	MonthlyAmount Money
	EffectiveDate time.Time
}
