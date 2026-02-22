package cmd

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"icekalt.dev/money-tracker/internal/api"
	authpkg "icekalt.dev/money-tracker/internal/auth"
	"icekalt.dev/money-tracker/internal/devmode"
	"icekalt.dev/money-tracker/internal/repository"
	"icekalt.dev/money-tracker/internal/service"

	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the HTTP server",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := repository.NewClient(cfg.Database)
		if err != nil {
			return fmt.Errorf("connecting to database: %w", err)
		}
		defer client.Close()

		// Auto-migrate
		if err := client.Schema.Create(context.Background()); err != nil {
			return fmt.Errorf("running migrations: %w", err)
		}

		// Repositories
		userRepo := repository.NewUserRepository(client)
		householdRepo := repository.NewHouseholdRepository(client)
		categoryRepo := repository.NewCategoryRepository(client)
		txRepo := repository.NewTransactionRepository(client)
		recurringRepo := repository.NewRecurringExpenseRepository(client)
		tokenRepo := repository.NewAPITokenRepository(client)

		// Services
		userSvc := service.NewUserService(userRepo)
		householdSvc := service.NewHouseholdService(householdRepo, categoryRepo, txRepo, recurringRepo)
		categorySvc := service.NewCategoryService(categoryRepo, householdSvc)
		txSvc := service.NewTransactionService(txRepo, householdSvc)
		recurringSvc := service.NewRecurringExpenseService(recurringRepo, householdSvc)
		summarySvc := service.NewSummaryService(txRepo, recurringRepo, categoryRepo, householdSvc)
		tokenSvc := service.NewAPITokenService(tokenRepo)

		svcs := &api.Services{
			User:             userSvc,
			Household:        householdSvc,
			Category:         categorySvc,
			Transaction:      txSvc,
			RecurringExpense: recurringSvc,
			Summary:          summarySvc,
			APIToken:         tokenSvc,
		}

		srv := api.NewServer(logger, cfg.Server.Host, cfg.Server.Port, svcs)

		// Session store
		sessionSecret := cfg.Auth.Session.Secret
		if sessionSecret == "" {
			b := make([]byte, 32)
			if _, err := rand.Read(b); err != nil {
				return fmt.Errorf("generating session secret: %w", err)
			}
			sessionSecret = hex.EncodeToString(b)
		}
		store := authpkg.NewSessionStore(sessionSecret, cfg.Auth.Session.MaxAge)

		// Setup auth based on build tag
		var devUserID int
		if devmode.Enabled {
			logger.Warn("DEV BUILD — running with auto-auth")
			devUserID, err = devmode.SetupUser(func() (int, error) {
				devUser, err := userSvc.GetOrCreate(context.Background(), "dev-user", "dev@localhost", "Dev User")
				if err != nil {
					return 0, err
				}
				return devUser.ID, nil
			})
			if err != nil {
				return fmt.Errorf("creating dev user: %w", err)
			}
			srv.SetupAuth(nil, store, devUserID)
		} else {
			oidcCfg, err := authpkg.NewOIDC(
				context.Background(),
				cfg.Auth.OIDC.Issuer,
				cfg.Auth.OIDC.ClientID,
				cfg.Auth.OIDC.ClientSecret,
				cfg.Auth.OIDC.RedirectURL,
			)
			if err != nil {
				return fmt.Errorf("setting up OIDC: %w", err)
			}
			srv.SetupAuth(oidcCfg, store, 0)
		}

		return srv.Start(context.Background())
	},
}

func init() {
	serveCmd.Flags().IntVar(&cfg.Server.Port, "port", 0, "server port (overrides config)")
	rootCmd.AddCommand(serveCmd)
}
