package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"icekalt.dev/money-tracker/internal/domain"
)

func (s *Server) handleListScheduleOverrides(c echo.Context) error {
	_, err := parseID(c, "id")
	if err != nil {
		return respondError(c, err)
	}

	recurringID, err := parseID(c, "recurringId")
	if err != nil {
		return respondError(c, err)
	}

	overrides, err := s.services.RecurringExpense.ListOverrides(c.Request().Context(), recurringID)
	if err != nil {
		return respondError(c, err)
	}

	resp := make([]ScheduleOverrideResponse, len(overrides))
	for i, o := range overrides {
		resp[i] = toScheduleOverrideResponse(o)
	}
	return c.JSON(http.StatusOK, resp)
}

func (s *Server) handleCreateScheduleOverride(c echo.Context) error {
	_, err := parseID(c, "id")
	if err != nil {
		return respondError(c, err)
	}

	recurringID, err := parseID(c, "recurringId")
	if err != nil {
		return respondError(c, err)
	}

	var req CreateScheduleOverrideRequest
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

	effectiveDate, err := time.Parse("2006-01-02", req.EffectiveDate)
	if err != nil {
		return respondError(c, fmt.Errorf("%w: invalid effective_date format", domain.ErrValidation))
	}

	override, err := s.services.RecurringExpense.CreateOverride(c.Request().Context(), recurringID, effectiveDate, amount, freq)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(http.StatusCreated, toScheduleOverrideResponse(override))
}

func (s *Server) handleUpdateScheduleOverride(c echo.Context) error {
	_, err := parseID(c, "id")
	if err != nil {
		return respondError(c, err)
	}

	overrideID, err := parseID(c, "overrideId")
	if err != nil {
		return respondError(c, err)
	}

	var req UpdateScheduleOverrideRequest
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

	effectiveDate, err := time.Parse("2006-01-02", req.EffectiveDate)
	if err != nil {
		return respondError(c, fmt.Errorf("%w: invalid effective_date format", domain.ErrValidation))
	}

	override, err := s.services.RecurringExpense.UpdateOverride(c.Request().Context(), overrideID, effectiveDate, amount, freq)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(http.StatusOK, toScheduleOverrideResponse(override))
}

func (s *Server) handleDeleteScheduleOverride(c echo.Context) error {
	_, err := parseID(c, "id")
	if err != nil {
		return respondError(c, err)
	}

	overrideID, err := parseID(c, "overrideId")
	if err != nil {
		return respondError(c, err)
	}

	if err := s.services.RecurringExpense.DeleteOverride(c.Request().Context(), overrideID); err != nil {
		return respondError(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}

func toScheduleOverrideResponse(o *domain.RecurringScheduleOverride) ScheduleOverrideResponse {
	return ScheduleOverrideResponse{
		ID:                 o.ID,
		RecurringExpenseID: o.RecurringExpenseID,
		EffectiveDate:      o.EffectiveDate.Format("2006-01-02"),
		Amount:             o.Amount.String(),
		Frequency:          string(o.Frequency),
		CreatedAt:          o.CreatedAt,
		UpdatedAt:          o.UpdatedAt,
	}
}
