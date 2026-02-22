# 004 — Edit Transactions & UI-Verbesserungen

## Kontext

Mehrere UX-Probleme beim Web-UI: fehlende Formulare zum Anlegen/Bearbeiten von Transaktionen und wiederkehrenden Transaktionen, keine gemeinsamen Tabs, Vorzeichen-basierte Ein-/Ausgaben-Unterscheidung war umständlich.

## Änderungen

### Shared Tabs (Commit: 01babe5)

- Neues Partial-Template `web/templates/household/tabs.html` mit Household-Header + Tab-Navigation
- Kategorie- und Recurring-Views nutzen jetzt dieselben Tabs (kein Back-Button mehr)
- `ActiveTab` Feld in `pageData` steuert aktiven Tab
- `Month` wird an alle Tab-Views übergeben

### Transaction-Formular (Commit: ec35ccd, 01babe5)

- `GET /households/:id/transactions/new` + `POST /households/:id/transactions`
- Einnahme/Ausgabe per Radio-Button (Default: Ausgabe) statt Vorzeichen
- "Today"-Button neben Datumspicker
- Inline neue Kategorie anlegen ("+ New category…" Option)
- `resolveCategory()` Shared Helper erstellt bei Bedarf neue Kategorie

### Recurring Transaction Formular (Commit: 01babe5)

- `GET /households/:id/recurring/new` + `POST /households/:id/recurring`
- Gleiche UX-Patterns wie Transaction-Formular (Radio, Today, New Category)
- Frequenz-Dropdown mit allen validen Werten aus `domain.AllFrequencies()`

### Edit-Support (Commit: 2a61658)

- **Schema**: `description` Feld zu RecurringExpense hinzugefügt (optional, max 500)
- **Transaction Update**: `Update` Methode in Repo-Interface, Repository und Service
- **Web-Routes**:
  - `GET/POST /households/:id/transactions/:txId/edit`
  - `GET/POST /households/:id/recurring/:recurringId/edit`
- Formulare werden im Edit-Modus mit bestehenden Werten vorausgefüllt
- Template-Funcs `absAmount` und `isNegative` für Betrag-Darstellung
- Klickbare Links in Tabellen (Datum/Beschreibung bei TX, Name bei Recurring)

### UI-Umbenennung

- "Recurring Expenses" → "Recurring Transactions" überall im Web-UI
- Household "Edit"-Button aus Header entfernt (Route existierte nie)

### Beschreibung-Feld

- Beide Formulare nutzen `<textarea>` für Beschreibung (max 500 Zeichen)
- Ent-Schema `recurringexpense` um `description` erweitert + `make generate`

## Geänderte Dateien

- `ent/schema/recurringexpense.go` + generierte Ent-Dateien
- `internal/domain/recurring_expense.go`, `internal/domain/repository.go`
- `internal/repository/convert.go`, `recurring_expense.go`, `transaction.go`
- `internal/service/transaction.go`, `recurring_expense.go`
- `internal/api/router.go`, `template.go`, `web_handler.go`, `recurring_expense_handler.go`
- `web/templates/household/tabs.html`, `detail.html`
- `web/templates/transaction/form.html`
- `web/templates/recurring/form.html`, `list.html`
- `web/templates/category/list.html`
