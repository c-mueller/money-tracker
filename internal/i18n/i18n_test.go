package i18n

import "testing"

func TestParseLocale(t *testing.T) {
	tests := []struct {
		input string
		want  Locale
	}{
		{"de", DE},
		{"de-DE", DE},
		{"de-AT", DE},
		{"en", EN},
		{"en-US", EN},
		{"en-GB", EN},
		{"DE", DE},
		{"EN", EN},
		{"fr", DE}, // unsupported → default DE
		{"", DE},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ParseLocale(tt.input)
			if got != tt.want {
				t.Errorf("ParseLocale(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseAcceptLanguage(t *testing.T) {
	tests := []struct {
		header string
		want   Locale
		ok     bool
	}{
		{"", DE, false},
		{"en", EN, true},
		{"de", DE, true},
		{"en-US,en;q=0.9,de;q=0.8", EN, true},
		{"de-DE,de;q=0.9,en;q=0.8", DE, true},
		{"fr-FR,fr;q=0.9", DE, false},                   // no match
		{"en;q=0.7,de;q=0.9", DE, true},                 // DE higher weight
		{"de;q=0.5,en;q=0.8", EN, true},                 // EN higher weight
		{"en-US,de-DE;q=0.9,fr;q=0.5", EN, true},
	}
	for _, tt := range tests {
		t.Run(tt.header, func(t *testing.T) {
			got, ok := ParseAcceptLanguage(tt.header)
			if got != tt.want || ok != tt.ok {
				t.Errorf("ParseAcceptLanguage(%q) = (%q, %v), want (%q, %v)", tt.header, got, ok, tt.want, tt.ok)
			}
		})
	}
}

func TestBundleT(t *testing.T) {
	b := NewBundle(DE)

	// German lookup
	if got := b.T(DE, "save"); got != "Speichern" {
		t.Errorf("T(DE, save) = %q, want Speichern", got)
	}

	// English lookup
	if got := b.T(EN, "save"); got != "Save" {
		t.Errorf("T(EN, save) = %q, want Save", got)
	}

	// Missing key returns key
	if got := b.T(DE, "nonexistent_key"); got != "nonexistent_key" {
		t.Errorf("T(DE, nonexistent_key) = %q, want nonexistent_key", got)
	}

	// Interpolation
	if got := b.T(DE, "delete_category_confirm", "Lebensmittel"); got != "Kategorie 'Lebensmittel' löschen?" {
		t.Errorf("T with interpolation = %q", got)
	}
}

func TestBundleTFallback(t *testing.T) {
	b := NewBundle(EN)

	// If a key exists only in EN (default) and we request DE, it should fall back
	// Both DE and EN have all keys, so let's test missing key fallback
	if got := b.T(DE, "unknown_key"); got != "unknown_key" {
		t.Errorf("expected key itself as fallback, got %q", got)
	}
}

func TestFrequencyName(t *testing.T) {
	b := NewBundle(DE)

	tests := []struct {
		locale Locale
		freq   string
		want   string
	}{
		{DE, "daily", "Täglich"},
		{DE, "monthly", "Monatlich"},
		{DE, "yearly", "Jährlich"},
		{EN, "daily", "Daily"},
		{EN, "monthly", "Monthly"},
		{EN, "weekly", "Weekly"},
		{DE, "unknown", "unknown"},
	}
	for _, tt := range tests {
		t.Run(string(tt.locale)+"_"+tt.freq, func(t *testing.T) {
			got := b.FrequencyName(tt.locale, tt.freq)
			if got != tt.want {
				t.Errorf("FrequencyName(%s, %s) = %q, want %q", tt.locale, tt.freq, got, tt.want)
			}
		})
	}
}

func TestDateFormat(t *testing.T) {
	b := NewBundle(DE)

	if got := b.DateFormat(DE); got != "02.01.2006" {
		t.Errorf("DateFormat(DE) = %q, want 02.01.2006", got)
	}
	if got := b.DateFormat(EN); got != "01/02/2006" {
		t.Errorf("DateFormat(EN) = %q, want 01/02/2006", got)
	}
}

func TestThousandsSep(t *testing.T) {
	b := NewBundle(DE)

	if got := b.ThousandsSep(DE); got != "." {
		t.Errorf("ThousandsSep(DE) = %q, want .", got)
	}
	if got := b.ThousandsSep(EN); got != "," {
		t.Errorf("ThousandsSep(EN) = %q, want ,", got)
	}
}

func TestDecimalSep(t *testing.T) {
	b := NewBundle(DE)

	if got := b.DecimalSep(DE); got != "," {
		t.Errorf("DecimalSep(DE) = %q, want ,", got)
	}
	if got := b.DecimalSep(EN); got != "." {
		t.Errorf("DecimalSep(EN) = %q, want .", got)
	}
}

func TestBundleParseLocale(t *testing.T) {
	b := NewBundle(DE)

	if got := b.ParseLocale("en-US"); got != EN {
		t.Errorf("bundle.ParseLocale(en-US) = %q, want en", got)
	}
	if got := b.ParseLocale("de-AT"); got != DE {
		t.Errorf("bundle.ParseLocale(de-AT) = %q, want de", got)
	}
	if got := b.ParseLocale("fr"); got != DE {
		t.Errorf("bundle.ParseLocale(fr) = %q, want de (default)", got)
	}
}
