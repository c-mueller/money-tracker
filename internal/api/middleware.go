package api

import (
	mw "icekalt.dev/money-tracker/internal/middleware"
)

func (s *Server) setupMiddleware() {
	s.echo.Use(mw.Recovery(s.logger))
	s.echo.Use(mw.RequestID())
	s.echo.Use(mw.Logger(s.logger))
	s.echo.Use(mw.InjectUserID())
}
