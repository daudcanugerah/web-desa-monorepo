# Security Audit — 2026 Q3

Original: `../recom.docs/security.md` (7 items, audit 2026-07-04).

| # | Priority | Issue | Status | Evidence |
|---|---|---|---|---|
| 1 | CRITICAL | Hardcoded secrets | `[RESOLVED]` | `config/app.go` validation rejects placeholders |
| 2 | CRITICAL | `/uploads/*` no auth | `[STILL OPEN]` | `interface/http/router.go:99` |
| 3 | HIGH | PPID dual token modes | `[PARTIAL]` | `usecase/ppid/service.go:780-799`; legacy fallback kept |
| 4 | HIGH | Reset rate-limit counts used tokens | `[RESOLVED]` | `interface/postgres/auth.go:169-176` |
| 5 | HIGH | No CSRF protection | `[STILL OPEN]` | No csrf middleware under `interface/http/middleware/` |
| 6 | MEDIUM | No request body size limit | `[RESOLVED]` | `interface/http/middleware/maxbytes.go` (60 MiB) |
| 7 | MEDIUM | JWT refresh no rotation | `[STILL OPEN]` | `pkg/jwt/jwt.go:39-41` (static 7d) |

---

## 1. Hardcoded Secrets — `[RESOLVED]`

**Original claim:** JWT secret, Metabase secret_key, SMTP password checked into `config.toml`.

**Fix applied (commit 3e8b0c6):** `config/app.go` now calls `validateSecrets()` after `viper.Unmarshal`. Startup refuses to boot when JWT/Metabase/SMTP secrets match a placeholder or historically leaked value (`gajah123`, `dev-secret-key-change-this`, `CHANGE_THIS_*`, `REPLACE_VIA_*_ENV_VAR`). Bypass flag `ALLOW_INSECURE_SECRETS=1` for integration tests. `config/prod.toml` and `config/dev.toml` updated with documented env-var placeholders.

**Remaining note:** The local `config.toml` (gitignored) still contains raw values. Operators must move them to env vars (`JWT_SECRET`, `METABASE_SECRET_KEY`, `SMTP_PASSWORD`) and clear the file.

---

## 2. `/uploads/*` No Auth — `[STILL OPEN]`

**Original claim:** `router.go:75-76` static file server exposes all uploaded files.

**Current state:** `interface/http/router.go:98-99`:
```go
fileServer := http.FileServer(http.Dir(cfg.UploadPublicDirectory))
r.Handle("/uploads/*", http.StripPrefix("/uploads/", fileServer))
```
No auth middleware on this route. Exposes user avatars, PPID documents, banner images.

**Action:** Either remove the route or apply signed URLs / per-resource auth check. Combined with `gallery/concept.md` Option A+C from original doc.

---

## 3. PPID Dual Token Modes — `[PARTIAL]`

**Original claim:** `DownloadDocument` accepts both PPID-request JWT and user JWT.

**Current state:** `usecase/ppid/service.go:780-799` `DownloadDocument` still has legacy user-JWT fallback (lines 783-785). New unified signed-URL flow exists via `galleryusecase.SignedURLService` (service.go:662-677) but legacy path kept.

**Action:** Deprecate user-JWT path or document the dual mode. Per `feature-docs/ppid/`, signed URLs are the intended path.

---

## 4. Reset Rate-Limit Counts Used Tokens — `[RESOLVED]`

**Original claim:** SQL counts all reset tokens including `used=true`.

**Fix applied (commit 1a77648):** `interface/postgres/auth.go:169-176` now includes `AND used = false` in the WHERE clause. Used tokens no longer count toward the per-user rate window.

---

## 5. No CSRF Protection — `[STILL OPEN]`

**Original claim:** No CSRF middleware.

**Current state:** `interface/http/middleware/` contains only `auth.go`, `cors.go`, `logger.go`, `ratelimit.go`, `rbac.go`, `recovery.go`, `request_id.go`, `middleware_deps.go`. No csrf middleware. Only `cors.go:19` lists `X-CSRF-Token` as allowed header.

**Action:** If any cookie-based auth is added, include `github.com/go-chi/csrf` middleware. Current bearer-only auth is OK but should be documented in security policy.

---

## 6. No Body Size Limit — `[RESOLVED]`

**Original claim:** Only `chi.Timeout(60s)`, no `MaxBytesReader`.

**Fix applied (commit 3e8b0c6):** `interface/http/middleware/maxbytes.go` adds global `MaxBytesMiddleware` (60 MiB default, configurable via `RouterConfig.MaxRequestBodyBytes`). Wired into `interface/http/router.go:93`. Per-route upload handlers still apply tighter limits via `MaxBytesReader`.

---

## 7. JWT Refresh Token Rotation — `[STILL OPEN]`

**Original claim:** 7-day refresh, no rotation.

**Current state:** `pkg/jwt/jwt.go:39-41` — `GenerateRefreshToken` returns static 7-day token. `usecase/auth/service.go:111-131` `RefreshToken` just mints new access token without DB-tracked chain.

**Action:** Implement refresh token table + rotation chain (per original doc recommendation).
