# Gallery — Design

> Technical design for the Gallery feature.
> Status: **IMPLEMENTED** — 2026-07-19.
> Last updated: 2026-07-19
> Companion docs: `requirements.md`, `tasks.md`

---

## 1. Overview

Gallery follows the existing Clean Architecture layout (`domain/`, `usecase/`, `interface/postgres/`, `interface/http/handler/gallery/`). It introduces two new entities (`Folder`, `Media`), two permissions (`gallery:read`, `gallery:write`), and a thumbnail pipeline. Folder and media visibility are independent; effective public visibility is `folder.is_public AND media.is_public`.

No external services are introduced — thumbnail generation uses stdlib `image` + `image/jpeg/png/webp` for images and `ffmpeg` (subprocess) for videos. ffmpeg is assumed present on the host (documented in README); absence degrades to `thumbnail_url = NULL` + `thumbnail_failed = true`.

---

## 2. Data Model

### 2.1 New Tables

```sql
-- 00041_create_gallery_folders_table.sql
CREATE TABLE gallery_folders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    is_public BOOLEAN NOT NULL DEFAULT FALSE,
    cover_media_id UUID,
    created_by UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX uq_gallery_folders_name ON gallery_folders (LOWER(name));
CREATE INDEX idx_gallery_folders_created_at ON gallery_folders(created_at DESC);
CREATE INDEX idx_gallery_folders_public_created ON gallery_folders(is_public, created_at DESC);

-- 00042_create_gallery_media_table.sql
CREATE TYPE gallery_media_type AS ENUM ('image', 'video');

CREATE TABLE gallery_media (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    folder_id UUID NOT NULL REFERENCES gallery_folders(id) ON DELETE CASCADE,
    media_type gallery_media_type NOT NULL,
    file_url VARCHAR(500) NOT NULL,             -- private original storage key, never a raw public URL
    thumbnail_url VARCHAR(500),                 -- private thumbnail storage key or NULL
    thumbnail_failed BOOLEAN NOT NULL DEFAULT FALSE,
    original_filename VARCHAR(255) NOT NULL,
    mime_type VARCHAR(100) NOT NULL,
    file_size BIGINT NOT NULL,                  -- bytes
    width INTEGER,                              -- nullable for unknown
    height INTEGER,                             -- nullable for unknown
    duration_seconds DOUBLE PRECISION,          -- nullable for images / unknown
    is_public BOOLEAN NOT NULL DEFAULT FALSE,
    uploaded_by UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_gallery_media_size CHECK (file_size > 0)
);
ALTER TABLE gallery_folders
    ADD CONSTRAINT fk_gallery_folders_cover_media
    FOREIGN KEY (cover_media_id) REFERENCES gallery_media(id) ON DELETE SET NULL;

CREATE INDEX idx_gallery_media_folder_id      ON gallery_media(folder_id);
CREATE INDEX idx_gallery_media_is_public      ON gallery_media(is_public);
CREATE INDEX idx_gallery_media_folder_public  ON gallery_media(folder_id, is_public, created_at DESC);
CREATE INDEX idx_gallery_media_created_at     ON gallery_media(created_at DESC);
```

### 2.2 Domain Entities

```go
// domain/gallery/gallery.go

type Folder struct {
    ID                 string
    Name               string
    Description        string
    IsPublic           bool
    CoverMediaID       *string  // persisted; nullable
    CoverThumbnailURL  *string  // derived per response scope; not persisted
    MediaCount         int      // populated by List queries; not persisted
    CreatedBy          string
    CreatedAt          time.Time
    UpdatedAt          time.Time
}

type Media struct {
    ID                string
    FolderID          string
    MediaType         MediaType  // "image" | "video"
    FileURL           string
    ThumbnailURL      *string
    ThumbnailFailed   bool
    OriginalFilename  string
    MimeType          string
    FileSize          int64
    Width             *int
    Height            *int
    DurationSeconds   *float64
    IsPublic          bool
    UploadedBy        string
    CreatedAt         time.Time
    UpdatedAt         time.Time
}

type MediaType string
const (
    MediaTypeImage MediaType = "image"
    MediaTypeVideo MediaType = "video"
)
```

### 2.3 Validation Rules

