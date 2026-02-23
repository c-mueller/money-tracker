package domain

import (
	"testing"
	"time"
)

func TestIsActiveInMonth(t *testing.T) {
	tests := []struct {
		name      string
		startDate time.Time
		endDate   *time.Time
		year      int
		month     time.Month
		want      bool
	}{
		{
			name:      "active - started before month",
			startDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			year:      2026, month: time.March,
			want: true,
		},
		{
			name:      "active - started same month",
			startDate: time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC),
			year:      2026, month: time.March,
			want: true,
		},
		{
			name:      "inactive - not started yet",
			startDate: time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
			year:      2026, month: time.March,
			want: false,
		},
		{
			name:      "active - no end date",
			startDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			year:      2026, month: time.December,
			want: true,
		},
		{
			name:      "inactive - ended before month",
			startDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			endDate:   timePtr(time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC)),
			year:      2026, month: time.March,
			want: false,
		},
		{
			name:      "active - ends during month",
			startDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			endDate:   timePtr(time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)),
			year:      2026, month: time.March,
			want: true,
		},
		{
			name:      "active - starts on last day of month",
			startDate: time.Date(2026, 3, 31, 0, 0, 0, 0, time.UTC),
			year:      2026, month: time.March,
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re := &RecurringExpense{
				StartDate: tt.startDate,
				EndDate:   tt.endDate,
			}
			got := re.IsActiveInMonth(tt.year, tt.month)
			if got != tt.want {
				t.Errorf("IsActiveInMonth(%d, %s) = %v, want %v", tt.year, tt.month, got, tt.want)
			}
		})
	}
}

func timePtr(t time.Time) *time.Time {
	return &t
}

