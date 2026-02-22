package api

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/shopspring/decimal"
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
	currencyByCode map[string]Currency
}

func NewTemplateRenderer() (*TemplateRenderer, error) {
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

	funcMap := template.FuncMap{
		"formatMoney": func(d decimal.Decimal) string {
			return d.StringFixed(2)
		},
		"formatDate": func(t time.Time) string {
			return t.Format("2006-01-02")
		},
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
		"dashboard":        "dashboard.html",
		"login":            "auth/login.html",
		"household_detail": "household/detail.html",
		"household_form":   "household/form.html",
		"category_list":    "category/list.html",
		"recurring_list":   "recurring/list.html",
		"recurring_form":   "recurring/form.html",
		"transaction_form": "transaction/form.html",
		"token_list":       "token/list.html",
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
		currencyByCode: currencyByCode,
	}, nil
}

func (r *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	t, ok := r.templates[name]
	if !ok {
		return fmt.Errorf("template %s not found", name)
	}
	return t.ExecuteTemplate(w, "layout", data)
}
