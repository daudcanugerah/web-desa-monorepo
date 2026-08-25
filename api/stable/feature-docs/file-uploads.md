# File Uploads Feature

Filesystem-backed file storage for images, documents, and Quill editor uploads.

**Module:** `webdesa/api`
**Last updated:** 2026-07-04

## Files

| File | Purpose |
|---|---|
| `interface/file/local_handler.go` | `LocalHandler` — main file storage adapter (JPEG/PNG/WebP images, PDF/DOC/DOCX/XLS/XLSX documents) |
| `interface/http/handler/file/file.go` | `FileHandler` — serves `/files/{filename}` (package `file`) |
| `interface/http/handler/file/route.go` | `RegisterRoutes` for `/files/{filename}` |
| `interface/http/handler/berita/berita_upload.go` | `BeritaUploadHandler` — Quill media upload |
| `pkg/handlerutil/handlerutil.go` | `InferContentType`, `ValidateStruct` |
| `config/fileupload.go` | `FileUploadConfig` — sizes, directories |

## Local File Handler

`interface/file/local_handler.go` — implements multiple per-feature `FileHandler` interfaces (banner, berita, ppid, struktur, user).

### Constructors

```go
NewLocalHandler(uploadDir string)                          // single dir
NewLocalHandlerWithSubdir(uploadDir, subdir string)         // nested subdir
```

Both call `MkdirAll(uploadDir, 0755)` on construction.

### Methods

| Method | Validation | Size Limit | Output |
|---|---|---|---|
| `SaveImage(ctx, filename, content, size, contentType)` | MIME: `image/jpeg\|jpg\|png\|webp` | ≤ 10 MB | Returns filename only (e.g., `<uuid>.<ext>`) |
| `SaveDocument(ctx, filename, content, size, contentType)` | MIME: PDF, DOC, DOCX, XLS, XLSX | ≤ 50 MB | Returns filename |
| `Delete(ctx, path) error` | — | — | Missing file is OK |
| `GetFilePath(relativePath) string` | — | — | Full path on disk |

Filename generation: `<uuid>.<ext>` where ext is inferred from content type.

### Save Flow (`saveFile` helper)

1. Validate content type and size
2. Generate filename with UUID
3. Open file for writing
4. Stream content via `io.Copy`
5. Close file
6. Delete file on any error

## Per-Feature Usage

| Feature | FileHandler | Subdirectory |
|---|---|---|
| Banner | `SaveImage` | `./uploads/` |
| Berita | `SaveImage` | `./uploads/` |
| UMKM | `SaveImage` | `./uploads/` |
| Fasilitas | `SaveImage` | `./uploads/` |
| Struktur | `SaveImage` | `./uploads/` |
| User | `SaveImage` (profile) | `./uploads/` |
| PPID | `SaveImage` (thumbnail) + `SaveDocument` | `./uploads/ppid/` |

### PPID Subdirectory

Created via `NewLocalHandlerWithSubdir("./uploads", "ppid")` in `cmd/serve.go`.

**Fixed** (`recom.docs/bugs.md#2`): `GetPPIDUploadDirectory()` now correctly returns `UploadPPIDDirectory` (was returning the public directory before).

## Endpoints

### Public

| Method | Path | Notes |
|---|---|---|
| GET | `/api/v1/files/{filename}` | Directory-traversal protected |

### Static file server

| Method | Path | Auth |
|---|---|---|
| GET | `/uploads/*` | **No auth** — serves everything under `./uploads/` directly |

> The `/uploads/*` static route exists alongside `/api/v1/files/{filename}`. The latter has directory-traversal protection; the former does not. See `recom.docs/security.md#2`.

## Public File Access

`GET /api/v1/files/:filename` (`interface/http/handler/file/file.go`):

### Security

- `isValidFilename` rejects `..`, `/`, `\\` and matches only `[a-zA-Z0-9._-]+` (line 100-114)
- `filepath.Abs` + `strings.HasPrefix(absPath, absUploadsDir)` check prevents directory traversal (line 60-64)
- `http.ServeFile` for delivery

### MIME Type Inference

`getContentType` (line 117-141) maps file extensions:
- `.jpg` → `image/jpeg`
- `.png` → `image/png`
- `.pdf` → `application/pdf`
- `.doc` → `application/msword`
- etc.

## Quill Temporary Uploads

`POST /api/v1/berita/upload-media` (`interface/http/handler/berita/berita_upload.go`):

- Multipart form, max 50 MB
- Field `type=image|video` (default `image`)
- Image max 10 MB; Video max 50 MB
- Allowed MIME: `image/jpeg|png|gif|webp`; `video/mp4|webm|ogg|quicktime`
- Filename: `tmp-yyyymmdd-<uuid>.<ext>`
- Stored directly in `uploadsDir` (no subdir)
- Returned URL is the bare filename: `"tmp-20260317-abc.jpg"`

The `tmp-` prefix signals `processDeltaImages` (in `usecase/berita/service.go:46-126`) to move the file on save.

**Files persist indefinitely** until next save or manual cleanup.

## Configuration

`config/fileupload.go`:

```go
type FileUploadConfig struct {
    MaxImageSizeMB         int    `mapstructure:"max_image_size_mb"`        // default 10
    MaxDocumentSizeMB      int    `mapstructure:"max_document_size_mb"`     // default 50
    UploadPublicDirectory  string `mapstructure:"upload_public_directory"`  // default "./uploads/public/"
    UploadPPIDDirectory    string `mapstructure:"upload_ppid_directory"`    // default "./uploads/ppid/"
    UploadPrivateDirectory string `mapstructure:"upload_private_directory"` // unused
}

func (f *FileUploadConfig) GetPublicUploadDirectory() string {
    if f.UploadPublicDirectory == "" {
        return "./uploads/public/"
    }
    return f.UploadPublicDirectory
}

func (f *FileUploadConfig) GetPPIDUploadDirectory() string {
    if f.UploadPPIDDirectory == "" {
        return filepath.Join(f.GetPublicUploadDirectory(), "..", "ppid")
    }
    return f.UploadPPIDDirectory
}
```

## Cleanup Behavior

- **Image cleanup on failure** — every Create/Update/Delete failure path calls `Delete` on saved files to prevent orphans
- **Multi-image partial failure** (UMKM, Fasilitas) — iterates prior saves and deletes on individual failure
- **Quill tmp files** — persist indefinitely; cleaned up only when berita is saved and the file is referenced

## Tests

**Integration**:
- `integration/file_test.go` (3 cases) — `TestGetFile`, `TestGetFileWithDifferentContentTypes`, `TestGetFileWithUUIDFilename`
- `integration/berita_upload_test.go` (11 cases) — image/video upload, MIME validation, tmp directory

## Known Issues

- `test_uploads/` accumulates thousands of UUID-named files in test runs
- Static `/uploads/*` has no auth — direct URL access works for any uploaded file
- No automatic cleanup of orphaned `tmp-` files

## Related

- `pkg/handlerutil.InferContentType` — MIME type inference (also in handlerutil)
- `config/fileupload.go` — directory configuration
- `db/migrations/*` — feature-specific file columns