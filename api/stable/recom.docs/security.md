# Security Recommendations

Security audit and improvements for the **webdesa/api** codebase.

---

## 1. Hardcoded Secrets in `config.toml` 🔴 CRITICAL

**Location:** `config.toml`

**Issue:**

Production secrets are committed (or currently present in) `config.toml`:
```toml
[jwt]
secret = "desa-api-secret-key-change-in-production-2024"

[metabase]
secret_key = "268948167f2d69b5840b651d1af2e216f750459fc92e86159d387186d291b5c5"

[smtp]
password = "gajah123"
```

**Impact:**
- Token signing secret exposed → can forge any JWT
- Metabase JWT key exposed → can access all Metabase embeds
- SMTP credentials exposed → can send email as the system

**Recommendation:**

1. **Verify `.gitignore`** excludes `config.toml` (currently does)
2. **Rotate ALL secrets immediately:**
   - Generate new JWT secret: `openssl rand -hex 32`
   - Generate new Metabase secret
   - Change SMTP password
3. **Use environment variables** for production (Viper supports `AutomaticEnv()`)
4. **Update `config/prod.toml`** to use env var references:
   ```toml
   [jwt]
   secret = "${JWT_SECRET}"
   
   [metabase]
   secret_key = "${METABASE_SECRET}"
   ```
5. **Use a secrets manager** for production (Vault, AWS Secrets Manager, etc.)
6. **Add a startup check** that fails if default secrets are detected:
   ```go
   if cfg.JWT.Secret == "desa-api-secret-key-change-in-production-2024" {
       log.Fatal("JWT secret must be changed in production")
   }
   ```

**Effort:** S (1 day)

---

## 2. Static `/uploads/*` Has No Authentication 🔴 CRITICAL

**Location:** `interface/http/router.go:75-76`

**Issue:**
```go
fileServer := http.FileServer(http.Dir("./uploads"))
r.Handle("/uploads/*", http.StripPrefix("/uploads/", fileServer))
```

The raw file server exposes ALL uploaded files including:
- User profile images
- PPID documents (potentially sensitive)
- Quill `tmp-` files (intermediate uploads)

While `/api/v1/files/:filename` has directory-traversal protection, the `/uploads/*` route does NOT.

**Impact:**
- Direct URL guessing can access any uploaded file
- PPID documents especially sensitive
- Profile images leak user identity

**Recommendation:**

**Option A — Remove the static route (best):**
```go
// Delete these lines from router.go
```

**Option B — Add path-based auth:**
```go
r.Handle("/uploads/*", AuthMiddleware, /* role check */, 
    http.StripPrefix("/uploads/", fileServer))
```

**Option C — Use opaque IDs:**
```go
// Generate unguessable IDs like uuid without extension
// Serve via /api/v1/files/<id>
```

**Recommendation:** Combine options A + C. Store files with UUID-only names (no extension), serve only through `/api/v1/files/:filename` with traversal protection.

**Effort:** S (1 day)

---

## 3. PPID Download Token Has Dual Modes — Increases Attack Surface 🟠 HIGH

**Location:** `usecase/ppid/service.go:617-675`

**Issue:**

`DownloadDocument` accepts BOTH:
1. **PPID-request-issued JWT** (10-min, single-document)
2. **Regular user JWT** (login token, can access any PPID)

This dual mode means:
- A regular user JWT has broader access than intended
- Hard to audit who accessed what
- Revoke doesn't revoke all access

**Impact:**
- Admins can download any PPID even without a request
- User JWTs become a target (long expiry = 24h)

**Recommendation:**

Consider removing user JWT access:

```go
// Refuse user JWT mode, require document-specific token only
func (s *Service) DownloadDocument(...) {
    token, err := s.ValidateAccessToken(input.Token)
    if err != nil {
        return nil, err  // No fallback to user JWT
    }
    // ...
}
```

Or audit log all downloads with attribution:
```go
type DownloadAuditLog struct {
    ID           string
    PPIDID       string
    TokenType    string  // "request" | "user"
    UserID       *string
    IPAddress    string
    UserAgent    string
    DownloadedAt time.Time
}
```

**Effort:** M (2-3 days)

---

## 4. Password Reset Rate-Limit Counts Wrong 🟠 HIGH

**Location:** `usecase/auth/service.go:130-136`

**Issue:**

```go
func (s *Service) countResetRequests(...) (int, error) {
    count, _ := s.repo.CountResetRequestsByEmail(ctx, email)
    if count >= 5 {
        return errorTooManyRequests
    }
    // ...
}
```

The SQL counts ALL reset tokens including those marked `used=true`:
```sql
SELECT COUNT(*) FROM password_reset_tokens
WHERE user_id = $1 AND created_at >= $2
```

