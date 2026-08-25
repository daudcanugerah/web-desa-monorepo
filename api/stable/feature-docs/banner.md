# Banner Feature

Banner carousel management with active/inactive status.

**Module:** `webdesa/api`
**Last updated:** 2026-07-04

## Files

| File | Purpose |
|---|---|
| `interface/http/handler/banner/banner.go` | `BannerHandler` struct + methods (package `banner`) |
| `interface/http/handler/banner/route.go` | `RegisterRoutes` for `/banners/*` + `/public/banner/*` (package `banner`) |
| `domain/banner/banner.go` | `Banner` entity + `Validate()` |
| `usecase/banner/service.go` | Create, List, Get, Update, UpdateStatus, Delete |
| `interface/postgres/banner.go` | sqlx implementation |
| `db/migrations/00038_enforce_active_banner_limit.sql` | Trigger enforcing 20-active limit |

## Database

`banners` table (`db/migrations/00004`, `00019`):

| Column | Type | Notes |
|---|---|---|
| `id` | UUID PK | `gen_random_uuid()` |
| `title` | VARCHAR(255) NOT NULL | |
| `description` | TEXT DEFAULT `''` | |
| `image_url` | VARCHAR(500) NOT NULL | |
| `link` | VARCHAR(500) DEFAULT `''` | added in 00019 |
| `status` | VARCHAR(20) NOT NULL DEFAULT `'inactive'` | CHECK `IN ('active','inactive')` |
| `metadata` | JSONB DEFAULT `'{}'` | |
| `created_at` | TIMESTAMP | |
| `updated_at` | TIMESTAMP | |

Index: `idx_banners_status`.

## Domain Entity

`domain/banner/banner.go:15-25` — `Banner{ID, Title, Description, ImageURL, Link, Status, Metadata map[string]interface{}, CreatedAt, UpdatedAt}`

### Validation Rules

- `ID` required
- `Title` ≤ 255 chars
- `ImageURL` required, ≤ 500 chars
- `Status` must be `active` or `inactive`
- `CreatedAt`, `UpdatedAt` required

## Service

`usecase/banner/service.go` — `Service{repo, fileHandler, clock}`

| Method | Purpose |
|---|---|
| `Create(CreateBannerInput)` | Saves image first, validates, persists; defaults status to `inactive` |
| `Update(UpdateBannerInput)` | Status NOT updated here — use `UpdateStatus` |
| `UpdateStatus(ctx, id, status)` | **20-active limit enforced by DB trigger** (see below). |
| `GetByID`, `List(ListBannersInput)`, `Delete` | Standard CRUD |

### Active Banner Limit (DB Trigger)

`db/migrations/00038_enforce_active_banner_limit.sql` creates a `BEFORE INSERT OR UPDATE OF status` trigger on `banners` that runs `enforce_active_banner_limit()`. The function counts active banners and raises `check_violation` if count >= 20.

**Previously** (before this fix): the limit was enforced in `usecase/banner/service.go:UpdateStatus` via a SELECT + check + UPDATE pattern, which had a TOCTOU race condition between concurrent requests.

**Now**: atomic enforcement at the database level. Even with N concurrent requests trying to activate a 21st banner, only 20 can succeed.

## Endpoints

### Public (no auth)

| Method | Path |
|---|---|
| GET | `/api/v1/public/banner` | Paginated public list |
| GET | `/api/v1/public/banner/{id}` | |
| GET | `/api/v1/banners/active` | Up to 100 active banners |

### Admin (RBAC: `banners:read`)

| Method | Path |
|---|---|
| GET | `/api/v1/banners` | Query: `page`, `limit`, `status` |
| GET | `/api/v1/banners/{id}` | |

### Admin (RBAC: `banners:write`)

| Method | Path | Notes |
|---|---|---|
| POST | `/api/v1/banners` | Multipart: `title`, `description`, `link`, `image` (required), `metadata` (JSON string). Defaults status to `inactive`. |
| PUT | `/api/v1/banners/{id}` | Multipart; image optional. |
| PATCH | `/api/v1/banners/{id}/status` | Body `{status: "active"\|"inactive"}`. **DB trigger enforces 20-active limit.** |
| DELETE | `/api/v1/banners/{id}` | Deletes image file too |

## Response Shape

```json
{
  "id": "uuid",
  "title": "string",
  "description": "string",
  "link": "string",
  "image_url": "string",
  "status": "active|inactive",
  "metadata": {},
  "created_at": "RFC3339",
  "updated_at": "RFC3339"
}
```

## RBAC Permissions

- `banners:read`
- `banners:write`

## Side Effects

- Image cleanup on Create/Update/Delete failure paths
- Active count is enforced atomically by DB trigger (no race window)

## Tests

**Integration** (`integration/banner_test.go`): 40+ cases including full CRUD, metadata persistence, status updates, link round-tripping, `TestBannerLink_ActiveBannersIncludesLink`.

## Related

- `interface/file/local_handler.go` — image storage adapter (10 MB max)
- `pkg/handlerutil.InferContentType` — MIME type inference
- `db/migrations/00038_enforce_active_banner_limit.sql` — atomic enforcement trigger