# Settings — Database

**Module:** `webdesa/api`
**Last updated:** 2026-08-23

## Migrations

| # | File | Change |
|---|---|---|
| 00014 | `db/migrations/00014_create_settings_table.sql` | Creates `settings` (key, value, updated_at) |
| 00005 | `db/migrations/00005_create_desa_table.sql` | Creates `desa` (unused — dead table) |

## Tables

### `settings` (migration 00014)

| Column | Type | Nullable | Constraints / FK |
|---|---|---|---|
| `key` | VARCHAR(255) | NO | **PRIMARY KEY**, e.g., `desa_profile` |
| `value` | JSONB | NO | DEFAULT `'{}'` |
| `updated_at` | TIMESTAMP | NO | DEFAULT `NOW()` |

> The `desa` table from migration 00005 is **unused** (dead table). Profile lives in `settings.desa_profile`.

## Indexes

Primary key on `key` is the only index. No secondary indexes; no unique constraint beyond the PK.

## Seeds

None — the application does not pre-populate `settings.desa_profile`. Operators create it via `PUT /api/v1/admin/desa` or directly via SQL.

## Known Issues

- `kind`/`schema` is not enforced on the JSONB blob — any shape is accepted by `Upsert`.
- The `desa` table from migration 00005 is dead code (never queried, never written to by the running app). Profile lives entirely in `settings.desa_profile`. Drop the migration in a future cleanup.
- No secondary indexes — single-row reads by PK are fast, but a future generic settings API scanning by key prefix would table-scan.

## Related

- [Concept](./concept.md)
- [Endpoints](./endpoint.md)
- `db/migrations/00014_create_settings_table.sql`
- `db/migrations/00005_create_desa_table.sql` — dead table
- `interface/postgres/desa.go` — current consumer
