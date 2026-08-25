# Bugs Audit — 2026 Q3

Original: `../recom.docs/bugs.md` (8 items, audit 2026-07-04).

| # | Priority | Issue | Status | Evidence |
|---|---|---|---|---|
| 1 | CRITICAL | Password reset token never emailed | `[RESOLVED]` | `usecase/auth/service.go:188-196`; `cmd/serve.go:158-162` |
| 2 | HIGH | PPID upload dir wrong | `[RESOLVED]` | `config/fileupload.go:44-49` |
| 3 | HIGH | Struktur.Description no DB column | `[RESOLVED]` | `db/migrations/00037_add_description_to_struktur.sql` |
| 4 | HIGH | Backup HTTP routes commented out | `[STILL OPEN]` | `cmd/serve.go:148,282-290,325` |
| 5 | HIGH | Banner count race condition | `[RESOLVED]` | `db/migrations/00038_enforce_active_banner_limit.sql` |
| 6 | MEDIUM | PPID email/DB inconsistency | `[RESOLVED]` | `usecase/ppid/service.go:695-711` |
| 7 | MEDIUM | Module name mismatch | `[RESOLVED]` | `go.mod:1`, `README.md:45,348` |
| 8 | LOW | PPID category nullable | `[STILL OPEN]` | `db/migrations/00032_convert_ppid_category_to_fk.sql:26` |

---

## 1. Password Reset Token Never Emailed — `[RESOLVED]`

**Original claim:** `usecase/auth/service.go:171` had TODO; reset token persisted but not emailed.

**Fix applied:** `usecase/auth/service.go:188-196` now calls `s.email.SendPasswordResetEmail(...)`. `cmd/serve.go:158-162` constructs `authEmailSvc`; wired into `auth.NewService(...)` at `serve.go:208-216`. `interface/email/auth_email_service.go` exists.

**Note:** Per `feature-docs/auth/concept.md:26-28`, the reset email body still hardcodes "1 hour" but TTL is 24h — new minor doc bug.

---

## 2. PPID Upload Directory Wrong — `[RESOLVED]`

**Original claim:** `config/fileupload.go:35-40` returned `UploadPublicDirectory` instead of `UploadPPIDDirectory`.

**Fix applied:** `config/fileupload.go:44-49` now correctly returns `c.UploadPPIDDirectory` with fallback. Migration 00045 dropped all legacy URL columns.

---

## 3. Struktur.Description No DB Column — `[RESOLVED]`

**Original claim:** `db:"-"` tag dropped bytes silently.

**Fix applied:** `db/migrations/00037_add_description_to_struktur.sql` adds column. `domain/struktur/struktur.go:23` uses `db:"description"` (no `db:"-"`).

**Note:** A separate `struktur/concept.md` bug still documented — `interface/postgres/struktur.go:139` has `phone = $4 = $5` typo + 7 placeholders / 8 args, breaking Update. NOT in original `bugs.md`; new finding.

---

## 4. Backup HTTP Routes Commented Out — `[STILL OPEN]`

**Original claim:** `cmd/serve.go:196-204, 228` had commented handler construction. `router.go:334-346` had commented routes.

**Current state:** Still commented. `cmd/serve.go:148` (`// backupRepo := ...`), `cmd/serve.go:282-290` (full backup service construction), `cmd/serve.go:325` (`// backupHandler := ...`). `interface/http/handler/backup.go` is dead code (still exists). `README.md:260-261` still references `POST /api/backup` and `POST /api/restore` (will 404).

**Action:** Either wire with RBAC + integration tests, or delete dead code + update README.

---

## 5. Banner Count Race Condition — `[RESOLVED]`

**Original claim:** Concurrent updates could exceed 20-active-banner limit.

**Fix applied:** `db/migrations/00038_enforce_active_banner_limit.sql` adds `enforce_active_banner_limit()` plpgsql trigger. `usecase/banner/service.go:289-294` docstring confirms DB-level enforcement. Per `feature-docs/banner/database.md:51-57`.

---

## 6. PPID Email/DB Inconsistency — `[RESOLVED]`

**Original claim:** Email sent before DB update; SMTP failure leaves request in pending state.

**Fix applied (commit 2e5b0a0):** `usecase/ppid/service.go` now persists approval BEFORE sending email. DB stays consistent on SMTP failure; error message changed to `"request approved but email send failed (resend required)"` so operators know the state. Future work: add `/ppid/requests/{id}/resend-email` endpoint for operator recovery.

---

## 7. Module Name Mismatch — `[RESOLVED]`

**Original claim:** `go.mod:1` = `module basic-service`.

**Fix applied (commit 27d0209):** `README.md:45,348` updated from `basic-service` → `webdesa/api`. Combined with prior `go.mod:1` fix, fully resolved.

---

## 8. PPID Category Nullable — `[STILL OPEN]`

**Original claim:** PPID `category` preserved as NULL while other features use NOT NULL.

**Current state:** `db/migrations/00032_convert_ppid_category_to_fk.sql:26` deliberately preserves NULL. No subsequent migration (00033-00046) makes it NOT NULL. Per `feature-docs/ppid/database.md:73-74` still pending decision.

**Action:** Business decision needed — keep nullable or backfill + NOT NULL.
