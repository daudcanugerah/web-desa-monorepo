# Banner — Endpoints

All HTTP routes, request/response shapes, validation, and RBAC for the banner feature. Public routes expose the active stream; admin routes gate CRUD + category management behind RBAC.

**Module:** `webdesa/api`
**Last updated:** 2026-08-23

## Routes

### Public (no auth)

| Method | Path | Notes |
|---|---|---|
| GET | `/api/v1/public/banner` | Paginated public list |
| GET | `/api/v1/public/banner/{id}` | |
| GET | `/api/v1/public/banner/categories` | Category list (q autocomplete optional) |
| GET | `/api/v1/banners/active` | Up to 100 active banners (feature-scoped public stream URLs) |

### Admin (RBAC: `banners:read`)

| Method | Path | Query |
|---|---|---|
| GET | `/api/v1/banners` | `page`, `limit`, `status`, `q`, `category` (UUID) |
| GET | `/api/v1/banners/{id}` | |
| GET | `/api/v1/banners/categories` | Category list |

### Admin (RBAC: `banners:write`)

| Method | Path | Notes |
|---|---|---|
| POST | `/api/v1/banners` | Multipart: `title` (req), `description`, `link`, `image` XOR `image_media_id`, `category` (UUID opt), `metadata` (JSON string). Defaults status to `inactive`. |
| PUT | `/api/v1/banners/{id}` | Multipart. Status NOT changed |
| PATCH | `/api/v1/banners/{id}/status` | Body `{status: "active"\|"inactive"}`. **DB trigger enforces 20-active limit.** |
| DELETE | `/api/v1/banners/{id}` | Calls `fileStore.Delete` after row delete |
| POST | `/api/v1/banners/upload-media` | Single-shot media upload (RBAC `banners:write`); returns `{url, media_id}` |
| POST | `/api/v1/banners/categories` | Body `{name}` |
| DELETE | `/api/v1/banners/categories/{id}` | 409 if in use |

## Request/Response

### Create / Update (multipart)

- Fields: `title` (req, ≤255), `description`, `link` (≤500), `category` (UUID, opt), `metadata` (JSON string)
- Exactly one of: `image` (file) OR `image_media_id` (UUID) — mutex enforced
- Defaults `status` to `inactive` on Create

### Patch status

```json
{ "status": "active" }
```
Returns 409 / `check_violation` if 20-active limit reached.

### Response shape

```json
{
  "id": "uuid",
  "title": "string",
  "description": "string",
  "link": "string",
  "media": {
    "media_id": "uuid",
    "url": "string",
    "thumbnail_url": "string"
  },
  "category_id": "uuid|null",
  "category": { "id": "uuid", "name": "string" },
  "status": "active|inactive",
  "metadata": {},
  "created_at": "RFC3339",
  "updated_at": "RFC3339"
}
```

Public endpoints emit feature-scoped public stream URLs (`/api/v1/public/banner/media/{id}/...`) so anonymous visitors can render system-folder media.

## Validation

Handler-level:
- `title` required, ≤ 255
- `image` XOR `image_media_id` mutex
- `status` ∈ {`active`,`inactive`} on PATCH
- `metadata` parsed as JSON if present

Domain-level (`domain/banner.Validate()`):
- `ID` required
- `Title` ≤ 255 chars
- `ImageMediaID` required (no length cap — UUID only)
- `Status` must be `active` or `inactive`
- `CreatedAt`, `UpdatedAt` required

## RBAC

| Permission | Gates |
|---|---|
| `banners:read` | admin list/get + both categories list endpoints |
| `banners:write` | create/update/delete + upload-media + category CRUD |

## Known Issues

- `usecase/banner/file_handler.go` is dead code — the service no longer depends on `FileHandler` (replaced by `fileStore: galleryUsecase.FileStore`). File kept for build stability but is unread.

## Related

- [Concept](./concept.md)
- [Database](./database.md)
- [Gallery](../../feature-docs/gallery.md) — file storage + media events
