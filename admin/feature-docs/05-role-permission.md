# 05 — Role & Permission Management

CRUD for roles + a permission matrix (resource × action) per role. The only fully i18n'd view in the app.

---

## 1. Purpose

Define roles and assign permissions to them so that admins can control what each role can do across every domain.

Permission shape: `{ resource, action }` where `resource` is a domain name (e.g. `users`, `berita`) and `action` is `read` or `write`.

---

## 2. Routes

| Path | Name | Breadcrumb | Guard |
|------|------|-----------|-------|
| `/roles` | `Roles` | Roles | `requiresAuth` |

---

## 3. Files

| File | Role |
|------|------|
| `src/views/users/RoleListView.vue` | Table + 3 modals (create, edit, permissions) |
| `src/services/role.service.js` | 6 endpoint wrappers |
| `src/components/common/AppTable.vue` | Role table |
| `src/components/common/AppModal.vue` | Create / edit / permissions modals |
| `src/components/common/LoadingSpinner.vue` | Spinner inside modals |

---

## 4. API Endpoints

| Method | Path | Body | Used by |
|--------|------|------|---------|
| `GET`    | `/roles`                              | —                              | List view |
| `POST`   | `/roles`                              | `{ name }`                     | Create modal |
| `DELETE` | `/roles/{role}`                       | —                              | List view (delete) |
| `GET`    | `/roles/{role}/permissions`           | —                              | Permissions modal (load) |
| `POST`   | `/roles/{role}/permissions`           | `{ resource, action }`         | Permissions modal (sync add) |
| `DELETE` | `/roles/{role}/permissions`           | `{ resource, action }`         | Permissions modal (sync remove) |

The list view also loads `GET /users` to compute per-role user counts (each user has a `roles[]`).

---

## 5. Data Model

```ts
type Role = { name: string }  // role name is the primary key

type Permission = {
  resource: string  // e.g. 'users', 'banners', 'berita', 'umkm', 'fasilitas',
                    //      'ppid', 'struktur', 'desa', 'roles', 'profile', 'infographic'
  action: 'read' | 'write'
}
```

### Hard-coded permission catalog

`RoleListView` declares a `KNOWN_PERMISSIONS` array (12 entries) used to seed the matrix UI. The current set:

| Resource | read | write |
|----------|------|-------|
| `users`      | ✓ | ✓ |
| `banners`    | ✓ | ✓ |
| `berita`     | ✓ | ✓ |
| `umkm`       | ✓ | ✓ |
| `fasilitas`  | ✓ | ✓ |
| `ppid`       | ✓ | ✓ |
| `struktur`   | ✓ | ✓ |
| `desa`       | ✓ | ✓ |
| `roles`      | ✓ | ✓ |
| `profile`    | ✓ | ✓ |
| `infographic`| ✓ | ✓ |

(`desa`, `profile`, `infographic` are likely implicitly supported — confirm with backend.)

---

## 6. Behaviors

### 6.1 List

- Columns: Name | User Count | Actions.
- Default roles (`admin`, `operator`) cannot be deleted — the trash icon is hidden.
- User count = number of users whose `roles[]` includes this role.

### 6.2 Create Role

- Modal with one field: `name` (required, unique).
- Submit → `roleService.create({ name })` → reload list.

### 6.3 Edit Role (rename)

- Modal with `name` field pre-filled.
- For default roles, the modal may not allow renaming (frontend decision; verify).

### 6.4 Manage Permissions

- Modal opens for a chosen role.
- Loads `GET /roles/{role}/permissions`.
- Renders a 2-column table (resource / read+write checkboxes).
- "Select all" + "Deselect all" buttons per row.
- On save:
  1. Compute diff: `toAdd` = (desired − current), `toRemove` = (current − desired).
  2. Fire all `POST /roles/{role}/permissions` (add) and `DELETE /roles/{role}/permissions` (remove) in parallel.
  3. On any failure → revert and show error toast.
  4. On success → success toast + close modal.

### 6.5 Delete Role

- Confirm dialog → `DELETE /roles/{role}`.
- Default roles skip the delete button entirely.

---

## 7. UI Notes

- This is the **only fully i18n'd view** — column headers, modal labels, and button text use `$t()` keys from `id.json` / `en.json`.
- The permissions matrix uses a sticky-header scrollable table inside the modal so it works on small screens.

---

## 8. Notes & Gotchas

- **`KNOWN_PERMISSIONS` is duplicated** — the same list is implicitly needed by the backend's auth middleware. Adding a new resource requires editing both sides.
- **Permission writes are 1-per-call** — for a role with N permissions, a full save can issue 2N requests. A batch endpoint would be faster.
- **Role rename** does not update existing user role references atomically — confirm backend behaviour.
- **Default-role protection** is client-side only.