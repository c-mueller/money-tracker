# Money Tracker

## Quick Reference

```bash
make build       # Build binary to bin/money-tracker
make run         # Build and start server
make test        # Run unit tests
make test-integration  # Run integration tests (requires -tags=integration)
make lint        # Run golangci-lint
make generate    # Run go generate (ent codegen)
make clean       # Remove build artifacts
```

## Architecture

- **Go 1.24.3**, module `icekalt.dev/money-tracker`
- **Ent** for ORM/schema management (schemas in `ent/schema/`)
- **Echo v4** for HTTP routing
- **SQLite** (default) or **PostgreSQL** for persistence
- **shopspring/decimal** for money — stored as strings in DB, never floats
- **Zap** for structured logging
- **Cobra/Viper** for CLI and config

## Project Layout

```
cmd/money-tracker/     CLI entrypoint + cobra commands
internal/
  api/                 HTTP handlers, DTOs, server, templates
  auth/                OIDC + session + token auth
  buildinfo/           Version info (set via ldflags)
  config/              Config structs + viper loader
  domain/              Pure domain types, validation, business rules
  logging/             Zap logger factory
  middleware/           Echo middleware (auth, logging, recovery, request ID)
  repository/          Ent-based repo implementations
  service/             Business logic layer
ent/schema/            Ent schema definitions
web/                   Embedded templates + static assets
tests/integration/     Integration tests
```

## Key Conventions

- Domain types are in `internal/domain/` — no external deps except shopspring/decimal
- Money amounts: use `decimal.Decimal`, stored as strings in DB
- Frequencies: daily, weekday, weekly, biweekly, monthly, quarterly, yearly
- Auth: OIDC in production, dev-mode auto-auth when `auth.oidc.issuer` is empty
- Config: ENV prefix `MONEY_TRACKER_`, e.g. `MONEY_TRACKER_SERVER_PORT=9090`
- After modifying ent schemas, run `make generate`

## Dev Mode

Server auto-creates a dev user and skips auth when no OIDC issuer is configured.
Just run `make run` and access http://localhost:8080.
