# 00 — Architecture & Cross-Cutting Concerns

The shared plumbing every feature relies on: tech stack, layering, auth/session lifecycle, axios interceptors, route guards, service layer pattern, error handling, i18n, theme, and file conventions.

---

## 1. Tech Stack

| Layer | Library | Version |
|-------|---------|---------|
| Framework | Vue.js | ^3.4.0 |
| Build | Vite | ^8.0.13 (`package.json`; `ARCHITECTURE.md` still says v5) |
| Devtools | `vite-plugin-vue-devtools` | ^8.1.2 |
| State | Pinia | ^2.1.0 |
| Router | Vue Router | ^4.3.0 |
| HTTP | Axios | ^1.6.0 |
| Styling | Tailwind CSS | ^3.4.0 |
| Validation | VeeValidate + Yup | ^4.12 / ^1.3 (declared, **not used in views**) |
| i18n | Vue-i18n | ^9.14.5 |
| Rich text | Quill + quill-better-table | ^2.0 / ^0.1.6 |
| Maps | Leaflet + leaflet-minimap | ^1.9.4 / ^3.6.1 |
| Linting | ESLint 10 + `eslint-plugin-vue` | ^10.4.0 |

Vite path alias `@` → `./src` is declared but unused — all imports use relative paths.

---

## 2. Directory Layout

```
src/
├── main.js                     # Bootstrap: Pinia + router + i18n + theme.init()
├── App.vue                     # Root: toast + confirm + <router-view>
├── assets/
│   ├── maps/area-desa-poly.json  # Village polygon for MapPicker
│   └── styles/main.css
├── components/
│   ├── common/                 # AppBadge, AppButton, AppModal, AppTable, ...
│   ├── forms/                  # FileUpload, ImageUpload, MapPicker, RichTextEditor
│   └── layout/                 # AppHeader, AppSidebar, AppBreadcrumb, AppLayout
├── composables/                # useConfirm, useImagePreview, useLocale, useMetabase
├── i18n/{index.js,locales/{id,en}.json}
├── router/index.js             # Routes + beforeEach guards
├── services/                   # api.js + one file per domain
├── stores/                     # auth, notification, theme, ui
├── utils/                      # imageUrl, storage, validators
└── views/                      # auth/, content/, dashboard/, organization/, users/
```

Conventions:
- **One service per domain** under `src/services/*.service.js`.
- **List + Form pattern** for content resources (`FooListView` + `FooFormView`).
- **Composables** wrap stateful logic that crosses multiple views (confirm, image preview, locale, metabase).
- **Stores** are thin: actions are the only writers, state is plain refs.

---

## 3. Layering

```
View  →  Service  →  api (axios)  →  REST backend
   ↘        ↘
    Store     Composables
```

- **Views** handle rendering, form state, and user interactions.
- **Services** are plain objects of arrow functions wrapping a single HTTP call.
- **`api`** is the single Axios instance with shared interceptors.
- **Stores** hold cross-view state (auth, toasts, theme, UI flags).
- **Composables** wrap reusable logic without owning global state.

---

## 4. Bootstrap Order (`src/main.js`)

1. Create app.
2. Register Pinia.
3. Register router.
4. Register i18n.
5. Call `themeStore.init()` to apply persisted dark mode before mount.
6. Mount to `#app`.

`App.vue` renders:

```vue
<AppNotification />
<ConfirmDialog />
<router-view />
```

The notification + confirm dialogs live at the app root so they survive route changes.

---

## 5. Authentication Flow

See `feature-docs/01-authentication.md` for the full user-facing flow. This section documents the plumbing only.

### 5.1 Token Storage (`utils/storage.js`)

| Key | Type | Purpose |
|-----|------|---------|
| `access_token` | string | JWT for `Authorization: Bearer` |
| `refresh_token` | string | Sent to `POST /auth/refresh` |
| `expires_at` | number | Unix seconds (informational) |

### 5.2 Auth Store (`stores/auth.js`)

State:
```js
{ accessToken, refreshToken, expiresAt, currentUser }
```

Actions:
- `setTokens({ access_token, refresh_token, expires_at })` — writes Pinia **and** `localStorage`.
- `clearTokens()` — clears both.
- `fetchCurrentUser()` — lazy-imports `services/api` and calls `GET /users/me`.

### 5.3 Request Interceptor (`services/api.js`)

```js
api.interceptors.request.use((config) => {
  const token = getAccessToken()
  if (token) config.headers.Authorization = `Bearer ${token}`
  return config
})
```

### 5.4 Response Interceptor — 401 Refresh

1. If `error.response.status === 401` and request is not already retried (`_retry` flag), set `originalRequest._retry = true`.
2. Call `POST ${VITE_API_BASE_URL}/auth/refresh` using **raw axios** (not the `api` instance — to avoid re-entering the interceptor).
3. Persist new tokens via `setTokens(data)`; also dynamically `import('../stores/auth')` and call `setTokens` on the Pinia store to keep it in sync (wrapped in try/catch because Pinia may not yet be installed).
4. Patch `originalRequest.headers.Authorization` with the new token and re-issue the request via `api(originalRequest)`.
5. If refresh fails: `clearTokens()`, sync the store, and `router.push('/login')`.

### 5.5 Route Guards (`router/index.js`)

```js
router.beforeEach(async (to) => {
  const auth = useAuthStore()

  if (to.meta.requiresAuth && !auth.accessToken) return '/login'
  if (to.name === 'Login' && auth.accessToken) return '/'

  if (to.meta.requiresAdmin) {
    if (!auth.currentUser) await auth.fetchCurrentUser()
    if (!auth.currentUser?.roles?.includes('admin')) return '/'
  }
})
```

