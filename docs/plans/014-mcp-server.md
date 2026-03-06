# 014 — MCP Server for Money Tracker

## Motivation

An MCP (Model Context Protocol) server enables LLM clients (Claude Desktop, Claude Code, etc.) to interact directly with Money Tracker — creating transactions, querying budgets, generating summaries, without having to use the web UI.

## Architecture

### Phase 1: Local MCP Server (stdio-based)

The MCP server is integrated into the existing binary as a **Cobra subcommand** (`money-tracker mcp`). It communicates via **stdio** (JSON-RPC) with the LLM client and internally calls the **REST API** of the running Money Tracker server.

```
┌─────────────────┐     stdio (JSON-RPC)     ┌──────────────────────────┐     HTTP/REST     ┌──────────────────┐
│  LLM Client     │ ◄──────────────────────► │  money-tracker mcp       │ ◄──────────────► │  money-tracker   │
│  (Claude, etc.) │                           │  (Subcommand, same bin)  │                  │  serve           │
└─────────────────┘                           └──────────────────────────┘                  └──────────────────┘
```

**Invocation:**
```bash
money-tracker mcp                                        # Default: localhost:8080
money-tracker mcp --url http://myserver:9090             # Custom URL
MONEY_TRACKER_API_TOKEN=mt_... money-tracker mcp         # Token via ENV
money-tracker mcp --token mt_...                         # Token via Flag
```

**Configuration:**
- `--url` / `MONEY_TRACKER_MCP_URL` — Base URL of the API server (default: `http://localhost:8080`)
- `--token` / `MONEY_TRACKER_API_TOKEN` — Bearer token (`mt_...`) for authentication

**Phase 1 Advantages:**
- No OAuth infrastructure required — uses existing API token auth
- Quick to implement, immediately usable locally
- Full functionality via existing REST API
- One binary for everything — config, logging, buildinfo are reused

**Disadvantages / Deliberate Trade-offs:**
- Reimplementation: MCP tool layer duplicates API client logic
- Token must be manually created and configured
- Only usable locally (no remote access for hosted LLM clients)

### Phase 2: Remote MCP Server with OAuth 2.1 (Future)

Long-term, the MCP server should run as a **Remote HTTP SSE/Streamable HTTP** endpoint directly in the Money Tracker server, using OAuth 2.1 for authorization.

```
┌─────────────────┐    HTTP SSE / Streamable HTTP    ┌──────────────────────────────┐
│  LLM Client     │ ◄──────────────────────────────► │  Money Tracker Server        │
│  (Claude, etc.) │         + OAuth 2.1               │  (integrated MCP endpoint)   │
└─────────────────┘                                   └──────────────────────────────┘
```

**Phase 2 Prerequisites:**
- OAuth 2.1 Authorization Server (either self-hosted or external provider)
- Dynamic Client Registration (RFC 7591) or preconfigured clients
- PKCE flow for LLM clients
- Token scoping (which households may a client access?)
- MCP endpoint integrated directly into the Echo router

**Migration Phase 1 → Phase 2:**
- The MCP tool definitions and descriptions remain identical
- Only the transport changes (stdio → HTTP SSE) and the auth (API token → OAuth)
- The API client layer from Phase 1 is replaced by direct service layer calls

---

## MCP Server Specification (Phase 1)

### Server Info

```json
{
  "name": "money-tracker",
  "version": "0.1.0"
}
```

### Tools

#### Household Management

##### `list_households`
List all households of the authenticated user.

- **Parameters:** none
- **Returns:** Array of Households (id, name, description, currency, icon)

##### `create_household`
Create a new household.

- **Parameters:**
  - `name` (string, required) — Name of the household
  - `description` (string, optional) — Description
  - `currency` (string, required) — ISO 4217 currency code (e.g. "EUR")
  - `icon` (string, optional) — Material Icon name
- **Returns:** The created household

