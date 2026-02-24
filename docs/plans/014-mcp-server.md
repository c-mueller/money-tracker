# 014 — MCP Server for Money Tracker

## Motivation

Ein MCP (Model Context Protocol) Server ermöglicht es LLM-Clients (Claude Desktop, Claude Code, etc.), direkt mit Money Tracker zu interagieren — Transaktionen anlegen, Budgets abfragen, Zusammenfassungen generieren, ohne die Web-UI nutzen zu müssen.

## Architektur

### Phase 1: Lokaler MCP Server (stdio-basiert)

Der MCP Server wird als **Cobra-Subcommand** (`money-tracker mcp`) in die bestehende Binary integriert. Er kommuniziert über **stdio** (JSON-RPC) mit dem LLM-Client und spricht intern die **REST API** des laufenden Money-Tracker-Servers an.

```
┌─────────────────┐     stdio (JSON-RPC)     ┌──────────────────────────┐     HTTP/REST     ┌──────────────────┐
│  LLM Client     │ ◄──────────────────────► │  money-tracker mcp       │ ◄──────────────► │  money-tracker   │
│  (Claude, etc.) │                           │  (Subcommand, selbe Bin) │                  │  serve           │
└─────────────────┘                           └──────────────────────────┘                  └──────────────────┘
```

**Aufruf:**
```bash
money-tracker mcp                                        # Default: localhost:8080
money-tracker mcp --url http://myserver:9090             # Custom URL
MONEY_TRACKER_API_TOKEN=mt_... money-tracker mcp         # Token via ENV
money-tracker mcp --token mt_...                         # Token via Flag
```

**Konfiguration:**
- `--url` / `MONEY_TRACKER_MCP_URL` — Base-URL des API-Servers (default: `http://localhost:8080`)
- `--token` / `MONEY_TRACKER_API_TOKEN` — Bearer-Token (`mt_...`) für Authentifizierung

**Vorteile Phase 1:**
- Kein OAuth-Infrastruktur nötig — nutzt existierende API-Token-Auth
- Schnell umsetzbar, sofort lokal nutzbar
- Volle Funktionalität über bestehende REST API
- Eine Binary für alles — Config, Logging, Buildinfo werden wiederverwendet

**Nachteile / Bewusste Trade-offs:**
- Reimplementierung: MCP-Tool-Layer dupliziert API-Client-Logik
- Token muss manuell erstellt und konfiguriert werden
- Nur lokal nutzbar (kein Remote-Zugriff für gehostete LLM-Clients)

### Phase 2: Remote MCP Server mit OAuth 2.1 (Zukunft)

Langfristig soll der MCP Server als **Remote HTTP SSE/Streamable HTTP** Endpoint direkt im Money-Tracker-Server laufen und OAuth 2.1 für die Autorisierung nutzen.

```
┌─────────────────┐    HTTP SSE / Streamable HTTP    ┌──────────────────────────────┐
│  LLM Client     │ ◄──────────────────────────────► │  Money Tracker Server        │
│  (Claude, etc.) │         + OAuth 2.1               │  (integrierter MCP Endpoint) │
└─────────────────┘                                   └──────────────────────────────┘
```

**Voraussetzungen für Phase 2:**
- OAuth 2.1 Authorization Server (entweder self-hosted oder externer Provider)
- Dynamic Client Registration (RFC 7591) oder vorkonfigurierte Clients
- PKCE-Flow für LLM-Clients
- Token-Scoping (welche Households darf ein Client sehen?)
- MCP-Endpoint direkt in den Echo-Router integriert

**Migration Phase 1 → Phase 2:**
- Die MCP-Tool-Definitionen und Beschreibungen bleiben identisch
- Nur der Transport wechselt (stdio → HTTP SSE) und die Auth (API-Token → OAuth)
- Der API-Client-Layer aus Phase 1 wird durch direkte Service-Layer-Aufrufe ersetzt

---

## MCP Server Spezifikation (Phase 1)

### Server-Info

```json
{
  "name": "money-tracker",
  "version": "0.1.0"
}
```

### Tools

#### Household Management

##### `list_households`
Alle Haushalte des authentifizierten Benutzers auflisten.

- **Parameter:** keine
- **Returns:** Array von Households (id, name, description, currency, icon)

##### `create_household`
Neuen Haushalt anlegen.

- **Parameter:**
  - `name` (string, required) — Name des Haushalts
  - `description` (string, optional) — Beschreibung
  - `currency` (string, required) — ISO 4217 Währungscode (z.B. "EUR")
  - `icon` (string, optional) — Material Icon Name
- **Returns:** Der erstellte Haushalt

##### `update_household`
Haushalt aktualisieren.

- **Parameter:**
  - `id` (integer, required) — Household ID
  - `name` (string, optional)
  - `description` (string, optional)
  - `currency` (string, optional)
  - `icon` (string, optional)
