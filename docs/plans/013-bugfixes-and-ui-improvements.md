# 013 — Bugfixes & UI Improvements

## Changes

### BF-001: Amount Input with Comma + Error Toasts
- `NewMoney` normalizes comma to dot (`12,50` → `12.50`)
- Form handlers re-render with error message instead of HTTP error pages
- Added `ErrorMessage` to `pageData` and alert-dismissible block to layout
- Added `error_invalid_amount`, `error_invalid_date` i18n keys

### BF-002: Recurring Date Filtering
- Added `IsActiveInMonth(year, month)` to `RecurringExpense`
- Summary service skips recurring entries not active in the queried month
- Prevents recurring expenses from appearing before their start date

### FE-004: Global Household Header with Summary
- Moved recurring income/expenses/total cards to `tabs.html` (visible on all tabs)
- `detail.html` only shows month nav + one-time income/expenses
- `handleWebRecurringList` and `handleWebHouseholdSettings` now load summary

### FE-003: Separate Income & Expense Tables
- Split transactions into `IncomeTransactions`/`ExpenseTransactions`
- Split recurring into `IncomeRecurring`/`ExpenseRecurring`
- Each table has per-table total in `<tfoot>`, result card below

### FE-005+006: "Since" Column + Recurring Display Date
- Transaction tables show `CreatedAt` in "Since" column
- Recurring tables show `StartDate` in "Since" column
- `data-sort-value` attributes added for sorting

### FE-001: Pseudo-Transactions (Grouped Recurring Summary)
- Added `RecurringFrequencyGroup` and `RecurringEntry` domain types
- Summary service builds frequency groups during recurring loop
- Collapsible accordion sections per frequency group in detail view
- `EffectiveDate` set to 1st of queried month

### FE-002: Client-Side Table Sorting
- Tables opt in with `data-sortable` attribute
- Clickable `<th>` headers with `data-sort-type` (text, number, date)
- CSS sort indicators (arrows)
- Date/amount cells use `data-sort-value` for locale-independent sorting

## Files Modified
- `internal/domain/money.go`, `money_test.go`
- `internal/domain/recurring_expense.go`, `recurring_expense_test.go`
- `internal/domain/summary.go`
- `internal/service/summary.go`, `summary_test.go`
- `internal/api/web_handler.go`
- `internal/i18n/locales/en.json`, `de.json`
- `web/templates/layout.html`
- `web/templates/household/tabs.html`, `detail.html`
- `web/templates/recurring/list.html`
- `web/static/app.js`, `app.css`
