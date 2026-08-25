# Settings — Endpoints

**Module:** `webdesa/api`
**Last updated:** 2026-08-23

## Routes

> **There are no settings-domain HTTP routes.** There is no `domain/settings/`, no `usecase/settings/`, and no `handler/settings/` package. All settings access today goes through the `/desa` endpoints below — they are the only feature wired against `settings.desa_profile`.

### Desa endpoints (the only consumer of `settings`)

Wired by `interface/http/handler/desa/desa.go` (used by both authenticated admin scope and public scope):

| Method | Path | Notes |
|---|---|---|
| GET | `/api/v1/desa` | Public read of `desa_profile` |
| GET | `/api/v1/admin/desa` | Admin read (same payload) |
| PUT | `/api/v1/admin/desa` | Admin upsert |

`Get` returns `404 "village profile not found"` when the row has never been written. Admin auth is enforced via the standard JWT middleware on `/admin/*` routes.

## Request/Response

### `GET /api/v1/desa` and `GET /api/v1/admin/desa`

Returns the `Desa` entity JSON (admin and public payloads are identical):

```json
{
  "id": "string",
  "name": "string",
  "code": "string",
  "address": "string",
  "phone": "string",
  "email": "string",
  "website": "string",
  "vision": "string",
  "mission": "string",
  "history": "string",
  "logo_media_id": "uuid|null",
  "updated_at": "RFC3339"
}
```

When the row is missing: `404` with `{"error":"village profile not found"}`.

### `PUT /api/v1/admin/desa`

Body is the same JSON shape (minus `updated_at`). Upserts `settings.value` (JSONB) and rewrites `updated_at` via `INSERT ... ON CONFLICT (key) DO UPDATE`. Returns the new `Desa` JSON.

## Validation

### Handler-level

- Standard `pkg/handlerutil.ValidateStruct` on request body.
- UUID validation on any media ID fields.

### Domain-level

`domain/desa/desa.go` — `Desa` entity field-level rules (string length caps, email format if enforced, required fields). Not enumerated here — see [feature-docs/desa.md](../desa.md).

## RBAC

- `GET /api/v1/desa` — public, no auth.
- `GET /api/v1/admin/desa` — JWT auth required (standard admin middleware).
- `PUT /api/v1/admin/desa` — JWT auth required; no specific permission beyond authentication. Suggested permission: `desa:write` (not currently enforced).

## Known Issues

- **No `domain/settings/`, `usecase/settings/`, or `handler/settings/` packages** — there is no generic settings API. Adding one would require creating those packages, plus a `settings:read`/`settings:write` RBAC pair. See [concept Potential Extensions](./concept.md) for the suggested shape.
- **`desa` table (migration 00005) is dead code** — never queried, never written to by the running app. Profile lives entirely in `settings.desa_profile`. Drop the migration in a future cleanup.
- **Missing `desa_profile` row at startup is silently swallowed** — `cmd/serve.go:248` does `if err == nil && d != nil`; defaults stand. No warning log — operators see stale emails without realizing.
- **`kind`/`schema` not validated at the boundary** — malformed JSONB can be written and only surfaces when something tries to unmarshal it.
- **Auth email service uses hardcoded defaults** (`"Village Administration"`, `support@village.go.id`, `App.DomainAddr`) and never reads the village profile. PPID email service does read it. Inconsistency per `feature-docs/desa.md` drift. Fix: thread the loaded `Desa` through `NewAuthEmailService` like `NewPPIDEmailService`.
- **No HTTP routes for settings domain** — every read/write today happens through the `/desa` endpoints. A future generic settings API would need new routes (see [concept](./concept.md)).

## Related

- [Concept](./concept.md)
- [Database](./database.md)
- `interface/http/handler/desa/desa.go` — the only HTTP surface that touches `settings`
- `interface/postgres/desa.go` — repository
- `cmd/serve.go:241-258` — startup read for PPID email config
- `feature-docs/desa.md` — actual feature consumer
