package i18n

import (
	"fmt"
	"strings"
)

// Locale represents a supported language.
type Locale string

const (
	DE Locale = "de"
	EN Locale = "en"
)

// Bundle holds translations for all supported locales.
type Bundle struct {
	translations  map[Locale]map[string]string
	defaultLocale Locale
}

// NewBundle creates a new translation bundle with the given default locale.
func NewBundle(defaultLocale Locale) *Bundle {
	b := &Bundle{
		translations:  make(map[Locale]map[string]string),
		defaultLocale: defaultLocale,
	}
	b.translations[DE] = deTranslations
	b.translations[EN] = enTranslations
	return b
}

// T returns the translated string for the given key and locale.
// If args are provided, fmt.Sprintf is used for interpolation.
// Falls back to default locale, then returns the key itself.
func (b *Bundle) T(locale Locale, key string, args ...interface{}) string {
	if msgs, ok := b.translations[locale]; ok {
		if msg, ok := msgs[key]; ok {
			if len(args) > 0 {
				return fmt.Sprintf(msg, args...)
			}
			return msg
		}
	}
	// Fallback to default locale
	if locale != b.defaultLocale {
		if msgs, ok := b.translations[b.defaultLocale]; ok {
			if msg, ok := msgs[key]; ok {
				if len(args) > 0 {
					return fmt.Sprintf(msg, args...)
				}
				return msg
			}
		}
	}
	return key
}

// DefaultLocale returns the bundle's default locale.
func (b *Bundle) DefaultLocale() Locale {
	return b.defaultLocale
}

// ParseLocale normalizes a locale string to a supported Locale.
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
