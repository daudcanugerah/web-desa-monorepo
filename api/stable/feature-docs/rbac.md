# RBAC Feature

Role-based access control via Casbin.

**Module:** `webdesa/api`
**Last updated:** 2026-07-04

## Files

| File | Purpose |
|---|---|
| `interface/http/handler/role/role.go` | `RoleHandler` struct + methods (package `role`) |
| `interface/http/handler/role/route.go` | `RegisterRoutes` (for `/roles/*` + `/permissions`) and `RegisterUserRoleRoutes` (for `/users/{id}/roles*`) |
| `usecase/role/service.go` | Role management business logic |
| `usecase/role/enforcer.go` | `Enforcer` interface (wraps Casbin methods) |
| `interface/rbac/casbin_enforcer.go` | Casbin enforcer wrapper |
| `interface/rbac/pg_adapter.go` | Custom PostgreSQL adapter for `casbin_rule` |
| `interface/http/middleware/rbac.go` | `RBACMiddleware` |
| `rbac/rbac_model.conf` | Active Casbin model (wildcard matcher) |
| `rbac/rbac_policy.csv` | Default policy (admin + operator) |

## Casbin Model

`rbac/rbac_model.conf`:
```ini
[request_definition]
r = sub, obj, act

[policy_definition]
p = role, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.role) && (p.obj == "*" || r.obj == p.obj) && (p.act == "*" || r.act == p.act)
```

Supports wildcards for both `obj` and `act`. Admin policy `{*, *}` covers everything.

## Database

`casbin_rule` table (`db/migrations/00013`):
| Column | Type | Notes |
|---|---|---|
| `id` | SERIAL PK | |
| `ptype` | VARCHAR(100) | `p` (policy) or `g` (grouping) |
| `v0..v5` | VARCHAR(100) | position-based policy fields |

## Service

`usecase/role/service.go` — `Service{enforcer Enforcer, repo Repository}`

| Method | Purpose |
|---|---|
| `CreateRole(ctx, CreateRoleInput)` | Validates non-empty name. |
| `ListRoles(ctx)` | Iterates `GetAllRoles` + permissions. |
| `GetRolePermissions(ctx, roleName)` | |
| `GetAvailablePermissions(ctx)` | Dedup keyed by `resource:action` |
| `AddPermissionToRole(ctx, roleName, perm)` | |
| `RemovePermissionFromRole(ctx, roleName, perm)` | |
| `DeleteRole(ctx, roleName)` | **Refuses if users assigned** (`"cannot delete role: %d user(s) are assigned"`) |
| `AssignRoleToUser(ctx, userID, roleName)` | Double-writes to Casbin + `user_roles` |
| `RemoveRoleFromUser(ctx, userID, roleName)` | |
| `GetUserRoles(ctx, userID)` | |

## Endpoints (RBAC: `roles:read`)

| Method | Path |
|---|---|
| GET | `/api/v1/roles` |
| GET | `/api/v1/roles/{role}/permissions` | `chi.URLParam` |
| GET | `/api/v1/permissions` | Available system permissions |

## Endpoints (RBAC: `roles:write`)

| Method | Path | Notes |
|---|---|---|
| POST | `/api/v1/roles` | Body `{name, permissions[]}` |
| POST | `/api/v1/roles/{role}/permissions` | Body `{resource, action}` |
| DELETE | `/api/v1/roles/{role}/permissions` | Permission in URL param `resource:action` (query string) |
| DELETE | `/api/v1/roles/{role}` | 400 if users assigned |

### Role assignment (also `roles:write`)

| Method | Path |
|---|---|
| POST | `/api/v1/users/{id}/roles` | Body `{role}` |
| DELETE | `/api/v1/users/{id}/roles/{role}` | |

## RBAC Middleware

`interface/http/middleware/rbac.go`:
```go
RBACMiddleware(enforcer, resource, action)
```

Flow:
1. Read `userID` from context (set by JWT `AuthMiddleware`)
2. Iterate `enforcer.GetRolesForUser(userID)`
3. For each role, call `enforcer.Enforce(role, resource, action)`
4. If any role passes → allow
5. No roles → 401; roles but no pass → 403

## Default Seeded Permissions

`cmd/seed.go:196-236` — `seedRolePermissions`:

- **`admin`** — `{Resource:"*", Action:"*"}` (wildcard)
- **`operator`** — read+write for `berita, umkm, fasilitas, ppid, struktur, banner, desa`

## Available System Permissions

Hardcoded list:
```
users, roles, banners, berita, berita_categories,
umkm, umkm_categories, fasilitas, fasilitas_categories,
ppid, ppid_categories, struktur, desa, profile, infographic,
infographic_categories, backup, restore
```

Each can be combined with actions `read` or `write`.

## Tests

**Unit** (`interface/http/middleware/rbac_test.go`): 5 scenarios with mock adapter

**Integration** (`integration/role_test.go`): 16 cases
- `TestListRoles_Success`, `TestCreateRole_*`
- `TestGetRolePermissions_Success`, `TestAddPermissionToRole_Success`
- `TestGetAvailablePermissions_Success`
- `TestAssignRoleToUser_Success`, `TestRemoveRoleFromUser_Success`
- `TestDeleteRole_Success`

## Double-Write Strategy

When `AssignRoleToUser` is called, the role is persisted in BOTH:
1. Casbin (for fast `Enforce()` lookups via `g` policy)
2. `user_roles` table (for backup queries, FK CASCADE with users)

Both must be kept in sync. On `RemoveRoleFromUser`, both are removed.

## Related

- `rbac/rbac_model.conf` — Casbin model
- `rbac/rbac_policy.csv` — default policy file
- `pkg/casbin/sqlx_adapter.go` — custom Casbin SQL adapter