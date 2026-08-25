# Desa API — Feature Documentation

Per-feature documentation for the **webdesa/api** Go REST API — a village information system backend.

> **Last updated:** 2026-08-23 (post gallery-migration + signed-URL refactor + docs split)
> **Module:** `webdesa/api`
> **Go:** 1.25.0 · **Router:** chi/v5 · **DB:** PostgreSQL via sqlx · **Auth:** JWT (HS256) + Casbin RBAC
>
> **Major changes since last update:**
>
> - Gallery public + admin binary routes **removed** (Task 7.4); replaced by unified `/api/v1/media/{id}/...?jwt=` signed URLs
> - All per-feature `*_url`/`images[]` legacy columns **dropped** (migration 00045); image refs use `*_media_id` UUIDs
> - All per-feature upload-media endpoints (`POST /<feature>/upload-media`) now route through `galleryusecase.FileStore`
> - PPID `/document/{id}/download` route removed; citizens use signed media
> - `cmd/seed.go` now invokes `galleryService.EnsureSystemFolders` (00043 system folders)
> - **Docs restructured**: each feature now has a subdir with `concept.md` + `database.md` + `endpoint.md`

---

## Architecture

This project follows **Clean Architecture** with strict dependency rules.

```
domain/                → Pure entities, zero external deps (Clean Architecture innermost)
usecase/               → Business logic + Repository interfaces (co-located, Go idiom)
interface/
  postgres/            → ALL sqlx-based repos (no _postgres suffix)
  rbac/                 → Casbin enforcer + PG adapter
  email/                → SMTP adapter (SMTPSender, PPIDEmailService, AuthEmailService)
  file/                 → Local file handler (gallery originals/thumbnails only — per-feature uploads via FileStore)
  http/
    middleware/         → Cross-cutting middleware (auth, RBAC, rate-limit, recovery, etc.)
      middleware_deps.go → MiddlewareDeps + RouteScope (shared by all route.go files)
      rbac.go, ratelimit.go, auth.go, cors.go, logger.go, ...
    handler/
      auth/             → package auth (AuthHandler + route.go)
      user/             → package user
      role/             → package role
      banner/           → package banner
      berita/           → package berita (handler + upload + category + route.go)
      umkm/             → package umkm
      fasilitas/        → package fasilitas
      ppid/             → package ppid (incl. signed-URL approval flow)
      struktur/         → package struktur
      desa/             → package desa (singleton profile via settings.desa_profile)
      profile/          → package profile
      infographic/      → package infographic (Metabase JWT embed)
      health/           → package health
      file/             → package file (legacy /api/v1/files/{filename})
      gallery/          → package gallery (folders + media + signed URLs + orphan sweeper)
    router.go           → Slim. Wires global middleware + scope boundaries
                          → calls each feature's `RegisterRoutes(...)` once per scope
pkg/
  handlerutil/          → ValidateStruct, IsValidUUID, InferContentType, IsDuplicateKeyError
  response/             → JSON response helpers
  pagination/           → Page/limit normalization
  clock/                → Clock interface (RealClock, FixedClock for tests)
  jwt/, password/, slug/, validator/, otel/, profiler/, casbin/
cmd/                    → Cobra CLI (serve, migrate, seed, backup, gallery_gc)
db/migrations/          → 46 SQL migrations (00001–00046; latest 00046 adds `backups.kind`)
rbac/                   → Casbin model + default policy (CSV is reference-only; DB adapter loads at startup)
integration/            → ~27 black-box HTTP test files (testcontainers Postgres)
scripts/swag-tags/      → Post-processor that adds top-level tags array to generated spec
docs/                   → GENERATED OpenAPI: swagger.json, swagger.yaml, docs.go
feature-docs/           → This folder
recom.docs/             → Code recommendations
```

### Per-Feature Package Layout

Every feature has its own Go package under `interface/http/handler/<feature>/`:

```
handler/auth/
  ├── auth.go       (package auth) → AuthHandler struct + handler methods
  └── route.go      (package auth) → RegisterRoutes(...) function
```

`route.go` exports **one** `RegisterRoutes(r chi.Router, deps *MiddlewareDeps, scope RouteScope)` function. When a feature has both public and protected endpoints, the function takes a `scope` parameter (`ScopePublic` or `ScopeProtected`).

`router.go` calls each RegisterRoutes **twice** — once from the public scope, once from the protected scope:

```go
// Public scope (no auth middleware)
banner.RegisterRoutes(r, cfg.BannerHandler, mw, middleware.ScopePublic)

// Protected scope (auth + RBAC + rate-limit already applied to r)
banner.RegisterRoutes(r, cfg.BannerHandler, mw, middleware.ScopeProtected)
```

---

## Per-Feature Documentation

