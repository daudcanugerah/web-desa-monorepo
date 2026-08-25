# Performance Audit — 2026 Q3

Original: `../recom.docs/performance.md` (8 items in summary table; 6 numbered issues).

| # | Priority | Issue | Status | Evidence |
|---|---|---|---|---|
| 1 | HIGH | In-memory rate limiter | `[STILL OPEN]` | `interface/http/middleware/ratelimit.go` |
| 2 | MEDIUM | No HTTP cache headers | `[PARTIAL]` | Gallery only (`gallery.go:1263,1265`, `signed_media.go:139,141`) |
| 3 | MEDIUM | No DB query caching | `[STILL OPEN]` | No LRU/Redis in repos |
| 4 | MEDIUM | Missing DB indexes | `[RESOLVED]` | Migration 00047 adds composite + pg_trgm |
| 5 | MEDIUM | N+1 queries | `[PARTIAL]` | UMKM prune fixed (Task 5.1); still per-row |
| 6 | LOW | Inconsistent pagination | `[RESOLVED]` | `/banners/active` 100 cap documented in OpenAPI |
| 7 | LOW | File upload streaming | `[RESOLVED]` | `galleryusecase.FileStore.SaveImage` streams |
| 8 | LOW | Connection pool tuning | `[SUPERSEDED]` | Config-driven (`cmd/serve.go:90-92`) |

---

## 1. In-Memory Rate Limiter — `[STILL OPEN]`

**Original claim:** Token-bucket per-process; multi-instance effective limit = N× configured.

**Current state:** `interface/http/middleware/ratelimit.go:18-56, 119-141` still uses `sync.RWMutex` + in-memory `map[string]*tokenBucket`. No `go-redis` dep in `go.mod`.

**Action:** Adopt `redis_rate` or similar. High impact for multi-instance deployments.

---

## 2. HTTP Cache Headers — `[PARTIAL]`

**Original claim:** Public endpoints have no cache headers.

**Current state:** Only gallery endpoints set `Cache-Control`:
- `interface/http/handler/gallery/gallery.go:1263,1265` — `public, max-age=300` and `private, max-age=300`
- `interface/http/handler/gallery/signed_media.go:139,141` — same

Public endpoints (`/public/banner`, `/public/berita`, etc.) have no cache headers.

**Action:** Add `Cache-Control` per resource type per original doc recommendations (desa 300s, banner 60s, berita 60s, ppid 300s).

---

## 3. No DB Query Caching — `[STILL OPEN]`

**Original claim:** Common queries hit DB every request.

**Current state:** `grep -r "lru.Cache\|redis" usecase/ interface/postgres/` returns 0 hits. No caching layer.

**Action:** LRU for hot reads (desa profile, active banners) with explicit invalidation.

---

## 4. Missing DB Indexes — `[RESOLVED]`

**Original claim:** No composite indexes; no `pg_trgm` GIN.

**Fix applied (commit 50c6253):** `db/migrations/00047_add_performance_indexes.sql` adds:
- `(category, created_at DESC)` composite on berita, umkm, fasilitas, ppid, ppid_requests
- GIN `pg_trgm` indexes on berita.title, umkm.name, fasilitas.name
- `pg_trgm` extension (created idempotently; requires superuser or managed-Postgres parameter group)

Partial index `WHERE publication_at IS NOT NULL` deferred — verify usage first.

---

## 5. N+1 Queries — `[PARTIAL]`

**Original claim:** Some list endpoints fetch related data per row.

**Current state:** `usecase/umkm/service.go:286-291` now prunes dangling media IDs (Gallery OnDelete sweep, Task 5.1) but uses per-row `fileStore.Open` — still N calls. Original `role.go` user-role N+1 not visible in current code (likely fixed via JOIN).

**Action:** Audit remaining N+1 patterns. Add `EXPLAIN ANALYZE` to integration tests for hot endpoints.

---

## 6. Pagination — `[RESOLVED]`

**Original claim:** Some lists lack pagination.

**Current state:** Most use `pagination.Paginate(...)`. `/api/v1/banners/active` (`handler/banner/banner.go:362-368`) returns up to 100 hardcoded; OpenAPI annotation `@Description Public endpoint returning up to 100 active banners` already documents the cap. Per `feature-docs/banner/database.md`, the DB trigger `enforce_active_banner_limit()` caps at 20 active, so the 100 limit only matters if the trigger is removed.

---

## 7. File Upload Streaming — `[RESOLVED]`

**Original claim:** `SaveImage` reads entire content into memory.

**Current state:** `usecase/gallery/filestore.go` + `usecase/umkm/service.go:124-138` use streaming via `fileStore.SaveImage` — no `io.ReadAll`.

---

## 8. Connection Pool Tuning — `[SUPERSEDED]`

**Original claim:** `MaxOpenConns: 25, MaxIdleConns: 5, ConnMaxLifetime: 5m` hardcoded.

**Current state:** `cmd/serve.go:90-92` uses config:
```go
db.SetMaxOpenConns(systemConfig.Postgres.GetMaxOpenConns())
```
Config-driven via `[postgres]` section. Tuning done via `config.toml` / env vars.
