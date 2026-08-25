# Testing Audit — 2026 Q3

Original: `../recom.docs/testing.md` (9 items, audit 2026-07-04).

| # | Priority | Issue | Status | Evidence |
|---|---|---|---|---|
| 1 | HIGH | Bbox tests for fasilitas | `[STILL OPEN]` | `grep -rn "BoundingBox\|BBox" integration/` = 0 |
| 2 | MEDIUM | Backup integration tests | `[STILL OPEN]` | No `integration/backup*.go` files |
| 3 | MEDIUM | Email notification tests for PPID | `[PARTIAL]` | `ppid_approval_test.go` exists but no mail catcher |
| 4 | MEDIUM | Chaos/failure tests | `[STILL OPEN]` | No `chaos_test.go` |
| 5 | MEDIUM | Property-based tests | `[STILL OPEN]` | No `quick.Check`/`gopter` |
| 6 | MEDIUM | Performance/benchmark tests | `[STILL OPEN]` | No `Benchmark*` funcs |
| 7 | MEDIUM | Coverage threshold enforcement | `[RESOLVED]` | `make test-coverage-check` (80%) |
| 8 | LOW | Mutation testing | `[STILL OPEN]` | No go-mutesting / gremlins |
| 9 | LOW | E2E tests | `[STILL OPEN]` | No `e2e/` directory |

---

## 1. Bounding-Box Tests for Fasilitas — `[STILL OPEN]`

**Original claim:** No tests for `List(ListFasilitasInput{BBox})`.

**Current state:** `grep -rn "BoundingBox\|BBox" /Users/ivosights/Office/3web/webdesa/api/integration/` returns 0 hits. `integration/fasilitas_test.go` has `TestFasilitasList_FilterByQuery` and `_FilterByCategory` but no BBox test. BBox query is admin-only per `feature-docs/fasilitas/endpoint.md`.

**Action:** Add the 4-quadrant test scaffold from original doc.

---

## 2. No Backup Integration Tests — `[STILL OPEN]`

**Original claim:** Backup feature completely untested.

**Current state:** 28 integration test files exist; no `backup_test.go`. Backup routes not wired in router (see `bugs.md#4`) — but CLI could be tested.

**Action:** Add `TestBackup_Create` / `_Restore` / `_Download` per original doc. Independent of HTTP wiring decision.

---

## 3. Email Notification Tests for PPID — `[PARTIAL]`

**Original claim:** Approval workflow tests don't verify email.

**Current state:** `integration/ppid_approval_test.go` exists. Tests use `noopEmailService{}` (setup_test.go:488) — no actual SMTP mock. Email is never verified to have been sent.

**Action:** Use MailHog or `github.com/stretchr/testify/mock` for SMTP. Verify message body + recipient.

---

## 4. No Chaos/Failure Tests — `[STILL OPEN]`

**Original claim:** Tests assume happy path.

**Current state:** `grep -rn "TestChaos\|DBDown" integration/` returns 0 hits. No DB-down, SMTP-timeout, or concurrent-stress tests.

**Action:** Add the two scaffold tests from original doc (`TestBerita_ConcurrentUpdate`, `TestAuth_DBDown`).

---

## 5. No Property-Based Tests — `[STILL OPEN]`

**Original claim:** Validation tested with hand-picked cases.

**Current state:** No `testing/quick` or `gopter` usage (`grep` returns 0 hits in code).

**Action:** Add property tests for complex validators per original doc.

---

## 6. No Performance/Benchmark Tests — `[STILL OPEN]`

**Original claim:** Zero benchmarks.

**Current state:** `grep -rn "Benchmark" *_test.go` returns 0 hits.

**Action:** Add `BenchmarkBeritaList`, `BenchmarkAuthLogin` per original doc.

---

## 7. No Coverage Threshold — `[RESOLVED]`

**Original claim:** `coverage.out` generated but no minimum enforced.

**Fix applied (commit 34749ab):** `make test-coverage-check` target added. Default threshold 80%, overridable via `COVERAGE_THRESHOLD=N make test-coverage-check`. Fails CI when total coverage drops below threshold.

---

## 8. No Mutation Testing — `[STILL OPEN]`

**Original claim:** Tests can pass with always-true asserts.

**Current state:** No `go-mutesting` / `gremlins` config or CI step.

**Action:** Setup + CI integration (LOW priority — defer until coverage threshold lands).

---

## 9. No E2E Tests — `[STILL OPEN]`

**Original claim:** No end-to-end user journey tests.

**Current state:** No `e2e/` directory. Integration tests cover API endpoints but no full frontend flow simulation.

**Action:** LOW priority — defer until CI/CD exists (see `operations.md#2`).
