# Infographic — Database

**Module:** `webdesa/api`
**Last updated:** 2026-08-23

## Migrations

| # | Purpose |
|---|---|
| `00021` | Create `infographic` table |
| `00022` | Add `state` boolean default `true` |
| `00023` | Index on `component_type` |
| `00035` | Create `infographic_categories` + seed `"Lainnya"` |
| `00036` | Add `category` FK → `infographic_categories(id)` (nullable) |
| `00040` | Create `infographic_access_log` |

## Tables

### `infographic` (migrations 00021, 00022, 00023, 00036)

| Column | Type | Nullable | Constraints / FK |
|---|---|---|---|
| `id` | UUID | NO | PK, `gen_random_uuid()` |
| `component_id` | BIGINT | NO | DEFAULT 0 — the Metabase component ID |
| `component_type` | VARCHAR(50) | NO | DEFAULT `'dashboard'` — `dashboard` or `question` |
| `section_name` | VARCHAR(100) | NO | |
| `section_endpoint` | VARCHAR(255) | NO | URL path, must start with `/` |
| `category` | UUID | YES | FK → `infographic_categories(id)` ON DELETE RESTRICT |
| `state` | BOOLEAN | NO | DEFAULT `true` — added 00022 |
| `created_at` | TIMESTAMP | NO | |
| `updated_at` | TIMESTAMP | NO | |

### `infographic_categories` (migration 00035)

Standard category table (UUID PK + name + timestamps). Seeded with `"Lainnya"`.

### `infographic_access_log` (migration 00040)

| Column | Type | Nullable | Notes |
|---|---|---|---|
| `id` | UUID | NO | PK, `gen_random_uuid()` |
| `infographic_id` | UUID | NO | FK → `infographic(id)` |
| `component_id` | BIGINT | NO | snapshot at log time |
| `component_type` | VARCHAR(50) | NO | snapshot at log time |
| `endpoint` | VARCHAR(255) | NO | `section_endpoint` resolved at request time |
| `ip_address` | INET | NO | |
| `user_agent` | TEXT | YES | |
| `referer` | TEXT | YES | |
| `token_issued_at` | TIMESTAMP | NO | |
| `token_expires_at` | TIMESTAMP | NO | |
| `created_at` | TIMESTAMP | NO | DEFAULT `NOW()` |

No FK on `infographic_id` CASCADE/SET NULL is enforced — log rows outlive deleted infographics (intentional audit trail).

## Indexes

- `idx_infographic_component_type` on `infographic(component_type)` — added 00023
- `idx_infographic_section_name` (implicit, used by `GetSectionNames`)
- `idx_infographic_access_log_infographic_id` on `infographic_access_log(infographic_id)`
- `idx_infographic_access_log_ip` on `infographic_access_log(ip_address)`
- `idx_infographic_access_log_created_at` on `infographic_access_log(created_at)`

## Seeds

`cmd/seed.go:286-292` seeds `infographic.section_name` (or categories, depending on which table the seed targets — verify against current `cmd/seed.go`):

```
"Statistik", "Keuangan", "Kependudukan", "Pendidikan", "Kesehatan"
```

> **Drift to verify:** the monolithic doc lists section names as seed content, but `cmd/seed.go:286-292` historically targets `infographic_categories`. Confirm before re-running seeds in a fresh env.

## Known Issues (DB-specific)

- **`category` column is nullable** (preserved from 00036). Domain accepts it; downstream consumers must tolerate null. See `recom.docs/bugs.md#8`.
- **No DB-level CHECK constraint on `component_type`** — invalid values caught only at domain `Validate()`.
- **No FK ON DELETE behavior specified for `infographic_access_log.infographic_id`** — orphans persist if `infographic` rows are deleted (by design, for audit).
- **No DB-level rate-limiting** — enforced in application layer via `rateLimiter`.

## Related

- `feature-docs/infographic/concept.md`
- `feature-docs/infographic/endpoint.md`