##### `update_household`
Update a household.

- **Parameters:**
  - `id` (integer, required) — Household ID
  - `name` (string, optional)
  - `description` (string, optional)
  - `currency` (string, optional)
  - `icon` (string, optional)
- **Returns:** The updated household

##### `delete_household`
Delete a household (cascades: all categories, transactions, recurring expenses).

- **Parameters:**
  - `id` (integer, required) — Household ID
- **Returns:** Confirmation

---

#### Category Management

##### `list_categories`
List categories of a household.

- **Parameters:**
  - `household_id` (integer, required)
- **Returns:** Array of Categories (id, name, icon)

##### `create_category`
Create a new category.

- **Parameters:**
  - `household_id` (integer, required)
  - `name` (string, required)
  - `icon` (string, optional)
- **Returns:** The created category

##### `update_category`
Update a category.

- **Parameters:**
  - `household_id` (integer, required)
  - `category_id` (integer, required)
  - `name` (string, optional)
  - `icon` (string, optional)
- **Returns:** The updated category

##### `delete_category`
Delete a category.

- **Parameters:**
  - `household_id` (integer, required)
  - `category_id` (integer, required)
- **Returns:** Confirmation

---

#### Transaction Management

##### `list_transactions`
List transactions of a household for a given month.

- **Parameters:**
  - `household_id` (integer, required)
  - `month` (string, optional) — Format "YYYY-MM", default: current month
- **Returns:** Array of Transactions (id, amount, description, date, category_id, category_name)

##### `create_transaction`
Create a new transaction.

- **Parameters:**
  - `household_id` (integer, required)
  - `category_id` (integer, required)
  - `amount` (string, required) — Decimal number as string, negative = expense, positive = income
  - `description` (string, required)
  - `date` (string, required) — Format "YYYY-MM-DD"
- **Returns:** The created transaction

##### `update_transaction`
Update a transaction.

- **Parameters:**
  - `household_id` (integer, required)
  - `transaction_id` (integer, required)
  - `category_id` (integer, optional)
  - `amount` (string, optional)
  - `description` (string, optional)
  - `date` (string, optional)
- **Returns:** The updated transaction

##### `delete_transaction`
Delete a transaction.

- **Parameters:**
  - `household_id` (integer, required)
  - `transaction_id` (integer, required)
- **Returns:** Confirmation

---

#### Recurring Expense Management

##### `list_recurring_expenses`
List recurring entries of a household.

- **Parameters:**
  - `household_id` (integer, required)
- **Returns:** Array of RecurringExpenses (id, name, description, amount, frequency, active, start_date, end_date, category_id, category_name)

##### `create_recurring_expense`
Create a new recurring entry.

- **Parameters:**
  - `household_id` (integer, required)
  - `category_id` (integer, required)
  - `name` (string, required)
  - `description` (string, optional)
  - `amount` (string, required) — negative = expense, positive = income
  - `frequency` (string, required) — daily|weekday|weekly|biweekly|monthly|quarterly|yearly
  - `start_date` (string, required) — Format "YYYY-MM-DD"
  - `end_date` (string, optional) — Format "YYYY-MM-DD", empty = indefinite
  - `active` (boolean, optional, default: true)
- **Returns:** The created entry

##### `update_recurring_expense`
Update a recurring entry.

- **Parameters:**
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
- **Returns:** The updated entry

##### `delete_recurring_expense`
Delete a recurring entry.

- **Parameters:**
  - `household_id` (integer, required)
  - `recurring_id` (integer, required)
- **Returns:** Confirmation

---

#### Schedule Overrides

##### `list_schedule_overrides`
List overrides for a recurring entry.

- **Parameters:**
  - `household_id` (integer, required)
  - `recurring_id` (integer, required)
- **Returns:** Array of Overrides (id, effective_date, amount, frequency)

##### `create_schedule_override`
Create a new override (changes amount/frequency from an effective date).

