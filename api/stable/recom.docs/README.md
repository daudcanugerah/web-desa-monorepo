# Recommendations

Prioritized recommendations for the **webdesa/api** codebase based on comprehensive analysis.

## Priority Legend

- 🔴 **CRITICAL** — Security, data loss, or production-blocking issues. Fix immediately.
- 🟠 **HIGH** — Correctness bugs, broken features, or significant code smells. Fix soon.
- 🟡 **MEDIUM** — Quality, maintainability, performance improvements. Plan within sprint.
- 🟢 **LOW** — Nice-to-have, polish, future-proofing. Backlog.

## Quick Reference — Top 10 Priorities

| # | Priority | Issue | Doc |
|---|---|---|---|
| 1 | 🔴 CRITICAL | Hardcoded secrets in `config.toml` | [security.md](./security.md#1-hardcoded-secrets-in-configtoml) |
| 2 | 🔴 CRITICAL | Password reset token never emailed (silent failure) | [bugs.md](./bugs.md#1-password-reset-token-never-emailed) |
| 3 | 🟠 HIGH | PPID upload directory misconfiguration | [bugs.md](./bugs.md#2-ppid-upload-directory-returns-wrong-path) |
| 4 | 🟠 HIGH | Struktur `Description` field has no DB column | [bugs.md](./bugs.md#3-struktur-description-field-has-no-db-column) |
| 5 | 🟠 HIGH | Backup HTTP routes commented out (not accessible) | [bugs.md](./bugs.md#4-backup-http-routes-commented-out) |
| 6 | 🟠 HIGH | Active banner count not transactional (race condition) | [bugs.md](./bugs.md#5-active-banner-count-race-condition) |
| 7 | 🟡 MEDIUM | No bounding-box integration tests | [testing.md](./testing.md#1-missing-bounding-box-tests-for-fasilitas) |
| 8 | 🟡 MEDIUM | Module name `basic-service` vs folder `webdesa/api` | [architecture.md](./architecture.md#1-module-name-mismatch) |
| 9 | 🟡 MEDIUM | Go version drift between `go.mod` and `Dockerfile` | [operations.md](./operations.md#1-go-version-drift) |
| 10 | 🟡 MEDIUM | No generic settings CRUD API despite K-V schema | [architecture.md](./architecture.md#2-generic-settings-api-missing) |
| 11 | 🔴 CRITICAL | OpenAPI: 6 critical path/shape errors, 53 total discrepancies | [openapi-consistency.md](./openapi-consistency.md) |

## Documents

### 🔧 Issues & Fixes

- **[bugs.md](./bugs.md)** — Known bugs, incorrect behavior, broken features, schema mismatches
- **[security.md](./security.md)** — Security vulnerabilities, secrets management, auth/RBAC gaps
- **[code-quality.md](./code-quality.md)** — Refactoring, code smells, consistency improvements
- **[performance.md](./performance.md)** — Performance optimizations, caching, query efficiency

### 📐 Engineering Practices

- **[testing.md](./testing.md)** — Missing test coverage, test infrastructure improvements
- **[architecture.md](./architecture.md)** — Structural improvements, design patterns, separation of concerns
- **[operations.md](./operations.md)** — Deployment, observability, CI/CD, dev experience

### 📋 Audit

- **[openapi-consistency.md](./openapi-consistency.md)** — 53 discrepancies between OpenAPI spec and implementation (6 CRITICAL, 18 MAJOR, 14 MINOR)

## Recommended Action Plan

### Sprint 1 (Critical & High Priority)

1. Fix `config.toml` secret hygiene — rotate `JWT_SECRET`, move secrets to env vars
2. Implement password reset email sending (or document limitation)
3. Resolve PPID upload directory bug
4. Either add migration for `struktur.description` or remove from domain entity
5. Decide on backup HTTP routes — wire them up or remove the dead code

### Sprint 2 (Quality & Coverage)

6. Add bounding-box integration tests
7. Fix active banner count race condition (wrap in transaction)
8. Standardize the `fmt.Println` debug logging (umkm/service.go)
9. Resolve module name mismatch (`basic-service` → `webdesa/api`)
10. Add `golangci-lint` config and enforce in CI

### Backlog (Architectural & Performance)

11. Add generic settings API
12. Add Redis for rate limiting (current in-memory doesn't scale)
13. Add OpenAPI generation from code annotations
14. Add request-level tracing via OpenTelemetry
15. Implement token rotation for refresh tokens

## Summary Statistics

| Category | Issues | Doc |
|---|---|---|
| Bugs (correctness) | 8 | [bugs.md](./bugs.md) |
| Security | 7 | [security.md](./security.md) |
| Code quality | 12 | [code-quality.md](./code-quality.md) |
| Performance | 6 | [performance.md](./performance.md) |
| Testing gaps | 9 | [testing.md](./testing.md) |
| Architecture | 7 | [architecture.md](./architecture.md) |
| Operations | 8 | [operations.md](./operations.md) |
| OpenAPI consistency | 53 | [openapi-consistency.md](./openapi-consistency.md) |

**Total: 110 recommendations across 8 categories**

## Methodology

Recommendations are based on:
- Comprehensive codebase analysis (see `/feature-docs/` for full feature documentation)
- Review of all 36 SQL migrations
- Inspection of all 21 HTTP handlers
- Review of all 24 integration test files
- Code smell detection and pattern consistency checks
- OWASP API Security Top 10 review
- Production-readiness audit

Each recommendation includes:
- **Issue** — clear description
- **Impact** — what could go wrong if not fixed
- **Location** — file paths and line numbers
- **Recommendation** — proposed solution
- **Effort** — rough estimate (S/M/L)