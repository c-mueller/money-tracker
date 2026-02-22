package api

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"icekalt.dev/money-tracker/internal/domain"
	"icekalt.dev/money-tracker/internal/service"
)

type pageData struct {
	Title              string
	User               *domain.User
	Households         []*domain.Household
	Household          *domain.Household
	Category           *domain.Category
	Categories         []*domain.Category
	Transactions       []*domain.Transaction
	Transaction        *domain.Transaction
	RecurringExpenses  []*domain.RecurringExpense
	RecurringExpense   *domain.RecurringExpense
	Summary            *domain.MonthlySummary
	Tokens             []*domain.APIToken
	NewToken           string
	Month              string
	PrevMonth          string
	NextMonth          string
	Currencies         []Currency
	Icons              []string
	ActiveTab          string
	Frequencies        []domain.Frequency
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
		ActiveTab:    "transactions",
	})
}

func (s *Server) handleWebHouseholdNew(c echo.Context) error {
	return c.Render(http.StatusOK, "household_form", pageData{
		Title:      "New Household",
		User:       s.getUserFromContext(c),
		Household:  &domain.Household{Currency: "EUR"},
		Currencies: s.renderer.Currencies,
		Icons:      s.renderer.Icons,
	})
}

func (s *Server) handleWebHouseholdCreate(c echo.Context) error {
	name := c.FormValue("name")
	currency := c.FormValue("currency")
	if currency == "" {
		currency = "EUR"
	}
	icon := c.FormValue("icon")

	_, err := s.services.Household.Create(c.Request().Context(), name, currency, icon)
	if err != nil {
		return err
	}

	return c.Redirect(http.StatusFound, "/")
}

func (s *Server) handleWebTransactionNew(c echo.Context) error {
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

	month := c.QueryParam("month")
	if month == "" {
		now := time.Now()
		month = fmt.Sprintf("%d-%02d", now.Year(), now.Month())
	}

	return c.Render(http.StatusOK, "transaction_form", pageData{
		Title:      "New Transaction",
		User:       s.getUserFromContext(c),
		Household:  hh,
		Categories: categories,
		Month:      month,
	})
}

