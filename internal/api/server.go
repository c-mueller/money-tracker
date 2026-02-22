package api

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"icekalt.dev/money-tracker/internal/service"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type Server struct {
	echo         *echo.Echo
	logger       *zap.Logger
	port         int
	host         string
	services     *Services
	sessionStore sessions.Store
	authHandler  *AuthHandler
	devUserID    int
	renderer     *TemplateRenderer
}

type Services struct {
	User             *service.UserService
	Household        *service.HouseholdService
	Category         *service.CategoryService
	Transaction      *service.TransactionService
	RecurringExpense *service.RecurringExpenseService
	Summary          *service.SummaryService
	APIToken         *service.APITokenService
}

func NewServer(logger *zap.Logger, host string, port int, svc *Services) *Server {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	s := &Server{
		echo:     e,
		logger:   logger,
		port:     port,
		host:     host,
		services: svc,
	}

	s.setupMiddleware()

	return s
}

// Echo returns the underlying echo instance (for testing).
func (s *Server) Echo() *echo.Echo {
	s.setupRoutes()
	return s.echo
}

func (s *Server) Start(ctx context.Context) error {
	// Setup routes after auth is configured
	s.setupRoutes()

	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	s.logger.Info("starting server", zap.String("addr", addr))

	errCh := make(chan error, 1)
	go func() {
		errCh <- s.echo.Start(addr)
	}()

	select {
	case <-ctx.Done():
		s.logger.Info("shutting down server")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return s.echo.Shutdown(shutdownCtx)
	case err := <-errCh:
		return err
	}
}
