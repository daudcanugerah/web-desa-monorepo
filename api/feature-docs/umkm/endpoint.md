# UMKM — Endpoints

All HTTP routes, request/response shapes, validation, and RBAC for the UMKM feature. Public routes expose listings + categories; admin routes gate CRUD + bulk-media handling.

**Module:** `webdesa/api`
**Last updated:** 2026-08-23

## Routes

### Public (no auth)

| Method | Path | Query |
|---|---|---|
| GET | `/api/v1/public/umkm/list` | `page`, `limit` (no `q`, no `category`) |
| GET | `/api/v1/public/umkm/{id}` | |
| GET | `/api/v1/public/umkm/categories` | With `q` autocomplete |

### Admin (RBAC: `umkm:read`)

| Method | Path | Query |
|---|---|---|
| GET | `/api/v1/umkm` | `q`, `category`, `page`, `limit` |
| GET | `/api/v1/umkm/{id}` | |

### Admin (RBAC: `umkm:write`)

| Method | Path | Notes |
|---|---|---|
| POST | `/api/v1/umkm` | Multipart OR JSON. Required: `name` (≤255), `description`, `category` (UUID). Opt: `owner`, `address`, `phone`, `email`, `website`. `images_media_ids` repeatable (multipart) or JSON array |
| PUT | `/api/v1/umkm/{id}` | Same; optional `images_media_ids` |
| DELETE | `/api/v1/umkm/{id}` | Deletes all referenced media |
| POST | `/api/v1/umkm/upload-media` | Single-shot media upload (RBAC `umkm:write`); returns `{url, media_id}` |
| POST | `/api/v1/umkm/categories` | Body `{name}` |
| DELETE | `/api/v1/umkm/categories/{id}` | 409 if in use |

## Request/Response

### Create / Update

Content types: `multipart/form-data` OR `application/json`.

Fields:
- Required: `name` (≤255), `description`, `category` (UUID)
- Optional: `owner`, `address`, `phone`, `email`, `website`
- Images: `images_media_ids` — repeatable multipart field OR JSON array of UUID strings

Returns 413 (`ErrBulkTooManyFiles` / `ErrBulkTotalTooLarge`) when bulk caps (20 files / 250 MB total) are exceeded.

### Single-shot upload

`POST /api/v1/umkm/upload-media` returns:
```json
{ "url": "string", "media_id": "uuid" }
```

### Response shape

```json
{
  "id": "uuid",
  "name": "string",
  "category_id": "uuid",
  "category": { "id": "uuid", "name": "string" },
  "description": "string",
  "owner": "string|null",
  "address": "string|null",
  "phone": "string|null",
  "email": "string|null",
  "website": "string|null",
  "media": [
    { "media_id": "uuid", "url": "string", "thumbnail_url": "string" }
  ],
  "created_at": "RFC3339",
  "updated_at": "RFC3339"
}
```

Dangling media pruned on read; sweeps run on media delete via `OnMediaDeleted`.

### BulkLimits response

413 responses for `ErrBulkTooManyFiles` (>20 files) and `ErrBulkTotalTooLarge` (>250 MB total).

## Validation

Handler-level:
- `name` required, ≤ 255
- `description` required
- `category` required, must be existing UUID
- `email` (when present) has `omitempty,email` validator (must contain `@` and `.`)
- `phone` ≤ 50
- `website` ≤ 255 (no URL format check)
- `images_media_ids` bulk caps: 20 files / 250 MB total

Domain-level (`domain/umkm.Validate()`):
| Field | Rule |
|---|---|
| `Name` | Required, ≤ 255 chars |
| `Category` | Required (UUID) |
| `Description` | Required |
| `Owner`, `Address`, `Phone`, `Email`, `Website` | Optional (pointers); non-empty if present |
| `Phone` | ≤ 50 chars |
| `Website` | ≤ 255 chars |

Both validations run on Create — handler validates first, then domain.

## RBAC

| Permission | Gates |
|---|---|
| `umkm:read` | list/get (admin) + read categories |
| `umkm:write` | create/update/delete + upload-media + category CRUD |

## Known Issues

- **Inline `images[]` upload was removed** — clients must first call `POST /umkm/upload-media` then attach returned `media_id`s via `images_media_ids`.
- **Bulk caps hardcoded at 20 files / 250 MB total** — exceeded → 413 (`ErrBulkTooManyFiles` / `ErrBulkTotalTooLarge`).
- **Sweep via `OnMediaDeleted`** keeps `images_media_ids` in sync with gallery deletions, but sweeper is best-effort (no read-time re-pruning safety net beyond the per-read prune).

## Related

- [Concept](./concept.md)
- [Database](./database.md)
- [Gallery](../../feature-docs/gallery.md) — file storage + media events
