# Fasilitas — Database

Schema for `fasilitas` and `fasilitas_categories`. Media lives in the gallery; `images_media_ids` stores gallery UUIDs.

**Module:** `webdesa/api`
**Last updated:** 2026-08-23

## Migrations

| Migration | Purpose |
|---|---|
| `00006` | initial `fasilitas` table (with `type VARCHAR(100)`, `images` JSONB) |
| `00015` | schema adjustments |
| `00033` | creates `fasilitas_categories`; seeds `"Lainnya"` |
| `00034` | drops `type`; adds `category` UUID FK (nullable) → `fasilitas_categories(id)` ON DELETE RESTRICT; adds `idx_fasilitas_category` |
| `00045_drop_legacy_upload_columns.sql` | drops `fasilitas.images` JSONB |

## Tables

### `fasilitas` (migrations 00006, 00015, 00034, 00045)

| Column | Type | Null | Notes |
|---|---|---|---|
| `id` | UUID | NO | PK, `gen_random_uuid()` |
| `name` | VARCHAR(255) | NO | |
| `category` | UUID | YES | **nullable**; FK → `fasilitas_categories(id)` ON DELETE RESTRICT (was `type VARCHAR(100)` until 00034) |
| `latitude` | DECIMAL(10,8) | NO | range `[-90, 90]` |
| `longitude` | DECIMAL(11,8) | NO | range `[-180, 180]` |
| `description` | TEXT | YES | |
| `images_media_ids` | UUID[] | NO | DEFAULT `[]` — gallery media UUIDs — replaces `images` JSONB (dropped 00045) |
| `created_at` | TIMESTAMP | NO | |
| `updated_at` | TIMESTAMP | NO | |

Indexes: `idx_fasilitas_location` (latitude, longitude), `idx_fasilitas_category` (00034).

### `fasilitas_categories` (migration 00033)

Standard category table. **Seeded**: `"Lainnya"`.

| Column | Type | Null | Notes |
|---|---|---|---|
| `id` | UUID | NO | PK |
| `name` | VARCHAR(100) | NO | UNIQUE index |
| `created_at` | TIMESTAMP | NO | |
| `updated_at` | TIMESTAMP | NO | |

## Indexes

| Index | Target |
|---|---|
| `idx_fasilitas_location` | `fasilitas(latitude, longitude)` |
| `idx_fasilitas_category` | `fasilitas(category)` (added 00034) |
| `fasilitas_categories_name_key` (UNIQUE) | `fasilitas_categories(name)` |

## Seeds

From migrations:
- `db/migrations/00033` — `"Lainnya"` (placeholder for legacy rows)

From `cmd/seed.go`:
- `"Pendidikan"`, `"Kesehatan"`, `"Ibadah"`, `"Olahraga"`, `"Pemerintahan"`, `"Pasar"`

## Known Issues

None at the DB layer beyond the legacy `type` and `images` columns already removed by migrations `00034` and `00045`.
