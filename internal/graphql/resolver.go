package graphql

import "icekalt.dev/money-tracker/internal/service"

// Resolver holds the service dependencies for GraphQL resolvers.
type Resolver struct {
	HouseholdSvc        *service.HouseholdService
	CategorySvc         *service.CategoryService
	TransactionSvc      *service.TransactionService
	RecurringExpenseSvc *service.RecurringExpenseService
	SummarySvc          *service.SummaryService
}
