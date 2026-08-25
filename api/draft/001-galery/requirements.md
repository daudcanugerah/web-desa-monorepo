# Gallery — Requirements

> Draft spec for the Gallery feature.
> Status: **IMPLEMENTED** — 2026-07-19.
> Last updated: 2026-07-19

## 1. Summary

A simple, admin-managed photo & video gallery for the village website. Admins organise media into named folders; both folders and individual media have independent `public` / `private` visibility. Public access uses a two-level gate: a media item is visible only when its folder is public **and** the media item itself is public.

The system automatically generates thumbnails for uploaded media and uses them to build a cover for each folder. The public website reads eligible public folders + their public media; the admin panel manages everything, including private folders and media.

## 2. In Scope (v1)

| # | User story |
|---|---|
| **R1** | An admin can **create a folder** (name + optional description) to group related media; new folders are private by default. |
| **R2** | An admin can **upload one or many photos or videos** into a folder. |
| **R3** | An admin can independently mark a **folder** and each **photo/video** as `public` or `private`; public access requires both to be public. |
| **R4** | The system **processes a thumbnail for every uploaded image and video** automatically. |
| **R5** | The system **chooses a thumbnail per folder** (cover image) automatically. |
| **R6** | The public website can **list folders** (with cover thumbnails) and **browse a folder's media** (public items only). |

## 3. Out of Scope (v1)

- Albums inside albums (folder hierarchy is flat).
- Comments / reactions / likes.
- Sharing links / expiring URLs.
- Public upload by visitors (admin-only).
- EXIF stripping / GPS privacy (best-effort only).
- AI tagging / face recognition.
- Video transcoding (videos served as uploaded — see Open Questions).

## 4. Detailed Requirements

### R1 — Create Folder

- **Inputs**: `name` (required, ≤ 255 chars), `description` (optional, ≤ 1000 chars), `is_public` (optional boolean, defaults to `false`).
- **Output**: Folder entity with `id`, `is_public`, `created_by`, timestamps.
- **Validation**:
  - Name must be unique per scope (case-insensitive).
  - Name must not be blank after trim.
- **Side effects**:
  - `created_by = authenticated user`.
  - `is_public = false` unless explicitly supplied.
  - `cover_media_id = NULL` until an eligible media item exists (R5).
- **Errors**:
  - 400 on validation failure.
  - 409 on duplicate name.

### R2 — Upload Photo/Video

- **Inputs**: `folder_id` (required UUID), one or more files via multipart `media[]`.
- **Allowed file types**:
  - Image: `image/jpeg`, `image/png`, `image/webp`, `image/gif`.
  - Video: `video/mp4`, `video/webm`, `video/quicktime`.
- **Size limits**:
  - Image: ≤ 10 MB per file.
  - Video: ≤ 100 MB per file.
- **Per-file processing**:
  1. Validate MIME + size.
  2. Generate UUID filename, save to `./uploads/private/gallery/originals/<id>.<ext>`.
  3. Generate thumbnail → `./uploads/private/gallery/thumbnails/<id>.webp` (always WebP for consistency).
  4. Extract `width`, `height`, `duration_seconds` (best-effort for video).
  5. Persist row with `is_public = false` (default; admin flips per R3).
- **Bulk upload**: accept `N` files in one request; return array of created media (success + per-file failures).
- **Errors**:
  - 404 if folder not found.
  - 413 if any file exceeds size limit.
  - 415 if MIME unsupported.
  - Thumbnail generation failure does not fail the upload; the created item reports `thumbnail_url = NULL` and `thumbnail_failed = true`.

### R3 — Folder and Media Visibility

- **Inputs**:
  - Folder: `folder_id`, `is_public` boolean.
  - Media: `media_id`, `is_public` boolean.
- **Output**: Updated folder or media entity.
- **Two-level gate**:
  - Folder private + media private → hidden.
  - Folder private + media public → hidden.
  - Folder public + media private → hidden.
  - Folder public + media public → visible on `/public/gallery/*`.
- **Semantics**:
  - New folders and newly uploaded media default to private.
  - Making a folder private immediately hides the folder and all contained media from public endpoints.
  - Folder visibility changes do not rewrite media visibility flags; making the folder public again restores access only to media already marked public.
  - A public folder with zero public media remains hidden from public folder listings.
  - **No `draft` / `archived` states** in v1 — strict public/private.
- **Operations**:
  - `PATCH /gallery/folders/{id}/visibility` with `{ is_public: bool }`.
  - `PATCH /gallery/media/{id}/visibility` with `{ is_public: bool }`.
  - `POST /gallery/folders/{id}/media/visibility` with `{ media_ids: [...], is_public: bool }` to flip many media items at once.
- **Errors**:
  - 404 if the folder or media does not exist.

### R4 — Thumbnail Processing

- **For images**:
  - Decode, resize to fit `MAX_THUMB_WIDTH × MAX_THUMB_HEIGHT` (default `400 × 400`, preserve aspect ratio, never upscale).
  - Encode as lossless **WebP**; configured quality maps to compression effort.
