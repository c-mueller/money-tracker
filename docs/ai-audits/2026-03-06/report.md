# Security Audit Report — Money Tracker

**Date:** 2026-03-06
**Scope:** Full application codebase (`icekalt.dev/money-tracker`)
**Auditor:** Claude Opus 4.6 (AI-assisted audit)
**Commit:** c185576 (master)

## Context & Threat Model

Money Tracker is a household budget tracking application designed for a **small, trusted user group** in a **local/private network** environment. It is not intended to be public-facing. This audit considers that context — findings like missing rate limiting or brute-force protection are **not flagged** as they are irrelevant for the intended deployment. The focus is on authorization correctness, data integrity, and defense-in-depth.

---

## Executive Summary

The codebase is generally well-structured from a security perspective. It uses an ORM (Ent) exclusively — **no SQL injection vectors exist**. Templates use Go's `html/template` with auto-escaping — **no XSS vulnerabilities found**. Static files are served from an embedded filesystem — **no path traversal risks**. API tokens use strong cryptography (256-bit entropy, SHA-256 hashed storage).

The main findings are:
- **Authorization gaps (IDOR)** in schedule overrides and API token deletion
- **No CSRF tokens** on web form routes (partially mitigated by SameSite=Lax cookies)
- **Missing security headers** (CSP, X-Frame-Options, etc.)
- **Minor CI/CD hygiene** issues

---

## Findings

### Critical

*None.*

---

### High

#### H1 — Schedule Override Endpoints Have No Authorization Checks — FIXED

**Status:** Fixed (2026-03-07)

**Location:** `internal/service/recurring_expense.go:135-181`, `internal/api/schedule_override_handler.go`

All four schedule override operations (List, Create, Update, Delete) performed **no ownership verification**. Any authenticated user could manipulate schedule overrides for recurring expenses belonging to other users by guessing/enumerating integer IDs.

**Resolution:** Added `authorizeRecurringExpense()` helper that verifies the recurring expense belongs to a household owned by the authenticated user. Applied to all four override operations (List, Create, Update, Delete).

---

### Medium

#### M1 — No CSRF Tokens on Web Form Routes

**Location:** `internal/api/router.go` (all web POST routes), `internal/auth/session.go:23`

Web mutation routes (creating transactions, households, categories, etc.) use session cookie authentication with no CSRF token. The session cookie has `SameSite=Lax`, which provides partial CSRF protection — modern browsers will not send the cookie on cross-site POST requests. However, `SameSite=Lax` alone is not a complete CSRF defense:

- Subdomain takeover scenarios bypass SameSite
- Older browser versions may not enforce SameSite correctly

**Context adjustment:** Given the local/private deployment model, the risk of a cross-site attack is significantly reduced. An attacker would need to be on the same network and trick a logged-in user into visiting a malicious page. `SameSite=Lax` provides adequate protection for this threat model.

**Recommendation (low priority):** Consider adding a synchronizer CSRF token or switching to `SameSite=Strict` for additional defense-in-depth.

#### M2 — API Token Deletion Has No Ownership Check — FIXED

**Status:** Fixed (2026-03-07)

**Location:** `internal/service/api_token.go:57-59`, `internal/api/token_handler.go:62-73`

`APITokenService.Delete` deleted any token by ID without verifying the authenticated user owned it.

**Resolution:** `Delete` now verifies the token belongs to the authenticated user by checking against their token list before deleting. Returns `ErrNotFound` if the token doesn't belong to the user.

#### M3 — No Security Headers — FIXED

**Status:** Fixed (2026-03-07)

**Location:** `internal/api/middleware.go`

The application set no security-related HTTP headers.

**Resolution:** Added `middleware.SecureWithConfig()` setting: `X-XSS-Protection`, `X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY`, `Referrer-Policy: strict-origin-when-cross-origin`, and a restrictive `Content-Security-Policy`.

#### M4 — API Token Expiry Not Enforced

**Location:** `internal/service/api_token.go:61-64`

The domain model has an `ExpiresAt` field, but `ValidateToken` does not check whether the token has expired. A token with a past `ExpiresAt` will be accepted. Additionally, `LastUsed` is never updated on token use despite the repository having an `UpdateLastUsed` method.

**Recommendation:** Check `ExpiresAt` during validation and call `UpdateLastUsed` on successful authentication.

#### M5 — Hardcoded Weak DB Credentials in docker-compose.yml — ACCEPTED

**Status:** Accepted (2026-03-07) — development artifact, not used in production.

---

### Low

#### L1 — Category Update Discards Household ID from URL

**Location:** `internal/api/category_handler.go:48`

`handleUpdateCategory` parses the household ID from the URL but discards it (`_, err := parseID(c, "id")`). The actual authorization happens via the category's stored `HouseholdID`, so this is not exploitable, but a request to `PUT /households/WRONG_ID/categories/VALID_CAT_ID` succeeds if the user owns the category — the URL is misleading.

**Recommendation:** Verify the URL's household ID matches the category's actual household.

