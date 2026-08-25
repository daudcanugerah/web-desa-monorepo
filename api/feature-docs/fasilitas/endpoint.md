# Fasilitas — Endpoints

All HTTP routes, request/response shapes, validation, and RBAC for the fasilitas (facility) feature. Public routes expose listings; admin routes gate CRUD + per-image removal + bbox queries.

**Module:** `webdesa/api`
**Last updated:** 2026-08-23

## Routes

### Public (no auth)

| Method | Path | Query |
|---|---|---|
| GET | `/api/v1/public/fasilitas/list` | `page`, `limit`, `q`, `category`. **No bbox.** |
| GET | `/api/v1/public/fasilitas/{id}` | |
| GET | `/api/v1/public/fasilitas/categories` | |

### Admin (RBAC: `fasilitas:read`)

| Method | Path | Query |
|---|---|---|
| GET | `/api/v1/fasilitas` | `minLat`, `maxLat`, `minLon`, `maxLon` (bbox, all-or-none), `q`, `category`, `page`, `limit` |
| GET | `/api/v1/fasilitas/{id}` | |

### Admin (RBAC: `fasilitas:write`)

| Method | Path | Notes |
|---|---|---|
| POST | `/api/v1/fasilitas` | Multipart OR JSON. Required: `name` (≤255), `latitude`, `longitude`. Optional: `category` (UUID), `description`, `images_media_ids` |
| PUT | `/api/v1/fasilitas/{id}` | Multipart. Same; optional `images_media_ids` |
| DELETE | `/api/v1/fasilitas/{id}` | |
| DELETE | `/api/v1/fasilitas/{id}/images/{imageIndex}` | Remove one image by zero-based index; service slices array, deletes gallery media |
| POST | `/api/v1/fasilitas/upload-media` | Single-shot media upload (RBAC `fasilitas:write`); returns `{url, media_id}` |
| POST | `/api/v1/fasilitas/categories` | Body `{name}` |
| DELETE | `/api/v1/fasilitas/categories/{id}` | 409 if in use |

## Request/Response

### Create / Update

Content types: `multipart/form-data` OR `application/json`.

Fields:
- Required: `name` (≤255), `latitude` (`[-90, 90]`), `longitude` (`[-180, 180]`)
- Optional: `category` (UUID, must exist), `description`
- Images: `images_media_ids` — repeatable multipart field OR JSON array of UUID strings

Returns 413 when bulk caps exceeded (>20 files / >250 MB total).

### Bounding Box Query Params

All four required if any provided (handler lines 284-288):
```
GET /api/v1/fasilitas?minLat=...&maxLat=...&minLon=...&maxLon=...&page=...&limit=...
```

If only some are supplied → 400 rejection. Service validates `MinLat < MaxLat` and `MinLon < MaxLon`.

### Per-image removal

`DELETE /api/v1/fasilitas/{id}/images/{imageIndex}` — `{imageIndex}` is zero-based. Service slices `ImagesMediaIDs` and deletes gallery media at that index.

### Single-shot upload

`POST /api/v1/fasilitas/upload-media` returns:
```json
{ "url": "string", "media_id": "uuid" }
```

### Response shape

```json
{
  "id": "uuid",
  "name": "string",
  "category_id": "uuid|null",
  "category": { "id": "uuid", "name": "string" },
  "latitude": -6.12345678,
  "longitude": 106.12345678,
  "description": "string|null",
  "media": [
    { "media_id": "uuid", "url": "string", "thumbnail_url": "string" }
  ],
  "created_at": "RFC3339",
  "updated_at": "RFC3339"
}
```

> **Critical field rename**: the response field is `category_id` (nullable UUID) + `category` ({id, name} object). The legacy `"type"` field (UUID) was removed — clients reading `response.type` will break. This was the single most dangerous drift in the upload migration.

### BBox Repository SQL

```sql
SELECT COUNT(*) FROM fasilitas WHERE 1=1
  AND latitude BETWEEN $1 AND $2
  AND longitude BETWEEN $3 AND $4

SELECT ... FROM fasilitas WHERE 1=1
  AND latitude BETWEEN $1 AND $2
  AND longitude BETWEEN $3 AND $4
ORDER BY created_at DESC
LIMIT $5 OFFSET $6
```

`$1=MinLat`, `$2=MaxLat`, `$3=MinLon`, `$4=MaxLon`. **BBox is required-all-or-none** — handler (lines 284-288) rejects requests where only some of the four params are supplied.

## Validation

Handler-level:
- `name` required, ≤ 255
- `latitude` required, `[-90, 90]`
- `longitude` required, `[-180, 180]`
- `category` (when present) must be existing UUID
- bbox params all-or-none
- bulk caps: 20 files / 250 MB total

Domain-level (`domain/fasilitas.Validate()`):
| Field | Rule |
|---|---|
| `Name` | Required, ≤ 255 chars |
| `Latitude` | `[-90, 90]` |
| `Longitude` | `[-180, 180]` |
| `Category` | If non-nil, must be non-empty UUID |
| `Description` | If non-nil, non-empty |
| `ImagesMediaIDs` | Each id non-empty, ≤ 64 chars (UUID cap, not 500-char URL) |

## RBAC

| Permission | Gates |
|---|---|
| `fasilitas:read` | list/get (admin) + read categories |
| `fasilitas:write` | create/update/delete + upload-media + per-image removal + category CRUD |

## Known Issues

- **Bounding-box query has no integration test coverage** — `integration/fasilitas_test.go` (13 cases) covers happy path / coord validation / list pagination / images array / empty images / timestamps, but NOT the geospatial bbox query. Add tests for the bbox (see `recom.docs/testing.md`).
- **Breaking field rename**: legacy API consumers reading `response.type` (UUID) will fail after deploys that adopt `category_id` + `category` — coordinate with frontend before rollout.
- **BBox is admin-only.** Public list endpoint does not expose bbox filtering (anonymous viewers see paginated list only).

## Related

- [Concept](./concept.md)
- [Database](./database.md)
- [Gallery](../../feature-docs/gallery.md) — file storage + media events
