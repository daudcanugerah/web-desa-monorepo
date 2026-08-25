# Gallery — Endpoints

**Module:** `webdesa/api`
**Last updated:** 2026-08-23

## Routes

All paths are under `/api/v1`. The legacy per-scope binary routes (`/public/gallery/...`, `/gallery/media/{id}/content`, `/gallery/media/{id}/thumbnail`) were removed in Task 7.4; every media binary stream now flows through the unified signed route.

### Unified Signed Media (`signed_media.go`)

| Method | Path | Notes |
|---|---|---|
| GET | `/media/{id}/content` | Streams original binary; JWT in `?jwt=` query string |
| GET | `/media/{id}/thumbnail` | Streams thumbnail; JWT in `?jwt=` query string |
| POST | `/media/refresh` | Body: existing JWT + optional `?id=`; returns fresh token with new `expires_at_unix` |

Scope is inferred from the token. Public token → only public-folder media. Admin token → any media. PPID token → only media in the `system/ppid` system folder. Returns `503` if `SignedURLService` is nil (i.e., `signed_urls_enabled=false` in config), `401` for invalid/expired/revoked tokens, `404` for missing media.

### Admin (RBAC: `gallery:read`)

| Method | Path | Notes |
|---|---|---|
| GET | `/gallery/folders` | All folders; query `q`, `is_public`, `page`, `limit` |
| GET | `/gallery/folders/{id}` | Folder detail with all media |
| GET | `/gallery/media` | Flat media list; query `folder_id`, `is_public`, `media_type`, `q`, `page`, `limit` |
| GET | `/gallery/media/{id}` | Metadata at any visibility |

### Admin (RBAC: `gallery:write`)

| Method | Path | Notes |
|---|---|---|
| POST | `/gallery/folders` | JSON `{name, description?, is_public?}`; private by default |
| PUT | `/gallery/folders/{id}` | JSON `{name, description}` |
| PATCH | `/gallery/folders/{id}/visibility` | JSON `{is_public}`; preserves media flags |
| PATCH | `/gallery/folders/{id}/cover` | JSON `{media_id}`; pins manual cover (403 on system folders) |
| DELETE | `/gallery/folders/{id}` | Cascades media rows and removes stored files |
| POST | `/gallery/folders/{id}/media` | Multipart field `media[]` or `media` (`pickMediaFiles`); returns per-file results |
| PATCH | `/gallery/media/{id}/visibility` | JSON `{is_public}` |
| POST | `/gallery/folders/{id}/media/visibility` | JSON `{media_ids, is_public}` |
| DELETE | `/gallery/media/{id}` | Removes media row and files |
| POST | `/gallery/media/{id}/regenerate-thumbnail` | Re-runs thumbnail generation |

## Request/Response

### Folder (`FolderResponse`)

```json
{
  "id": "uuid",
  "name": "string",
  "description": "string",
  "is_public": true,
  "cover_media_id": "uuid|null",
  "cover_thumbnail_url": "string",   // only when cover exists
  "media_count": 0,
  "public_media_count": 0,
  "creator": { ... },                // admin scope only
  "created_at": "RFC3339",
  "updated_at": "RFC3339"
}
```

### Media (`MediaResponse`)

```json
{
  "id": "uuid",
  "folder_id": "uuid",
  "media_type": "image|video|document",
  "original_filename": "string",
  "mime_type": "string",
  "file_size": 0,
  "width": 0,
  "height": 0,
  "duration_seconds": 0.0,
  "is_public": true,
  "thumbnail_url": "string?",        // null when thumbnail_failed
  "thumbnail_failed": false,
  "content_url": "string",
  "uploaded_by": "uuid",             // admin only; omitted from public
  "created_at": "RFC3339",
  "updated_at": "RFC3339"
}
```

List responses wrap `folders` or `media` plus a `pagination` block.

### Binary Cache-Control

| Scope | Header |
|---|---|
| Public (signed) | `Cache-Control: public, max-age=300` |
| Admin (signed or legacy) | `Cache-Control: private, max-age=300` |

