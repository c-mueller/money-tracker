package api

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"icekalt.dev/money-tracker/internal/domain"
)

type ErrorResponse struct {
	Error   string            `json:"error"`
	Details map[string]string `json:"details,omitempty"`
}

func respondError(c echo.Context, err error) error {
	var ve *domain.ValidationError
	if errors.As(err, &ve) {
		return c.JSON(http.StatusUnprocessableEntity, ErrorResponse{
			Error:   "validation error",
			Details: map[string]string{ve.Field: ve.Message},
		})
	}

	switch {
	case errors.Is(err, domain.ErrValidation):
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	case errors.Is(err, domain.ErrNotFound):
		return c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
	case errors.Is(err, domain.ErrForbidden):
		return c.JSON(http.StatusForbidden, ErrorResponse{Error: err.Error()})
	case errors.Is(err, domain.ErrConflict):
		return c.JSON(http.StatusConflict, ErrorResponse{Error: err.Error()})
	default:
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
	}
}
