# Users — Endpoints

Self-service, admin CRUD, password, and avatar upload routes. The `/users/{id}/roles*` endpoints are mounted by the role package — see [rbac/endpoint.md](../rbac/endpoint.md).

**Module:** webdesa/api · **Last updated:** 2026-08-23

## Routes

### Self-service (JWT only, no RBAC)

| Method | Path | Scope | RBAC | Summary |
|---|---|---|---|---|
| GET | `/api/v1/users/me` | protected | — | Current user with roles |
| PUT | `/api/v1/users/me` | protected | — | Multipart OR JSON — **broken via repo SQL** |
| POST | `/api/v1/users/me/upload-media` | protected | — | Multipart `file`. Returns `{url, media_id, filename}`. |

### Read (RBAC: `users:read`)

| Method | Path | Scope | RBAC | Summary |
|---|---|---|---|---|
| GET | `/api/v1/users` | protected | `users:read` | List + pagination (query: `page`, `limit`) |
| GET | `/api/v1/users/{id}` | protected | `users:read` | Single user (UUID-validated) |

### Write (RBAC: `users:write`)

| Method | Path | Scope | RBAC | Summary |
|---|---|---|---|---|
| POST | `/api/v1/users` | protected | `users:write` | **JSON-only** — `{name, email, password}` (no image, no media_id) |
| PUT | `/api/v1/users/{id}` | protected | `users:write` | Multipart OR JSON — **broken via repo SQL** |
| DELETE | `/api/v1/users/{id}` | protected | `users:write` | CASCADE removes roles + tokens |
| PUT | `/api/v1/users/{id}/password` | protected | `users:write` | `{old_password?, new_password}` — admin bypass; `{id}` not UUID-validated |
| POST | `/api/v1/users/upload-media` | protected | `users:write` | Multipart `file` (admin upload) |

### Role assignment (RBAC: `roles:write`, mounted from `role/route.go` → `RegisterUserRoleRoutes`)

| Method | Path | Scope | RBAC | Summary |
|---|---|---|---|---|
| POST | `/api/v1/users/{id}/roles` | protected | `roles:write` | Body `{role}` |
| DELETE | `/api/v1/users/{id}/roles/{role}` | protected | `roles:write` | Remove role from user |

## Request/Response

### GET `/api/v1/users/me`

Success 200:
```json
{ "success": true, "data": { "id": "uuid", "name": "string", "email": "string", "profile_media": {"media_id":"uuid","url":"string","thumbnail_url":"string|null"}, "roles": ["admin"], "created_at": "RFC3339", "updated_at": "RFC3339" } }
```

`profile_media` is **omitempty** — absent when the user has no avatar.

### PUT `/api/v1/users/me` — multipart

Form fields:
- `name` (required)
- `email` (required)
- `profile_image` (file, optional)
- `profile_image_media_id` (string, optional)

400 if both `profile_image` and `profile_image_media_id` are provided.

### PUT `/api/v1/users/me` — application/json

Body:
```json
{ "name": "string", "email": "string", "profile_image_media_id": "uuid?" }
```

No `profile_image` field — uploads must go through `/upload-media` first.

### POST `/api/v1/users/me/upload-media`

Multipart `file` (max 50 MB body). Delegates to gallery `FileStore.Upload`.

Success 200:
```json
{ "success": true, "data": { "url": "string", "media_id": "uuid", "filename": "string" } }
```

### POST `/api/v1/users` (admin create)

Body:
```json
{ "name": "string", "email": "string", "password": "string (min 6)" }
```

**JSON-only** — no image, no media_id. See Known Issues.

### PUT `/api/v1/users/{id}` — multipart OR JSON

