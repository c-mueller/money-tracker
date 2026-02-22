# Plan 008: i18n Übersetzungen in JSON-Dateien auslagern

## Status: Done

## Kontext
Die i18n-Übersetzungen lagen als Go-Maps in `internal/i18n/de.go`, `en.go`, `frequencies.go`. Um neue Sprachen ohne Code-Änderung hinzufügen zu können, wurden die Übersetzungen in JSON-Dateien ausgelagert.

## Änderungen

### Neue Dateien
- `internal/i18n/locales/de.json` — Deutsche Übersetzungen + Frequenzen + Formatierung
- `internal/i18n/locales/en.json` — Englische Übersetzungen + Frequenzen + Formatierung

### Gelöschte Dateien
- `internal/i18n/de.go`
- `internal/i18n/en.go`
- `internal/i18n/frequencies.go`

### Geänderte Dateien
- `internal/i18n/i18n.go` — embed.FS + localeData struct + auto-load via fs.Glob + neue Getter (DateFormat, ThousandsSep, DecimalSep)
- `internal/api/template.go` — formatMoney/formatDate nutzen jetzt Bundle-Getter statt hardcoded if/else
- `internal/i18n/i18n_test.go` — Tests erweitert um neue Getter

## JSON-Struktur pro Sprache
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

Neue Sprache hinzufügen = nur neue JSON-Datei in `internal/i18n/locales/` anlegen.
