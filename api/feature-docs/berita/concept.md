# Berita — Concept

News/article management with Quill editor, gallery-backed media, and a category system. Articles store a Quill delta JSON in `content` and a single cover image via the gallery. Public readers see rewritten media URLs.

**Module:** `webdesa/api`
**Last updated:** 2026-08-23

## Purpose

Provide admins a Quill-authored news feed with cover images, category tagging, search, sort, and date filters, served both admin-side and publicly.

## Business Rules

- **Single cover image, mutex upload.** `image` (multipart file) and `image_media_id` (UUID) are mutex — at most one supplied on Create/Update.
- **Cover image is optional.** `image_media_id` is nullable; not setting it leaves the row with no image.
- **Category is required** and FK-validated via `categoryRepo.FindByID` on every Create/Update. Category is `NOT NULL` at the column level.
- **Quill media upload is image-only.** `POST /berita/upload-media` accepts `image/jpeg|png|gif|webp` only.
- **Quill upload filename = `media_id` (UUID).** No `tmp-yyyymmdd-` prefix; files are not staged in `uploads/tmp/`. The legacy `processDeltaImages` move-on-save flow was deleted.
- **`sort` and `order` are whitelisted.** `sort ∈ {created_at, title}`, `order ∈ {asc, desc}`. Validation rejects any other value.
- **Default ordering is `created_at DESC`.**
- **`since` / `until` filters** accept `YYYY-MM-DD`.
- **`GetBeritaPublic` rewrites embedded media paths** in content from `/api/v1/gallery/media/` → `/api/v1/public/berita/media/` so anonymous visitors can render embedded images.
- **Public list endpoint omits `content`** — only the detail endpoint returns it.
- **`DeleteBerita` returns 200 + message** (not 204).
- **Categories cannot be renamed.** No `Update` op exists — comment: *"intentionally no Update operation; to rename, delete and re-create"*.
- **Category deletion is in-use-guarded** — returns 409 if any article references the category.
- **Service retains `uploadsDir` field** for handler signature compatibility but no longer reads it after the `tmp-` flow was removed.

## Architecture

### Files

| File | Purpose |
|---|---|
| `interface/http/handler/berita/berita.go` | `BeritaHandler` struct + methods (package `berita`) |
| `interface/http/handler/berita/berita_upload.go` | `BeritaUploadHandler` — Quill media upload |
| `interface/http/handler/berita/beritacategory.go` | `BeritaCategoryHandler` |
| `interface/http/handler/berita/route.go` | `RegisterRoutes` for `/berita/*` + `/public/berita/*` |
| `domain/berita/berita.go` | `Berita` entity + `Validate()` |
| `domain/beritacategory/beritacategory.go` | `Category` entity (no Update op — delete and re-create to rename) |
| `usecase/berita/service.go` | CRUD; legacy `uploadsDir` field kept for signature compat |
| `usecase/berita/repository.go` | `Repository` interface |
| `usecase/berita/category_lookup.go` | `CategoryLookup` interface |
| `usecase/beritacategory/service.go` | Category CRUD + usage count |
| `interface/postgres/berita.go` | sqlx implementation |
| `interface/postgres/beritacategory.go` | sqlx implementation |
| `db/migrations/00007`, `00018`, `00027`, `00029`, `00045` | schema + FK conversion + drops `image_url` |

### Service Struct Dependencies

```go
type Service struct {
    repo        Repository
    fileStore   galleryUsecase.FileStore
    clock       func() time.Time
    uploadsDir  string                  // legacy; kept for signature compat only
    categoryRepo CategoryLookup
}
```

### Repository Interface Methods

```go
type Repository interface {
    Create(ctx context.Context, b Berita) error
    GetByID(ctx context.Context, id string) (*Berita, error)
    List(ctx context.Context, in ListBeritaInput) ([]Berita, int, error)
    Update(ctx context.Context, b Berita) error
    Delete(ctx context.Context, id string) error
}

type CategoryLookup interface {
    FindByID(ctx context.Context, id string) (*Category, error)
}
```

## Glossary

- **Quill delta** — JSON document format produced by the Quill rich-text editor; stored verbatim in `berita.content`.
- **Cover image** — single gallery media referenced by `image_media_id`.
- **Public content rewrite** — `GetBeritaPublic` substitutes gallery media paths with feature-scoped public stream URLs in returned `content`.
- **Whitelisted sort/order** — only `created_at`/`title` for sort, `asc`/`desc` for order.

## Related

- [Database](./database.md)
- [Endpoints](./endpoint.md)
- `interface/file/local_handler.go` — image storage
- `usecase/gallery` — `FileStore`
- `pkg/handlerutil.IsValidUUID`, `ValidateStruct` — input validation
- `pkg/pagination.Paginate` — pagination normalization
- `db/migrations/00045_drop_legacy_upload_columns.sql` — drops `berita.image_url`