| Meta | Applied to |
|------|-----------|
| `requiresAuth` | All `/banners`, `/users`, `/berita`, `/umkm`, `/fasilitas`, `/ppid`, `/struktur`, `/profile`, `/roles`, `/infographic`, `/me`, `/` |
| `requiresAdmin` | `/users`, `/users/create`, `/users/:id/edit` |

### 5.6 Layout-Level Bootstrap

`AppLayout.onMounted` calls `authStore.fetchCurrentUser()` so the header avatar, name, and role badges are populated for every authenticated route.

---

## 6. Service Layer Pattern

Every domain service is a plain object of arrow functions:

```js
// services/foo.service.js
import api from './api'

export const fooService = {
  list:  (params)        => api.get('/foo', { params }),
  get:   (id)            => api.get(`/foo/${id}`),
  create: (payload)      => api.post('/foo', payload),
  update: (id, payload)  => api.put(`/foo/${id}`, payload),
  remove: (id)           => api.delete(`/foo/${id}`),
}
```

### Conventions

- Services return the raw `AxiosResponse`. Views destructure `res.data.data`.
- For multipart uploads the service builds `FormData` and (mostly) sets `Content-Type: multipart/form-data` explicitly. `facility.service` and `umkm.service` rely on Axios to set the boundary automatically — small inconsistency.
- Pagination envelope: `{ data: { <resource>: [...], pagination: { page, limit, total, total_pages } } }`.

---

## 7. Error Handling

- **Global**: 401 refresh is the only global concern (handled in `api.js`).
- **Per-view**: each view catches errors and calls `notificationStore.error(err.response?.data?.error || fallbackMessage)`.
- **Toast semantics**: success toasts auto-dismiss after 3 s; error toasts are sticky (`duration: 0`).

---

## 8. Validation

- Validation is **hand-rolled** in every view (email regex, required checks, length checks, lat/lng range).
- `utils/validators.js` exports:
  - `isValidImageType(file)` — JPEG / PNG / WebP only.
  - `isValidDocumentType(file)` — PDF / DOC / DOCX / XLS / XLSX.
  - `isWithinSize(file, maxMB)` — defaults to 50 MB.
  - `formatFileSize(bytes)` — pretty-print.
- VeeValidate + Yup are installed but unused.

---

## 9. Image / File Handling

- **`utils/imageUrl.js`** → `getImageUrl(nameOrUrl)`:
  - If `nameOrUrl` starts with `http://` or `https://`, return as-is.
  - Otherwise return `${VITE_API_BASE_URL}/files/${name}`.
- **`ImageUpload.vue`** + `useImagePreview` composable → drag-and-drop with type/size validation + `URL.createObjectURL` preview.
- **`FileUpload.vue`** → drag-and-drop for PDF/DOC/XLS, default 50 MB.
- **`RichTextEditor.vue`** → Quill image/video handlers call `newsService.uploadMedia` and inline the returned URL.
- **`UmkmFormView`** has a quirk: on update it re-fetches existing images via `fetch(getImageUrl(img)).then(r => r.blob())` to re-upload them — fragile if `/files/...` is not CORS-enabled.

---

## 10. Pagination

- `AppPagination.vue` shows up to 5 page numbers + prev/next.
- Most views use `pagination.total_pages` from the backend.
- `ProfileSectionListView` and `InfographicListView` compute `Math.ceil(pagination.total / pagination.limit)` themselves — minor inconsistency.

---

## 11. i18n

- `i18n/index.js` uses `createI18n({ legacy: false })` (Composition mode).
- Locales: `id` (default + fallback), `en`.
- Persistence: `localStorage.locale`; also updates `document.documentElement.lang`.
- `useLocale()` composable exposes a writable `currentLocale`.
- Only `RoleListView` is fully i18n'd; most views hardcode Indonesian strings.

---

## 12. Theme (Dark Mode)

- `stores/theme.js` toggles `dark` (persisted in `localStorage.darkMode`).
- Applies `<html class="dark">` via `applyTheme()`.
- `themeStore.init()` runs in `main.js` **before** mount so there is no flash.
- Toggled from `ThemeToggle.vue` in the header.

---

## 13. Notification & Confirm

- **`stores/notification.js`** holds `toasts: []`. `add({ type, message, duration })` pushes a toast; `remove(id)` removes it. Convenience: `success(msg)` / `error(msg)`.
- **`stores/ui.js`** holds `confirmDialog: { open, title, message, ... }`. `showConfirm({...})` opens it.
- **`components/common/ConfirmDialog.vue`** renders `AppModal` driven by the `ui` store.
- **`composables/useConfirm.js`** wraps `uiStore.showConfirm` in a `Promise<boolean>` so views can `await confirm(...)`.

---

## 14. Known Inconsistencies / TODOs

1. **Vite version mismatch** — `package.json` says ^8.0.13, `ARCHITECTURE.md` says v5.
2. **`@` alias** configured in `vite.config.js` but never used.
3. **VeeValidate + Yup** unused.
4. **PPID download path** — service calls `/ppid/documents/{id}/download` (plural); swagger defines `/ppid/document/{id}/download` (singular).
5. **Pagination source** — see §10.
6. **i18n coverage** — only RoleListView is i18n'd.
7. **`theme.js`** has stray `console.log` calls (debug remnants).
8. **`LoginView`** does not `await authStore.fetchCurrentUser()` — relies on `AppLayout.onMounted` to populate the user.
9. **`UmkmFormView`** re-fetches images as blobs to re-upload — fragile under CORS restrictions.
10. **No global error toast** — every view handles its own.

---

## 15. Environment Variables

`.env.example`:

```
VITE_API_BASE_URL=http://localhost:8080/api/v1
VITE_METABASE_URL=http://localhost:3000
```

Both are read via `import.meta.env.*` and used by `services/api.js` and `composables/useMetabase.js`.