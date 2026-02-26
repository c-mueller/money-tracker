# Plan 015: Erweiterte Transaktionsbeschreibung (Details-Feld)

## Motivation

Transaktionen haben aktuell nur ein kurzes `description`-Feld (max 500 Zeichen). Für Kassenzettel-Inhalte, Rechnungsdetails oder sonstige Zusatzinformationen reicht das nicht aus. Wir fügen ein optionales `details`-Textfeld (max 5000 Zeichen) zu Transaction und RecurringExpense hinzu — durch alle Layer hindurch.

## Änderungen

### Schema
- `ent/schema/transaction.go`: `field.String("details").Optional().MaxLen(5000)` nach `description`
- `ent/schema/recurringexpense.go`: `field.String("details").Optional().MaxLen(5000).Default("")` nach `description`

### Domain
- `internal/domain/transaction.go`: `Details string` Feld
- `internal/domain/recurring_expense.go`: `Details string` Feld
- `internal/domain/validate.go`: `ValidateDetails()` (max 5000 Zeichen)

### Repository
- `internal/repository/convert.go`: Details-Mapping in Konvertierungsfunktionen
- `internal/repository/transaction.go`: SetDetails in Create/Update
- `internal/repository/recurring_expense.go`: SetDetails in Create/Update

### Service
- `internal/service/transaction.go`: `details string` Parameter in Create/Update
- `internal/service/recurring_expense.go`: `details string` Parameter in Create/Update

### API
- `internal/api/dto.go`: Details in allen Transaction/RecurringExpense DTOs
- `internal/api/transaction_handler.go`: Details durchreichen
- `internal/api/recurring_expense_handler.go`: Details durchreichen
- `internal/api/web_handler.go`: FormValue("details") in Web-Handlern

### GraphQL
- `internal/graphql/schema.graphqls`: details Feld in Types und Inputs
- `internal/graphql/schema.resolvers.go`: details an Services übergeben
- `internal/graphql/helpers.go`: Details in Konvertierungsfunktionen

### MCP
- `internal/mcp/server.go`: Details in Arg-Structs
- `internal/mcp/client.go`: Details in Response-Structs

### Frontend
- i18n: details + details_placeholder Strings
- Templates: Textarea für Details in Formularen, Info-Icon in Listen
- OpenAPI: details Property in Schemas

## Design-Entscheidungen

- **Max 5000 Zeichen**: Genug für Kassenzettel/Rechnungsinhalte
- **Optional, leerer String als Default**: Konsistent mit bestehendem `description`-Feld
- **Einfaches Textfeld**: Reicht für den aktuellen Usecase
- **Native `title`-Tooltip mit Info-Icon**: Einfachste Lösung für Hover-Anzeige