Each feature has its own subdir under `feature-docs/` containing **three** files:

| File | Content |
| --- | --- |
| `concept.md` | Purpose · business rules · status flags · architecture (files, service deps, repo interface) · glossary · cross-refs |
| `database.md` | Migrations (chronological) · tables (columns, types, constraints, FKs, ON DELETE) · indexes · seeds · DB-only known issues |
| `endpoint.md` | Routes (full table) · request/response per endpoint · validation (handler + domain) · RBAC · endpoint-only known issues |

| Feature | Subdir |
| --- | --- |
| Auth | [auth/](./auth/) |
| Banner | [banner/](./banner/) |
| Berita | [berita/](./berita/) |
| Users | [users/](./users/) |
| Roles | [rbac/](./rbac/) |
| UMKM | [umkm/](./umkm/) |
| Fasilitas | [fasilitas/](./fasilitas/) |
| PPID | [ppid/](./ppid/) |
| Struktur | [struktur/](./struktur/) |
| Desa | [desa/](./desa/) |
| Profile | [profile/](./profile/) |
| Infographic | [infographic/](./infographic/) |
| Gallery | [gallery/](./gallery/) |
| File Uploads | [file-uploads/](./file-uploads/) |
| Backup | [backup/](./backup/) |
| Settings | [settings/](./settings/) |

Cross-cutting concerns (RBAC matrix, rate-limit, token strategies, category pattern, file cleanup, UUID strategy) stay at this root level — they are not duplicated per feature.

---

## Routes Overview

All endpoints live under `/api/v1`. Public endpoints add `/public/` prefix or use `/health`, `/auth/*`, `/files/{filename}`.

**Per-feature breakdown** (public + protected, includes category/upload-media where applicable):

| Feature | Public | Protected | Total |
| --- | --- | --- | --- |
| Auth | 5 | — | 5 |
| Banner (+ categories + upload-media) | 3 | 10 | 13 |
| Berita (+ categories + upload-media) | 3 | 8 | 11 |
| Users (+ /me + upload-media) | — | 10 | 10 |
| Roles (+ user-role) | — | 7 | 7 |
| UMKM (+ categories + upload-media) | 3 | 8 | 11 |
| Fasilitas (+ categories + upload-media) | 3 | 9 | 12 |
| PPID (+ categories + requests + uploads; download via signed media) | 4 | 13 | 17 |
| Struktur (+ upload-media) | 2 | 6 | 8 |
| Desa | 1 | 2 | 3 |
| Profile | 2 | 6 | 8 |
| Infographic (+ categories + access-logs) | 3 | 10 | 13 |
| Gallery (admin only; public via signed media) | — | 12 | 12 |
| Signed Media (unified, JWT-gated) | 3 | — | 3 |
| Files | 1 | — | 1 |
| Health | 1 | — | 1 |
| **TOTAL** | **31** | **101** | **132** |

**Notes:**

- Gallery public routes removed (Task 7.4). All gallery media access via `/api/v1/media/{id}/?jwt=` (signed URL) or `/api/v1/public/media/{id}/?jwt=` (public-signed scope).
- PPID download route removed; `service.DownloadDocument` is dead code.
- Backup HTTP handlers implemented but **not mounted** (`backupHandler` commented at `cmd/serve.go:325`); CLI-only.
- Infographic public list silently ignores `q`/`category`/`section_name`/`state` filters — admin list honors them.
- Roles `DELETE /roles/{role}/permissions` is broken (route reads non-existent `{permission}` URL param).
- Users `PUT /users/{id}` and `PUT /users/me` are broken (`interface/postgres/user.go:138` malformed SQL).
- Struktur `Update` SQL is broken (`interface/postgres/struktur.go:139` typo `phone = $4 = $5`).

---

## OpenAPI Spec

Source: **`docs/swagger.yaml`** (auto-generated, **do not edit by hand**)

Generated from Go source annotations on each handler method (`@Summary`, `@Param`, `@Success`, `@Router`, etc.).

### Regenerate

```bash
make swagger-gen
```

Steps performed by the Makefile target:

1. `swag init -g main.go -o docs --parseDependency --parseInternal --parseDepth 10`
2. `go run ./scripts/swag-tags` — injects top-level `tags:` array (swag v1.x does not auto-generate this from `@Tags`)
3. Output: `docs/swagger.json`, `docs/swagger.yaml`, `docs/docs.go`

### Validate

```bash
make swagger-validate
```

Uses `redocly.yaml` config which relaxes rules that don't apply to code-generated specs (e.g. `security-defined`, `operation-summary`).

### Serve via UI

Two options:

1. **Embedded** (recommended) — Swagger UI served by the API itself:

   ```bash
   make serve
   # open http://localhost:8080/swagger/index.html
   ```

