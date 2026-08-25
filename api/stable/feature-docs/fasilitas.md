# Fasilitas (Facilities) Feature

Village facility management with geolocation and bounding-box queries.

**Module:** `webdesa/api`
**Last updated:** 2026-07-04

## Files

| File | Purpose |
|---|---|
| `interface/http/handler/fasilitas/fasilitas.go` | `FasilitasHandler` struct + methods (package `fasilitas`) |
| `interface/http/handler/fasilitas/fasilitascategory.go` | `FasilitasCategoryHandler` |
| `interface/http/handler/fasilitas/route.go` | `RegisterRoutes` for `/fasilitas/*` + `/public/fasilitas/*` |
| `domain/fasilitas/fasilitas.go` | `Fasilitas` entity + `Validate()` |
| `domain/fasilitascategory/fasilitascategory.go` | `Category` entity |
| `usecase/fasilitas/service.go` | CRUD + bbox queries |
| `usecase/fasilitas/repository.go` | `Repository` interface |
| `usecase/fasilitascategory/service.go` | Category CRUD |
| `interface/postgres/fasilitas.go` | sqlx implementation with bbox SQL |
| `interface/postgres/fasilitascategory.go` | sqlx implementation |
| `db/migrations/00015`, `00033`, `00034` | images, category table, FK conversion |

## Database

### `fasilitas` (migrations 00006, 00015, 00034)

| Column | Type | Notes |
|---|---|---|
| `id` | UUID PK | `gen_random_uuid()` |
| `name` | VARCHAR(255) | |
| `category` | UUID NOT NULL | FK → `fasilitas_categories(id)` ON DELETE RESTRICT (was `type VARCHAR(100)` until 00034) |
| `latitude` | DECIMAL(10,8) NOT NULL | range `[-90, 90]` |
| `longitude` | DECIMAL(11,8) NOT NULL | range `[-180, 180]` |
| `description` | TEXT | |
| `images` | JSONB DEFAULT `'[]'` | added 00015 |
| `created_at` | TIMESTAMP | |
| `updated_at` | TIMESTAMP | |

Index: `idx_fasilitas_location` (latitude, longitude).

### `fasilitas_categories` (migration 00033)

Standard category table. **Seeded**: `"Lainnya"`.

## Domain Entity

`domain/fasilitas/fasilitas.go:15-25` — `Fasilitas{ID, Name, Category, Latitude, Longitude, Description, Images []string, CreatedAt, UpdatedAt}`

### Validation Rules

| Field | Rule |
|---|---|
| `Name` | Required |
| `Latitude` | `[-90, 90]` |
| `Longitude` | `[-180, 180]` |
| Image URLs | ≤ 500 chars each |

## Service

`usecase/fasilitas/service.go` — `Service{repo, fileHandler, categoryRepo CategoryLookup, clock}`

| Method | Purpose |
|---|---|
| `Create(CreateFasilitasInput)` | Validates coordinates, optional category FK existence check |
| `Update(ctx, id, UpdateFasilitasInput)` | Deletes old images only when new files uploaded |
| `GetByID`, `Delete` | Standard CRUD |
| `List(ListFasilitasInput)` | Supports `BBox *BoundingBox`, `page`, `limit` |
| `RemoveImage(ctx, id, imageIndex int)` | Slices the images array |

### Bounding Box

```go
type BoundingBox struct {
    MinLon, MinLat, MaxLon, MaxLat float64
}
```

Validation: `MinLat < MaxLat`, `MinLon < MaxLon`, all within valid ranges.

### Repository SQL (BBox Query)

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

## Endpoints

### Public (no auth)

| Method | Path |
|---|---|
| GET | `/api/v1/public/fasilitas/list` | Query: bbox + page + limit |
| GET | `/api/v1/public/fasilitas/{id}` | |
| GET | `/api/v1/public/fasilitas/categories` | |

### Admin (RBAC: `fasilitas:read`)

| Method | Path | Query |
|---|---|---|
| GET | `/api/v1/fasilitas` | bbox + page + limit |
| GET | `/api/v1/fasilitas/{id}` | |

### Admin (RBAC: `fasilitas:write`)

| Method | Path | Notes |
|---|---|---|
| POST | `/api/v1/fasilitas` | Multipart. Required: `name`, `latitude`, `longitude`. Optional: `category` (UUID), `description`, `images[]` |
| PUT | `/api/v1/fasilitas/{id}` | Multipart |
| DELETE | `/api/v1/fasilitas/{id}` | |
| DELETE | `/api/v1/fasilitas/{id}/images/{imageIndex}` | Remove one image by zero-based index |
| POST | `/api/v1/fasilitas/categories` | Body `{name}` |
| DELETE | `/api/v1/fasilitas/categories/{id}` | 409 if in use |

### Bounding Box Query Params

All four required if any provided (line 222 of handler):
```
GET /api/v1/fasilitas?minLat=...&maxLat=...&minLon=...&maxLon=...&page=...&limit=...
```

## Response Shape

```json
{
  "id": "uuid",
  "name": "string",
  "type": "uuid",         // ← JSON field name is "type" for API stability
  "latitude": -6.12345678,
  "longitude": 106.12345678,
  "description": "string|null",
  "images": ["url1"],
  "created_at": "RFC3339",
  "updated_at": "RFC3339"
}
```

> The JSON field for category is `"type"` (Go field `Category *string` mapped via `json:"type"`) for API stability with older versions.

## Seed Data

`cmd/seed.go:275-282`:
```
"Pendidikan", "Kesehatan", "Ibadah", "Olahraga", "Pemerintahan", "Pasar"
```

## RBAC Permissions

- `fasilitas:read`
- `fasilitas:write`

## Tests

**Integration**:
- `integration/fasilitas_test.go` (13 cases) — happy path, coord validation, list pagination, **bounding-box query not yet covered** ⚠️, images array, empty images, timestamps
- `integration/fasilitascategory_test.go` (8 cases) — list, autocomplete, CRUD

## Known Gaps

- **No bounding-box integration test coverage** — add tests for the geospatial query (see `recom.docs/testing.md`).

## Related

- `interface/file/local_handler.go` — image storage
- `pkg/handlerutil.ValidateStruct` — struct validation
- `usecase/fasilitas/service.go` — bbox validation logic