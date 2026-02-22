package api

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"icekalt.dev/money-tracker/internal/domain"
)

func (s *Server) handleListHouseholds(c echo.Context) error {
	households, err := s.services.Household.List(c.Request().Context())
	if err != nil {
		return respondError(c, err)
	}

	resp := make([]HouseholdResponse, len(households))
	for i, h := range households {
		resp[i] = toHouseholdResponse(h)
	}
	return c.JSON(http.StatusOK, resp)
}

func (s *Server) handleCreateHousehold(c echo.Context) error {
	var req CreateHouseholdRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
	}

	if req.Currency == "" {
		req.Currency = "EUR"
	}

	h, err := s.services.Household.Create(c.Request().Context(), req.Name, req.Description, req.Currency, req.Icon)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(http.StatusCreated, toHouseholdResponse(h))
}

func (s *Server) handleUpdateHousehold(c echo.Context) error {
	id, err := parseID(c, "id")
	if err != nil {
		return respondError(c, err)
	}

	var req UpdateHouseholdRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
	}

	h, err := s.services.Household.Update(c.Request().Context(), id, req.Name, req.Description, req.Currency, req.Icon)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(http.StatusOK, toHouseholdResponse(h))
}

func (s *Server) handleDeleteHousehold(c echo.Context) error {
	id, err := parseID(c, "id")
	if err != nil {
		return respondError(c, err)
	}

	if err := s.services.Household.Delete(c.Request().Context(), id); err != nil {
		return respondError(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}

func toHouseholdResponse(h *domain.Household) HouseholdResponse {
	return HouseholdResponse{
		ID:          h.ID,
		Name:        h.Name,
		Description: h.Description,
		Currency:    h.Currency,
		Icon:        h.Icon,
		OwnerID:     h.OwnerID,
		CreatedAt:   h.CreatedAt,
		UpdatedAt:   h.UpdatedAt,
	}
}
