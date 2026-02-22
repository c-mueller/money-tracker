package api

import (
	"io/fs"
	"net/http"

	"github.com/labstack/echo/v4"
	"icekalt.dev/money-tracker/web"
)

func (s *Server) setupStatic() {
	staticFS, err := fs.Sub(web.Content, "static")
	if err != nil {
		s.logger.Fatal("failed to access static files")
	}
	s.echo.GET("/static/*", echo.WrapHandler(http.StripPrefix("/static/", http.FileServer(http.FS(staticFS)))))
}
