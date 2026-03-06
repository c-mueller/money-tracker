# 002 — Dev Mode via Build Flag

## Context

Dev mode was migrated from runtime config (`cfg.Auth.OIDC.Issuer == ""`) to Go build tags. Production binaries no longer contain any dev code.

## Changes

### New Package: `internal/devmode/`
- `devmode_dev.go` (`//go:build dev`) — `Enabled = true`, actual `SetupUser()` logic
- `devmode_prod.go` (`//go:build !dev`) — `Enabled = false`, no-op stubs

### `internal/buildinfo/buildinfo.go`
- `String()` outputs `[DEV BUILD]` suffix when `devmode.Enabled`

### `internal/middleware/auth.go`
- `devMode bool` parameter removed, uses `devmode.Enabled` directly

### `internal/api/server.go`
- `devMode` field removed from server struct, `devUserID` retained

### `internal/api/router.go`
- `SetupAuth()` signature simplified: `devMode bool` parameter removed
- Auth middleware call updated (no more `devMode`)

### `cmd/money-tracker/cmd/serve.go`
- Runtime check `cfg.Auth.OIDC.Issuer == ""` removed
- Dev user setup via `devmode.Enabled` and `devmode.SetupUser()`
- OIDC setup only in prod build (`!devmode.Enabled`)

### `tests/integration/testutil.go`
- `SetupAuth()` call updated to new signature

### `Makefile`
- `build-dev` target: `-tags=dev`, output `bin/money-tracker-dev`
- `run-dev` target: builds with dev tag and starts

### `CLAUDE.md`
- Quick reference updated with `make build-dev` / `make run-dev`
- Dev mode docs updated: build tag instead of runtime config

## Verification

```bash
make build && ./bin/money-tracker version
# → money-tracker ... (no [DEV BUILD])

make build-dev && ./bin/money-tracker-dev version
# → money-tracker ... [DEV BUILD]

make test                    # ✓
go test -tags=dev ./... -count=1  # ✓
```
