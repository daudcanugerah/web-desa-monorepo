# Profile Feature

CMS-style content sections for the village website.

**Module:** `webdesa/api`
**Last updated:** 2026-07-04

## Files

| File | Purpose |
|---|---|
| `interface/http/handler/profile/profile.go` | `ProfileHandler` struct + methods (package `profile`) |
| `interface/http/handler/profile/route.go` | `RegisterRoutes` for `/profile/*` + `/public/profile/*` (package `profile`) |
| `domain/profile/profile.go` | `Profile` entity + `Validate()` |
| `usecase/profile/service.go` | CRUD + section-name listing |
| `interface/postgres/profile.go` | sqlx implementation |

## Database

`profile` table (`db/migrations/00020`):

| Column | Type | Notes |
|---|---|---|
| `id` | UUID PK | `gen_random_uuid()` |
| `content` | TEXT NOT NULL | HTML/text content (≤ 10000 chars) |
| `section_name` | VARCHAR(100) NOT NULL | indexed `idx_profile_section_name` |
| `section_endpoint` | VARCHAR(255) NOT NULL | URL path (must start with `/`) |
| `state` | BOOLEAN NOT NULL DEFAULT `true` | indexed `idx_profile_state` |
| `created_at` | TIMESTAMP | |
| `updated_at` | TIMESTAMP | |

## Domain Entity

`domain/profile/profile.go:15-23` — `Profile{ID, Content, SectionName, SectionEndpoint, State, CreatedAt, UpdatedAt}`

### Validation Rules

| Field | Rule |
|---|---|
| `Content` | Required, ≤ 10000 chars |
| `SectionName` | Required, ≤ 100 chars |
| `SectionEndpoint` | Required, ≤ 255 chars, **must start with `/`** |
| `State` | Required |

## Service

`usecase/profile/service.go` — standard CRUD service.

| Method | Purpose |
|---|---|
| `Create(ctx, CreateProfileInput)` | |
| `GetByID(ctx, id)` | |
| `List(ListProfilesInput)` | Supports `section_name` and `state` filters, page, limit |
| `Update(ctx, id, UpdateProfileInput)` | |
| `Delete(ctx, id)` | |
| `GetSectionNames(ctx)` | Returns `SELECT DISTINCT section_name FROM profile ORDER BY section_name` |

## Endpoints

### Public (no auth)

| Method | Path |
|---|---|
| GET | `/api/v1/public/profile/list` | Paginated |
| GET | `/api/v1/public/profile/{id}` | |

### Admin (RBAC: `profile:read`)

| Method | Path |
|---|---|
| GET | `/api/v1/profile` | Query: `page`, `limit`, `section_name`, `state` |
| GET | `/api/v1/profile/{id}` | |
| GET | `/api/v1/profile/sections/names` | Distinct section names |

### Admin (RBAC: `profile:write`)

| Method | Path |
|---|---|
| POST | `/api/v1/profile` | Body `{content, section_name, section_endpoint, state}` |
| PUT | `/api/v1/profile/{id}` | |
| DELETE | `/api/v1/profile/{id}` | |

## Implementation Note

`ListProfile` handler does NOT use httpin — it parses query strings manually via `r.URL.Query().Get("page")` + `strconv.Atoi` (line 117-130). `UpdateProfile` uses httpin normally.

## RBAC Permissions

- `profile:read`
- `profile:write`

## Tests

**Integration** (`integration/profile_test.go`): 6 cases
- Happy path
- Sections names
- List with filters
- No-auth
- Empty content
- Not-found