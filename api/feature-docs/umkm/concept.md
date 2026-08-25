# UMKM — Concept

Small business directory management with category system and multi-image support, all media gallery-backed. Admins attach pre-uploaded gallery media IDs to each business; clients no longer upload images inline.

**Module:** `webdesa/api`
**Last updated:** 2026-08-23

## Purpose

Provide a curated small-business directory with categorized listings, contact details, and up to 20 gallery images per business, served both admin-side and publicly.

## Business Rules

- **Category is required** and FK-validated via `categoryRepo.FindByID` on every Create/Update.
- **Multi-image via `images_media_ids`, not inline uploads.** Clients first `POST /umkm/upload-media`, then attach returned `media_id`s via `images_media_ids` (repeatable multipart or JSON array). Inline file uploads were removed.
- **Bulk media caps (enforced via `respondIfBulkLimit`):**
  - 20 files per record → 413 `ErrBulkTooManyFiles`
  - 250 MB total → 413 `ErrBulkTotalTooLarge`
- **`OnMediaDeleted` sweep.** The repository implements `OnMediaDeleted(ctx, mediaID)`; `galleryService.Hub().AddListener(umkmRepo)` registers it. When gallery media is deleted, the repo sweeps any row referencing that media and updates the row to drop the now-dangling id.
- **Reads prune dangling UUIDs.** `GetByID` / `List` reconcile `images_media_ids` against the gallery so deleted media never surfaces in API output.
- **Delete cleans up all referenced media.** Service walks `ImagesMediaIDs` and calls `fileStore.Delete` for each after the row delete.
- **Default placeholder category is `"Umum"`** (seeded by `00028`), NOT `"Lainnya"`. `"Lainnya"` is for other features.
- **Handler validates first, then domain.** Email format check is handler-side (`omitempty,email`); domain `Validate()` enforces content rules.
- **Service cleans up uploaded files on partial failure** during Create.

## Architecture

### Files

| File | Purpose |
|---|---|
| `interface/http/handler/umkm/umkm.go` | `UMKMHandler` struct + methods (package `umkm`) |
| `interface/http/handler/umkm/umkmcategory.go` | `UMKMCategoryHandler` |
| `interface/http/handler/umkm/route.go` | `RegisterRoutes` for `/umkm/*` + `/public/umkm/*` |
| `domain/umkm/umkm.go` | `UMKM` entity + `Validate()` |
| `domain/umkmcategory/umkmcategory.go` | `Category` entity |
| `usecase/umkm/service.go` | CRUD + bulk-media handling |
| `usecase/umkm/repository.go` | `Repository` interface; implements `OnMediaDeleted(ctx, mediaID)` |
| `usecase/umkm/category_lookup.go` | `CategoryLookup` interface |
| `usecase/umkmcategory/service.go` | Category CRUD + usage count |
| `interface/postgres/umkm.go` | sqlx implementation |
| `interface/postgres/umkmcategory.go` | sqlx implementation |
| `db/migrations/00028`, `00030`, `00045` | categories table, FK conversion, drops `images` JSONB |

### Service Struct Dependencies

```go
type Service struct {
    repo        Repository
    fileStore   galleryUsecase.FileStore
    limits      BulkLimits
    clock       func() time.Time
    categoryRepo CategoryLookup
}
```

### Repository Interface Methods

```go
type Repository interface {
    Create(ctx context.Context, b UMKM) error
    GetByID(ctx context.Context, id string) (*UMKM, error)
    List(ctx context.Context, in ListUMKMInput) ([]UMKM, int, error)
    Update(ctx context.Context, b UMKM) error
    Delete(ctx context.Context, id string) error
    OnMediaDeleted(ctx context.Context, mediaID string) error
}

type CategoryLookup interface {
    FindByID(ctx context.Context, id string) (*Category, error)
}
```

## Glossary

- **`images_media_ids`** — `UUID[]` column replacing the dropped `images` JSONB; gallery UUIDs attached to a row.
- **`BulkLimits`** — config struct with max-files + max-total-bytes caps.
- **`OnMediaDeleted`** — repo-side listener method invoked by gallery hub when media is deleted.
- **`Umum`** — UMKM-specific placeholder category seeded by `00028`.

## Related

- [Database](./database.md)
- [Endpoints](./endpoint.md)
- `interface/file/local_handler.go` — image storage
- `usecase/gallery` — `FileStore` + media-event hub
- `pkg/handlerutil.ValidateStruct` — struct validation (handler-level email check)
- `db/migrations/00045_drop_legacy_upload_columns.sql` — drops `umkm.images` JSONB
