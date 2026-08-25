# Banner — Database

Schema for `banners` and `banner_categories`, the active-banner-limit trigger, and seed categories. All banner data lives in PostgreSQL; media lives in the gallery (see `feature-docs/gallery`).

**Module:** `webdesa/api`
**Last updated:** 2026-08-23

## Migrations

| Migration | Purpose |
|---|---|
| `00004` | initial `banners` table (with `image_url`) |
| `00019` | adds `banners.link` |
| `00038_enforce_active_banner_limit.sql` | `BEFORE INSERT OR UPDATE OF status` trigger + `enforce_active_banner_limit()` function; raises `check_violation` if active count ≥ 20 |
| `00039_banner_categories.sql` | creates `banner_categories`; adds `banners.category` UUID FK (nullable, ON DELETE RESTRICT); seeds `"Lainnya"` |
| `00045_drop_legacy_upload_columns.sql` | drops `banners.image_url` (gallery now owns media) |

## Tables

### `banners`

| Column | Type | Null | Default | Notes |
|---|---|---|---|---|
| `id` | UUID | NO | `gen_random_uuid()` | PK |
| `title` | VARCHAR(255) | NO | | |
| `description` | TEXT | NO | `''` | |
| `image_media_id` | UUID | YES | | FK → `media(id)` ON DELETE SET NULL — added in upload-migration, replaces `image_url` |
| `link` | VARCHAR(500) | NO | `''` | added in 00019 |
| `category` | UUID | YES | | FK → `banner_categories(id)` ON DELETE RESTRICT, added 00039 |
| `status` | VARCHAR(20) | NO | `'inactive'` | CHECK `IN ('active','inactive')` |
| `metadata` | JSONB | NO | `'{}'` | |
| `created_at` | TIMESTAMP | NO | | |
| `updated_at` | TIMESTAMP | NO | | |

### `banner_categories`

| Column | Type | Null | Notes |
|---|---|---|---|
| `id` | UUID | NO | PK |
| `name` | VARCHAR(100) | NO | UNIQUE index |
| `created_at` | TIMESTAMP | NO | |
| `updated_at` | TIMESTAMP | NO | |

## Indexes

| Index | Target |
|---|---|
| `idx_banners_status` | `banners(status)` |
| `banner_categories_name_key` (UNIQUE) | `banner_categories(name)` |

### Active Banner Limit Trigger

`db/migrations/00038_enforce_active_banner_limit.sql` creates a `BEFORE INSERT OR UPDATE OF status` trigger on `banners` that runs `enforce_active_banner_limit()`. The function counts active banners and raises `check_violation` if count ≥ 20.

**Previously**: limit was enforced in `usecase/banner/service.go:UpdateStatus` via a SELECT + check + UPDATE pattern, with TOCTOU race window.

**Now**: atomic enforcement at the database level.

## Seeds

From migrations:
- `db/migrations/00039` — `"Lainnya"` (placeholder for legacy rows)

From `cmd/seed.go`:
- `"Promo"`, `"Pengumuman"`, `"Event"`

## Known Issues

None at the DB layer beyond the legacy `image_url` column already removed by migration `00045`.
