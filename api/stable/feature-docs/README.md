# Desa API — Feature Documentation

Per-feature documentation for the **webdesa/api** Go REST API — a village information system backend.

> **Last updated:** 2026-07-04 (post-refactor freeze)
> **Module:** `webdesa/api`
> **Go:** 1.25.0 · **Router:** chi/v5 · **DB:** PostgreSQL via sqlx · **Auth:** JWT (HS256) + Casbin RBAC

---

## Architecture

This project follows **Clean Architecture** with strict dependency rules.

```
domain/                → Pure entities, zero external deps (Clean Architecture innermost)
usecase/               → Business logic + Repository interfaces (co-located, Go idiom)
interface/
  postgres/            → ALL sqlx-based repos (19 files, no _postgres suffix)
  rbac/                 → Casbin enforcer + PG adapter
  email/                → SMTP adapter (SMTPSender, PPIDEmailService, AuthEmailService)
  file/                 → Local file handler
  http/
    middleware/         → Cross-cutting middleware (auth, RBAC, rate-limit, recovery, etc.)
      middleware_deps.go → MiddlewareDeps + RouteScope (shared by all route.go files)
      rbac.go, ratelimit.go, auth.go, cors.go, logger.go, ...
    handler/
      auth/             → package auth (AuthHandler + route.go)
      user/             → package user
      role/             → package role
      banner/           → package banner
      berita/           → package berita (3 handler files + route.go)
      umkm/             → package umkm
      fasilitas/        → package fasilitas
      ppid/             → package ppid
      struktur/         → package struktur
      desa/             → package desa
      profile/          → package profile
      infographic/      → package infographic
      health/           → package health
      file/             → package file
    router.go           → Slim (~149 lines). Wires global middleware + scope boundaries
                          → calls each feature's `RegisterRoutes(...)` once per scope
pkg/
  handlerutil/          → ValidateStruct, IsValidUUID, InferContentType, IsDuplicateKeyError
  response/             → JSON response helpers
  pagination/           → Page/limit normalization
  clock/                → Clock interface (RealClock, FixedClock for tests)
  jwt/, password/, slug/, validator/, otel/, profiler/, casbin/
cmd/                    → Cobra CLI (serve, migrate, seed, backup)
db/migrations/          → 38 SQL migrations
rbac/                   → Casbin model + default policy
integration/            → 24 integration test files
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

## Routes Overview (77 paths)

All endpoints live under `/api/v1`. Public endpoints add `/public/` prefix or use `/health`, `/auth/*`, `/files/{filename}`.

| Feature | Public | Protected | Total |
|---|---|---|---|
| Auth | 5 | — | 5 |
| Banner | 3 | 5 | 8 |
| Berita (+ category + upload) | 3 | 7 | 10 |
| Users | — | 8 | 8 |
| Roles (+ user-role) | — | 7 | 7 |
| UMKM (+ category) | 3 | 4 | 7 |
| Fasilitas (+ category) | 3 | 8 | 11 |
| PPID (+ category + request + download) | 5 | 8 | 13 |
| Struktur | 2 | 5 | 7 |
| Desa | 1 | 2 | 3 |
| Profile | 2 | 6 | 8 |
| Infographic (+ category) | 3 | 6 | 9 |
| Files | 1 | — | 1 |
| Health | 1 | — | 1 |
| **TOTAL** | **32** | **67** | **99** |

(Counts include each unique path × HTTP method; backups are CLI-only and not in HTTP.)

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
- **`operator`** — read+write for berita, umkm, fasilitas, ppid, struktur, banner, desa

Permissions stored in `casbin_rule` table (migration 00013).

---

## Default Seeded Categories

| Category | Seed Values |
|---|---|
| Berita | Berita Desa, Pengumuman, Kegiatan |
| UMKM | Kuliner, Kerajinan, Pertanian, Jasa, Perdagangan |
| Fasilitas | Pendidikan, Kesehatan, Ibadah, Olahraga, Pemerintahan, Pasar |
| PPID | Anggaran, Peraturan, Laporan, Profil, Keuangan |
| Infographic | Statistik, Keuangan, Kependudukan, Pendidikan, Kesehatan |

Each has a `"Lainnya"` / `"Umum"` / `"Tanpa Kategori"` placeholder from FK-conversion migrations.

---

## Rate Limiting

| Bucket | Limit | Scope |
|---|---|---|
| `auth_requests_per_minute` | 5 | `/auth/*` |
| `public_requests_per_minute` | 100 | `/public/*` |
| `protected_requests_per_minute` | 30 | Authenticated routes |

Token-bucket in-memory implementation at `interface/http/middleware/ratelimit.go`. Per-process; for multi-instance, use Redis (see `recom.docs/performance.md`).

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

| Use | Algorithm | Expiry |
|---|---|---|
| Auth access | JWT HS256 | 24h |
| Auth refresh | JWT HS256 | 7d |
| PPID download | JWT HS256 | 10 min |
| Metabase embed | JWT HS256 | 10 min |

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
├── domain/                     # 15 pure entity packages
├── usecase/                    # 15 service packages + interfaces
├── interface/
│   ├── postgres/               # ALL sqlx-based repos (19 files, no _postgres suffix)
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
├── db/migrations/              # 38 SQL migrations
├── rbac/                       # Casbin model + default policy
├── integration/                # 24 integration test files
├── scripts/swag-tags/          # Post-processor for top-level tags array
│
├── docs/                       # GENERATED OpenAPI spec (do not edit)
├── feature-docs/               # This folder
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