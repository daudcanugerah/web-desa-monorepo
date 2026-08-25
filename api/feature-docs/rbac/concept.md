# RBAC — Concept

Role-based access control via Casbin with a PostgreSQL adapter. Wildcard matcher policy; admin role covers everything. Role assignments are double-written (Casbin + `user_roles`) without a transaction.

**Module:** webdesa/api · **Last updated:** 2026-08-23

## Purpose

- Manage roles (create / list / delete) and their permissions (`resource:action` pairs).
- List available permissions to seed UI dropdowns.
- Assign and remove roles from users.
- Gate every protected route via a single middleware (`RBACMiddleware`).

Casbin is configured at startup with `rbac/rbac_model.conf` (wildcard matcher) and a PostgreSQL adapter (`casbin_rule` table). The CSV-based policy file `rbac/rbac_policy.csv` is **on disk but not loaded** — see Known Issues.

## Business Rules

### Casbin model — wildcard matcher

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

Both `obj` and `act` support `*`. Admin policy `{*, *}` covers all (resource, action) combinations.

### Middleware contract (401 vs 403)

`RBACMiddleware` (`interface/http/middleware/rbac.go`):
1. Read `userID` from context (set by JWT `AuthMiddleware`).
2. Iterate `enforcer.GetRolesForUser(userID)`.
3. For each role, call `enforcer.Enforce(role, resource, action)`.
4. If any role passes → allow.
5. **`userID` missing → 401**.
6. **Otherwise (no roles, or roles but none pass) → 403**.

`RBACMiddlewareWithContext` is also 401-only-if-no-userID; all other failure modes (missing ctx values, no roles, no pass) are 403 or 500 depending on the inner error.

### Role deletion guard

