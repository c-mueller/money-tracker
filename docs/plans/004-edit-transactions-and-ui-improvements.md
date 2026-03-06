# 004 — Edit Transactions & UI Improvements

## Context

Several UX issues in the web UI: missing forms for creating/editing transactions and recurring transactions, no shared tabs, sign-based income/expense distinction was cumbersome.

## Changes

### Shared Tabs (Commit: 01babe5)

- New partial template `web/templates/household/tabs.html` with household header + tab navigation
- Category and recurring views now use the same tabs (no more back button)
- `ActiveTab` field in `pageData` controls the active tab
- `Month` is passed to all tab views

### Transaction Form (Commit: ec35ccd, 01babe5)

- `GET /households/:id/transactions/new` + `POST /households/:id/transactions`
- Income/expense via radio button (default: expense) instead of sign
- "Today" button next to date picker
- Inline new category creation ("+ New category…" option)
- `resolveCategory()` shared helper creates new category when needed

### Recurring Transaction Form (Commit: 01babe5)

- `GET /households/:id/recurring/new` + `POST /households/:id/recurring`
- Same UX patterns as transaction form (radio, today, new category)
- Frequency dropdown with all valid values from `domain.AllFrequencies()`

### Edit Support (Commit: 2a61658)

- **Schema**: `description` field added to RecurringExpense (optional, max 500)
- **Transaction Update**: `Update` method in repo interface, repository, and service
- **Web Routes**:
  - `GET/POST /households/:id/transactions/:txId/edit`
  - `GET/POST /households/:id/recurring/:recurringId/edit`
- Forms are pre-filled with existing values in edit mode
- Template funcs `absAmount` and `isNegative` for amount display
- Clickable links in tables (date/description for TX, name for recurring)

### UI Renaming

- "Recurring Expenses" → "Recurring Transactions" throughout the web UI
- Household "Edit" button removed from header (route never existed)

### Description Field

- Both forms use `<textarea>` for description (max 500 characters)
- Ent schema `recurringexpense` extended with `description` + `make generate`

## Modified Files

- `ent/schema/recurringexpense.go` + generated Ent files
- `internal/domain/recurring_expense.go`, `internal/domain/repository.go`
- `internal/repository/convert.go`, `recurring_expense.go`, `transaction.go`
- `internal/service/transaction.go`, `recurring_expense.go`
- `internal/api/router.go`, `template.go`, `web_handler.go`, `recurring_expense_handler.go`
- `web/templates/household/tabs.html`, `detail.html`
- `web/templates/transaction/form.html`
- `web/templates/recurring/form.html`, `list.html`
- `web/templates/category/list.html`
