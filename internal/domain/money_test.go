package domain

import (
	"testing"
)

func TestNewMoney(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		want    string
		wantErr bool
	}{
		{"positive integer", "100", "100", false},
		{"positive decimal", "100.50", "100.5", false},
		{"negative", "-50.00", "-50", false},
		{"zero", "0", "0", false},
		{"large number", "999999999.99", "999999999.99", false},
		{"small decimal", "0.01", "0.01", false},
		{"invalid string", "abc", "", true},
		{"empty string", "", "", true},
		{"comma decimal", "12,50", "12.5", false},
		{"comma zero", "0,01", "0.01", false},
		{"comma large", "1234,56", "1234.56", false},
		{"double dot", "1.2.3", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewMoney(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewMoney(%q) error = %v, wantErr %v", tt.value, err, tt.wantErr)
				return
			}
			if err == nil && got.String() != tt.want {
				t.Errorf("NewMoney(%q) = %s, want %s", tt.value, got.String(), tt.want)
			}
		})
	}
}

func TestMoneyFromInt(t *testing.T) {
	tests := []struct {
		name  string
		cents int64
		want  string
	}{
		{"one dollar", 100, "1.00"},
		{"zero", 0, "0.00"},
		{"negative", -5000, "-50.00"},
		{"small", 1, "0.01"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MoneyFromInt(tt.cents)
			if got.StringFixed(2) != tt.want {
				t.Errorf("MoneyFromInt(%d) = %s, want %s", tt.cents, got.StringFixed(2), tt.want)
			}
		})
	}
}

func TestZeroMoney(t *testing.T) {
	z := ZeroMoney()
	if !z.IsZero() {
		t.Errorf("ZeroMoney() = %s, want 0", z.String())
	}
}
