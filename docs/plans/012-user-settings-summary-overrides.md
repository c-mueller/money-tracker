# Implementation Plan: User Settings, Household Menu, Currency Display, Summary Restructure, Schedule Overrides

## Context

The money-tracker needs several UI and backend enhancements:
1. User profile settings page (change display name, view email)
2. Household settings split into sidebar navigation (Household, Categories, Danger Zone)
3. Currency symbol/code shown after all amounts
4. Monthly summary restructured: separate recurring income/expenses, one-time income/expenses, monthly total
5. Recurring transactions support price+frequency changes over time via schedule overrides

---

## Phase 1: User Settings Page

### Files to modify
- `internal/service/user.go` — add `UpdateName` method
- `internal/api/web_handler.go` — add `handleWebUserSettings`, `handleWebUserSettingsUpdate`
- `internal/api/router.go` — register GET/POST `/settings`
- `internal/api/template.go` — register `"user_settings"` template name
- `web/templates/layout.html` — add "Settings" dropdown item
- NEW: `web/templates/user/settings.html` — user settings form
- `internal/i18n/locales/en.json`, `de.json` — add keys: `user_settings`, `display_name`, `email_address`
- NEW: `internal/service/user_test.go` — test `UpdateName`

### Details
- `UpdateName(ctx, name)` validates name (reuse `domain.ValidateHouseholdName`), gets userID from context, fetches user, sets name, calls `repo.Update`
- Settings template: form with editable name, read-only email (disabled input), save button
- Navbar dropdown order: API-Tokens, Settings, divider, Logout

---

## Phase 2: Household Settings Side Menu

### Files to modify
- `internal/api/web_handler.go` — add `ActiveSection` to `pageData`, read `?section=` query param in settings handler
- `web/templates/household/settings.html` — rewrite with Bootstrap sidebar layout
- `internal/api/web_handler.go` — update category redirects to include `?section=categories`
- `internal/i18n/locales/en.json`, `de.json` — add key: `danger_zone_menu` (if needed)

### Details
- `pageData` gets new field `ActiveSection string`
- `handleWebHouseholdSettings` reads `section` query param, defaults to `"household"`
- Template layout: `col-md-3` sidebar with 3 `list-group` items:
  - "Haushalt" (`?section=household`) — name, description, currency, icon form
  - "Kategorien" (`?section=categories`) — category add form + table
  - "Gefahrenzone" (`?section=danger`) — delete household button
- Active item gets `active` class
- Category create/update/delete redirects append `?section=categories`

---

## Phase 3: Currency Symbol After Amounts

### Files to modify
- `internal/api/template.go` — add `formatMoneyWithCurrency` template function
- `web/templates/dashboard.html` — use `formatMoneyWithCurrency`
- `web/templates/household/detail.html` — use `formatMoneyWithCurrency`
- `web/templates/recurring/list.html` — use `formatMoneyWithCurrency`

### Details
- Logic for currency display: look up symbol via `currencyByCode[code]`. If symbol is a single character (len <= 1 rune, e.g. `$`, `EUR-symbol`, `GBP-symbol`), display it. Otherwise display the 3-letter code (e.g. "CHF", "SEK").
- New template func `formatMoneyWithCurrency(amount, currencyCode)` → e.g. "500,00 EUR-symbol" or "500,00 CHF"
- Must be overridden in `Render()` for locale-aware formatting (same pattern as `formatMoney`)
- Dashboard: `.Currency` is available inside `{{range .Households}}`
- Detail page: `.Household.Currency` available
- Recurring list: `$.Household.Currency` available

---

## Phase 4: Transaction Overview Restructure

### Files to modify
- `internal/domain/summary.go` — add fields: `RecurringIncome`, `RecurringExpenses`, `OneTimeIncome`, `OneTimeExpenses`, `MonthlyTotal`
- `internal/service/summary.go` — compute new fields in `GetMonthlySummary`
- `internal/api/dto.go` — add fields to `SummaryResponse`
- `internal/api/summary_handler.go` — map new fields in `toSummaryResponse`
- `internal/graphql/schema.graphqls` — add fields to `MonthlySummary` type
- `internal/graphql/helpers.go` — map new fields in `toGQLMonthlySummary`
- `web/templates/household/detail.html` — show 6 cards: Month nav, Recurring Income, Recurring Expenses, One-time Income, One-time Expenses, Monthly Total
- `web/templates/dashboard.html` — show 5 values per household card
- `web/static/openapi.yaml` — add fields to Summary schema
- `internal/i18n/locales/en.json`, `de.json` — add keys: `recurring_income`, `recurring_expenses`, `one_time_income`, `one_time_expenses`, `monthly_total`
- `internal/service/summary_test.go` — test new field calculations

### Details
- In the recurring expense loop: track `recurringIncome` (positive amounts) and `recurringExpenses` (negative amounts) separately
- `OneTimeIncome` = existing `TotalIncome`, `OneTimeExpenses` = existing `TotalExpenses` (rename for clarity but keep old fields)
- `MonthlyTotal` = `RecurringTotal + OneTimeTotal`
- Keep backward-compatible fields: `RecurringTotal`, `OneTimeTotal`, `TotalIncome`, `TotalExpenses`