`DeleteRole(ctx, roleName)` refuses if users are assigned. Error string is `"cannot delete role: %d user(s) are assigned to this role"` — but the handler compares against the literal `"cannot delete role: users are assigned to this role"` (no `%d`), so the match **never succeeds** → the handler falls through to 500 instead of returning 400. See [endpoint.md](./endpoint.md#known-issues).

### Available permissions source

`GET /api/v1/permissions` returns **deduped rows from the `casbin_rule` table** via `enforcer.GetAllPermissions()`. There is **no hardcoded resource list** — what the API returns is exactly what has been seeded. After a fresh seed: admin `{*,*}` + operator's 16 pairs.

### Fabricated timestamps

`created_at` / `updated_at` on role responses are **fabricated per-request** (`time.Now()` inside `ListRoles`/`CreateRole`) — Casbin stores no timestamps. They are not useful as audit signals.

### Double-write strategy

`AssignRoleToUser` writes the role in BOTH:
1. Casbin (for fast `Enforce()` lookups via `g` policy)
2. `user_roles` table (for backup queries, FK CASCADE with users)

Both must be kept in sync. `RemoveRoleFromUser` removes both. **There is no transaction wrapping the two writes** — see Known Issues.

### Read-from-Casbin-only quirk

`GetUserRoles(ctx, userID)` **reads from Casbin only**, never from `user_roles`. If the double-write diverges (e.g. Casbin insert failed, `user_roles` insert succeeded), lookups miss real assignments.

### Default seeded permissions

`cmd/seed.go:246-288` — `seedRolePermissions`:

- **`admin`** — `{Resource:"*", Action:"*"}` (wildcard)
- **`operator`** — 16 pairs across 8 resources:
  - `berita:read,write`
  - `umkm:read,write`
  - `fasilitas:read,write`
  - `ppid:read,write`
  - `struktur:read,write`
  - `banner:read,write`
  - `desa:read,write`
  - `gallery:read,write`

`admin` covers all 16 plus any future resource/action.

## Architecture

### Files

| File | Purpose |
|---|---|
| `interface/http/handler/role/role.go` | `RoleHandler` struct + methods (package `role`) |
| `interface/http/handler/role/route.go` | `RegisterRoutes` (for `/roles/*` + `/permissions`) and `RegisterUserRoleRoutes` (for `/users/{id}/roles*`) |
| `usecase/role/service.go` | Role management business logic |
| `usecase/role/enforcer.go` | `Enforcer` interface (wraps Casbin methods) |
| `usecase/role/repository.go` | `Repository` interface (`CountUsersWithRole`, `AssignUserRole`, `RemoveUserRole`) |
| `usecase/role/types.go` | `Role`, `Permission`, `CreateRoleInput` |
| `interface/rbac/casbin_enforcer.go` | Casbin enforcer wrapper; uses `interface/rbac/pg_adapter.go` |
| `interface/rbac/pg_adapter.go` | In-use PostgreSQL adapter for `casbin_rule` |
| `interface/postgres/role.go` | sqlx impl of `Repository` |
| `interface/http/middleware/rbac.go` | `RBACMiddleware`, `RBACMiddlewareWithContext`, `WithRBAC` |
| `interface/http/middleware/middleware_deps.go` | `MiddlewareDeps` bundling for per-route wiring |
| `rbac/rbac_model.conf` | Active Casbin model (wildcard matcher — loaded by `cmd/serve.go`) |

> **Not loaded at runtime:** `rbac/rbac_policy.csv` is on disk but the enforcer uses the `casbin_rule` table. The CSV content is stale (predates `gallery`/wildcard policy).
>
> **Dead:** `rbac/model.conf` + `rbac/policy.csv` (different model — `p = sub, obj, act` instead of `p = role, obj, act` — incompatible with `rbac_model.conf`; nothing imports them). `pkg/casbin/sqlx_adapter.go` also has zero importers in production code (only its own test).

### Service struct deps

`usecase/role/service.go` — `Service{enforcer Enforcer, repo Repository}`

### Repository interface methods

`usecase/role/repository.go` (`Repository`):
- `CountUsersWithRole(ctx, roleName string) (int, error)`
- `AssignUserRole(ctx, userID string, roleName string) error`
- `RemoveUserRole(ctx, userID string, roleName string) error`

`usecase/role/enforcer.go` (`Enforcer` — Casbin wrapper):
- `AddPolicy(ctx, role, resource, action string) error`
- `RemovePolicy(ctx, role, resource, action string) error`
- `AddRoleForUser(ctx, userID, role string) error`
- `RemoveRoleForUser(ctx, userID, role string) error`
- `GetRolesForUser(ctx, userID string) ([]string, error)`
- `GetAllRoles(ctx) ([]string, error)`
- `GetPermissionsForRole(ctx, role string) ([][]string, error)`
- `GetAllPermissions(ctx) ([][]string, error)`
- `Enforce(ctx, role, resource, action string) (bool, error)`
- `DeleteRole(ctx, role string) error`

### RBAC middleware surface

| Function | Signature | Notes |
|---|---|---|
| `RBACMiddleware` | `(enforcer *casbin.Enforcer, resource, action string) func(http.Handler) http.Handler` | Concrete `*casbin.Enforcer`, not an interface |
| `RBACMiddlewareWithContext` | `(enforcer *casbin.Enforcer) func(http.Handler) http.Handler` | Reads `rbac_resource` / `rbac_action` from `context.Context` |
| `WithRBAC` | `(resource, action string) func(http.Handler) http.Handler` | Writes raw string keys (`"rbac_resource"`, `"rbac_action"`) into ctx |

`MiddlewareDeps.RBAC(resource, action)` is the route-level wrapper that calls `RBACMiddleware` with the bundled enforcer.

## Glossary

- **Casbin** — policy engine (`github.com/casbin/casbin`). Loaded with a model + adapter.
- **`casbin_rule`** — PostgreSQL table backing the policy + grouping storage.
- **`g` policy** — Casbin grouping (role inheritance / user→role assignment).
- **`p` policy** — Casbin permission rule `(role, resource, action)`.
- **`user_roles`** — backup PostgreSQL table for user→role assignments. FK CASCADE with `users`.
- **Wildcard policy** — `{Resource:"*", Action:"*"}` matches every (resource, action) per the model matcher.
- **Double-write** — the practice of writing to both Casbin and `user_roles` on assignment/removal.

## Related

- [database.md](./database.md)
- [endpoint.md](./endpoint.md)
- `rbac/rbac_model.conf` — Casbin model (active)
- `pkg/casbin/sqlx_adapter.go` — alternate Casbin adapter (dead, kept for tests)
- `interface/rbac/pg_adapter.go` — live PostgreSQL Casbin adapter
- [users/concept.md](../users/concept.md) — `/users/{id}/roles*` endpoints are mounted from the role package
