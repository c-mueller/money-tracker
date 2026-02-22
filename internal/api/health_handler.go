package api

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (s *Server) handleHealth(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}