func TestEffectiveSchedule(t *testing.T) {
	baseAmount, _ := NewMoney("-800.00")
	baseFreq := FrequencyMonthly

	t.Run("no overrides", func(t *testing.T) {
		amount, freq := EffectiveSchedule(baseAmount, baseFreq, nil, 2026, time.January)
		if !amount.Equal(baseAmount) {
			t.Errorf("amount = %s, want %s", amount.String(), baseAmount.String())
		}
		if freq != baseFreq {
			t.Errorf("freq = %s, want %s", freq, baseFreq)
		}
	})

	t.Run("single override before month", func(t *testing.T) {
		overrideAmount, _ := NewMoney("-900.00")
		overrides := []*RecurringScheduleOverride{
			{
				EffectiveDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
				Amount:        overrideAmount,
				Frequency:     FrequencyMonthly,
			},
		}

		amount, freq := EffectiveSchedule(baseAmount, baseFreq, overrides, 2026, time.February)
		if !amount.Equal(overrideAmount) {
			t.Errorf("amount = %s, want %s", amount.String(), overrideAmount.String())
		}
		if freq != FrequencyMonthly {
			t.Errorf("freq = %s, want %s", freq, FrequencyMonthly)
		}
	})

	t.Run("single override in same month", func(t *testing.T) {
		overrideAmount, _ := NewMoney("-950.00")
		overrides := []*RecurringScheduleOverride{
			{
				EffectiveDate: time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC),
				Amount:        overrideAmount,
				Frequency:     FrequencyMonthly,
			},
		}

		amount, _ := EffectiveSchedule(baseAmount, baseFreq, overrides, 2026, time.March)
		if !amount.Equal(overrideAmount) {
			t.Errorf("amount = %s, want %s", amount.String(), overrideAmount.String())
		}
	})

	t.Run("override after month returns base", func(t *testing.T) {
		overrideAmount, _ := NewMoney("-1000.00")
		overrides := []*RecurringScheduleOverride{
			{
				EffectiveDate: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
				Amount:        overrideAmount,
				Frequency:     FrequencyMonthly,
			},
		}

		amount, freq := EffectiveSchedule(baseAmount, baseFreq, overrides, 2026, time.March)
		if !amount.Equal(baseAmount) {
			t.Errorf("amount = %s, want %s", amount.String(), baseAmount.String())
		}
		if freq != baseFreq {
			t.Errorf("freq = %s, want %s", freq, baseFreq)
		}
	})

	t.Run("multiple overrides picks latest applicable", func(t *testing.T) {
		amount1, _ := NewMoney("-850.00")
		amount2, _ := NewMoney("-900.00")
		amount3, _ := NewMoney("-1000.00")
		overrides := []*RecurringScheduleOverride{
			{
				EffectiveDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
				Amount:        amount1,
				Frequency:     FrequencyMonthly,
			},
			{
				EffectiveDate: time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
				Amount:        amount2,
				Frequency:     FrequencyMonthly,
			},
			{
				EffectiveDate: time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
				Amount:        amount3,
				Frequency:     FrequencyMonthly,
			},
		}

		// March: only first override applies
		amount, _ := EffectiveSchedule(baseAmount, baseFreq, overrides, 2026, time.March)
		if !amount.Equal(amount1) {
			t.Errorf("March: amount = %s, want %s", amount.String(), amount1.String())
		}

		// May: second override applies
		amount, _ = EffectiveSchedule(baseAmount, baseFreq, overrides, 2026, time.May)
		if !amount.Equal(amount2) {
			t.Errorf("May: amount = %s, want %s", amount.String(), amount2.String())
		}

		// August: third override applies
		amount, _ = EffectiveSchedule(baseAmount, baseFreq, overrides, 2026, time.August)
		if !amount.Equal(amount3) {
			t.Errorf("August: amount = %s, want %s", amount.String(), amount3.String())
		}
	})

	t.Run("override changes frequency", func(t *testing.T) {
		overrideAmount, _ := NewMoney("-400.00")
		overrides := []*RecurringScheduleOverride{
			{
				EffectiveDate: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
				Amount:        overrideAmount,
				Frequency:     FrequencyBiweekly,
			},
		}

		_, freq := EffectiveSchedule(baseAmount, baseFreq, overrides, 2026, time.April)
		if freq != FrequencyBiweekly {
			t.Errorf("freq = %s, want %s", freq, FrequencyBiweekly)
		}
	})

	t.Run("unsorted overrides are handled correctly", func(t *testing.T) {
		amount1, _ := NewMoney("-850.00")
		amount2, _ := NewMoney("-900.00")
		// Pass overrides in reverse order
		overrides := []*RecurringScheduleOverride{
			{
				EffectiveDate: time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
				Amount:        amount2,
				Frequency:     FrequencyMonthly,
			},
			{
				EffectiveDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
				Amount:        amount1,
				Frequency:     FrequencyMonthly,
			},
		}

		// March: first override (Jan) should apply
		amount, _ := EffectiveSchedule(baseAmount, baseFreq, overrides, 2026, time.March)
		if !amount.Equal(amount1) {
			t.Errorf("March: amount = %s, want %s", amount.String(), amount1.String())
		}

		// May: second override (Apr) should apply
		amount, _ = EffectiveSchedule(baseAmount, baseFreq, overrides, 2026, time.May)
		if !amount.Equal(amount2) {
			t.Errorf("May: amount = %s, want %s", amount.String(), amount2.String())
		}
	})

	t.Run("boundary last day of month", func(t *testing.T) {
		overrideAmount, _ := NewMoney("-999.00")
		// Override effective on the last day of January
		overrides := []*RecurringScheduleOverride{
			{
				EffectiveDate: time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC),
				Amount:        overrideAmount,
				Frequency:     FrequencyMonthly,
			},
		}

		// Should apply in January (effective date is within the month)
		amount, _ := EffectiveSchedule(baseAmount, baseFreq, overrides, 2026, time.January)
		if !amount.Equal(overrideAmount) {
			t.Errorf("January: amount = %s, want %s", amount.String(), overrideAmount.String())
		}
	})
}
