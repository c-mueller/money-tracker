# Plan 011: Swagger UI für REST API

## Übersicht
Swagger UI zur interaktiven Dokumentation und Test der REST-API, ohne zusätzliche Go-Dependencies.

## Ansatz
- **OpenAPI 3.0.1 YAML** manuell geschrieben, via `go:embed` eingebettet
- **Swagger UI** als minimale HTML-Seite mit CDN-geladenem `swagger-ui-dist@5.18.2`
- **Null neue Go-Dependencies**

## Neue Dateien

| Datei | Beschreibung |
|---|---|
| `web/static/openapi.yaml` | OpenAPI 3.0.1 Spec — alle 21 REST-Endpoints |
| `web/static/swagger/index.html` | Swagger UI HTML (CDN: swagger-ui-dist@5.18.2) |
| `internal/api/swagger_handler.go` | Handler für Spec + UI |
| `tests/integration/swagger_test.go` | Integration-Tests |

## Geänderte Dateien

| Datei | Änderung |
|---|---|
| `internal/api/router.go` | 2 neue Routen hinzugefügt |

## Routing

| Route | Auth | Beschreibung |
|---|---|---|
| `GET /api/v1/openapi.yaml` | Nein | OpenAPI Spec (YAML) |
| `GET /swagger` | Session/Token | Swagger UI HTML-Seite |

## Details

### OpenAPI Spec
- Alle Endpoints aus `router.go` dokumentiert
- Schemas exakt aus `dto.go` + `token_handler.go` abgeleitet
- Security: `bearerAuth` (HTTP Bearer), global angewandt
- Money: `type: string` (nie float)
- `/health` mit `security: []` (kein Auth)

### Swagger UI
- CDN: `unpkg.com/swagger-ui-dist@5.18.2` (gepinnte Version)
- `tryItOutEnabled: true` — direkt API-Calls testen
- Bearer Token Auth via "Authorize" Button

### Tests
1. `GET /api/v1/openapi.yaml` ohne Auth → 200
2. `GET /swagger` ohne Auth → 401
3. `GET /swagger` mit Session-Cookie → 200
4. `GET /swagger` mit Bearer Token → 200
