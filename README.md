# Money Tracker

A self-hosted household budget tracker for managing income, expenses, and recurring transactions across multiple households.

## Features

- **Multi-Household Support** — Manage separate budgets for different households, each with its own currency (ISO 4217)
- **Transaction Tracking** — Record income and expenses with categories, descriptions, and dates
- **Recurring Transactions** — Define recurring income/expenses with flexible frequencies (daily, weekday, weekly, biweekly, monthly, quarterly, yearly)
- **Schedule Overrides** — Temporarily adjust amount or frequency for specific dates on recurring transactions
- **Monthly Summaries** — Dashboard with income/expense breakdown, category analysis, and net result per month
- **REST API** — Full CRUD API with OpenAPI/Swagger documentation at `/swagger/`
- **GraphQL API** — Alternative GraphQL endpoint at `/graphql` with playground at `/playground`
- **MCP Server** — Model Context Protocol integration for AI assistants (Claude Desktop, Claude Code, etc.)
- **Internationalization** — German and English UI
- **OIDC Authentication** — Production-ready authentication via any OpenID Connect provider (Keycloak, Authentik, Auth0, etc.)
- **API Tokens** — Token-based authentication for programmatic access and MCP

## Quick Start

### Docker Compose (recommended)

```yaml
services:
  app:
    image: ghcr.io/c-mueller/money-tracker:latest
    ports:
      - "8080:8080"
    environment:
      MONEY_TRACKER_DATABASE_DRIVER: postgres
      MONEY_TRACKER_DATABASE_DSN: postgres://money:money@db:5432/money_tracker?sslmode=disable
      MONEY_TRACKER_AUTH_SESSION_SECRET: change-me-to-a-random-string
      MONEY_TRACKER_AUTH_OIDC_ISSUER: https://auth.example.com/realms/myrealm
      MONEY_TRACKER_AUTH_OIDC_CLIENT_ID: money-tracker
      MONEY_TRACKER_AUTH_OIDC_CLIENT_SECRET: your-client-secret
      MONEY_TRACKER_AUTH_OIDC_REDIRECT_URL: https://money.example.com/auth/callback
    depends_on:
      db:
        condition: service_healthy

  db:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: money
      POSTGRES_PASSWORD: money
      POSTGRES_DB: money_tracker
    volumes:
      - pgdata:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-ONLY", "pg_isready", "-U", "money", "-d", "money_tracker"]
      interval: 5s
      timeout: 5s
      retries: 5

volumes:
  pgdata:
```

### Binary