func (s *Server) handleWebTransactionCreate(c echo.Context) error {
	id, err := parseID(c, "id")
	if err != nil {
		return err
	}

	amount, err := domain.NewMoney(c.FormValue("amount"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid amount")
	}

	if c.FormValue("type") == "expense" {
		amount = amount.Neg()
	}

	categoryID, err := s.resolveCategory(c, id)
	if err != nil {
		return err
	}

	date, err := time.Parse("2006-01-02", c.FormValue("date"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid date")
	}

	description := c.FormValue("description")

	_, err = s.services.Transaction.Create(c.Request().Context(), id, categoryID, amount, description, date)
	if err != nil {
		return err
	}

	month := fmt.Sprintf("%d-%02d", date.Year(), date.Month())
	return c.Redirect(http.StatusFound, fmt.Sprintf("/households/%d?month=%s", id, month))
}

func (s *Server) handleWebTransactionEdit(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := parseID(c, "id")
	if err != nil {
		return err
	}

	txID, err := parseID(c, "transactionId")
	if err != nil {
		return err
	}

	hh, err := s.services.Household.GetByID(ctx, id)
	if err != nil {
		return err
	}

	tx, err := s.services.Transaction.GetByID(ctx, txID)
	if err != nil {
		return err
	}

	categories, err := s.services.Category.List(ctx, id)
	if err != nil {
		return err
	}

	month := fmt.Sprintf("%d-%02d", tx.Date.Year(), tx.Date.Month())

	return c.Render(http.StatusOK, "transaction_form", pageData{
		Title:       "Edit Transaction",
		User:        s.getUserFromContext(c),
		Household:   hh,
		Transaction: tx,
		Categories:  categories,
		Month:       month,
	})
}

func (s *Server) handleWebTransactionUpdate(c echo.Context) error {
	id, err := parseID(c, "id")
	if err != nil {
		return err
	}

	txID, err := parseID(c, "transactionId")
	if err != nil {
		return err
	}

	amount, err := domain.NewMoney(c.FormValue("amount"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid amount")
	}

	if c.FormValue("type") == "expense" {
		amount = amount.Neg()
	}

	categoryID, err := s.resolveCategory(c, id)
	if err != nil {
		return err
	}

	date, err := time.Parse("2006-01-02", c.FormValue("date"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid date")
	}

	description := c.FormValue("description")

	_, err = s.services.Transaction.Update(c.Request().Context(), id, txID, categoryID, amount, description, date)
	if err != nil {
		return err
	}

	month := fmt.Sprintf("%d-%02d", date.Year(), date.Month())
	return c.Redirect(http.StatusFound, fmt.Sprintf("/households/%d?month=%s", id, month))
}

func (s *Server) handleWebCategoryList(c echo.Context) error {
	id, err := parseID(c, "id")
	if err != nil {
		return err
	}
	return c.Redirect(http.StatusFound, fmt.Sprintf("/households/%d/settings", id))
}

func (s *Server) handleWebCategoryCreate(c echo.Context) error {
	id, err := parseID(c, "id")
	if err != nil {
		return err
	}

	name := c.FormValue("name")
	icon := c.FormValue("icon")
	_, err = s.services.Category.Create(c.Request().Context(), id, name, icon)
	if err != nil {
		return err
	}

	return c.Redirect(http.StatusFound, fmt.Sprintf("/households/%d/settings", id))
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

	now := time.Now()
	month := fmt.Sprintf("%d-%02d", now.Year(), now.Month())

	return c.Render(http.StatusOK, "recurring_list", pageData{
		Title:             "Recurring Transactions",
		User:              s.getUserFromContext(c),
		Household:         hh,
		RecurringExpenses: expenses,
		Month:             month,
		ActiveTab:         "recurring",
	})
}

func (s *Server) handleWebRecurringNew(c echo.Context) error {
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

	return c.Render(http.StatusOK, "recurring_form", pageData{
		Title:       "New Recurring Transaction",
		User:        s.getUserFromContext(c),
		Household:   hh,
		Categories:  categories,
		Frequencies: domain.AllFrequencies(),
	})
}

func (s *Server) handleWebRecurringCreate(c echo.Context) error {
	id, err := parseID(c, "id")
	if err != nil {
		return err
	}

	// Handle new category creation
	categoryID, err := s.resolveCategory(c, id)
	if err != nil {
		return err
	}

	amount, err := domain.NewMoney(c.FormValue("amount"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid amount")
	}

	if c.FormValue("type") == "expense" {
		amount = amount.Neg()
	}

	freq := domain.Frequency(c.FormValue("frequency"))

	startDate, err := time.Parse("2006-01-02", c.FormValue("start_date"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid start date")
	}

	var endDate *time.Time
	if ed := c.FormValue("end_date"); ed != "" {
		t, err := time.Parse("2006-01-02", ed)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid end date")
		}
		endDate = &t
	}

	name := c.FormValue("name")
	description := c.FormValue("description")

	_, err = s.services.RecurringExpense.Create(c.Request().Context(), id, categoryID, name, description, amount, freq, startDate, endDate)
	if err != nil {
		return err
	}

	return c.Redirect(http.StatusFound, fmt.Sprintf("/households/%d/recurring", id))
}

func (s *Server) handleWebRecurringEdit(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := parseID(c, "id")
	if err != nil {
		return err
	}

	recurringID, err := parseID(c, "recurringId")
	if err != nil {
		return err
	}

	hh, err := s.services.Household.GetByID(ctx, id)
	if err != nil {
		return err
	}

	re, err := s.services.RecurringExpense.GetByID(ctx, recurringID)
	if err != nil {
		return err
	}

	categories, err := s.services.Category.List(ctx, id)
	if err != nil {
		return err
	}

	return c.Render(http.StatusOK, "recurring_form", pageData{
		Title:            "Edit Recurring Transaction",
		User:             s.getUserFromContext(c),
		Household:        hh,
		RecurringExpense: re,
		Categories:       categories,
		Frequencies:      domain.AllFrequencies(),
	})
}

func (s *Server) handleWebRecurringUpdate(c echo.Context) error {
	id, err := parseID(c, "id")
	if err != nil {
		return err
	}

	recurringID, err := parseID(c, "recurringId")
	if err != nil {
		return err
	}

	categoryID, err := s.resolveCategory(c, id)
	if err != nil {
		return err
	}

	amount, err := domain.NewMoney(c.FormValue("amount"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid amount")
	}

	if c.FormValue("type") == "expense" {
		amount = amount.Neg()
	}

	freq := domain.Frequency(c.FormValue("frequency"))

	startDate, err := time.Parse("2006-01-02", c.FormValue("start_date"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid start date")
	}

	var endDate *time.Time
	if ed := c.FormValue("end_date"); ed != "" {
		t, err := time.Parse("2006-01-02", ed)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid end date")
		}
		endDate = &t
	}

	name := c.FormValue("name")
	description := c.FormValue("description")
	active := c.FormValue("active") == "on"

	_, err = s.services.RecurringExpense.Update(c.Request().Context(), recurringID, categoryID, name, description, amount, freq, active, startDate, endDate)
	if err != nil {
		return err
	}

	return c.Redirect(http.StatusFound, fmt.Sprintf("/households/%d/recurring", id))
}

func (s *Server) handleWebHouseholdSettings(c echo.Context) error {
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

	now := time.Now()
	month := fmt.Sprintf("%d-%02d", now.Year(), now.Month())

	return c.Render(http.StatusOK, "household_settings", pageData{
		Title:      "Settings",
		User:       s.getUserFromContext(c),
		Household:  hh,
		Categories: categories,
		Currencies: s.renderer.Currencies,
		Icons:      s.renderer.Icons,
		Month:      month,
		ActiveTab:  "settings",
	})
}

func (s *Server) handleWebHouseholdSettingsUpdate(c echo.Context) error {
	id, err := parseID(c, "id")
	if err != nil {
		return err
	}

	name := c.FormValue("name")
	currency := c.FormValue("currency")
	icon := c.FormValue("icon")

	_, err = s.services.Household.Update(c.Request().Context(), id, name, currency, icon)
	if err != nil {
		return err
	}

	return c.Redirect(http.StatusFound, fmt.Sprintf("/households/%d/settings", id))
}

func (s *Server) handleWebCategoryEdit(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := parseID(c, "id")
	if err != nil {
		return err
	}

	categoryID, err := parseID(c, "categoryId")
	if err != nil {
		return err
	}

	hh, err := s.services.Household.GetByID(ctx, id)
	if err != nil {
		return err
	}

	cat, err := s.services.Category.GetByID(ctx, categoryID)
	if err != nil {
		return err
	}

	now := time.Now()
	month := fmt.Sprintf("%d-%02d", now.Year(), now.Month())

	return c.Render(http.StatusOK, "category_form", pageData{
		Title:     "Edit Category",
		User:      s.getUserFromContext(c),
		Household: hh,
		Category:  cat,
		Icons:     s.renderer.Icons,
		Month:     month,
		ActiveTab: "settings",
	})
}

func (s *Server) handleWebCategoryUpdate(c echo.Context) error {
	id, err := parseID(c, "id")
	if err != nil {
		return err
	}

	categoryID, err := parseID(c, "categoryId")
	if err != nil {
		return err
	}

	name := c.FormValue("name")
	icon := c.FormValue("icon")

	_, err = s.services.Category.Update(c.Request().Context(), categoryID, name, icon)
	if err != nil {
		return err
	}

	return c.Redirect(http.StatusFound, fmt.Sprintf("/households/%d/settings", id))
}

// resolveCategory returns the category ID from the form, creating a new category if "NEW" was selected.
func (s *Server) resolveCategory(c echo.Context, householdID int) (int, error) {
	catVal := c.FormValue("category_id")
	if catVal == "NEW" {
		name := c.FormValue("new_category_name")
		if name == "" {
			return 0, echo.NewHTTPError(http.StatusBadRequest, "category name required")
		}
		cat, err := s.services.Category.Create(c.Request().Context(), householdID, name, "category")
		if err != nil {
			return 0, err
		}
		return cat.ID, nil
	}
	categoryID, err := strconv.Atoi(catVal)
	if err != nil {
		return 0, echo.NewHTTPError(http.StatusBadRequest, "invalid category")
	}
	return categoryID, nil
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