- **Returns:** Der aktualisierte Haushalt

##### `delete_household`
Haushalt löschen (kaskadiert: alle Kategorien, Transaktionen, Recurring Expenses).

- **Parameter:**
  - `id` (integer, required) — Household ID
- **Returns:** Bestätigung

---

#### Category Management

##### `list_categories`
Kategorien eines Haushalts auflisten.

- **Parameter:**
  - `household_id` (integer, required)
- **Returns:** Array von Categories (id, name, icon)

##### `create_category`
Neue Kategorie anlegen.

- **Parameter:**
  - `household_id` (integer, required)
  - `name` (string, required)
  - `icon` (string, optional)
- **Returns:** Die erstellte Kategorie

##### `update_category`
Kategorie aktualisieren.

- **Parameter:**
  - `household_id` (integer, required)
  - `category_id` (integer, required)
  - `name` (string, optional)
  - `icon` (string, optional)
- **Returns:** Die aktualisierte Kategorie

##### `delete_category`
Kategorie löschen.

- **Parameter:**
  - `household_id` (integer, required)
  - `category_id` (integer, required)
- **Returns:** Bestätigung

---

#### Transaction Management

##### `list_transactions`
Transaktionen eines Haushalts für einen Monat auflisten.

- **Parameter:**
  - `household_id` (integer, required)
  - `month` (string, optional) — Format "YYYY-MM", default: aktueller Monat
- **Returns:** Array von Transactions (id, amount, description, date, category_id, category_name)

##### `create_transaction`
Neue Transaktion anlegen.

- **Parameter:**
  - `household_id` (integer, required)
  - `category_id` (integer, required)
  - `amount` (string, required) — Dezimalzahl als String, negativ = Ausgabe, positiv = Einnahme
  - `description` (string, required)
  - `date` (string, required) — Format "YYYY-MM-DD"
- **Returns:** Die erstellte Transaktion

##### `update_transaction`
Transaktion aktualisieren.

- **Parameter:**
  - `household_id` (integer, required)
  - `transaction_id` (integer, required)
  - `category_id` (integer, optional)
  - `amount` (string, optional)
  - `description` (string, optional)
  - `date` (string, optional)
- **Returns:** Die aktualisierte Transaktion

##### `delete_transaction`
Transaktion löschen.

- **Parameter:**
  - `household_id` (integer, required)
  - `transaction_id` (integer, required)
- **Returns:** Bestätigung

---

#### Recurring Expense Management

##### `list_recurring_expenses`
Wiederkehrende Einträge eines Haushalts auflisten.

- **Parameter:**
  - `household_id` (integer, required)
- **Returns:** Array von RecurringExpenses (id, name, description, amount, frequency, active, start_date, end_date, category_id, category_name)

##### `create_recurring_expense`
Neuen wiederkehrenden Eintrag anlegen.

- **Parameter:**
  - `household_id` (integer, required)
  - `category_id` (integer, required)
  - `name` (string, required)
  - `description` (string, optional)
  - `amount` (string, required) — negativ = Ausgabe, positiv = Einnahme
  - `frequency` (string, required) — daily|weekday|weekly|biweekly|monthly|quarterly|yearly
  - `start_date` (string, required) — Format "YYYY-MM-DD"
  - `end_date` (string, optional) — Format "YYYY-MM-DD", leer = unbefristet
  - `active` (boolean, optional, default: true)
- **Returns:** Der erstellte Eintrag

##### `update_recurring_expense`
Wiederkehrenden Eintrag aktualisieren.

- **Parameter:**
  - `household_id` (integer, required)
  - `recurring_id` (integer, required)
  - `category_id` (integer, optional)
  - `name` (string, optional)
  - `description` (string, optional)
  - `amount` (string, optional)
  - `frequency` (string, optional)
  - `start_date` (string, optional)
  - `end_date` (string, optional)
  - `active` (boolean, optional)
- **Returns:** Der aktualisierte Eintrag

##### `delete_recurring_expense`
Wiederkehrenden Eintrag löschen.

- **Parameter:**
  - `household_id` (integer, required)
  - `recurring_id` (integer, required)
- **Returns:** Bestätigung

---

#### Schedule Overrides

##### `list_schedule_overrides`
Overrides für einen wiederkehrenden Eintrag auflisten.

- **Parameter:**
  - `household_id` (integer, required)
  - `recurring_id` (integer, required)
- **Returns:** Array von Overrides (id, effective_date, amount, frequency)

##### `create_schedule_override`
Neuen Override anlegen (ändert Betrag/Frequenz ab einem Stichtag).

- **Parameter:**
  - `household_id` (integer, required)
  - `recurring_id` (integer, required)
  - `effective_date` (string, required) — Format "YYYY-MM-DD"
  - `amount` (string, required)
  - `frequency` (string, required)
