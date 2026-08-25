# Infographic — Concept

Metabase dashboard/question embedding with JWT-signed tokens + access-log audit trail.

**Module:** `webdesa/api`
**Last updated:** 2026-08-23

## Purpose

Lets admins register Metabase dashboards/questions as named "sections" that citizens can view via a signed JWT token. Token issuance is logged (IP, user-agent, referer, token TTL) for audit. Rate-limited public access prevents token-farming abuse.

## Business Rules

- **Public vs admin TTL split**: `PublicTokenTTL = 5 minutes` for `/api/v1/public/infographic/{id}`; `AdminTokenTTL = 10 minutes` for `GET /api/v1/infographic/{id}` and `/preview/token`.
- **Rate limit**: 30 requests/hour/IP on `GetPublicAccess` (configurable via `rateLimiter`).
- **Access log on every public detail hit**: `LogAccess("public_detail", 5m)` writes one `infographic_access_log` row per request.
- **Access log on admin detail hit**: `LogAccess("admin_detail", 10m)` for audit traceability.
- **State flag**: `state` boolean gates visibility — `false` hides from public lists.
- **Category nullable**: `category` column preserved nullable (00036); see `recom.docs/bugs.md#8`.
- **Section endpoint must be a path**: `section_endpoint` must start with `/`, ≤ 255 chars.
- **Two `component_type`s**: `dashboard` or `question`. Anything else → `Validate()` rejects.
- **`token` field omitempty**: only present on **detail** responses (admin + public). List responses omit it.

## Architecture

### Files

| File | Purpose |
|---|---|
| `interface/http/handler/infographic/infographic.go` | `InfographicHandler` (package `infographic`) + `ListAccessLogs` |
| `interface/http/handler/infographic/infographiccategory.go` | `InfographicCategoryHandler` |
| `interface/http/handler/infographic/route.go` | `RegisterRoutes` for `/infographic/*` + `/public/infographic/*` |
| `domain/infographic/infographic.go` | `Infographic` entity + `ComponentType` + `Validate()` |
| `domain/infographic/access_log.go` | `AccessLog` entity |
| `domain/infographiccategory/infographiccategory.go` | `Category` entity |
| `usecase/infographic/service.go` | CRUD + `GenerateMetabaseToken` + `LogAccess` + `ListAccessLogs` |
| `usecase/infographic/access_log_repository.go` | `AccessLogRepository` interface |
| `usecase/infographiccategory/service.go` | Category CRUD |
| `interface/postgres/infographic.go` | sqlx implementation (LEFT JOIN on `infographic_categories`) |
| `interface/postgres/infographic_access_log.go` | sqlx implementation |
| `interface/postgres/infographiccategory.go` | sqlx implementation |
| `db/migrations/00021`, `00022`, `00023`, `00035`, `00036`, `00040` | table + FK + access log |

### Service Dependencies

`Service{repo, categoryRepo, accessLogRepo, clock, metabaseConfig config.MetabaseConfig, rateLimiter, publicTTL, adminTTL}`.

- `metabaseConfig.SecretKey` signs every Metabase JWT — must match Metabase's JWT Signing Key exactly.
- `rateLimiter` enforces 30/h/IP on public detail.
- `clock` injected for testable `created_at` / `token_issued_at` timestamps.

### Repository Interface Methods

`usecase/infographic/repository.go`:

- `Create(ctx, *Infographic) error`
- `Update(ctx, *Infographic) error`
- `Delete(ctx, id string) error`
- `GetByID(ctx, id string) (*Infographic, error)`
- `List(ctx, ListInfographicsInput) ([]Infographic, int64, error)`
- `GetSectionNames(ctx) ([]string, error)`

`AccessLogRepository`:

- `Create(ctx, *AccessLog) error`
- `List(ctx, ListAccessLogsInput) ([]AccessLog, int64, error)` — filters AND-combine on `infographic_id`, `ip`, page/limit

`CategoryLookup` (shared with `usecase/infographiccategory`): `GetByID(ctx, id) (*Category, error)`.

## Glossary

- **Metabase JWT** — HS256 token signed with `config.MetabaseConfig.SecretKey`. Header `{"alg":"HS256","typ":"JWT"}`. Payload `{resource: {dashboard|question: id}, params, exp}`. Signature `hmac.New(sha256.New, []byte(secret))` over `header_b64.payload_b64`. Encoded with `base64.RawURLEncoding` (no padding). See `pkg/metabase/jwt.go`.
- **Metabase Signing Key** — the secret configured under `Admin → Authentication → JWT` in Metabase's UI. MUST exactly match `metabase.secret_key`; mismatch → Metabase rejects with `400 Message seems corrupt or manipulated`.
- **Component type** — `dashboard` or `question`. Determines which Metabase resource type the JWT's `resource` field references.
- **Section endpoint** — public path segment citizens use to address this infographic. URL-friendly.
- **Access log event** — `"public_detail"` or `"admin_detail"`. Stored on `infographic_access_log` for audit.

## Related

- `feature-docs/ppid/concept.md` — sibling JWT signing (gallery scope, different claim set)
- `pkg/metabase/jwt.go` — Metabase JWT signing helper
- `config/metabase.go` — `MetabaseConfig`
- `config/config.example.toml` — `[metabase]` section
- `feature-docs/infographic/database.md`
- `feature-docs/infographic/endpoint.md`
