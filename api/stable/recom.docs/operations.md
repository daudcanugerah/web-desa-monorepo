# Operations Recommendations

Deployment, observability, CI/CD, dev experience.

---

## 1. Go Version Drift 🟠 HIGH

**Location:** `go.mod:2`, `Dockerfile`

**Issue:**

```
go.mod:    go 1.25.0
Dockerfile: golang:1.24.1-alpine
```

**Impact:**
- Local builds may use newer features than Dockerfile
- CI may use different version than production

**Recommendation:**

1. **Use `golang-version-file` in CI**: read version from `go.mod`
2. **Pin Dockerfile** to match:
   ```dockerfile
   FROM golang:1.25.0-alpine AS builder
   ```
3. **Add CI check** that Dockerfile version matches go.mod:
   ```bash
   REQUIRED=$(grep '^go ' go.mod | awk '{print $2}')
   ACTUAL=$(grep '^FROM golang' Dockerfile | grep -oP 'golang:\K[0-9.]+')
   if [ "$REQUIRED" != "$ACTUAL" ]; then exit 1; fi
   ```

**Effort:** XS (1 hour)

---

## 2. No CI/CD Pipeline 🟠 HIGH

**Location:** Repo (no `.github/`, `.gitlab-ci.yml`, etc.)

**Issue:**

No automated CI/CD. Tests, lint, deploy are manual.

**Recommendation:**

Create `.github/workflows/ci.yml`:
```yaml
name: CI

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
          cache: true
      
      - name: Install dev tools
        run: |
          go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
      
      - name: Lint
        run: make lint
      
      - name: Unit tests
        run: make test-unit
      
      - name: Integration tests
        run: make test-integration
        env:
          DATABASE_URL: postgres://test:test@localhost:5432/test
      
      - name: Build
        run: make build
      
      - name: Docker build
        run: docker build -t webdesa-api:${{ github.sha }} .
  
  deploy:
    needs: test
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'
    steps:
      - uses: actions/checkout@v4
      - name: Deploy
        run: |
          # Deploy commands here
```

**Effort:** S (1-2 days)

---

## 3. No Health Check Endpoint Standardization 🟡 MEDIUM

**Location:** `interface/http/handler/health.go`

**Issue:**

Health check exists at `/health` but may only return 200 without verifying dependencies.

**Recommendation:**

Multi-level health checks:
```go
// Simple liveness: process is alive
GET /health/liveness → 200 OK

// Readiness: can handle requests
GET /health/readiness
→ 200 OK with:
{
    "status": "healthy",
    "checks": {
        "database": {"status": "up", "latency_ms": 5},
        "smtp": {"status": "configured", "reachable": true},
        "metabase": {"status": "unknown", "url": "..."},
        "files": {"status": "writable"}
    },
    "version": "1.2.3",
    "uptime": "1h23m"
}
```

Plus use for Kubernetes:
```yaml
livenessProbe:
  httpGet:
    path: /health/liveness
    port: 8080
readinessProbe:
  httpGet:
    path: /health/readiness
    port: 8080
```

**Effort:** S (1 day)

---

## 4. OpenTelemetry Not Fully Utilized 🟡 MEDIUM

**Location:** `pkg/otel/`, `cmd/serve.go`

**Issue:**

OTel is integrated (`pkg/otel` setup in `cmd/root.go`) but it's unclear if all HTTP handlers have tracing spans, all DB queries are instrumented, etc.

**Recommendation:**

1. **Add trace spans per request:**
   ```go
   r.Use(otelgin.Middleware("desa-api"))  // or chi equivalent
   ```

2. **Instrument all DB queries:**
   ```go
   db, err := otelsqlx.Open("postgres", dsn, otelsqlx.WithAttributes(...))
   ```

3. **Add custom spans in business logic:**
   ```go
   ctx, span := otel.Tracer("ppid").Start(ctx, "ApproveRequest")
   defer span.End()
   ```

4. **Verify metrics are exported:**
   - Request rate
   - Error rate
   - DB pool utilization
   - Response time percentiles

**Effort:** M (3-5 days)

---

## 5. pprofiler Not Exposed 🟡 MEDIUM

**Location:** `pkg/profiler/`

**Issue:**

Profiling is set up via `pkg/profiler` but there's no way to access profiles externally.

**Recommendation:**