| Field | Rule |
|---|---|
| Folder.Name | required, ≤ 255 chars, unique (case-insensitive) |
| Folder.Description | ≤ 1000 chars |
| Folder.IsPublic | defaults to `false`; public access still requires public media |
| Media.FileURL | required, ≤ 500 chars |
| Media.MimeType | must be in allowlist (see §4.2) |
| Media.FileSize | > 0, ≤ 10 MB (image) / ≤ 100 MB (video) |
| Media.IsPublic | required |

---

## 3. Storage Layout

```
./uploads/
├── public/                     # existing static files exposed by /uploads/*
├── ppid/                       # existing protected documents
└── private/
    └── gallery/
        ├── originals/          # UUID.<ext>
        │   ├── 9f2a1b…c4.jpg
        │   └── 7c8d3e…f5.mp4
        └── thumbnails/         # always <uuid>.webp (or absent on failure)
            └── 9f2a1b…c4.webp
```

All gallery originals and thumbnails stay outside `FileUpload.GetPublicUploadDirectory()`, so the existing `/uploads/*` static route cannot bypass visibility checks. `FileUploadConfig.GetPrivateUploadDirectory()` defaults to `./uploads/private/`; `LocalHandlerWithSubdir(privateRoot, "gallery/originals")` and `LocalHandlerWithSubdir(privateRoot, "gallery/thumbnails")` are wired in `cmd/serve.go`.

Files do not move when visibility changes. Dedicated gallery binary endpoints resolve media by ID, enforce the two-level visibility gate for public requests, then stream from private storage. Admin binary endpoints require JWT + `gallery:read`. Unknown, private, and cross-scope IDs all return 404.

---

## 4. Thumbnail Pipeline

### 4.1 For Images

```
func GenerateImageThumbnail(src io.Reader, maxW, maxH int) ([]byte, error)
```

- Decode via `image.Decode` (jpeg/png/webp/gif — auto-detected).
- Resize using `golang.org/x/image/draw` `CatmullRom` (high quality).
- Encode as lossless WebP via `github.com/HugoSmits86/nativewebp`; `thumbnail_quality` maps to encoder compression effort.
- Bounds: fit within `maxW × maxH`, preserve aspect, never upscale.

### 4.2 For Videos

```
func GenerateVideoThumbnail(ctx context.Context, srcPath, dstPath string) error
```

- Probe with `ffprobe -v error -show_entries format=duration -of csv=p=0 <src>` → duration.
- Extract frame: `ffmpeg -y -ss 1 -i <src> -frames:v 1 -vf scale=400:400:force_original_aspect_ratio=decrease -c:v libwebp -lossless 0 -q:v 80 <dst>`.
- Capture `width × height` from ffprobe output for the stream.
- If `ffmpeg`/`ffprobe` missing → return `ErrFFmpegUnavailable`; service persists media with `thumbnail_url = NULL` + `thumbnail_failed = true`.

### 4.3 Configuration

```toml
# config/gallery.go
[gallery]
thumbnail_max_width   = 400
thumbnail_max_height  = 400
thumbnail_quality     = 80              # 0-100
image_max_size_mb     = 10
video_max_size_mb     = 100
bulk_upload_max_files = 20
bulk_upload_max_total_mb = 250
ffmpeg_path           = "ffmpeg"
ffprobe_path          = "ffprobe"
video_thumbnail_frame_seconds = 1       # ignored if duration < this
```

All overridable via env: `GALLERY_THUMBNAIL_MAX_WIDTH`, `GALLERY_FFMPEG_PATH`, etc.

### 4.4 Sync vs Async

v1: **synchronous**. Upload request blocks until thumbnail is generated (typical < 500 ms for images, 1–3 s for video frame extraction).

Future: a `pkg/worker` background queue with `asynq` or similar. Marked as a follow-up — see `tasks.md §6`.

---

## 5. Folder Cover Selection

The service calls `Repository.RecomputeFolderCover(ctx, folderID)`. The PostgreSQL implementation performs the lock, candidate selection, and cover update atomically in one transaction.

```sql
-- pseudo
SELECT m.id
FROM gallery_media m
JOIN gallery_folders f ON f.id = m.folder_id
WHERE m.folder_id = $1
  AND m.thumbnail_url IS NOT NULL
  AND (f.is_public = FALSE OR m.is_public = TRUE)
ORDER BY
    CASE WHEN m.media_type = 'image' THEN 0 ELSE 1 END,
    m.created_at DESC
LIMIT 1;
```

