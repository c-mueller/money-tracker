package api

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"icekalt.dev/money-tracker/internal/domain"
)

func (s *Server) handleGetSummary(c echo.Context) error {
	householdID, err := parseID(c, "id")
	if err != nil {
		return respondError(c, err)
	}

	year, month, err := parseMonth(c)
	if err != nil {
		return respondError(c, err)
	}

	summary, err := s.services.Summary.GetMonthlySummary(c.Request().Context(), householdID, year, month)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(http.StatusOK, toSummaryResponse(summary))
}

func toSummaryResponse(s *domain.MonthlySummary) SummaryResponse {
	breakdown := make([]CategorySummaryResponse, len(s.CategoryBreakdown))
	for i, cs := range s.CategoryBreakdown {
		breakdown[i] = CategorySummaryResponse{
			CategoryID:   cs.CategoryID,
			CategoryName: cs.CategoryName,
			Recurring:    cs.Recurring.String(),
			OneTime:      cs.OneTime.String(),
			Total:        cs.Total.String(),
		}
	}

	return SummaryResponse{
		Month:             s.Month,
		HouseholdID:       s.HouseholdID,
		TotalIncome:       s.TotalIncome.String(),
		TotalExpenses:     s.TotalExpenses.String(),
		RecurringTotal:    s.RecurringTotal.String(),
		OneTimeTotal:      s.OneTimeTotal.String(),
		CategoryBreakdown: breakdown,
	}
}
