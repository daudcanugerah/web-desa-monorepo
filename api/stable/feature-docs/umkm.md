# UMKM Feature

Small business directory management with category system and multi-image support.

**Module:** `webdesa/api`
**Last updated:** 2026-07-04

## Files

| File | Purpose |
|---|---|
| `interface/http/handler/umkm/umkm.go` | `UMKMHandler` struct + methods (package `umkm`) |
| `interface/http/handler/umkm/umkmcategory.go` | `UMKMCategoryHandler` |
| `interface/http/handler/umkm/route.go` | `RegisterRoutes` for `/umkm/*` + `/public/umkm/*` |
| `domain/umkm/umkm.go` | `UMKM` entity + `Validate()` |
| `domain/umkmcategory/umkmcategory.go` | `Category` entity |
| `usecase/umkm/service.go` | CRUD + image handling |
| `usecase/umkm/repository.go` | `Repository` interface |
| `usecase/umkm/category_lookup.go` | `CategoryLookup` interface |
| `usecase/umkmcategory/service.go` | Category CRUD + usage count |
| `interface/postgres/umkm.go` | sqlx implementation |
| `interface/postgres/umkmcategory.go` | sqlx implementation |
| `db/migrations/00016`, `00028`, `00030` | images, category table, FK conversion |

## Database

### `umkm` (migrations 00008, 00016, 00030)

| Column | Type | Notes |
|---|---|---|
| `id` | UUID PK | `gen_random_uuid()` |
| `name` | VARCHAR(255) NOT NULL | indexed `idx_umkm_name` |
| `owner` | VARCHAR(255) | |
| `address` | VARCHAR(255) | |
| `phone` | VARCHAR(50) | |
| `email` | VARCHAR(255) | |
| `website` | VARCHAR(255) | |
| `category` | UUID NOT NULL | FK → `umkm_categories(id)` ON DELETE RESTRICT |
| `description` | TEXT NOT NULL DEFAULT `''` | added 00016 |
| `images` | JSONB DEFAULT `'[]'` | array of URLs — added 00016 |
| `created_at` | TIMESTAMP | |
| `updated_at` | TIMESTAMP | |

### `umkm_categories` (migration 00028)

| Column | Type | Notes |
|---|---|---|
| `id` | UUID PK | |
| `name` | VARCHAR(100) NOT NULL | UNIQUE index |
| `created_at` | TIMESTAMP | |
| `updated_at` | TIMESTAMP | |

**Seeded**: `"Umum"` (placeholder for legacy rows).

## Domain Entity

`domain/umkm/umkm.go:15-28` — `UMKM{ID, Name, Category, Description, Owner, Address, Phone, Email, Website, Images []string, CreatedAt, UpdatedAt}`

### Validation Rules

| Field | Rule |
|---|---|
| `Name` | Required, ≤ 255 chars |
| `Category` | Required (UUID) |
| `Description` | Required |
| Optional fields | Non-empty if present |
| `Phone` | ≤ 50 chars |
| `Email` | Must contain `@` and `.` |
| `Website` | ≤ 255 chars (no URL format check) |

## Service

`usecase/umkm/service.go` — `Service{repo, fileHandler, clock, categoryRepo CategoryLookup}`

| Method | Purpose |
|---|---|
| `Create(CreateUMKMInput)` | Validates category; saves images (deletes prior saves on partial failure) |
| `Update(ctx, id, UpdateUMKMInput)` | Replaces all images when new files provided (deletes old) |
| `GetByID`, `List(ListUMKMInput)`, `Delete` | Standard CRUD. Delete also removes all images |

### Input

```go
type CreateUMKMInput struct {
    Name, Category, Description string
    Owner, Address, Phone, Email, Website *string
    Images []string      // existing URLs
    ImageFiles []ImageInput  // new uploads
}

type ImageInput struct {
    Filename, ContentType string
    Content io.Reader
    Size int64
}
```

`Images` and `ImageFiles` are mutually exclusive paths.

## Endpoints

### Public (no auth)

| Method | Path |
|---|---|
| GET | `/api/v1/public/umkm/list` | Query: `q`, `category` (UUID), `page`, `limit` |
| GET | `/api/v1/public/umkm/{id}` | |
| GET | `/api/v1/public/umkm/categories` | With `q` autocomplete |

### Admin (RBAC: `umkm:read`)

| Method | Path |
|---|---|
| GET | `/api/v1/umkm` | Query: `q`, `category`, `page`, `limit` |
| GET | `/api/v1/umkm/{id}` | |

### Admin (RBAC: `umkm:write`)

| Method | Path | Notes |
|---|---|---|
| POST | `/api/v1/umkm` | **Multipart only**. Required: `name`, `description`, `category` (UUID). Multiple images via `images[]` files |
| PUT | `/api/v1/umkm/{id}` | Multipart. Optional new image files |
| DELETE | `/api/v1/umkm/{id}` | Deletes all images |
| POST | `/api/v1/umkm/categories` | Body `{name}` |
| DELETE | `/api/v1/umkm/categories/{id}` | 409 if in use |

## Response Shape

```json
{
  "id": "uuid",
  "name": "string",
  "category": "uuid",
  "description": "string",
  "owner": "string|null",
  "address": "string|null",
  "phone": "string|null",
  "email": "string|null",
  "website": "string|null",
  "images": ["url1", "url2"],
  "created_at": "RFC3339",
  "updated_at": "RFC3339"
}
```

## Seed Data

`cmd/seed.go:254-260`:
```
"Kuliner", "Kerajinan", "Pertanian", "Jasa", "Perdagangan"
```

## RBAC Permissions

- `umkm:read`
- `umkm:write`

## Tests

**Integration**:
- `integration/umkm_test.go` (12 cases) — happy path, validation, list pagination, categories, image upload, description+images, timestamps
- `integration/umkmcategory_test.go` (9 cases) — list, autocomplete, in-use (409), success, usage count

## Related

- `interface/file/local_handler.go` — image storage
- `pkg/handlerutil.ValidateStruct` — struct validation