# Profile — Endpoints

Public read endpoints plus admin CRUD. Two distinct list-handler implementations: admin uses manual parsing, public uses httpin.

**Module:** webdesa/api · **Last updated:** 2026-08-23

## Routes

### Public (no auth, no RBAC)

| Method | Path | Scope | RBAC | Summary |
|---|---|---|---|---|
| GET | `/api/v1/public/profile/list` | public | — | Paginated. **`state` not filtered** — admin-disabled sections leak. |
| GET | `/api/v1/public/profile/{id}` | public | — | **No `state` filter**; `{id}` not UUID-validated. |

### Read (RBAC: `profile:read`)

| Method | Path | Scope | RBAC | Summary |
|---|---|---|---|---|
| GET | `/api/v1/profile` | protected | `profile:read` | Query: `page`, `limit`, `section_name`, `state`, `q` (search) |
| GET | `/api/v1/profile/{id}` | protected | `profile:read` | Single |
| GET | `/api/v1/profile/sections/names` | protected | `profile:read` | Distinct section names |

### Write (RBAC: `profile:write`)

| Method | Path | Scope | RBAC | Summary |
|---|---|---|---|---|
| POST | `/api/v1/profile` | protected | `profile:write` | Body `{content, section_name, section_endpoint, state}` |
| PUT | `/api/v1/profile/{id}` | protected | `profile:write` | Body `{content, section_name, section_endpoint, state}` |
| DELETE | `/api/v1/profile/{id}` | protected | `profile:write` | Hard delete |

## Request/Response

### Single (any `GET .../{id}`, `POST /profile`, `PUT /profile/{id}`, public `GET /public/profile/{id}`)

Success 200:
```json
{ "success": true, "data": { "id": "uuid", "content": "string", "section_name": "string", "section_endpoint": "/string", "state": true, "created_at": "RFC3339", "updated_at": "RFC3339" } }
```

### List (`GET /profile`, `GET /public/profile/list`)

Success 200:
```json
{ "success": true, "data": { "profile": [{...ProfileResponse...}, ...], "pagination": {"page":1, "limit":10, "total":42, "total_pages":5} } }
```

Singular key is **`profile`** — not `profiles` or `items`.

### Sections (`GET /profile/sections/names`)

Success 200:
```json
{ "success": true, "data": { "section_names": ["string", ...] } }
```

### POST `/api/v1/profile` / PUT `/api/v1/profile/{id}`

Body:
```json
{ "content": "string (1-10000)", "section_name": "string (1-100)", "section_endpoint": "/string (1-255, must start with /)", "state": true }
```

`state` is optional; absent = `false` (no default true).

## Validation

Handler-level (`httpin` tags):
- `CreateProfileInput.content` — `required,min=1,max=10000`
- `CreateProfileInput.section_name` — `required,min=1,max=100`
- `CreateProfileInput.section_endpoint` — `required,min=1,max=255`
- `CreateProfileInput.state` — **no validation**
- `UpdateProfileInput.*` — same tags as Create
- `ListProfilesInput.SectionName` — `omitempty`
- `ListProfilesInput.State` — **no validation** (parsed manually)
- `ListProfilesInput.Query` — `omitempty` (`q` param)
- `ListProfilesInput.Page` — `omitempty` (manual parse; bad → default 1)
- `ListProfilesInput.Limit` — `omitempty` (manual parse; bad → default 10; **no max check**)

Domain rules (`domain/profile.Validate()`):
- `Content` — required, 1–10000 chars
- `SectionName` — required, 1–100 chars
- `SectionEndpoint` — required, 1–255 chars, **must start with `/`**
- `State` — **no check**

## RBAC

| Permission | Routes |
|---|---|
| (none — public) | `GET /public/profile/list`, `GET /public/profile/{id}` |
| `profile:read` | `GET /profile`, `GET /profile/{id}`, `GET /profile/sections/names` |
| `profile:write` | `POST /profile`, `PUT /profile/{id}`, `DELETE /profile/{id}` |

## Known Issues

- **`q` query param is undocumented** in the OpenAPI spec (`GET /profile`). The admin handler parses it (`profile.go:174-196`) and applies ILIKE `%q%` on `section_name` OR `content`. The public `ListProfilePublic` does **not** support `q`. Coverage: `TestProfile_List_FilterByQuery`.
- **`state` has no validation** — neither struct tag on `CreateProfileInput.State`/`UpdateProfileInput.State` nor a check in `Profile.Validate()`. Omitting the field silently yields `false`. DB column default of `true` is dead (see [database.md](./database.md#known-issues)).
- **`GET /public/profile/list` and `GET /public/profile/{id}` do not filter by `state`** — admin-disabled sections leak to the public site. Public handler ignores `state` entirely (no parameter parsed, no service filter set). **Design issue** → see [concept.md](./concept.md#public-endpoints-never-filter-by-state).
- **`{id}` path param is not UUID-validated** on profile handlers — unlike `user.GetUser`/`DeleteUser`/`UpdateUser`. Malformed id surfaces as 500 from `sql: invalid UUID` rather than 400. **Code-level bug** — present in GET admin, GET public, PUT, DELETE.
- **`limit > 100` returns 500, not 400** in admin `ListProfile` — manual parse bypasses `pagination.MaxLimit`; `pagination.Paginate` error is wrapped and returned as `StatusInternalServerError`. `ListProfilePublic` does the right thing (400).
- **Bad/non-numeric `page` / `limit` silently default** in admin `ListProfile` — should return 400. `ListProfilePublic` validates `Page > 0` / `Limit > 0` and returns 400.
- **Hard delete only** — no soft-delete column. `DELETE /profile/{id}` is irreversible.
- **`profile.go:174-196`** — the manual-parse implementation sits inside `ListProfile`; the line citation reflects the actual code.

## Related

- [concept.md](./concept.md)
- [database.md](./database.md)
