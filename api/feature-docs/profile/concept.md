# Profile — Concept

CMS-style content sections for the village website. Single table, simple CRUD plus a public read path. `state` flag exists but is not enforced anywhere — admin-disabled sections leak to the public site.

**Module:** webdesa/api · **Last updated:** 2026-08-23

## Purpose

- Manage named content sections (e.g. "Profil Desa", "Visi Misi") with HTML/text content and a `section_endpoint` URL path.
- Expose a public read API for the village website frontend.
- Filter listings by section name, state, and a free-text `q` query (ILIKE on `section_name` OR `content`).
- Surface a distinct list of section names for UI navigation.

## Business Rules

### Single table — `profile`

Only one migration touches `profile` (`00020_create_profile_table.sql`). No subsequent migrations add or drop columns.

### `state` is not validated

- No `validate` struct tag on `CreateProfileInput.State` / `UpdateProfileInput.State`.
- No check in `Profile.Validate()`.
- Omitting the field silently yields `false`.
- The DB column default of `true` is **dead**: the repo binds `$5` always, so the default never fires for new rows.

### Public endpoints never filter by `state`

`GET /api/v1/public/profile/list` and `GET /api/v1/public/profile/{id}` do **not** filter on `state`. Admin-disabled sections leak to the public site. The public handler ignores `state` entirely — no parameter is parsed, no service filter is set.

### `{id}` path param is not UUID-validated

Unlike `user.GetUser` / `DeleteUser` / `UpdateUser` which call `handlerutil.IsValidUUID`, profile handlers do not. A malformed id surfaces as a generic 500 from `sql: invalid UUID` rather than 400.

### List filters

`List(ctx, ListProfilesInput)`:
- `SectionName *string` — exact match on `section_name`
- `State *bool` — pointer, absent = no filter
- `Query *string` — ILIKE `%q%` on `section_name` OR `content` (undocumented in OpenAPI but used by the admin UI — see [endpoint.md](./endpoint.md#known-issues))
- `Page`, `Limit` — pagination

### Section name listing

`GetSectionNames(ctx)` returns `SELECT DISTINCT section_name FROM profile ORDER BY section_name`. Used for UI navigation.

### Response envelope — singular key

List response uses `data.profile` (singular) — not `profiles` or `items`. Section names response uses `data.section_names`.

### Validation rules

| Field | Rule |
|---|---|
| `Content` | Required, 1–10000 chars |
| `SectionName` | Required, 1–100 chars |
| `SectionEndpoint` | Required, 1–255 chars, **must start with `/`** |
| `State` | **No validation** — see Known Issues |

### Implementation note — manual parse vs httpin

The admin `ListProfile` handler (`profile.go:174-196`) parses `page`, `limit`, `section_name`, `state`, `q` **manually** via `r.URL.Query().Get(...)` + `strconv.Atoi` rather than using httpin:
- `limit > 100` does **not** return 400 — manual `strconv.Atoi` parses any positive int, and `pagination.Paginate` surfaces `pagination.MaxLimit` errors as a generic 500.
- Bad/non-numeric `page` / `limit` are silently ignored (fall back to defaults) instead of 400.

`ListProfilePublic` uses httpin with defaults and validates `Page > 0` / `Limit > 0` explicitly (returns 400 on bad input).

`UpdateProfile` uses httpin normally.

## Architecture

### Files

| File | Purpose |
|---|---|
| `interface/http/handler/profile/profile.go` | `ProfileHandler` struct + methods (package `profile`) |
| `interface/http/handler/profile/route.go` | `RegisterRoutes` for `/profile/*` + `/public/profile/*` |
| `domain/profile/profile.go` | `Profile` entity + `Validate()` |
| `usecase/profile/service.go` | `Service{repo, clock}` + `CreateProfileInput`, `UpdateProfileInput`, `ListProfilesInput{SectionName *string, State *bool, Query *string, Page, Limit}` |
| `usecase/profile/repository.go` | `Repository` interface |
| `interface/postgres/profile.go` | sqlx implementation |

### Service struct deps

`usecase/profile/service.go` — standard CRUD service.

### Repository interface methods

`usecase/profile/repository.go` (`Repository`):
- `Create(ctx, *Profile) error`
- `GetByID(ctx, id string) (*Profile, error)`
- `List(ctx, ListProfilesInput) ([]*Profile, int, error)`
- `Update(ctx, id string, *Profile) error`
- `Delete(ctx, id string) error`
- `GetSectionNames(ctx) ([]string, error)`

## Glossary

- **Section** — a named content block (`section_name`) with HTML/text content and a URL endpoint (`section_endpoint`).
- **`state`** — boolean column on `profile`. Not validated, not filtered on public routes.
- **`q`** — admin-only filter param; ILIKE on `section_name` OR `content`. Undocumented in OpenAPI.
- **`section_endpoint`** — VARCHAR(255) URL path; must start with `/`.
- **`ListProfilesInput`** — service input struct with `SectionName *string`, `State *bool`, `Query *string`, `Page`, `Limit`.
- **Public vs admin routes** — `/api/v1/public/profile/*` is unauthenticated; `/api/v1/profile/*` requires JWT + RBAC (`profile:read`/`profile:write`).

## Related

- [database.md](./database.md)
- [endpoint.md](./endpoint.md)
- `pkg/pagination/pagination.go` — pagination normalization (validation bypassed by `ListProfile`)
- `pkg/handlerutil/ValidateStruct` — httpin-side validation
