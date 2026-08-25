# Desa (Village Profile) Feature

Village profile management, stored as JSON in the `settings` table.

**Module:** `webdesa/api`
**Last updated:** 2026-07-04

## Files

| File | Purpose |
|---|---|
| `interface/http/handler/desa/desa.go` | `DesaHandler` struct + methods (package `desa`) |
| `interface/http/handler/desa/route.go` | `RegisterRoutes` for `/desa` + `/public/desa` (package `desa`) |
| `domain/desa/desa.go` | `Desa` entity + `Validate()` |
| `usecase/desa/service.go` | `Service` — Get + Upsert business logic |
| `interface/postgres/desa.go` | Repository using `settings` table |
| `interface/http/handler/auth/route.go` (init) | Reads `settings.desa_profile` at startup to populate PPID email config |
| `cmd/serve.go:175-204` | Loads village profile for email senders (Auth + PPID) |

## Database

`settings` table (`db/migrations/00014`):

| Column | Type | Notes |
|---|---|---|
| `key` | VARCHAR(255) PK | e.g., `desa_profile` |
| `value` | JSONB NOT NULL DEFAULT `'{}'` | The actual profile data |
| `updated_at` | TIMESTAMP NOT NULL DEFAULT NOW() | |

> The `desa` table from migration 00005 is **unused**. Profile lives in `settings` under key `desa_profile`.

## Domain Entity

`domain/desa/desa.go:15-24` — `Desa{Name, Description, Address, Phone, Email, Website, VisionMission}` (all optional fields are `*string`)

### Validation Rules

| Field | Rule |
|---|---|
| `Name` | Required, ≤ 255 chars |
| `Description` | Non-empty if present |
| `Address` | Non-empty if present |
| `Phone` | ≤ 50 chars |
| `Email` | Must contain `@` and `.` |
| `Website` | Must start with `http://` or `https://` |
| `VisionMission` | Non-empty if present |

## Service

`usecase/desa/service.go:13-21` — `Service{repo, clock}`

| Method | Purpose |
|---|---|
| `Get(ctx)` | Returns `"village profile not found"` wrapper error if missing |
| `Update(ctx, UpdateDesaInput)` | Validates, calls `repo.Upsert` |

## Endpoints

### Public (no auth)

| Method | Path |
|---|---|
| GET | `/api/v1/public/desa` | Full profile |

### Admin (RBAC: `desa:read` / `desa:write`)

| Method | Path | Permission |
|---|---|---|
| GET | `/api/v1/desa` | `desa:read` |
| PUT | `/api/v1/desa` | `desa:write` |

## Response Shape

```json
{
  "name": "Desa Sukamaju",
  "description": "...",
  "address": "Jl. Raya No.1",
  "phone": "+62...",
  "email": "admin@desa.sukamaju.id",
  "website": "https://desa.sukamaju.id",
  "vision_mission": "Maju bersama...",
  "updated_at": "2026-01-15T10:00:00+07:00"
}
```

## Side Effects

The `desa_profile` row is read at server startup (`cmd/serve.go`) to populate:
- **PPID email** — `VillageName`, `SupportEmail`, `WebsiteURL`
- **Auth email** — same fields used for password reset emails

**Changes to Desa require a server restart to take effect on outbound emails.**

## RBAC Permissions

- `desa:read`
- `desa:write`

## Tests

**Integration** (`integration/desa_test.go`): 5 cases
- `TestDesa_HappyPath`
- `TestDesa_UpdateWithoutAuth`
- `TestDesa_UpdateWithEmptyName`
- `TestDesa_UpdateWithInvalidEmail`
- `TestDesa_UpdateWithInvalidWebsite`

## Repository

`interface/postgres/desa.go` — backed by the `settings` table:
- Constant: `desaSettingKey = "desa_profile"`
- `Get` — `SELECT key, value, updated_at FROM settings WHERE key = $1`, then `json.Unmarshal` into `desa.Desa`
- `Upsert` — `INSERT ... ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = EXCLUDED.updated_at`

## Related

- `cmd/serve.go:175-204` — startup loading of village profile for email senders
- `interface/email/auth_email_service.go` — consumer of village profile
- `interface/email/ppid_email_service.go` — consumer of village profile
- `integration/desa_test.go` — integration tests