# Struktur Organisasi Feature

Organization chart member management.

**Module:** `webdesa/api`
**Last updated:** 2026-07-04

## Files

| File | Purpose |
|---|---|
| `interface/http/handler/struktur/struktur.go` | `StrukturHandler` struct + methods (package `struktur`) |
| `interface/http/handler/struktur/route.go` | `RegisterRoutes` for `/struktur/*` + `/public/struktur/*` (package `struktur`) |
| `domain/struktur/struktur.go` | `Struktur` entity + `Validate()` |
| `usecase/struktur/service.go` | CRUD business logic |
| `interface/postgres/struktur.go` | sqlx implementation |
| `db/migrations/00011`, `00037` | table + `description` column |

## Database

### `struktur_organisasi` (migrations 00011, 00037)

| Column | Type | Notes |
|---|---|---|
| `id` | UUID PK | `gen_random_uuid()` |
| `name` | VARCHAR(255) NOT NULL | indexed `idx_struktur_name` |
| `position` | VARCHAR(255) | indexed `idx_struktur_position` |
| `email` | VARCHAR(255) | |
| `phone` | VARCHAR(50) | |
| `profile_image_url` | VARCHAR(500) | |
| `description` | TEXT | **added in 00037** (previously missing — see `recom.docs/bugs.md#3`) |
| `created_at` | TIMESTAMP | |
| `updated_at` | TIMESTAMP | |

## Domain Entity

`domain/struktur/struktur.go:16-26` — `Struktur{ID, Name, Position, Email, Phone, ProfileImageURL, Description, CreatedAt, UpdatedAt}`

### Validation Rules

- `Name` required (≤ 255 chars)
- Other fields optional but validated if present
- `Description` optional, no strict length limit

## Bug Fix History (`bugs.md#3`)

Previously the `Description` field existed in the domain entity with `db:"-"` (explicitly excluded from DB scan) but `Validate()` included it — so the field was **silently dropped** at scan time.

**Now**: column exists in DB (`db/migrations/00037_add_description_to_struktur.sql`), struct uses `db:"description"`, INSERT/UPDATE/SELECT queries all include it.

## Service

`usecase/struktur/service.go` — `Service{repo, fileHandler, clock}`

| Method | Purpose |
|---|---|
| `Create(CreateStrukturInput)` | Image optional |
| `List(ListStrukturInput)` | `Query`, `page`, `limit` |
| `GetByID`, `Delete` | Standard CRUD. Delete also removes profile image |
| `Update` | Replaces image if new file provided |

### Input

```go
type CreateStrukturInput struct {
    Name string
    Position, Email, Phone, Description *string
    ImageFile io.Reader
    ImageName, ContentType string
    ImageSize int64
}
```

## Endpoints

### Public (no auth)

| Method | Path |
|---|---|
| GET | `/api/v1/public/struktur/list` | Paginated |
| GET | `/api/v1/public/struktur/{id}` | |

### Admin (RBAC: `struktur:read`)

| Method | Path |
|---|---|
| GET | `/api/v1/struktur` | Query: `q`, `page`, `limit` |
| GET | `/api/v1/struktur/{id}` | |

### Admin (RBAC: `struktur:write`)

| Method | Path | Notes |
|---|---|---|
| POST | `/api/v1/struktur` | Multipart. Required: `name`. Optional: `position`, `email`, `phone`, `description`, `profile_image` |
| PUT | `/api/v1/struktur/{id}` | Multipart |
| DELETE | `/api/v1/struktur/{id}` | Deletes profile image too |

## RBAC Permissions

- `struktur:read`
- `struktur:write`

## Tests

**Integration** (`integration/struktur_test.go`): 7 cases
- Happy path
- No auth
- Empty name validation
- Not-found
- Invalid UUID
- Update non-existent
- Delete non-existent

## Related

- `interface/file/local_handler.go` — image storage adapter
- `pkg/handlerutil.ValidateStruct` — struct validation