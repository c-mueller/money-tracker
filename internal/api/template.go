package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/shopspring/decimal"
	"icekalt.dev/money-tracker/internal/i18n"
	mw "icekalt.dev/money-tracker/internal/middleware"
	"icekalt.dev/money-tracker/web"
)

type Currency struct {
	Code   string `json:"code"`
	Symbol string `json:"symbol"`
	Label  string `json:"label"`
}

type TemplateRenderer struct {
	templates      map[string]*template.Template
	Currencies     []Currency
	Icons          []string
	currencyByCode map[string]Currency
	bundle         *i18n.Bundle
	defaultLocale  i18n.Locale
}

func NewTemplateRenderer(bundle *i18n.Bundle, defaultLocale i18n.Locale) (*TemplateRenderer, error) {
	// Load currencies from embedded JSON
	currencyData, err := fs.ReadFile(web.Content, "static/currencies.json")
	if err != nil {
		return nil, fmt.Errorf("reading currencies.json: %w", err)
	}

	var currencies []Currency
	if err := json.Unmarshal(currencyData, &currencies); err != nil {
		return nil, fmt.Errorf("parsing currencies.json: %w", err)
	}

	currencyByCode := make(map[string]Currency, len(currencies))
	for _, c := range currencies {
		currencyByCode[c.Code] = c
	}

	// Load icons from embedded JSON
	iconData, err := fs.ReadFile(web.Content, "static/icons.json")
	if err != nil {
		return nil, fmt.Errorf("reading icons.json: %w", err)
	}

	var icons []string
	if err := json.Unmarshal(iconData, &icons); err != nil {
		return nil, fmt.Errorf("parsing icons.json: %w", err)
	}

	// Base funcMap with placeholder t/tf — overridden per-request in Render()
	funcMap := template.FuncMap{
		"t": func(key string, args ...interface{}) string {
			return bundle.T(defaultLocale, key, args...)
		},
		"tf": func(freq string) string {
			return bundle.FrequencyName(defaultLocale, freq)
		},
		"formatMoney": formatMoneyForLocale(defaultLocale),
		"formatDate":  formatDateForLocale(defaultLocale),
		"derefTime": func(t *time.Time) time.Time {
			if t == nil {
				return time.Time{}
			}
			return *t
		},
		"or": func(a, b string) string {
			if a != "" {
				return a
			}
			return b
		},
		"currencySymbol": func(code string) string {
			if c, ok := currencyByCode[code]; ok {
				return c.Symbol
			}
			return code
		},
		"isCurrencyCode": func(code string, currencies []Currency) bool {
			for _, c := range currencies {
				if c.Code == code {
					return true
				}
			}
			return false
		},
		"absAmount": func(d decimal.Decimal) string {
			return d.Abs().StringFixed(2)
		},
		"isNegative": func(d decimal.Decimal) bool {
			return d.IsNegative()
		},
		"formatDateISO": func(t time.Time) string {
			if t.IsZero() {
				return ""
			}
			return t.Format("2006-01-02")
		},
		"dict": func(pairs ...interface{}) map[string]interface{} {
			m := make(map[string]interface{}, len(pairs)/2)
			for i := 0; i+1 < len(pairs); i += 2 {
				m[pairs[i].(string)] = pairs[i+1]
			}
			return m
		},
	}

	templatesFS, err := fs.Sub(web.Content, "templates")
	if err != nil {
		return nil, fmt.Errorf("accessing templates: %w", err)
	}

	layoutBytes, err := fs.ReadFile(templatesFS, "layout.html")
	if err != nil {
		return nil, fmt.Errorf("reading layout: %w", err)
	}

	// Load partials (shared template fragments)
	partials := []string{
		"household/tabs.html",
		"partials/icon-picker.html",
	}
	var partialBytes [][]byte
	for _, p := range partials {
		b, err := fs.ReadFile(templatesFS, p)
		if err != nil {
			return nil, fmt.Errorf("reading partial %s: %w", p, err)
		}
		partialBytes = append(partialBytes, b)
	}

	pages := map[string]string{
		"dashboard":          "dashboard.html",
		"login":              "auth/login.html",
		"household_detail":   "household/detail.html",
		"household_form":     "household/form.html",
		"category_list":      "category/list.html",
		"category_form":      "category/form.html",
		"household_settings": "household/settings.html",
		"recurring_list":     "recurring/list.html",
		"recurring_form":     "recurring/form.html",
		"transaction_form":   "transaction/form.html",
		"token_list":         "token/list.html",
	}

	templates := make(map[string]*template.Template)
	for name, file := range pages {
		pageBytes, err := fs.ReadFile(templatesFS, file)
		if err != nil {
			return nil, fmt.Errorf("reading template %s: %w", file, err)
		}
		t, err := template.New("layout").Funcs(funcMap).Parse(string(layoutBytes))
		if err != nil {
			return nil, fmt.Errorf("parsing layout: %w", err)
		}
		for _, pb := range partialBytes {
			t, err = t.Parse(string(pb))
			if err != nil {
				return nil, fmt.Errorf("parsing partial for %s: %w", file, err)
			}
		}
		t, err = t.Parse(string(pageBytes))
		if err != nil {
			return nil, fmt.Errorf("parsing %s: %w", file, err)
		}
		templates[name] = t
	}

	return &TemplateRenderer{
		templates:      templates,
		Currencies:     currencies,
		Icons:          icons,
		currencyByCode: currencyByCode,
		bundle:         bundle,
		defaultLocale:  defaultLocale,
	}, nil
}

