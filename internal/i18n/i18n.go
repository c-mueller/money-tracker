package i18n

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"strings"
)

//go:embed locales/*.json
var localesFS embed.FS

// Locale represents a supported language.
type Locale string

const (
	DE Locale = "de"
	EN Locale = "en"
)

// localeData represents the JSON structure for a single locale file.
type localeData struct {
	Locale       string            `json:"locale"`
	DateFormat   string            `json:"date_format"`
	ThousandsSep string            `json:"thousands_sep"`
	DecimalSep   string            `json:"decimal_sep"`
	Frequencies  map[string]string `json:"frequencies"`
	Messages     map[string]string `json:"messages"`
}

// Bundle holds translations for all supported locales.
type Bundle struct {
	locales       map[Locale]*localeData
	defaultLocale Locale
}

// NewBundle creates a new translation bundle by loading all JSON locale files.
func NewBundle(defaultLocale Locale) *Bundle {
	b := &Bundle{
		locales:       make(map[Locale]*localeData),
		defaultLocale: defaultLocale,
	}

	matches, err := fs.Glob(localesFS, "locales/*.json")
	if err != nil {
		panic(fmt.Sprintf("i18n: failed to glob locale files: %v", err))
	}

	for _, path := range matches {
		data, err := fs.ReadFile(localesFS, path)
		if err != nil {
			panic(fmt.Sprintf("i18n: failed to read %s: %v", path, err))
		}

		var ld localeData
		if err := json.Unmarshal(data, &ld); err != nil {
			panic(fmt.Sprintf("i18n: failed to parse %s: %v", path, err))
		}

		b.locales[Locale(ld.Locale)] = &ld
	}

	return b
}

// T returns the translated string for the given key and locale.
// If args are provided, fmt.Sprintf is used for interpolation.
// Falls back to default locale, then returns the key itself.
func (b *Bundle) T(locale Locale, key string, args ...interface{}) string {
	if ld, ok := b.locales[locale]; ok {
		if msg, ok := ld.Messages[key]; ok {
			if len(args) > 0 {
				return fmt.Sprintf(msg, args...)
			}
			return msg
		}
	}
	// Fallback to default locale
	if locale != b.defaultLocale {
		if ld, ok := b.locales[b.defaultLocale]; ok {
			if msg, ok := ld.Messages[key]; ok {
				if len(args) > 0 {
					return fmt.Sprintf(msg, args...)
				}
				return msg
			}
		}
	}
	return key
}

// FrequencyName returns the localized display name for a frequency.
func (b *Bundle) FrequencyName(locale Locale, freq string) string {
	if ld, ok := b.locales[locale]; ok {
		if name, ok := ld.Frequencies[freq]; ok {
			return name
		}
	}
	// Fallback to default locale
	if ld, ok := b.locales[b.defaultLocale]; ok {
		if name, ok := ld.Frequencies[freq]; ok {
			return name
		}
	}
	return freq
}

// DateFormat returns the date format string for the given locale.
func (b *Bundle) DateFormat(locale Locale) string {
	if ld, ok := b.locales[locale]; ok {
		return ld.DateFormat
	}
	if ld, ok := b.locales[b.defaultLocale]; ok {
		return ld.DateFormat
	}
	return "2006-01-02"
}

// ThousandsSep returns the thousands separator for the given locale.
func (b *Bundle) ThousandsSep(locale Locale) string {
	if ld, ok := b.locales[locale]; ok {
		return ld.ThousandsSep
	}
	if ld, ok := b.locales[b.defaultLocale]; ok {
		return ld.ThousandsSep
	}
	return ","
}

// DecimalSep returns the decimal separator for the given locale.
func (b *Bundle) DecimalSep(locale Locale) string {
	if ld, ok := b.locales[locale]; ok {
		return ld.DecimalSep
	}
	if ld, ok := b.locales[b.defaultLocale]; ok {
		return ld.DecimalSep
	}
	return "."
}

// DefaultLocale returns the bundle's default locale.
func (b *Bundle) DefaultLocale() Locale {
	return b.defaultLocale
}

// ParseLocale normalizes a locale string to a supported Locale.
func (b *Bundle) ParseLocale(s string) Locale {
	s = strings.ToLower(strings.TrimSpace(s))
	for code := range b.locales {
		if strings.HasPrefix(s, string(code)) {
			return code
		}
	}
	return b.defaultLocale
}

// ParseLocale normalizes a locale string to a supported Locale (package-level).
func ParseLocale(s string) Locale {
	s = strings.ToLower(strings.TrimSpace(s))
	switch {
	case strings.HasPrefix(s, "de"):
		return DE
	case strings.HasPrefix(s, "en"):
		return EN
	default:
		return DE
	}
}

// ParseAcceptLanguage parses an Accept-Language header and returns the best matching locale.
func ParseAcceptLanguage(header string) (Locale, bool) {
	if header == "" {
		return DE, false
	}

	// Parse tags like "en-US,en;q=0.9,de;q=0.8"
	parts := strings.Split(header, ",")
	type langWeight struct {
		locale Locale
		weight float64
	}

	var matches []langWeight
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Split on ;q=
		var lang string
		weight := 1.0
		if idx := strings.Index(part, ";"); idx >= 0 {
			lang = strings.TrimSpace(part[:idx])
			qPart := strings.TrimSpace(part[idx+1:])
			if strings.HasPrefix(qPart, "q=") {
				fmt.Sscanf(qPart[2:], "%f", &weight)
			}
		} else {
			lang = part
		}

		lang = strings.ToLower(lang)
		switch {
		case strings.HasPrefix(lang, "de"):
			matches = append(matches, langWeight{DE, weight})
		case strings.HasPrefix(lang, "en"):
			matches = append(matches, langWeight{EN, weight})
		}
	}

	if len(matches) == 0 {
		return DE, false
	}

	// Return highest weight match
	best := matches[0]
	for _, m := range matches[1:] {
		if m.weight > best.weight {
			best = m
		}
	}
	return best.locale, true
}
