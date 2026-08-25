# Struktur — Database

**Module:** `webdesa/api`
**Last updated:** 2026-08-23

## Migrations

| # | Purpose |
|---|---|
| `00011` | Create `struktur_organisasi` table |
| `00037` | Add `description` column to `struktur_organisasi` |
| `00043` | Add `profile_image_media_id` FK → `gallery_media(id)` |
| `00045` | Drop `profile_image_url` column |

## Tables

### `struktur_organisasi` (migrations 00011, 00037, 00043, 00045)

| Column | Type | Nullable | Constraints / FK |
|---|---|---|---|
| `id` | UUID | NO | PK, `gen_random_uuid()` |
| `name` | VARCHAR(255) | NO | indexed `idx_struktur_name` |
| `position` | VARCHAR(255) | YES | indexed `idx_struktur_position` |
| `email` | VARCHAR(255) | YES | |
| `phone` | VARCHAR(50) | YES | |
| `profile_image_media_id` | UUID | NO (app-level) | FK → `gallery_media(id)` ON DELETE SET NULL — added 00043 |
| `description` | TEXT | YES | added 00037 |
| `created_at` | TIMESTAMP | NO | |
| `updated_at` | TIMESTAMP | NO | |

> `profile_image_url` was dropped by 00045. Service `Validate()` requires non-nil `ProfileImageMediaID`.

## Indexes

- `idx_struktur_name` on `struktur_organisasi(name)`
- `idx_struktur_position` on `struktur_organisasi(position)`
- `idx_struktur_profile_image_media_id` (implicit via FK)

## Seeds

None. Admin creates rows via admin API.

## Known Issues (DB-specific)

- **Broken Update SQL** in `interface/postgres/struktur.go:136-167`:
  ```sql
  UPDATE struktur_organisasi
  SET name = $1, position = $2, email = $3, phone = $4 = $5, profile_image_media_id = $6, description = $7, updated_at = NOW()
  WHERE id = $8
  ```
  Two problems:
  1. **Syntax error**: `phone = $4 = $5` — PostgreSQL rejects the statement.
  2. **Placeholder count mismatch**: 7 placeholders bind 8 Exec args (`Name, Position, Email, Phone, ProfileImageMediaID, Description, ID`). Even fixing the typo, `*sql.DB.ExecContext` errors with `bind count mismatch`.
  3. The row is never updated end-to-end. Existing test (`integration/struktur_test.go` "Update non-existent") hit the `RowsAffected==0` branch without ever exercising the SQL path.
  4. **Fix**: `phone = $4, profile_image_media_id = $5, description = $6, updated_at = NOW() WHERE id = $7`, then bind 7 args.
- **No DB-level CHECK constraint** on `email` shape — domain validates only.
- **`description` column nullable, no length cap** — domain enforces non-empty if present.

## Related

- `feature-docs/struktur/concept.md`
- `feature-docs/struktur/endpoint.md`
