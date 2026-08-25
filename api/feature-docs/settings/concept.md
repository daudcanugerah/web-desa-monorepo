# Settings — Concept

Generic K-V JSONB settings storage. Currently used **only** for storing the village profile under key `desa_profile`. There is no generic CRUD API.

**Module:** `webdesa/api`
**Last updated:** 2026-08-23

## Purpose

Provide a flexible JSONB-backed key-value store. The schema is generic enough to hold arbitrary settings, but the application only writes one key today: `desa_profile`. The feature is therefore functionally identical to a single-row village-profile table — there's no `domain/settings/`, no `usecase/settings/`, and no `handler/settings/` package, and access is funneled entirely through `DesaRepository`.

## Business Rules

- **Desa-only settings constraint** — only `desa_profile` is read/written, via `interface/postgres/desa.go`. There is no `/api/v1/settings` route, no `domain/settings/`, no `usecase/settings/`, and no `handler/settings/` package. Every other proposed use case has to land in the `desa_profile` JSONB blob or a separate migration.
- **Startup read for PPID email config** — at server startup (`cmd/serve.go:243-258`), the `desa_profile` row is read to populate PPID email config (`VillageName`, `SupportEmail` (`d.Email`), `WebsiteURL` (`d.Website`)). The auth email service does not consume this read — it uses hardcoded defaults (`"Village Administration"`, `support@village.go.id`, `App.DomainAddr`). PPID email picks up the village profile fields; auth email ignores them. Changes to `desa_profile` require a server restart to take effect on outbound emails.
- **No `kind`/`schema` validation** — the JSONB blob is accepted as-is by `Upsert`. Malformed shapes can be written and only surface when something tries to unmarshal them.
- **Missing row at startup is silently swallowed** — `cmd/serve.go:248` does `if err == nil && d != nil`, so the defaults stand when the row is missing. No warning log — operators see stale emails without realizing.
- **Dead `desa` table (migration 00005)** — never queried, never written to by the running app. Profile lives entirely in `settings.desa_profile`. Drop in a future cleanup.
- **Test isolation** — `CleanDatabase` truncates `settings` between tests but does not pre-populate a default village profile. Each test that uses the endpoint does its own `INSERT INTO settings ...`.

## Architecture

### Files

| File | Purpose |
|---|---|
| `interface/postgres/desa.go` | `DesaRepository` — reads/writes the village profile under key `desa_profile` |
| `usecase/desa/service.go` | Get + Upsert business logic |
| `interface/http/handler/desa/desa.go` | HTTP handlers (`GetDesa`, `UpdateDesa`, `GetDesaPublic`) |
| `domain/desa/desa.go` | `Desa` entity |
| `cmd/serve.go` | Reads `desa_profile` at startup for PPID email config |
| `db/migrations/00014_create_settings_table.sql` | Creates `settings` |

> Despite the schema being generic, there is no CRUD API for arbitrary keys. Only `desa_profile` is read/written, via `interface/postgres/desa.go`. There is no `domain/settings/`, no `usecase/settings/`, and no `handler/settings/` package.

### Service Deps

`usecase/desa/service.go` — depends on `DesaRepository` (interface in `interface/postgres/desa.go:17-77`). Constant `desaSettingKey = "desa_profile"` is declared at line 17.

### Repository Interface Methods

```go
// Get — SELECT key, value, updated_at FROM settings WHERE key = $1
//       then json.Unmarshal into desa.Desa
//       finally overwrites d.UpdatedAt from the row timestamp
//       returns error "village profile not found" when the row is missing

// Upsert — INSERT ... ON CONFLICT (key) DO UPDATE
//           SET value = EXCLUDED.value, updated_at = EXCLUDED.updated_at
```

### Startup Read Flow

At server startup (`cmd/serve.go:243-258`), the `desa_profile` row is read to populate:

- **PPID email config**: `VillageName`, `SupportEmail` (`d.Email`), `WebsiteURL` (`d.Website`)
- **Auth email config**: defaults (`"Village Administration"`, `support@village.go.id`, `App.DomainAddr`) — see Known Issues

The PPID email service picks up the village profile fields; the auth email service does not (defaults are used unconditionally; the desa profile is read only for PPID). The actual SMTP credentials come from `config.SMTP`.

**Changes to `desa_profile` require a server restart to take effect on outbound emails** — both PPID and Auth read their copy of the values at process start.

## Glossary

- **`desa_profile`** — the single key currently written into `settings`. Constant `desaSettingKey` in `interface/postgres/desa.go:17`.
- **JSONB blob** — `settings.value` is `jsonb` with no schema enforcement. Anything unmarshals, anything serializes.
- **dead table** — `desa` (migration 00005) is unused; profile lives in `settings.desa_profile`.
- **startup read** — the `desa_profile` row is read once at process start. Subsequent writes don't affect in-process email services.

## Related

- [Database](./database.md)
- [Endpoints](./endpoint.md)
- `db/migrations/00014_create_settings_table.sql` — schema
- `interface/postgres/desa.go` — current consumer
- `cmd/serve.go:241-258` — startup read for PPID email config
- `feature-docs/desa.md` — actual feature consumer
