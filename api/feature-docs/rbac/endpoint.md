# RBAC — Endpoints

Routes for role management, permission catalog, and user→role assignment. Mounted from `interface/http/handler/role/route.go`.

**Module:** webdesa/api · **Last updated:** 2026-08-23

## Routes

### Read (RBAC: `roles:read`)

| Method | Path | Scope | RBAC | Summary |
|---|---|---|---|---|
| GET | `/api/v1/roles` | protected | `roles:read` | List all roles + permissions |
| GET | `/api/v1/roles/{role}/permissions` | protected | `roles:read` | Permissions for one role |
| GET | `/api/v1/permissions` | protected | `roles:read` | Deduped `(resource, action)` pairs from `casbin_rule` |

### Write (RBAC: `roles:write`)

| Method | Path | Scope | RBAC | Summary |
|---|---|---|---|---|
| POST | `/api/v1/roles` | protected | `roles:write` | Create role `{name, permissions[]}` |
| POST | `/api/v1/roles/{role}/permissions` | protected | `roles:write` | Add one permission `{resource, action}` |
| DELETE | `/api/v1/roles/{role}/permissions` | protected | `roles:write` | **BROKEN** — see Known Issues |
| DELETE | `/api/v1/roles/{role}` | protected | `roles:write` | Delete role — **500 on in-use**, not 400 |

### Role assignment (RBAC: `roles:write`, mounted from `role/route.go` → `RegisterUserRoleRoutes`)

| Method | Path | Scope | RBAC | Summary |
|---|---|---|---|---|
| POST | `/api/v1/users/{id}/roles` | protected | `roles:write` | Body `{role}` |
| DELETE | `/api/v1/users/{id}/roles/{role}` | protected | `roles:write` | Remove role from user |

## Request/Response

### POST `/api/v1/roles`

Request:
```json
{ "name": "string", "permissions": [{"resource": "string", "action": "string"}, ...] }
```

Success 200:
```json
{ "success": true, "data": { "name": "string", "permissions": [{"resource": "string", "action": "string"}], "created_at": "RFC3339 (fabricated)", "updated_at": "RFC3339 (fabricated)" } }
```

Errors: `400` (empty name), `401` (no user), `403` (no `roles:write`).

### GET `/api/v1/roles`

Success 200:
```json
{ "success": true, "data": [ {"name": "string", "permissions": [{"resource": "string", "action": "string"}], "created_at": "RFC3339 (fabricated)", "updated_at": "RFC3339 (fabricated)"}, ... ] }
```

### GET `/api/v1/roles/{role}/permissions`

Success 200:
```json
{ "success": true, "data": [ {"resource": "string", "action": "string"}, ... ] }
```

### GET `/api/v1/permissions`

Success 200:
```json
{ "success": true, "data": [ {"resource": "string", "action": "string"}, ... ] }
```

Returns deduped `(resource, action)` rows from `casbin_rule`. **No hardcoded list** — see [concept.md](./concept.md#available-permissions-source).

### POST `/api/v1/roles/{role}/permissions`

Request:
```json
{ "resource": "string", "action": "string" }
```

Success 200:
```json
{ "success": true, "data": { "message": "..." } }
```

### DELETE `/api/v1/roles/{role}/permissions` — **BROKEN**

The chi route is `/roles/{role}/permissions` — **no `{permission}` segment** — so `chi.URLParam(r, "permission")` always returns `""` → handler returns `400 "Permission is required"` on every call (100% failure).

Swagger annotation `@Param permission query string` is also wrong — should be a path param.

Even if the URL were fixed, the parse uses `fmt.Sscanf(permissionStr, "%s:%s", ...)` which cannot split on `:` because `%s` stops at whitespace, not `:`.

### DELETE `/api/v1/roles/{role}`

Success 200:
```json
{ "success": true, "data": { "message": "..." } }
```

Errors: `500` (not `400`) when role is in use — see Known Issues.

### POST `/api/v1/users/{id}/roles`

Request:
```json
{ "role": "string" }
```

Success 200:
```json
{ "success": true, "data": { "message": "..." } }
```

Non-transactional double-write to Casbin + `user_roles` — see [concept.md](./concept.md#double-write-strategy).

### DELETE `/api/v1/users/{id}/roles/{role}`

Success 200:
```json
{ "success": true, "data": { "message": "..." } }
```

## Validation

- `name` — `required` (CreateRole)
- `resource`, `action` — `required` (AddPermission)
- `role` (in `{role}` path) — no UUID validation (free-form string)

## RBAC

| Permission | Routes |
|---|---|
| `roles:read` | `GET /roles`, `GET /roles/{role}/permissions`, `GET /permissions` |
| `roles:write` | `POST /roles`, `POST /roles/{role}/permissions`, `DELETE /roles/{role}/permissions` (broken), `DELETE /roles/{role}`, `POST /users/{id}/roles`, `DELETE /users/{id}/roles/{role}` |

Middleware contract: `userID` missing → 401; otherwise (no roles, no pass) → 403. See [concept.md](./concept.md#middleware-contract-401-vs-403).

## Known Issues

- **`DELETE /roles/{role}/permissions` is broken** (`role.go:323`): no `{permission}` path segment → 100% `400 "Permission is required"`. Swagger annotation is also wrong (query string vs path). Even with a URL fix, `fmt.Sscanf("%s:%s")` can't split on `:` because `%s` stops at whitespace.
- **`DELETE /roles/{role}` returns 500 on in-use**, not 400 (`role.go:376`): handler compares service error string against `"cannot delete role: users are assigned to this role"` but the service emits `"cannot delete role: %d user(s) are assigned to this role"` — never matches → falls through to 500.
- **Non-transactional double-write** (`role/service.go:240-246`): `AssignRoleToUser` writes Casbin then `user_roles` without a transaction. If the second write fails, the two stores diverge.
- **`GetUserRoles` reads Casbin only** (`role/service.go:283`): `user_roles` is write-only on this path. Casbin/`user_roles` divergence has no read-side detection.
- **`pgAdapter.AddPolicy` has no uniqueness** (`interface/rbac/pg_adapter.go:76`): `ON CONFLICT DO NOTHING` against a table with no unique constraint → duplicate `(ptype, v0..v5)` rows accumulate. See [database.md](./database.md#known-issues).
- **Dead code**: `pkg/casbin/sqlx_adapter.go` has zero production importers. `rbac/model.conf` + `rbac/policy.csv` are incompatible second Casbin model — nothing loads them. `rbac/rbac_policy.csv` is on disk but not loaded.
- **`created_at` / `updated_at` are per-request fabrications** — see [concept.md](./concept.md#fabricated-timestamps).

## Related

- [concept.md](./concept.md)
- [database.md](./database.md)
- [users/endpoint.md](../users/endpoint.md) — `/users/{id}/roles*` is mounted from the role package
