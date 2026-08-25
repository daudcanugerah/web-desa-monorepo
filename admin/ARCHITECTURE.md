# Architecture

## Tech Stack

| Layer | Technology |
|-------|------------|
| Framework | Vue.js 3.4 (Composition API with `<script setup>`) |
| Build Tool | Vite 5 |
| State Management | Pinia 2.1 |
| Routing | Vue Router 4.3 |
| HTTP Client | Axios |
| Styling | Tailwind CSS 3.4 |
| Form Validation | VeeValidate 4.12 + Yup 1.3 |
| i18n | Vue-i18n 9.14 |
| Rich Text | Quill 2.0 + quill-better-table-plus |
| Maps | Leaflet 1.9.4 |

## Directory Structure

```
src/
├── main.js                 # App initialization
├── App.vue                 # Root component
├── assets/styles/main.css  # Global styles
├── components/
│   ├── common/             # UI components (AppButton, AppModal, etc.)
│   ├── forms/             # Form components (FileUpload, MapPicker, etc.)
│   └── layout/            # Layout (AppLayout, AppSidebar, AppHeader)
├── composables/           # Vue composables (useConfirm, useLocale, etc.)
├── i18n/
│   ├── index.js           # i18n config
│   └── locales/          # Translation files (id.json, en.json)
├── router/index.js        # Routes with guards
├── services/             # API modules (13 service files)
├── stores/               # Pinia stores (auth, notification, ui)
├── utils/                # Helpers (storage, imageUrl, validators)
└── views/                # Page components by feature
```

## State Management

### Stores

| Store | Responsibility |
|-------|----------------|
| `auth.js` | Access/refresh tokens, current user, auth state |
| `notification.js` | Toast notification queue |
| `ui.js` | UI state (sidebar, etc.) |

## Data Flow

```
User Action
    ↓
View Component
    ↓
Service Layer (api.js + domain services)
    ↓
Axios Instance (interceptors for auth, errors)
    ↓
REST API (swagger.yaml spec)
```

## API Layer

- Axios instance with Bearer token authentication
- Request interceptor: attaches JWT token from storage
- Response interceptor: handles 401 by refreshing token; 429 is not handled globally (each view branches on `err.response?.status === 429` to show a friendly rate-limit message)
- Base URL from `VITE_API_BASE_URL` env variable
- Backend spec: `swagger.yaml` (Swagger 2.0 — `basePath: /api/v1`, `host: localhost:8080`, `schemes: [http]`). RBAC enforced server-side via Casbin; every admin operation declares `RBAC: <resource>:<action>` in its description
- JWT lifetimes: access 24 h, refresh 7 d
- Rate limiting: 5 password-reset requests / hour / email (HTTP 429); general rate limiting on most write endpoints

## Route Guards

| Guard | Logic |
|-------|-------|
| `requiresAuth` | Redirect to `/login` if not authenticated |
| `requiresAdmin` | Check user has `admin` role, redirect to `/` if not |

## Form Validation

Hand-rolled in each view (`utils/validators.js` provides shared helpers — type/size validation, file-size formatting). VeeValidate + Yup are declared in `package.json` but no view imports them.

## Authentication Flow

1. User submits credentials
2. API returns access_token + refresh_token + expires_at
3. Tokens stored in localStorage via `utils/storage.js`
4. Axios interceptor attaches token to requests
5. On 401, interceptor attempts token refresh via `POST /auth/refresh`; retries the original request once
6. On second 401, clears tokens and routes to `/login`
7. Auth state persisted in Pinia `auth` store

## Coverage & Swagger Compliance

See `feature-docs/SWAGGER_DIFF.md` for the live diff between `swagger.yaml` and the frontend implementation. Headline: 88 admin endpoints defined, 78 consumed (88.6 %); 0 divergences; 10 informational gaps (category CRUD for resources without a category field, plus 3 helper endpoints).