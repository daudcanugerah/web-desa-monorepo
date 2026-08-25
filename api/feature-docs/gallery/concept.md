# Gallery — Concept

Gallery is the resource layer for folders, images, videos, documents, thumbnails, and feature-owned uploads. Public media requires both the containing folder and the media item to be public. All media binary access flows through the unified `/api/v1/media/{id}/...?jwt=` signed-URL route (Task 7.1) when `signed_urls_enabled` is on, and falls back to per-scope legacy routes when the flag is off.

**Module:** `webdesa/api`
**Last updated:** 2026-08-23

## Purpose

Provide a single canonical store for all uploaded media (images, videos, documents) and expose them to every feature through a uniform folder/media contract. Keep binaries private at rest, gate visibility through folder + media flags, and stream them through JWT-signed URLs so the public never sees the storage path.

## Business Rules

- **System folder immutability** — folders where `is_system=true` cannot have `PATCH /gallery/folders/{id}/cover` invoked against them (`interface/http/handler/gallery/gallery.go:537` returns `403`). Mutating `is_public` or `feature_slug` on a system folder is similarly blocked at the service layer. System folders are seeded once via `EnsureSystemFolders` and stay linked 1:1 with their owning feature via the partial unique index on `feature_slug`.
- **Status flags** — every media row carries `is_public` (default `false`), `thumbnail_failed` (default `false`), and per-type metadata. Folders carry `is_public`, `is_system`, `cover_manual`, `cover_media_id`. New folders are private by default. New media is private by default. `cover_manual=TRUE` short-circuits `RecomputeFolderCover` and commits without changing `cover_media_id`.
- **Public visibility gate** — effective public visibility is `folder.is_public AND media.is_public`; a public folder with no public media is excluded from public listings. Public repository queries enforce this gate, not only HTTP handlers.
- **Signed URL scopes + deny-list** — `SignedURLService` mints HS256 JWTs bound to `(media_id, scope, sub)`. Scope TTLs: `public` 24h (`anonymous`), `admin` 1h (`user:<id>`), `ppid` 1h (`ppid_request:<id>`). `Verify` rejects mismatched `media_id`, expired tokens, and denied `sub`s. `Refresh` re-mints a fresh token but the **old token is not added to the deny list** — expiry is the only invalidation. `DenyList` is an in-memory `map[string]struct{}` (synchronized); the PPID service calls `Revoke(sub)` to invalidate a previously-issued token when a citizen's request is denied. Scope match enforcement: `public` token → only public-folder media; `admin` → any media; `ppid` → only media in the `system/ppid` system folder.
- **Manual cover pinning** — `SetFolderCover(ctx, folderID, mediaID)` pins a media as cover. Passing `NULL` clears the pin and resets `cover_manual = FALSE`. Targeting a system folder returns `403`.
- **Folder cover auto-detection** — `RecomputeFolderCover` prefers images over videos, then newer media (`ORDER BY CASE WHEN m.media_type = 'image' THEN 0 ELSE 1 END, m.created_at DESC, m.id DESC`). Public folder covers cannot use private media (`f.is_public = FALSE OR m.is_public = TRUE`). Short-circuits when `cover_manual = TRUE`.
- **Document MIME support** — `document` is now a valid `MediaType` (migration 00043 added the enum value, `mediaTypeFromMIME` maps document MIMEs). Allowed at the storage layer.
- **`signed_urls_enabled` flag** — when `true`, all response mappers emit the unified `/api/v1/media/{id}/...?jwt=` URL pattern and `SignedMediaHandler` serves those routes; when `false`, mappers fall back to legacy per-scope URL builders and the unified route stays mounted but returns `503`.

## Architecture

### Files

| File | Purpose |
|---|---|
| `domain/gallery/folder.go` | `Folder` entity and validation (includes `CoverManual`) |
| `domain/gallery/media.go` | `Media`, `MediaType` (`image`/`video`/`document`), and validation |
| `usecase/gallery/service.go` | Folder/media CRUD, upload processing, visibility, thumbnails, cover recomputation, orphan sweep |
| `usecase/gallery/repository.go` | Repository contract for folders and media |
| `usecase/gallery/filestore.go` | Narrow `FileStore` port used by other features |
| `usecase/gallery/filestore_impl.go` | `FileStore` adapter backed by the Gallery service |
| `usecase/gallery/system_folders.go` | Canonical feature system-folder slugs and seed specs |
| `usecase/gallery/list_inputs.go` | `FolderListInput`, `MediaListInput` query types |
| `usecase/gallery/signed_url.go` | `SignedURLService` — JWT mint/verify, `DenyList`, scope TTLs |
| `usecase/gallery/thumbnail_image.go` | Image thumbnail generation |
| `usecase/gallery/thumbnail_video.go` | FFmpeg/FFprobe video thumbnail generation |
| `usecase/gallery/url.go` | URL builders (legacy + signed) |
| `usecase/gallery/storage.go` | `MediaBinary` stream type |
| `usecase/gallery/deletion_hub.go` | Cross-feature listener hub for media delete sweeps |
| `config/gallery.go` | Thumbnail, upload, bulk-upload, FFmpeg, and `SignedURLsEnabled` config |
| `interface/http/handler/gallery/gallery.go` | HTTP handlers, request/response types, binary streaming, cover pinning |
| `interface/http/handler/gallery/route.go` | Public + protected gallery route registration |
| `interface/http/handler/gallery/signed_media.go` | Unified signed-URL handler (`/api/v1/media/{id}/...`) |
| `interface/http/handler/gallery/binary.go` | Shared `streamBinary` helper |
| `interface/postgres/gallery_folder.go` | Folder SQL repository, `RecomputeFolderCover`, `SetFolderCover` |
| `interface/postgres/gallery_media.go` | Media SQL repository, `FindOrphanMedia` |
| `cmd/seed.go` | Seeds system folders via `EnsureSystemFolders` (line 136) |
| `cmd/gallery_gc.go` | `gallery-gc` CLI for orphan-media sweeper |
| `db/migrations/00041_create_gallery_folders_table.sql` | Creates `gallery_folders` |
| `db/migrations/00042_create_gallery_media_table.sql` | Creates `gallery_media` and its enum |
| `db/migrations/00043_add_gallery_system_folders_and_media_refs.sql` | Adds `document` enum value, system folders, feature `*_media_id` FKs |
| `db/migrations/00044_add_gallery_folder_cover_manual.sql` | Adds `cover_manual` column |

