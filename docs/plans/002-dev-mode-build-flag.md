# 002 — Dev Mode via Build Flag

## Kontext

Dev Mode wurde von Runtime-Config (`cfg.Auth.OIDC.Issuer == ""`) auf Go Build Tags umgestellt. Produktivbinaries enthalten keinen Dev-Code mehr.

## Änderungen

### Neues Package: `internal/devmode/`
- `devmode_dev.go` (`//go:build dev`) — `Enabled = true`, echte `SetupUser()`-Logik
- `devmode_prod.go` (`//go:build !dev`) — `Enabled = false`, No-Op Stubs

### `internal/buildinfo/buildinfo.go`
- `String()` gibt `[DEV BUILD]` Suffix aus wenn `devmode.Enabled`

### `internal/middleware/auth.go`
- `devMode bool` Parameter entfernt, nutzt `devmode.Enabled` direkt

### `internal/api/server.go`
- `devMode` Feld aus Server-Struct entfernt, `devUserID` beibehalten

### `internal/api/router.go`
- `SetupAuth()` Signatur vereinfacht: `devMode bool` Parameter entfernt
- Auth-Middleware-Aufruf angepasst (kein `devMode` mehr)

### `cmd/money-tracker/cmd/serve.go`
- Runtime-Check `cfg.Auth.OIDC.Issuer == ""` entfernt
- Dev-User-Setup über `devmode.Enabled` und `devmode.SetupUser()`
- OIDC-Setup nur in Prod-Build (`!devmode.Enabled`)

### `tests/integration/testutil.go`
- `SetupAuth()` Aufruf an neue Signatur angepasst

### `Makefile`
- `build-dev` Target: `-tags=dev`, Output `bin/money-tracker-dev`
- `run-dev` Target: baut mit dev-Tag und startet

### `CLAUDE.md`
- Quick Reference um `make build-dev` / `make run-dev` ergänzt
- Dev Mode Doku aktualisiert: Build-Tag statt Runtime-Config

## Verifikation

```bash
make build && ./bin/money-tracker version
# → money-tracker ... (kein [DEV BUILD])

make build-dev && ./bin/money-tracker-dev version
# → money-tracker ... [DEV BUILD]

make test                    # ✓
go test -tags=dev ./... -count=1  # ✓
```
