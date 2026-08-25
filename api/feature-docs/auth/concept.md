# Auth — Concept

JWT-based authentication with password reset flow. Public endpoints only — no RBAC required; rate-limited at HTTP and business levels.

**Module:** webdesa/api · **Last updated:** 2026-08-23

## Purpose

Single sign-on entry point for the API. Provides:
- Login → access/refresh JWT pair
- Refresh → new access token from valid refresh token
- Password reset request → email with token link
- Password reset confirm → new password
- Token validity check before user clicks the link

Auth is the only feature where business and HTTP rate limits are both active and chained (5/min auth bucket sits inside the 100/min public bucket).

## Business Rules

### JWT TTL source

Access-token expiration is wired from `systemConfig.JWT.GetExpiration()` (`cmd/serve.go:214`). The default is **24h**, not 1h. The email body hardcodes "1 hour" — see [endpoint.md](./endpoint.md#known-issues).

Refresh-token TTL lives in `pkg/jwt/jwt.go:34-41` (7 days).

### Reset-token TTL bug

The password reset token's TTL is also wired from the same `systemConfig.JWT.GetExpiration()` — meaning reset tokens live **24h**, not the documented "1 hour" implied by the email body. Users can take up to a full day to use the link.

### Login credential failure

`Login` returns the generic string `"Invalid credentials"` for *every* failure mode: unknown email, wrong password, missing user. No enumeration possible.

### Login password length

`LoginRequest` enforces `validate:"required,min=6"` on `password`. A password < 6 chars returns **400** (handler validation), not 401. Legacy short passwords fail at login.

### Reset token behavior

- 32-byte hex token, stored in `password_reset_tokens` with `used=false`.
- `CheckResetToken` **never errors** — returns `{Valid:false}` for missing/expired/used. UI uses this to decide whether to show the confirm form.
- `ResetPassword` bcrypt-hashes new password, marks the token `used=true`. Multiple `used=true` flags on the same token are not prevented by the schema; only `token` has uniqueness.

### Email-send failure surface

If `email.SendPasswordResetEmail` fails (SMTP down, misconfigured), the handler returns **500** with the actual SMTP error in the log. Previously was swallowed silently → 200. Regression test `TestRequestPasswordReset_EmailFailureSurfacesError` locks this in.

`IsConfigured()` returns true only if host/port/username/password are all non-empty. When nil `EmailSender` is passed (legacy mode) `RequestPasswordReset` still returns 200 silently — covered by `TestRequestPasswordReset_NoEmailSender_LegacyBehavior`.

### Rate-limit key behavior

The token bucket key is `r.RemoteAddr` = `IP:port` (`ratelimit.go:125`). Go's `http.Request.RemoteAddr` includes the source port → each TCP connection gets its **own** bucket rather than per-IP.

Behind a reverse proxy that reuses source ports this collapses to per-IP. Behind direct connections the limit is effectively per-connection — see [endpoint.md](./endpoint.md#known-issues).

### Rate-limit inheritance chain

```
PublicRateLimitMiddleware (100 req/min)   ← inherited by all /api/v1/*
   └─ AuthRateLimitMiddleware (5 req/min) ← /auth/* only (middleware_deps.go:35)
       └─ handler
```

A client that exhausts the public bucket is blocked **before** reaching the auth bucket.

Business-layer rate limit (separate from HTTP): **5 reset requests per hour per email**, DB-counted against `password_reset_tokens.created_at`, not in-memory.

### Response envelope

All responses use `{success:true, data:{...}}`. Errors use `{success:false, error:"..."}`.

## Architecture

### Files

| File | Purpose |
|---|---|
| `interface/http/handler/auth/auth.go` | `AuthHandler` + handler methods (package `auth`). Unused `isValidEmail` at line 311. |
| `interface/http/handler/auth/route.go` | `RegisterRoutes` for `/auth/*` |
| `usecase/auth/service.go` | Login, RefreshToken, RequestPasswordReset, CheckResetToken, ResetPassword |
| `usecase/auth/repository.go` | `Repository` interface (reset tokens + password update) |
| `usecase/auth/user_repository.go` | `UserRepository` interface (`FindByEmail`) |
| `usecase/auth/email_sender.go` | `EmailSender` interface (auth-specific subset) |
| `usecase/auth/service_test.go` | Unit tests (4 cases) |
| `interface/email/auth_email_service.go` | `AuthEmailService` — sends password reset email. Body hardcodes "1 hour". |
| `interface/email/smtp_sender.go` | `SMTPSender` — shared SMTP client wrapper |
| `interface/email/ppid_email_service.go` | `PPIDEmailService` — also uses `SMTPSender` |
| `interface/postgres/auth.go` | sqlx impl. Dead `FindUserByEmail` method (not in any interface). |
| `domain/auth/types.go` | `TokenPair`, `AccessToken`, `ResetToken`, `ResetTokenInfo`, `ResetMetadata` |
| `domain/user/user.go` | `User` entity + `Validate()` |
| `pkg/password/bcrypt.go` | bcrypt wrapper (cost 12) |
| `pkg/jwt/jwt.go` | JWT HS256 generation/validation |

### Service struct deps

`usecase/auth/service.go` — `Service{repo, userRepo, email, clock, jwtSecret, resetTokenTTL, resetURLBase, rateLimit{5, 1h}}`

### Repository interface methods

`usecase/auth/repository.go` (`Repository`):
- `CreateResetToken(ctx, *ResetToken) error`
- `GetResetTokenByToken(ctx, string) (*ResetToken, error)`
- `MarkResetTokenUsed(ctx, string) error`
- `UpdateUserPassword(ctx, userID string, hashedPassword string) error`
- `CountRecentResetRequests(ctx, email string, since time.Time) (int, error)`

`usecase/auth/user_repository.go` (`UserRepository`):
- `FindByEmail(ctx, email string) (*domain.User, error)`

`usecase/auth/email_sender.go` (`EmailSender`):
- `SendPasswordResetEmail(ctx, to, token, baseURL string) error`

### Password Reset Flow

```
1. POST /auth/password-reset/request → HTTP 5/min + business 5/hour per email
2. Token persisted to password_reset_tokens table
3. Email sent with reset link: {resetURLBase}/auth/password-reset/confirm?token=...
4. User clicks link → POST /auth/password-reset/confirm {token, new_password}
5. Token marked used; password updated
```

If email send fails (SMTP down, etc.), the handler returns **500** instead of the previous silent 200. Operators see the actual SMTP error in logs.

## Glossary

- **Access token** — short-lived JWT (24h, wired from `systemConfig.JWT.GetExpiration()`). Carried as `Authorization: Bearer`.
- **Refresh token** — long-lived JWT (7d). Used at `/auth/refresh` to mint new access tokens.
- **Reset token** — opaque 32-byte hex, stored in `password_reset_tokens`. Distinct from JWT; not a JWT.
- **Public bucket** — 100 req/min rate-limit inherited by every `/api/v1/*` route via `PublicRateLimitMiddleware`.
- **Auth bucket** — 5 req/min rate-limit on `/auth/*` via `AuthRateLimitMiddleware`. Sits inside the public bucket.
- **EmailSender** — auth-package interface; nil = legacy silent-success mode.

## Related

- [database.md](./database.md)
- [endpoint.md](./endpoint.md)
- `pkg/jwt/jwt.go:34-41` — Token generation (24h access / 7d refresh)
- `pkg/password/bcrypt.go` — bcrypt wrapper (cost 12)
- `interface/http/middleware/auth.go` — JWT bearer parsing
- `interface/http/middleware/ratelimit.go` — Token-bucket rate limiter
- `cmd/seed.go` — creates `admin@desa.local` / `operator@desa.local` defaults
- [users.md (profile_image_media_id)](../users/concept.md)
- [rbac.md (admin role)](../rbac/concept.md)
