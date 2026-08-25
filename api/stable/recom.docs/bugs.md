# Bug Fixes

Known bugs, incorrect behavior, schema mismatches, and broken features.

---

## 1. Password Reset Token Never Emailed 🔴 CRITICAL

**Location:** `usecase/auth/service.go:171`

**Issue:**
```go
// TODO: Send email with reset token (email service integration)
```

The `RequestPasswordReset` method persists a reset token to the `password_reset_tokens` table but **never emails it to the user**. The HTTP handler returns 204 No Content, masking the issue from operators. Users cannot actually reset their passwords because they have no way to retrieve the token.

**Impact:**
- Users locked out of accounts with no recovery path
- Silent failure — operators have no visibility
- Production-blocking for any real deployment

**Recommendation:**

1. **Short term:** Reuse the existing `PPIDEmailService` pattern. Create a generic `EmailService` interface and an `AuthEmailService` implementation.
2. **Wire it up:** Inject the email service into `auth.Service`, call it after token generation, return error if SMTP fails.
3. **Update handler:** Return 500 if email fails (don't pretend success).
4. **Add template:** "Click here to reset your password" with the reset link.

**Implementation sketch:**

```go
// domain/auth/email_service.go (new)
type EmailService interface {
    SendPasswordResetEmail(ctx, recipient, token string) error
}

// usecase/auth/service.go
func (s *Service) RequestPasswordReset(ctx, email, ip, ua string) error {
    // ... existing token generation ...
    
    resetLink := fmt.Sprintf("%s/reset-password?token=%s", s.domainAddr, token)
    if err := s.emailService.SendPasswordResetEmail(ctx, user.Email, resetLink); err != nil {
        return fmt.Errorf("failed to send reset email: %w", err)
    }
    return nil
}
```

**Effort:** S (1-2 days)

---

## 2. PPID Upload Directory Returns Wrong Path 🟠 HIGH

**Location:** `config/fileupload.go:35-40`

**Issue:**
```go
func (c *FileUploadConfig) GetPPIDUploadDirectory() string {
    return c.UploadPublicDirectory  // ← BUG: should be UploadPPIDDirectory
}
```

The method intended to return the PPID-specific upload subdirectory (`./uploads/ppid`) instead returns the same path as the public directory. However, in `cmd/serve.go:97-101`, the PPID file handler is correctly initialized via:

```go
ppidFileHandler, err := filehandler.NewLocalHandlerWithSubdir(
    systemConfig.FileUpload.GetPPIDUploadDirectory(),
    "",  // subdir
)
```

Because of the bug, PPID files actually go into `./uploads/` (same as public), not `./uploads/ppid/`.

**Impact:**
- PPID documents not isolated from public images
- Directory traversal protection weakens (public endpoint can serve PPID files)
- Defeats the intent of having a separate PPID directory

**Recommendation:**

1. **Fix the method:**
   ```go
   func (c *FileUploadConfig) GetPPIDUploadDirectory() string {
       if c.UploadPPIDDirectory == "" {
           return filepath.Join(c.UploadDirectory, "ppid")
       }
       return c.UploadPPIDDirectory
   }
   ```

2. **Add `UploadPPIDDirectory` field** to config struct.

3. **Migrate existing PPID files** from `./uploads/` to `./uploads/ppid/` (if any).

4. **Add integration test** verifying PPID files land in the correct directory.

**Effort:** XS (1 hour)

---

## 3. Struktur `Description` Field Has No DB Column 🟠 HIGH

**Location:** `domain/struktur/struktur.go:16-26`

**Issue:**
```go
type Struktur struct {
    // ...
    Description *string  // db:"-" — NOT in DB schema
    // ...
}
```

The `Description` field exists in the domain entity with `db:"-"` (explicitly excluded from DB scan), but `Validate()` includes it. The field is **silently dropped** at scan time and never persisted.

**Impact:**
- Operators may set descriptions via API (validated) but they never save
- Confusing behavior — no error, just data loss
- Schema/code drift

**Recommendation:**

**Option A — Add the column (preferred):**
```sql
-- db/migrations/00037_add_description_to_struktur.sql
ALTER TABLE struktur_organisasi
    ADD COLUMN description TEXT;
```

Then remove `db:"-"` from the struct.

**Option B — Remove the field:**
```go
type Struktur struct {
    ID, Name, Position, Email, Phone string
    ProfileImageURL *string
    CreatedAt, UpdatedAt time.Time
}
// Remove Description and update Validate()
```

**Decision criteria:**
- If the frontend displays `description` → Option A
- If unused → Option B (cleaner)

**Effort:** XS (1-2 hours)

---

## 4. Backup HTTP Routes Commented Out 🟠 HIGH

**Location:** `interface/http/router.go:334-346`, `cmd/serve.go:196-204, 228`

**Issue:**
- `interface/http/handler/backup.go` exists (202 lines, fully implemented)
- `cmd/serve.go` has commented-out `BackupHandler` construction: `// Will be used when backup service is enabled`
- `router.go` has commented route registrations

The handler is dead code. Backup functionality is **CLI-only** despite the handler being ready.

**Impact:**
- Misleading code — looks complete but isn't wired
- Frontend cannot trigger backups via HTTP
- Operator must SSH into server to run backups

**Recommendation:**

**Option A — Wire it up:**
1. Uncomment handler construction in `cmd/serve.go:196-204, 228`
2. Uncomment route registrations in `router.go:334-346`
3. Add `backup:read` and `backup:write` RBAC permissions
4. Add to seed.go default permissions
5. Add integration tests

**Option B — Remove dead code:**
1. Delete `interface/http/handler/backup.go`
2. Remove commented lines from `cmd/serve.go` and `router.go`
3. Update `openapi.yaml` to reflect CLI-only status

**Decision criteria:**
- If admin UI needs backup → Option A
- If CLI is sufficient → Option B

**Effort:** S (1-2 days for Option A)

---

## 5. Active Banner Count Race Condition 🟠 HIGH

**Location:** `usecase/banner/service.go:220-228`

**Issue:**
```go
func (s *Service) UpdateStatus(ctx, id, status) error {
    if status == "active" {
        count, _ := s.repo.CountActiveBanners(ctx)
        if count >= 20 {
            return errors.New("cannot activate banner: maximum of 20 active banners reached")
        }
    }
    // ... update happens later ...
}
```

The count + update is **not transactional**. Two concurrent requests can both pass the check (count=19) and both activate, resulting in 21 active banners.

**Impact:**
- 20-banner limit can be exceeded under concurrent load
- Inconsistent state possible

**Recommendation:**

Wrap in a transaction with `SELECT FOR UPDATE`:

```go
func (s *Service) UpdateStatus(ctx, id, status) error {
    tx, err := s.db.BeginTxx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback()
    
    if status == "active" {
        var count int
        err := tx.QueryRowxContext(ctx, 
            "SELECT COUNT(*) FROM banners WHERE status = $1 FOR UPDATE", 
            "active").Scan(&count)
        if err != nil {
            return err
        }
        if count >= 20 {
            return errors.New("maximum of 20 active banners reached")
        }
    }
    
    _, err = tx.ExecContext(ctx, 
        "UPDATE banners SET status=$1, updated_at=NOW() WHERE id=$2", 
        status, id)
    if err != nil {
        return err
    }
    
    return tx.Commit()
}
```

**Alternative — Postgres-level constraint:**

```sql
-- db/migrations/00037_add_active_banner_limit.sql
CREATE OR REPLACE FUNCTION enforce_active_banner_limit()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.status = 'active' AND 
       (SELECT COUNT(*) FROM banners WHERE status = 'active' AND id != NEW.id) >= 20 THEN
        RAISE EXCEPTION 'cannot activate: maximum of 20 active banners reached';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_active_banner_limit
BEFORE INSERT OR UPDATE ON banners
FOR EACH ROW
EXECUTE FUNCTION enforce_active_banner_limit();
```

**Effort:** S (2-4 hours)

---

## 6. PPID Email Sent Before DB Update — Inconsistent State 🟡 MEDIUM

**Location:** `usecase/ppid/service.go:507-571`

**Issue:**

The approval flow sends the email BEFORE updating the DB:

```go
func (s *Service) ApproveRequest(ctx, requestID, adminUserID) error {
    // ... load request, generate token ...
    
    // Email sent FIRST
    if err := s.emailService.SendApprovalEmail(...); err != nil {
        return err  // Status stays "pending"
    }
    
    // Then status updated
    s.repo.UpdateRequest(ctx, requestID, "approved", ...)
}
```

If the DB update fails after the email is sent, the user receives a download link for a request that's still marked "pending". Conversely, if the email fails, the operator sees an error and the request stays pending — but the user never gets the link.

**Impact:**
- Inconsistent state possible (email sent but DB not updated)
- Operator confusion when retries are needed

**Recommendation:**

**Option A — Outbox pattern (best):**
```sql
CREATE TABLE email_outbox (
    id UUID PK,
    recipient VARCHAR(255),
    subject VARCHAR(500),
    body TEXT,
    status VARCHAR(20),  -- pending|sent|failed
    retry_count INT DEFAULT 0,
    next_attempt_at TIMESTAMP,
    created_at TIMESTAMP
);
```

1. Write email to outbox in same transaction as DB update
2. Background worker sends from outbox
3. Failed sends retried with backoff

**Option B — Transactional email + compensate:**

1. Send email
2. If email succeeds, update DB in transaction
3. If DB fails, send "apology" email
4. Accept eventual consistency

**Option C — Reverse the order (simplest):**

1. Update DB first (mark approved, save token)
2. Try to send email
3. If email fails, log + return error, but DB stays approved (operator can manually resend)

**Effort:** M (2-3 days for Option A; 1 day for Option C)

---

## 7. Module Name Mismatch 🟡 MEDIUM

**Location:** `go.mod:1`

**Issue:**
```
module basic-service
```

But the folder is `webdesa/api/`. Likely from a template that was renamed.

**Impact:**
- Import paths are `basic-service/...` throughout the code
- Confusing for new contributors
- Wrong module name in published artifacts

**Recommendation:**

Rename the module:
```bash
# 1. Update go.mod
module github.com/<org>/webdesa-api
# or
module webdesa/api

# 2. Update all imports (in all .go files)
find . -name "*.go" -type f -exec sed -i '' 's|basic-service/|webdesa/api/|g' {} +

# 3. Verify
grep -r "basic-service" --include="*.go" .
go build ./...
go test ./...
```

**Effort:** XS (1 hour)

---

## 8. PPID `category` Column Nullable But Used Everywhere 🟢 LOW

**Location:** `db/migrations/00032`, `domain/ppid/ppid.go`

**Issue:**

The PPID `category` column is nullable (preserved from migration 00032), but every other content feature uses NOT NULL. This creates inconsistency.

**Impact:**
- Frontend must handle null category for PPID only
- Inconsistent validation behavior

**Recommendation:**

If business logic allows PPID without category, keep as-is. Otherwise:
```sql
-- db/migrations/00038_make_ppid_category_not_null.sql
UPDATE ppid SET category = (SELECT id FROM ppid_categories WHERE name = 'Tanpa Kategori')
WHERE category IS NULL;
ALTER TABLE ppid ALTER COLUMN category SET NOT NULL;
```

**Effort:** XS (1 hour)

---

## Summary Table

| # | Issue | Priority | Effort | Location |
|---|---|---|---|---|
| 1 | Password reset never emails | 🔴 CRITICAL | S | `usecase/auth/service.go:171` |
| 2 | PPID upload dir wrong | 🟠 HIGH | XS | `config/fileupload.go:35-40` |
| 3 | Struktur.Description no column | 🟠 HIGH | XS | `domain/struktur/struktur.go:16-26` |
| 4 | Backup routes commented out | 🟠 HIGH | S | `router.go:334-346` |
| 5 | Banner count race | 🟠 HIGH | S | `usecase/banner/service.go:220-228` |
| 6 | PPID email/DB inconsistency | 🟡 MEDIUM | M | `usecase/ppid/service.go:507-571` |
| 7 | Module name mismatch | 🟡 MEDIUM | XS | `go.mod:1` |
| 8 | PPID category nullable | 🟢 LOW | XS | `db/migrations/00032` |