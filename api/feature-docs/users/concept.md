# Users — Concept

User account management — CRUD, self-profile, password reset, avatar upload via gallery. The avatar pipeline no longer touches raw URLs or a separate `FileHandler`; uploads route through the gallery `FileStore`.

**Module:** webdesa/api · **Last updated:** 2026-08-23

## Purpose

- Admin CRUD for user accounts (list, create, get, update, delete).
- Self-service endpoints (`/users/me*`) for the authenticated user.
- Avatar upload via two upload routes (`/users/me/upload-media`, `/users/upload-media`) that delegate to the gallery `FileStore`.
- Password update with admin-bypass semantics.
- Cross-feature role assignment endpoints mounted by the **role package** (avoids import cycles).

## Business Rules

### Avatar pipeline

User avatars do **not** accept a raw URL string. The flow is:

1. Client uploads an image to `POST /users/me/upload-media` or `POST /users/upload-media` (multipart `file`, 50 MB max body).
2. Handler delegates to gallery `FileStore.Upload(...)` → returns `{url, media_id, filename}`.
3. Client calls `PUT /users/me` or `PUT /users/{id}` with `profile_image_media_id` (JSON) or `profile_image` (multipart).
4. Service `UpdateProfileImage` / `UpdateProfileImageMediaID` replaces the old avatar (deletes the previous gallery media).

There is **no** `FileHandler` interface, **no** `UpdateProfileInput`, **no** `UpdateProfile` method, **no** `usecase/user/file_handler.go`. The avatar pipeline lives entirely on `galleryUsecase.FileStore` + `UpdateProfileImage` / `UpdateProfileImageMediaID`.

### Email uniqueness

Email uniqueness is checked across both `Create` and `Update`. The DB has a `UNIQUE` constraint on `email` (added in `00001`).

### Password update — admin bypass

`UpdatePassword(ctx, id, UpdatePasswordInput, isAdmin)`:
- Non-admin: requires `old_password`. Verifies against `bcrypt`.
- Admin: `old_password` may be empty. **No role check** — the bypass key is the **absence of `old_password`**, not the user's actual role. Any caller with `users:write` can reset any user's password by sending `{new_password: "x"}` with `old_password` empty. See [endpoint.md](./endpoint.md#known-issues).

### `PUT /users/me` vs `PUT /users/{id}`

Both branch on `Content-Type`:
- **multipart/form-data**: `name` (required), `email` (required), `profile_image` (file), `profile_image_media_id` (string, optional). 400 if both image fields provided.
- **application/json**: `{name, email, profile_image_media_id?}`. No `profile_image` field — file uploads must go through the `/upload-media` endpoints first.

`PUT /users/me` requires **both** `name` AND `email` — omitting either returns 400.

### Profile image response shape

Response uses the shared `profile_media` object:
```json
"profile_media": { "media_id": "uuid", "url": "string", "thumbnail_url": "string|null" }
```

`profile_media` is **omitempty** — absent when the user has no avatar. The legacy `profile_image_url` string field **no longer exists** (column dropped in migration `00045`, entity field never existed in the current revision).

### List response

`GET /users` returns:
```json
{ "success": true, "data": { "users": [...], "pagination": {"page":1, "limit":10, "total":42, "total_pages":5} } }
```

### Delete cascade

`Delete` on a user cascades to `user_roles` and `password_reset_tokens` (FK `ON DELETE CASCADE`). Roles in Casbin (`g` rows) are **not** automatically cleaned — `AssignRoleToUser` does not run on user delete.

### Cross-feature mount

`/users/{id}/roles*` endpoints are owned by the **role package** (`role/route.go` exports `RegisterUserRoleRoutes`). The `user` package does not import `role` — this avoids import cycles.

### Domain entity

`domain/user/user.go` — `User{ID, Name, Email, HashedPassword, ProfileImageMediaID *string, CreatedAt, UpdatedAt}`. **No `ProfileImageURL` field** — the legacy column is gone. The stale comment in `domain/user.Validate()` claiming an invariant about `ProfileImageURL` OR `ProfileImageMediaID` is dead/wrong and should be ignored.

## Architecture

### Files

| File | Purpose |
|---|---|
| `interface/http/handler/user/user.go` | `UserHandler` struct + methods (package `user`) |
| `interface/http/handler/user/user_upload.go` | `UserUploadHandler` — `POST /users/upload-media` + `POST /users/me/upload-media`. 50 MB max body. |
| `interface/http/handler/user/route.go` | `RegisterRoutes` for `/users/*` |
| `interface/http/handler/role/role.go` | `RoleHandler` — owns `AssignRoleToUser` / `RemoveRoleFromUser` |
| `interface/http/handler/role/route.go` | `RegisterUserRoleRoutes` mounts `/users/{id}/roles*` |
| `usecase/user/service.go:19-33` | `Service{repo, fileStore galleryUsecase.FileStore, clock}` |
| `usecase/user/repository.go` | `Repository` interface |
| `interface/postgres/user.go` | sqlx implementation |
| `domain/user/user.go` | `User` entity + `Validate()` |

### Service struct deps

`usecase/user/service.go:19-33` — `Service{repo Repository, fileStore galleryUsecase.FileStore, clock clock.Clock}`

### Repository interface methods

`usecase/user/repository.go` (`Repository`):
- `Create(ctx, *User) error`
- `FindByEmail(ctx, email string) (*User, error)`
- `List(ctx, page, limit int) ([]*User, int, error)`
- `GetByID(ctx, id string) (*User, error)`
- `Update(ctx, id string, name, email string, profileImageMediaID *string) error`
- `Delete(ctx, id string) error`
- `UpdatePassword(ctx, id string, hashedPassword string) error`
- `GetProfileImageMediaID(ctx, id string) (*string, error)`

## Glossary

- **Avatar / profile image** — gallery media used as the user's avatar. Stored as `users.profile_image_media_id` referencing `gallery_media(id)`.
- **`profile_image_media_id`** — UUID column on `users`. Nullable. FK CASCADE behavior is `ON DELETE SET NULL` (from gallery side).
- **`profile_image_url`** — legacy column, **dropped in migration 00045**. The entity never had a corresponding field.
- **`profile_media`** — response-side object: `{media_id, url, thumbnail_url}`. `omitempty` in JSON.
- **`FileStore`** — gallery package's upload interface, used by the user service for avatar uploads.
- **`isAdmin`** — boolean parameter to `UpdatePassword`. Computed by the handler from `req.Payload.OldPassword == ""` (not from the user's actual role).
- **`UserUploadHandler`** — separate handler type from `UserHandler`; routes are registered in the same `user/route.go`.

## Related

- [database.md](./database.md)
- [endpoint.md](./endpoint.md)
- [rbac/endpoint.md](../rbac/endpoint.md) — `/users/{id}/roles*` (mounted from role package)
- [auth/endpoint.md](../auth/endpoint.md) — `/auth/password-reset/*` (separate, not user package)
- `usecase/gallery` — `FileStore` (avatar upload + delete)
- `interface/file/local_handler.go` — legacy local file handler; **no longer the user avatar path**
- `pkg/password/bcrypt.go` — bcrypt wrapper (cost 12)
- `pkg/pagination/pagination.go` — pagination normalization
- `pkg/handlerutil/ValidateStruct` — struct validation
