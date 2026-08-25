# Desa — Endpoints

**Module:** `webdesa/api`
**Last updated:** 2026-08-23

## Routes

### Public (no auth)

| Method | Path | Notes |
|---|---|---|
| GET | `/api/v1/public/desa` | Full profile |

### Admin (RBAC: `desa:read` / `desa:write`)

| Method | Path | Permission |
|---|---|---|
| GET | `/api/v1/desa` | `desa:read` |
| PUT | `/api/v1/desa` | `desa:write` |

No Create / Delete / List endpoints — singleton profile, single upsert endpoint.

## Request/Response

### Get (public + admin)

```
GET /api/v1/public/desa
GET /api/v1/desa
```

**Response**:

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

Missing row → `404 Not Found` (`"village profile not found"`).

### Update (admin)

```
PUT /api/v1/desa
{
  "name": "Desa Sukamaju",
  "description": "...",
  "address": "Jl. Raya No.1",
  "phone": "+62...",
  "email": "admin@desa.sukamaju.id",
  "website": "https://desa.sukamaju.id",
  "vision_mission": "Maju bersama..."
}
```

**Response**: `200 OK` + the full updated profile JSON.

## Validation

### Handler-level

- `Email` field has `validate:"omitempty,email"` tag — extra structural check beyond the domain `@+.` rule.
- All other fields validated against the rules below via `pkg/handlerutil.ValidateStruct`.

### Domain-level (`domain/desa/desa.go`)

| Field | Rule |
|---|---|
| `Name` | Required, ≤ 255 chars |
| `Description` | Non-empty if present |
| `Address` | Non-empty if present |
| `Phone` | ≤ 50 chars |
| `Email` | Must contain `@` and `.` |
| `Website` | Must start with `http://` or `https://` |
| `VisionMission` | Non-empty if present |

## RBAC

- `desa:read` — admin Get
- `desa:write` — admin Update

## Known Issues (Endpoint-specific)

- **Auth email sender ignores village profile.** `cmd/serve.go:158-162` constructs the auth `EmailConfig` with literal `"Village Administration"` / `support@village.go.id`. Password-reset emails always come from this generic identity, regardless of which village is configured in `desa_profile`. PPID email path is the only one that actually consumes the profile.
- **Handler `Email` `omitempty` tag.** The handler-side `Email` field uses `validate:"omitempty,email"` — when the field is empty/zero, the email validator is skipped entirely, allowing a PUT that clears the email without complaint. Domain-level `@+.` check fires only if a non-empty value is provided.
- **Changes require restart.** `cmd/serve.go` snapshots the `desa_profile` row once at boot into email senders. Live updates via `PUT /api/v1/desa` do not propagate to outbound email until the server restarts. No live-reload mechanism.
- **No partial-update / PATCH endpoint.** `PUT /api/v1/desa` replaces the full profile — clients must re-send all fields.

## Related

- `feature-docs/desa/concept.md`
- `feature-docs/desa/database.md`
- `interface/http/handler/desa/desa.go` — handler
- `usecase/desa/service.go` — service
