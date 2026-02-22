package domain

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

func TestNormalizeToMonthly(t *testing.T) {
	amount := decimal.NewFromInt(120)
	jan2026 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	feb2026 := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	feb2024 := time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC) // leap year

	tests := []struct {
		name     string
		amount   Money
		freq     Frequency
		ref      time.Time
		expected string
	}{
		{"monthly", amount, FrequencyMonthly, jan2026, "120"},
		{"quarterly", amount, FrequencyQuarterly, jan2026, "40"},
		{"yearly", amount, FrequencyYearly, jan2026, "10"},
		{"weekly", amount, FrequencyWeekly, jan2026, "520"},
		{"biweekly", amount, FrequencyBiweekly, jan2026, "260"},
		// Jan 2026: 31 days
		{"daily_jan", decimal.NewFromInt(10), FrequencyDaily, jan2026, "310"},
		// Feb 2026: 28 days
		{"daily_feb", decimal.NewFromInt(10), FrequencyDaily, feb2026, "280"},
		// Feb 2024 (leap year): 29 days
		{"daily_feb_leap", decimal.NewFromInt(10), FrequencyDaily, feb2024, "290"},
		// Jan 2026: starts Thursday → 22 weekdays
		{"weekday_jan", decimal.NewFromInt(10), FrequencyWeekday, jan2026, "220"},
		// Feb 2026: starts Sunday → 20 weekdays
		{"weekday_feb", decimal.NewFromInt(10), FrequencyWeekday, feb2026, "200"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := NormalizeToMonthly(tt.amount, tt.freq, tt.ref)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			expected, _ := decimal.NewFromString(tt.expected)
			if !result.Equal(expected) {
				t.Errorf("expected %s, got %s", tt.expected, result.String())
			}
		})
	}
}

func TestNormalizeToMonthlyInvalidFrequency(t *testing.T) {
	_, err := NormalizeToMonthly(decimal.NewFromInt(100), Frequency("invalid"), time.Now())
	if err == nil {
		t.Error("expected error for invalid frequency")
	}
}
