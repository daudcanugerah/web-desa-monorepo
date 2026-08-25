# Auth — Database

Schema for password reset tokens. Users table lives under the [users](../users/database.md) feature.

**Module:** webdesa/api · **Last updated:** 2026-08-23

## Migrations

| Migration | Purpose |
|---|---|
| `db/migrations/00003_create_password_reset_tokens_table.sql` | Initial `password_reset_tokens` table for opaque 32-byte hex reset tokens. |

The reset-token table is touched by exactly one migration. No subsequent migrations add or drop columns.

## Tables

### `password_reset_tokens`

| Column | Type | Constraints | Default | Nullable | FK / ON DELETE |
|---|---|---|---|---|---|
| `id` | UUID | PK | `gen_random_uuid()` | NO | — |
| `user_id` | UUID | — | — | NO | FK → `users(id)` ON DELETE CASCADE |
| `token` | VARCHAR(255) | UNIQUE | — | NO | — |
| `expires_at` | TIMESTAMP | — | — | NO | — |
| `ip_address` | VARCHAR(45) | — | — | NO | — |
| `user_agent` | TEXT | — | — | NO | — |
| `used` | BOOLEAN | — | `FALSE` | NO | — |
| `created_at` | TIMESTAMP | — | — | NO | — |

Notes:
- `token` is the only UNIQUE column — there is no uniqueness on `(user_id, used)` so the same token can be flagged `used=true` multiple times without DB error (the handler just updates the existing row).
- `used` defaults to `FALSE`; the repo `MarkResetTokenUsed` does `UPDATE ... SET used = TRUE WHERE token = $1`.
- No `updated_at` column — the `used` flip is not timestamped.

## Indexes

| Index | Column(s) | Migration | Purpose |
|---|---|---|---|
| `idx_reset_tokens_token` | `token` | `00003` | Lookup by token on confirm/check |
| `idx_reset_tokens_user_id` | `user_id` | `00003` | CASCADE + audit queries |
| `idx_reset_tokens_expires_at` | `expires_at` | `00003` | Cleanup of expired rows |

## Seeds

None. Reset tokens are created on demand by `POST /auth/password-reset/request`. Default user accounts (`admin@desa.local`, `operator@desa.local`) are seeded elsewhere — see `cmd/seed.go`.

## Known Issues

- **`expires_at` is computed but not DB-enforced**: nothing queries or deletes rows past `expires_at`. Expired tokens remain in the table until manual cleanup. The application layer (`CheckResetToken`) compares against `time.Now()`.
- **No uniqueness on `(user_id, used)`**: the same token can be marked used multiple times without DB error. The handler relies on application-layer idempotency (`MarkResetTokenUsed` is called once per request).
- **`ip_address VARCHAR(45)`**: sized for IPv6 in `IP:port` form (max 45 chars). IPv4-in-IPv6 mapped addresses still fit.
- **`user_agent TEXT`**: unbounded — agents can be very long; consider `VARCHAR(255)` if size matters.

See [endpoint.md](./endpoint.md#known-issues) for runtime-level issues (email TTL bug, rate-limit key behavior).

## Related

- [concept.md](./concept.md)
- [endpoint.md](./endpoint.md)
- [users/database.md](../users/database.md) — `users` table (FK target)
