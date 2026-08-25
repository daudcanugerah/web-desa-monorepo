# PPID Feature

Public Information Disclosure (Pejabat Pengelola Informasi dan Dokumentasi) — public document requests with approval workflow, email notifications, and JWT-signed download links.

**Module:** `webdesa/api`
**Last updated:** 2026-07-04

## Files

| File | Purpose |
|---|---|
| `interface/http/handler/ppid/ppid.go` | `PPIDHandler` struct + methods (package `ppid`) |
| `interface/http/handler/ppid/ppidcategory.go` | `PPIDCategoryHandler` |
| `interface/http/handler/ppid/route.go` | `RegisterRoutes` for `/ppid/*` + `/public/ppid/*` |
| `domain/ppid/ppid.go` | `PPID` + `PPIDRequest` entities + `Validate()` |
| `domain/ppidcategory/ppidcategory.go` | `Category` entity |
| `usecase/ppid/service.go` | Approval workflow + JWT generation |
| `usecase/ppid/email_service.go` | `EmailService` interface |
| `usecase/ppid/file_handler.go` | `FileHandler` interface |
| `usecase/ppid/repository.go` | `Repository` interface |
| `usecase/ppid/category_lookup.go` | `CategoryLookup` interface |
| `usecase/ppidcategory/service.go` | Category CRUD |
| `interface/postgres/ppid.go` | sqlx implementation |
| `interface/postgres/ppidcategory.go` | sqlx implementation |
| `interface/email/ppid_email_service.go` | `PPIDEmailService` — sends approval email via `SMTPSender` |
| `interface/email/smtp_sender.go` | Shared `SMTPSender` |
| `db/migrations/00009`, `00017`, `00024`, `00026`, `00031`, `00032` | |

## Database

### `ppid` (migrations 00009, 00017, 00024, 00026, 00032)

| Column | Type | Notes |
|---|---|---|
| `id` | UUID PK | `gen_random_uuid()` |
| `title` | VARCHAR(255) NOT NULL | indexed |
| `category` | UUID | FK → `ppid_categories(id)` ON DELETE RESTRICT — **nullable** |
| `file_url` | VARCHAR(500) | nullable — actual document |
| `thumbnail_url` | VARCHAR(500) | nullable, added 00026 |
| `description` | TEXT | added 00024 |
| `publication_at` | TIMESTAMP | nullable, added 00017 |
| `created_at` | TIMESTAMP | |
| `updated_at` | TIMESTAMP | |

### `ppid_requests` (migrations 00010, 00025)

| Column | Type | Notes |
|---|---|---|
| `id` | UUID PK | `gen_random_uuid()` |
| `ppid_id` | UUID FK | → `ppid(id)` ON DELETE CASCADE |
| `requester_name` | VARCHAR(255) NOT NULL | |
| `requester_email` | VARCHAR(255) NOT NULL | |
| `notes` | TEXT | (domain field: `purpose`) |
| `status` | VARCHAR(20) NOT NULL DEFAULT `'pending'` | `pending\|approved\|revoked` |
| `approved_at` | TIMESTAMP | nullable |
| `approved_by` | UUID FK | → `users(id)` ON DELETE SET NULL |
| `revoked_at` | TIMESTAMP | nullable |
| `revoked_by` | UUID FK | → `users(id)` ON DELETE SET NULL |
| `created_at` | TIMESTAMP | |

### `ppid_categories` (migration 00031)

Standard category table. **Seeded**: `"Tanpa Kategori"`.

## Domain Entities

### `PPID`
```go
type PPID struct {
    ID, Title string
    Category, DocumentURL, ThumbnailURL *string  // db:"file_url" for DocumentURL
    Description string
    PublicationAt *time.Time
    CreatedAt, UpdatedAt time.Time
}
```

**Validation** (`domain/ppid/ppid.go:199-225`):
- Thumbnail must end in `.jpg/.jpeg/.png/.gif/.webp`
- Description non-empty if present

### `PPIDRequest`
```go
type PPIDRequest struct {
    ID, PPIDId, RequesterName, RequesterEmail string
    Purpose *string                              // db:"notes"
    Status string                                 // pending|approved|revoked
    ApprovedAt *time.Time
    ApprovedBy *string
    RevokedAt *time.Time
    RevokedBy *string
    CreatedAt time.Time
}
```

Status validation: if `approved` then both `ApprovedAt` + `ApprovedBy` set; same for `revoked`.

## Service

`usecase/ppid/service.go` — `Service{repo, fileHandler, emailService EmailService, categoryRepo CategoryLookup, clock, jwtSecret, emailConfig}`

### Document Management

| Method | Purpose |
|---|---|
| `Create(CreatePPIDInput)` | Saves document (PDF/DOC/DOCX/XLS/XLSX ≤50MB), saves thumbnail if provided; rolls back document if thumbnail save fails |
| `Update(ctx, id, UpdatePPIDInput)` | Handles `ClearPublicationAt bool` flag to set `publication_at=nil` |
| `Delete` | Removes file |

### Request Workflow

