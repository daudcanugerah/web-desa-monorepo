# Code Quality Recommendations

Refactoring, code smells, consistency improvements.

---

## 1. `fmt.Println` Debug Logging in UMKM Service 🟡 MEDIUM

**Location:** `usecase/umkm/service.go:96-99, 108`

**Issue:**
```go
fmt.Println("DEBUG: received", len(input.ImageFiles), "images")
fmt.Println("DEBUG: result of rename", err)
fmt.Println("DEBUG: result of save images", err)
```

**Impact:**
- Pollutes stdout in production
- Not structured, no log levels
- Cannot be filtered or correlated

**Recommendation:**

Replace with structured logger:
```go
s.logger.Info("creating UMKM with images",
    "count", len(input.ImageFiles),
    "name", input.Name,
)
```

Use the existing `pkg/otel` logger or introduce `slog` (already in stdlib).

**Effort:** XS (1 hour)

---

## 2. Inconsistent Error Wrapping 🟡 MEDIUM

**Location:** Throughout codebase

**Issue:**

Some places use `braces.dev/errtrace`, others use `fmt.Errorf("...: %w", err)`, others use bare returns. Inconsistent.

**Impact:**
- Stack traces inconsistent
- Hard to debug

**Recommendation:**

Pick ONE convention and enforce it:

```go
// Always use this pattern
return fmt.Errorf("operation %s failed: %w", op, err)

// Or always use errtrace
import "braces.dev/errtrace"
return errtrace.Wrap(fmt.Errorf("...: %w", err))
```

Add a `golangci-lint` rule to enforce.

**Effort:** S (1-2 days)

---

## 3. Two Repository Directories (`interface/postgres/` and `interface/postgres/`) 🟡 MEDIUM

**Location:** `interface/postgres/`, `interface/postgres/`

**Issue:**

Both directories contain `package postgres` files:
- `interface/postgres/auth.go` — auth-specific
- `interface/postgres/user.go` — user
- `interface/postgres/<feature>.go` — all others (no _postgres suffix)

Same Go package name, different directories. Confusing for navigation.

**Recommendation:**

Consolidate into ONE directory:
```
interface/postgres/
  auth_repository.go
  user_repository.go
  banner_repository.go
  // ... all other repos
```

Or use the names as the package discriminator:
```
interface/postgres/auth/    → package auth
interface/postgres/content/ → package content
```

**Effort:** S (refactor, ~1 day)

---

## 4. Long Handler Files (>500 lines) 🟡 MEDIUM

**Location:** 
- `interface/http/handler/ppid.go` (838 lines)
- `interface/http/handler/banner.go` (547 lines)
- `interface/http/handler/berita.go` (451 lines)
- `interface/http/handler/user.go` (~400 lines)

**Issue:**

These handlers mix:
- Request parsing
- Validation
- Business logic invocation
- Response shaping
- Error translation

**Impact:**
- Hard to test individually
- Hard to read
- Multiple developers editing same file

**Recommendation:**

Split each handler into per-resource files:
```
interface/http/handler/ppid/
  document.go        // CRUD endpoints
  request.go         // Request workflow
  approval.go        // Approve/revoke
  category.go        // Category CRUD
```

Plus extract response shapes to separate files:
```
interface/http/handler/ppid/responses.go
interface/http/handler/ppid/requests.go
```

**Effort:** M (3-5 days total)

---

## 5. Magic Strings for Permission/Resource Names 🟡 MEDIUM

**Location:** Throughout handlers, services

**Issue:**
```go
RBACMiddleware(enforcer, "berita", "write")
RBACMiddleware(enforcer, "users", "read")
```

Magic strings everywhere. Typo-prone, hard to grep.

**Recommendation:**

Define constants:
```go
// pkg/permissions/permissions.go
package permissions

const (
    ResourceBerita      = "berita"
    ResourceUsers       = "users"
    ResourceRoles       = "roles"
    ResourceBanners     = "banners"
    // ...
    
    ActionRead  = "read"
    ActionWrite = "write"
)

func BeritaRead() string  { return ResourceBerita + ":" + ActionRead }
func BeritaWrite() string { return ResourceBerita + ":" + ActionWrite }
// ...
```

Use at registration:
```go
r.With(RBACMiddleware(enforcer, permissions.BeritaRead()))
```

**Effort:** S (1 day)

---

## 6. Duplicated Handler Boilerplate 🟡 MEDIUM

**Location:** All handlers

**Issue:**

Every handler repeats:
```go
userID, ok := middleware.GetUserIDFromContext(r.Context())
if !ok {
    response.Error(w, http.StatusUnauthorized, "unauthorized", nil)
    return
}
```

