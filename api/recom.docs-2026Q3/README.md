# Recommendations Audit — 2026 Q3

Fresh audit of every recommendation in `recom.docs/` against current code state. Each item is reclassified with current evidence.

**Audit date:** 2026-08-25
**Codebase root:** `/Users/ivosights/Office/3web/webdesa/api`

## Status Legend

- `[RESOLVED]` — fix applied, verified in code
- `[STILL OPEN]` — recommendation still valid, not addressed
- `[PARTIAL]` — some work done, incomplete or only partial coverage
- `[SUPERSEDED]` — feature redesigned / claim no longer accurate
- `[CANT VERIFY]` — requires live system / external state

## Priority Legend (preserved from original docs)

- `CRITICAL` — security, data loss, production-blocking
- `HIGH` — correctness bugs, broken features
- `MEDIUM` — quality, maintainability, performance
- `LOW` — polish, future-proofing

## Summary

| Category | Total | Resolved | Still Open | Partial | Superseded | Cant Verify |
|---|---|---|---|---|---|---|
| Security | 7 | 3 | 2 | 1 | 0 | 1 |
| Bugs | 8 | 7 | 1 | 0 | 0 | 0 |
| Code Quality | 12 | 4 | 4 | 3 | 1 | 0 |
| Performance | 8 | 3 | 3 | 2 | 0 | 0 |
| Testing | 9 | 1 | 7 | 0 | 0 | 1 |
| Architecture | 7 | 3 | 1 | 2 | 1 | 0 |
| Operations | 8 | 2 | 3 | 2 | 0 | 1 |
| OpenAPI | 53 | ~18 | ~5 | ~0 | ~30 | 0 |
| **Total** | **112** | **~41** | **24** | **10** | **~32** | **3** |

## Resolved in 2026 Q3 audit pass

**Round 1 (commits 1a77648 .. df504d2):**
- `security.md#4` — Reset rate-limit SQL fix
- `bugs.md#7` — README module name fix
- `code-quality.md#8` — `.golangci.yml` config
- `code-quality.md#9` — Dead RBAC files removed
- `performance.md#6` — `/banners/active` cap already documented
- `testing.md#7` — `make test-coverage-check` target
- `openapi-consistency.md#6` — 6 obsolete PPID tests deleted

**Round 2 (commits 3e8b0c6 .. 50c6253):**
- `security.md#1` — Startup secret validation
- `security.md#6` — Global body size cap (60 MiB)
- `bugs.md#6` — PPID email/DB order fix (DB first)
- `code-quality.md#6` — `RequireUserIDFromCtx` + `ParseUUIDParam` helpers
- `performance.md#4` — Composite + pg_trgm indexes (migration 00047)
- `operations.md#2` — GitHub Actions CI workflow

Note: `openapi-consistency.md` is heavily stale — every `openapi.yaml:NNN` line ref is now invalid because the file was replaced with code-generated `docs/swagger.yaml`. Most items there are `[RESOLVED]` by the swag migration but cannot be verified line-by-line.

## Top Priorities (Still Open, CRITICAL/HIGH)

1. `security.md#2` `[STILL OPEN]` — Static `/uploads/*` no auth. `router.go:99`.
2. `bugs.md#4` `[STILL OPEN]` — Backup HTTP routes commented out. `cmd/serve.go:148,282-290,325`.
3. `performance.md#1` `[STILL OPEN]` — In-memory rate limiter. `interface/http/middleware/ratelimit.go`.
4. `testing.md#1` `[STILL OPEN]` — No bbox tests for fasilitas.
5. `testing.md#6` `[STILL OPEN]` — No benchmark tests.
6. `security.md#3` `[PARTIAL]` — PPID dual token mode (legacy fallback).
7. `security.md#7` `[STILL OPEN]` — JWT refresh rotation missing.

## Documents

- `security.md` — 7 items
- `bugs.md` — 8 items
- `code-quality.md` — 12 items
- `performance.md` — 8 items
- `testing.md` — 9 items
- `architecture.md` — 7 items
- `operations.md` — 8 items
- `openapi-consistency.md` — 53 items (heavily stale)
