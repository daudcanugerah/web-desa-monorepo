# 01 — Authentication

Public-facing authentication flows: login, forgot password, reset password. JWT lifecycle, session restoration, and token refresh are also handled here (alongside `00-architecture.md` §5).

---

## 1. Purpose

Allow village admins to:
- Log in with email + password.
- Recover their password via email.
- Reset their password via a one-time token link.
- Stay logged in across page reloads (JWT persisted in `localStorage`).
- Automatically refresh access tokens in the background.

---

## 2. Routes

| Path | Name | Breadcrumb | Guard |
|------|------|-----------|-------|
| `/login` | `Login` | — | none (redirects to `/` if already authed) |
| `/forgot-password` | `ForgotPassword` | — | none |
| `/reset-password` | `ResetPassword` | — | none (validates `?token=` on mount) |

Catch-all: `/:pathMatch(.*)*` → `NotFoundView`.

---

## 3. Files

| File | Role |
|------|------|
| `src/views/auth/LoginView.vue` | Email + password form, calls `authService.login` |
| `src/views/auth/ForgotPasswordView.vue` | Email form, calls `authService.requestPasswordReset` |
| `src/views/auth/ResetPasswordView.vue` | New-password form, validates token then calls `confirmPasswordReset` |
| `src/services/auth.service.js` | 5 endpoint wrappers |
| `src/stores/auth.js` | Token state + `fetchCurrentUser` |
| `src/services/api.js` | Request + 401 refresh interceptors |
| `src/utils/storage.js` | `getAccessToken`, `setTokens`, `clearTokens` |

---

## 4. API Endpoints

| Method | Path | Body / Params | Used by |
|--------|------|---------------|---------|
| `POST` | `/auth/login` | `{ email, password }` | `LoginView` |
| `POST` | `/auth/refresh` | `{ refresh_token }` | `api.js` interceptor (401 path) |
| `POST` | `/auth/password-reset/request` | `{ email }` | `ForgotPasswordView` |
| `GET`  | `/auth/password-reset/check?token=...` | query string | `ResetPasswordView` (`onMounted`) |
| `POST` | `/auth/password-reset/confirm` | `{ token, new_password }` | `ResetPasswordView` |

### Login response shape

```json
{
  "success": true,
  "data": {
    "access_token": "...",
    "refresh_token": "...",
    "expires_at": 1717500000
  }
}
```

`authStore.setTokens(data)` writes both Pinia state and `localStorage` keys `access_token` / `refresh_token` / `expires_at`.

---

## 5. Behaviors

### 5.1 Login (`LoginView.vue`)

- Form fields: `email`, `password`.
- Validation: email regex + password length ≥ 8 (manual).
- On submit:
  1. Calls `authService.login({ email, password })`.
  2. On success → `authStore.setTokens(response.data.data)`.
  3. Navigates to `/` (or to `?redirect` query param if present).
- **Note**: the view does **not** `await authStore.fetchCurrentUser()` — the avatar / role are populated lazily by `AppLayout.onMounted`.
- On 401 → `notification.error(...)`.

### 5.2 Forgot Password (`ForgotPasswordView.vue`)

- Single field: `email`.
- On submit → `authService.requestPasswordReset({ email })`.
- Shows success toast + "check your email" message; user is expected to follow the link.

### 5.3 Reset Password (`ResetPasswordView.vue`)

- Reads `?token=` from `route.query`.
- On mount → `authService.checkResetToken(token)` to verify validity. If invalid, blocks the form.
- Form fields: `password`, `password_confirmation`.
- Validation: both ≥ 8 chars and must match.
- On submit → `authService.confirmPasswordReset({ token, new_password })` → success toast + redirect to `/login`.

### 5.4 Session Lifecycle

1. **Login** writes tokens to Pinia + `localStorage`.
2. **App reload** → `auth` store re-hydrates from `localStorage` on first access (`accessToken` getter).
3. **Every request** → request interceptor attaches `Authorization: Bearer <accessToken>`.
4. **Access token expired (401)** → response interceptor refreshes via `POST /auth/refresh`, retries the original request once.
5. **Refresh failed** → `clearTokens()` + `router.push('/login')`.

---

## 6. UI Notes

- All three auth views use a centered card layout (no `AppLayout`).
- Theme is fully respected — the same dark mode applies to login screens.
- Locale switcher is available in the header even on auth screens (no header on auth views currently — locale toggle not exposed there; would need a layout change).

---

## 7. Notes & Gotchas

- **`LoginView` does not pre-fetch the user.** If a feature relies on `auth.currentUser` being set immediately after login, it may need an explicit `await authStore.fetchCurrentUser()`.
- **Logout** is fired from `AppHeader` avatar dropdown (`authStore.clearTokens()` + `router.push('/login')`). There is no dedicated logout endpoint in `auth.service.js`.
- **Refresh token storage** is plain `localStorage`. A malicious script with access to the page can exfiltrate it.
- **Token revocation** is not surfaced in the UI — there is no "log out all devices" feature.