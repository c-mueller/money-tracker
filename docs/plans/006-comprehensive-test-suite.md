# Plan 006: Comprehensive Test Suite

## Status: Done

## Kontext
Vorher: 3 Testdateien (config loader, domain NormalizeToMonthly, integration mit 401-Fehler).
Nachher: Vollständige Tests über alle Schichten.

## Änderungen

### Phase 1: Domain Unit Tests (4 Dateien)
- `internal/domain/validate_test.go` — Currency, Email, Amount, DateRange, Month, HouseholdName, CategoryName, Description
- `internal/domain/money_test.go` — NewMoney (parse), MoneyFromInt, ZeroMoney
- `internal/domain/frequency_test.go` — Valid(), Validate(), AllFrequencies()
- `internal/domain/errors_test.go` — ValidationError, Unwrap, Sentinel-Fehler

### Phase 2: Service Unit Tests (7 Dateien)
- `internal/service/testutil_test.go` — In-memory SQLite Setup, Helper für User/Household/Category
- `internal/service/household_test.go` — Create (inkl. Defaults), List, Update, Delete (inkl. Cascade), Auth
- `internal/service/category_test.go` — CRUD, Household-Isolation, Default-Icon
- `internal/service/transaction_test.go` — CRUD, Monats-Filter, Household-Zugehörigkeit
- `internal/service/recurring_expense_test.go` — CRUD, Frequency-Validierung, Active-Flag, EndDate
- `internal/service/summary_test.go` — Leerer Monat, nur OneTime, nur Recurring, gemischt, Category-Breakdown
- `internal/service/api_token_test.go` — Create/Validate/List/Delete, Token-Format

### Phase 3: Integration Tests Fix + Erweiterung
- `tests/integration/testutil.go` — Bearer-Token erstellt im Setup, `token`-Feld in `testEnv`
- `tests/integration/api_test.go` — `doRequest()` setzt `Authorization: Bearer` Header
  - Fix: Bestehende Tests (TestFullFlow, TestValidation) nutzen jetzt Auth
  - Neu: TestHouseholdCRUD, TestCategoryCRUD, TestTransactionCRUD, TestRecurringExpenseCRUD
  - Neu: TestSummaryEndpoint, TestTokenManagement, TestUnauthorized

## Verifikation
```
make test                    # ✅ Alle Unit Tests grün
make test-integration        # ✅ Alle Integration Tests grün
```
