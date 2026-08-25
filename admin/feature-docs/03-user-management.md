# 03 — User Management

Admin-only CRUD for backend users + role assignment + password reset.

---

## 1. Purpose

Allow `admin` users to:
- List, search, and paginate users.
- Create users and assign one or more roles.
- Edit user details and adjust their roles.
- Reset a user's password without knowing the old one.
- Delete users (cannot delete self).

---

## 2. Routes

| Path | Name | Breadcrumb | Guard |
|------|------|-----------|-------|
| `/users` | `Users` | Users | `requiresAuth` + `requiresAdmin` |
| `/users/create` | `UserCreate` | Tambah User | `requiresAuth` + `requiresAdmin` |
| `/users/:id/edit` | `UserEdit` | Edit User | `requiresAuth` + `requiresAdmin` |

The `requiresAdmin` guard fetches `/users/me` if needed and checks `currentUser.roles.includes('admin')`. Non-admins are redirected to `/`.

---

## 3. Files

| File | Role |
|------|------|
| `src/views/users/UserListView.vue` | Searchable, paginated table with row actions |
| `src/views/users/UserFormView.vue` | Create + edit (mode from `route.name`) |
| `src/services/user.service.js` | 8 endpoint wrappers |
| `src/components/common/AppTable.vue` | Generic table with slot cells + loading skeleton |
| `src/components/common/AppPagination.vue` | Numbered pager |
| `src/components/common/AppBadge.vue` | Role pill badges |
| `src/components/forms/ImageUpload.vue` | Avatar upload |
| `src/composables/useConfirm.js` | Confirm-before-delete |

---

## 4. API Endpoints

| Method | Path | Body | Used by |
|--------|------|------|---------|
| `GET`    | `/users?page&limit&q`            | —                     | List view |
| `GET`    | `/users/{id}`                     | —                     | Form view (edit) |
| `POST`   | `/users`                          | `{ email, name, password?, avatar?, roles[] }` | Form view (create) |
| `PUT`    | `/users/{id}`                     | `{ email?, name?, avatar?, roles[] }` | Form view (edit) |
| `DELETE` | `/users/{id}`                     | —                     | List view (delete) |
| `PUT`    | `/users/{id}/password`            | `{ old_password?, new_password }` | Form view (admin reset) |
| `POST`   | `/users/{id}/roles`               | `{ role }`            | Form view (sync role add) |
| `DELETE` | `/users/{id}/roles/{role}`        | —                     | Form view (sync role remove) |
| `GET`    | `/users/me`                       | —                     | AppLayout / router guard |

### Response envelope

```json
{
  "success": true,
  "data": {
    "users": [
      { "id": "...", "email": "...", "name": "...", "avatar_url": "...", "roles": ["admin"] }
    ],
    "pagination": { "page": 1, "limit": 10, "total": 42, "total_pages": 5 }
  }
}
```

---

## 5. Data Model

```ts
type User = {
  id: string
  email: string
  name: string
  avatar?: string | null
  avatar_url?: string | null   // server-rendered absolute or /files/... relative
  roles: string[]              // e.g. ["admin"], ["operator"]
  created_at?: string
  updated_at?: string
}
```

---

## 6. Behaviors

### 6.1 List (`UserListView.vue`)

- Search input (debounced 300 ms) hits `?q=` on every change after debounce.
- Pagination uses `pagination.total_pages`.
- Table columns: Avatar (40x40 round) | Nama | Email | Roles (badges) | Aksi.
- Row actions:
  - **Edit** → `/users/:id/edit`.
  - **Delete** → confirm dialog → `userService.remove(id)`. The currently logged-in user's own row hides the delete button (client-side guard).
- Empty state: "Tidak ada data pengguna."

### 6.2 Form (`UserFormView.vue`)

- Mode is determined by `route.name` (`UserCreate` vs `UserEdit`).
- **Create mode**:
  - Required: `email`, `name`, `password` (≥ 8 chars).
  - Optional: `avatar` (image), `roles[]`.
  - If avatar present → multipart upload; otherwise JSON.
- **Edit mode**:
  - Loads existing user.
  - Email may be readonly depending on backend contract; current UI allows editing.
  - On submit:
    1. `PUT /users/{id}` with the edited fields.
    2. **Sync roles** by diffing the new `roles[]` against the original and calling `POST /users/{id}/roles` for additions and `DELETE /users/{id}/roles/{role}` for removals. Runs in parallel.
- **Admin password reset** is rendered as a separate collapsible section on the edit form: optional `old_password` + required `new_password` (≥ 8 chars) → `PUT /users/{id}/password`.

### 6.3 Validation

- Email: regex `^[^\s@]+@[^\s@]+\.[^\s@]+$`.
- Password: length ≥ 8.
- Avatar: `useImagePreview` enforces JPEG/PNG/WebP and size limit.

---

## 7. UI Notes

- Form is a single column, max-width container with `AppInput` fields.
- Submit button shows `LoadingSpinner` while in-flight.
- Role selection uses checkboxes (multi-select). For small fixed role sets this is fine; for many roles a searchable multi-select would be better.

---

## 8. Notes & Gotchas

- **Self-delete guard is client-side only.** The backend should also reject it.
- **No bulk actions.** Selection + bulk delete / bulk role assign is not implemented.
- **Role sync** uses individual `POST` + `DELETE` calls per role diff — for many roles this could be slow.
- **Avatar uploads** rebuild `FormData` on the form — if switching between JSON and multipart on edit, ensure all fields are included.
- **Search** is exact-prefix on backend (`q=`) — confirm by reading swagger.