---

## Phase 5: Recurring Schedule Overrides

### New files
- `ent/schema/recurringscheduleoverride.go` — new Ent schema
- `internal/repository/recurring_schedule_override.go` — repo implementation
- `internal/api/schedule_override_handler.go` — REST API handlers
- `internal/domain/recurring_expense_test.go` — test `EffectiveSchedule`

### Files to modify
- `ent/schema/recurringexpense.go` — add edge `To("schedule_overrides", ...)`
- `internal/domain/recurring_expense.go` — add `RecurringScheduleOverride` type + `EffectiveSchedule()` helper
- `internal/domain/repository.go` — add `RecurringScheduleOverrideRepo` interface
- `internal/repository/convert.go` — add `overrideToDomain` conversion
- `internal/service/recurring_expense.go` — add `overrideRepo` field, override CRUD methods, update constructor
- `internal/service/summary.go` — add `overrideRepo`, use `EffectiveSchedule` in monthly calculation
- `internal/api/dto.go` — add override DTOs
- `internal/api/router.go` — register override routes
- `internal/api/recurring_expense_handler.go` — include overrides in response
- `internal/api/web_handler.go` — add `ScheduleOverrides` to `pageData`, load in edit handler, add web handlers for override create/delete
- `web/templates/recurring/form.html` — add override section (only on edit)
- `internal/graphql/schema.graphqls` — add `ScheduleOverride` type, inputs, query, mutations
- `internal/graphql/schema.resolvers.go` — implement resolver stubs
- `internal/graphql/helpers.go` — add `toGQLScheduleOverride`
- `web/static/openapi.yaml` — add override schemas + endpoints
- `internal/i18n/locales/en.json`, `de.json` — add keys
- `cmd/money-tracker/cmd/serve.go` — wire new repo into services
- `internal/service/testutil_test.go` — wire new repo in test setup
- `tests/integration/testutil.go` — wire new repo in integration test setup

### Data Model
```
RecurringScheduleOverride:
  id            int (auto)
  effective_date time.Time (required) — from when this override applies
  amount        string (required) — decimal amount
  frequency     string (required) — frequency enum value
  created_at    time.Time
  updated_at    time.Time
  Edge: belongs to RecurringExpense (required)
```

### EffectiveSchedule Logic
```go
func EffectiveSchedule(baseAmount Money, baseFreq Frequency, overrides []*RecurringScheduleOverride, year int, month time.Month) (Money, Frequency) {
    // overrides sorted by effective_date ascending
    // Find latest override where effective_date <= last day of queried month
    // If found → return override's amount/frequency
    // Else → return base amount/frequency
}
```

### REST API Endpoints
```
GET    /api/v1/households/:id/recurring-expenses/:recurringId/overrides
POST   /api/v1/households/:id/recurring-expenses/:recurringId/overrides
PUT    /api/v1/households/:id/recurring-expenses/:recurringId/overrides/:overrideId
DELETE /api/v1/households/:id/recurring-expenses/:recurringId/overrides/:overrideId
```

### Web Routes
```
POST   /households/:id/recurring/:recurringId/overrides           (create)
POST   /households/:id/recurring/:recurringId/overrides/:overrideId/delete  (delete)
```

### UI: Override Section in Edit Form
- Only shown when editing (inside `{{if .RecurringExpense}}`)
- Table: effective date, amount, frequency, delete button
- Inline form: date picker, amount input, frequency dropdown, add button

### Tests
- **Domain test** (`internal/domain/recurring_expense_test.go`): `EffectiveSchedule` with no overrides, single override, multiple overrides, boundary cases
- **Service test** (`internal/service/recurring_expense_test.go`): Override CRUD, validation
- **Summary test** (`internal/service/summary_test.go`): Summary with overrides applying different amounts per month
- **Integration test** (`tests/integration/api_test.go`): Override REST API CRUD + summary verification

---

## Implementation Order

1. **Phase 1**: User Settings (independent)
2. **Phase 2**: Household Side Menu (independent)
3. **Phase 3**: Currency Symbol (independent)
4. **Phase 4**: Summary Restructure (depends on Phase 3 for template helper)
5. **Phase 5**: Schedule Overrides (depends on Phase 4 for summary changes)

Each phase ends with a commit. After Phase 5, run `make generate`, `make test`, `make test-integration`, `make lint`.

---

## Verification

1. `make generate` — Ent codegen succeeds after new schema
2. `make build-dev` — compiles without errors
3. `make test` — all unit tests pass (including new ones)
4. `make test-integration` — all integration tests pass
5. `make lint` — no lint errors
6. Manual: `make run-dev` → verify:
   - Click username → "Settings" appears, page lets you change name
   - Household settings has sidebar with 3 sections
   - All amounts show currency symbol/code suffix
   - Transaction overview shows 5 separate cards + monthly total
   - Edit recurring → override section visible, can add/delete overrides
   - Summary reflects overrides for different months

---

## Plan file: `docs/plans/007-user-settings-summary-overrides.md`
