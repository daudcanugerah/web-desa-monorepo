# UMKM — Database

Schema for `umkm` and `umkm_categories`. Media lives in the gallery; `images_media_ids` stores gallery UUIDs.

**Module:** `webdesa/api`
**Last updated:** 2026-08-23

## Migrations

| Migration | Purpose |
|---|---|
| `00008` | initial `umkm` table (with `images` JSONB) |
| `00016` | adds `description` (NOT NULL DEFAULT '') |
| `00028` | creates `umkm_categories`; seeds `"Umum"` (placeholder for legacy rows) |
| `00030` | converts `umkm.category` to UUID FK → `umkm_categories(id)` ON DELETE RESTRICT |
| `00045_drop_legacy_upload_columns.sql` | drops `umkm.images` JSONB column |

## Tables

### `umkm` (migrations 00008, 00016, 00030, 00045)

| Column | Type | Null | Notes |
|---|---|---|---|
| `id` | UUID | NO | PK, `gen_random_uuid()` |
| `name` | VARCHAR(255) | NO | indexed `idx_umkm_name` |
| `owner` | VARCHAR(255) | YES | |
| `address` | TEXT | YES | |
| `phone` | VARCHAR(50) | YES | |
| `email` | VARCHAR(255) | YES | |
| `website` | VARCHAR(255) | YES | |
| `category` | UUID | NO | FK → `umkm_categories(id)` ON DELETE RESTRICT |
| `description` | TEXT | NO | DEFAULT `''`, added 00016 |
| `images_media_ids` | UUID[] | NO | DEFAULT `[]` — gallery media UUIDs — replaces `images` JSONB (dropped 00045) |
| `created_at` | TIMESTAMP | NO | |
| `updated_at` | TIMESTAMP | NO | |

### `umkm_categories` (migration 00028)

| Column | Type | Null | Notes |
|---|---|---|---|
| `id` | UUID | NO | PK |
| `name` | VARCHAR(100) | NO | UNIQUE index |
| `created_at` | TIMESTAMP | NO | |
| `updated_at` | TIMESTAMP | NO | |

**Seeded**: `"Umum"` (placeholder for legacy rows).

## Indexes

| Index | Target |
|---|---|
| `idx_umkm_name` | `umkm(name)` |
| `umkm_categories_name_key` (UNIQUE) | `umkm_categories(name)` |

## Seeds

From migrations:
- `db/migrations/00028` — `"Umum"` (placeholder for legacy rows)

From `cmd/seed.go`:
- `"Kuliner"`, `"Kerajinan"`, `"Pertanian"`, `"Jasa"`, `"Perdagangan"`

## Known Issues

None at the DB layer beyond the legacy `images` JSONB column already removed by migration `00045`.
