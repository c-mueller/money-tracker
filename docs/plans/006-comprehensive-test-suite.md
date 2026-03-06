# Plan 006: Comprehensive Test Suite

## Status: Done

## Context
Before: 3 test files (config loader, domain NormalizeToMonthly, integration with 401 error).
After: Comprehensive tests across all layers.

## Changes

### Phase 1: Domain Unit Tests (4 files)
- `internal/domain/validate_test.go` — Currency, Email, Amount, DateRange, Month, HouseholdName, CategoryName, Description
- `internal/domain/money_test.go` — NewMoney (parse), MoneyFromInt, ZeroMoney
- `internal/domain/frequency_test.go` — Valid(), Validate(), AllFrequencies()
- `internal/domain/errors_test.go` — ValidationError, Unwrap, sentinel errors

### Phase 2: Service Unit Tests (7 files)
- `internal/service/testutil_test.go` — In-memory SQLite setup, helpers for User/Household/Category
- `internal/service/household_test.go` — Create (incl. defaults), List, Update, Delete (incl. cascade), Auth
- `internal/service/category_test.go` — CRUD, household isolation, default icon
- `internal/service/transaction_test.go` — CRUD, month filter, household membership
- `internal/service/recurring_expense_test.go` — CRUD, frequency validation, active flag, EndDate
- `internal/service/summary_test.go` — Empty month, one-time only, recurring only, mixed, category breakdown
- `internal/service/api_token_test.go` — Create/Validate/List/Delete, token format

### Phase 3: Integration Tests Fix + Extension
- `tests/integration/testutil.go` — Bearer token created in setup, `token` field in `testEnv`
- `tests/integration/api_test.go` — `doRequest()` sets `Authorization: Bearer` header
  - Fix: Existing tests (TestFullFlow, TestValidation) now use auth
  - New: TestHouseholdCRUD, TestCategoryCRUD, TestTransactionCRUD, TestRecurringExpenseCRUD
  - New: TestSummaryEndpoint, TestTokenManagement, TestUnauthorized

## Verification
```
make test                    # All unit tests green
make test-integration        # All integration tests green
```
