//go:build integration

package integration

import (
	"context"
	"net/http/httptest"
	"testing"

	"icekalt.dev/money-tracker/ent"
	"icekalt.dev/money-tracker/internal/api"
	"icekalt.dev/money-tracker/internal/auth"
	"icekalt.dev/money-tracker/internal/config"
	"icekalt.dev/money-tracker/internal/logging"
	"icekalt.dev/money-tracker/internal/repository"
	"icekalt.dev/money-tracker/internal/service"

	_ "modernc.org/sqlite"
)

type testEnv struct {
	server   *httptest.Server
	client   *ent.Client
	services *api.Services
}

func setupTestEnv(t *testing.T) *testEnv {
	t.Helper()

	dbCfg := config.DatabaseConfig{
		Driver: "sqlite",
		DSN:    "file::memory:?cache=shared&_pragma=foreign_keys(1)",
	}

	client, err := repository.NewClient(dbCfg)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	if err := client.Schema.Create(context.Background()); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	userRepo := repository.NewUserRepository(client)
	householdRepo := repository.NewHouseholdRepository(client)
	categoryRepo := repository.NewCategoryRepository(client)
	txRepo := repository.NewTransactionRepository(client)
	recurringRepo := repository.NewRecurringExpenseRepository(client)
	tokenRepo := repository.NewAPITokenRepository(client)

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

	logger, _ := logging.New("error")
	srv := api.NewServer(logger, "127.0.0.1", 0, svcs)

	devUser, err := userSvc.GetOrCreate(context.Background(), "test-user", "test@example.com", "Test User")
	if err != nil {
		t.Fatalf("failed to create dev user: %v", err)
	}

	store := auth.NewSessionStore("test-secret-key-for-testing-only", 3600)
	srv.SetupAuth(nil, store, true, devUser.ID)

	ts := httptest.NewServer(srv.Echo())

	t.Cleanup(func() {
		ts.Close()
		client.Close()
	})

	return &testEnv{
		server:   ts,
		client:   client,
		services: svcs,
	}
}
