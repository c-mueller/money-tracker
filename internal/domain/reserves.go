package domain

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

// NormalizeToMonthly converts an amount with a given frequency to a monthly equivalent
// for a specific reference month. The reference month is needed for day-based frequencies
// to know how many days/weekdays are in the month.
func NormalizeToMonthly(amount Money, freq Frequency, refMonth time.Time) (Money, error) {
	if err := freq.Validate(); err != nil {
		return ZeroMoney(), err
	}

	switch freq {
	case FrequencyMonthly:
		return amount, nil

	case FrequencyQuarterly:
		return amount.Div(decimal.NewFromInt(3)), nil

	case FrequencyYearly:
		return amount.Div(decimal.NewFromInt(12)), nil

	case FrequencyWeekly:
		// 52 weeks / 12 months ≈ 4.333...
		return amount.Mul(decimal.NewFromInt(52)).Div(decimal.NewFromInt(12)), nil

	case FrequencyBiweekly:
		// 26 bi-weeks / 12 months ≈ 2.166...
		return amount.Mul(decimal.NewFromInt(26)).Div(decimal.NewFromInt(12)), nil

	case FrequencyDaily:
		days := daysInMonth(refMonth)
		return amount.Mul(decimal.NewFromInt(int64(days))), nil

	case FrequencyWeekday:
		weekdays := weekdaysInMonth(refMonth)
		return amount.Mul(decimal.NewFromInt(int64(weekdays))), nil

	default:
		return ZeroMoney(), fmt.Errorf("%w: unknown frequency %q", ErrValidation, freq)
	}
}

func daysInMonth(t time.Time) int {
	y, m, _ := t.Date()
	return time.Date(y, m+1, 0, 0, 0, 0, 0, t.Location()).Day()
}

func weekdaysInMonth(t time.Time) int {
	y, m, _ := t.Date()
	days := daysInMonth(t)
	count := 0
	for d := 1; d <= days; d++ {
		wd := time.Date(y, m, d, 0, 0, 0, 0, t.Location()).Weekday()
		if wd != time.Saturday && wd != time.Sunday {
			count++
		}
	}
	return count
}