## Validation

### Handler-level

- Folder name/description length, required fields via `pkg/handlerutil.ValidateStruct` and per-field checks in `interface/http/handler/gallery/gallery.go`.
- Multipart uploads: `pickMediaFiles` accepts field names `media[]` or `media`; per-file validation rejects disallowed MIME types and oversize files.
- SetFolderCover request validates `media_id` is a valid UUID and that the target media belongs to the folder; service-layer rejects system folders with `403`.

### Domain-level

`domain/gallery/folder.go` and `domain/gallery/media.go` validation rules:

| Field | Rule |
|---|---|
| `Folder.Name` | Trimmed, required, max 255 chars, case-insensitively unique in persistence |
| `Folder.Description` | Max 1000 chars |
| `Folder.CreatedBy` | Required |
| `Media.FolderID` | Required |
| `Media.MediaType` | Must be `image`, `video`, or `document` |
| `Media.FileURL` | Required, max 500 chars |
| `Media.ThumbnailURL` | If present, non-empty and max 500 chars |
| `Media.OriginalFilename` | Required, max 255 chars |
| `Media.MimeType` | Required, max 100 chars |
| `Media.FileSize` | Greater than zero |
| `Media.UploadedBy` | Required |

### Upload Limits and Status Codes

| Failure | Code |
|---|---|
| Total bulk size > 250 MB | `413` |
| Number of files > 20 | `400` |
| MIME not allowed | `415` |
| Per-file size > limit | `413` |
| Duplicate folder name | `409` |
| Invalid ID | `400` |
| Missing resource | `404` |
| Targeting system folder (cover) | `403` |

Defaults: 10 MB per image, 100 MB per video, 20 files per bulk request, 250 MB total per bulk request (`config/gallery.go:30-97`).

## RBAC

- `gallery:read` — protected folder/media reads and metadata.
- `gallery:write` — folder CRUD, uploads, visibility changes, cover pinning, deletion, thumbnail regeneration.

The `operator` role receives both permissions in `cmd/seed.go:270-271`; `admin` receives wildcard access. Public signed route and the public folder/media listings are not gated by RBAC — they are gated by signed-URL scope and public visibility flags only.

## Known Issues

- **Public gallery routes removed (Task 7.4)** — `/public/gallery/...`, `/gallery/media/{id}/content`, and `/gallery/media/{id}/thumbnail` no longer exist. Clients still hitting them get `404`. The unified `/api/v1/media/{id}/{content|thumbnail}?jwt=` is the only binary route.
- **Unified signed media routes added** — `/api/v1/media/{id}/content`, `/api/v1/media/{id}/thumbnail`, `/api/v1/media/refresh`. Returns `503` when `signed_urls_enabled=false` (handler is mounted but service is nil).
- **`cover_manual` blocks auto-recompute** — once an admin pins a cover via `PATCH /gallery/folders/{id}/cover`, `RecomputeFolderCover` short-circuits on every subsequent visibility change until the admin clears the pin (sets `media_id: null`). Cover persists even after deleting the underlying media (`cover_media_id` would be set NULL by the FK cascade, leaving `cover_manual=TRUE` with no media — known edge).
- **Document MIME validation error string** — see [database Known Issues](./database.md); the runtime accepts `document` correctly, the validation error message in `domain/gallery/media.go:55` is stale.
- **Signed URL deny-list is single-instance** — `DenyList` is in-memory only. A multi-instance deployment needs a shared backend (Redis) for cross-instance revocation of PPID tokens. See [concept Glossary](./concept.md).

## Related

- [Concept](./concept.md)
- [Database](./database.md)
- `interface/http/handler/gallery/gallery.go` — admin handlers
- `interface/http/handler/gallery/signed_media.go` — unified signed route
- `interface/http/handler/gallery/binary.go` — `streamBinary` helper
- `config/gallery.go` — `SignedURLsEnabled` config
