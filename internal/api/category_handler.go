package api

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"icekalt.dev/money-tracker/internal/domain"
)

func (s *Server) handleListCategories(c echo.Context) error {
	householdID, err := parseID(c, "id")
	if err != nil {
		return respondError(c, err)
	}

	categories, err := s.services.Category.List(c.Request().Context(), householdID)
	if err != nil {
		return respondError(c, err)
	}

	resp := make([]CategoryResponse, len(categories))
	for i, cat := range categories {
		resp[i] = toCategoryResponse(cat)
	}
	return c.JSON(http.StatusOK, resp)
}

func (s *Server) handleCreateCategory(c echo.Context) error {
	householdID, err := parseID(c, "id")
	if err != nil {
		return respondError(c, err)
	}

	var req CreateCategoryRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
	}

	cat, err := s.services.Category.Create(c.Request().Context(), householdID, req.Name, "")
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(http.StatusCreated, toCategoryResponse(cat))
}

func (s *Server) handleUpdateCategory(c echo.Context) error {
	_, err := parseID(c, "id")
	if err != nil {
		return respondError(c, err)
	}

	categoryID, err := parseID(c, "categoryId")
	if err != nil {
		return respondError(c, err)
	}

	var req CreateCategoryRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
	}

	cat, err := s.services.Category.Update(c.Request().Context(), categoryID, req.Name, "")
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(http.StatusOK, toCategoryResponse(cat))
}

func (s *Server) handleDeleteCategory(c echo.Context) error {
	householdID, err := parseID(c, "id")
	if err != nil {
		return respondError(c, err)
	}

	categoryID, err := parseID(c, "categoryId")
	if err != nil {
		return respondError(c, err)
	}

	if err := s.services.Category.Delete(c.Request().Context(), householdID, categoryID); err != nil {
		return respondError(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}

func toCategoryResponse(cat *domain.Category) CategoryResponse {
	return CategoryResponse{
		ID:          cat.ID,
		HouseholdID: cat.HouseholdID,
		Name:        cat.Name,
		CreatedAt:   cat.CreatedAt,
		UpdatedAt:   cat.UpdatedAt,
	}
}
