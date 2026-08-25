# Profile — Database

Single `profile` table. No migrations beyond `00020_create_profile_table.sql`.

**Module:** webdesa/api · **Last updated:** 2026-08-23

## Migrations

| Migration | Purpose |
|---|---|
| `db/migrations/00020_create_profile_table.sql` | Initial `profile` table — id, content, section_name, section_endpoint, state, timestamps. **The only migration touching `profile`.** |

## Tables

### `profile` (`db/migrations/00020_create_profile_table.sql`)

| Column | Type | Constraints | Default | Nullable | FK / ON DELETE |
|---|---|---|---|---|---|
| `id` | UUID | PK | `gen_random_uuid()` | NO | — |
| `content` | TEXT | NOT NULL | — | NO | — |
| `section_name` | VARCHAR(100) | NOT NULL | — | NO | — |
| `section_endpoint` | VARCHAR(255) | NOT NULL | — | NO | — |
| `state` | BOOLEAN | NOT NULL | `true` | NO | — |
| `created_at` | TIMESTAMP | — | `NOW()` | NO | — |
| `updated_at` | TIMESTAMP | — | `NOW()` | NO | — |

Notes:
- No FK from `section_endpoint` to anything — `section_endpoint` is a free-form URL path string validated only at the application layer (must start with `/`).
- `state` default of `true` is **dead** at the application layer — see [concept.md](./concept.md#state-is-not-validated).

## Indexes

| Index | Column(s) | Migration | Purpose |
|---|---|---|---|
| `idx_profile_section_name` | `section_name` | `00020` | Filter by section name + `GetSectionNames` DISTINCT |
| `idx_profile_state` | `state` | `00020` | Filter by state (unused — public routes don't filter) |

## Seeds

None. Profiles are created on demand via `POST /api/v1/profile`.

## Known Issues

- **`state` default `true` is dead**: the repo binds `$5` always in the INSERT, so the column default never fires. New rows always get the value the handler passes (often `false` because the field has no validation default).
- **`section_endpoint` has no DB-level constraint** that it must start with `/`. Validated only at the application layer (`Profile.Validate()`).
- **No length check on `content`**: column is `TEXT` (unbounded). Application validates 1–10000 chars at the handler/struct level.
- **`idx_profile_state` is unused**: public routes do not filter by `state`; admin routes can filter but rarely do. Index supports `List` queries with `State *bool` set.
- **No soft-delete column**: deletion is hard `DELETE FROM profile WHERE id = $1`. No `deleted_at`.
- **No FK from `section_name` to a sections catalog**: free-form strings. Typos produce orphan sections.

See [concept.md](./concept.md#known-issues) and [endpoint.md](./endpoint.md#known-issues) for application-level issues.

## Related

- [concept.md](./concept.md)
- [endpoint.md](./endpoint.md)
- `db/migrations/00020_create_profile_table.sql`
