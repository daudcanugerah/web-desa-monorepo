# Desa — Database

**Module:** `webdesa/api`
**Last updated:** 2026-08-23

## Migrations

| # | Purpose |
|---|---|
| `00005` | Original `desa` table (now dead — no code reads/writes it) |
| `00014` | Create `settings` table (JSONB blob storage) |

## Tables

### `settings` (migration 00014) — **the actual storage**

| Column | Type | Nullable | Constraints |
|---|---|---|---|
| `key` | VARCHAR(255) | NO | PK |
| `value` | JSONB | NO | DEFAULT `'{}'::jsonb` |
| `updated_at` | TIMESTAMP | NO | DEFAULT `NOW()` |

Constant `desaSettingKey = "desa_profile"` at `interface/postgres/desa.go:17`.

### `desa` (migration 00005) — **DEAD TABLE**

Originally created with the project. No code reads from or writes to `desa.*` — the village profile lives in `settings.value` under key `desa_profile` instead. Drop or repurpose.

## Indexes

- Primary key on `settings(key)` — sufficient for the singleton Get/Upsert path.

No additional indexes on `settings.value` (JSONB). Lookups are always by exact `key`.

## Seeds

None. First admin write seeds the row via `PUT /api/v1/desa`. Subsequent writes upsert.

## Known Issues (DB-specific)

- **`desa` table from migration 00005 is dead.** Profile lives in `settings` under key `desa_profile`. No code reads/writes `desa.*`. Drop or repurpose.
- **No schema validation on `settings.value` JSONB** — domain-level `Validate()` is the only check on field shape/length.
- **No `created_at` on `settings`** — only `updated_at`. Audit trail of initial creation is lost.
- **No FK or unique constraint** beyond PK — by design (singleton).

## Related

- `feature-docs/desa/concept.md`
- `feature-docs/desa/endpoint.md`
