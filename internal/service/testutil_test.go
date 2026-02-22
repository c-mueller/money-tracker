package service_test

import (
	"context"
	"testing"

	"icekalt.dev/money-tracker/ent"
	"icekalt.dev/money-tracker/internal/config"
	"icekalt.dev/money-tracker/internal/domain"
	"icekalt.dev/money-tracker/internal/repository"
	"icekalt.dev/money-tracker/internal/service"

	_ "modernc.org/sqlite"
)

type testServices struct {
	client           *ent.Client
	User             *service.UserService
	Household        *service.HouseholdService
	Category         *service.CategoryService
	Transaction      *service.TransactionService
	RecurringExpense *service.RecurringExpenseService
	Summary          *service.SummaryService
	APIToken         *service.APITokenService
}

func setupTestServices(t *testing.T) *testServices {
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

	t.Cleanup(func() {
		client.Close()
	})

	return &testServices{
		client:           client,
		User:             userSvc,
		Household:        householdSvc,
		Category:         categorySvc,
		Transaction:      txSvc,
		RecurringExpense: recurringSvc,
		Summary:          summarySvc,
		APIToken:         tokenSvc,
	}
}

// createTestUser creates a user and returns a context with the user ID set.
func createTestUser(t *testing.T, svc *testServices) (context.Context, *domain.User) {
	t.Helper()
	user, err := svc.User.GetOrCreate(context.Background(), "test-sub", "test@example.com", "Test User")
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}
	ctx := service.WithUserID(context.Background(), user.ID)
	return ctx, user
}

// createTestHousehold creates a household and returns it.
func createTestHousehold(t *testing.T, svc *testServices, ctx context.Context) *domain.Household {
	t.Helper()
	hh, err := svc.Household.Create(ctx, "Test Household", "", "EUR", "")
	if err != nil {
		t.Fatalf("failed to create test household: %v", err)
	}
	return hh
}

// createTestCategory creates a category and returns it.
func createTestCategory(t *testing.T, svc *testServices, ctx context.Context, householdID int) *domain.Category {
	t.Helper()
	cat, err := svc.Category.Create(ctx, householdID, "Test Category", "")
	if err != nil {
		t.Fatalf("failed to create test category: %v", err)
	}
	return cat
}
