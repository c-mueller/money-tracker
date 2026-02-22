package api

import (
	"fmt"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"icekalt.dev/money-tracker/internal/domain"
)

func parseID(c echo.Context, param string) (int, error) {
	id, err := strconv.Atoi(c.Param(param))
	if err != nil {
		return 0, fmt.Errorf("%w: invalid id parameter", domain.ErrValidation)
	}
	return id, nil
}

func parseMonth(c echo.Context) (int, time.Month, error) {
	monthStr := c.QueryParam("month")
	if monthStr == "" {
		now := time.Now()
		return now.Year(), now.Month(), nil
	}

	t, err := time.Parse("2006-01", monthStr)
	if err != nil {
		return 0, 0, fmt.Errorf("%w: invalid month format, expected YYYY-MM", domain.ErrValidation)
	}

	return t.Year(), t.Month(), nil
}