Add a debugging-only endpoint:
```go
import _ "net/http/pprof"

// Only register in dev/staging
if !config.IsProduction() {
    r.HandleFunc("/debug/pprof/", pprof.Index)
    r.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
    r.HandleFunc("/debug/pprof/profile?seconds=30", pprof.Profile)
    r.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
    r.HandleFunc("/debug/pprof/trace", pprof.Trace)
}
```

Protect behind admin RBAC.

**Effort:** XS (1 hour)

---

## 6. Docker Image Not Optimized 🟡 MEDIUM

**Location:** `Dockerfile`

**Issue:**

Currently uses multi-stage Alpine build. Verify optimization:
- Distroless base image?
- Reproducible builds?
- CVE scanning?

**Recommendation:**

Verify Dockerfile:
- ✅ Multi-stage build (already done)
- ✅ Runs as non-root (already done)
- ✅ Stripped binary (already done)
- Add `HEALTHCHECK` (already done)
- Add `.dockerignore`:
  ```
  coverage.out
  coverage_handler.out
  coverage_ppid.out
  docs/
  recom.docs/
  .git/
  .vscode/
  tmp/
  ```
- Use BuildKit cache mounts:
  ```dockerfile
  RUN --mount=type=cache,target=/go/pkg/mod \
      --mount=type=cache,target=/root/.cache/go-build \
      go build ...
  ```
- Pin base image by SHA, not tag:
  ```dockerfile
  FROM golang:1.25.0-alpine@sha256:... AS builder
  ```

**Effort:** S (1 day)

---

## 7. No Database Backup Strategy 🟡 MEDIUM

**Location:** `cmd/backup.go`, ops

**Issue:**

Backup CLI exists but:
- No scheduled backups (cron)
- No S3/cloud upload
- No restoration drill documentation

**Recommendation:**

1. **Schedule cron:**
   ```cron
   0 2 * * * /opt/desa-api/desa-api backup create
   ```

2. **Upload to S3:**
   ```bash
   aws s3 cp /var/backups/desa/$(date +\%Y\%m\%d).sql s3://desa-backups/
   ```

3. **Add backup verification:**
   ```bash
   pg_restore --list backup.sql  # verify file
   ```

4. **Document restoration drill:**
   ```markdown
   # Disaster Recovery
   
   1. Spin up new server
   2. Run migrations
   3. Download latest backup from S3
   4. Run `desa-api backup restore <id>`
   5. Verify checksum
   6. Smoke test endpoints
   ```

**Effort:** M (1 week)

---

## 8. No Graceful Shutdown Verification 🟢 LOW

**Location:** `cmd/serve.go`

**Issue:**

Graceful shutdown exists (`SIGINT`/`SIGTERM` with 30s timeout) but consider:
- Are in-flight requests waited for?
- Are connections to DB closed properly?
- Are background workers stopped?

**Recommendation:**

Verify and add:
```go
// Track in-flight requests
inflight := sync.WaitGroup{}
r.Use(func(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        inflight.Add(1)
        defer inflight.Done()
        next.ServeHTTP(w, r)
    })
})

// On shutdown:
go func() {
    sigint := make(chan os.Signal, 1)
    signal.Notify(sigint, syscall.SIGINT, syscall.SIGTERM)
    <-sigint
    
    // Stop accepting new requests
    server.Shutdown(ctx)
    
    // Wait for in-flight to complete
    inflight.Wait()
    
    // Close DB
    db.Close()
    
    // Stop background workers
}()
```

**Effort:** S (1 day)

---

## Summary Table

| # | Issue | Priority | Effort |
|---|---|---|---|
| 1 | Go version drift | 🟠 HIGH | XS |
| 2 | No CI/CD | 🟠 HIGH | S |
| 3 | Health check standardization | 🟡 MEDIUM | S |
| 4 | OTel not fully utilized | 🟡 MEDIUM | M |
| 5 | pprofiler not exposed | 🟡 MEDIUM | XS |
| 6 | Docker optimization | 🟡 MEDIUM | S |
| 7 | DB backup strategy | 🟡 MEDIUM | M |
| 8 | Graceful shutdown | 🟢 LOW | S |

## Recommended Stack Additions

- **CI/CD:** GitHub Actions (already suggested)
- **Container registry:** GitHub Container Registry or AWS ECR
- **Secret management:** HashiCorp Vault or AWS Secrets Manager
- **Monitoring:** Prometheus + Grafana
- **Logging:** ELK stack or Loki
- **Tracing:** Jaeger or Tempo
- **Error tracking:** Sentry
- **Uptime monitoring:** Better Uptime or UptimeRobot
- **Performance:** k6 load testing in CI