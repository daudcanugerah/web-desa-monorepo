# Gallery — Implementation Tasks

> Step-by-step task list for the Gallery feature.
> Status: **IMPLEMENTED** — 2026-07-19.
> Last updated: 2026-07-19
> Companion docs: `requirements.md`, `design.md`

Each task lists the files to create/modify, the work to do, and the verification command.

---

## Phase 1 — Schema & Domain (no behaviour yet)

### Task 1.1 — Add Go dependencies

- **Files**: `go.mod`, `go.sum`
- **Work**:
  - Add `github.com/HugoSmits86/nativewebp` for pure-Go lossless WebP encoding.
  - Add `golang.org/x/image` for Catmull-Rom resize and WebP decoding.
- **Verify**: `go mod tidy && go build ./...`

### Task 1.2 — Migration: folders table

- **File**: `db/migrations/00041_create_gallery_folders_table.sql` (new)
- **Work**: Create `gallery_folders` table per `design.md §2.1`, including `is_public BOOLEAN NOT NULL DEFAULT FALSE`, case-insensitive unique-name index, and public-list index.
- **Verify**: `go run main.go migrate up && psql -c "\d gallery_folders"`

### Task 1.3 — Migration: media table + ENUM

- **File**: `db/migrations/00042_create_gallery_media_table.sql` (new)
- **Work**: Create `gallery_media_type` ENUM + `gallery_media` table + indexes per `design.md §2.1`, then add the nullable `gallery_folders.cover_media_id` FK after `gallery_media` exists.
- **Verify**: `go run main.go migrate up && psql -c "\d gallery_media"`

### Task 1.4 — Domain entities

- **Files** (new):
  - `domain/gallery/folder.go` — `Folder` struct + `Validate()`.
  - `domain/gallery/media.go` — `Media` struct, `MediaType` constants + `Validate()`.
- **Work**: Implement structs exactly as in `design.md §2.2`, including `Folder.IsPublic`, `Folder.CoverMediaID`, and `Media.IsPublic`. Validation rules in §2.3.
- **Verify**: `go build ./domain/gallery/...`

---

## Phase 2 — Persistence Layer

### Task 2.1 — Gallery repository (interface + sqlx)

- **Files**:
  - `usecase/gallery/repository.go` (new) — `Repository` interface per `design.md §7`.
  - `interface/postgres/gallery_folder.go` (new) — sqlx impl for folders.
  - `interface/postgres/gallery_media.go` (new) — sqlx impl for media.
- **Work**:
  - All hand-written SQL (no ORM).
  - Implement `RecomputeFolderCover` in one transaction: lock the folder with `SELECT ... FOR UPDATE`, select an eligible candidate using folder/media visibility, update `cover_media_id`, then commit.
  - Use `db:"..."` tags; nullable fields → `*string` / `*int`.
  - JSONB / typed columns consistent with rest of project.
- **Verify**: `go build ./...`

### Task 2.2 — sqlx test fixtures

- **File**: `integration/testdata/gallery.sql` (new, optional)
- **Work**: Optional — only if testcontainers benefit from a fixture file.
- **Verify**: N/A (covered in Task 5.x).

---

## Phase 3 — Service Layer (folder CRUD first)

### Task 3.1 — Gallery service skeleton

- **File**: `usecase/gallery/service.go` (new)
- **Work**:
  - `Service` struct (see `design.md §7`) — start with `repo`, `clock`, `cfg` only; file handlers added in Phase 4.
  - Define input/output types:
    - `CreateFolderInput`, `UpdateFolderInput`, `UpdateFolderVisibilityInput`, `FolderListInput`
    - `CreateMediaInput`, `UpdateMediaVisibilityInput`, `MediaListInput`
  - Default folder and media visibility to private.
- **Verify**: `go build ./usecase/gallery/...`

### Task 3.2 — Folder service methods

- **File**: `usecase/gallery/service.go`
- **Work**: Implement `CreateFolder`, `UpdateFolder`, `UpdateFolderVisibility`, `DeleteFolder`, `GetFolderByID`, `GetPublicFolderByID`, `ListFolders`, `ListFoldersWithPublicMedia`.
- **Edge cases**:
  - Case-insensitive name uniqueness (use `LOWER()` in repo query).
  - Delete cascades to media (FK CASCADE).
  - `ListFoldersWithPublicMedia` requires `folder.is_public = true` and `COUNT(public media) > 0`.
  - Folder visibility toggle preserves every media visibility flag and recomputes the cover.
- **Verify**: `go test ./usecase/gallery/...`

