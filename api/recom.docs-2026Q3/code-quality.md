# Code Quality Audit — 2026 Q3

Original: `../recom.docs/code-quality.md` (12 items, audit 2026-07-04).

| # | Priority | Issue | Status | Evidence |
|---|---|---|---|---|
| 1 | MEDIUM | `fmt.Println` debug in UMKM | `[RESOLVED]` | `grep fmt.Println usecase/` returns 0 hits |
| 2 | MEDIUM | Inconsistent error wrapping | `[STILL OPEN]` | `errtrace.Wrap` + bare `fmt.Errorf` coexist |
| 3 | MEDIUM | Two repo directories | `[SUPERSEDED]` | `interface/postgres/` is single dir |
| 4 | MEDIUM | Long handler files (>500 lines) | `[STILL OPEN]` | `handler/ppid/ppid.go` ≈900 lines |
| 5 | MEDIUM | Magic permission strings | `[STILL OPEN]` | No `pkg/permissions/` |
| 6 | MEDIUM | Duplicated handler boilerplate | `[RESOLVED]` | `pkg/handlerutil.RequireUserIDFromCtx`, `ParseUUIDParam` |
| 7 | LOW | Inconsistent handler naming | `[PARTIAL]` | Most use Verb+Resource; auth.go bare verbs |
| 8 | MEDIUM | Missing `.golangci-lint` config | `[RESOLVED]` | `.golangci.yml` at root |
| 9 | LOW | Two Casbin model files | `[RESOLVED]` | `rbac/model.conf`, `rbac/policy.csv` deleted |
| 10 | LOW | Commented-out code in router | `[STILL OPEN]` | `cmd/serve.go:148,282-290,325` |
| 11 | LOW | Hardcoded path strings | `[PARTIAL]` | Mixed; `/uploads/*` still hardcoded |
| 12 | LOW | Inconsistent JSON field naming | `[PARTIAL]` | Mostly consistent, no linter |

---

## 1. `fmt.Println` Debug — `[RESOLVED]`

**Original claim:** `usecase/umkm/service.go:96-99, 108` had `fmt.Println` debug statements.

**Current state:** `grep -r "fmt.Println" usecase/` returns 0 hits. Only `cmd/root.go:108` has `fmt.Fprintf(os.Stderr,...)` for fatal error path — intentional.

---

## 2. Inconsistent Error Wrapping — `[STILL OPEN]`

**Original claim:** Mixed `errtrace.Wrap`, `fmt.Errorf("...: %w", err)`, bare returns.

**Current state:** Both styles still coexist:
- `usecase/backup/service.go:69,93,99,114...` — `errtrace.Wrap(fmt.Errorf(...))`
- `usecase/banner/service.go:102,135,146,170...` — bare `fmt.Errorf("...: %w", err)`

No `golangci-lint` rule to enforce (see item #8).

**Action:** Pick one convention. Add lint rule once #8 lands.

---

## 3. Two Repo Directories — `[SUPERSEDED]`

**Original claim:** `interface/postgres/` and `interface/postgres/` both contain `package postgres`.

**Current state:** Single `interface/postgres/` directory (22 files). Claim no longer accurate — auth.go and user.go were unified in this single dir.

---

## 4. Long Handler Files — `[STILL OPEN]`

**Original claim:** `handler/ppid.go` = 838 lines, `banner.go` = 547.

**Current state:** `interface/http/handler/ppid/ppid.go` ≈ 35.8 KB (~900 lines), `banner.go` = 27.7 KB. Per-feature split into subdirs done for routes (`route.go`) but main handlers still large.

**Action:** Split into `document.go`, `request.go`, `approval.go`, `category.go` per original doc recommendation.

---

## 5. Magic Permission Strings — `[STILL OPEN]`

**Original claim:** Bare strings `RBACMiddleware(enforcer, "berita", "write")`.

**Current state:** `interface/http/handler/ppid/route.go:33,40` still uses bare `mw.RBAC("ppid", "read")`. No `pkg/permissions/` constants package.

**Action:** Add `pkg/permissions/permissions.go` with constants.

---

## 6. Handler Boilerplate — `[RESOLVED]`

**Original claim:** `UserIDFromCtx`, `ParseUUID` helpers needed.

**Fix applied (commit 743a031):** `pkg/handlerutil/handlerutil.go` now exposes `RequireUserIDFromCtx` and `ParseUUIDParam`. Existing call sites can migrate incrementally (61 `chi.URLParam(r,"id")` sites + 4 `GetUserIDFromContext` sites still using the long form).

---

## 7. Handler Naming — `[PARTIAL]`

**Original claim:** Mixed `GetUser`/`CreateUser` vs bare `Login`/`RefreshToken`.

**Current state:** Most follow Verb+Resource (`CreateBanner`, `ListBanners`, `UpdateStruktur`). Exceptions: `auth.go:89,136` use `Login`/`RefreshToken`.

**Action:** Rename `Login`→`CreateSession` or accept as convention.

---

## 8. Missing `golangci-lint` Config — `[RESOLVED]`

**Original claim:** No `.golangci.yml` at repo root.

**Fix applied (commit a79dc3c):** `.golangci.yml` added with errcheck, gosimple, govet, ineffassign, staticcheck, unused, bodyclose, errorlint, nilerr, gocritic. Test files excluded from errcheck/gocritic. `Makefile lint` target now enforces rules.

---

## 9. Two Casbin Model Files — `[RESOLVED]`

**Original claim:** `rbac/model.conf` and `rbac/rbac_model.conf` both exist.

**Fix applied (commit af5033c):** `rbac/model.conf` and `rbac/policy.csv` deleted. Only `rbac/rbac_model.conf` + `rbac/rbac_policy.csv` remain (the pair loaded by `cmd/serve.go:105` and `cmd/seed.go:109`).

---

## 10. Commented-Out Code — `[STILL OPEN]`

**Original claim:** Backup routes commented in router + serve.go.

**Current state:** Same as `bugs.md#4` — `cmd/serve.go:148,282-290,325` all commented.

**Action:** See `bugs.md#4`.

---

## 11. Hardcoded Path Strings — `[PARTIAL]`

**Original claim:** `http.ServeFile(w, r, filepath.Join("./uploads", filename))`.

**Current state:** Mixed:
- `cmd/serve.go:324` uses config: `filehandlerpkg.NewFileHandler(systemConfig.FileUpload.GetPublicUploadDirectory(), ...)`
- `cmd/serve.go:314` uses config: `systemConfig.FileUpload.GetPPIDUploadDirectory()`
- `interface/http/router.go:98` still hardcodes: `http.Dir(cfg.UploadPublicDirectory)` — but `cfg` is the RouterConfig, so this is config-driven. Original doc's specific concern addressed.

**Action:** Audit remaining hardcoded paths.

---

## 12. JSON Field Naming — `[PARTIAL]`

**Original claim:** Inconsistent snake_case vs camelCase.

**Current state:** Mostly consistent snake_case. `auth/auth.go:39-42` uses `access_token`, `refresh_token`, `expires_at`. No linter to enforce.

**Action:** Add `revive` or similar lint rule once #8 lands.