Download the binary for your platform from the [releases page](https://github.com/c-mueller/money-tracker/releases) and run:

```bash
# SQLite (default, no database setup needed)
export MONEY_TRACKER_AUTH_SESSION_SECRET=change-me
export MONEY_TRACKER_AUTH_OIDC_ISSUER=https://auth.example.com/realms/myrealm
export MONEY_TRACKER_AUTH_OIDC_CLIENT_ID=money-tracker
export MONEY_TRACKER_AUTH_OIDC_CLIENT_SECRET=your-client-secret
export MONEY_TRACKER_AUTH_OIDC_REDIRECT_URL=http://localhost:8080/auth/callback

./money-tracker serve
```

The server starts on port 8080 by default.

## Configuration

All options are configured via environment variables with the `MONEY_TRACKER_` prefix.

### Server

| Variable | Default | Description |
|---|---|---|
| `MONEY_TRACKER_SERVER_HOST` | `0.0.0.0` | Listen address |
| `MONEY_TRACKER_SERVER_PORT` | `8080` | Listen port |

### Database

| Variable | Default | Description |
|---|---|---|
| `MONEY_TRACKER_DATABASE_DRIVER` | `sqlite` | Database driver (`sqlite` or `postgres`) |
| `MONEY_TRACKER_DATABASE_DSN` | `money-tracker.db?_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)` | Connection string |

PostgreSQL DSN example: `postgres://user:pass@host:5432/dbname?sslmode=disable`

### Authentication (OIDC)

Money Tracker uses OpenID Connect for authentication in production. Any OIDC-compliant provider works (Keycloak, Authentik, Auth0, Authelia, Kanidm, etc.).

| Variable | Required | Description |
|---|---|---|
| `MONEY_TRACKER_AUTH_OIDC_ISSUER` | Yes | OIDC issuer URL (e.g. `https://auth.example.com/realms/myrealm`) |
| `MONEY_TRACKER_AUTH_OIDC_CLIENT_ID` | Yes | OAuth2 client ID |
| `MONEY_TRACKER_AUTH_OIDC_CLIENT_SECRET` | Yes | OAuth2 client secret |
| `MONEY_TRACKER_AUTH_OIDC_REDIRECT_URL` | Yes | Callback URL — must be `https://<your-domain>/auth/callback` |

#### OIDC Provider Setup

1. Create a new **confidential/private client** in your OIDC provider
2. Set the **redirect URI** to `https://<your-domain>/auth/callback`
3. Ensure the client has access to the scopes: `openid`, `profile`, `email`
4. Copy the issuer URL, client ID, and client secret into the environment variables above

**Example: Keycloak**
- Create a new client in your realm with "Client authentication" enabled
- Set Valid Redirect URIs to `https://money.example.com/auth/callback`
- Use issuer: `https://keycloak.example.com/realms/<your-realm>`

**Example: Authentik**
- Create a new OAuth2/OpenID Provider and Application
- Set redirect URI to `https://money.example.com/auth/callback`
- Use issuer: `https://authentik.example.com/application/o/<slug>/`

### Session

| Variable | Default | Description |
|---|---|---|
| `MONEY_TRACKER_AUTH_SESSION_SECRET` | — | Secret key for session cookies (required, use a random string) |
| `MONEY_TRACKER_AUTH_SESSION_MAX_AGE` | `86400` | Session lifetime in seconds (default: 24h) |

### Other

| Variable | Default | Description |
|---|---|---|
| `MONEY_TRACKER_LANGUAGE` | `de` | Default language (`de` or `en`) |
| `MONEY_TRACKER_LOGGING_LEVEL` | `info` | Log level (`debug`, `info`, `warn`, `error`) |

## MCP Server

Money Tracker includes a [Model Context Protocol](https://modelcontextprotocol.io/) server for integration with AI assistants like Claude.

### Setup

1. Create an API token in the Money Tracker UI under user settings
2. Run the MCP server:

```bash
money-tracker mcp --url http://localhost:8080 --token <your-api-token>
```

Or via environment variables:
```bash
export MONEY_TRACKER_MCP_URL=http://localhost:8080
export MONEY_TRACKER_MCP_TOKEN=your-api-token
money-tracker mcp
```

### Claude Desktop / Claude Code

Add to your MCP configuration:

```json
{
  "mcpServers": {
    "money-tracker": {
      "command": "/path/to/money-tracker",
      "args": ["mcp"],
      "env": {
        "MONEY_TRACKER_MCP_URL": "http://localhost:8080",
        "MONEY_TRACKER_MCP_TOKEN": "your-api-token"
      }
    }
  }
}
```

### Capabilities

**Tools:** Full CRUD for households, categories, transactions, recurring expenses, schedule overrides, and monthly summaries.

**Prompts:**
- `monthly_report` — Generate a formatted monthly financial report
- `budget_analysis` — Analyze spending patterns across multiple months
- `categorize_transaction` — Suggest the best category for a transaction

## API Documentation

- **Swagger UI:** `http://localhost:8080/swagger/`
- **GraphQL Playground:** `http://localhost:8080/playground`

API tokens for programmatic access can be created in the UI under user settings and used via `Authorization: Bearer <token>` header.

## Development

```bash
make build-dev    # Build with auto-auth (no OIDC needed)
make run-dev      # Build and start dev server
make test         # Run unit tests
make lint         # Run linter
make generate     # Regenerate Ent + gqlgen code
```

Dev builds (`-tags=dev`) skip OIDC and auto-authenticate as a dev user.

## License

All rights reserved.