2. **Standalone** — Docker container:

   ```bash
   make swagger-up    # http://localhost:8081
   make swagger-down
   ```

---

## Default Seeded Users

```
admin@desa.local     / admin123      (role: admin)
operator@desa.local  / operator123   (role: operator)
```

Created idempotently by `cmd/seed.go`.

---

## RBAC

Casbin enforcer with `g(r.sub, p.role) && (p.obj == "*" || r.obj == p.obj) && (p.act == "*" || r.act == p.act)` matcher.

- **`admin`** — wildcard `{*, *}` for all resources/actions
- **`operator`** — read+write for berita, umkm, fasilitas, ppid, struktur, banner, desa, **gallery**

Permissions stored in `casbin_rule` table (migration 00013). Seeded by `cmd/seed.go:246-288` (`seedRolePermissions`).

> ⚠️ `rbac/rbac_policy.csv` is **NOT loaded** at runtime — the DB-backed `pgAdapter` is the source of truth. The CSV exists for reference only and contains stale entries.
>
> ⚠️ `rbac/model.conf` + `rbac/policy.csv` are a second, incompatible policy pair — never loaded, safe to ignore.
>
> ⚠️ `pkg/casbin/sqlx_adapter.go` has zero importers (dead code).
>
> ⚠️ `rbac_policy.csv` is missing an `admin,infographic,write` row, so admin write on infographic returns 403 until seed is re-run or the row is added.

---

## Default Seeded Categories

| Category | Seed Values | Placeholder |
| --- | --- | --- |
| Banner | Promo, Pengumuman, Event | "Lainnya" (00039) |
| Berita | Berita Desa, Pengumuman, Kegiatan | "Lainnya" (00027) |
| UMKM | Kuliner, Kerajinan, Pertanian, Jasa, Perdagangan | **"Umum"** (00028) |
| Fasilitas | Pendidikan, Kesehatan, Ibadah, Olahraga, Pemerintahan, Pasar | "Lainnya" (00033) |
| PPID | Anggaran, Peraturan, Laporan, Profil, Keuangan | "Tanpa Kategori" (00031) |
| Infographic | Statistik, Keuangan, Kependudukan, Pendidikan, Kesehatan | "Lainnya" (00035) |

**System gallery folders** (created by `galleryService.EnsureSystemFolders` during seed at `cmd/seed.go:136`): `banner`, `berita`, `struktur`, `umkm`, `fasilitas`, `user`, `ppid` — each marked `is_system=true`, immutable via API (403).

---

## Rate Limiting

| Bucket | Limit | Scope | Key |
| --- | --- | --- | --- |
| `auth_requests_per_minute` | 5 | `/auth/*` (also inherits public 100/min) | `r.RemoteAddr` (`IP:port` — bucket per TCP conn, **not** per IP) |
| `public_requests_per_minute` | 100 | `/public/*` | `r.RemoteAddr` |
| `protected_requests_per_minute` | 30 | Authenticated routes | `userID` (from JWT) |

Token-bucket in-memory implementation at `interface/http/middleware/ratelimit.go`. Per-process; for multi-instance, use Redis (see `recom.docs/performance.md`).

> ⚠️ `r.RemoteAddr` includes port → every new TCP connection gets a fresh bucket. The documented per-IP guarantee is **not** enforced.

---

## Cross-Cutting Patterns

### Category Pattern

All 5 category tables share an identical Repository contract (`Create/Delete/FindByID/List/CountByCategoryIDs`) and common error types:

- `CategoryInUseError` → 409
- `CategoryNotFoundError` → 404
- `DuplicateNameError` → 409

Category FKs use `ON DELETE RESTRICT` — categories cannot be deleted while referenced.

### File Cleanup

Every Create/Update/Delete failure path calls `Delete` on saved files to prevent orphans. Multi-image features (UMKM, Fasilitas) iterate prior saves and delete on partial failure.

### UUIDs Everywhere

All entity primary keys are UUIDs via PostgreSQL `gen_random_uuid()`.

### Token Strategies

| Use | Algorithm | Expiry | Notes |
| --- | --- | --- | --- |
| Auth access | JWT HS256 | 24h | from `systemConfig.JWT.GetExpiration()` |
| Auth refresh | JWT HS256 | 7d | |
| Password reset | JWT (none — opaque) | **24h** (mislabeled as "1 hour" in email body) | Email body says 1h; actual TTL = JWT expiry. **Bug**. |
| PPID signed media | JWT HS256 (gallery) | per `signedURL.Sign(ttl)` | scoped `gallery.ScopePPID` + `ppid_request:<id>`; revoke deny-lists sub |
| Gallery signed media (admin) | JWT HS256 (gallery) | configurable | revoke via `signedURL.DenyList` |
| Gallery signed media (public) | JWT HS256 (gallery) | configurable | rate-limited per `gallery.PublicRateLimit` |
| Metabase embed (admin) | JWT HS256 | 10 min | `GenerateMetabaseToken` |
| Metabase embed (public) | JWT HS256 | **5 min** (`PublicTokenTTL`) | `GetPublicAccess` |

