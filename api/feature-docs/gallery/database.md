# Gallery — Database

**Module:** `webdesa/api`
**Last updated:** 2026-08-23

## Migrations

| # | File | Change |
|---|---|---|
| 00041 | `db/migrations/00041_create_gallery_folders_table.sql` | Creates `gallery_folders` |
| 00042 | `db/migrations/00042_create_gallery_media_table.sql` | Creates `gallery_media` and the `gallery_media_type` enum |
| 00043 | `db/migrations/00043_add_gallery_system_folders_and_media_refs.sql` | Adds `document` enum value; adds system folders seed mechanism; adds nullable media references to `users`, `struktur_organisasi`, `berita`, `banners`, `ppid` and JSONB `*_media_ids` arrays to `umkm` and `fasilitas` |
| 00044 | `db/migrations/00044_add_gallery_folder_cover_manual.sql` | Adds `cover_manual` column (pins manual folder cover) |
| 00045 | `db/migrations/00045_drop_legacy_upload_columns.sql` | Drops legacy URL columns (`image_url`, `images`, `profile_image_url`, `file_url`, `thumbnail_url`) from the feature tables migrated in 00043 |

## Tables

### `gallery_folders` (migrations 00041, 00043, 00044)

| Column | Type | Nullable | Constraints / FK |
|---|---|---|---|
| `id` | UUID | NO | PK, `gen_random_uuid()` |
| `name` | VARCHAR(255) | NO | Case-insensitive unique index |
| `description` | TEXT | NO | Defaults to empty string; max 1000 chars in domain validation |
| `is_public` | BOOLEAN | NO | Defaults to `false` |
| `is_system` | BOOLEAN | NO | Defaults to `false`; marks feature-owned folders |
| `feature_slug` | VARCHAR(32) | YES | **Partial unique index** `WHERE NOT NULL` |
| `cover_media_id` | UUID | YES | FK to `gallery_media(id)`, `ON DELETE SET NULL` |
| `cover_manual` | BOOLEAN | NO | DEFAULT `FALSE` (migration 00044); pins manual cover |
| `created_by` | UUID | NO | FK to `users(id)`, `ON DELETE RESTRICT` |
| `created_at` | TIMESTAMP | NO | |
| `updated_at` | TIMESTAMP | NO | |

### `gallery_media` (migrations 00042, 00043)

| Column | Type | Nullable | Constraints / FK |
|---|---|---|---|
| `id` | UUID | NO | PK, `gen_random_uuid()` |
| `folder_id` | UUID | NO | FK to `gallery_folders(id)`, `ON DELETE CASCADE` |
| `media_type` | `gallery_media_type` | NO | Enum: `image`, `video`, `document` (00043 added `document`) |
| `file_url` | VARCHAR(500) | NO | Private storage key |
| `thumbnail_url` | VARCHAR(500) | YES | Private thumbnail key |
| `thumbnail_failed` | BOOLEAN | NO | Defaults to `false` |
| `original_filename` | VARCHAR(255) | NO | Original client filename metadata |
| `mime_type` | VARCHAR(100) | NO | |
| `file_size` | BIGINT | NO | Must be greater than zero |
| `width` | INTEGER | YES | Image/video metadata |
| `height` | INTEGER | YES | Image/video metadata |
| `duration_seconds` | DOUBLE PRECISION | YES | Primarily video metadata |
| `is_public` | BOOLEAN | NO | Defaults to `false` |
| `uploaded_by` | UUID | NO | FK to `users(id)`, `ON DELETE RESTRICT` |
| `created_at` | TIMESTAMP | NO | |
| `updated_at` | TIMESTAMP | NO | |

### Feature `cover_manual` column (added by 00044)

The `cover_manual` column on `gallery_folders` is the only new column in 00044. When `TRUE`, `RecomputeFolderCover` short-circuits and commits the transaction without changing `cover_media_id`. `SetFolderCover(ctx, folderID, NULL)` clears the pin and resets `cover_manual = FALSE`.

### Cross-table media references (added by 00043, legacy columns dropped by 00045)

Migration 00043 adds nullable media FK columns to `users`, `struktur_organisasi`, `berita`, `banners`, `ppid`, and JSONB `*_media_ids` arrays to `umkm` and `fasilitas`. Migration 00045 drops the legacy URL columns (`image_url`, `images`, `profile_image_url`, `file_url`, `thumbnail_url`) from the same tables. After 00045 every feature media reference is a UUID column or JSONB array of UUIDs pointing at `gallery_media(id)`.

## Indexes

| Table | Index | Notes |
|---|---|---|
| `gallery_folders` | Unique case-insensitive on `name` | Migration 00041 |
| `gallery_folders` | `created_at DESC` | Listing |
| `gallery_folders` | `(is_public, created_at DESC)` | Public listing |
| `gallery_folders` | **Partial unique** `uq_gallery_folders_feature_slug ON (feature_slug) WHERE feature_slug IS NOT NULL` | Migration 00043 — one folder per feature slug |
| `gallery_folders` | FK index on `cover_media_id` | `ON DELETE SET NULL` target |
| `gallery_folders` | FK index on `created_by` | `ON DELETE RESTRICT` target |
| `gallery_media` | FK index on `folder_id` | `ON DELETE CASCADE` target |
| `gallery_media` | FK index on `uploaded_by` | `ON DELETE RESTRICT` target |

## Seeds

System folders are seeded via `cmd/seed.go:136 EnsureSystemFolders`. The seven canonical slugs (defined in `usecase/gallery/system_folders.go:7-15`):

```
banner, berita, struktur, umkm, fasilitas, user, ppid
```

Every system folder row is private (`is_public=false`), feature-owned (`is_system=true`), and carries one of the seven `feature_slug` values. `EnsureSystemFolders` is idempotent — it resolves each slug through `GetSystemFolderByFeature` and only inserts when the partial unique index would otherwise be empty, so a re-run on an already-seeded DB is a no-op. A fresh `make db-reset` produces all seven system folders before any feature upload runs.

## Known Issues

- `domain/gallery/media.go:55` validation error string still says `"media_type must be one of: image, video"` and omits `document` even though `MediaType.Valid()` accepts it and `mediaTypeFromMIME` maps document MIMEs. Truthful enum value but misleading error message.
- `kind` is a free-form `VARCHAR(16)` with no DB-level check constraint; anything not `"db"` or `"full"` is accepted. (Backups table only — see [backup database](./backup/database.md).)
- Cover recompute SQL `JOIN gallery_folders f ON f.id = m.folder_id` (`interface/postgres/gallery_folder.go:270-330`) does a per-folder join on every cover recomputation; for very large folder lists the `RecomputeFolderCover` loop adds N queries to visibility updates. Acceptable today but worth measuring if folder counts grow.
- The JSONB `*_media_ids` arrays on `umkm` and `fasilitas` are queried by `FindOrphanMedia` with `NOT EXISTS (... WHERE m.id = ANY(f.images_media_ids))`. No GIN index on those arrays — orphan sweep scans the rows instead of using the index. Negligible until row counts climb into the millions.

## Related

- [Concept](./concept.md)
- [Endpoints](./endpoint.md)
- `db/migrations/00041_create_gallery_folders_table.sql`
- `db/migrations/00042_create_gallery_media_table.sql`
- `db/migrations/00043_add_gallery_system_folders_and_media_refs.sql`
- `db/migrations/00044_add_gallery_folder_cover_manual.sql`
- `db/migrations/00045_drop_legacy_upload_columns.sql`
- `cmd/seed.go:136` — `EnsureSystemFolders`
