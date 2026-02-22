package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"icekalt.dev/money-tracker/internal/domain"
)

func (s *Server) handleListRecurringExpenses(c echo.Context) error {
	householdID, err := parseID(c, "id")
	if err != nil {
		return respondError(c, err)
	}

	expenses, err := s.services.RecurringExpense.List(c.Request().Context(), householdID)
	if err != nil {
		return respondError(c, err)
	}

	resp := make([]RecurringExpenseResponse, len(expenses))
	for i, re := range expenses {
		resp[i] = toRecurringExpenseResponse(re)
	}
	return c.JSON(http.StatusOK, resp)
}

func (s *Server) handleCreateRecurringExpense(c echo.Context) error {
	householdID, err := parseID(c, "id")
	if err != nil {
		return respondError(c, err)
	}

	var req CreateRecurringExpenseRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
	}

	amount, err := domain.NewMoney(req.Amount)
	if err != nil {
		return respondError(c, fmt.Errorf("%w: invalid amount", domain.ErrValidation))
	}

	freq := domain.Frequency(req.Frequency)
	if err := freq.Validate(); err != nil {
		return respondError(c, err)
	}

	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return respondError(c, fmt.Errorf("%w: invalid start_date format", domain.ErrValidation))
	}

	var endDate *time.Time
	if req.EndDate != nil {
		t, err := time.Parse("2006-01-02", *req.EndDate)
		if err != nil {
			return respondError(c, fmt.Errorf("%w: invalid end_date format", domain.ErrValidation))
		}
		endDate = &t
	}

	re, err := s.services.RecurringExpense.Create(c.Request().Context(), householdID, req.CategoryID, req.Name, "", amount, freq, startDate, endDate)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(http.StatusCreated, toRecurringExpenseResponse(re))
}

func (s *Server) handleUpdateRecurringExpense(c echo.Context) error {
	_, err := parseID(c, "id")
	if err != nil {
		return respondError(c, err)
	}

	recurringID, err := parseID(c, "recurringId")
	if err != nil {
		return respondError(c, err)
	}

	var req UpdateRecurringExpenseRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
	}

	amount, err := domain.NewMoney(req.Amount)
	if err != nil {
		return respondError(c, fmt.Errorf("%w: invalid amount", domain.ErrValidation))
	}

	freq := domain.Frequency(req.Frequency)
	if err := freq.Validate(); err != nil {
		return respondError(c, err)
	}

	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return respondError(c, fmt.Errorf("%w: invalid start_date format", domain.ErrValidation))
	}

	var endDate *time.Time
	if req.EndDate != nil {
		t, err := time.Parse("2006-01-02", *req.EndDate)
		if err != nil {
			return respondError(c, fmt.Errorf("%w: invalid end_date format", domain.ErrValidation))
		}
		endDate = &t
	}

	re, err := s.services.RecurringExpense.Update(c.Request().Context(), recurringID, req.CategoryID, req.Name, "", amount, freq, req.Active, startDate, endDate)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(http.StatusOK, toRecurringExpenseResponse(re))
}

func (s *Server) handleDeleteRecurringExpense(c echo.Context) error {
	householdID, err := parseID(c, "id")
	if err != nil {
		return respondError(c, err)
	}

	recurringID, err := parseID(c, "recurringId")
	if err != nil {
		return respondError(c, err)
	}

	if err := s.services.RecurringExpense.Delete(c.Request().Context(), householdID, recurringID); err != nil {
		return respondError(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}

func toRecurringExpenseResponse(re *domain.RecurringExpense) RecurringExpenseResponse {
	resp := RecurringExpenseResponse{
		ID:          re.ID,
		HouseholdID: re.HouseholdID,
		CategoryID:  re.CategoryID,
		Name:        re.Name,
		Amount:      re.Amount.String(),
		Frequency:   string(re.Frequency),
		Active:      re.Active,
		StartDate:   re.StartDate.Format("2006-01-02"),
		CreatedAt:   re.CreatedAt,
		UpdatedAt:   re.UpdatedAt,
	}
	if re.EndDate != nil {
		s := re.EndDate.Format("2006-01-02")
		resp.EndDate = &s
	}
	return resp
}
