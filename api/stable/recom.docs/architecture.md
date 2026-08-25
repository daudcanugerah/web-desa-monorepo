# Architecture Recommendations

Structural improvements, design patterns, separation of concerns.

---

## 1. Module Name Mismatch 🟡 MEDIUM

**Location:** `go.mod:1`

**Issue:** See [bugs.md](./bugs.md#7)

**Recommendation:** Rename `basic-service` to `webdesa/api` or proper org path.

**Effort:** XS

---

## 2. Generic Settings API Missing 🟡 MEDIUM

**Location:** `db/migrations/00014` (settings table)

**Issue:**

The `settings` table is generic (K-V JSONB) but only one row exists (`desa_profile`). No generic CRUD API.

**Impact:**
- Cannot store other settings (theme, footer text, contact info)
- Forces new tables for each new config

**Recommendation:**

Add full settings feature:
```go
// domain/settings/settings.go
type Setting struct {
    Key       string
    Value     json.RawMessage
    UpdatedAt time.Time
}

func (s *Setting) Validate() error {
    if s.Key == "" {
        return errors.New("key required")
    }
    if s.Value == nil {
        return errors.New("value required")
    }
    return nil
}
```

```go
// usecase/settings/service.go
type Service interface {
    Get(ctx, key string) (*Setting, error)
    Upsert(ctx, key string, value interface{}) error
    Delete(ctx, key string) error
    List(ctx) ([]Setting, error)
}
```

Plus HTTP endpoints:
- `GET /api/v1/settings` (admin only, RBAC: `settings:read`)
- `PUT /api/v1/settings/{key}` (RBAC: `settings:write`)
- `DELETE /api/v1/settings/{key}` (RBAC: `settings:write`)
- `GET /api/v1/public/settings/{key}` (public read of explicitly whitelisted keys)

**Effort:** M (3-5 days)

---

## 3. No Specification Layer (OpenAPI as Single Source of Truth) 🟡 MEDIUM

**Location:** `openapi.yaml`, handlers

**Issue:**

`openapi.yaml` (1,896+ lines) is maintained separately from the code. They can drift. There's `openapi-verification.md` to check alignment, but it's manual.

**Recommendation:**

**Option A — Generate OpenAPI from code annotations:**

Use a Go OpenAPI generator like [`ogen`](https://github.com/ogen-go/ogen) or [`swag`](https://github.com/swaggo/swag):

```go
// @Summary Create a berita
// @Description Create a new berita article
// @Tags berita
// @Accept multipart/form-data
// @Produce json
// @Param title formData string true "Title"
// @Param content formData string true "Content"
// @Param category formData string true "Category UUID"
// @Param image formData file false "Cover image"
// @Success 201 {object} BeritaResponse
// @Failure 400 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/v1/berita [post]
func (h *Handler) CreateBerita(w http.ResponseWriter, r *http.Request) { ... }
```

Then `swag init` regenerates OpenAPI from code annotations.

**Option B — Generate code from OpenAPI:**

Use `ogen` to generate handler interfaces and types from OpenAPI spec, then implement.

**Effort:** L (1-2 weeks for either option)

---

## 4. Domain Entities Have External Dependencies (Some) 🟡 MEDIUM

**Location:** Multiple domain files

**Issue:**

Per Clean Architecture, `domain/` should be pure. But some files import `time`, `errors`, `fmt`. Check for any that import external packages.

**Recommendation:**

Audit and ensure `domain/` imports ONLY:
- `errors`
- `fmt`
- `strings` (sometimes)
- `time`
- Standard library

Never `pkg/`, `usecase/`, `interface/`.

**Effort:** XS (audit + cleanup)

---

## 5. Service Layer Mixes Business Logic with Repository Calls 🟡 MEDIUM

**Location:** Multiple `usecase/*/service.go`

**Issue:**

Some services directly call `s.repo.X()` AND make business decisions in same method. This is fine, but consider introducing:

- **Application services** (orchestration only)
- **Domain services** (pure business logic)
- **Repository adapters** (DB only)

For complex features like PPID approval flow, separation would help:
```go
// pkg/approval/engine.go (pure logic, no DB)
type ApprovalEngine interface {
    ValidateRequest(*PPIDRequest) error
    GenerateToken(secret string, ppidID string) (string, error)
    BuildDownloadLink(domainAddr, ppidID, token string) string
}

// usecase/ppid/service.go (orchestration)
func (s *Service) ApproveRequest(...) error {
    if err := s.engine.ValidateRequest(req); err != nil {
        return err
    }
    token, err := s.engine.GenerateToken(s.jwtSecret, req.PPIDId)
    // ... DB, email ...
}
```

**Effort:** L (significant refactor)

---

## 6. Email Service Tied to PPID Feature 🟡 MEDIUM

**Location:** `interface/email/ppid_email_service.go`

**Issue:**

`PPIDEmailService` is in `ppid_email_service.go` but the auth password reset will also need email. The interface is in `usecase/ppid/email_service.go`.

**Recommendation:**

Generalize:
```
interface/email/  → package email
  client.go       → SMTPClient interface
  smtp_client.go   → SMTPSender implementation (HTML, plain auth)
  templates.go    → Email template registry

usecase/<feature>/email_service.go → feature-specific EmailService interface
```

This way, auth and PPID both use the same SMTP client.

**Effort:** M (3-5 days)

---

## 7. File Handler Has Multiple Implementations Hidden 🟢 LOW

**Location:** Each feature's `usecase/<feature>/file_handler.go`

**Issue:**

Every feature defines its own `FileHandler` interface. While each has the same methods (`SaveImage`, `Delete`), they are duplicate types:

```go
// usecase/berita/file_handler.go
type FileHandler interface {
    SaveImage(ctx, name string, content io.Reader, size int64, contentType string) (string, error)
    Delete(ctx, path string) error
}

// usecase/banner/file_handler.go  
type FileHandler interface {  // IDENTICAL
    SaveImage(...)
    Delete(...)
}
```

**Recommendation:**

Define ONE interface in `pkg/filehandler/`:
```go
// pkg/filehandler/handler.go
type Handler interface {
    SaveImage(ctx, name string, content io.Reader, size int64, contentType string) (string, error)
    SaveDocument(...) (string, error)
    Delete(ctx, path string) error
    GetFilePath(relPath string) string
}
```

All features reference this. Single `LocalHandler` from `interface/file/` implements it.

**Effort:** S (2-3 days)

---

## Summary Table

| # | Issue | Priority | Effort |
|---|---|---|---|
| 1 | Module name mismatch | 🟡 MEDIUM | XS |
| 2 | Generic settings API | 🟡 MEDIUM | M |
| 3 | OpenAPI drift | 🟡 MEDIUM | L |
| 4 | Domain dep audit | 🟡 MEDIUM | XS |
| 5 | Service layer purity | 🟡 MEDIUM | L |
| 6 | Email service generalization | 🟡 MEDIUM | M |
| 7 | File handler consolidation | 🟢 LOW | S |

## Architectural Strengths (Keep)

- ✅ Strict Clean Architecture boundaries
- ✅ Interface co-location in usecase layer
- ✅ "Accept interfaces, return structs" pattern
- ✅ Time abstraction via `pkg/clock`
- ✅ Single-shared test container pattern
- ✅ Token-bucket rate limiting
- ✅ pg_dump/pg_restore for backups (vs application-level)
- ✅ Domain validates entities at construction