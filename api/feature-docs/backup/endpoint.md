# Backup — Endpoints

**Module:** `webdesa/api`
**Last updated:** 2026-08-23

## Routes

> **HTTP handlers are NOT mounted.** `cmd/serve.go:325` keeps `backupHandler := handler.NewBackupHandler(backupService)` commented out. The handler compiles and binds to the service, but no route registration calls it. The only working entry point today is the CLI (`./desa-api backup ...`). Decision pending — wire with RBAC + integration tests, or remove dead code (`recom.docs/bugs.md#4`).

| Method | Path | Handler method | Notes |
|---|---|---|---|
| POST | `/api/v1/backups` | `CreateBackup` | Creates db-only backup |
| GET | `/api/v1/backups` | `ListBackups` | Query `page` (default 1), `limit` (default 20, max 100) |
| GET | `/api/v1/backups/{id}/download` | `DownloadBackup` | Streams binary; **Content-Disposition bug** — see Known Issues |
| POST | `/api/v1/backups/{id}/restore` | `RestoreBackup` | Triggers `RestoreFromBackup` |
| POST | `/api/v1/backups/restore/upload` | `UploadAndRestore` | Multipart `file` field; saves to disk + restores |
| GET | `/api/v1/backups/{id}` | `GetBackup` | **Always returns `501 Not Implemented`** (handler stub) |

### CLI Usage (the only live entry point)

```bash
./desa-api backup create                  # db-only pg_dump custom-format file
./desa-api backup create --with-files    # tar.gz bundling pg_dump + uploads/ (Task 5.2)
./desa-api backup list
./desa-api backup restore <backup-id>    # auto-detects kind from filename suffix
```

CLI list (`cmd/backup.go:138`) requests up to 100 rows, overriding the service-layer 20-cap (`service.go:132-134`). The CLI ignores pagination params.

Wires the same `usecase/backup.NewService` directly to `interface/postgres.NewBackupRepository`. Parses DSN via custom `parseDSN` helper in `cmd/serve.go:415-501` (not `url.Parse`).

## Request/Response

> None of the HTTP endpoints are reachable today. The shapes below describe what the (unmounted) handler would return if wired.

### `POST /api/v1/backups` — CreateBackup

No body. Returns `201 Created` with the new `Backup` JSON:

```json
{
  "id": "uuid",
  "filename": "backup_YYYYMMDD_HHMMSS.sql",
  "size": 0,
  "kind": "db",
  "created_at": "RFC3339"
}
```

### `GET /api/v1/backups` — ListBackups

Query: `page` (default 1), `limit` (default 20, max 100). Returns list + pagination:

```json
{
  "backups": [ { "id": "...", "filename": "...", "size": 0, "kind": "db|full", "created_at": "RFC3339" } ],
  "pagination": { "page": 1, "limit": 20, "total": 0 }
}
```

### `GET /api/v1/backups/{id}/download` — DownloadBackup

Streams the binary file. Content-Type is `application/octet-stream` (or `application/gzip` for tar.gz). **Content-Disposition is hardcoded** `attachment; filename=backup.sql` — see Known Issues.

### `POST /api/v1/backups/{id}/restore` — RestoreBackup

No body. Returns `200 OK` with empty body on success.

### `POST /api/v1/backups/restore/upload` — UploadAndRestore

Multipart `file` field, `application/sql` or `application/octet-stream`. Persists metadata with `kind="db"` and runs plain `pg_restore`. Returns the new metadata JSON.

### `GET /api/v1/backups/{id}` — GetBackup

Stub. Always returns `501 Not Implemented`.

## Validation

### Handler-level

- Path parameter `id` must be a valid UUID (handler returns `400` on parse failure).
- `RestoreFromFile` filename is run through `safeStorageName` before being joined to `backupDir`.
- Status codes: `404` for missing metadata, `500` for `pg_dump` / `pg_restore` non-zero exit, `400` for invalid upload payload.

### Domain-level

`domain/backup/backup.go:16-22` — `Backup{ID, Filename, Size int64, Kind string, CreatedAt}`.

`Validate()`:

- `Filename` required (≤ 255 chars)
- `Size > 0`
- `CreatedAt` set
- `Kind` **not** validated by `Validate()` — free-form string, `db`/`full` by convention

## RBAC

**None wired.** Even if the routes were mounted, no RBAC middleware is applied — `BackupHandler` accepts requests from any caller. Wiring must include the middleware before exposing the routes.

## Known Issues

- **`DownloadBackup` Content-Disposition filename bug** — hardcodes `Content-Disposition: attachment; filename=backup.sql` even when streaming a `.tar.gz` file (`backup.go:177`). Clients will save the full archive as `backup.sql`. Fix: derive filename from `b.Filename` and set `attachment; filename="<actual>"`.
- **HTTP handler routes are not mounted** — see top of [Routes](#routes). Decision pending — wire with RBAC + integration tests, or remove dead code (`recom.docs/bugs.md#4`).
- **`GetBackup` always returns 501** — stub endpoint, never implemented. Either implement or drop.
- CLI list silently caps at 100 rows via direct repo call (`cmd/backup.go:138`), bypassing the service-layer 20-cap. A backup-heavy environment can hide older entries behind this hardcoded limit.
- No RBAC middleware applied to any handler method. Wiring must add it before exposure.
- `kind` is a free-form `VARCHAR(16)` with no DB-level check constraint; anything not `"db"` or `"full"` is accepted.

## Related

- [Concept](./concept.md)
- [Database](./database.md)
- `cmd/backup.go` — active CLI
- `usecase/backup/service.go` — service implementation
- `interface/http/handler/backup.go` — handler (not mounted)
- `recom.docs/bugs.md#4` — wiring decision pending
