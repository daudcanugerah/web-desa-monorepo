# Auth Feature

JWT-based authentication with password reset flow.

**Module:** `webdesa/api`
**Last updated:** 2026-07-04

## Files

| File | Purpose |
|---|---|
| `interface/http/handler/auth/auth.go` | `AuthHandler` struct + handler methods (package `auth`) |
| `interface/http/handler/auth/route.go` | `RegisterRoutes` for `/auth/*` (package `auth`) |
| `usecase/auth/service.go` | Login, RefreshToken, PasswordReset business logic |
| `usecase/auth/email_sender.go` | `EmailSender` interface (auth-specific subset) |
| `usecase/auth/service_test.go` | Unit tests (4 cases including email-send failure regression) |
| `interface/email/auth_email_service.go` | `AuthEmailService` — sends password reset email |
| `interface/email/smtp_sender.go` | `SMTPSender` — shared SMTP client wrapper |
| `interface/email/ppid_email_service.go` | `PPIDEmailService` — also uses `SMTPSender` |
| `domain/auth/types.go` | `TokenPair`, `AccessToken`, `ResetToken` |
| `domain/user/user.go` | `User` entity + `Validate()` |
| `pkg/password/bcrypt.go` | bcrypt wrapper (cost 12) |
| `pkg/jwt/jwt.go` | JWT HS256 generation/validation |

## Service

`usecase/auth/service.go` — `Service{repo, userRepo, email, clock, jwtSecret, resetTokenTTL, resetURLBase, rateLimit{5, 1h}}`

| Method | Purpose |
|---|---|
| `Login(ctx, email, password)` | Returns `*TokenPair` (access + refresh tokens). Always returns `"Invalid credentials"` on failure. |
| `RefreshToken(ctx, refreshToken)` | Validates refresh JWT, issues new access token (24h). |
| `RequestPasswordReset(ctx, email, ip, ua)` | Generates 32-byte hex token. **Sends reset email via `email.SendPasswordResetEmail`**. Surfaces errors to operator (was previously swallowed — see `recom.docs/bugs.md#1`). |
| `CheckResetToken(ctx, token)` | Returns `*ResetTokenInfo{Valid, ExpiresAt}`. Never errors. |
| `ResetPassword(ctx, token, newPassword)` | bcrypt hashes, persists, marks tokens used. |

### Password Reset Flow (fixed)

```
1. POST /auth/password-reset/request → rate-limited (5/hour per email)
2. Token persisted to password_reset_tokens table
3. Email sent with reset link: {resetURLBase}/auth/password-reset/confirm?token=...
4. User clicks link → POST /auth/password-reset/confirm {token, new_password}
5. Token marked used; password updated
```

If email send fails (SMTP down, etc.), the handler returns **500** instead of the previous silent 200. Operators see the actual SMTP error in logs.

## Endpoints

All endpoints are **public** (no auth required) and rate-limited via `AuthRateLimitMiddleware` (5 req/min).

| Method | Path | Body | Response |
|---|---|---|---|
| POST | `/api/v1/auth/login` | `{email, password}` | `{access_token, refresh_token, expires_at}` |
| POST | `/api/v1/auth/refresh` | `{refresh_token}` | `{access_token, expires_at}` |
| POST | `/api/v1/auth/password-reset/request` | `{email}` | 200 `{message}` (or 429 if rate-limited) |
| GET | `/api/v1/auth/password-reset/check?token=...` | — | `{valid, expires_at?}` |
| POST | `/api/v1/auth/password-reset/confirm` | `{token, new_password}` | 200 |

## Rate Limiting

- HTTP: `AuthRateLimitMiddleware` (5 req/min, per IP)
- Business: 5 reset requests per hour per email

## Tests

**Unit** (`usecase/auth/service_test.go`, 4 cases):
- `TestRequestPasswordReset_EmailSentOnSuccess` — link includes token + base URL
- `TestRequestPasswordReset_EmailFailureSurfacesError` — regression for `bugs.md#1`
- `TestRequestPasswordReset_NoEmailSender_LegacyBehavior` — nil sender = silent success
- `TestRequestPasswordReset_RateLimit`

**Integration** (`integration/auth_test.go`): 13 cases covering login, refresh, password reset happy paths, invalid credentials, rate limits.

## Configuration

`config/smtp.go` — `IsConfigured()` returns true only if host/port/username/password are all non-empty.

## Related

- `pkg/jwt/jwt.go:34-41` — Token generation (24h access / 7d refresh)
- `pkg/password/bcrypt.go` — bcrypt wrapper (cost 12)
- `interface/http/middleware/auth.go` — JWT bearer parsing
- `interface/http/middleware/ratelimit.go` — Token-bucket rate limiter
- `cmd/seed.go` — creates `admin@desa.local` / `operator@desa.local` defaults