Same body shapes as `PUT /users/me`. **Currently broken** — every request that reaches `repo.Update` returns 500 due to malformed SQL. See [Known Issues](#known-issues) and [database.md](./database.md#known-issues).

### PUT `/api/v1/users/{id}/password`

Body:
```json
{ "old_password": "string (optional)", "new_password": "string (min 6)" }
```

- `old_password` empty ⇒ admin bypass. **No role check** — see Known Issues.
- `{id}` is **not UUID-validated** — malformed UUID causes DB error → 500 instead of 400.

Success 200:
```json
{ "success": true, "data": { "message": "..." } }
```

### GET `/api/v1/users`

Query: `page` (default 1), `limit` (default 10).

Success 200:
```json
{ "success": true, "data": { "users": [{...UserResponse...}, ...], "pagination": {"page":1, "limit":10, "total":42, "total_pages":5} } }
```

### GET `/api/v1/users/{id}`

Success 200: same `UserResponse` shape as `GET /users/me`.

## Validation

Handler-level (`httpin` tags):
- `CreateUserInput.name` — `required`
- `CreateUserInput.email` — `required,email`
- `CreateUserInput.password` — `required,min=6`
- `UpdateUserInput.name` — `required` (multipart and JSON)
- `UpdateUserInput.email` — `required,email` (multipart and JSON)
- `UpdatePasswordInput.new_password` — `required,min=6`
- `UpdatePasswordInput.old_password` — **no `required` tag** (untagged JSON)

Domain rules:
- `domain/user.User.Validate()` enforces name/email/password presence and format. **Stale comment** in `Validate()` references deleted `ProfileImageURL` field.
- Email uniqueness checked in both `Create` and `Update`.

## RBAC

| Permission | Routes |
|---|---|
| `users:read` | `GET /users`, `GET /users/{id}` |
| `users:write` | `POST /users`, `PUT /users/{id}`, `DELETE /users/{id}`, `PUT /users/{id}/password`, `POST /users/upload-media` |
| (none — JWT only) | `GET /users/me`, `PUT /users/me`, `POST /users/me/upload-media` |
| `roles:write` (mounted from role pkg) | `POST /users/{id}/roles`, `DELETE /users/{id}/roles/{role}` |

## Known Issues

- **Broken `Update` SQL** at `interface/postgres/user.go:138` (DB-level detail in [database.md](./database.md#known-issues)):
  ```sql
  UPDATE users SET name = $1, email = $2 = $3, profile_image_media_id = $4, updated_at = NOW() WHERE id = $5
  ```
  `email = $2 = $3` is malformed; 5 placeholders, 4 args. Every `PUT /users/{id}` and `PUT /users/me` that reaches `repo.Update` fails with SQL parse error → 500. Repo never reaches the "email already exists" path because Update blows up first.
- **Admin bypass via empty `old_password`** (`user.go:525`): `isAdmin := req.Payload.OldPassword == ""`. Handler treats omission of `old_password` as proof of admin status, but there is **no role check** — anyone with `users:write` can reset any other user's password by sending `{new_password: "x"}` with `old_password` empty. `OldPassword` is also untagged (`json:"old_password,omitempty"` only — no `validate:"required"`).
- **`PUT /users/{id}/password` does not UUID-validate `{id}`** (`user.go:499`): unlike `GET/PUT/DELETE /users/{id}`, this handler skips `handlerutil.IsValidUUID`. Malformed UUID causes DB error → 500 instead of 400.
- **`UpdatePasswordRequest.OldPassword` is untagged** (`user.go:69`): `json:"old_password,omitempty"`. No `validate:"required"` for the non-admin case. The whole "old password required unless admin" contract is enforced by the string-equality check, not by validation.
- **`POST /users` is JSON-only** — no image/media_id support. Admins must upload an avatar separately via `POST /users/upload-media` and then `PUT /users/{id}` to attach it.
- **Stale integration assertion** at `integration/user_test.go:731`: asserts `data["profile_image_url"]` after avatar upload. Field no longer exists in `UserResponse`. Assertion never fires (key absent → nil) but should be rewritten to `data["profile_media"]`.

## Related

- [concept.md](./concept.md)
- [database.md](./database.md)
- [rbac/endpoint.md](../rbac/endpoint.md) — `/users/{id}/roles*`
- [auth/endpoint.md](../auth/endpoint.md) — `/auth/password-reset/*`
