# Plan 015: Extended Transaction Description (Details Field)

## Motivation

Transactions currently only have a short `description` field (max 500 characters). For receipt contents, invoice details, or other supplementary information, this is not sufficient. We add an optional `details` text field (max 5000 characters) to Transaction and RecurringExpense — across all layers.

## Changes

### Schema
- `ent/schema/transaction.go`: `field.String("details").Optional().MaxLen(5000)` after `description`
- `ent/schema/recurringexpense.go`: `field.String("details").Optional().MaxLen(5000).Default("")` after `description`

### Domain
- `internal/domain/transaction.go`: `Details string` field
- `internal/domain/recurring_expense.go`: `Details string` field
- `internal/domain/validate.go`: `ValidateDetails()` (max 5000 characters)

### Repository
- `internal/repository/convert.go`: Details mapping in conversion functions
- `internal/repository/transaction.go`: SetDetails in Create/Update
- `internal/repository/recurring_expense.go`: SetDetails in Create/Update

### Service
- `internal/service/transaction.go`: `details string` parameter in Create/Update
- `internal/service/recurring_expense.go`: `details string` parameter in Create/Update

### API
- `internal/api/dto.go`: Details in all Transaction/RecurringExpense DTOs
- `internal/api/transaction_handler.go`: Pass details through
- `internal/api/recurring_expense_handler.go`: Pass details through
- `internal/api/web_handler.go`: FormValue("details") in web handlers

### GraphQL
- `internal/graphql/schema.graphqls`: details field in types and inputs
- `internal/graphql/schema.resolvers.go`: Pass details to services
- `internal/graphql/helpers.go`: Details in conversion functions

### MCP
- `internal/mcp/server.go`: Details in arg structs
- `internal/mcp/client.go`: Details in response structs

### Frontend
- i18n: details + details_placeholder strings
- Templates: Textarea for details in forms, info icon in lists
- OpenAPI: details property in schemas

## Design Decisions

- **Max 5000 characters**: Sufficient for receipt/invoice contents
- **Optional, empty string as default**: Consistent with existing `description` field
- **Simple text field**: Sufficient for the current use case
- **Native `title` tooltip with info icon**: Simplest solution for hover display
