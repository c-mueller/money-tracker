package domain

import (
	"errors"
	"testing"
)

func TestFrequencyValid(t *testing.T) {
	tests := []struct {
		name string
		freq Frequency
		want bool
	}{
		{"daily", FrequencyDaily, true},
		{"weekday", FrequencyWeekday, true},
		{"weekly", FrequencyWeekly, true},
		{"biweekly", FrequencyBiweekly, true},
		{"monthly", FrequencyMonthly, true},
		{"quarterly", FrequencyQuarterly, true},
		{"yearly", FrequencyYearly, true},
		{"empty", Frequency(""), false},
		{"invalid", Frequency("hourly"), false},
		{"uppercase", Frequency("Daily"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.freq.Valid(); got != tt.want {
				t.Errorf("Frequency(%q).Valid() = %v, want %v", tt.freq, got, tt.want)
			}
		})
	}
}

func TestFrequencyValidate(t *testing.T) {
	tests := []struct {
		name    string
		freq    Frequency
		wantErr bool
	}{
		{"valid monthly", FrequencyMonthly, false},
		{"valid daily", FrequencyDaily, false},
		{"invalid", Frequency("invalid"), true},
		{"empty", Frequency(""), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.freq.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Frequency(%q).Validate() error = %v, wantErr %v", tt.freq, err, tt.wantErr)
			}
			if err != nil && !errors.Is(err, ErrValidation) {
				t.Errorf("expected ErrValidation, got %v", err)
			}
		})
	}
}

func TestAllFrequencies(t *testing.T) {
	freqs := AllFrequencies()
	if len(freqs) != 7 {
		t.Errorf("AllFrequencies() returned %d, want 7", len(freqs))
	}

	for _, f := range freqs {
		if !f.Valid() {
			t.Errorf("AllFrequencies() contains invalid frequency %q", f)
		}
	}
}
