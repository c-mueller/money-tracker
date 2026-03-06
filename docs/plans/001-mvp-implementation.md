# 001 — MVP Implementation

**Date**: 2026-02-22
**Commit**: e748c17

## Context

Greenfield Go project — deterministic household budget system with reserve logic, multi-household support, and REST API. Backend-first, SQLite default, AI-ready (but no AI in MVP).

**Module**: `icekalt.dev/money-tracker` | **Go**: 1.24.3

---

## Phases

### Phase 1: Build Tooling, CLI Skeleton, Config

- `Makefile` with targets: build, run, test, lint, clean, generate (ldflags for version)
- `internal/buildinfo/buildinfo.go` — Version/Commit/BuildDate via ldflags
- `internal/config/` — Config struct + Viper loader (ENV prefix `MONEY_TRACKER_`)
- `internal/logging/logger.go` — Zap logger factory
- `cmd/money-tracker/` — Cobra CLI with `serve`, `migrate`, `version` commands
- **Deps**: cobra, viper, zap

### Phase 2: Domain Types

- `internal/domain/money.go` — `Money = decimal.Decimal` + helpers
- `internal/domain/frequency.go` — Enum: monthly, quarterly, yearly, weekly, biweekly, daily, weekday
- Entity structs: User, Household, Category, Transaction, RecurringExpense, APIToken
- `internal/domain/reserves.go` — `NormalizeToMonthly()` for all frequencies
- `internal/domain/errors.go` — Sentinel errors + ValidationError type
- Table-driven tests for reserve normalization
- **Deps**: shopspring/decimal

### Phase 3: Ent Schema & Repository Layer

- Ent schemas in `ent/schema/` for all 6 entities
- `internal/domain/repository.go` — Interfaces
- `internal/repository/` — Ent-based implementations + converters
- `internal/repository/db.go` — SQLite/Postgres client factory
- Decimal strategy: Amount as string in DB, conversion in repository layer
- **Deps**: ent, modernc.org/sqlite, jackc/pgx

### Phase 4: Service Layer

- Business logic for all entities in `internal/service/`
- Authorization: User ID from context, ownership check
- Summary service: Normalize recurring + sum one-time transactions
- API token: Generate → SHA-256 hash → store, plaintext only on creation

### Phase 5: HTTP Server & Middleware

- `internal/api/server.go` — Echo server with graceful shutdown (SIGINT/SIGTERM)
- `internal/middleware/` — Request ID (UUID v4), logging (Zap), recovery, context injection
- `internal/api/response.go` — Domain error → HTTP status mapping
- `internal/api/request.go` — ID parsing, month parsing
- **Deps**: echo/v4, google/uuid

### Phase 6: REST API Handlers (CRUD)

- All CRUD endpoints in `internal/api/`
- DTOs in `internal/api/dto.go`
- **Routes**:
  - `GET/POST /api/v1/households`, `PUT/DELETE .../households/{id}`
  - `GET/POST .../categories`, `PUT/DELETE .../categories/{id}`
  - `GET/POST .../transactions`, `DELETE .../transactions/{id}` (GET with `?month=YYYY-MM`)
  - `GET/POST/PUT/DELETE .../recurring-expenses`
  - `GET .../summary?month=YYYY-MM`
  - `GET/POST/DELETE /api/v1/tokens`
- **HTTP Status Codes**: 200, 201, 204, 400, 404, 422, 500

### Phase 7: Authentication (OIDC + API Tokens)

- `internal/auth/` — OIDC provider, session store (secure cookie), token generate/hash
- `internal/middleware/auth.go` — Session cookie OR Bearer token
- Auth handlers: `/auth/login`, `/auth/callback`, `/auth/logout`
- Dev mode: Auto-auth when `auth.oidc.issuer` is empty
- **Deps**: go-oidc, oauth2, gorilla/sessions

### Phase 8: Server-Rendered Frontend

- `web/embed.go` — `//go:embed` for templates + static
- Templates: layout, dashboard, household detail/form, category list, recurring list, token list
- Bootstrap 5 (CDN), app.js (confirm dialogs, copy-to-clipboard)
- Template engine with helpers: formatMoney, formatDate

### Phase 9: Input Validation & Edge Cases

- `internal/domain/validate.go` — ValidateCurrency, ValidateAmount, ValidateHouseholdName, etc.
- Validation in all service methods
- Rules: Household name 1-100, Category name 1-50, Amount max 999999999.99, Description max 500

### Phase 10: Integration Tests, Docs, Polish

- `tests/integration/` — Full-flow test (Health → Household → Category → Recurring → Transaction → Summary → Delete)
- Validation tests
- `CLAUDE.md`, `.golangci.yml`
- `Dockerfile` (multi-stage, distroless), `docker-compose.yml` (Postgres + App)

---

## Dependency Overview

| Phase | New Dependencies |
|-------|------------------|
| 1 | cobra, viper, zap |
| 2 | shopspring/decimal |
| 3 | ent, modernc.org/sqlite, jackc/pgx |
| 5 | echo/v4, google/uuid |
| 7 | go-oidc, oauth2, gorilla/sessions |
