package domain

import (
	"sort"
	"time"
)

type RecurringExpense struct {
	ID          int
	HouseholdID int
	CategoryID  int
	Name        string
	Description string
	Amount      Money
	Frequency   Frequency
	Active      bool
	StartDate   time.Time
	EndDate     *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type RecurringScheduleOverride struct {
	ID                 int
	RecurringExpenseID int
	EffectiveDate      time.Time
	Amount             Money
	Frequency          Frequency
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// EffectiveSchedule returns the amount and frequency in effect for a given month,
// considering any schedule overrides. Overrides are sorted by effective_date ascending.
// The latest override where effective_date <= last day of queried month is used.
func EffectiveSchedule(baseAmount Money, baseFreq Frequency, overrides []*RecurringScheduleOverride, year int, month time.Month) (Money, Frequency) {
	if len(overrides) == 0 {
		return baseAmount, baseFreq
	}

	// Sort by effective_date ascending
	sorted := make([]*RecurringScheduleOverride, len(overrides))
	copy(sorted, overrides)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].EffectiveDate.Before(sorted[j].EffectiveDate)
	})

	// Last day of the queried month
	lastDay := time.Date(year, month+1, 0, 23, 59, 59, 0, time.UTC)

	var effective *RecurringScheduleOverride
	for _, o := range sorted {
		if !o.EffectiveDate.After(lastDay) {
			effective = o
		}
	}

	if effective != nil {
		return effective.Amount, effective.Frequency
	}
	return baseAmount, baseFreq
}
