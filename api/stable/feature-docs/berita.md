# Berita (News) Feature

News/article management with Quill editor and category system.

**Module:** `webdesa/api`
**Last updated:** 2026-07-04

## Files

| File | Purpose |
|---|---|
| `interface/http/handler/berita/berita.go` | `BeritaHandler` struct + methods (package `berita`) |
| `interface/http/handler/berita/berita_upload.go` | `BeritaUploadHandler` — Quill media upload |
| `interface/http/handler/berita/beritacategory.go` | `BeritaCategoryHandler` |
| `interface/http/handler/berita/route.go` | `RegisterRoutes` for `/berita/*` + `/public/berita/*` |
| `domain/berita/berita.go` | `Berita` entity + `Validate()` |
| `domain/beritacategory/beritacategory.go` | `Category` entity (no Update op — delete and re-create to rename) |
| `usecase/berita/service.go` | CRUD + image processing |
| `usecase/berita/file_handler.go` | `FileHandler` interface |
| `usecase/berita/repository.go` | `Repository` interface |
| `usecase/berita/category_lookup.go` | `CategoryLookup` interface |
| `usecase/beritacategory/service.go` | Category CRUD + usage count |
| `interface/postgres/berita.go` | sqlx implementation |
| `interface/postgres/beritacategory.go` | sqlx implementation |
| `db/migrations/00027`, `00029`, `00018` | berita_categories table + FK conversion |

## Database

### `berita` (migrations 00007, 00018, 00029)

| Column | Type | Notes |
|---|---|---|
| `id` | UUID PK | `gen_random_uuid()` |
| `title` | VARCHAR(255) NOT NULL | |
| `content` | TEXT NOT NULL | Quill delta JSON |
| `image_url` | VARCHAR(500) | nullable |
| `category` | UUID NOT NULL | FK → `berita_categories(id)` ON DELETE RESTRICT |
| `created_at` | TIMESTAMP | indexed DESC |
| `updated_at` | TIMESTAMP | |

### `berita_categories` (migration 00027)

| Column | Type | Notes |
|---|---|---|
| `id` | UUID PK | |
| `name` | VARCHAR(100) NOT NULL | UNIQUE index |
| `created_at` | TIMESTAMP | |
| `updated_at` | TIMESTAMP | |

**Seeded**: `"Lainnya"` (placeholder for legacy rows).

## Domain Entities

### `Berita`
```go
type Berita struct {
    ID, Title, Content, Category string
    ImageURL *string
    CreatedAt, UpdatedAt time.Time
}
```

### `Category`
```go
type Category struct {
    ID, Name string
    CreatedAt, UpdatedAt time.Time
}
```
**No `Update` operation** — comment: *"intentionally no Update operation; to rename, delete and re-create"*.

## Services

### Berita Service

`usecase/berita/service.go` — `Service{repo, fileHandler, clock, uploadsDir, categoryRepo CategoryLookup}`

| Method | Purpose |
|---|---|
| `Create(CreateBeritaInput)` | Validates category via `categoryRepo.FindByID`, saves image, calls `processDeltaImages`, persists. Cleans up image on failure. |
| `Update(ctx, id, UpdateBeritaInput)` | Same pattern, deletes old image after successful new save |
| `GetByID`, `List(ListBeritaInput)`, `Delete` | Standard CRUD |

#### Image Processing — `processDeltaImages`

Walks Quill delta JSON `ops[]` looking for `insert.image` and `insert.video` URLs, moves files from `uploads/tmp/` to `uploads/` permanent storage at save time.

### Category Service

`usecase/beritacategory/service.go` — `Service{repo, clock}`

| Method | Purpose |
|---|---|
| `Create(name)` | Errors: `DuplicateNameError{Name}` |
| `Delete(id)` | Checks `FindByID`, then `repo.CountByCategoryIDs`, returns `*CategoryInUseError` if >0 |
| `List(query, page, limit)` | Returns `*CategoryWithCount` enriched with `UsageCount` |
| `FindByID` | |

## Endpoints

### Public (no auth)

| Method | Path |
|---|---|
| GET | `/api/v1/public/berita/list` | Query: `page`, `limit` |
| GET | `/api/v1/public/berita/{id}` | |
| GET | `/api/v1/public/berita/categories` | With `q` search for autocomplete |

### Admin (RBAC: `berita:read`)

| Method | Path |
|---|---|
| GET | `/api/v1/berita` | Query: `q`, `category` (UUID), `since`, `until` (YYYY-MM-DD), `page`, `limit` |
| GET | `/api/v1/berita/{id}` | |

### Admin (RBAC: `berita:write`)

| Method | Path | Notes |
|---|---|---|
| POST | `/api/v1/berita` | Multipart: `title` (required), `content` (Quill delta JSON, required), `category` (UUID required), optional `image` |
| PUT | `/api/v1/berita/{id}` | Multipart. Missing fields fall back to existing values |
| DELETE | `/api/v1/berita/{id}` | Deletes image file too |
| POST | `/api/v1/berita/upload-media` | Quill media upload (see below) |
| POST | `/api/v1/berita/categories` | Body `{name}` |
| DELETE | `/api/v1/berita/categories/{id}` | 409 if in use |

## Quill Media Upload

`POST /api/v1/berita/upload-media` (`interface/http/handler/berita/berita_upload.go`):

- Multipart form, max 50 MB total
- Field `type=image|video` (default `image`)
- Image max 10 MB; Video max 50 MB
- Allowed MIME: `image/jpeg|png|gif|webp`; `video/mp4|webm|ogg|quicktime`
- Filename: `tmp-yyyymmdd-<uuid>.<ext>`
- Stored directly in `uploadsDir` (no subdir)
- Returned URL: bare filename, e.g., `"tmp-20260317-abc.jpg"`

The `tmp-` prefix signals `processDeltaImages` to move the file on save.

## Business Logic Highlights

- Search filter: `ILIKE '%q%'` on title and content
- TTL of `tmp/` images: not enforced — persist until next save
- Default ordering: `created_at DESC`

## Seed Data

`cmd/seed.go:245-249`:
```
"Berita Desa", "Pengumuman", "Kegiatan"
```

## RBAC Permissions

- `berita:read`
- `berita:write`

## Tests

**Integration**:
- `integration/berita_test.go` (14 cases) — happy path, auth, validation, categories, list vs detail
- `integration/berita_upload_test.go` (11 cases) — image/video upload, MIME validation, tmp dir
- `integration/beritacategory_test.go` (12 cases) — list, autocomplete, pagination, create, in-use (409)

## Related

- `interface/file/local_handler.go` — image storage
- `pkg/handlerutil.IsValidUUID`, `ValidateStruct` — input validation
- `pkg/pagination.Paginate` — pagination normalization