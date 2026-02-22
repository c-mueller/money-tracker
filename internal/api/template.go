package api

import (
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/shopspring/decimal"
	"icekalt.dev/money-tracker/web"
)

type TemplateRenderer struct {
	templates map[string]*template.Template
}

func NewTemplateRenderer() (*TemplateRenderer, error) {
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
	}

	templatesFS, err := fs.Sub(web.Content, "templates")
	if err != nil {
		return nil, fmt.Errorf("accessing templates: %w", err)
	}

	layoutBytes, err := fs.ReadFile(templatesFS, "layout.html")
	if err != nil {
		return nil, fmt.Errorf("reading layout: %w", err)
	}

	pages := map[string]string{
		"dashboard":        "dashboard.html",
		"login":            "auth/login.html",
		"household_detail": "household/detail.html",
		"household_form":   "household/form.html",
		"category_list":    "category/list.html",
		"recurring_list":   "recurring/list.html",
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
		t, err = t.Parse(string(pageBytes))
		if err != nil {
			return nil, fmt.Errorf("parsing %s: %w", file, err)
		}
		templates[name] = t
	}

	return &TemplateRenderer{templates: templates}, nil
}

func (r *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	t, ok := r.templates[name]
	if !ok {
		return fmt.Errorf("template %s not found", name)
	}
	return t.ExecuteTemplate(w, "layout", data)
}
