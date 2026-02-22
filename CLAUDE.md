# Money Tracker

## Quick Reference

```bash
make build       # Build production binary to bin/money-tracker
make build-dev   # Build dev binary to bin/money-tracker-dev (with auto-auth)
make run         # Build and start production server
make run-dev     # Build and start dev server (auto-auth, no OIDC needed)
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
  devmode/             Build-tag-controlled dev mode (dev vs prod stubs)
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
- Auth: OIDC in production builds; dev builds (`-tags=dev`) use auto-auth
- Config: ENV prefix `MONEY_TRACKER_`, e.g. `MONEY_TRACKER_SERVER_PORT=9090`
- After modifying ent schemas, run `make generate`

## Dev Mode

Dev mode is controlled via Go build tag `dev` — production binaries contain no dev code.

- `make run-dev` builds with `-tags=dev` and starts with auto-auth (no OIDC needed)
- `make run` builds a production binary that requires OIDC configuration
- `./bin/money-tracker version` shows `[DEV BUILD]` only for dev builds
