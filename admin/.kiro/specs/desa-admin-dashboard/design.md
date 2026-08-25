# Technical Design Document: Desa Admin Dashboard

## Overview

A Vue.js 3 SPA providing an administrative interface for the Village (Desa) Information Management System. Communicates with a REST API at base URL `/api/v1` (see `swagger.yaml` for full endpoint reference). Authentication uses JWT Bearer tokens with automatic refresh.

**Tech stack**: Vue 3 + Composition API, Vite, Tailwind CSS, Pinia, Vue Router 4, Axios, VeeValidate/Yup, Leaflet (map), TipTap (rich text).

## Architecture

```
Presentation  →  Composables / Pinia Stores  →  Service Layer (Axios)  →  REST API
(Views/Components)   (business logic, state)     (api.js + *service.js)
```

### Directory Structure

```
src/
├── assets/styles/main.css
├── components/
│   ├── common/        # AppButton, AppInput, AppModal, AppTable, AppPagination,
│   │                  # AppNotification, LoadingSpinner, ConfirmDialog
│   ├── layout/        # AppLayout, AppSidebar, AppHeader, AppBreadcrumb
│   └── forms/         # ImageUpload, FileUpload, RichTextEditor, MapPicker
├── views/
│   ├── auth/          # LoginView, ForgotPasswordView, ResetPasswordView
│   ├── dashboard/     # DashboardView
│   ├── users/         # UserListView, UserFormView, RoleListView
│   ├── content/       # Banner*, News*, Umkm*, Facility*, Ppid* (List + Form each)
│   ├── organization/  # StructureView, ProfileView
│   └── NotFoundView
├── composables/       # useAuth, useApi, useNotification, usePagination,
│                      # useConfirm, useImagePreview
├── stores/            # auth.js, notification.js, ui.js
├── services/          # api.js, auth/user/role/banner/news/umkm/
│                      # facility/ppid/structure/profile .service.js
├── router/index.js
└── utils/             # validators, formatters, constants, storage
```

## API Reference

Base URL: `/api/v1` — servers: `localhost:8080`, staging, production. See `swagger.yaml` for full spec.

All list responses follow:
```
{ success, data: { [resource]: [], pagination: { page, limit, total, total_pages } } }
```

Dashboard stats have no dedicated endpoint — derive counts from `pagination.total` of each list endpoint.

## Authentication Flow

- `POST /auth/login` — body: `{ email, password }` → `{ access_token, refresh_token, expires_at }`
- `POST /auth/refresh` — body: `{ refresh_token }` → `{ access_token, expires_at }`
- No logout endpoint. Clear tokens client-side on logout.
- Password reset: `POST /auth/password-reset/request` → `GET /auth/password-reset/check?token=` → `POST /auth/password-reset/confirm` with `{ token, new_password }`

Axios request interceptor attaches `Authorization: Bearer <access_token>`. Response interceptor retries on 401 after refreshing; redirects to `/login` if refresh fails.

## Data Models

All IDs are UUIDs. Timestamps are ISO 8601 strings. `?` = nullable/optional.

```typescript
// Auth
LoginRequest:          { email: string, password: string }
LoginResponse:         { access_token, refresh_token, expires_at }
RefreshTokenResponse:  { access_token, expires_at }

// User
User:                  { id, name, email, profile_image_url?: string|null, roles?: string[], created_at, updated_at }
CreateUserRequest:     { name, email, password }          // JSON
UpdateUserRequest:     { name, email }                    // JSON
UpdatePasswordRequest: { old_password?, new_password }    // PUT /users/{id}/password
// Roles assigned via: POST /users/{id}/roles { role: string }
//                     DELETE /users/{id}/roles/{role}

// Role & Permission
Role:                  { name: string }                   // no id, no permissions array
Permission:            { resource: string, action: string }  // e.g. "users", "read"
// Permissions are per-role: GET/POST/DELETE /roles/{role}/permissions
// Known permissions: users:read/write, banners:read/write, berita:write,
//   umkm:write, fasilitas:write, ppid:write, struktur:write, desa:write, roles:read/write

// Banner
Banner:                { id, title, image_url, status: 'active'|'inactive', created_at, updated_at }
// Create/Update: multipart/form-data { title, image (binary, max 10MB) }
// Status toggle: PUT /banners/{id}/status { status }
// List filter: GET /banners?status=active|inactive

// Berita (News)
Berita:                { id, title, content, image_url?: string|null, created_at, updated_at }
// Create: multipart/form-data { title, content, image? (max 10MB) }
// Search: GET /berita?q=&since=YYYY-MM-DD&until=YYYY-MM-DD&page=&limit=

// UMKM
UMKM:                  { id, name, owner?, address?, phone?, email?, website?, created_at, updated_at }
// Create/Update: JSON { name, owner?, address?, phone?, email?, website? }
// Search: GET /umkm?q=&page=&limit=

// Fasilitas
Fasilitas:             { id, name, type?, latitude: number, longitude: number, description?, created_at, updated_at }
// Create/Update: JSON { name, type?, latitude, longitude, description? }
// Filter: GET /fasilitas?minLat=&maxLat=&minLon=&maxLon=&page=&limit= (bounding box, no text search)

// PPID
PPID:                  { id, title, category?, document_url?, description?, created_at, updated_at }
// Create: multipart/form-data { title, category?, description?, document (binary, max 50MB) }
// Update: multipart/form-data { title, category?, description?, document? }
// Public request: POST /ppid/{id}/requests { requester_name, requester_email, purpose? }
// Admin view requests: GET /ppid/requests
// Search: GET /ppid?category=&q=&page=&limit=

// Struktur (flat list, not hierarchical)
Struktur:              { id, name, position?, email?, phone?, profile_image_url?, description?, created_at, updated_at }
// Create/Update: multipart/form-data { name, position?, email?, phone?, description?, image? (max 10MB) }
// Search: GET /struktur?q=&page=&limit=

// Desa (Village Profile)
Desa:                  { id, name, description?, address?, phone?, email?, website?, vision_mission?, created_at, updated_at }
// Update: JSON { name, description?, address?, phone?, email?, website?, vision_mission? }
```

