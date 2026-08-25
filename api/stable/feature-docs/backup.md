# Backup Feature

Database backup/restore via `pg_dump`/`pg_restore` shell execution. **CLI-only** — HTTP routes are intentionally not wired.

**Module:** `webdesa/api`
**Last updated:** 2026-07-04

## Status: CLI-Only (HTTP routes commented out)

The handler implementation exists (`interface/http/handler/backup.go`, 202 lines) but **is not used**. See `recom.docs/bugs.md#4` for the two options:
- **Option A**: Wire up HTTP routes + add RBAC + integration tests
- **Option B**: Remove dead code (delete handler, remove commented routes)

Decision pending.

## Files

| File | Purpose |
|---|---|
| `cmd/backup.go` | **Active** — Cobra CLI commands `create`, `list`, `restore` |
| `usecase/backup/service.go` | **Active** — wraps `pg_dump`/`pg_restore` via `os/exec` |
| `domain/backup/backup.go` | `Backup` entity + `Validate()` |
| `interface/postgres/backup.go` | Metadata repository (filename, size, created_at) |
| `interface/http/handler/backup.go` | **DEAD** — handler exists, never wired |
| `interface/http/router.go` | **DEAD** — route registrations commented out |

## Database

### `backups` (migration 00012)

| Column | Type | Notes |
|---|---|---|
| `id` | UUID PK | `gen_random_uuid()` |
| `filename` | VARCHAR(255) NOT NULL | e.g., `backup_20260101_120000.sql` |
| `size` | BIGINT NOT NULL | bytes |
| `created_at` | TIMESTAMP NOT NULL DEFAULT NOW() | indexed DESC |

## Domain Entity

`domain/backup/backup.go:16-21` — `Backup{ID, Filename, Size int64, CreatedAt}`

### Validation

- `Filename` required (≤ 255 chars)
- `Size > 0`
- `CreatedAt` set

## Service

`usecase/backup/service.go` — `Service{repo, backupDir, dbHost, dbPort, dbName, dbUser, dbPassword}`. Depends on `os/exec` for `pg_dump`/`pg_restore`. Requires these binaries on `$PATH`.

### Methods

| Method | Purpose |
|---|---|
| `CreateBackup(ctx) (*Backup, error)` | Runs `pg_dump`, persists metadata. Rolls back file on DB failure. |
| `ListBackups(offset, limit)` | Paginated, ORDER BY `created_at DESC`. Clamps limit to 20 if > 100 |
| `RestoreFromBackup(ctx, backupID)` | Loads metadata, verifies file exists, runs `pg_restore --clean --if-exists` |
| `RestoreFromFile(ctx, file, filename)` | Saves upload to `<dir>/restore_<ts>_<safeFilename>`, runs restore, persists metadata |
| `GetBackupFile(ctx, backupID) (io.ReadCloser, int64, error)` | Returns `*os.File` + size for download |

### CreateBackup Flow

1. `MkdirAll(backupDir, 0755)`
2. Generate filename `backup_YYYYMMDD_HHMMSS.sql`
3. Run `pg_dump -h host -p port -U user -d dbname -F c -f filepath` (custom compressed format)
4. Set `PGPASSWORD` via `cmd.Env = append(os.Environ(), "PGPASSWORD=...")`
5. `os.Stat` to get size
6. Build entity + `repo.Create`
7. Roll back file on persistence failure

### Restore Command

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

## CLI Usage

```bash
./desa-api backup create
./desa-api backup list
./desa-api backup restore <backup-id>
```

Wires the same `usecase/backup.NewService` directly to `interface/postgres.NewBackupRepository`. Parses DSN via custom `parseDSN` helper (not `url.Parse`).

## HTTP Endpoints (NOT WIRED — commented in router)

| Method | Path | Notes |
|---|---|---|
| POST | `/api/v1/backups` | `CreateBackup` |
| GET | `/api/v1/backups` | `ListBackups` (page, limit) |
| GET | `/api/v1/backups/{id}/download` | Streams binary with `Content-Type: application/octet-stream` |
| POST | `/api/v1/backups/{id}/restore` | `RestoreBackup` |
| POST | `/api/v1/backups/restore/upload` | Multipart `UploadAndRestore` |
| GET | `/api/v1/backups/{id}` | **Returns 501 Not Implemented** |

`BackupHandler` construction is commented out in `cmd/serve.go:196-204, 228`.

## Required Environment

- `pg_dump` binary on `$PATH`
- `pg_restore` binary on `$PATH`
- Write access to `backupDir`
- `PGPASSWORD` env var injection at runtime

## Configuration

`config/backup.go`:
```go
type BackupConfig struct {
    BackupDir string  // "./backups"
}
```

DB connection details come from `config/postgres.go` DSN (parsed by `parseDSN` in `cmd/backup.go`).

## Tests

**No HTTP integration tests**. CLI is manual via `./desa-api backup ...`.

## Related

- `cmd/backup.go` — active CLI
- `usecase/backup/service.go` — service implementation
- `recom.docs/bugs.md#4` — wiring decision pending