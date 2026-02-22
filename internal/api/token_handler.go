package api

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

type CreateTokenRequest struct {
	Name string `json:"name"`
}

type TokenResponse struct {
	ID        int        `json:"id"`
	Name      string     `json:"name"`
	Token     string     `json:"token,omitempty"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	LastUsed  *time.Time `json:"last_used,omitempty"`
}

func (s *Server) handleListTokens(c echo.Context) error {
	tokens, err := s.services.APIToken.List(c.Request().Context())
	if err != nil {
		return respondError(c, err)
	}

	resp := make([]TokenResponse, len(tokens))
	for i, t := range tokens {
		resp[i] = TokenResponse{
			ID:        t.ID,
			Name:      t.Name,
			ExpiresAt: t.ExpiresAt,
			CreatedAt: t.CreatedAt,
			LastUsed:  t.LastUsed,
		}
	}
	return c.JSON(http.StatusOK, resp)
}

func (s *Server) handleCreateToken(c echo.Context) error {
	var req CreateTokenRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
	}

	plaintext, token, err := s.services.APIToken.Create(c.Request().Context(), req.Name)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(http.StatusCreated, TokenResponse{
		ID:        token.ID,
		Name:      token.Name,
		Token:     plaintext,
		ExpiresAt: token.ExpiresAt,
		CreatedAt: token.CreatedAt,
	})
}

func (s *Server) handleDeleteToken(c echo.Context) error {
	id, err := parseID(c, "tokenId")
	if err != nil {
		return respondError(c, err)
	}

	if err := s.services.APIToken.Delete(c.Request().Context(), id); err != nil {
		return respondError(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}
