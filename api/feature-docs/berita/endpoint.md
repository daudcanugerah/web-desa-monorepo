# Berita — Endpoints

All HTTP routes, request/response shapes, validation, and RBAC for the berita (news) feature. Public routes expose the article feed; admin routes gate CRUD + Quill media upload behind RBAC.

**Module:** `webdesa/api`
**Last updated:** 2026-08-23

## Routes

### Public (no auth)

| Method | Path | Notes |
|---|---|---|
| GET | `/api/v1/public/berita/list` | Query: `q`, `category`, `since`, `until`, `page`, `limit`. Only `active` articles. List response omits `content`. |
| GET | `/api/v1/public/berita/{id}` | 404 when the article is `inactive` |
| GET | `/api/v1/public/berita/categories` | With `q` search for autocomplete |

### Admin (RBAC: `berita:read`)

| Method | Path | Query |
|---|---|---|
| GET | `/api/v1/berita` | `q`, `category` (UUID), `status` (`active`\|`inactive`), `since`, `until`, `sort`, `order`, `page`, `limit` |
| GET | `/api/v1/berita/{id}` | |

### Admin (RBAC: `berita:write`)

| Method | Path | Notes |
|---|---|---|
| POST | `/api/v1/berita` | Multipart: `title` (req), `content` (Quill delta JSON, req), `category` (UUID req), opt `status` (default `active`), opt `image` XOR `image_media_id` |
| PUT | `/api/v1/berita/{id}` | Multipart. Same mutex. Status is NOT changed here — use PATCH |
| PATCH | `/api/v1/berita/{id}/status` | JSON `{status}` (`active`\|`inactive`) |
| DELETE | `/api/v1/berita/{id}` | Returns **200 + message** |
| POST | `/api/v1/berita/upload-media` | Quill media upload (see below) |
| POST | `/api/v1/berita/categories` | Body `{name}` |
| DELETE | `/api/v1/berita/categories/{id}` | 409 if in use |

## Request/Response

### Create / Update (multipart)

- Fields: `title` (req, ≤255), `content` (Quill delta JSON, req), `category` (UUID, req)
- Optional: `status` (`active`|`inactive`, Create only, defaults to `active`)
- Optional: `image` (file) XOR `image_media_id` (UUID) — mutex enforced
- Service cleans up uploaded file on failure; deletes old media after successful new save on Update

### Quill Media Upload

`POST /api/v1/berita/upload-media` (`interface/http/handler/berita/berita_upload.go`):

- Multipart form, max **50 MB total**
- Image type only (`type=image`)
- Image cap from `config.GetImageMaxSizeMB()` (default 10 MB)
- Allowed MIME: `image/jpeg|png|gif|webp`
- Filename = `media_id` (UUID, no `tmp-yyyymmdd-` prefix) — no longer stored in `uploads/tmp/`
- Returned: `{url, media_id}`

The legacy `tmp-` prefix and `processDeltaImages` move-on-save flow were deleted; Quill delta content must already reference gallery media URLs.

### Response shape

```json
{
  "id": "uuid",
  "title": "string",
  "content": "Quill delta JSON",
  "media": {
    "media_id": "uuid",
    "url": "string",
    "thumbnail_url": "string"
  },
  "category_id": "uuid",
  "category": { "id": "uuid", "name": "string" },
  "status": "active",
  "created_at": "RFC3339",
  "updated_at": "RFC3339"
}
```

List endpoint omits `content`. Public `GetBeritaPublic` silently rewrites embedded media paths in `content` from gallery URLs to feature-scoped public stream URLs (`/api/v1/public/berita/media/{id}/...`).

### List filters

`q` (ILIKE on `title`), `category` (UUID), `status` (`active`|`inactive`), `since`/`until` (YYYY-MM-DD), `sort` (whitelist: `created_at`|`title`), `order` (whitelist: `asc`|`desc`), `page`, `limit`. Default ordering: `created_at DESC`. Public list always applies `status=active`.

### Delete response

```json
{ "message": "..." }
```
(200, not 204)

## Validation

Handler-level:
- `title` required, ≤ 255
- `content` required (Quill delta JSON)
- `category` required, must be existing UUID
- `status` optional on Create, `oneof active inactive`; PATCH body requires it
- `image` XOR `image_media_id` mutex
- `sort` whitelist: `created_at`|`title`
- `order` whitelist: `asc`|`desc`
- `since`/`until` validated as `YYYY-MM-DD`

Domain-level (`domain/berita.Validate()`):
- `Title` required, ≤ 255
- `Content` required
- `Category` required UUID
- `Status` must be `active` or `inactive`

## RBAC

| Permission | Gates |
|---|---|
| `berita:read` | list/get (admin) + read categories |
| `berita:write` | create/update/delete + upload-media + category CRUD |

## Known Issues

- **Quill upload is image-only.** No video / file MIME allowlist — by design (Quill image embed).
- **Filename pattern is `media_id`, not `tmp-yyyymmdd-`.** Quill delta content must reference gallery media URLs directly; the old move-on-save helper is gone.
- `usecase/berita.service.uploadsDir` field is retained for handler signature compatibility but is no longer read by the service after the `tmp-` flow was removed.

## Related

- [Concept](./concept.md)
- [Database](./database.md)
- [Gallery](../../feature-docs/gallery.md) — file storage
