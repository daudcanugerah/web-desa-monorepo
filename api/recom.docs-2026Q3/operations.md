# Operations Audit — 2026 Q3

Original: `../recom.docs/operations.md` (8 items, audit 2026-07-04).

| # | Priority | Issue | Status | Evidence |
|---|---|---|---|---|
| 1 | HIGH | Go version drift | `[RESOLVED]` | `Dockerfile:2` = `golang:1.25-alpine` |
| 2 | HIGH | No CI/CD pipeline | `[RESOLVED]` | `.github/workflows/ci.yml` (5 jobs) |
| 3 | MEDIUM | Health check standardization | `[PARTIAL]` | `/health` returns 200; no DB ping / readiness split |
| 4 | MEDIUM | OTel not fully utilized | `[PARTIAL]` | `otelhttp` indirect; not in router middleware |
| 5 | MEDIUM | pprofiler not exposed | `[STILL OPEN]` | No `/debug/pprof/*` HTTP route |
| 6 | MEDIUM | Docker image optimization | `[PARTIAL]` | Multi-stage + non-root + HEALTHCHECK; no SHA pin |
| 7 | MEDIUM | DB backup strategy | `[PARTIAL]` | CLI exists; no cron / S3 / restoration drill |
| 8 | LOW | Graceful shutdown | `[PARTIAL]` | `http.Server.Shutdown(ctx)` with 30s timeout; no WaitGroup |

---

## 1. Go Version Drift — `[RESOLVED]`

**Original claim:** `go.mod` 1.25.0 vs `Dockerfile` 1.24.1.

**Current state:** `Dockerfile:2` = `FROM golang:1.25-alpine AS builder`. `go.mod:3` = `go 1.25.0`. Aligned.

---

## 2. No CI/CD Pipeline — `[RESOLVED]`

**Original claim:** No `.github/`, `.gitlab-ci.yml`, etc.

**Fix applied (commit 40768f5):** `.github/workflows/ci.yml` defines 5 jobs gated on lint success:
- **lint** — golangci-lint-action with `.golangci.yml` config
- **unit** — `make test-unit` + `make test-coverage-check` (80% threshold)
- **build** — `make build` + `go vet ./...`
- **integration** — `make test-integration` against `postgres:16-alpine` service
- **docker** — BuildKit build with GHA cache

`ALLOW_INSECURE_SECRETS=1` env var lets `config.InitConfig` validation accept test secrets.

---

## 3. Health Check Standardization — `[PARTIAL]`

**Original claim:** `/health` may only return 200 without verifying deps.

**Current state:** `interface/http/handler/health/health.go:24-28` `HealthCheck` returns `{"status": "ok"}` only. No DB ping, no SMTP check, no liveness/readiness split.

**Action:** Add `/health/liveness` and `/health/readiness` per original doc.

---

## 4. OTel Not Fully Utilized — `[PARTIAL]`

**Original claim:** OTel setup unclear; HTTP handlers may not be instrumented.

**Current state:**
- `pkg/otel/tracer.go:18` exposes `otel.Tracer(name)`
- `go.mod` has `otelhttp v0.54.0` as `// indirect` dep
- NO `r.Use(otelhttp.NewMiddleware(...))` in `interface/http/router.go:85-94`

Tracer helper exists but unused for HTTP middleware. DB queries likely uninstrumented.

**Action:** Wire `otelhttp.NewMiddleware` into router. Use `otelsqlx` for DB instrumentation.

---

## 5. pprofiler Not Exposed — `[STILL OPEN]`

**Original claim:** Profiling setup but no external access.

**Current state:** `pkg/profiler/profiler.go:15-34` uses `bygui86/multi-profile` via CLI flags (`cmd/root.go:122-124` `--run-profile`, `--run-realtime-profile`). No `/debug/pprof/*` HTTP route.

**Action:** Add dev-only route per original doc, behind admin RBAC.

---

## 6. Docker Image Optimization — `[PARTIAL]`

**Original claim:** Verify Dockerfile optimization.

**Current state:**
- `Dockerfile:67-68` — HEALTHCHECK ✓
- `Dockerfile:38-39` — non-root user ✓
- `Dockerfile:29` — stripped binary ✓
- `.dockerignore` exists ✓
- NOT pinned by SHA — `FROM golang:1.25-alpine` uses tag
- No BuildKit cache mounts

**Action:** Pin base by SHA + add cache mounts.

---

## 7. DB Backup Strategy — `[PARTIAL]`

**Original claim:** CLI exists but no cron, S3, restoration drill.

**Current state:**
- `cmd/backup.go` — CLI works
- `Makefile:132-159` — `backup-create`/`backup-list`/`backup-restore`
- No cron entry
- No S3/cloud upload script
- No restoration drill doc

**Action:** Add cron, S3 upload, and DR runbook.

---

## 8. Graceful Shutdown — `[PARTIAL]`

**Original claim:** SIGINT/SIGTERM with 30s timeout, but in-flight tracking?

**Current state:** `cmd/serve.go:382-403` does signal-driven `server.Shutdown(ctx)` with 30s timeout. No explicit `sync.WaitGroup` for in-flight tracking. `http.Server.Shutdown` waits for active connections by design.

**Action:** Optional — add `sync.WaitGroup` if explicit drain logging needed.
