# Infographic — Endpoints

**Module:** `webdesa/api`
**Last updated:** 2026-08-23

## Routes

### Public (no auth)

| Method | Path | Notes |
|---|---|---|
| GET | `/api/v1/public/infographic/list` | **Only accepts `page` + `limit`.** Other query params silently dropped (see Known Issues). |
| GET | `/api/v1/public/infographic/{id}` | Issues 5m token, rate-limited 30/h/IP, writes access log |
| GET | `/api/v1/public/infographic/categories` | |

### Admin (RBAC: `infographic:read`)

| Method | Path | Notes |
|---|---|---|
| GET | `/api/v1/infographic` | Query: `section_name`, `state`, `q`, `category`, `page`, `limit` |
| GET | `/api/v1/infographic/{id}` | Issues 10m token, `LogAccess("admin_detail", 10m)` |
| GET | `/api/v1/infographic/sections/names` | Distinct section names |
| GET | `/api/v1/infographic/access-logs` | Query: `infographic_id` (UUID, optional), `ip` (optional), `page`, `limit` (default 20). Filters AND-combine. |

### Admin (RBAC: `infographic:write`)

| Method | Path | Notes |
|---|---|---|
| POST | `/api/v1/infographic` | Body `{component_id, component_type, section_name, section_endpoint, category?, state}` |
| POST | `/api/v1/infographic/preview/token` | Body `{component_id, component_type}` — generates JWT without creating record, 10m |
| PUT | `/api/v1/infographic/{id}` | |
| DELETE | `/api/v1/infographic/{id}` | |
| POST | `/api/v1/infographic/categories` | Body `{name}` |
| DELETE | `/api/v1/infographic/categories/{id}` | 409 if in use |

## Request/Response

### Detail response shape (admin + public)

```json
{
  "id": "uuid",
  "component_id": 123,
  "component_type": "dashboard",
  "section_name": "Statistik",
  "section_endpoint": "/statistik",
  "category_id": "uuid",
  "category": {"id": "uuid", "name": "Statistik"},
  "state": true,
  "created_at": "RFC3339",
  "updated_at": "RFC3339",
  "token": "<jwt>"
}
```

**`token` field is `omitempty`** — only present on **detail** responses (admin + public). List responses omit it.

### List response (admin)

Same shape per item but **omits `token`**.

### List response (public)

Same shape per item, **omits `token`**. Filters silently dropped (see Known Issues).

### Preview token (admin)

```
POST /api/v1/infographic/preview/token
{"component_id": 123, "component_type": "dashboard"}
```

**Response**:
```json
{"token": "<jwt>"}
```

10-minute TTL. Does **not** persist any record or log access.

### Create / Update (admin)

```
POST /api/v1/infographic
PUT  /api/v1/infographic/{id}
{
  "component_id": 123,
  "component_type": "dashboard",
  "section_name": "Statistik",
  "section_endpoint": "/statistik",
  "category": "uuid (optional)",
  "state": true
}
```

### Access log list (admin)

```
GET /api/v1/infographic/access-logs?infographic_id=...&ip=...&page=1&limit=20
```

**Response**: paginated list of `AccessLog` objects (each carries `infographic_id`, `component_id`, `component_type`, `endpoint`, `ip_address`, `user_agent`, `referer`, `token_issued_at`, `token_expires_at`, `created_at`).

## Validation

### Handler-level

- `component_id` must be `> 0`
- `component_type` must be `dashboard` or `question`
- `section_name` non-empty, ≤ 100 chars
- `section_endpoint` ≤ 255 chars, must start with `/`
- `category` must be valid UUID (if present)
- Multipart file constraints not applicable — body is JSON

### Domain-level (`domain/infographic/infographic.go`)

Same rules as handler, plus:

- `state` must be BOOLEAN
- `ComponentType` wraps the string and validates against `ComponentTypeQuestion` / `ComponentTypeDashboard` constants

## RBAC

- `infographic:read` — admin list, detail, section names, access log list
- `infographic:write` — create, update, delete, preview token, category CRUD

> **Drift**: `rbac/rbac_policy.csv` declares `admin,infographic,read` but **no `admin,infographic,write` row**. Routes guarded with `infographic:write` reject the built-in `admin` user with 403. See Known Issues.

## Known Issues (Endpoint-specific)

- **Public list filters silently dropped.** `GET /api/v1/public/infographic/list` parses only `page` + `limit`. `q`, `category`, `section_name`, `state` are accepted by the handler but never threaded into the repo query — admin list has them, public does not. Fix pending.
- **Admin RBAC bug: write-gated routes return 403 for `admin`.** `rbac/rbac_policy.csv` declares `admin,infographic,read` but no `admin,infographic,write` row. Routes guarded with `infographic:write` (POST `/api/v1/infographic`, PUT, DELETE, `/preview/token`, category CRUD) reject the built-in admin user with 403. Either add the row or relax the guard.
- **`state` filter on public list silently ignored** — same root cause as the filters bug above.

## Related

- `feature-docs/infographic/concept.md`
- `feature-docs/infographic/database.md`
- `interface/http/handler/infographic/infographic.go` — handler
- `usecase/infographic/service.go` — `GenerateMetabaseToken` / `GenerateMetabaseTokenWithExpiry`
