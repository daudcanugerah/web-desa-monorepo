# Galeri

Public gallery. Routes `/galeri` (album grid) and `/galeri/:id` (album detail).
Gated by `VITE_FEATURE_GALLERY_PUBLIC`; set it to `true` to render.

Visibility is **folder-level**: the API lists folders flagged public, and a
public folder exposes **all** of its media on the web (per-item `is_public`
is not applied to the public gallery). Publishing an album means making the
folder public.

## Data sources

| Source | Function | Endpoint |
|---|---|---|
| Album list | `getPublicGalleryFolders({ limit })` | `GET /public/gallery/folders` |
| Album detail | `getPublicGalleryFolderById(id)` | `GET /public/gallery/folders/{id}` |

Response shapes:
- List: `{ folders: [{id,name,description,is_public,cover_thumbnail_url,media_count,public_media_count,created_at,updated_at}], pagination }`
- Detail: `{ folder, media: [{id,media_type,original_filename,mime_type,thumbnail_url,content_url,width,height,duration_seconds,created_at}], pagination }`

## Media URLs

`cover_thumbnail_url`, `thumbnail_url`, `content_url` are signed paths
(`/api/v1/media/{id}/thumbnail|content?jwt=...`) resolved against the API origin
via `resolveGalleryAssetUrl` — never the SPA origin.

## Galeri.vue

- Grid of album cards; client-side search on `name`/`description`.
- `folderCoverUrl(folder)` → `resolveFolderCoverUrl`.
- Shows `public_media_count || media_count`.
- Empty/loading states via `EmptyState`/`LoadingSpinner`.

## GaleriDetail.vue

- Header + counters (foto/video).
- Type filter: Semua / Foto / Video (`media_type` or `mime_type` prefix).
- Thumbnail grid; click opens a `Teleport` lightbox with prev/next, keyboard
  (`Escape`/`ArrowLeft`/`ArrowRight`) and video playback.
- `resolveMediaThumbnailUrl` / `resolveMediaContentUrl` for assets.

## Backend

Public routes (`api/interface/http/handler/gallery/route.go`, `ScopePublic`):
- `GET /public/gallery/folders`
- `GET /public/gallery/folders/{id}`
- `GET /public/gallery/media/{id}`

Binary streaming is handled by the unified
`/api/v1/media/{id}/{content,thumbnail}?jwt=` endpoints, not the gallery routes.