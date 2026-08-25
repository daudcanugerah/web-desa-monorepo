# PPID — Endpoints

**Module:** `webdesa/api`
**Last updated:** 2026-08-23

## Routes

### Public (no auth)

| Method | Path | Notes |
|---|---|---|
| GET | `/api/v1/public/ppid/list` | Paginated, description truncated to 500 chars |
| GET | `/api/v1/public/ppid/{id}` | Full description |
| GET | `/api/v1/public/ppid/categories` | With `q` autocomplete |
| POST | `/api/v1/public/ppid/{id}/requests` | Body `{requester_name, requester_email, purpose}` |

> `GET /api/v1/ppid/document/{documentId}/download` was **removed** by Task 7.4. Citizens use `/api/v1/media/{id}/content?jwt=` instead.

### Admin (RBAC: `ppid:read`)

| Method | Path | Query |
|---|---|---|
| GET | `/api/v1/ppid` | `page`, `limit`, `category`, `q` |
| GET | `/api/v1/ppid/{id}` | Full `PPIDResponse` |
| GET | `/api/v1/ppid/requests` | `status` (repeatable), `page`, `limit` |

### Admin (RBAC: `ppid:write`)

| Method | Path | Notes |
|---|---|---|
| POST | `/ppid/upload-media` | Multipart `file` (PDF/DOC/DOCX/XLS/XLSX ≤50MB). Returns `{media_id, url}` |
| POST | `/ppid/upload-thumbnail` | Multipart `file` (image). Returns `{media_id, url}` |
| POST | `/ppid/upload` | Multipart `file` (alias/legacy form of `upload-media`) |
| POST | `/api/v1/ppid` | Multipart. Either `document` file OR `document_media_id`; same mutex for `thumbnail` / `thumbnail_media_id`. Other fields: `title` (req), `category` (UUID opt), `description`, `publication_at` (RFC3339) |
| PUT | `/api/v1/ppid/{id}` | Multipart. Mutex same as Create. Required: `title` |
| DELETE | `/api/v1/ppid/{id}` | |
| POST | `/api/v1/ppid/requests/{id}/approve` | |
| POST | `/api/v1/ppid/requests/{id}/revoke` | |
| POST | `/api/v1/ppid/categories` | Body `{name}` |
| DELETE | `/api/v1/ppid/categories/{id}` | 409 if in use |

## Request/Response

### Create PPID (admin)

```
POST /api/v1/ppid   (multipart/form-data)
```

| Form field | Type | Required | Notes |
|---|---|---|---|
| `title` | string | yes | ≤ 255 |
| `category` | UUID | no | must exist in `ppid_categories` |
| `description` | string | no | non-empty if present |
| `publication_at` | RFC3339 string | no | |
| `document` | file | mutex | raw upload → creates gallery media |
| `document_media_id` | UUID | mutex | references existing gallery media |
| `thumbnail` | file | mutex | raw upload → creates gallery media |
| `thumbnail_media_id` | UUID | mutex | references existing gallery media |

Exactly one of (`document` / `document_media_id`) must be supplied. Same for `thumbnail` / `thumbnail_media_id`.

**Response** — `PPIDResponse`:

```json
{
  "id": "uuid",
  "title": "...",
  "category_id": "uuid",
  "document": {"media_id": "uuid", "url": "/api/v1/media/{id}/content?jwt=..."},
  "thumbnail": {"media_id": "uuid", "url": "/api/v1/media/{id}/thumbnail?jwt=..."},
  "description": "...",
  "publication_at": "RFC3339",
  "created_at": "RFC3339",
  "updated_at": "RFC3339"
}
```

### Public list — `PPIDResponseTruncated`

```json
{
  "id": "uuid",
  "title": "...",
  "description": "...<truncated to 500 chars + '...'>",
  "thumbnail": {...}
}
```

`truncateDescription(*string) *string` clamps to 500 chars at last space, appends "...".

### Admin list — `PPIDResponsePrivate`

Same as truncated but **omits** the `description` field entirely.

### Detail — `PPIDResponse`

Full variant — see above. Returned by admin detail (`GET /api/v1/ppid/{id}`) and public detail (`GET /api/v1/public/ppid/{id}`).

### Create request (public)

```
POST /api/v1/public/ppid/{id}/requests
{
  "requester_name": "Budi",
  "requester_email": "budi@example.com",
  "purpose": "Untuk audit"
}
```

**Response**: `201 Created` + `PPIDRequest` echo with `status="pending"`.

### Approve request (admin)

```
POST /api/v1/ppid/requests/{id}/approve
```

Side effect: mints `signedURL.Sign(ScopePPID, *DocumentMediaID, "ppid_request:"+req.ID, 0)` → email sent with `?jwt=` link → status flips to `approved`. JWT TTL = `GALLERY_SIGNED_URL_PUBLIC_TTL` (default 1 hour).

### Revoke request (admin)

```
POST /api/v1/ppid/requests/{id}/revoke
```

Sets `revoked_at`, `revoked_by`, status=`revoked`. Pushes `sub=ppid_request:<id>` onto the gallery signed-URL deny-list — already-issued JWTs fail verification.

### Download (citizen)

```
GET /api/v1/media/{media_id}/content?jwt=<signed-token>
```

Handler is `gallery.SignedMediaHandler`. Verifies scope=`ppid` and that `sub` is not on the deny-list. **This is the only path** — Task 7.4 removed the legacy PPID-scoped download route.

## Validation

### Handler-level (multipart)

- `title` non-empty, ≤ 255 chars
- `category` must be valid UUID (if present)
- `publication_at` must parse as RFC3339 (if present)
- Each upload slot (document / thumbnail) accepts exactly one of `file` vs `media_id`

### Domain-level (`domain/ppid/ppid.go`)

- `DocumentMediaID` required (non-nil after `Validate`)
- `ThumbnailMediaID` optional
- `Description` non-empty if present
- Status invariants on `PPIDRequest`: `approved` requires both `ApprovedAt` + `ApprovedBy` set; `revoked` requires `RevokedAt` + `RevokedBy` set

## RBAC

- `ppid:read` — admin list + detail + request list
- `ppid:write` — admin create/update/delete + upload endpoints + approve/revoke + category CRUD

Legacy `policy.csv` (older second policy file, **inactive**) granted operator create/read/update/delete on PPID. Active policy is `rbac/rbac_policy.csv` + DB adapter (`rbac_policy` table).

## Known Issues (Endpoint-specific)

- **Removed download route** (`/api/v1/ppid/document/{documentId}/download`) — Task 7.4. Citizens must use `/api/v1/media/{id}/content?jwt=`.
- **`DownloadDocument` (`service.go:772`) is dead code.** No caller references the method in the active codebase. The dual-token logic (`ValidateAccessToken` / `ValidateUserToken`) is unreachable. Remove or repoint.
- **`GenerateAccessToken` (legacy HS256, `service.go:530`) is dead code.** Approval flow now uses `signedURL.Sign(ScopePPID, …)`. Still compiles because `SignedURLService` is nilable; bypassed when wired.
- **`GET /ppid/upload` is redundant with `/ppid/upload-media`.** Both call `PPIDUploadHandler.UploadDocument` → `fileStore.SaveDocument`. Kept for legacy form compatibility.

## Related

- `feature-docs/ppid/concept.md`
- `feature-docs/ppid/database.md`
- `interface/http/handler/ppid/ppid.go` — handler
- `usecase/ppid/service.go` — approval flow (`service.go:625-710`)
