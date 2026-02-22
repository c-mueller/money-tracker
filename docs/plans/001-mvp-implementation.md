# 001 — MVP Implementation

**Datum**: 2026-02-22
**Commit**: e748c17

## Kontext

Greenfield Go-Projekt — deterministisches Haushaltsbuch-System mit Rücklagenlogik, Multi-Haushalt-Fähigkeit und REST-API. Backend-first, SQLite-Default, KI-ready vorbereitet (aber ohne KI in MVP).

**Modul**: `icekalt.dev/money-tracker` | **Go**: 1.24.3

---

## Phasen

### Phase 1: Build-Tooling, CLI-Skeleton, Config

- `Makefile` mit Targets: build, run, test, lint, clean, generate (ldflags für Version)
- `internal/buildinfo/buildinfo.go` — Version/Commit/BuildDate via ldflags
- `internal/config/` — Config-Struct + Viper-Loader (ENV-Prefix `MONEY_TRACKER_`)
- `internal/logging/logger.go` — Zap-Logger-Factory
- `cmd/money-tracker/` — Cobra CLI mit `serve`, `migrate`, `version` Commands
- **Deps**: cobra, viper, zap

### Phase 2: Domain-Typen

- `internal/domain/money.go` — `Money = decimal.Decimal` + Helpers
- `internal/domain/frequency.go` — Enum: monthly, quarterly, yearly, weekly, biweekly, daily, weekday
- Entity-Structs: User, Household, Category, Transaction, RecurringExpense, APIToken
- `internal/domain/reserves.go` — `NormalizeToMonthly()` für alle Frequenzen
- `internal/domain/errors.go` — Sentinel Errors + ValidationError-Typ
- Table-driven Tests für Reserven-Normalisierung
- **Deps**: shopspring/decimal

### Phase 3: Ent-Schema & Repository-Layer

- Ent-Schemas in `ent/schema/` für alle 6 Entities
- `internal/domain/repository.go` — Interfaces
- `internal/repository/` — Ent-basierte Implementierungen + Converter
- `internal/repository/db.go` — SQLite/Postgres Client-Factory
- Decimal-Strategie: Amount als String in DB, Konvertierung im Repository-Layer
- **Deps**: ent, modernc.org/sqlite, jackc/pgx

### Phase 4: Service-Layer

- Business-Logic für alle Entities in `internal/service/`
- Autorisierung: User-ID aus Context, Ownership-Check
- Summary-Service: Recurring normalisieren + Einmalbuchungen summieren
- API-Token: Generate → SHA-256 Hash → Store, Klartext nur bei Erstellung

### Phase 5: HTTP-Server & Middleware

- `internal/api/server.go` — Echo-Server mit Graceful Shutdown (SIGINT/SIGTERM)
- `internal/middleware/` — Request-ID (UUID v4), Logging (Zap), Recovery, Context-Injection
- `internal/api/response.go` — Domain-Error → HTTP-Status Mapping
- `internal/api/request.go` — ID-Parsing, Month-Parsing
- **Deps**: echo/v4, google/uuid

### Phase 6: REST-API Handler (CRUD)

- Alle CRUD-Endpoints in `internal/api/`
- DTOs in `internal/api/dto.go`
- **Routen**:
  - `GET/POST /api/v1/households`, `PUT/DELETE .../households/{id}`
  - `GET/POST .../categories`, `PUT/DELETE .../categories/{id}`
  - `GET/POST .../transactions`, `DELETE .../transactions/{id}` (GET mit `?month=YYYY-MM`)
  - `GET/POST/PUT/DELETE .../recurring-expenses`
  - `GET .../summary?month=YYYY-MM`
  - `GET/POST/DELETE /api/v1/tokens`
- **HTTP-Statuscodes**: 200, 201, 204, 400, 404, 422, 500

### Phase 7: Authentifizierung (OIDC + API-Tokens)

- `internal/auth/` — OIDC-Provider, Session-Store (Secure Cookie), Token Generate/Hash
- `internal/middleware/auth.go` — Session-Cookie ODER Bearer-Token
- Auth-Handler: `/auth/login`, `/auth/callback`, `/auth/logout`
- Dev-Modus: Auto-Auth wenn `auth.oidc.issuer` leer
- **Deps**: go-oidc, oauth2, gorilla/sessions

### Phase 8: Server-Rendered Frontend

- `web/embed.go` — `//go:embed` für Templates + Static
- Templates: layout, dashboard, household detail/form, category list, recurring list, token list
- Bootstrap 5 (CDN), app.js (Confirm-Dialogs, Copy-to-Clipboard)
- Template-Engine mit Helpers: formatMoney, formatDate

### Phase 9: Input-Validation & Edge Cases

- `internal/domain/validate.go` — ValidateCurrency, ValidateAmount, ValidateHouseholdName, etc.
- Validierung in allen Service-Methoden
- Regeln: Household-Name 1-100, Category-Name 1-50, Amount max 999999999.99, Description max 500

### Phase 10: Integration-Tests, Doku, Polish

- `tests/integration/` — Full-Flow Test (Health → Household → Category → Recurring → Transaction → Summary → Delete)
- Validation-Tests
- `CLAUDE.md`, `.golangci.yml`
- `Dockerfile` (Multi-Stage, distroless), `docker-compose.yml` (Postgres + App)

---

## Dependency-Übersicht

| Phase | Neue Dependencies |
|-------|------------------|
| 1 | cobra, viper, zap |
| 2 | shopspring/decimal |
| 3 | ent, modernc.org/sqlite, jackc/pgx |
| 5 | echo/v4, google/uuid |
| 7 | go-oidc, oauth2, gorilla/sessions |
