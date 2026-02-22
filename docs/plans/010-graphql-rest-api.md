# Plan 010: GraphQL API + REST API Vervollständigung

## Zusammenfassung
- REST-API um fehlende Felder und Update-Endpoints erweitert
- GraphQL-API mit gqlgen hinzugefügt (Token-only Auth)
- Umfassende Integration-Tests für beides

## Änderungen

### REST API Fixes
- **DTOs**: `CreateHouseholdRequest` +description/icon, `HouseholdResponse` +description/icon, neues `UpdateHouseholdRequest`
- **DTOs**: `CreateCategoryRequest` +icon, `CategoryResponse` +icon, neues `UpdateCategoryRequest`
- **DTOs**: Neues `UpdateTransactionRequest`
- **DTOs**: `CreateRecurringExpenseRequest` +description, `UpdateRecurringExpenseRequest` +description, `RecurringExpenseResponse` +description
- **Handler**: Alle Handler reichen neue Felder an Service-Layer durch
- **Route**: `PUT /api/v1/households/:id/transactions/:transactionId` hinzugefügt

### GraphQL API
- Schema-first mit gqlgen v0.17.87
- Queries: households, household, categories, transactions, recurringExpenses, monthlySummary
- Mutations: create/update für Household, Category, Transaction, RecurringExpense (kein Delete)
- Money als String, Dates als String (YYYY-MM-DD), optionale Felder nullable
- Token-only Auth Middleware (kein Session-Cookie)
- Playground unter `/playground` nur im Dev-Mode

### Tests
- REST: TestTransactionUpdate, TestHouseholdFullFields, TestCategoryWithIcon, TestRecurringExpenseDescription
- GraphQL: TestGraphQLHouseholds, TestGraphQLCategories, TestGraphQLTransactions, TestGraphQLRecurringExpenses, TestGraphQLSummary, TestGraphQLTokenOnlyAuth, TestGraphQLNoDeleteMutations, TestGraphQLValidationErrors, TestGraphQLFullFlow

## Neue Dateien
- `internal/graphql/schema.graphqls` — GraphQL-Schema
- `internal/graphql/gqlgen.yml` — gqlgen-Konfiguration
- `internal/graphql/generate.go` — go:generate-Direktive
- `internal/graphql/generated.go` — generierter Code
- `internal/graphql/model/models_gen.go` — generierte Models
- `internal/graphql/resolver.go` — Root-Resolver
- `internal/graphql/helpers.go` — Domain→GraphQL-Konvertierung
- `internal/graphql/schema.resolvers.go` — Resolver-Implementierung
- `tests/integration/graphql_test.go` — GraphQL-Integration-Tests
- `docs/plans/010-graphql-rest-api.md` — Dieses Dokument