For a public folder, only public media may become the cover. For a private folder, any media may become the admin-only cover. The selected media ID is stored as `gallery_folders.cover_media_id`; handlers derive the correct public or admin thumbnail URL.

Called from:
- `CreateMedia` (after insert).
- `DeleteMedia`.
- `UpdateMediaVisibility`.
- `UpdateFolderVisibility`.

`RecomputeFolderCover` starts a transaction, locks the `gallery_folders` row with `SELECT ... FOR UPDATE`, selects the candidate, updates `cover_media_id`, and commits. No lock handle escapes the repository, preventing later queries from accidentally running on a different connection.

---

## 6. API Surface

All endpoints live under `/api/v1` in two gallery scopes: public and protected (JWT + RBAC).

### 6.1 Public

| Method | Path | Notes |
|---|---|---|
| GET | `/public/gallery/folders` | Folders where `folder.is_public = true` and public media count > 0. Query: `page`, `limit`, `q`. |
| GET | `/public/gallery/folders/{id}` | Public folder detail + public media. 404 if folder is private or has no public media. |
| GET | `/public/gallery/media/{id}` | Metadata only when both folder and media are public; otherwise 404. |
| GET | `/public/gallery/media/{id}/content` | Stream original when both visibility gates pass; otherwise 404. |
| GET | `/public/gallery/media/{id}/thumbnail` | Stream thumbnail when both visibility gates pass; otherwise 404. |

### 6.2 Admin (gallery:read)

| Method | Path | Notes |
|---|---|---|
| GET | `/gallery/folders` | All folders; filter by `is_public`, with total and public media counts. |
| GET | `/gallery/folders/{id}` | Full folder detail with all media and both visibility flags. |
| GET | `/gallery/media` | Flat list of all media (filter: `folder_id`, `is_public`, `media_type`, `q`). |
| GET | `/gallery/media/{id}` | Single media metadata at any visibility. |
| GET | `/gallery/media/{id}/content` | Stream original with JWT + `gallery:read`. |
| GET | `/gallery/media/{id}/thumbnail` | Stream thumbnail with JWT + `gallery:read`. |

### 6.3 Admin (gallery:write)

| Method | Path | Notes |
|---|---|---|
| POST   | `/gallery/folders` | Create folder. Body: `{name, description?, is_public?}`; visibility defaults private. |
| PUT    | `/gallery/folders/{id}` | Update name/description. |
| PATCH  | `/gallery/folders/{id}/visibility` | Toggle folder `is_public`; preserve media flags and recompute cover. |
| DELETE | `/gallery/folders/{id}` | Cascade delete media + files. |
| POST   | `/gallery/folders/{id}/media` | Multipart `media[]` upload (bulk). |
| PATCH  | `/gallery/media/{id}/visibility` | Body: `{is_public: bool}`. |
| POST   | `/gallery/folders/{id}/media/visibility` | Body: `{media_ids: [...], is_public: bool}`. Bulk. |
| DELETE | `/gallery/media/{id}` | Delete single media + file + thumbnail. |
| POST   | `/gallery/media/{id}/regenerate-thumbnail` | Re-run thumbnail pipeline. |

### 6.4 Response Shapes

**Folder (public list)**:
```json
{
  "id": "uuid",
  "name": "Panen Raya 2026",
  "description": "...",
  "cover_thumbnail_url": "/api/v1/public/gallery/media/9f2a1b…c4/thumbnail",
  "media_count": 24,
  "created_at": "2026-07-19T10:00:00+07:00"
}
```

**Folder (public detail)**:
```json
{
  "id": "uuid",
  "name": "...",
  "description": "...",
  "cover_thumbnail_url": "...",
  "media_count": 24,
  "media": [
    { "id": "...", "media_type": "image", "thumbnail_url": "...", "is_public": true, "width": 1920, "height": 1080, "created_at": "..." },
    ...
  ],
  "pagination": { "page": 1, "limit": 20, "total": 24, "total_pages": 2 }
}
```

