# Infographic Feature

Metabase dashboard embedding with JWT-signed tokens.

**Module:** `webdesa/api`
**Last updated:** 2026-07-05 (FROZEN)
**Status:** FROZEN — no further changes accepted without review

## Files

| File | Purpose |
|---|---|
| `interface/http/handler/infographic/infographic.go` | `InfographicHandler` struct + methods (package `infographic`) |
| `interface/http/handler/infographic/infographiccategory.go` | `InfographicCategoryHandler` |
| `interface/http/handler/infographic/route.go` | `RegisterRoutes` for `/infographic/*` + `/public/infographic/*` |
| `domain/infographic/infographic.go` | `Infographic` entity + `ComponentType` |
| `domain/infographic/infographic.go` | `CategoryName *string` (joined from `infographic_categories`) |
| `domain/infographiccategory/infographiccategory.go` | `Category` entity |
| `usecase/infographic/service.go` | CRUD + `GenerateMetabaseToken` |
| `usecase/infographiccategory/service.go` | Category CRUD |
| `interface/postgres/infographic.go` | sqlx implementation (LEFT JOIN on `infographic_categories`) |
| `interface/postgres/infographiccategory.go` | sqlx implementation |
| `db/migrations/00021`, `00022`, `00023`, `00035`, `00036` | infographic table + FK conversion |

## Database

### `infographic` (migrations 00021, 00022, `00023_change_component_id_to_integer`, 00036)

| Column | Type | Notes |
|---|---|---|
| `id` | UUID PK | `gen_random_uuid()` |
| `component_id` | BIGINT NOT NULL DEFAULT 0 | The Metabase component ID (integer) |
| `component_type` | VARCHAR(50) NOT NULL DEFAULT `'dashboard'` | `dashboard` or `question` |
| `section_name` | VARCHAR(100) NOT NULL | |
| `section_endpoint` | VARCHAR(255) NOT NULL | URL path (must start with `/`) |
| `category` | UUID | FK → `infographic_categories(id)` ON DELETE RESTRICT — **nullable** |
| `state` | BOOLEAN DEFAULT `true` | |
| `created_at` | TIMESTAMP | |
| `updated_at` | TIMESTAMP | |

Index: `idx_infographic_component_type`.

### `infographic_categories` (migration 00035)

Standard category table. **Seeded**: `"Lainnya"`.

## Domain Entity

`domain/infographic/infographic.go`:

```go
type ComponentType string

const (
    ComponentTypeQuestion   ComponentType = "question"
    ComponentTypeDashboard ComponentType = "dashboard"
)

type Infographic struct {
    ID              string
    ComponentID     int64
    ComponentType   ComponentType
    SectionName     string
    SectionEndpoint string
    Category        *string  // FK UUID
    CategoryName    *string  // joined from category table (omitempty in response)
    State           bool
    CreatedAt       time.Time
    UpdatedAt       time.Time
}
```

### Validation Rules

| Field | Rule |
|---|---|
| `ComponentID` | Must be > 0 |
| `ComponentType` | Must be `question` or `dashboard` |
| `SectionName` | ≤ 100 chars |
| `SectionEndpoint` | ≤ 255 chars, must start with `/` |

