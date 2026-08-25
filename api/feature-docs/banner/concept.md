# Banner — Concept

Banner carousel management with active/inactive status, gallery-backed media, and a category system. Active banners surface via a public stream; admins manage the pool via status toggles that are atomically capped by a DB trigger.

**Module:** `webdesa/api`
**Last updated:** 2026-08-23

## Purpose

Provide an admin-curated banner carousel with bounded active capacity, category organization, and a single image per banner drawn from the gallery.

## Business Rules

- **Single image per banner.** `image` (multipart file) and `image_media_id` (UUID) are mutex — exactly one must be supplied on Create/Update.
- **Image is required.** `image_media_id` is a non-empty UUID; no legacy URL string is accepted (replaces the dropped `image_url` column).
- **Status default is `inactive`.** Create does not auto-activate; admins must explicitly PATCH.
- **20-active hard cap, enforced at the database level** via a `BEFORE INSERT OR UPDATE OF status` trigger (`enforce_active_banner_limit()`). The legacy service-side SELECT+UPDATE pattern was removed to eliminate the TOCTOU race window.
- **Status is updated only via `PATCH /banners/{id}/status`**, never via the multipart PUT.
- **Category lookup is FK-validated** via `CategoryLookup.FindByID` on every Create/Update.
- **Delete cleans up media.** The service calls `fileStore.Delete(ctx, *b.ImageMediaID)` after the row delete.
- **Public endpoints rewrite media URLs** to feature-scoped public stream URLs (`/api/v1/public/banner/media/{id}/...`) so anonymous visitors can render system-folder media without gallery auth.
- **`GetActiveBanners` is hard-coded to a 100-row ceiling**, ordered `created_at DESC`.

## Architecture

### Files

| File | Purpose |
|---|---|
| `interface/http/handler/banner/banner.go` | `BannerHandler` struct + methods (package `banner`) |
| `interface/http/handler/banner/bannercategory.go` | `BannerCategoryHandler` |
| `interface/http/handler/banner/route.go` | `RegisterRoutes` for `/banners/*` + `/public/banner/*` |
| `domain/banner/banner.go` | `Banner` entity + `Validate()` |
| `domain/bannercategory/bannercategory.go` | `Category` entity |
| `usecase/banner/service.go` | CRUD + status toggle |
| `usecase/banner/repository.go` | `Repository` interface |
| `usecase/banner/category_lookup.go` | `CategoryLookup` interface |
| `domain/bannercategory/service.go` (or `usecase/bannercategory/`) | Category CRUD + usage count |
| `interface/postgres/banner.go` | sqlx implementation |
| `interface/postgres/bannercategory.go` | sqlx implementation |
| `db/migrations/00039_banner_categories.sql` | categories table + `banners.category` FK |
| `db/migrations/00045_drop_legacy_upload_columns.sql` | drops `banners.image_url` |
| `db/migrations/00038_enforce_active_banner_limit.sql` | atomic 20-active trigger |

### Service Struct Dependencies

```go
type Service struct {
    repo       Repository
    fileStore  galleryUsecase.FileStore
    categories CategoryLookup
    clock      func() time.Time
}
```

### Repository Interface Methods

```go
type Repository interface {
    Create(ctx context.Context, b Banner) error
    GetByID(ctx context.Context, id string) (*Banner, error)
    List(ctx context.Context, in ListBannersInput) ([]Banner, int, error)
    Update(ctx context.Context, b Banner) error
    UpdateStatus(ctx context.Context, id, status string) error
    Delete(ctx context.Context, id string) error
    GetActiveBanners(ctx context.Context) ([]Banner, error)
}
```

## Glossary

- **Active banner** — row with `status='active'`. Counted by the 20-cap trigger.
- **Category** — FK target on `banners.category`; managed via `/banners/categories` endpoints.
- **Media** — gallery-stored image referenced by `image_media_id`. Single image per banner.
- **Public stream URL** — `/api/v1/public/banner/media/{id}/...` URL form used by anonymous visitors.

## Related

- [Database](./database.md)
- [Endpoints](./endpoint.md)
- `interface/http/handler/file/local_handler.go` — gallery file storage adapter
- `usecase/gallery` — `FileStore` + `Media` events
- `pkg/handlerutil.InferContentType` — MIME type inference
- `db/migrations/00038_enforce_active_banner_limit.sql` — atomic enforcement trigger
- `db/migrations/00045_drop_legacy_upload_columns.sql` — drops `banners.image_url`
