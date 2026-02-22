package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"icekalt.dev/money-tracker/internal/domain"
	"icekalt.dev/money-tracker/internal/service"
)

type pageData struct {
	Title             string
	User              *domain.User
	Households        []*domain.Household
	Household         *domain.Household
	Categories        []*domain.Category
	Transactions      []*domain.Transaction
	RecurringExpenses []*domain.RecurringExpense
	Summary           *domain.MonthlySummary
	Tokens            []*domain.APIToken
	NewToken          string
	Month             string
	PrevMonth         string
	NextMonth         string
	Currencies        []Currency
}

func (s *Server) handleWebDashboard(c echo.Context) error {
	ctx := c.Request().Context()
	households, err := s.services.Household.List(ctx)
	if err != nil {
		return err
	}

	return c.Render(http.StatusOK, "dashboard", pageData{
		Title:      "Dashboard",
		User:       s.getUserFromContext(c),
		Households: households,
	})
}

func (s *Server) handleWebHouseholdDetail(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := parseID(c, "id")
	if err != nil {
		return err
	}

	hh, err := s.services.Household.GetByID(ctx, id)
	if err != nil {
		return err
	}

	year, month, err := parseMonth(c)
	if err != nil {
		return err
	}

	monthStr := fmt.Sprintf("%d-%02d", year, month)
	ref := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	prev := ref.AddDate(0, -1, 0)
	next := ref.AddDate(0, 1, 0)

	transactions, err := s.services.Transaction.ListByMonth(ctx, id, year, month)
	if err != nil {
		return err
	}

	summary, err := s.services.Summary.GetMonthlySummary(ctx, id, year, month)
	if err != nil {
		return err
	}

	return c.Render(http.StatusOK, "household_detail", pageData{
		Title:        hh.Name,
		User:         s.getUserFromContext(c),
		Household:    hh,
		Transactions: transactions,
		Summary:      summary,
		Month:        monthStr,
		PrevMonth:    fmt.Sprintf("%d-%02d", prev.Year(), prev.Month()),
		NextMonth:    fmt.Sprintf("%d-%02d", next.Year(), next.Month()),
	})
}

func (s *Server) handleWebHouseholdNew(c echo.Context) error {
	return c.Render(http.StatusOK, "household_form", pageData{
		Title:      "New Household",
		User:       s.getUserFromContext(c),
		Household:  &domain.Household{Currency: "EUR"},
		Currencies: s.renderer.Currencies,
	})
}

func (s *Server) handleWebHouseholdCreate(c echo.Context) error {
	name := c.FormValue("name")
	currency := c.FormValue("currency")
	if currency == "" {
		currency = "EUR"
	}

	_, err := s.services.Household.Create(c.Request().Context(), name, currency)
	if err != nil {
		return err
	}

	return c.Redirect(http.StatusFound, "/")
}

func (s *Server) handleWebCategoryList(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := parseID(c, "id")
	if err != nil {
		return err
	}

	hh, err := s.services.Household.GetByID(ctx, id)
	if err != nil {
		return err
	}

	categories, err := s.services.Category.List(ctx, id)
	if err != nil {
		return err
	}

	return c.Render(http.StatusOK, "category_list", pageData{
		Title:      "Categories",
		User:       s.getUserFromContext(c),
		Household:  hh,
		Categories: categories,
	})
}

func (s *Server) handleWebCategoryCreate(c echo.Context) error {
	id, err := parseID(c, "id")
	if err != nil {
		return err
	}

	name := c.FormValue("name")
	_, err = s.services.Category.Create(c.Request().Context(), id, name)
	if err != nil {
		return err
	}

	return c.Redirect(http.StatusFound, fmt.Sprintf("/households/%d/categories", id))
}

func (s *Server) handleWebRecurringList(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := parseID(c, "id")
	if err != nil {
		return err
	}

	hh, err := s.services.Household.GetByID(ctx, id)
	if err != nil {
		return err
	}

	expenses, err := s.services.RecurringExpense.List(ctx, id)
	if err != nil {
		return err
	}

	return c.Render(http.StatusOK, "recurring_list", pageData{
		Title:             "Recurring Expenses",
		User:              s.getUserFromContext(c),
		Household:         hh,
		RecurringExpenses: expenses,
	})
}

func (s *Server) handleWebTokenList(c echo.Context) error {
	tokens, err := s.services.APIToken.List(c.Request().Context())
	if err != nil {
		return err
	}

	return c.Render(http.StatusOK, "token_list", pageData{
		Title:  "API Tokens",
		User:   s.getUserFromContext(c),
		Tokens: tokens,
	})
}

func (s *Server) handleWebTokenCreate(c echo.Context) error {
	name := c.FormValue("name")
	plaintext, _, err := s.services.APIToken.Create(c.Request().Context(), name)
	if err != nil {
		return err
	}

	tokens, err := s.services.APIToken.List(c.Request().Context())
	if err != nil {
		return err
	}

	return c.Render(http.StatusOK, "token_list", pageData{
		Title:    "API Tokens",
		User:     s.getUserFromContext(c),
		Tokens:   tokens,
		NewToken: plaintext,
	})
}

func (s *Server) getUserFromContext(c echo.Context) *domain.User {
	userID, ok := service.UserIDFromContext(c.Request().Context())
	if !ok {
		return nil
	}
	user, err := s.services.User.GetByID(c.Request().Context(), userID)
	if err != nil {
		return nil
	}
	return user
}