**Media** (admin):
```json
{
  "id": "uuid",
  "folder_id": "uuid",
  "media_type": "image|video",
  "file_url": "...",
  "thumbnail_url": "...",
  "thumbnail_failed": false,
  "original_filename": "DSC_1234.jpg",
  "mime_type": "image/jpeg",
  "file_size": 4521984,
  "width": 4032,
  "height": 3024,
  "duration_seconds": null,
  "is_public": true,
  "uploaded_by": "uuid",
  "created_at": "...",
  "updated_at": "..."
}
```

### 6.5 File Serving

- All gallery binaries live under the private upload root and are inaccessible through `/uploads/*`.
- Public originals and thumbnails use `/api/v1/public/gallery/media/{id}/content` and `/thumbnail`; repository lookup must confirm `folder.is_public = true AND media.is_public = true`.
- Admin originals and thumbnails use `/api/v1/gallery/media/{id}/content` and `/thumbnail` with JWT + `gallery:read`.
- Public failures return 404 for not found, private folder, private media, or missing file to avoid visibility oracles.
- Handlers stream files; they never expose filesystem paths or accept a client-supplied filename.

---

## 7. Repository / Service Contracts

```go
// usecase/gallery/repository.go
type Repository interface {
    CreateFolder(ctx context.Context, f *Folder) error
    UpdateFolder(ctx context.Context, f *Folder) error
    UpdateFolderVisibility(ctx context.Context, id string, isPublic bool) error
    DeleteFolder(ctx context.Context, id string) error
    GetFolderByID(ctx context.Context, id string) (*Folder, error)
    GetPublicFolderByID(ctx context.Context, id string) (*Folder, error)
    GetFolderByName(ctx context.Context, name string) (*Folder, error)
    ListFolders(ctx context.Context, q FolderListInput) ([]Folder, int, error)
    ListFoldersWithPublicMedia(ctx context.Context, q FolderListInput) ([]Folder, int, error)

    CreateMedia(ctx context.Context, m *Media) error
    UpdateMediaVisibility(ctx context.Context, id string, isPublic bool) error
    BulkUpdateMediaVisibility(ctx context.Context, folderID string, mediaIDs []string, isPublic bool) error
    DeleteMedia(ctx context.Context, id string) error
    GetMediaByID(ctx context.Context, id string) (*Media, error)
    GetPublicMediaByID(ctx context.Context, id string) (*Media, error)
    ListMediaByFolder(ctx context.Context, folderID string, publicOnly bool, q MediaListInput) ([]Media, int, error)
    ListAllMedia(ctx context.Context, q MediaListInput) ([]Media, int, error)

    RecomputeFolderCover(ctx context.Context, folderID string) error
}
```

```go
// usecase/gallery/service.go
type Service struct {
    repo            Repository
    originalHandler filehandler.SaveDeleter       // ./uploads/private/gallery/originals
    thumbHandler    filehandler.SaveDeleter       // ./uploads/private/gallery/thumbnails
    imageProcessor  ImageProcessor                // image → webp bytes
    videoProcessor  VideoProcessor                // filepath → webp bytes + metadata
    cfg             config.GalleryConfig
    clock           clock.Clock
}

type ImageProcessor interface {
    GenerateThumbnail(src io.Reader, maxW, maxH, quality int) (webpBytes []byte, width, height int, err error)
}

type VideoProcessor interface {
    GenerateThumbnail(ctx context.Context, srcPath, dstPath string) (width, height int, duration float64, err error)
    Available() bool
}
```

---

## 8. RBAC

Two new permissions:
- `gallery:read` — list/admin endpoints.
- `gallery:write` — create/update/delete/upload.

Add to `cmd/seed.go:seedRolePermissions` for `operator`:

```go
permissions = append(permissions,
    casbinPolicy{Resource: "gallery", Action: "read"},
    casbinPolicy{Resource: "gallery", Action: "write"},
)
```

`backup` and `restore` style reserved: `gallery:read` and `gallery:write` are both seeded for `operator` in v1.

---

## 9. Configuration Addition

New file `config/gallery.go` (mirroring `config/fileupload.go`). Also implement the currently missing `FileUploadConfig.GetPrivateUploadDirectory()` getter with default `./uploads/private/`:

