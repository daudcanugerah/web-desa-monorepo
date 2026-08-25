# Users — Database

`users` table is the canonical user record. `user_roles` and `password_reset_tokens` are owned by [rbac](../rbac/database.md) and [auth](../auth/database.md) respectively but live close to the user model.

**Module:** webdesa/api · **Last updated:** 2026-08-23

## Migrations

| Migration | Purpose |
|---|---|
| `db/migrations/00001_create_users_table.sql` | Initial `users` table (id, name, email, hashed_password, profile_image_url, timestamps). |
| `db/migrations/00002_create_user_roles_table.sql` | `user_roles` table — PK `(user_id, role)`, FK CASCADE to `users`. |
| `db/migrations/00003_create_password_reset_tokens_table.sql` | `password_reset_tokens` table — FK CASCADE to `users`. |
| `db/migrations/00043_add_profile_image_media_id_to_users.sql` | Add `profile_image_media_id UUID` (FK to `gallery_media(id)`, `ON DELETE SET NULL`). |
| `db/migrations/00045_drop_legacy_upload_columns.sql` | Drop legacy `profile_image_url` column. |

## Tables

### `users` (`db/migrations/00001`, modified by `00043` and `00045`)

| Column | Type | Constraints | Default | Nullable | FK / ON DELETE |
|---|---|---|---|---|---|
| `id` | UUID | PK | `gen_random_uuid()` | NO | — |
| `name` | VARCHAR(255) | NOT NULL | — | NO | — |
| `email` | VARCHAR(255) | UNIQUE, NOT NULL | — | NO | — |
| `hashed_password` | VARCHAR(255) | NOT NULL | — | NO | — |
| `profile_image_url` | VARCHAR(500) | — | — | NO | **DROPPED in migration 00045** |
| `profile_image_media_id` | UUID | — | — | YES | FK → `gallery_media(id)` ON DELETE SET NULL (added 00043) |
| `created_at` | TIMESTAMP | — | `NOW()` | NO | — |
| `updated_at` | TIMESTAMP | — | `NOW()` | NO | — |

### `user_roles` (`db/migrations/00002`)

| Column | Type | Constraints | Default | Nullable | FK / ON DELETE |
|---|---|---|---|---|---|
| `user_id` | UUID | PK part, NOT NULL | — | NO | FK → `users(id)` ON DELETE CASCADE |
| `role` | VARCHAR(100) | PK part, NOT NULL | — | NO | — |
| `created_at` | TIMESTAMP | — | `NOW()` | NO | — |

PK `(user_id, role)`. See [rbac/database.md](../rbac/database.md) for indexes.

### `password_reset_tokens` (`db/migrations/00003`)

| Column | Type | Constraints | Default | Nullable | FK / ON DELETE |
|---|---|---|---|---|---|
| `id` | UUID | PK | `gen_random_uuid()` | NO | — |
| `user_id` | UUID | NOT NULL | — | NO | FK → `users(id)` ON DELETE CASCADE |
| `token` | VARCHAR(255) | UNIQUE, NOT NULL | — | NO | — |
| `expires_at` | TIMESTAMP | NOT NULL | — | NO | — |
| `ip_address` | VARCHAR(45) | NOT NULL | — | NO | — |
| `user_agent` | TEXT | NOT NULL | — | NO | — |
| `used` | BOOLEAN | NOT NULL | `FALSE` | NO | — |
| `created_at` | TIMESTAMP | NOT NULL | — | NO | — |

See [auth/database.md](../auth/database.md) for indexes.

## Indexes

| Index | Table | Column(s) | Migration | Purpose |
|---|---|---|---|---|
| `idx_users_email` | `users` | `email` | `00001` | Login + uniqueness check |
| `idx_user_roles_user_id` | `user_roles` | `user_id` | `00002` | Reverse lookup |
| `idx_user_roles_role` | `user_roles` | `role` | `00002` | Count users with role + delete guard |
| `idx_reset_tokens_token` | `password_reset_tokens` | `token` | `00003` | Lookup by token |
| `idx_reset_tokens_user_id` | `password_reset_tokens` | `user_id` | `00003` | CASCADE + audit |
| `idx_reset_tokens_expires_at` | `password_reset_tokens` | `expires_at` | `00003` | Cleanup |

The `profile_image_media_id` column has no dedicated index — lookups by media id are rare (avatar replacement goes through `GetByID`).

## Seeds

Default users (in `cmd/seed.go`, not in this feature's migration list):
- `admin@desa.local` (password: from env / default) → `admin` role
- `operator@desa.local` (password: from env / default) → `operator` role

Role definitions live in the RBAC feature seed (see [rbac/database.md](../rbac/database.md#seeds)).

## Known Issues

- **Dropped column `profile_image_url`** (migration 00045): kept here for historical reference only. Any code or test referencing `users.profile_image_url` is broken — the column is gone. The entity `domain/user.User` never carried a `ProfileImageURL` field.
- **Stale comment in `domain/user.Validate()`** still references the deleted `ProfileImageURL` field's invariant. Comment is dead/wrong — the entity only has `ProfileImageMediaID`.
- **Broken `Update` SQL** at `interface/postgres/user.go:138`:
  ```sql
  UPDATE users SET name = $1, email = $2 = $3, profile_image_media_id = $4, updated_at = NOW() WHERE id = $5
  ```
  The `email = $2 = $3` clause is malformed, and the statement has 5 placeholders but only 4 args are bound. Every `PUT /users/{id}` and `PUT /users/me` that reaches `repo.Update` fails with a SQL parse error → 500. **See [endpoint.md](./endpoint.md#known-issues) for endpoint impact.**
- **Stale integration assertion** at `integration/user_test.go:731`: asserts `data["profile_image_url"]` after avatar upload. Field no longer exists in `UserResponse`; the assertion never fires (key absent → nil is returned by `NotNil`'s lookup) but should be rewritten to `data["profile_media"]`.
- **No index on `profile_image_media_id`**: avatar lookups by media id go through `users.id` first; no direct path.
- **No CHECK constraint on email format**: validated only at application layer (`domain/user.Validate`).
- **`hashed_password` is `VARCHAR(255)`**: bcrypt-12 produces ~60-byte hashes; 255 is comfortable but oversized.

## Related

- [concept.md](./concept.md)
- [endpoint.md](./endpoint.md)
- [rbac/database.md](../rbac/database.md) — `user_roles`
- [auth/database.md](../auth/database.md) — `password_reset_tokens`
- `db/migrations/00001_create_users_table.sql`
- `db/migrations/00043_add_profile_image_media_id_to_users.sql`
- `db/migrations/00045_drop_legacy_upload_columns.sql`