This means once a user resets their password 5 times in an hour, they're locked out for the rest of the hour even after successful resets.

**Impact:**
- Legitimate users may hit the limit
- Used tokens shouldn't count toward rate limit

**Recommendation:**

Update the WHERE clause:
```sql
SELECT COUNT(*) FROM password_reset_tokens
WHERE user_id = $1
  AND created_at >= $2
  AND used = false  -- ← exclude used tokens
```

Or use a different time window after a successful reset:
```go
// Reset the count when user successfully resets
```

**Effort:** XS (30 minutes)

---

## 5. No CSRF Protection on Cookie-Based Requests 🟠 HIGH

**Location:** `interface/http/router.go`, all handlers

**Issue:**

If the frontend uses cookies for auth (which JWT in Authorization header typically doesn't), there's no CSRF protection. Even with bearer tokens, some endpoints may set cookies (e.g., last admin login).

**Recommendation:**

- **If using Authorization header only**: This is mitigated. JWT in localStorage is vulnerable to XSS; in httpOnly cookies requires CSRF.
- **Add CSRF middleware** if any cookie is set:
  ```go
  import "github.com/go-chi/csrf"
  
  r.Use(csrf.Protect([]byte("csrf-secret-key")))
  ```
- **Use SameSite=Strict** on any session cookies
- **Document auth strategy** in security policy

**Effort:** S (1 day if needed)

---

## 6. No Request Size Limit Beyond chi Timeout 🟡 MEDIUM

**Location:** `interface/http/router.go:64-69`

**Issue:**

```go
r.Use(chi.Timeout(60s))
```

Only timeout is enforced, not max body size. A malicious client could upload gigabytes and exhaust disk.

**Recommendation:**

Add `http.MaxBytesReader` middleware:
```go
r.Use(func(next http.Handler) http.Handler {
    return http.MaxBytesHandler(next, 50<<20)  // 50 MB
})
```

Or per-route limits via httpin/validator:
```go
r.With(MaxBodySize(50 << 20)).Post("/berita/upload-media", ...)
```

**Effort:** XS (1 hour)

---

## 7. JWT Refreshs Don't Rotate 🟡 MEDIUM

**Location:** `pkg/jwt/jwt.go:34-41`

**Issue:**

Refresh tokens are valid for **7 days** with no rotation. If a refresh token is stolen, attacker has 7 days of access even after password change.

**Recommendation:**

Implement refresh token rotation:
```go
type RefreshToken struct {
    ID        string  // UUID
    UserID    string
    Used      bool
    CreatedAt time.Time
    ExpiresAt time.Time
}

// On refresh:
func (s *Service) RefreshToken(ctx, oldRefresh string) (newAccess, newRefresh string, err error) {
    // Validate old
    // Mark old as used
    // Issue new pair
    // Return new pair
}
```

Old tokens become invalid after use. If a used token is presented again, invalidate the entire chain (potential theft detected).

**Effort:** M (2-3 days)

---

## Summary Table

| # | Issue | Priority | Effort | OWASP |
|---|---|---|---|---|
| 1 | Hardcoded secrets | 🔴 CRITICAL | S | A02:2021 |
| 2 | Static uploads no auth | 🔴 CRITICAL | S | A01:2021 |
| 3 | PPID dual token modes | 🟠 HIGH | M | A01:2021 |
| 4 | Reset rate limit wrong | 🟠 HIGH | XS | A04:2021 |
| 5 | No CSRF protection | 🟠 HIGH | S | A01:2021 |
| 6 | No body size limit | 🟡 MEDIUM | XS | A05:2021 |
| 7 | JWT no rotation | 🟡 MEDIUM | M | A07:2021 |

## OWASP API Top 10 Checklist

- [ ] **A01: Broken Access Control** — Static uploads exposed (item 2)
- [ ] **A02: Cryptographic Failures** — Hardcoded secrets (item 1)
- [ ] **A03: Injection** — Uses sqlx with parameterized queries ✓
- [ ] **A04: Insecure Design** — Rate limit issue (item 4)
- [ ] **A05: Security Misconfiguration** — Body size (item 6)
- [ ] **A07: Auth Failures** — JWT rotation (item 7)
- [ ] **A08: Data Integrity** — Email/DB consistency
- [ ] **A09: Logging Failures** — Add structured audit logs

## Additional Recommendations

- **Add security headers middleware** (`X-Content-Type-Options`, `X-Frame-Options`, `Strict-Transport-Security`)
- **Add request ID correlation** to all logs (already partially done via middleware)
- **Add rate limiting per-user** not just per-IP for authenticated routes
- **Add login attempt lockout** (separate from rate limit)
- **Regular dependency updates** via `go list -m -u all`
- **Add SAST** (gosec, staticcheck) to CI
