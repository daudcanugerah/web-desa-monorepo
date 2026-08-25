# Fasilitas — Concept

Village facility management with geolocation, bounding-box queries, and gallery-backed media. Admins pin facilities with lat/lon and category; clients query by area or list all.

**Module:** `webdesa/api`
**Last updated:** 2026-08-23

## Purpose

Provide a facility directory with optional category, exact coordinates, multi-image gallery, and geospatial bounding-box filtering for map UIs.

## Business Rules

- **Category is nullable** at both schema and domain layer (`*string`, FK ON DELETE RESTRICT). Facilities without a category are allowed.
- **Coordinates required.** `latitude ∈ [-90, 90]`, `longitude ∈ [-180, 180]`, stored as `DECIMAL(10,8)` / `DECIMAL(11,8)`.
- **Multi-image via `images_media_ids`.** Same model as UMKM — gallery UUIDs attached post-upload, not inline uploads. Legacy `images` JSONB column dropped in `00045`.
- **Bulk media caps:** 20 files per record → 413, 250 MB total → 413.
- **`OnMediaDeleted` sweep.** Repository implements `OnMediaDeleted(ctx, mediaID)`; gallery hub invokes it. Reads also prune dangling UUIDs.
- **Bounding-box struct field order is `MinLon, MinLat, MaxLon, MaxLat`.** Service validates `MinLat < MaxLat` and `MinLon < MaxLon`; all four must be within valid coordinate ranges.
- **BBox is required all-or-none.** Handler (lines 284-288) rejects requests where only some of the four params are supplied.
- **Per-image UUID cap is 64 chars** (UUID length, not the legacy 500-char URL cap).
- **`RemoveImage(ctx, id, imageIndex)`** slices `ImagesMediaIDs` by zero-based index and deletes gallery media at that index.
- **Critical response field rename:** response emits `category_id` (nullable UUID) + `category` ({id, name} object). Legacy `"type"` UUID field is gone — clients reading `response.type` will break.
- **Public list has no bbox** (anonymous viewers don't get geospatial filtering).

## Architecture

### Files

| File | Purpose |
|---|---|
| `interface/http/handler/fasilitas/fasilitas.go` | `FasilitasHandler` struct + methods (package `fasilitas`) |
| `interface/http/handler/fasilitas/fasilitascategory.go` | `FasilitasCategoryHandler` |
| `interface/http/handler/fasilitas/route.go` | `RegisterRoutes` for `/fasilitas/*` + `/public/fasilitas/*` |
| `domain/fasilitas/fasilitas.go` | `Fasilitas` entity + `Validate()` |
| `domain/fasilitascategory/fasilitascategory.go` | `Category` entity |
| `usecase/fasilitas/service.go` | CRUD + bbox queries + bulk media |
| `usecase/fasilitas/repository.go` | `Repository` interface; implements `OnMediaDeleted(ctx, mediaID)` |
| `usecase/fasilitascategory/service.go` | Category CRUD |
| `interface/postgres/fasilitas.go` | sqlx implementation with bbox SQL |
| `interface/postgres/fasilitascategory.go` | sqlx implementation |
| `db/migrations/00006`, `00015`, `00033`, `00034`, `00045` | schema + FK + drops `images` JSONB |

### Service Struct Dependencies

```go
type Service struct {
    repo         Repository
    fileStore    galleryUsecase.FileStore
    categoryRepo CategoryLookup
    limits       BulkLimits
    clock        func() time.Time
}
```

### Repository Interface Methods

```go
type Repository interface {
    Create(ctx context.Context, f Fasilitas) error
    GetByID(ctx context.Context, id string) (*Fasilitas, error)
    List(ctx context.Context, in ListFasilitasInput) ([]Fasilitas, int, error)
    Update(ctx context.Context, f Fasilitas) error
    Delete(ctx context.Context, id string) error
    OnMediaDeleted(ctx context.Context, mediaID string) error
}
```

### Bounding Box

```go
type BoundingBox struct {
    MinLon, MinLat, MaxLon, MaxLat float64
}
```

Handler binds query params → sets `MinLat, MaxLat, MinLon, MaxLon`. Service validates `MinLat < MaxLat` and `MinLon < MaxLon`, all within valid coordinate ranges.

## Glossary

- **BBox** — bounding box defined by `minLat`, `maxLat`, `minLon`, `maxLon`. All-or-none on admin list endpoint.
- **`category_id` / `category`** — current response fields. Legacy `type` was removed.
- **`OnMediaDeleted`** — repo-side listener method invoked by gallery hub.
- **`images_media_ids`** — `UUID[]` column; gallery UUIDs.

## Related

- [Database](./database.md)
- [Endpoints](./endpoint.md)
- `interface/file/local_handler.go` — image storage
- `usecase/gallery` — `FileStore` + media-event hub
- `pkg/handlerutil.ValidateStruct` — struct validation
- `usecase/fasilitas/service.go` — bbox validation logic
- `db/migrations/00045_drop_legacy_upload_columns.sql` — drops `fasilitas.images` JSONB
