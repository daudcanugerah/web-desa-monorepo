# Backup — Concept

Database backup/restore via `pg_dump`/`pg_restore` shell execution, plus tar.gz full backups that bundle the uploads tree. The HTTP handler exists and is built against the service, but **the routes are not mounted in the router**; the only working entry point today is the CLI.

**Module:** `webdesa/api`
**Last updated:** 2026-08-23

## Purpose

Provide a one-shot disaster-recovery path for the entire deployment: the database schema and data, plus the on-disk uploads tree (gallery originals + thumbnails, plus any historical public/ppid files). Two backup kinds: `db` (pg_dump-only, smallest, fastest) and `full` (pg_dump + uploads tar.gz). Restore auto-detects kind from the filename suffix.

## Business Rules

- **HTTP handlers not mounted** — `cmd/serve.go:325` keeps `backupHandler := handler.NewBackupHandler(backupService)` commented out. The routes compiled in `interface/http/handler/backup.go` are reachable only through the CLI today. Decision pending — wire with RBAC + integration tests, or remove dead code (`recom.docs/bugs.md#4`).
- **Backup kind auto-detection** — `RestoreFromBackup` switches on the filename suffix: `.tar.gz` → full restore (extract `db.pgdump`, run `pg_restore --clean --if-exists`, walk every other tar entry to its original relative path so `uploads/public/`, `uploads/private/`, and `uploads/ppid/` land back where the running app expects them); anything else → plain `pg_restore` (db-only flow unchanged).
- **`IncludePaths` default `["./uploads/"]`** — `NewService` falls back to that slice when `IncludePaths` is empty. Full backups bundle `uploads/public/`, `uploads/private/` (gallery originals + thumbnails), and `uploads/ppid/`.
- **`DownloadBackup` Content-Disposition bug** — hardcodes `attachment; filename=backup.sql` even when streaming a `.tar.gz` file (`backup.go:177`). Clients will save the full archive as `backup.sql`. Fix: derive filename from `b.Filename` and set `attachment; filename="<actual>"`.
- **No RBAC wired** — even if the routes were mounted, no RBAC middleware is applied — `BackupHandler` accepts requests from any caller. Wiring must include the middleware before exposing the routes.
- **`kind` is free-form** — `VARCHAR(16)` with no DB-level check constraint. By convention `db` or `full`; anything else is accepted by the metadata row.
- **CLI list hardcoded 100-cap** — `cmd/backup.go:138` requests up to 100 rows via direct repo call, bypassing the service-layer 20-cap. Older backups can hide behind the hardcoded limit.
- **Rollback on persistence failure** — both `CreateBackup` and `CreateFullBackup` remove the produced file/archive when the metadata insert fails.

## Architecture

### Files

| File | Purpose |
|---|---|
| `cmd/backup.go` | **Active** — Cobra CLI commands `create`, `list`, `restore` |
| `cmd/serve.go:148 + 325` | Backup service construction + handler wiring — **commented out** |
| `usecase/backup/service.go` | `Service` — `pg_dump`/`pg_restore` via `os/exec`, tar.gz bundling, restore auto-detect |
| `domain/backup/backup.go` | `Backup` entity + `Validate()` (includes `Kind`) |
| `interface/postgres/backup.go` | Metadata repository (filename, size, kind, created_at) |
| `interface/http/handler/backup.go` | **Implemented, not mounted** — handler compiles + binds to service |
| `config/backup.go` | `BackupConfig` — Directory + IncludePaths |
| `db/migrations/00012_create_backups_table.sql` | Creates `backups` |
| `db/migrations/00046_add_backup_kind.sql` | Adds `kind` column |

### Service Deps

`usecase/backup/service.go` — `Service{repo, backupDir, includePaths, dbDSN, dbHost, dbPort, dbName, dbUser, dbPassword}`. Depends on `os/exec` for `pg_dump`/`pg_restore` and on `archive/tar` + `compress/gzip` for the full bundle. Requires these binaries on `$PATH`.

### Repository Interface Methods

From `interface/postgres/backup.go`:

- `Create(ctx, *Backup) error`
- `GetByID(ctx, id) (*Backup, error)`
- `List(ctx, offset, limit) ([]Backup, error)`

### Service Methods

