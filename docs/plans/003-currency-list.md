# 003 — Predefined Currency List

## Summary

Replaced the free-text currency input on the household creation form with a dropdown of 13 European currencies + USD. Custom currencies are still supported via a "Custom…" option that reveals a text input.

## Changes

- **`web/static/currencies.json`** — New file with 14 predefined currencies (code, symbol, label)
- **`internal/api/template.go`** — Added `Currency` struct, currency loading from embedded JSON, `currencySymbol` and `isCurrencyCode` template functions
- **`internal/api/server.go`** — Added `renderer *TemplateRenderer` field to Server struct
- **`internal/api/router.go`** — Store renderer reference on server after creation
- **`internal/api/web_handler.go`** — Added `Currencies` field to `pageData`, pass currencies in `handleWebHouseholdNew`
- **`web/templates/household/form.html`** — Select dropdown with predefined currencies, "Custom…" option with toggle JS

## Not Changed

- `internal/domain/validate.go` — `^[A-Z]{3}$` regex unchanged
- `web/embed.go` — already embeds `static/`
- `ent/schema/household.go` — `MaxLen(3)` unchanged
- `formatMoney` template function — unchanged
