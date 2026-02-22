package api

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"icekalt.dev/money-tracker/web"
)

func (s *Server) handleOpenAPISpec(c echo.Context) error {
	data, err := web.Content.ReadFile("static/openapi.yaml")
	if err != nil {
		return c.String(http.StatusInternalServerError, "spec not found")
	}
	return c.Blob(http.StatusOK, "application/yaml", data)
}

func (s *Server) handleSwaggerUI(c echo.Context) error {
	data, err := web.Content.ReadFile("static/swagger/index.html")
	if err != nil {
		return c.String(http.StatusInternalServerError, "swagger ui not found")
	}
	return c.HTMLBlob(http.StatusOK, data)
}