#### L2 — RecurringExpense.GetByID Has No Ownership Check (Data Leak via Edit Form)

**Location:** `internal/api/web_handler.go:510-552`, `internal/service/recurring_expense.go:61-63`

The web edit form handler for recurring expenses fetches the expense without ownership verification. While the POST update endpoint does check ownership, the GET form leaks the expense data to unauthorized users who can guess the ID.

**Recommendation:** Add ownership verification in `GetByID` or in the handler before rendering.

#### L3 — Client-Side Cookie Sessions Cannot Be Revoked Server-Side

**Location:** `internal/auth/session.go`

Sessions use `gorilla/sessions.CookieStore` (client-side). Logout sets `MaxAge = -1` to instruct the browser to delete the cookie, but a copy of the cookie remains valid until its original MaxAge expires. There is no server-side session store to revoke.

**Context adjustment:** For a small trusted user group, this is acceptable. Server-side session management adds complexity.

#### L4 — Session Secret Auto-Generated on Each Restart

**Location:** `cmd/money-tracker/cmd/serve.go:65-71`

When `MONEY_TRACKER_AUTH_SESSION_SECRET` is not configured, a random secret is generated at startup. This invalidates all sessions on restart. No log warning is emitted.

**Recommendation:** Log a warning when the session secret is auto-generated, so operators are aware.

#### L5 — Secure Cookie Flag Hardcoded to `true`

**Location:** `internal/auth/session.go:22`

The session cookie always has `Secure: true`, meaning it will not be sent over plain HTTP. This is correct for production behind a TLS proxy, but breaks session auth when accessing the app directly over HTTP (even in dev mode).

**Recommendation:** Make the `Secure` flag configurable or derive it from the environment (e.g., `false` in dev builds).

#### L6 — CI: Overly Broad Permissions and Mutable Action Tags

**Location:** `.github/workflows/ci.yml`

- Top-level `permissions: { contents: write, packages: write }` applies to all jobs including test/build jobs that need only read access
- Third-party actions are pinned to mutable version tags (e.g., `actions/checkout@v4`) rather than commit SHAs
- Integration test job is a no-op placeholder that always passes

**Recommendation:** Set minimal permissions at top level; override per-job. Pin actions to commit SHAs.

#### L7 — Missing Input Validation in Service Layer

**Locations:**
- `internal/service/household.go:31` — `description` not validated (DB constraint exists but error is raw)
- `internal/service/household.go` / `category.go` — `icon` field has no allowlist
- `internal/service/api_token.go:36` — token `name` not validated before DB insert

These all have Ent schema constraints as a safety net, but service-layer validation would provide cleaner error messages.

---

### Informational

| Finding | Location | Notes |
|---------|----------|-------|
| OAuth2 state comparison is not constant-time | `internal/api/auth_handler.go:53` | Not practically exploitable for OAuth state |
| OIDC logout does not revoke upstream IdP session | `internal/api/auth_handler.go:97` | Common trade-off; document as known limitation |
| `/api/v1/openapi.yaml` is unauthenticated | `internal/api/router.go:43` | **Accepted** — intentional for local deployment |
| No CORS headers configured | `internal/api/middleware.go` | Restricts cross-origin access (secure default) |
| `InjectUserID` middleware is effectively dead code | `internal/middleware/context.go` | Runs before auth; always a no-op |
| Full request URI logged (incl. query params) | `internal/middleware/logging.go:25` | **Accepted** — no sensitive data in URLs currently |
| Postgres port bound to all interfaces in compose | `docker-compose.yml:22` | **Accepted** — development artifact |

---

## What's Done Well

- **No SQL injection** — Ent ORM used exclusively, no raw queries
- **No XSS** — `html/template` with auto-escaping throughout
- **No path traversal** — embedded filesystem with `http.FS`
- **Strong token cryptography** — 256-bit `crypto/rand`, SHA-256 hashed storage, `mt_` prefix for scanner detection
- **Dev mode via build tags** — `const Enabled = true/false` is compiled out; no runtime risk of accidental dev mode in production
- **Distroless Docker image** running as non-root
- **Proper OIDC implementation** — state parameter, audience validation, `go-oidc` library
- **Clean error responses** — internal errors not leaked to clients
- **No open redirects**
- **Dependencies up to date** — no known vulnerable packages

---

## Recommended Priority Actions

1. ~~**Fix authorization in schedule override endpoints** (H1)~~ — **FIXED**
2. ~~**Fix API token delete ownership check** (M2)~~ — **FIXED**
3. ~~**Add `middleware.Secure()` for security headers** (M3)~~ — **FIXED**
4. **Enforce token expiry in validation** (M4) — small change
5. **Log a warning when session secret is auto-generated** (L4) — small quality-of-life improvement

---

*This audit was performed via static code analysis. No dynamic testing, penetration testing, or dependency vulnerability scanning (`govulncheck`) was performed. Running `govulncheck ./...` is recommended as a follow-up.*
