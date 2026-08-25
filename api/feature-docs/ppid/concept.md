# PPID — Concept

Public Information Disclosure (Pejabat Pengelola Informasi dan Dokumentasi) — public document requests with approval workflow, email notifications, and signed-URL download links (gallery-backed).

**Module:** `webdesa/api`
**Last updated:** 2026-08-23

## Purpose

Lets citizens browse public-information documents and submit disclosure requests. Admins review requests, approve them, and the citizen receives a signed-URL download link valid for one hour. Approvals can be revoked — already-minted links die immediately via a JWT deny-list.

## Business Rules

- **Approval status**: `pending | approved | revoked`. Transition requires both `*_at` + `*_by` to be set together (`pending → approved`, `approved → revoked`).
- **Email-before-DB on approval**: `emailService.SendApprovalEmail` must succeed before the row flips to `approved`. SMTP failure → status stays `pending`, handler returns 500 (`strings.Contains "failed to send approval email"`).
- **Revoke kills live tokens**: `RevokeRequest` pushes the request id onto the gallery signed-URL deny-list (`sub=ppid_request:<id>`) so any JWT already minted becomes invalid.
- **Document media required**: `DocumentMediaID` non-nil at `Validate()` time. Legacy `file_url` column dropped in 00045.
- **Thumbnail optional**: `ThumbnailMediaID` may be nil. Whitelist (`jpg/jpeg/png/gif/webp`) is dead code — column dropped.
- **Category nullable**: `category` column preserved nullable (00032); see `recom.docs/bugs.md#8`.
- **Upload mutex**: each upload slot accepts either a raw multipart file OR a prior `media_id` — both supplied → rejected.
- **3 upload endpoints**: `/ppid/upload-media`, `/ppid/upload-thumbnail`, `/ppid/upload` (legacy alias).

## Architecture

### Files

| File | Purpose |
|---|---|
| `interface/http/handler/ppid/ppid.go` | `PPIDHandler` struct + methods (package `ppid`) |
| `interface/http/handler/ppid/ppid_upload.go` | `PPIDUploadHandler` — multipart upload to gallery |
| `interface/http/handler/ppid/ppidcategory.go` | `PPIDCategoryHandler` |
| `interface/http/handler/ppid/route.go` | `RegisterRoutes` for `/ppid/*` + `/public/ppid/*` |
| `domain/ppid/ppid.go` | `PPID` + `PPIDRequest` entities + `Validate()` |
| `domain/ppidcategory/ppidcategory.go` | `Category` entity |
| `usecase/ppid/service.go` | Approval workflow + signed-URL minting |
| `usecase/ppid/email_service.go` | `EmailService` interface |
| `usecase/ppid/file_handler.go` | `FileHandler` interface |
| `usecase/ppid/repository.go` | `Repository` interface |
| `usecase/ppid/category_lookup.go` | `CategoryLookup` interface |
| `usecase/ppidcategory/service.go` | Category CRUD |
| `interface/postgres/ppid.go` | sqlx implementation |
| `interface/postgres/ppidcategory.go` | sqlx implementation |
| `interface/email/ppid_email_service.go` | `PPIDEmailService` — sends approval email via `SMTPSender` |
| `interface/email/smtp_sender.go` | Shared `SMTPSender` |
| `db/migrations/00009`, `00010`, `00017`, `00024`, `00025`, `00026`, `00031`, `00032`, `00043`, `00045` | |

### Service Dependencies

`Service{repo, fileHandler, emailService EmailService, categoryRepo CategoryLookup, clock, signedURL *galleryUsecase.SignedURLService, emailConfig}`.

- `signedURL` is nilable — when wired (default `cmd/serve.go`), approval flow uses `signedURL.Sign(ScopePPID, …)`. When nil, legacy HS256 fallback path activates.
- `emailService` interface: `SendApprovalEmail(ctx, SendApprovalEmailInput) error`.
- `clock` injected for testable timestamps on `approved_at` / `revoked_at`.

### Repository Interface Methods

`usecase/ppid/repository.go` — covers both `PPID` and `PPIDRequest`:

- `Create(ctx, *PPID) error`
- `Update(ctx, *PPID) error`
- `Delete(ctx, id string) error`
- `GetByID(ctx, id string) (*PPID, error)`
- `List(ctx, ListPPIDInput) ([]PPID, int64, error)`
- `CreateRequest(ctx, *PPIDRequest) error`
- `GetRequestByID(ctx, id string) (*PPIDRequest, error)`
- `ListRequests(ctx, ListRequestsInput) ([]PPIDRequest, int64, error)`
- `UpdateRequestStatus(ctx, id string, status string, actorID string, at time.Time) error`

`CategoryLookup`: `GetByID(ctx, id string) (*Category, error)` (shared with `usecase/ppidcategory`).

## Glossary

- **Metabase JWT** — HS256 token signed with `MetabaseConfig.SecretKey`. Carries `resource: {dashboard|question: id}`, `params`, `exp`. Used for embedding Metabase dashboards/questions directly. See `feature-docs/infographic/concept.md`.
- **Signed-URL scope** — every gallery signed JWT carries a `scope` claim (`public`, `ppid`, `internal`, etc.) and a `sub` claim. The handler verifies scope matches the route; `sub` controls deny-list hits.
- **Request ID lifecycle** — `ppid_requests.id` flows through three states. Pending → `ApproveRequest` mints JWT with `sub=ppid_request:<id>` and emails link. Revoked → that exact `sub` is pushed onto the gallery deny-list, invalidating all live JWTs for that request.
- **Legacy HS256 fallback** — `GenerateAccessToken` (`service.go:530-551`) builds the old `AccessTokenClaims`. Bypassed when `SignedURLService` is wired; only reachable in tests that construct `Service` with `signedURL == nil`.

## Related

- `feature-docs/infographic/concept.md` — Metabase JWT sibling (different scope, same signing primitive)
- `interface/email/ppid_email_service.go` — approval email sender
- `interface/email/smtp_sender.go` — shared SMTP wrapper
- `usecase/gallery/signed_url.go` — `SignedURLService` + `ScopePPID` constant
- `interface/http/handler/gallery/route.go` — `/api/v1/media/{id}/content|thumbnail` handler
- `pkg/jwt/jwt.go` — HS256 JWT primitives (still used by `GenerateAccessToken` legacy fallback)
- `feature-docs/ppid/database.md`
- `feature-docs/ppid/endpoint.md`