## Service Layer

Each service module wraps Axios calls. Key signatures (full details in swagger.yaml):

- `authService`: `login(email, password)`, `refresh(refreshToken)`, `requestPasswordReset(email)`, `checkResetToken(token)`, `confirmPasswordReset(token, newPassword)`
- `userService`: `list(page, limit)`, `get(id)`, `create(data)`, `update(id, data)`, `delete(id)`, `updatePassword(id, data)`, `assignRole(id, role)`, `removeRole(id, role)`
- `roleService`: `list()`, `create(name)`, `delete(role)`, `getPermissions(role)`, `addPermission(role, resource, action)`, `removePermission(role, resource, action)`
- `bannerService`: `list(page, limit, status?)`, `get(id)`, `create(formData)`, `update(id, formData)`, `delete(id)`, `updateStatus(id, status)`
- `newsService`: `list(page, limit, q?, since?, until?)`, `get(id)`, `create(formData)`, `update(id, formData)`, `delete(id)`
- `umkmService`: `list(page, limit, q?)`, `get(id)`, `create(data)`, `update(id, data)`, `delete(id)`
- `facilityService`: `list(page, limit, bbox?)`, `get(id)`, `create(data)`, `update(id, data)`, `delete(id)`
- `ppidService`: `list(page, limit, category?, q?)`, `get(id)`, `create(formData)`, `update(id, formData)`, `delete(id)`, `listRequests(page, limit)`
- `structureService`: `list(page, limit, q?)`, `get(id)`, `create(formData)`, `update(id, formData)`, `delete(id)`
- `desaService`: `get()`, `update(data)`

## State Management

**Auth store** (`stores/auth.js`): `accessToken`, `refreshToken`, `expiresAt`. No user object from login — fetch `/users/me` separately for current user info (name, roles).

**Notification store**: toast queue with type/message/duration.

**UI store**: sidebar collapsed state, confirm dialog state.

Local component state handles form data, pagination, search/filter values.

## Components

### Common
- `AppTable` — columns config + data array, loading skeleton, empty state, slot for custom cell rendering
- `AppPagination` — driven by `{ page, total_pages }` from API response
- `AppModal` — backdrop + ESC close, focus trap
- `ConfirmDialog` — used before all delete operations
- `AppNotification` — success auto-dismisses (3s), errors persist until dismissed

### Forms
- `ImageUpload` — drag-drop, preview, validates MIME + size (warn >5MB, max 10MB)
- `FileUpload` — for PPID documents, max 50MB, accepts PDF/DOC/DOCX/XLS/XLSX
- `RichTextEditor` — TipTap, used for Berita content
- `MapPicker` — Leaflet map, click-to-set coordinates, syncs with lat/lng inputs

### Views
- `LoginView` — email + password form
- `ForgotPasswordView` — email form → request reset
- `ResetPasswordView` — reads `?token=` from query, validates token via check endpoint, submits new password
- `DashboardView` — stat cards; counts fetched from `pagination.total` of users/berita/umkm/fasilitas list calls
- `UserListView` / `UserFormView` — list with search; form handles create (name/email/password) and update (name/email) separately; password change via separate form calling `PUT /users/{id}/password`
- `RoleListView` — list roles, manage permissions per role (add/remove `{ resource, action }` pairs), assign/remove roles to users
- `BannerListView` / `BannerFormView` — status filter, inline status toggle via `PUT /banners/{id}/status`
- `NewsListView` / `NewsFormView` — date range filter (since/until), rich text editor for content
- `UmkmListView` / `UmkmFormView` — text search, JSON body (no file upload)
- `FacilityListView` / `FacilityFormView` — map view with Leaflet, bounding box filter on list
- `PpidListView` / `PpidFormView` — category filter, file upload; admin can view public requests
- `StructureView` — flat list (not tree), multipart form with optional photo
- `ProfileView` — single Desa record, JSON update

## Routing & Guards

All routes under `/` require authentication (checked via auth store). Unauthenticated requests redirect to `/login`. Auth pages (`/login`, `/forgot-password`, `/reset-password`) are public. Vue Router navigation guard reads `accessToken` from auth store.

## Error Handling

- 401 on any request → attempt token refresh → retry → redirect to `/login` on failure
- 403 → show error notification ("Insufficient permissions")
- 4xx validation errors → display API `error` message near relevant field
- 5xx → generic error notification
- Network timeout (30s Axios default) → error notification
