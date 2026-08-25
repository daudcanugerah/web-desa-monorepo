# Berita — Database

Schema for `berita` (articles) and `berita_categories`. Media lives in the gallery.

**Module:** `webdesa/api`
**Last updated:** 2026-08-23

## Migrations

| Migration | Purpose |
|---|---|
| `00007` | initial `berita` table (with `image_url`) |
| `00018` | schema adjustments |
| `00027` | creates `berita_categories`; seeds `"Lainnya"` |
| `00029` | converts `berita.category` to UUID FK → `berita_categories(id)` ON DELETE RESTRICT |
| `00045_drop_legacy_upload_columns.sql` | drops `berita.image_url` |

## Tables

### `berita` (migrations 00007, 00018, 00027, 00029, 00045)

| Column | Type | Null | Notes |
|---|---|---|---|
| `id` | UUID | NO | PK, `gen_random_uuid()` |
| `title` | VARCHAR(255) | NO | |
| `content` | TEXT | NO | Quill delta JSON |
| `image_media_id` | UUID | YES | FK → `media(id)` ON DELETE SET NULL, nullable — replaces `image_url` (dropped 00045) |
| `category` | UUID | NO | FK → `berita_categories(id)` ON DELETE RESTRICT |
| `created_at` | TIMESTAMP | NO | indexed DESC |
| `updated_at` | TIMESTAMP | NO | |

### `berita_categories` (migration 00027)

| Column | Type | Null | Notes |
|---|---|---|---|
| `id` | UUID | NO | PK |
| `name` | VARCHAR(100) | NO | UNIQUE index |
| `created_at` | TIMESTAMP | NO | |
| `updated_at` | TIMESTAMP | NO | |

**Seeded**: `"Lainnya"` (placeholder for legacy rows).

## Indexes

| Index | Target |
|---|---|
| `berita_categories_name_key` (UNIQUE) | `berita_categories(name)` |
| `idx_berita_created_at` (DESC) | `berita(created_at)` |

## Seeds

From migrations:
- `db/migrations/00027` — `"Lainnya"` (placeholder for legacy rows)

From `cmd/seed.go:245-249`:
- `"Berita Desa"`, `"Pengumuman"`, `"Kegiatan"`

## Known Issues

None at the DB layer beyond the legacy `image_url` column already removed by migration `00045`.