## Response Shape

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
  "token": "<jwt>"   // ← only on detail responses (admin + public)
}
```

## Metabase JWT Generation

`usecase/infographic/service.go:199-242` — `GenerateMetabaseToken(componentID int64, componentType string)` manually builds HS256 JWT (no external SDK).

**Header**:
```json
{"alg":"HS256","typ":"JWT"}
```

**Payload** (matches [Metabase JWT signing protocol](https://www.metabase.com/docs/latest/people-and-groups/authenticating-with-jwt)):
```json
{
  "resource": {"dashboard": 123} or {"question": 123},
  "params": {},
  "exp": <now+10m unix timestamp>
}
```

**Signature**: HMAC-SHA256 over `header_b64.payload_b64` using `metabaseConfig.SecretKey`.

**Encoding**: `base64.RawURLEncoding` (no padding, JWT-spec compliant).

**Returns**: `<header_b64>.<payload_b64>.<signature_b64>`.

## Service

`usecase/infographic/service.go` — `Service{repo, categoryRepo, clock, metabaseConfig config.MetabaseConfig}`

| Method | Purpose |
|---|---|
| `Create(CreateInfographicInput)` | Wraps string in `ComponentType(componentType)` |
| `Update`, `Delete` | Standard |
| `List(ListInfographicsInput)` | `section_name`, `state`, `q`, `category`, page, limit |
| `GetByID` | Generates token on detail response |
| `GetSectionNames` | `SELECT DISTINCT section_name ...` |
| `GenerateMetabaseToken(componentID, componentType)` | The JWT signing function (HS256) |

## Endpoints

### Public (no auth)

| Method | Path | Notes |
|---|---|---|
| GET | `/api/v1/public/infographic/list` | Paginated, no `token` field. Query: `q`, `category`, `section_name`, `state` |
| GET | `/api/v1/public/infographic/{id}` | Generates token |
| GET | `/api/v1/public/infographic/categories` | |

### Admin (RBAC: `infographic:read`)

| Method | Path | Notes |
|---|---|---|
| GET | `/api/v1/infographic` | Query: `section_name`, `state`, `q`, `category`, `page`, `limit` |
| GET | `/api/v1/infographic/{id}` | Generates token, includes in response |
| GET | `/api/v1/infographic/sections/names` | Distinct section names |

### Admin (RBAC: `infographic:write`)

| Method | Path | Notes |
|---|---|---|
| POST | `/api/v1/infographic` | Body `{component_id, component_type, section_name, section_endpoint, category?, state}` |
| POST | `/api/v1/infographic/preview/token` | Body `{component_id, component_type}` — generates JWT without creating record |
| PUT | `/api/v1/infographic/{id}` | |
| DELETE | `/api/v1/infographic/{id}` | |
| POST | `/api/v1/infographic/categories` | Body `{name}` |
| DELETE | `/api/v1/infographic/categories/{id}` | 409 if in use |

## List Filters (added 2026-07-05)

All public/admin list endpoints accept:
- `q=<text>` — case-insensitive substring search on `section_name` + `component_id` (cast to text)
- `category=<uuid>` — exact category UUID match

Both optional. Combine with existing `section_name`, `state` filters.

## Seed Data

`cmd/seed.go:286-292`:
```
"Statistik", "Keuangan", "Kependudukan", "Pendidikan", "Kesehatan"
```

## Configuration

`config/metabase.go`:
```go
type MetabaseConfig struct {
    SecretKey string `mapstructure:"secret_key"`
    URL       string `mapstructure:"url"`
}
```

`config/config.example.toml` example:
```toml
[metabase]
secret_key = "your-metabase-jwt-secret-key"
url = "http://metabase:3000"
```

**Critical:** The `secret_key` MUST exactly match the JWT signing secret configured in Metabase's admin settings (`Admin → Authentication → JWT → Signing Key`). If the keys don't match, Metabase rejects the token with `400 Message seems corrupt or manipulated`.

Set via env var if preferred: `DESA_METABASE_SECRET_KEY`, `DESA_METABASE_URL`.

## RBAC Permissions

- `infographic:read`
- `infographic:write`

## Tests

**Integration** (added 2026-07-05):
- `integration/infographic_test.go` (12+2 cases) — happy path, section names, list with filters, validation, generate preview token, **FilterByQuery**, **FilterByCategory**
- `integration/infographiccategory_test.go` (6 cases)

Full suite: 309 PASS / 0 FAIL / 0 PANICS (testcontainers DB).

## Known Quirks

- `category` column is **nullable** (preserved from migration 00036) — see `recom.docs/bugs.md#8`

## Freeze Notes (2026-07-05)

This feature is FROZEN. The following have been verified and locked:
- JWT payload matches Metabase spec (`resource` + `params` + `exp`)
- HS256 signing with `base64.RawURLEncoding` (unpadded)
- Header includes `alg: HS256`, `typ: JWT`
- `exp` set to `now + 10m`
- Response includes `category_id` (raw UUID) + `category` (object)
- List filters standardized: `q` + `category`
- Category lookups use LEFT JOIN to populate `CategoryName`

To un-freeze: must explicitly bump version in `feature-docs/infographic.md` and document reason in `recom.docs/`.
