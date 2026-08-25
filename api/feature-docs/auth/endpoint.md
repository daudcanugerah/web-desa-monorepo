# Auth — Endpoints

Five public endpoints. All under `/api/v1/auth/*`. No JWT required, no RBAC. Rate-limited via `AuthRateLimitMiddleware` (5 req/min) which sits inside `PublicRateLimitMiddleware` (100 req/min).

**Module:** webdesa/api · **Last updated:** 2026-08-23

## Routes

| Method | Path | Scope | RBAC | Summary |
|---|---|---|---|---|
| POST | `/api/v1/auth/login` | public | — | Email + password → access/refresh JWT pair |
| POST | `/api/v1/auth/refresh` | public | — | Valid refresh token → new access token |
| POST | `/api/v1/auth/password-reset/request` | public | — | Send reset email if email exists |
| GET | `/api/v1/auth/password-reset/check` | public | — | Validate a token before user opens the form |
| POST | `/api/v1/auth/password-reset/confirm` | public | — | Set new password using a valid token |

## Request/Response

### POST `/api/v1/auth/login`

Request:
```json
{ "email": "string", "password": "string (min 6)" }
```

Success 200:
```json
{ "success": true, "data": { "access_token": "jwt", "refresh_token": "jwt", "expires_at": "RFC3339" } }
```

Errors:
- `400` — missing/short password (handler validation, not auth) / malformed JSON
- `401` — invalid credentials (unknown email OR wrong password — same string either way)
- `429` — rate-limited (auth bucket or public bucket)

### POST `/api/v1/auth/refresh`

Request:
```json
{ "refresh_token": "jwt" }
```

Success 200:
```json
{ "success": true, "data": { "access_token": "jwt", "expires_at": "RFC3339" } }
```

Errors:
- `400` — missing field
- `401` — invalid/expired/malformed refresh token
- `429` — rate-limited

### POST `/api/v1/auth/password-reset/request`

Request:
```json
{ "email": "string" }
```

Success 200 (always returns 200 if request is well-formed):
```json
{ "success": true, "data": { "message": "..." } }
```

Errors:
- `400` — missing/malformed email JSON
- `429` — rate-limited (auth bucket AND/OR business 5/hour per email)
- `500` — **non-rate-limit failures** (e.g. SMTP send failure, DB error) — `auth.go:211`

### GET `/api/v1/auth/password-reset/check?token=...`

Query: `token` (string, required).

Success 200 (never errors on bad token — returns `valid:false`):
```json
{ "success": true, "data": { "valid": true, "expires_at": "RFC3339" } }
```

or for invalid/expired/used:
```json
{ "success": true, "data": { "valid": false } }
```

Errors:
- `400` — missing `token` query param
- `429` — rate-limited

### POST `/api/v1/auth/password-reset/confirm`

Request:
```json
{ "token": "string", "new_password": "string (min 6)" }
```

Success 200:
```json
{ "success": true, "data": { "message": "..." } }
```

Errors:
- `400` — missing fields / short password
- `401` — invalid/expired/already-used token
- `429` — rate-limited
- `500` — DB or bcrypt failure

## Validation

Handler-level (`httpin` tags):
- `LoginRequest.email` — `required`
- `LoginRequest.password` — `required,min=6`
- `RefreshRequest.refresh_token` — `required`
- `RequestResetRequest.email` — `required`
- `ConfirmResetRequest.token` — `required`
- `ConfirmResetRequest.new_password` — `required,min=6`

Domain rules:
- Login `password < 6` → 400 (handler, not auth).
- Login failure → generic `"Invalid credentials"` regardless of cause.
- Reset token uniqueness enforced at DB layer (`UNIQUE` on `token`).
- Reset-token TTL = `systemConfig.JWT.GetExpiration()` (default 24h).

## RBAC

None. All auth endpoints are public — JWT is not required, and Casbin policies do not gate them.

Rate-limit chain is the only access control:

| Layer | Limit | Middleware | Key |
|---|---|---|---|
| Public | 100 req/min | `PublicRateLimitMiddleware` | `r.RemoteAddr` |
| Auth | 5 req/min | `AuthRateLimitMiddleware` | `r.RemoteAddr` |
| Business reset | 5/hour per email | inline in service | `email` |

## Known Issues

- **Email body hardcodes "1 hour"** (`interface/email/auth_email_service.go:102`): actual reset-token TTL is 24h (driven by `systemConfig.JWT.GetExpiration()`). Users are misinformed about expiry. **Design issue** → see [concept.md](./concept.md#business-rules).
- **Rate limit key includes port** (`ratelimit.go:125` uses `r.RemoteAddr` which is `IP:port`): bucket is per TCP connection, not per IP. Documented honestly above. **Design issue** → see [concept.md](./concept.md#rate-limit-key-behavior).
- **`RequestPasswordReset` returns 500 on non-rate-limit errors** (`auth.go:211`): correct behavior (was previously silent 200), but documented because it surfaces SMTP/auth failures to the operator.
- **Dead code** — `isValidEmail` at `auth.go:311` is unused. `interface/postgres/auth.go:34` (`FindUserByEmail`) is also dead — `AuthRepository` is wired only via the `Repository` interface for reset tokens, not the `UserRepository` interface. **Code-level** → tracked in code, no DB or design impact.
- **`GET /auth/password-reset/check` never errors on bad token**: returns `valid:false` instead. UI consumers must check the `valid` flag, not the HTTP status.

## Related

- [concept.md](./concept.md)
- [database.md](./database.md)
- [users/endpoint.md](../users/endpoint.md) — `PUT /users/{id}/password` (admin-bypass variant)
