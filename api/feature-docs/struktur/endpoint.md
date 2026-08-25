# Struktur — Endpoints

**Module:** `webdesa/api`
**Last updated:** 2026-08-23

## Routes

### Public (no auth)

| Method | Path | Notes |
|---|---|---|
| GET | `/api/v1/public/struktur/list` | Paginated |
| GET | `/api/v1/public/struktur/{id}` | Profile image rendered via unified `/api/v1/media/{id}/...` (see Response Shape) |

### Admin (RBAC: `struktur:read`)

| Method | Path | Notes |
|---|---|---|
| GET | `/api/v1/struktur` | Query: `q`, `page`, `limit` |
| GET | `/api/v1/struktur/{id}` | |

### Admin (RBAC: `struktur:write`)

| Method | Path | Notes |
|---|---|---|
| POST | `/struktur/upload-media` | Multipart `file` (image). Returns `{media_id, url}` |
| POST | `/api/v1/struktur` | Multipart. Required: `name`. Mutex: `profile_image` file OR `profile_image_media_id`. Optional: `position`, `email`, `phone`, `description` |
| PUT | `/api/v1/struktur/{id}` | Multipart, same mutex |
| DELETE | `/api/v1/struktur/{id}` | |

## Request/Response

### Create / Update (admin)

```
POST /api/v1/struktur   (multipart/form-data)
PUT  /api/v1/struktur/{id}
```

| Form field | Type | Required | Notes |
|---|---|---|---|
| `name` | string | yes | ≤ 255 |
| `position` | string | no | |
| `email` | string | no | basic shape check if present |
| `phone` | string | no | |
| `description` | string | no | |
| `profile_image` | file | mutex | raw upload → creates gallery media |
| `profile_image_media_id` | UUID | mutex | references existing gallery media |

Exactly one of `profile_image` / `profile_image_media_id` must be supplied.

### Response Shape

```json
{
  "id": "uuid",
  "name": "...",
  "position": "...",
  "email": "...",
  "phone": "...",
  "profile_image_url": "...",        // deprecated, omitempty
  "profile_media": {                  // canonical
    "media_id": "uuid",
    "url": "/api/v1/media/{id}/content?jwt=...",
    "thumbnail_url": "/api/v1/media/{id}/thumbnail?jwt=..."
  },
  "description": "...",
  "created_at": "RFC3339",
  "updated_at": "RFC3339"
}
```

Public endpoints embed the unified `/api/v1/media/{id}/...?jwt=` URL (gallery `SignedURLService`, `scope=public`). No feature-scoped public media stream routes for struktur.

### Upload media

```
POST /struktur/upload-media   (multipart/form-data)
```

**Response**:
```json
{"media_id": "uuid", "url": "/api/v1/media/{id}/content?jwt=..."}
```

## Validation

### Handler-level

- `name` non-empty, ≤ 255 chars
- `email` valid email format (if present)
- `phone` ≤ 50 chars (if present)
- `description` non-empty if present
- Multipart mutex: `profile_image` file vs `profile_image_media_id` — exactly one

### Domain-level (`domain/struktur/struktur.go`)

- `Name` required, ≤ 255 chars
- `Position` optional
- `Email` optional (basic shape check if present)
- `Phone` optional
- `Description` optional
- `ProfileImageMediaID` **required** (non-nil after `Validate`)

## RBAC

- `struktur:read` — admin list + detail
- `struktur:write` — gates `/struktur/upload-media` and Create/Update/Delete

## Known Issues (Endpoint-specific)

- **Broken Update SQL.** `PUT /api/v1/struktur/{id}` and the `Service.Update` repo call both fail end-to-end:
  - Typo: `phone = $4 = $5` is a syntax error — PostgreSQL rejects the statement, every Update call fails at the database driver.
  - Arg count vs placeholder mismatch: 7 placeholders bind 8 Exec args (`Name, Position, Email, Phone, ProfileImageMediaID, Description, ID`) — `*sql.DB.ExecContext` errors with `bind count mismatch`.
  - Fix: `phone = $4, profile_image_media_id = $5, description = $6, updated_at = NOW() WHERE id = $7`, then bind 7 args. See `feature-docs/struktur/database.md` for full SQL block.
- **No regression test for Update on existing row.** Existing integration test `TestStruktur_UpdateNonExistent` exercises the `RowsAffected==0` branch and never trips the broken SQL. Add a test that calls Update on an actual row.

## Related

- `feature-docs/struktur/concept.md`
- `feature-docs/struktur/database.md`
- `interface/http/handler/struktur/struktur.go` — handler
- `usecase/struktur/service.go` — service
