# File Uploads — Concept

Filesystem-backed file storage for gallery originals/thumbnails, plus the legacy public `/api/v1/files/{filename}` route used by any pre-gallery-migration feature code that still references a UUID filename.

**Module:** `webdesa/api`
**Last updated:** 2026-08-23

> **Note (2026-08-23, draft/002-upload-migration complete):** the legacy public-upload flow this doc used to describe has been replaced by the gallery media system (`./gallery.md`). Feature services now persist only `*_media_id` UUIDs; the actual file lives in a per-feature system folder under `uploads/private/gallery/...` and is streamed through the unified signed media route. The `uploads/public/` and `uploads/ppid/` directories still exist for back-compat reads but **no new writes happen there** for any feature. Migration `00045_drop_legacy_upload_columns.sql` removed `image_url`, `images`, `profile_image_url`, `file_url`, `thumbnail_url` from every feature table.

## Purpose

Provide the on-disk primitives the gallery feature depends on (`LocalHandler` for filename generation, safe-path resolution, byte streaming) plus the one surviving public route (`/api/v1/files/{filename}`) that still serves any historical UUID-named file. Quill rich-text image uploads also live here as a thin wrapper that delegates the actual persistence to `galleryusecase.FileStore`.

## Business Rules

- **`LocalHandler.SaveImage` / `SaveDocument` are dead code for per-feature flows** — every active feature routes uploads through `galleryusecase.FileStore` (`usecase/gallery/filestore.go`). `cmd/serve.go` constructs a `LocalHandler` only for gallery originals + thumbnails. `cmd/seed.go` constructs one but discards the result with `_ = fileHandler` "retained for legacy /uploads/* static route scaffolding". Kept for the gallery adapter but never invoked on the per-feature upload path.
- **Quill upload is image-only with `media_id` filename** — `POST /api/v1/berita/upload-media` accepts only images, capped at 50 MB total, with the response filename set to the gallery `media_id` UUID (no `tmp-` prefix). Historical video upload retired; clients should embed a YouTube link instead.
- **Filename generation** — `<uuid>.<ext>` where `ext` is taken from the source filename or inferred from content type. Extension map: jpg/jpeg/png/webp/pdf/doc/docx/xls/xlsx.
- **Public route directory-traversal protection** — `GET /api/v1/files/{filename}` rejects `..`, `/`, `\\` and matches only `^[a-zA-Z0-9._-]+$`, then performs `filepath.Abs` + `strings.HasPrefix(absPath, absUploadsDir)` to keep reads inside the public uploads directory.
- **Static `/uploads/*` has no auth** — direct URL access works for any uploaded file. Static FS does not validate path bounds (see Known Issues).
- **Cleanup on failure** — every Create/Update/Delete failure path calls `Delete` on saved files to prevent orphans. Multi-image partial failure (UMKM, Fasilitas) iterates prior saves and deletes on individual failure.
- **Directory isolation** — `ValidateDirectoryIsolation()` (called at the top of `runServer`) ensures private and ppid directories resolve outside the public directory tree.

## Architecture

### Files

| File | Purpose |
|---|---|
| `interface/file/local_handler.go` | `LocalHandler` — generic file storage adapter (SaveImage / SaveDocument / Save / Path) |
| `interface/http/handler/file/file.go` | `FileHandler` — serves `/api/v1/files/{filename}` (package `file`) |
| `interface/http/handler/file/route.go` | `RegisterRoutes` for `/files/{filename}` |
| `interface/http/handler/berita/berita_upload.go` | `BeritaUploadHandler` — Quill image upload via gallery `FileStore` |
| `pkg/handlerutil/handlerutil.go` | `InferContentType`, `ValidateStruct` |
| `config/fileupload.go` | `FileUploadConfig` — sizes, directories |

### Local Handler Service Deps

`interface/file/local_handler.go` — implements file storage using the local filesystem.

### Local Handler Methods

| Method | Validation | Size Limit | Output |
|---|---|---|---|
| `SaveImage(ctx, filename, content, size, contentType)` | MIME: `image/jpeg`, `image/jpg`, `image/png`, `image/webp` | ≤ 10 MB (`local_handler.go:62`) | Returns filename `<uuid>.<ext>` |
| `SaveDocument(ctx, filename, content, size, contentType)` | MIME: `application/pdf`, `application/msword`, `application/vnd.openxmlformats-officedocument.wordprocessingml.document`, `application/vnd.ms-excel`, `application/vnd.openxmlformats-officedocument.spreadsheetml.spreadsheetml.document` | ≤ 50 MB (`local_handler.go:97`) | Returns filename `<uuid>.<ext>` |
| `Save(ctx, name, content, size, contentType)` | `safeStorageName` (rejects `..`/`.`/non-basename) | declared size must match written bytes | Returns safe name |
| `Path(name string) string` | — | — | `GetFilePath(filepath.Base(name))` |
| `Delete(ctx, path) error` | No-op when file missing | — | Removes file |
| `GetFilePath(relativePath) string` | — | — | Full path on disk |

Constructors:

```go
NewLocalHandler(uploadDir string)                    // single dir; MkdirAll(uploadDir, 0755)
NewLocalHandlerWithSubdir(uploadDir, subdir string)  // nested; MkdirAll(<full>, 0755)
```

### Save Flow (`saveFile` helper)

1. Validate content type and size
2. Generate filename with UUID
3. Open file for writing
4. Stream content via `io.Copy`
5. Close file
6. `os.Remove(fullPath)` on any error

## Glossary

- **LocalHandler** — `interface/file/local_handler.go` adapter. After the gallery migration, only used by gallery originals + thumbnails; per-feature paths go through `galleryusecase.FileStore`.
- **FileStore** — narrow port in `usecase/gallery/filestore.go` exposed by the gallery service so other feature services can persist media without knowing about folders, thumbnails, or storage layout. Returns gallery media UUIDs.
- **system folder** — private `gallery_folders` row that owns media for a single feature (see [gallery concept](./gallery/concept.md)). The Quill upload path stores inside the `system/berita` folder.
- **public route** — `/api/v1/files/{filename}` and `/uploads/*` static. The first has directory-traversal protection; the second does not.
- **`tmp-` prefix** — historical filename marker from before the gallery migration. No longer produced; Quill uploads now return `media_id` directly.

## Related

- [Database](./database.md)
- [Endpoints](./endpoint.md)
- `pkg/handlerutil.InferContentType` — MIME type inference (used by Quill upload)
- `config/fileupload.go` — directory configuration
- `db/migrations/00045_drop_legacy_upload_columns.sql` — legacy URL column cleanup
- [gallery](./gallery/concept.md) — gallery feature (current upload path)
- `recom.docs/security.md#2` — `/uploads/*` static FS recommendation