### Task 3.3 — Folder service unit tests

- **File**: `usecase/gallery/folder_service_test.go` (new)
- **Work**: Cover happy paths, private defaults, folder visibility toggles, preserved media flags, validation, and duplicate names with a mock repo.
- **Verify**: `go test -run Gallery ./usecase/gallery/...`

---

## Phase 4 — File Storage + Image Thumbnails

### Task 4.1 — Gallery config struct

- **Files**: `config/gallery.go` (new), `config/gallery_defaults.go` (new), `config/fileupload.go` (modify).
- **Work**: Define `GalleryConfig` + `Get*()` methods with defaults (see `design.md §9`) and add `FileUploadConfig.GetPrivateUploadDirectory()` defaulting to `./uploads/private/`.
- **Verify**: `go build ./config/...`

### Task 4.2 — Wire file handlers in cmd/serve.go

- **File**: `cmd/serve.go`
- **Work**:
  - Resolve `privateRoot := cfg.FileUpload.GetPrivateUploadDirectory()`.
  - Construct two `LocalHandler` instances:
    - `galleryOriginals := file.NewLocalHandlerWithSubdir(privateRoot, "gallery/originals")`
    - `galleryThumbs := file.NewLocalHandlerWithSubdir(privateRoot, "gallery/thumbnails")`
  - Pass both to `galleryService`; never mount either directory under `/uploads/*`.
- **Verify**: `go build ./cmd/...`

### Task 4.3 — Image thumbnail processor

- **File**: `usecase/gallery/thumbnail_image.go` (new)
- **Work**:
  - Implement `ImageProcessor` interface (see `design.md §7`).
  - Use `x/image/draw` to resize and `nativewebp` to encode lossless WebP.
  - Return width/height of *resized* thumbnail (not original).
- **Verify**: `go test ./usecase/gallery/...`

### Task 4.4 — Thumbnail processor unit tests

- **File**: `usecase/gallery/thumbnail_image_test.go` (new)
- **Work**:
  - Golden-file test: encode a known small JPEG → known-ish WebP bytes (assert dimensions).
  - Verify upscaling is skipped (input smaller than max → output == input).
- **Verify**: `go test ./usecase/gallery/...`

---

## Phase 5 — Media Upload + Visibility

### Task 5.1 — Media service: CreateMedia (image)

- **File**: `usecase/gallery/service.go`
- **Work**:
  - Validate MIME + size.
  - Save original via `originalHandler`.
  - Generate image thumbnail → save via `thumbHandler`.
  - Build entity with `is_public = false` + `repo.CreateMedia`.
  - Call `repo.RecomputeFolderCover` using folder visibility.
  - On any failure: delete saved files before returning error.
- **Verify**: integration test (Task 7.x)

### Task 5.2 — Media service: CreateMedia (video)

- Same as 5.1 but route through `VideoProcessor`.
- **Work**: If `videoProcessor.Available() == false` → persist with `thumbnail_url = NULL`, `thumbnail_failed = true` (no error).

### Task 5.3 — Visibility methods

- **File**: `usecase/gallery/service.go`
- **Work**:
  - `UpdateFolderVisibility(ctx, id, isPublic)` — update only the folder flag, preserve all media flags, recompute cover.
  - `UpdateMediaVisibility(ctx, id, isPublic)` — single media toggle, recompute cover.
  - `BulkUpdateMediaVisibility(ctx, folderID, mediaIDs, isPublic)` — validate all media belong to the folder, update all in one tx, recompute cover once.
  - Public lookup methods enforce `folder.is_public = true AND media.is_public = true` in SQL, not only in handlers.
  - `DeleteMedia(ctx, id)` — delete files + thumbnail, recompute cover.
- **Verify**: unit tests + integration tests for all four folder/media visibility combinations.

### Task 5.4 — Folder cover recompute

- **Files**: `usecase/gallery/service.go`, `interface/postgres/gallery_folder.go`.
- **Work**:
  - Service calls `repo.RecomputeFolderCover(ctx, folderID)` after upload, delete, and folder/media visibility changes.
  - Repository performs `BEGIN` → `SELECT ... FOR UPDATE` folder → select eligible candidate → update nullable `cover_media_id` → `COMMIT` on one transaction-bound connection.
  - Public folders select only public media; private folders may select any media for their admin-only cover.
- **Verify**: integration tests for concurrent uploads and cover changes after folder/media visibility toggles.

---

## Phase 6 — Video Thumbnail Processor

### Task 6.1 — FFmpeg subprocess wrapper