- **Returns:** Der erstellte Override

##### `update_schedule_override`
Override aktualisieren.

- **Parameter:**
  - `household_id` (integer, required)
  - `recurring_id` (integer, required)
  - `override_id` (integer, required)
  - `effective_date` (string, optional)
  - `amount` (string, optional)
  - `frequency` (string, optional)
- **Returns:** Der aktualisierte Override

##### `delete_schedule_override`
Override löschen.

- **Parameter:**
  - `household_id` (integer, required)
  - `recurring_id` (integer, required)
  - `override_id` (integer, required)
- **Returns:** Bestätigung

---

#### Summary / Analytics

##### `get_monthly_summary`
Monatliche Finanzübersicht für einen Haushalt.

- **Parameter:**
  - `household_id` (integer, required)
  - `month` (string, optional) — Format "YYYY-MM", default: aktueller Monat
- **Returns:** MonthlySummary mit:
  - Gesamteinnahmen / -ausgaben (einmalig + wiederkehrend)
  - Brutto-Einnahmen / -Ausgaben
  - Monatstotal
  - Aufschlüsselung nach Kategorien
  - Wiederkehrende Einträge gruppiert nach Frequenz

---

### Resources

Der MCP Server stellt folgende Read-Only Resources bereit:

##### `money-tracker://households`
Übersicht aller Haushalte mit Basis-Infos. Erlaubt dem LLM, Kontext über die verfügbaren Haushalte zu haben, ohne explizit ein Tool aufrufen zu müssen.

##### `money-tracker://households/{id}/summary/{month}`
Monatszusammenfassung als strukturierte Resource. Nützlich, damit der LLM-Client automatisch relevanten Finanzkontext laden kann.

---

### Prompts

Vordefinierte Prompt-Templates, die LLM-Clients dem Nutzer anbieten können:

##### `monthly_report`
Erstellt einen formatierten Monatsbericht.

- **Argumente:**
  - `household_id` (integer, required)
  - `month` (string, optional)
- **Prompt-Template:** Lädt Summary + Transaktionen + Recurring und erzeugt einen strukturierten Finanzbericht

##### `budget_analysis`
Analysiert die Ausgaben und gibt Empfehlungen.

- **Argumente:**
  - `household_id` (integer, required)
  - `months` (integer, optional, default: 3) — Anzahl der zu analysierenden Monate
- **Prompt-Template:** Lädt Summaries der letzten N Monate und analysiert Trends

##### `categorize_transaction`
Schlägt eine Kategorie für eine Transaktion vor.

- **Argumente:**
  - `household_id` (integer, required)
  - `description` (string, required)
  - `amount` (string, required)
- **Prompt-Template:** Lädt existierende Kategorien und schlägt basierend auf Beschreibung und Betrag die passendste vor

---

## Projektstruktur (Phase 1)

```
cmd/money-tracker/cmd/
  mcp.go                   Cobra "mcp" Subcommand (stdio transport setup)
internal/
  mcp/
    server.go              MCP Server Setup, Tool/Resource/Prompt Registration
    client.go              HTTP Client für Money Tracker REST API
    tools.go               Tool-Handler (list_households, create_transaction, ...)
    resources.go           Resource-Handler
    prompts.go             Prompt-Templates
```

Das `mcp` Subcommand wird in die bestehende Binary integriert — kein separates Build-Target nötig:
```bash
make build       # → bin/money-tracker        (enthält serve + mcp + migrate + version)
make build-dev   # → bin/money-tracker-dev    (dto., mit Dev-Mode)
```

### MCP Client Konfiguration (Claude Desktop / Claude Code)

```json
{
  "mcpServers": {
    "money-tracker": {
      "command": "/path/to/money-tracker",
      "args": ["mcp"],
      "env": {
        "MONEY_TRACKER_API_TOKEN": "mt_..."
      }
    }
  }
}
```

---

## Implementierungsreihenfolge

1. **MCP-001:** Projekt-Scaffolding — Cobra `mcp` Subcommand, MCP SDK Dependency, stdio Transport
2. **MCP-002:** API Client — HTTP Client mit Token-Auth für alle REST Endpoints
3. **MCP-003:** Household & Category Tools — list/create/update/delete
4. **MCP-004:** Transaction Tools — list/create/update/delete
5. **MCP-005:** Recurring Expense & Override Tools — list/create/update/delete
6. **MCP-006:** Summary Tool — get_monthly_summary
7. **MCP-007:** Resources — households Übersicht, monthly summary
8. **MCP-008:** Prompts — monthly_report, budget_analysis, categorize_transaction
9. **MCP-009:** Docs — README-Abschnitt, Konfigurationsanleitung
