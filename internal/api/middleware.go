package api

import (
	"github.com/labstack/echo/v4/middleware"

	mw "icekalt.dev/money-tracker/internal/middleware"
)

func (s *Server) setupMiddleware() {
	s.echo.Use(mw.Recovery(s.logger))
	s.echo.Use(mw.RequestID())
	s.echo.Use(middleware.SecureWithConfig(middleware.SecureConfig{
		XSSProtection:         "1; mode=block",
		ContentTypeNosniff:    "nosniff",
		XFrameOptions:         "DENY",
		ReferrerPolicy:        "strict-origin-when-cross-origin",
		ContentSecurityPolicy: "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self'",
	}))
	s.echo.Use(mw.Logger(s.logger))
	s.echo.Use(mw.InjectUserID())
}