| Method | Purpose |
|---|---|
| `CreateBackup(ctx) (*Backup, error)` | `pg_dump -F c` to `<dir>/backup_YYYYMMDD_HHMMSS.sql`; persists `kind="db"`. Rolls back file on DB failure. |
| `CreateFullBackup(ctx) (*Backup, error)` | `pg_dump -F c` into a temp file, then `gzip+tar` wraps `db.pgdump` + every path in `includePaths` (default `./uploads/`) into `<dir>/full_YYYYMMDD_HHMMSS.tar.gz`; persists `kind="full"`. (Task 5.2) |
| `ListBackups(offset, limit)` | Paginated, ORDER BY `created_at DESC`. Clamps `limit` to 20 if `<=0` or `>100`. |
| `RestoreFromBackup(ctx, backupID)` | Loads metadata, verifies file exists. Auto-detects kind from filename suffix `.tar.gz` → full restore, else plain `pg_restore --clean --if-exists`. |
| `RestoreFromFile(ctx, file, filename)` | Saves upload to `<dir>/restore_<ts>_<safeFilename>`, runs plain `pg_restore`, persists metadata with `kind="db"`. (Used only by the unmounted HTTP handler.) |
| `GetBackupFile(ctx, backupID) (io.ReadCloser, int64, error)` | Returns `*os.File` + size for download. |

### `CreateBackup` Flow

1. `MkdirAll(backupDir, 0755)`
2. Generate filename `backup_YYYYMMDD_HHMMSS.sql`
3. Run `pg_dump -h host -p port -U user -d dbname -F c -f filepath` (custom compressed format)
4. Set `PGPASSWORD` via `cmd.Env = append(os.Environ(), "PGPASSWORD=...")`
5. `os.Stat` to get size
6. Build entity (`kind="db"`) + `repo.Create`
7. Roll back file on persistence failure

### `CreateFullBackup` Flow (Task 5.2)

1. `pg_dump -F c` into a temp file (`os.CreateTemp("", "desa-full-dump-*.pgdump")`)
2. Open `<dir>/full_YYYYMMDD_HHMMSS.tar.gz` and wrap in `gzip.Writer → tar.Writer`
3. Write the temp dump as the `db.pgdump` tar entry
4. Walk each `includePaths` entry and add every regular file at its repo-relative path (skipping missing roots so a fresh install still works)
5. Close writers, `os.Stat` for size
6. Build entity (`kind="full"`) + `repo.Create`
7. Roll back archive on persistence failure

### Restore Auto-Detection

`RestoreFromBackup` switches behaviour on the filename suffix:

- `.tar.gz` → `restoreFull`: extract `db.pgdump` to a temp file → `pg_restore --clean --if-exists` → walk every other tar entry and re-emit it to its original relative path so `uploads/public/`, `uploads/private/` (gallery originals + thumbnails), and `uploads/ppid/` land back where the running app expects them.
- Anything else → plain `pg_restore` (db-only flow unchanged).

### `pg_restore` Invocation

```go
exec.CommandContext(ctx, "pg_restore",
    "-h", host,
    "-p", port,
    "-U", user,
    "-d", dbname,
    "--clean", "--if-exists",
    "-F", "c",
    filepath,
).Run()
```

`PGPASSWORD` injected via `cmd.Env` like `pg_dump`.

### Required Environment

- `pg_dump` binary on `$PATH`
- `pg_restore` binary on `$PATH`
- Write access to `BackupDir`
- `PGPASSWORD` env var injection at runtime (handled by service)

## Glossary

- **kind** — `backups.kind` column (`db` or `full`). Drives the restore code path; nothing else in the system cares.
- **`IncludePaths`** — `config.BackupConfig.IncludePaths` slice, default `["./uploads/"]`. Tells the full backup tar walker which on-disk roots to bundle.
- **`db.pgdump`** — tar entry name inside a full backup archive. Always the first entry written by `CreateFullBackup`; the only entry read by `restoreFull` for the database step.
- **`RestoreFromFile`** — restore-from-upload path used only by the unmounted HTTP `UploadAndRestore` handler. Persists metadata with `kind="db"`.

## Related

- [Database](./database.md)
- [Endpoints](./endpoint.md)
- `cmd/backup.go` — active CLI
- `usecase/backup/service.go` — service implementation
- `interface/http/handler/backup.go` — handler (not mounted)
- `recom.docs/bugs.md#4` — wiring decision pending
