# PPID — Database

**Module:** `webdesa/api`
**Last updated:** 2026-08-23

## Migrations

| # | Purpose |
|---|---|
| `00009` | Create `ppid` table |
| `00010` | Create `ppid_requests` table |
| `00017` | Add `publication_at` to `ppid` |
| `00024` | Add `description` to `ppid` |
| `00025` | Revoked-by columns on `ppid_requests` (`revoked_at`, `revoked_by`) |
| `00026` | Index on `ppid.title` |
| `00031` | Create `ppid_categories` + seed `"Tanpa Kategori"` |
| `00032` | Make `ppid.category` reference `ppid_categories(id)` (nullable) |
| `00043` | Add `document_media_id` + `thumbnail_media_id` (FK → `gallery_media(id)`) |
| `00045` | Drop `file_url` + `thumbnail_url` columns from `ppid` |

## Tables

### `ppid` (migrations 00009, 00017, 00024, 00026, 00032, 00043, 00045)

| Column | Type | Nullable | Constraints / FK |
|---|---|---|---|
| `id` | UUID | NO | PK, `gen_random_uuid()` |
| `title` | VARCHAR(255) | NO | indexed (00026) |
| `category` | UUID | YES | FK → `ppid_categories(id)` ON DELETE RESTRICT |
| `document_media_id` | UUID | NO (app-level) | FK → `gallery_media(id)` ON DELETE RESTRICT — added 00043 |
| `thumbnail_media_id` | UUID | YES | FK → `gallery_media(id)` ON DELETE RESTRICT — added 00043 |
| `description` | TEXT | YES | added 00024 |
| `publication_at` | TIMESTAMP | YES | added 00017 |
| `created_at` | TIMESTAMP | NO | |
| `updated_at` | TIMESTAMP | NO | |

> **`file_url` and `thumbnail_url` columns were dropped by 00045.** Service requires non-nil `DocumentMediaID` in `Create` + `Validate`.

### `ppid_requests` (migrations 00010, 00025)

| Column | Type | Nullable | Constraints / FK |
|---|---|---|---|
| `id` | UUID | NO | PK, `gen_random_uuid()` |
| `ppid_id` | UUID | NO | FK → `ppid(id)` ON DELETE CASCADE |
| `requester_name` | VARCHAR(255) | NO | |
| `requester_email` | VARCHAR(255) | NO | |
| `notes` | TEXT | YES | (domain field: `purpose`) |
| `status` | VARCHAR(20) | NO | DEFAULT `'pending'`; `pending\|approved\|revoked` |
| `approved_at` | TIMESTAMP | YES | |
| `approved_by` | UUID | YES | FK → `users(id)` ON DELETE SET NULL |
| `revoked_at` | TIMESTAMP | YES | added 00025 |
| `revoked_by` | UUID | YES | FK → `users(id)` ON DELETE SET NULL — added 00025 |
| `created_at` | TIMESTAMP | NO | |

### `ppid_categories` (migration 00031)

Standard category table (UUID PK + name + timestamps). Seeded with `"Tanpa Kategori"`.

## Indexes

- `idx_ppid_title` on `ppid(title)` — added 00026
- `idx_ppid_category` on `ppid(category)` (implicit via FK)
- `idx_ppid_document_media_id` on `ppid(document_media_id)` (implicit via FK)
- `idx_ppid_thumbnail_media_id` on `ppid(thumbnail_media_id)` (implicit via FK)

## Seeds

- Migration `00031` seeds `ppid_categories."Tanpa Kategori"`.
- `cmd/seed.go:326-332` seeds `ppid_categories`: `"Anggaran", "Peraturan", "Laporan", "Profil", "Keuangan"`.

## Known Issues (DB-specific)

- **`category` column is nullable** (preserved from 00032). Domain accepts it; downstream consumers must tolerate null. See `recom.docs/bugs.md#8`.
- **`thumbnail_media_id` nullable by design**, but the legacy `validateThumbnailURL` whitelist (`jpg/jpeg/png/gif/webp`) is dead code — column dropped in 00045.
- No DB-level CHECK constraint on `ppid_requests.status` — invalid values are caught only at domain validation.

## Related

- `feature-docs/ppid/concept.md`
- `feature-docs/ppid/endpoint.md`