- **For videos**:
  - Extract the **first frame** (or the frame at `t = 1s` if probe succeeds and duration ≥ 1s) using **ffmpeg** as a subprocess (`ffmpeg -ss 1 -i input -frames:v 1 -q:v 2 thumb.webp`).
  - If ffmpeg is unavailable or fails → row persists with `thumbnail_url = NULL` + `thumbnail_failed = true` (logged with stderr).
- **Async vs sync** (decision pending — see `design.md` §4.4):
  - v1 recommendation: **sync** for simplicity; switch to background worker later if it becomes a bottleneck.
- **Regeneration**:
  - Admin endpoint `POST /gallery/media/{id}/regenerate-thumbnail` re-runs the pipeline (useful after config changes).

### R5 — Folder Thumbnail

- **Auto-selection rule** (deterministic):
  - For a **public folder**: latest public image, else latest public video, else `NULL`. Private media must never become its public cover.
  - For a **private folder**: latest image, else latest video, else `NULL`; this cover is admin-only.
  - Candidates without a generated thumbnail are skipped.
- **Re-evaluation triggers**:
  - Media uploaded to folder.
  - Media deleted from folder.
  - Media `is_public` toggled.
  - Folder `is_public` toggled.
- **Stored**: As nullable `gallery_folders.cover_media_id`; handlers derive a scope-safe thumbnail URL from the selected media ID.
- **No manual override in v1** — keep it pure automatic to match the requirement that the *system* chooses.

### R6 — Public Display

- **Listing folders**: `GET /public/gallery/folders`
  - Returns paginated folders where `folder.is_public = true` and at least one contained media item has `media.is_public = true`.
  - Private and empty-public folders are hidden.
  - Each entry includes: `id`, `name`, `description`, `cover_thumbnail_url`, `media_count` (public items only), `created_at`.
- **Folder detail**: `GET /public/gallery/folders/{id}`
  - Returns folder metadata + paginated media where both folder and media are public.
  - Returns 404 when the folder is private or has zero public media; do not reveal its existence.
- **Media direct view**: `GET /public/gallery/media/{id}`
  - Returns metadata only when both folder and media are public; otherwise 404.
- **Binary access**:
  - `GET /public/gallery/media/{id}/content` streams the original only when both visibility gates pass.
  - `GET /public/gallery/media/{id}/thumbnail` streams the thumbnail only when both visibility gates pass.
  - Gallery binaries are never served directly from the generic `/uploads/*` static route.

## 5. Non-Functional Requirements

- **Performance**: P95 list endpoint < 200 ms with 100 folders / 5000 media items on a single instance.
- **Concurrency**: 5 concurrent uploads to the same folder must not corrupt the folder cover (use a row-level lock or single-flight select-for-update when recomputing R5).
- **Storage**: Originals and thumbnails live outside the static public root under `./uploads/private/gallery/`, split into `originals/` and `thumbnails/`.
- **Privacy**: Gallery metadata and binaries are public only when `folder.is_public = true AND media.is_public = true`; all other combinations return 404 even if an ID or filename is guessed.
- **Audit**: `created_by` on every folder + media; surfaced in admin responses.

## 6. Open Questions (to resolve before implementation)

1. **Video thumbnail timing**: 0s vs 1s frame? (1s usually more representative, but videos < 1s would black-frame.)
2. **Bulk upload limits**: max files per request (suggest 20)? max total payload?
3. **Soft delete vs hard delete** for media? (Hard delete is simpler; soft delete adds `deleted_at` and complicates queries.)
4. **Reordering**: should folders be reorderable (`display_order` int)? v1 says no — sorted by `created_at DESC`.
5. **Folder rename**: rename endpoint, or delete+recreate like categories?
6. **Quota / disk usage**: track `total_bytes` per folder for admin display?
7. **Image dimension enforcement**: should we reject > N megapixel uploads to prevent OOM during thumbnail gen?

---

## 7. Acceptance Checklist

- [ ] Admin can create a folder with valid name.
- [ ] Admin gets 400/409 on invalid/duplicate name.
- [ ] Admin can upload image(s) and/or video(s) to a folder via multipart.
- [ ] Oversized or wrong-MIME files return 413/415.
- [ ] Each uploaded media has a generated thumbnail (WebP).
- [ ] Folder cover auto-updates on upload, delete, visibility toggle.
- [ ] New folders and media default to private.
- [ ] Admin can toggle `is_public` independently on a folder and media item.
- [ ] Bulk media visibility flip works.
- [ ] A private folder hides all contained media regardless of media flags.
- [ ] Re-publishing a folder preserves prior media visibility flags.
- [ ] Public list returns only public folders with ≥ 1 public media item.
- [ ] Public folder detail hides private media.
- [ ] Public metadata, original, and thumbnail endpoints return 404 unless both folder and media are public.
- [ ] Thumbnail regen endpoint works.
- [ ] Integration tests cover all of the above.