| Method | Purpose |
|---|---|
| `CreateRequest(CreateRequestInput)` | **Public**, no auth. Creates request with `status="pending"` |
| `ListRequests(ListRequestsInput)` | `statuses []string`, `page`, `limit` — `status IN (...)` |
| `ApproveRequest(ctx, requestID, adminUserID)` | See approval flow below |
| `RevokeRequest(ctx, requestID, adminUserID)` | Sets `revoked_at`, `revoked_by`, status=revoked |

### Approval Flow (`service.go:507-571`)

1. Load request → verify `pending`
2. `GenerateAccessToken(req.PPIDId)` — HS256 JWT, 10-min expiry
3. Load PPID (for title)
4. Build `downloadLink = {DomainAddr}/api/v1/ppid/document/{docID}/download?token={jwt}`
5. **`emailService.SendApprovalEmail` happens BEFORE DB update** — if SMTP fails, status stays pending
6. Set `status="approved"`, `approved_at=now`, `approved_by=adminUserID`, persist

### Download (`service.go:617-675`)

`DownloadDocument(DownloadDocumentInput)` accepts either:
1. **PPID-request-issued JWT** (10-min, single-document)
2. **Regular user JWT** (admin access path)

Returns `*DownloadDocumentResult{FilePath, ContentType, Filename}`.

### JWT Token Generation

`GenerateAccessToken(ppidID)` (`service.go:423-445`):
```go
type AccessTokenClaims struct {
    DocumentId string  // set to ppidID
    UserID     string  // also set to ppidID (legacy field)
    jwt.RegisteredClaims  // iat, exp, jti (random UUID)
}
```

HS256 signed, 10-minute expiry.

## Email Service Interface

`usecase/ppid/email_service.go`:
```go
type EmailService interface {
    SendApprovalEmail(ctx, recipient, requesterName, ppidTitle, downloadLink string) error
}
```

Implementation: `interface/email/ppid_email_service.go` — `PPIDEmailService` (HTML email via `net/smtp`, configured via `config.SMTPConfig`).

## Endpoints

### Public (no auth)

| Method | Path | Notes |
|---|---|---|
| GET | `/api/v1/public/ppid/list` | Description truncated to 500 chars |
| GET | `/api/v1/public/ppid/{id}` | Full description |
| GET | `/api/v1/public/ppid/categories` | With `q` autocomplete |
| POST | `/api/v1/public/ppid/{id}/requests` | Body `{requester_name, requester_email, purpose}` |
| GET | `/api/v1/ppid/document/{documentId}/download` | Dual token via `?token=` OR `Authorization: Bearer` |

### Admin (RBAC: `ppid:read`)

| Method | Path | Query |
|---|---|---|
| GET | `/api/v1/ppid` | `page`, `limit`, `category`, `q` |
| GET | `/api/v1/ppid/{id}` | Full PPIDResponse |
| GET | `/api/v1/ppid/requests` | `status` (repeatable), `page`, `limit` |

### Admin (RBAC: `ppid:write`)

| Method | Path | Notes |
|---|---|---|
| POST | `/api/v1/ppid` | Multipart. Required: `title`, `document` (PDF/DOC/DOCX/XLS/XLSX ≤50MB). Optional: `category` (UUID), `description`, `publication_at` (RFC3339), `thumbnail` |
| PUT | `/api/v1/ppid/{id}` | Multipart. Required: `title` |
| DELETE | `/api/v1/ppid/{id}` | |
| POST | `/api/v1/ppid/requests/{id}/approve` | |
| POST | `/api/v1/ppid/requests/{id}/revoke` | |
| POST | `/api/v1/ppid/categories` | Body `{name}` |
| DELETE | `/api/v1/ppid/categories/{id}` | 409 if in use |

## Response Variants

| Variant | Used in | Description |
|---|---|---|
| `PPIDResponse` | Detail (admin + public) | Full description included |
| `PPIDResponseTruncated` | Public list | Description included, ≤500 chars |
| `PPIDResponsePrivate` | Admin list | No description field |

`truncateDescription(*string) *string` at handler.go:576-594 clamps to 500 chars at last space, appends "...".

## Seed Data

`cmd/seed.go:265-271`:
```
"Anggaran", "Peraturan", "Laporan", "Profil", "Keuangan"
```

## RBAC Permissions

- `ppid:read`
- `ppid:write`

## Business Logic

- Email sent BEFORE DB update in approval flow — if SMTP fails, status stays pending
- PPID `category` column is **nullable** (preserved from migration 00032) — see `recom.docs/bugs.md#8`
- Dual-token download accepts either PPID-request JWT (10-min) or user JWT (admin)

## Tests

**Integration**:
- `integration/ppid_test.go` (16 cases) — happy path, categories, publication_at, timestamps
- `integration/ppid_approval_test.go` (16 cases) — full workflow, approve/revoke, expired token, download with/without token
- `integration/ppidcategory_test.go` (9 cases) — standard category CRUD

## Related

- `interface/email/ppid_email_service.go` — approval email sender
- `interface/email/smtp_sender.go` — shared SMTP wrapper
- `pkg/jwt/jwt.go` — HS256 JWT signing/verification