- **File**: `usecase/gallery/thumbnail_video.go` (new)
- **Work**:
  - `VideoProcessor` impl using `os/exec` to invoke ffmpeg + ffprobe.
  - `GenerateThumbnail(ctx, srcPath, dstPath)`:
    1. Probe duration (ignore error → use 0).
    2. Pick frame time: `min(cfg.VideoThumbnailFrameSeconds, duration/2)`.
    3. Run ffmpeg with `-y -ss <t> -i <src> -frames:v 1 -vf scale=400:-2 -c:v libwebp -lossless 0 -q:v 80 <dst>`.
    4. Run ffprobe on src for width/height.
  - `Available()` — check `exec.LookPath(ffmpegPath)` on init.
- **Verify**: integration test with sample mp4 (use `ffmpeg`-generated test fixture if available; otherwise skip when ffmpeg missing).

### Task 6.2 — Container image update

- **File**: `Dockerfile`
- **Work**: Add `RUN apk add --no-cache ffmpeg` to runtime stage.
- **Verify**: `docker build .`

---

## Phase 7 — HTTP Layer

### Task 7.1 — Visibility-gated gallery binary endpoints

- **Files**: `interface/http/handler/gallery/gallery.go`, `interface/http/handler/gallery/route.go`.
- **Work**:
  - Add public `GET /public/gallery/media/{id}/content` and `/thumbnail`.
  - Resolve media by ID with repository queries requiring both folder and media public.
  - Add protected `GET /gallery/media/{id}/content` and `/thumbnail` under `gallery:read`.
  - Stream only from the configured private gallery roots; never accept a filename or expose filesystem paths.
  - Return the same 404 for missing, private-folder, private-media, and missing-file cases.
- **Verify**: integration matrix proves public binaries require both visibility gates and admin binaries allow either visibility.

### Task 7.2 — Gallery handler + route.go

- **Files** (new):
  - `interface/http/handler/gallery/gallery.go` — `GalleryHandler` struct + methods.
  - `interface/http/handler/gallery/route.go` — `RegisterRoutes(r, mw, scope)`.
- **Work**:
  - Implement every endpoint from `design.md §6`, including `PATCH /gallery/folders/{id}/visibility`.
  - Use httpin for body binding (multipart for upload endpoint).
  - Multipart parsing with `r.ParseMultipartForm(maxTotal)`.
  - Validate bulk upload count + total size early.
  - On visibility flip: respond 200 with updated entity.
- **Verify**: `go build ./...`

### Task 7.3 — Mount in router

- **File**: `interface/http/router.go`
- **Work**:
  - Call `gallery.RegisterRoutes(r, cfg.GalleryHandler, mw, middleware.ScopePublic)`.
  - Call `gallery.RegisterRoutes(r, cfg.GalleryHandler, mw, middleware.ScopeProtected)`.
- **Verify**: `go build ./...`

### Task 7.4 — DI wiring

- **File**: `cmd/serve.go`
- **Work**:
  - Construct `galleryRepo`, `imageProcessor`, `videoProcessor`, `galleryService`, `galleryHandler`.
  - Pass to router config.
- **Verify**: `go build ./cmd/...`

### Task 7.5 — OpenAPI annotations

- **File**: `interface/http/handler/gallery/gallery.go`
- **Work**: Add swag annotations (`@Summary`, `@Tags`, `@Param`, `@Success`, `@Router`, `@Security`) on every handler.
- **Verify**: `make swagger-gen && make swagger-validate`

---

## Phase 8 — RBAC & Seed

### Task 8.1 — Update available permissions list

- **File**: `usecase/role/service.go` (or wherever the hardcoded permission list lives — see `feature-docs/rbac.md`).
- **Work**: Add `"gallery"` to the resources list.
- **Verify**: `GET /api/v1/permissions` includes `gallery:read` and `gallery:write`.

### Task 8.2 — Seed operator permissions

- **File**: `cmd/seed.go`
- **Work**: In `seedRolePermissions`, add `gallery:read` + `gallery:write` to `operator`'s grant list.
- **Verify**: re-run `go run main.go seed`, check `casbin_rule` rows.

---

## Phase 9 — Integration Tests

### Task 9.1 — Folder integration tests

- **File**: `integration/gallery_folder_test.go` (new)
- **Coverage**:
  - Create defaults to private; list / get / update / delete.
  - Toggle folder visibility without changing contained media flags.
  - Filter admin list by `is_public`.
  - Duplicate name 409.
  - Update without auth 401.
  - List with `q` filter.
  - Pagination.

### Task 9.2 — Media upload integration tests

