# Desa — Concept

Village profile management, stored as JSON in the `settings` table.

**Module:** `webdesa/api`
**Last updated:** 2026-08-23

## Purpose

Single source of truth for the village identity (name, contact info, vision/mission, etc.). Stored as JSONB under a known key in the `settings` table so the schema can evolve without migrations. Read at server startup to populate outbound email sender identity (PPID path only — auth email hard-codes defaults).

## Business Rules

- **Singleton via settings key**: profile lives under `settings.key = 'desa_profile'`. Repository `Get` returns wrapper error `"village profile not found"` if row missing (→ 404).
- **Upsert semantics**: `INSERT ... ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = EXCLUDED.updated_at` — single-row write, no separate Create endpoint.
- **Validation** (`domain/desa/desa.go`):
  - `Name` required, ≤ 255 chars
  - `Description` non-empty if present
  - `Address` non-empty if present
  - `Phone` ≤ 50 chars
  - `Email` must contain `@` and `.`
  - `Website` must start with `http://` or `https://`
  - `VisionMission` non-empty if present
- **Handler `Email` field has `validate:"omitempty,email"` tag** — extra structural check beyond the domain `@+.` rule.
- **Snapshot at startup**: `cmd/serve.go` reads the row once at boot and caches values into email senders. Changes to Desa require a **server restart** to take effect on outbound emails — no live-reload.

## Architecture

### Files

| File | Purpose |
|---|---|
| `interface/http/handler/desa/desa.go` | `DesaHandler` struct + methods (package `desa`) |
| `interface/http/handler/desa/route.go` | `RegisterRoutes` for `/desa` + `/public/desa` |
| `domain/desa/desa.go` | `Desa` entity + `Validate()` |
| `usecase/desa/service.go` | `Service` — Get + Upsert business logic |
| `interface/postgres/desa.go` | Repository using `settings` table |
| `cmd/serve.go:175-204` | Loads village profile for PPID email sender |
| `cmd/serve.go:158-162` | Auth email sender — **hard-coded defaults** (see Side Effects) |

### Service Dependencies

`Service{repo, clock}`. No external integrations beyond the `settings` repository.

### Repository Interface Methods

`interface/postgres/desa.go` — uses `settings` table, not a dedicated `desa` table:

- `Get(ctx) (*Desa, time.Time, error)` — returns `("village profile not found", …)` wrapper error when row missing
- `Upsert(ctx, *Desa) error` — `INSERT ... ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = EXCLUDED.updated_at`

Constant `desaSettingKey = "desa_profile"` at `interface/postgres/desa.go:17`.

## Glossary

- **Singleton via settings key** — the entire village profile lives under one JSONB blob in the generic `settings` table. The `desa` table from migration 00005 exists but is dead — no code reads/writes `desa.*`. See Known Issues.
- **Email sender snapshot** — `cmd/serve.go` reads `desa_profile` once at boot, copies `name` + `email` + `website` into the email-sender config, then runs cold for the lifetime of the process.
- **Auth email hard-coded defaults** — `cmd/serve.go:158-162` constructs the auth `EmailConfig` with literal `"Village Administration"` / `support@village.go.id`. PPID email path is the only one that actually consumes the profile.

## Related

- `feature-docs/ppid/concept.md` — primary consumer of the village profile (email sender)
- `cmd/serve.go:158-162` — hard-coded auth email defaults
- `cmd/serve.go:175-204` — startup loading of village profile for PPID email sender
- `interface/email/auth_email_service.go` — consumer (does NOT consume profile today)
- `interface/email/ppid_email_service.go` — consumer (consumes profile)
- `integration/desa_test.go` — integration tests
- `feature-docs/desa/database.md`
- `feature-docs/desa/endpoint.md`
