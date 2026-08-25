# File Uploads — Endpoints

**Module:** `webdesa/api`
**Last updated:** 2026-08-23

## Routes

### Public file route

| Method | Path | Notes |
|---|---|---|
| GET | `/api/v1/files/{filename}` | Directory-traversal protected (`isValidFilename` + `filepath.Abs` + `HasPrefix`) |

### Static file server

| Method | Path | Auth |
|---|---|---|
| GET | `/uploads/*` | **No auth** — `http.FileServer(http.Dir(UploadPublicDirectory))` |

> The `/uploads/*` static route still exists alongside `/api/v1/files/{filename}`. The latter has directory-traversal protection (`file.go:60-78`); the former does not. See `recom.docs/security.md#2`.

### Quill image upload

| Method | Path | Notes |
|---|---|---|
| POST | `/api/v1/berita/upload-media` | Multipart `file` field; image-only; 50 MB cap; returns `media_id` as filename |

## Request/Response

### Public file access

`GET /api/v1/files/:filename` (`interface/http/handler/file/file.go`):

### Security

- `isValidFilename` rejects `..`, `/`, `\\` and matches only `^[a-zA-Z0-9._-]+$` (lines 114-128).
- `filepath.Abs` + `strings.HasPrefix(absPath, absUploadsDir)` check prevents directory traversal (lines 58-78).
- `http.ServeFile` for delivery.

### MIME Type Inference

`getContentType` (lines 131-148) maps file extensions:

| Extension | Content-Type |
|---|---|
| `.jpg` / `.jpeg` | `image/jpeg` |
| `.png` | `image/png` |
| `.gif` | `image/gif` |
| `.webp` | `image/webp` |
| `.pdf` | `application/pdf` |
| `.doc` | `application/msword` |
| `.docx` | `application/vnd.openxmlformats-officedocument.wordprocessingml.document` |
| `.xls` | `application/vnd.ms-excel` |
| `.xlsx` | `application/vnd.openxmlformats-officedocument.spreadsheetml.sheet` |
| `.txt` | `text/plain` |
| `.csv` | `text/csv` |
| `.zip` | `application/zip` |
| (other) | `application/octet-stream` |

### Quill Image Uploads

`POST /api/v1/berita/upload-media` (`interface/http/handler/berita/berita_upload.go`):

- Multipart form, max 50 MB cap (`ParseMultipartForm(50 << 20)`).
- Field name: `file` (single image — video upload removed).
- Image max 10 MB; enforced by `FileStore.SaveImage` against gallery config (`image_max_size_mb`).
- Allowed MIME: `image/jpeg`, `image/png`, `image/gif`, `image/webp` (gallery berita allowlist).
- Filename returned = `media_id` (no `tmp-` prefix). URL is the admin binary URL: `galleryuc.URLFor(URLScopeAdmin, "content", saved.MediaID)`.
- Backed by `galleryusecase.FileStore.SaveImage(ctx, FeatureBerita, ...)`. Stored in the `system/berita` system folder.

Response shape:

```json
{ "filename": "<media_id>", "url": "<admin binary url>" }
```

**No video support** in the Quill upload path — historical video upload was retired; clients should embed a YouTube link instead.

## Validation

### Handler-level

- `isValidFilename` regex `^[a-zA-Z0-9._-]+$` plus `filepath.Abs` traversal check.
- `ParseMultipartForm(50 << 20)` enforces the 50 MB total cap before any per-file check.
- `FileStore.SaveImage` enforces the per-image 10 MB cap and the berita image MIME allowlist.

### Domain-level

None — file-uploads has no `domain/` package. All validation lives in the handler.

## RBAC

No RBAC. Both `/api/v1/files/{filename}` (public) and `/uploads/*` (static) bypass authentication. The Quill `/api/v1/berita/upload-media` route applies the standard JWT auth middleware (inherited from the berita route group); specific permissions are not checked.

## Known Issues

- **`LocalHandler.SaveImage` / `SaveDocument` are DEAD code** for per-feature flows — every active feature routes uploads through `galleryusecase.FileStore`. The `LocalHandler` is still constructed by `cmd/serve.go` for gallery originals + thumbnails and by `cmd/seed.go` (discarded with `_ = fileHandler`). Cleanup recommended once the gallery adapter stabilizes, to prevent accidental re-introduction.
- **Quill upload image-only** — `POST /api/v1/berita/upload-media` accepts only images. Filename returned is the gallery `media_id` UUID (no `tmp-` prefix). Historical video upload was retired.
- Static `/uploads/*` has no auth — direct URL access works for any uploaded file. Static FS does not validate path bounds.
- `/api/v1/files/{filename}` only blocks traversal via regex + `Abs`/`HasPrefix`. A symlink inside `./uploads/public/` that resolves outside that directory would still serve — no symlink resolution check.
- `test_uploads/` accumulates thousands of UUID-named files in test runs because every integration test that hits `LocalHandler` creates new dirs.

## Related

- [Concept](./concept.md)
- [Database](./database.md)
- `interface/file/local_handler.go` — `LocalHandler`
- `interface/http/handler/file/file.go` — public route handler
- `interface/http/handler/berita/berita_upload.go` — Quill upload
- `config/fileupload.go` — directory configuration
