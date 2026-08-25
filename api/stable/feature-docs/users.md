# Users Feature

User account management — CRUD, self-profile, password reset.

**Module:** `webdesa/api`
**Last updated:** 2026-07-04

## Files

| File | Purpose |
|---|---|
| `interface/http/handler/user/user.go` | `UserHandler` struct + methods (package `user`) |
| `interface/http/handler/user/route.go` | `RegisterRoutes` for `/users/*` (package `user`) |
| `interface/http/handler/role/role.go` | `RoleHandler` — owns `AssignRoleToUser` / `RemoveRoleFromUser` (cross-feature) |
| `interface/http/handler/role/route.go` | `RegisterUserRoleRoutes` mounts `/users/{id}/roles*` |
| `usecase/user/service.go` | Create, List, Get, Update, Delete, password, profile |
| `usecase/user/file_handler.go` | `FileHandler` interface for profile image |
| `usecase/user/repository.go` | `Repository` interface |
| `interface/postgres/user.go` | sqlx implementation |
| `domain/user/user.go` | `User` entity + `Validate()` |

## Service

`usecase/user/service.go:18-32` — `Service{repo, fileHandler, clock}`

| Method | Purpose |
|---|---|
| `Create(ctx, CreateUserInput)` | Blocks duplicate email, bcrypt-hashes, generates UUID |
| `List(ctx, page, limit)` | Paginated, ORDER BY `created_at DESC` |
| `GetByID(ctx, id)` | Single lookup |
| `Update(ctx, id, UpdateUserInput)` | Updates name/email with uniqueness check |
| `Delete(ctx, id)` | FK CASCADE removes roles + tokens |
| `UpdatePassword(ctx, id, UpdatePasswordInput, isAdmin)` | Non-admin requires old password |
| `UpdateProfile(ctx, id, UpdateProfileInput)` | Updates name + profile_image, then re-fetches |

## Endpoints

### Self-service (auth only, no RBAC)

| Method | Path | Notes |
|---|---|---|
| GET | `/api/v1/users/me` | Current user with roles |
| PUT | `/api/v1/users/me` | Multipart (image) or JSON |

### Admin (RBAC: `users:read`)

| Method | Path |
|---|---|
| GET | `/api/v1/users` | Query: `page`, `limit` |
| GET | `/api/v1/users/{id}` | |

### Admin (RBAC: `users:write`)

| Method | Path |
|---|---|
| POST | `/api/v1/users` | Multipart (profile_image) or JSON |
| PUT | `/api/v1/users/{id}` | Multipart or JSON |
| DELETE | `/api/v1/users/{id}` | |
| PUT | `/api/v1/users/{id}/password` | Admin can skip old password |

### Role assignment (RBAC: `roles:write`, mounted from `role/route.go`)

| Method | Path |
|---|---|
| POST | `/api/v1/users/{id}/roles` | Body `{role}` |
| DELETE | `/api/v1/users/{id}/roles/{role}` | |

## Response Shape

```json
{
  "id": "uuid",
  "name": "string",
  "email": "string",
  "profile_image_url": "string|null",
  "roles": ["admin"],
  "created_at": "RFC3339",
  "updated_at": "RFC3339"
}
```

## Business Logic

- Email uniqueness check across both Create and Update
- Old-password verification in UpdatePassword unless caller is admin
- Multipart + JSON dual parsing in Create/Update
- Profile-image cleanup on replace
- FK CASCADE on Delete removes roles + tokens

## RBAC Permissions

- `users:read` — ListUsers, GetUser
- `users:write` — CreateUser, UpdateUser, DeleteUser, UpdatePassword
- `/users/me` requires only JWT (no RBAC)

## Tests

**Unit** (`domain/user/user_test.go`): 13 validation cases

**Integration** (`integration/user_test.go`): 30 cases including Create, DuplicateEmail, List, Update, UpdatePassword (with/without old password), UpdateCurrentUserProfile, GetCurrentUser

## Cross-Feature

`/users/{id}/roles*` endpoints are owned by the **role package** (`role/route.go` exports `RegisterUserRoleRoutes`), not the user package. This avoids cross-package import cycles — `user` doesn't import `role`.

## Related

- `interface/file/local_handler.go` — file storage adapter (10 MB max)
- `pkg/password/bcrypt.go` — bcrypt wrapper
- `pkg/pagination/pagination.go` — pagination normalization
- `pkg/handlerutil/ValidateStruct` — struct validation