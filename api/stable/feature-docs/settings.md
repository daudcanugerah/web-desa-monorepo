# Settings Feature

Generic K-V JSONB settings storage. Currently used **only** for storing the village profile under key `desa_profile`.

**Module:** `webdesa/api`
**Last updated:** 2026-07-04

## Status: Desa-Only (no generic API)

Despite the schema being generic, there is no CRUD API for arbitrary keys. Only `desa_profile` is read/written, via `interface/postgres/desa.go`.

## Files

| File | Purpose |
|---|---|
| `interface/postgres/desa.go` | Uses `settings` table under key `desa_profile` |
| `usecase/desa/service.go` | Get + Upsert business logic |
| `interface/http/handler/desa/desa.go` | HTTP handlers (`GetDesa`, `UpdateDesa`, `GetDesaPublic`) |
| `domain/desa/desa.go` | `Desa` entity |

## Database

### `settings` (migration 00014)

| Column | Type | Notes |
|---|---|---|
| `key` | VARCHAR(255) PRIMARY KEY | e.g., `desa_profile` |
| `value` | JSONB NOT NULL DEFAULT `'{}'` | The actual data |
| `updated_at` | TIMESTAMP NOT NULL DEFAULT NOW() | |

> The `desa` table from migration 00005 is **unused**. Profile lives in `settings.desa_profile`.

## Repository Access

Only through `DesaRepository`:

```go
const desaSettingKey = "desa_profile"

// Get — SELECT key, value, updated_at FROM settings WHERE key = $1
//       then json.Unmarshal into desa.Desa
//       finally overwrites d.UpdatedAt from the row timestamp

// Upsert — INSERT ... ON CONFLICT (key) DO UPDATE
//           SET value = EXCLUDED.value, updated_at = EXCLUDED.updated_at
```

## Startup Read

At server startup (`cmd/serve.go:175-204`), the `desa_profile` row is read to populate:
- **PPID email config**: `VillageName`, `SupportEmail`, `WebsiteURL`
- **Auth email config**: same fields for password reset emails

**Changes to Desa require a server restart to take effect on outbound emails.**

## Test Cleanup

`CleanDatabase` test teardown in `integration/setup_test.go` truncates `settings` between tests, but doesn't pre-populate a default village profile. Each test that uses the endpoint does its own `INSERT INTO settings ...`.

## Limitations

- No generic settings CRUD API
- No migration for non-desa keys
- No separate `usecase/settings/` package — access is via `DesaRepository` only

## Potential Extensions

If generic settings API is needed, add:

```go
// domain/settings/settings.go
type Setting struct {
    Key       string
    Value     json.RawMessage
    UpdatedAt time.Time
}

func (s *Setting) Validate() error {
    if s.Key == "" {
        return errors.New("key required")
    }
    if s.Value == nil {
        return errors.New("value required")
    }
    return nil
}

// usecase/settings/service.go
type Service interface {
    Get(ctx, key string) (*Setting, error)
    Upsert(ctx, key string, value interface{}) error
    Delete(ctx, key string) error
    List(ctx) ([]Setting, error)
}
```

Plus HTTP routes:
- `GET /api/v1/settings` (admin only, RBAC: `settings:read`)
- `PUT /api/v1/settings/{key}` (RBAC: `settings:write`)
- `DELETE /api/v1/settings/{key}` (RBAC: `settings:write`)
- `GET /api/v1/public/settings/{key}` (public read of explicitly whitelisted keys)

## Related

- `db/migrations/00014_create_settings_table.sql` — schema
- `interface/postgres/desa.go` — current consumer
- `cmd/serve.go:175-204` — startup read for PPID + Auth email config
- `feature-docs/desa.md` — actual feature consumer