- **File**: `integration/gallery_media_upload_test.go` (new)
- **Coverage**:
  - Single image upload → 201, media private by default, thumbnail created, private-folder admin cover set.
  - Single video upload → 201, media private by default, ffmpeg call attempted.
  - Bulk upload (5 files) → 201 array.
  - Oversized image → 413.
  - Wrong MIME → 415.
  - Missing folder → 404.

### Task 9.3 — Visibility integration tests

- **File**: `integration/gallery_visibility_test.go` (new)
- **Coverage**:
  - Folder private + media private → public metadata/content/thumbnail all 404.
  - Folder private + media public → all public endpoints still 404.
  - Folder public + media private → all public endpoints still 404.
  - Folder public + media public → metadata/content/thumbnail succeed.
  - Toggling folder private hides everything without rewriting media flags; toggling public restores only previously public media.
  - Bulk media toggle updates all requested items atomically and recomputes cover once.

### Task 9.4 — Cover recompute tests

- **File**: `integration/gallery_cover_test.go` (new)
- **Coverage**:
  - Private folder may use latest private image/video as its admin-only `cover_media_id`.
  - Public folder ignores private media and chooses the latest public image, then public video.
  - Public folder with no public thumbnail has `cover_media_id = NULL` and remains absent from the public list.
  - Delete the cover media → recompute picks the next eligible item.
  - Folder or current-cover media visibility toggle recomputes without exposing a private thumbnail.

### Task 9.5 — Thumbnail degradation test

- **File**: `integration/gallery_thumbnail_test.go` (new)
- **Coverage**:
  - Video upload with ffmpeg missing → `thumbnail_failed = true`, upload still 201.
  - Image upscale skipped → thumbnail dimensions ≤ max.
- **Verify**: full integration suite passes — `make test-integration`

---

## Phase 10 — Documentation

### Task 10.1 — Feature doc

- **File**: `feature-docs/gallery.md` (new)
- **Work**: Mirror existing `feature-docs/*.md` template (Files → DB → domain → service → endpoints → response → RBAC → tests → related).
- **Verify**: markdown lints clean.

### Task 10.2 — Update specs.md

- **File**: `specs.md`
- **Work**: Add Gallery to the feature inventory table (Appendix A); add `gallery:read|write` to RBAC matrix (§6); add table to DB schema list (§5); add public endpoint count to §4.

### Task 10.3 — README quick-start

- **File**: `README.md`
- **Work**: Add `apk add ffmpeg` note (or `apt-get install ffmpeg`) under dev requirements.

### Task 10.4 — Recom docs entry

- **File**: `recom.docs/gallery.md` (new, optional)
- **Work**: List any open questions still unresolved post-implementation (e.g. async pipeline, soft delete, EXIF stripping).

---

## Phase 11 — Verification & Cleanup

### Task 11.1 — Full test run

```bash
make test-unit
make test-integration
make lint
make swagger-validate
```

All must pass.

### Task 11.2 — Manual smoke test

- Start server.
- Login as `admin@desa.local`.
- Create folder and confirm it defaults private.
- Upload 3 images + 1 video and confirm every media item defaults private.
- Toggle one media item public; confirm public endpoints still return 404 while the folder is private.
- Toggle the folder public; confirm the folder appears with only that media and a public-safe cover.
- Fetch its public metadata, original, and thumbnail; all succeed.
- Toggle the folder private; confirm all public endpoints return 404 while the media flag remains public.
- Delete folder → all private originals and thumbnails are removed from disk.

### Task 11.3 — Update AGENTS.md / CLAUDE.md (optional)

- Add `gallery` to keywords / file lists if relevant.

---

## Effort Estimate (rough)

| Phase | Effort |
|---|---|
| 1. Schema & domain | XS |
| 2. Persistence | S |
| 3. Folder service | S |
| 4. File + image thumb | M |
| 5. Media + visibility | M |
| 6. Video thumb | M |
| 7. HTTP layer | M |
| 8. RBAC | XS |
| 9. Tests | L |
| 10. Docs | S |

**Total**: ~1.5–2 days of focused work for a single engineer (tests included).

---

## Follow-ups (post-v1, backlog)

- Async thumbnail pipeline (`asynq` or river).
- Folder reordering (`display_order` int).
- EXIF stripping for privacy.
- Video transcoding to web-optimised MP4.
- Bulk upload UI with drag-drop progress.
- Per-folder quota / disk usage dashboard.
- Album-of-albums hierarchy.
- S3/MinIO storage adapter (deferred; current local FS works for v1).