func (r *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	t, ok := r.templates[name]
	if !ok {
		return fmt.Errorf("template %s not found", name)
	}

	// Determine locale from context
	locale := r.defaultLocale
	if l, ok := c.Get(mw.LocaleContextKey).(i18n.Locale); ok {
		locale = l
	}

	// Clone template and override locale-sensitive functions
	tc, err := t.Clone()
	if err != nil {
		return fmt.Errorf("cloning template %s: %w", name, err)
	}

	tc.Funcs(template.FuncMap{
		"t": func(key string, args ...interface{}) string {
			return r.bundle.T(locale, key, args...)
		},
		"tf": func(freq string) string {
			return r.bundle.FrequencyName(locale, freq)
		},
		"formatMoney": formatMoneyForLocale(locale),
		"formatDate":  formatDateForLocale(locale),
	})

	var buf bytes.Buffer
	if err := tc.ExecuteTemplate(&buf, "layout", data); err != nil {
		return err
	}
	_, err = buf.WriteTo(w)
	return err
}

func formatMoneyForLocale(locale i18n.Locale) func(decimal.Decimal) string {
	return func(d decimal.Decimal) string {
		s := d.StringFixed(2)
		if locale == i18n.DE {
			// 1234.50 → 1.234,50
			s = strings.ReplaceAll(s, ".", "POINT")
			// We need to handle the grouping for DE locale
			parts := strings.SplitN(s, "POINT", 2)
			intPart := parts[0]
			decPart := parts[1]

			// Add thousands separator
			negative := false
			if len(intPart) > 0 && intPart[0] == '-' {
				negative = true
				intPart = intPart[1:]
			}

			if len(intPart) > 3 {
				var groups []string
				for len(intPart) > 3 {
					groups = append([]string{intPart[len(intPart)-3:]}, groups...)
					intPart = intPart[:len(intPart)-3]
				}
				groups = append([]string{intPart}, groups...)
				intPart = strings.Join(groups, ".")
			}

			if negative {
				return "-" + intPart + "," + decPart
			}
			return intPart + "," + decPart
		}
		// EN: 1234.50 → 1,234.50
		parts := strings.SplitN(s, ".", 2)
		intPart := parts[0]
		decPart := parts[1]

		negative := false
		if len(intPart) > 0 && intPart[0] == '-' {
			negative = true
			intPart = intPart[1:]
		}

		if len(intPart) > 3 {
			var groups []string
			for len(intPart) > 3 {
				groups = append([]string{intPart[len(intPart)-3:]}, groups...)
				intPart = intPart[:len(intPart)-3]
			}
			groups = append([]string{intPart}, groups...)
			intPart = strings.Join(groups, ",")
		}

		if negative {
			return "-" + intPart + "." + decPart
		}
		return intPart + "." + decPart
	}
}

func formatDateForLocale(locale i18n.Locale) func(time.Time) string {
	return func(t time.Time) string {
		if t.IsZero() {
			return ""
		}
		if locale == i18n.DE {
			return t.Format("02.01.2006")
		}
		return t.Format("01/02/2006")
	}
}
