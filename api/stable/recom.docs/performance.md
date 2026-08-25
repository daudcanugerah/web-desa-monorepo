# Performance Recommendations

Performance optimizations, caching, query efficiency, scalability.

---

## 1. In-Memory Rate Limiter Doesn't Scale 🟠 HIGH

**Location:** `interface/http/middleware/ratelimit.go`

**Issue:**

The token-bucket rate limiter is in-memory per-process. With multiple API instances behind a load balancer, each instance has its own bucket. Effective limit = `N × configured_limit`.

**Impact:**
- Rate limits effectively N times higher than configured
- Some users getting throttled, others not

**Recommendation:**

**Option A — Redis-backed limiter (preferred for multi-instance):**
```go
import "github.com/go-redis/redis_rate/v10"

func RateLimitMiddleware(redis *redis.Client) func(http.Handler) http.Handler {
    limiter := redis_rate.NewLimiter(redis)
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            key := getKey(r)
            res, err := limiter.Allow(r.Context(), key, redis_rate.PerMinute(30))
            if err != nil {
                next.ServeHTTP(w, r)
                return
            }
            if res.Allowed == 0 {
                http.Error(w, "rate limit exceeded", 429)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}
```

**Option B — Sliding window via SQL:**
```sql
SELECT COUNT(*) FROM rate_limit_log
WHERE key = $1 AND created_at > NOW() - INTERVAL '1 minute'
```

**Effort:** M (2-3 days)

---

## 2. No HTTP Response Caching Headers 🟡 MEDIUM

**Location:** All handlers

**Issue:**

Public endpoints (e.g., `/api/v1/public/banners/active`) return same data for many users but have no cache headers. Every request hits the DB.

**Recommendation:**

Add cache headers for public endpoints:
```go
func PublicListHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Cache-Control", "public, max-age=60")  // 1 min
    w.Header().Set("ETag", generateETag(data))
    // ... serve content ...
}
```

For content that changes:
- Village profile: `max-age=300` (5 min)
- Active banners: `max-age=60`
- News list: `max-age=60`
- PPID list: `max-age=300`

**Effort:** S (1 day)

---

## 3. No Database Query Caching 🟡 MEDIUM

**Location:** Repositories

**Issue:**

Common queries (e.g., "list active banners", "village profile") hit DB every request.

**Recommendation:**

Add in-memory or Redis cache layer:
```go
type CachedBannerRepository struct {
    repo  BannerRepository
    cache *lru.Cache[string, []Banner]
    ttl   time.Duration
}

func (r *CachedBannerRepository) ListActive(ctx) ([]Banner, error) {
    if cached, ok := r.cache.Get("active"); ok {
        return cached, nil
    }
    banners, err := r.repo.ListActive(ctx)
    if err != nil {
        return nil, err
    }
    r.cache.Add("active", banners)
    return banners, nil
}
```

Invalidate on Update/Delete.

**Effort:** M (2-3 days)

---

## 4. Missing Database Indexes on Common Queries 🟡 MEDIUM

**Location:** `db/migrations/`

**Issue:**

Common queries may not be indexed:

```sql
-- berita: search by query, filter by category + created_at
SELECT ... FROM berita
WHERE (title ILIKE $1 OR content ILIKE $1)
  AND category = $2
  AND created_at BETWEEN $3 AND $4
```

Current index: `idx_berita_created_at DESC`. Missing composite.

**Recommendation:**

Add migration `db/migrations/00037_add_performance_indexes.sql`:
```sql
-- Composite indexes for common queries
CREATE INDEX idx_berita_category_created_at 
    ON berita (category, created_at DESC);

CREATE INDEX idx_berita_title_trgm 
    ON berita USING gin (title gin_trgm_ops);

CREATE INDEX idx_published_berita 
    ON berita (created_at DESC) 
    WHERE publication_at IS NOT NULL;

-- Similarly for other features
CREATE INDEX idx_umkm_category_created_at 
    ON umkm (category, created_at DESC);

CREATE INDEX idx_fasilitas_category_location 
    ON fasilitas (category, latitude, longitude);

-- PPID by status
CREATE INDEX idx_ppid_requests_status_created 
    ON ppid_requests (status, created_at DESC);
```

**Requires extension:** `CREATE EXTENSION IF NOT EXISTS pg_trgm;`

**Effort:** S (1 day)

---

## 5. N+1 Queries in List Endpoints 🟡 MEDIUM

**Location:** Several repositories

**Issue:**

Some list endpoints may fetch N+1 queries for related data:

```go
// List banners with metadata — already in same row, OK
// But list users with roles:
for _, user := range users {
    roles, _ := userRepo.GetRoles(user.ID)  // N+1!
}
```

**Recommendation:**

Use JOINs or batch queries:
```sql
-- Single query with subquery for roles
SELECT 
    u.*, 
    COALESCE(array_agg(ur.role) FILTER (WHERE ur.role IS NOT NULL), '{}') as roles
FROM users u
LEFT JOIN user_roles ur ON ur.user_id = u.id
GROUP BY u.id
ORDER BY u.created_at DESC
LIMIT $1 OFFSET $2
```

**Effort:** M (per endpoint, ~1 day each)

---

## 6. No Pagination on Some List Endpoints 🟢 LOW

**Location:** Select handlers

**Issue:**

`/api/v1/public/banners/active` may return up to 100 banners. If village has more, they're cut off silently.

**Recommendation:**

Ensure all list endpoints have explicit pagination defaults and max limits. Already mostly done — verify each.

**Effort:** XS (audit + fix)

---

## 7. File Uploads Not Streamed 🟢 LOW

**Location:** `interface/file/local_handler.go`

**Issue:**

`SaveImage` reads entire content into memory via `io.Copy`. For 50MB documents, this doubles memory usage.

**Recommendation:**

Already uses streaming via `io.Copy` — verify no buffered reads elsewhere.

**Effort:** XS (audit)

---

## 8. No Connection Pooling Verified 🟢 LOW

**Location:** `cmd/serve.go`

**Issue:**

Database connection pool config:
```go
MaxOpenConns:    25
MaxIdleConns:    5
ConnMaxLifetime: 5m
```

For production traffic, may need tuning.

**Recommendation:**

Profile and tune:
```go
// For high-traffic:
db.SetMaxOpenConns(100)
db.SetMaxIdleConns(25)
db.SetConnMaxLifetime(30 * time.Minute)
```

**Effort:** XS (config only)

---

## Summary Table

| # | Issue | Priority | Effort |
|---|---|---|---|
| 1 | In-memory rate limiter (multi-instance) | 🟠 HIGH | M |
| 2 | No HTTP cache headers | 🟡 MEDIUM | S |
| 3 | No DB query caching | 🟡 MEDIUM | M |
| 4 | Missing indexes | 🟡 MEDIUM | S |
| 5 | N+1 queries | 🟡 MEDIUM | M |
| 6 | Inconsistent pagination | 🟢 LOW | XS |
| 7 | Streaming verification | 🟢 LOW | XS |
| 8 | Connection pool tuning | 🟢 LOW | XS |

## Profiling Recommendations

1. Add `pprof` endpoint in non-production:
   ```go
   import _ "net/http/pprof"
   go func() {
       log.Println(http.ListenAndServe("localhost:6060", nil))
   }()
   ```
2. Use `go test -bench` for performance regression detection
3. Add k6 or vegeta load testing scripts
4. Monitor with Prometheus + Grafana (already have OTel infrastructure)