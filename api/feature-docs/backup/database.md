# Backup — Database

**Module:** `webdesa/api`
**Last updated:** 2026-08-23

## Migrations

| # | File | Change |
|---|---|---|
| 00012 | `db/migrations/00012_create_backups_table.sql` | Creates `backups` (filename, size, created_at) |
| 00046 | `db/migrations/00046_add_backup_kind.sql` | Adds `kind` column (defaults to `'db'`) |

## Tables

### `backups` (migrations 00012, 00046)

| Column | Type | Nullable | Constraints / FK |
|---|---|---|---|
| `id` | UUID | NO | PK, `gen_random_uuid()` |
| `filename` | VARCHAR(255) | NO | `backup_YYYYMMDD_HHMMSS.sql` or `full_YYYYMMDD_HHMMSS.tar.gz` |
| `size` | BIGINT | NO | bytes |
| `kind` | VARCHAR(16) | NO | DEFAULT `'db'`. Migration 00046. `db` (pg_dump-only) or `full` (tar.gz with uploads). No DB-level check constraint. |
| `created_at` | TIMESTAMP | NO | DEFAULT `NOW()`, indexed DESC |

## Indexes

| Table | Index | Notes |
|---|---|---|
| `backups` | `created_at DESC` | Listing order |

## Seeds

None.

## Known Issues

- `kind` is a free-form `VARCHAR(16)` with no DB-level check constraint; anything not `"db"` or `"full"` is accepted. Convention only.
- `BackupService.GetBackupFile` returns a raw `*os.File` from `b.Filename`; the handler never closes it on early errors and there's no path validation (e.g., `b.Filename = "../etc/passwd"`). Low risk because the ID comes from a UUID and we filter on the filename in the metadata, but no defense in depth.
- `GetBackupFile` opens from `backupDir` joined with `b.Filename`. If `b.Filename` contains `..`, it would escape the directory; no defense-in-depth check.

## Related

- [Concept](./concept.md)
- [Endpoints](./endpoint.md)
- `db/migrations/00012_create_backups_table.sql`
- `db/migrations/00046_add_backup_kind.sql`
- `interface/postgres/backup.go` — metadata repository