```go
type GalleryConfig struct {
    ThumbnailMaxWidth         int     `mapstructure:"thumbnail_max_width"`
    ThumbnailMaxHeight        int     `mapstructure:"thumbnail_max_height"`
    ThumbnailQuality          int     `mapstructure:"thumbnail_quality"`
    ImageMaxSizeMB            int     `mapstructure:"image_max_size_mb"`
    VideoMaxSizeMB            int     `mapstructure:"video_max_size_mb"`
    BulkUploadMaxFiles        int     `mapstructure:"bulk_upload_max_files"`
    BulkUploadMaxTotalMB      int     `mapstructure:"bulk_upload_max_total_mb"`
    FfmpegPath                string  `mapstructure:"ffmpeg_path"`
    FfprobePath               string  `mapstructure:"ffprobe_path"`
    VideoThumbnailFrameSeconds int    `mapstructure:"video_thumbnail_frame_seconds"`
}
```

Defaults applied in `Get*()` methods (mirror `fileupload.go` pattern).

---

## 10. Dependencies

New Go imports (already common to the project's ecosystem):

- `golang.org/x/image/draw` for Catmull-Rom resize.
- `golang.org/x/image/webp` for WebP decoding.
- `github.com/HugoSmits86/nativewebp` for pure-Go lossless WebP encoding.

Decision: use `x/image/draw` directly and `nativewebp`; no CGO or external image codec is required.

Binary dependency: **ffmpeg + ffprobe** on `$PATH` (documented in README + `Dockerfile` install line: `apk add ffmpeg`).

---

## 11. Error Mapping

| Service error | HTTP status |
|---|---|
| `folder not found` | 404 |
| `media not found` | 404 |
| `duplicate folder name` | 409 |
| `invalid mime type` | 415 |
| `file too large` | 413 |
| `bulk upload: too many files` | 400 |
| `bulk upload: total size exceeded` | 413 |
| folder private, media private, or no public media (public request) | 404 |
| thumbnail generation failed | logged; upload still succeeds (200/201) with `thumbnail_failed: true` |

This continues the existing fragile-error-string pattern from `specs.md §11 #16`. Acceptable for v1 consistency; full refactor tracked separately.

---

## 12. Test Plan

**Unit** (alongside code):
- `domain/gallery/gallery_test.go` — validation rules, MediaType constants.
- `usecase/gallery/service_test.go` — folder CRUD, visibility toggle, cover selection logic with mock repo.
- `usecase/gallery/thumbnail_test.go` — image resize produces WebP bytes; ffmpeg fallback when `Available() == false`.

**Integration** (`integration/gallery_test.go`):
- Folder: create, list, update, delete, duplicate-name 409, case-insensitive uniqueness.
- Media upload: single image, single video, bulk 5 files, oversize 413, wrong MIME 415, missing folder 404.
- Visibility: cover the four folder/media visibility combinations, folder toggle preserving media flags, single media toggle, and bulk media toggle.
- Binary privacy: public metadata, original, and thumbnail return 404 unless both folder and media are public; admin read can stream either visibility.
- Cover auto-update: upload triggers cover set; delete clears/updates; folder or media visibility toggle recomputes using only eligible media.
- Thumbnail: image produces WebP, video without ffmpeg produces `thumbnail_failed = true`.
- Cascade delete: folder delete removes all media + thumbnail files from disk.

---

## 13. Risks & Mitigations

| Risk | Mitigation |
|---|---|
| ffmpeg not installed | Document in README + Dockerfile; service gracefully degrades to `thumbnail_failed = true` |
| Large video files exhaust memory | Stream `ffmpeg` output to disk, never load full video into memory |
| Concurrent cover recompute | `SELECT ... FOR UPDATE` on `gallery_folders` row |
| Guessed IDs exposing private originals or thumbnails | Keep all gallery files outside static root; stream only after ID lookup and two-level visibility/RBAC checks |
| Slow thumbnail gen on big folder batch uploads | Sync in v1; plan async worker as follow-up |
| WebP encoding memory blow-up on huge images | Cap input decode at 50 MP via image.DecodeConfig first |

---

## 14. Implementation Order

See `tasks.md` for the full breakdown. High-level:

1. Migrations + domain entities.
2. Repository interface + sqlx impl.
3. Service: folder CRUD.
4. File handler wiring + image thumbnail processor.
5. Service: media upload + folder/media visibility.
6. Video thumbnail processor (ffmpeg).
7. Cover recompute logic.
8. HTTP handlers + route.go + middleware RBAC.
9. Seed update + public endpoints.
10. Tests.
11. Docs (`feature-docs/gallery.md`).
