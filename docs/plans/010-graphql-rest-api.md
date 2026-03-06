# Plan 010: GraphQL API + REST API Completion

## Summary
- REST API extended with missing fields and update endpoints
- GraphQL API added with gqlgen (token-only auth)
- Comprehensive integration tests for both

## Changes

### REST API Fixes
- **DTOs**: `CreateHouseholdRequest` +description/icon, `HouseholdResponse` +description/icon, new `UpdateHouseholdRequest`
- **DTOs**: `CreateCategoryRequest` +icon, `CategoryResponse` +icon, new `UpdateCategoryRequest`
- **DTOs**: New `UpdateTransactionRequest`
- **DTOs**: `CreateRecurringExpenseRequest` +description, `UpdateRecurringExpenseRequest` +description, `RecurringExpenseResponse` +description
- **Handlers**: All handlers pass new fields to service layer
- **Route**: `PUT /api/v1/households/:id/transactions/:transactionId` added

### GraphQL API
- Schema-first with gqlgen v0.17.87
- Queries: households, household, categories, transactions, recurringExpenses, monthlySummary
- Mutations: create/update for Household, Category, Transaction, RecurringExpense (no delete)
- Money as string, dates as string (YYYY-MM-DD), optional fields nullable
- Token-only auth middleware (no session cookie)
- Playground at `/playground` only in dev mode

### Tests
- REST: TestTransactionUpdate, TestHouseholdFullFields, TestCategoryWithIcon, TestRecurringExpenseDescription
- GraphQL: TestGraphQLHouseholds, TestGraphQLCategories, TestGraphQLTransactions, TestGraphQLRecurringExpenses, TestGraphQLSummary, TestGraphQLTokenOnlyAuth, TestGraphQLNoDeleteMutations, TestGraphQLValidationErrors, TestGraphQLFullFlow

## New Files
- `internal/graphql/schema.graphqls` — GraphQL schema
- `internal/graphql/gqlgen.yml` — gqlgen configuration
- `internal/graphql/generate.go` — go:generate directive
- `internal/graphql/generated.go` — Generated code
- `internal/graphql/model/models_gen.go` — Generated models
- `internal/graphql/resolver.go` — Root resolver
- `internal/graphql/helpers.go` — Domain → GraphQL conversion
- `internal/graphql/schema.resolvers.go` — Resolver implementation
- `tests/integration/graphql_test.go` — GraphQL integration tests
- `docs/plans/010-graphql-rest-api.md` — This document
