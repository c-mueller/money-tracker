package domain

import "time"

type MonthlySummary struct {
	Month             string // YYYY-MM
	HouseholdID       int
	TotalIncome       Money // one-time income
	TotalExpenses     Money // one-time expenses
	RecurringTotal    Money
	RecurringIncome   Money
	RecurringExpenses Money
	OneTimeTotal      Money
	OneTimeIncome     Money
	OneTimeExpenses   Money
	GrossIncome       Money // one-time + recurring income
	GrossExpenses     Money // one-time + recurring expenses
	MonthlyTotal      Money
	CategoryBreakdown []CategorySummary
	RecurringGroups   []RecurringFrequencyGroup
	IncomeRecurringEntries  []RecurringEntry
	ExpenseRecurringEntries []RecurringEntry
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
	CategoryID    int
	Amount        Money
	Frequency     Frequency
	MonthlyAmount Money
	EffectiveDate time.Time
}
