# RBAC — Database

Two tables: `casbin_rule` (active Casbin storage) and `user_roles` (backup for user→role assignment with FK CASCADE).

**Module:** webdesa/api · **Last updated:** 2026-08-23

## Migrations

| Migration | Purpose |
|---|---|
| `db/migrations/00002_create_user_roles_table.sql` | Initial `user_roles` table — PK `(user_id, role)`, FK CASCADE to `users`. |
| `db/migrations/00013_create_casbin_rule_table.sql` | `casbin_rule` table — Casbin's PostgreSQL adapter storage. |

## Tables

### `casbin_rule` (`db/migrations/00013`)

| Column | Type | Constraints | Default | Nullable | FK / ON DELETE |
|---|---|---|---|---|---|
| `id` | SERIAL | PK | auto-increment | NO | — |
| `ptype` | VARCHAR(100) | — | — | NO | — |
| `v0` | VARCHAR(100) | — | — | NO | — |
| `v1` | VARCHAR(100) | — | — | NO | — |
| `v2` | VARCHAR(100) | — | — | NO | — |
| `v3` | VARCHAR(100) | — | — | YES | — |
| `v4` | VARCHAR(100) | — | — | YES | — |
| `v5` | VARCHAR(100) | — | — | YES | — |

Notes:
- `ptype` is `p` (policy) or `g` (grouping).
- `v0..v5` are position-based policy fields per the standard Casbin PostgreSQL adapter.
- **No unique constraint** on `(ptype, v0..v5)` — see Known Issues.

### `user_roles` (`db/migrations/00002`)

| Column | Type | Constraints | Default | Nullable | FK / ON DELETE |
|---|---|---|---|---|---|
| `user_id` | UUID | PK part | — | NO | FK → `users(id)` ON DELETE CASCADE |
| `role` | VARCHAR(100) | PK part, NOT NULL | — | NO | — |
| `created_at` | TIMESTAMP | — | `NOW()` | NO | — |

PK `(user_id, role)`. No FK on `role` to a separate `roles` table — role names are free-form strings managed entirely in Casbin.

## Indexes

| Index | Table | Column(s) | Purpose |
|---|---|---|---|
| `idx_casbin_rule_ptype` | `casbin_rule` | `ptype` | Filter by policy vs grouping |
| `idx_casbin_rule_v0` | `casbin_rule` | `v0` | Lookup by role/user (v0 = role for `p`, user for `g`) |
| `idx_casbin_rule_v1` | `casbin_rule` | `v1` | Lookup by resource (for `p`) or role (for `g`) |
| `idx_casbin_rule_v2` | `casbin_rule` | `v2` | Lookup by action (for `p`) |
| `idx_user_roles_user_id` | `user_roles` | `user_id` | Reverse lookup (rarely used; `GetUserRoles` reads Casbin) |
| `idx_user_roles_role` | `user_roles` | `role` | Count users with role + role delete guard |

## Seeds

`cmd/seed.go:246-288` — `seedRolePermissions`:

- **`admin`** — `{Resource:"*", Action:"*"}` (one `p` rule)
- **`operator`** — 16 `p` rules across 8 resources (`berita`, `umkm`, `fasilitas`, `ppid`, `struktur`, `banner`, `desa`, `gallery`), each with `read` + `write`

User→role assignments (`g` rules) for the seeded users (`admin@desa.local` → `admin`, `operator@desa.local` → `operator`) are also inserted into `casbin_rule` and mirrored to `user_roles`.

## Known Issues

- **`casbin_rule` has no unique constraint** (`interface/rbac/pg_adapter.go:76`): `AddPolicy` uses `ON CONFLICT DO NOTHING` but `casbin_rule` has nothing to conflict with. Duplicate `(ptype, v0..v5)` rows accumulate across restarts/seed runs.
- **No FK from `user_roles.role` to a roles catalog**: role names are free-form. Typos and case-mismatches silently fail to match in Casbin.
- **No timestamp on `casbin_rule`**: impossible to tell when a policy was added or removed. The API fabricates `created_at`/`updated_at` per-request — see [concept.md](./concept.md#fabricated-timestamps).
- **`user_roles` is write-only on the read path**: `GetUserRoles` reads Casbin only, so a successful `user_roles` insert paired with a failed Casbin write leaves the two stores inconsistent with no read-side detection.
- **No cleanup of `casbin_rule` rows when a role is deleted**: deleting a role via Casbin leaves its `p` rows, but `g` rows for users pointing at it may still resolve during `Enforce`.

See [concept.md](./concept.md#double-write-strategy) and [endpoint.md](./endpoint.md#known-issues) for application-level issues.

## Related

- [concept.md](./concept.md)
- [endpoint.md](./endpoint.md)
- [users/database.md](../users/database.md) — `users` table (FK target for `user_roles`)
- [auth/database.md](../auth/database.md) — `password_reset_tokens` (FK target for reset tokens)
