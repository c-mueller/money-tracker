# Plan 011: Swagger UI for REST API

## Overview
Swagger UI for interactive documentation and testing of the REST API, without additional Go dependencies.

## Approach
- **OpenAPI 3.0.1 YAML** manually written, embedded via `go:embed`
- **Swagger UI** as a minimal HTML page with CDN-loaded `swagger-ui-dist@5.18.2`
- **Zero new Go dependencies**

## New Files

| File | Description |
|---|---|
| `web/static/openapi.yaml` | OpenAPI 3.0.1 spec — all 21 REST endpoints |
| `web/static/swagger/index.html` | Swagger UI HTML (CDN: swagger-ui-dist@5.18.2) |
| `internal/api/swagger_handler.go` | Handler for spec + UI |
| `tests/integration/swagger_test.go` | Integration tests |

## Modified Files

| File | Change |
|---|---|
| `internal/api/router.go` | 2 new routes added |

## Routing

| Route | Auth | Description |
|---|---|---|
| `GET /api/v1/openapi.yaml` | No | OpenAPI spec (YAML) |
| `GET /swagger` | Session/Token | Swagger UI HTML page |

## Details

### OpenAPI Spec
- All endpoints from `router.go` documented
- Schemas derived exactly from `dto.go` + `token_handler.go`
- Security: `bearerAuth` (HTTP Bearer), applied globally
- Money: `type: string` (never float)
- `/health` with `security: []` (no auth)

### Swagger UI
- CDN: `unpkg.com/swagger-ui-dist@5.18.2` (pinned version)
- `tryItOutEnabled: true` — test API calls directly
- Bearer token auth via "Authorize" button

### Tests
1. `GET /api/v1/openapi.yaml` without auth → 200
2. `GET /swagger` without auth → 401
3. `GET /swagger` with session cookie → 200
4. `GET /swagger` with Bearer token → 200