### Service deps

`usecase/gallery/service.go` — `Service{repo, originalStorage, thumbStorage, imageProcessor, videoProcessor, cfg, clock, deletionHub, signedURL}`.

### Repository interface methods

From `usecase/gallery/repository.go` and implementations in `interface/postgres/gallery_folder.go` + `interface/postgres/gallery_media.go`:

- `CreateFolder(ctx, *Folder) error`
- `UpdateFolder(ctx, *Folder) error`
- `DeleteFolder(ctx, id) error` (cascades media)
- `GetFolderByID(ctx, id, includeAllMedia) (*Folder, error)`
- `ListFolders(ctx, FolderListInput) ([]Folder, error)`
- `ListFoldersWithPublicMedia(ctx, FolderListInput) ([]Folder, error)`
- `GetPublicFolderByID(ctx, id) (*Folder, error)`
- `SetFolderCover(ctx, folderID, mediaID *uuid.UUID) error`
- `RecomputeFolderCover(ctx, folderID) error`
- `CreateMedia(ctx, *Media) error`
- `BulkCreateMedia(ctx, []*Media) error`
- `UpdateMediaVisibility(ctx, mediaID, isPublic) error`
- `BulkUpdateMediaVisibility(ctx, mediaIDs, isPublic) error`
- `DeleteMedia(ctx, id) error`
- `DeleteMediaForFeature(ctx, folderSlug, mediaID) error`
- `GetMediaByID(ctx, id) (*Media, error)`
- `ListMedia(ctx, MediaListInput) ([]Media, error)`
- `FindOrphanMedia(ctx, cutoff, limit) ([]Media, error)`

### SignedURLService

`usecase/gallery/signed_url.go` — HS256 JWTs bound to `(media_id, scope, sub)`. Methods: `MintPublic(ctx, mediaID)`, `MintAdmin(ctx, mediaID, userID)`, `MintPPID(ctx, mediaID, requestID)`, `Verify(ctx, token, mediaID)`, `Refresh(ctx, token, mediaID)`, `Revoke(sub)`, `IsDenied(sub) bool`.

### System folder specs

`usecase/gallery/system_folders.go:7-15` defines canonical slugs:

```
banner, berita, struktur, umkm, fasilitas, user, ppid
```

Private (`is_public=false`), feature-owned (`is_system=true`), `feature_slug` carries a partial unique index so each slug maps to at most one folder. Feature services resolve them through `FileStore` rather than creating ordinary folders. `EnsureSystemFolders` is called by `cmd/seed.go:136` immediately after seeding the admin user so a fresh `make db-reset` produces all seven system folders before any feature upload runs.

## Glossary

- **signed URL** — JWT-minted, time-boxed token bound to `(media_id, scope, sub)`. Carried in the `?jwt=` query parameter on `/api/v1/media/{id}/{content|thumbnail}`. Replaces legacy direct URL access.
- **deny list** — in-memory `map[string]struct{}` tracked by `SignedURLService`. PPID service calls `Revoke(sub)` to remove a `ppid_request:<id>` token when the citizen's request is denied. Single-instance only.
- **system folder** — `gallery_folders` row with `is_system=TRUE` and a `feature_slug`. Always private, never deletable through normal CRUD, holds media owned by a specific feature. Seeded by `EnsureSystemFolders`.
- **media type** — enum `gallery_media_type` value: `image`, `video`, `document` (added in 00043). Drives thumbnail routing (`image` → image processor, `video` → FFmpeg/FFprobe, `document` → no thumbnail).
- **`signedURLsEnabled` flag** — config key `signed_urls_enabled` (`config/gallery.go:24-127`). When `true`, unified signed route serves binaries and response mappers emit signed URLs. When `false`, unified route returns `503` and mappers emit legacy per-scope URLs (admin binary, public gallery, PPID route).

## Related

- [Database](./database.md)
- [Endpoints](./endpoint.md)
- `draft/001-galery/design.md` — technical design (implemented)
- `draft/001-galery/requirements.md` — feature requirements (implemented)
- `draft/001-galery/tasks.md` — implementation checklist (implemented)
- [file-uploads](./file-uploads.md) — legacy file storage behavior
- [ppid](./ppid.md) — PPID document workflow
- `recom.docs/security.md` — upload/static-file security recommendations
- `recom.docs/testing.md` — test recommendations
