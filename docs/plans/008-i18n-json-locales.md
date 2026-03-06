# Plan 008: Move i18n Translations to JSON Files

## Status: Done

## Context
The i18n translations were stored as Go maps in `internal/i18n/de.go`, `en.go`, `frequencies.go`. To allow adding new languages without code changes, translations were moved to JSON files.

## Changes

### New Files
- `internal/i18n/locales/de.json` — German translations + frequencies + formatting
- `internal/i18n/locales/en.json` — English translations + frequencies + formatting

### Deleted Files
- `internal/i18n/de.go`
- `internal/i18n/en.go`
- `internal/i18n/frequencies.go`

### Modified Files
- `internal/i18n/i18n.go` — embed.FS + localeData struct + auto-load via fs.Glob + new getters (DateFormat, ThousandsSep, DecimalSep)
- `internal/api/template.go` — formatMoney/formatDate now use bundle getters instead of hardcoded if/else
- `internal/i18n/i18n_test.go` — Tests extended with new getters

## JSON Structure per Language
```json
{
  "locale": "de",
  "date_format": "02.01.2006",
  "thousands_sep": ".",
  "decimal_sep": ",",
  "frequencies": { ... },
  "messages": { ... }
}
```

Adding a new language = just create a new JSON file in `internal/i18n/locales/`.
