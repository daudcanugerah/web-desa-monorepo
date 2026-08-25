# File Uploads — Database

**Module:** `webdesa/api`
**Last updated:** 2026-08-23

## Migrations

| # | File | Change |
|---|---|---|
| 00045 | `db/migrations/00045_drop_legacy_upload_columns.sql` | Drops legacy URL columns (`image_url`, `images`, `profile_image_url`, `file_url`, `thumbnail_url`) from the feature tables migrated in 00043 — `users`, `struktur_organisasi`, `berita`, `banners`, `ppid`, `umkm`, `fasilitas` |

> File uploads does not own any tables. The feature is purely a filesystem adapter plus two HTTP routes. Migration 00045 belongs to the gallery feature's table-evolution arc but is included here because the legacy URL columns it removed are the only schema artifacts this feature previously touched.

## Tables

None.

## Indexes

None.

## Seeds

None.

## Known Issues

- The `uploads/ppid/` directory is no longer written to by any feature. Historical files may still exist on disk; safe to delete after confirming no in-flight reads remain. (See `feature-docs/backup.md` for the full backup/restore path that does preserve it.)
- `test_uploads/` accumulates thousands of UUID-named files in test runs because every integration test that hits `LocalHandler` creates new dirs. No automated teardown.

## Related

- [Concept](./concept.md)
- [Endpoints](./endpoint.md)
- `db/migrations/00045_drop_legacy_upload_columns.sql`
- [gallery database](./gallery/database.md) — owns all media storage