---

## Development Setup

```bash
make dev              # db-up + migrate + seed + air hot-reload
make serve           # run server (foreground)
make test-unit       # unit tests
make test-integration # requires Docker (testcontainers)
make test-integration-local INTEGRATION_TEST_DB_URL=...   # against local Postgres
make swagger-gen     # regenerate OpenAPI
make swagger-validate # lint with redocly
```

---

## File Structure

```
api/
├── main.go                     # Entry point — also has swag global info annotations
├── go.mod                      # Module: webdesa/api
├── Makefile
├── Dockerfile
├── docker-compose.yml
├── docker-compose.swagger.yml
├── redocly.yaml                # Linter config for generated OpenAPI
├── .air.toml
├── .env.example
├── config.example.toml
│
├── cmd/                        # Cobra CLI: serve, migrate, seed, backup
├── config/                     # Viper config structs (fileupload, smtp, etc.)
├── domain/                     # 22 pure entity packages
├── usecase/                    # 20 service packages + interfaces
├── interface/
│   ├── postgres/               # ALL sqlx-based repos (no _postgres suffix)
│   ├── rbac/                   # Casbin enforcer
│   ├── email/                  # SMTP (SMTPSender, PPIDEmailService, AuthEmailService)
│   ├── file/                   # Local file handler
│   └── http/
│       ├── middleware/         # Cross-cutting middleware (incl. MiddlewareDeps)
│       ├── handler/<feature>/  # Per-feature packages
│       └── router.go           # Slim root router
│
├── pkg/
│   ├── handlerutil/            # ValidateStruct, IsValidUUID, etc.
│   ├── response/               # JSON response helpers
│   ├── pagination/, clock/, jwt/, password/, slug/, validator/, otel/, profiler/, casbin/
│
├── db/migrations/              # 46 SQL migrations (00001–00046)
├── rbac/                       # Casbin model + reference-only policy CSV
├── integration/                # ~27 black-box HTTP integration test files
├── scripts/swag-tags/          # Post-processor for top-level tags array
│
├── docs/                       # GENERATED OpenAPI spec (do not edit)
├── feature-docs/               # This folder — per-feature subdirs (concept/database/endpoint)
│   ├── README.md               # This file (cross-cutting)
│   ├── auth/{concept,database,endpoint}.md
│   ├── banner/{concept,database,endpoint}.md
│   ├── berita/{concept,database,endpoint}.md
│   ├── users/{concept,database,endpoint}.md
│   ├── rbac/{concept,database,endpoint}.md
│   ├── umkm/{concept,database,endpoint}.md
│   ├── fasilitas/{concept,database,endpoint}.md
│   ├── ppid/{concept,database,endpoint}.md
│   ├── struktur/{concept,database,endpoint}.md
│   ├── desa/{concept,database,endpoint}.md
│   ├── profile/{concept,database,endpoint}.md
│   ├── infographic/{concept,database,endpoint}.md
│   ├── gallery/{concept,database,endpoint}.md
│   ├── file-uploads/{concept,database,endpoint}.md
│   ├── backup/{concept,database,endpoint}.md
│   └── settings/{concept,database,endpoint}.md
└── recom.docs/                 # Code recommendations
```

---

## Contributing

When adding a new feature:

1. **SQL migration** → `db/migrations/00XXX_<name>.sql`
2. **Domain entity** → `domain/<feature>/`
3. **Repository interface** → `usecase/<feature>/`
4. **Use case service** → `usecase/<feature>/service.go`
5. **Repository impl** → `interface/postgres/<feature>.go`
6. **Handler package** → `interface/http/handler/<feature>/`
   - `<feature>.go` — handler struct + methods
   - `route.go` — `RegisterRoutes(r, deps, scope)` function
7. **Swag annotations** on handler methods (`@Summary`, `@Tags`, `@Router`, etc.)
8. **Register route** in `interface/http/router.go`
9. **RBAC permission** in seed.go (if applicable)
10. **Tests**:
    - Unit: `usecase/<feature>/service_test.go`
    - Integration: `integration/<feature>_test.go`
11. **Regenerate OpenAPI**: `make swagger-gen && make swagger-validate`
12. **Documentation** (if feature is non-trivial):
    - `feature-docs/<feature>/concept.md` — purpose + business rules + architecture
    - `feature-docs/<feature>/database.md` — migrations + schema + indexes + seeds
    - `feature-docs/<feature>/endpoint.md` — routes + request/response + RBAC + validation