- **Parameters:**
  - `household_id` (integer, required)
  - `recurring_id` (integer, required)
  - `effective_date` (string, required) — Format "YYYY-MM-DD"
  - `amount` (string, required)
  - `frequency` (string, required)
- **Returns:** The created override

##### `update_schedule_override`
Update an override.

- **Parameters:**
  - `household_id` (integer, required)
  - `recurring_id` (integer, required)
  - `override_id` (integer, required)
  - `effective_date` (string, optional)
  - `amount` (string, optional)
  - `frequency` (string, optional)
- **Returns:** The updated override

##### `delete_schedule_override`
Delete an override.

- **Parameters:**
  - `household_id` (integer, required)
  - `recurring_id` (integer, required)
  - `override_id` (integer, required)
- **Returns:** Confirmation

---

#### Summary / Analytics

##### `get_monthly_summary`
Monthly financial overview for a household.

- **Parameters:**
  - `household_id` (integer, required)
  - `month` (string, optional) — Format "YYYY-MM", default: current month
- **Returns:** MonthlySummary with:
  - Total income / expenses (one-time + recurring)
  - Gross income / expenses
  - Monthly total
  - Breakdown by category
  - Recurring entries grouped by frequency

---

### Resources

The MCP server provides the following read-only resources:

##### `money-tracker://households`
Overview of all households with basic info. Allows the LLM to have context about available households without explicitly calling a tool.

##### `money-tracker://households/{id}/summary/{month}`
Monthly summary as a structured resource. Useful for the LLM client to automatically load relevant financial context.

---

### Prompts

Predefined prompt templates that LLM clients can offer to the user:

##### `monthly_report`
Creates a formatted monthly report.

- **Arguments:**
  - `household_id` (integer, required)
  - `month` (string, optional)
- **Prompt Template:** Loads summary + transactions + recurring and generates a structured financial report

##### `budget_analysis`
Analyzes spending and provides recommendations.

- **Arguments:**
  - `household_id` (integer, required)
  - `months` (integer, optional, default: 3) — Number of months to analyze
- **Prompt Template:** Loads summaries of the last N months and analyzes trends

##### `categorize_transaction`
Suggests a category for a transaction.

- **Arguments:**
  - `household_id` (integer, required)
  - `description` (string, required)
  - `amount` (string, required)
- **Prompt Template:** Loads existing categories and suggests the best match based on description and amount

---

## Project Structure (Phase 1)

```
cmd/money-tracker/cmd/
  mcp.go                   Cobra "mcp" Subcommand (stdio transport setup)
internal/
  mcp/
    server.go              MCP Server Setup, Tool/Resource/Prompt Registration
    client.go              HTTP Client for Money Tracker REST API
    tools.go               Tool handlers (list_households, create_transaction, ...)
    resources.go           Resource handlers
    prompts.go             Prompt templates
```

The `mcp` subcommand is integrated into the existing binary — no separate build target needed:
```bash
make build       # → bin/money-tracker        (includes serve + mcp + migrate + version)
make build-dev   # → bin/money-tracker-dev    (same, with dev mode)
```

### MCP Client Configuration (Claude Desktop / Claude Code)

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

## Implementation Order

1. **MCP-001:** Project scaffolding — Cobra `mcp` subcommand, MCP SDK dependency, stdio transport
2. **MCP-002:** API Client — HTTP client with token auth for all REST endpoints
3. **MCP-003:** Household & Category Tools — list/create/update/delete
4. **MCP-004:** Transaction Tools — list/create/update/delete
5. **MCP-005:** Recurring Expense & Override Tools — list/create/update/delete
6. **MCP-006:** Summary Tool — get_monthly_summary
7. **MCP-007:** Resources — households overview, monthly summary
8. **MCP-008:** Prompts — monthly_report, budget_analysis, categorize_transaction
9. **MCP-009:** Docs — README section, configuration guide