Plus duplicate UUID validation, similar error handling.

**Recommendation:**

Create helper functions:
```go
// pkg/handlerutil/helper.go
func UserIDFromCtx(w http.ResponseWriter, r *http.Request) (string, bool) {
    userID, ok := middleware.GetUserIDFromContext(r.Context())
    if !ok {
        response.Error(w, http.StatusUnauthorized, "unauthorized", nil)
        return "", false
    }
    return userID, true
}

func ParseUUID(w http.ResponseWriter, r *http.Request, paramName string) (string, bool) {
    id := chi.URLParam(r, paramName)
    if !isValidUUID(id) {
        response.Error(w, http.StatusBadRequest, "invalid id", nil)
        return "", false
    }
    return id, true
}
```

**Effort:** S (2 days)

---

## 7. Inconsistent Naming (Handler Methods) 🟢 LOW

**Location:** Various handlers

**Issue:**

Some handlers use:
- `h.GetUser(w, r)` — verb + resource
- `h.CreateUser(w, r)` — verb + resource
- `h.ListUsers(w, r)` — verb + plural

Others use:
- `h.Login(w, r)` — bare verb
- `h.RefreshToken(w, r)` — verb + resource

Inconsistent.

**Recommendation:**

Standardize: `h.<Verb><Resource>(w, r)` for all.

**Effort:** XS (1-2 hours)

---

## 8. Missing `golangci-lint` Config 🟡 MEDIUM

**Location:** Repository root

**Issue:**

No `.golangci.yml`. Code style is enforced only by IDE defaults.

**Impact:**
- No automated enforcement of conventions
- No SAST integration
- Slower code review

**Recommendation:**

Create `.golangci.yml`:
```yaml
linters:
  enable:
    - errcheck
    - gosimple
    - govet
    - ineffassign
    - staticcheck
    - unused
    - gosec
    - gocritic
    - bodyclose
    - errorlint
    - nilerr
    - testifylint

linters-settings:
  gocritic:
    enabled-tags:
      - diagnostic
      - performance
      - style

run:
  timeout: 5m
```

The `Makefile` already has `make lint` target that auto-installs it — just needs the config.

**Effort:** XS (1 hour)

---

## 9. Two Casbin Model Files 🟢 LOW

**Location:** `rbac/model.conf`, `rbac/rbac_model.conf`

**Issue:**

Legacy `model.conf` exists alongside active `rbac_model.conf`. Confusing.

**Recommendation:**

Delete `rbac/model.conf` (legacy) after verifying `rbac_model.conf` is the only one referenced in `cmd/serve.go:80`.

**Effort:** XS (15 minutes)

---

## 10. Commented-Out Code in Router 🟢 LOW

**Location:** `interface/http/router.go:334-346`, `cmd/serve.go:196-204, 228`

**Issue:**

Backup routes are commented out. Same in serve.go.

**Recommendation:**

Either wire them up ([bugs.md](./bugs.md#4)) or delete. Version control remembers deleted code.

**Effort:** S (covered in bug fix)

---

## 11. Hardcoded Path Strings in Handlers 🟢 LOW

**Location:** Throughout

**Issue:**
```go
http.ServeFile(w, r, filepath.Join("./uploads", filename))
```

Hardcoded relative paths.

**Recommendation:**

Use config:
```go
http.ServeFile(w, r, cfg.FileUpload.GetFullPath(filename))
```

**Effort:** S (1 day)

---

## 12. Inconsistent JSON Field Naming 🟢 LOW

**Location:** Handlers

**Issue:**

Some fields use snake_case:
```json
"profile_image_url": "..."
"section_name": "..."
```

Others use camelCase inconsistently or are missing.

**Recommendation:**

Add a JSON tag linter or generate from struct tags via `go generate`.

**Effort:** S (1-2 days)

---

## Summary Table

| # | Issue | Priority | Effort |
|---|---|---|---|
| 1 | fmt.Println debug | 🟡 MEDIUM | XS |
| 2 | Inconsistent error wrapping | 🟡 MEDIUM | S |
| 3 | Two repo directories | 🟡 MEDIUM | S |
| 4 | Long handler files | 🟡 MEDIUM | M |
| 5 | Magic permission strings | 🟡 MEDIUM | S |
| 6 | Handler boilerplate | 🟡 MEDIUM | S |
| 7 | Handler naming | 🟢 LOW | XS |
| 8 | No golangci-lint | 🟡 MEDIUM | XS |
| 9 | Two Casbin models | 🟢 LOW | XS |
| 10 | Commented code | 🟢 LOW | S |
| 11 | Hardcoded paths | 🟢 LOW | S |
| 12 | JSON naming | 🟢 LOW | S |