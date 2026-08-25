# 06 — Banner Management

CRUD for hero / carousel banners shown on the public site.

---

## 1. Purpose

Maintain a set of banners (image + title + caption + status) used on the public-facing village website. Admins create, edit, delete, and toggle active/inactive status.

---

## 2. Routes

| Path | Name | Breadcrumb | Guard |
|------|------|-----------|-------|
| `/banners` | `Banners` | Banners | `requiresAuth` |
| `/banners/create` | `BannerCreate` | Tambah Banner | `requiresAuth` |
| `/banners/:id/edit` | `BannerEdit` | Edit Banner | `requiresAuth` |

---

## 3. Files

| File | Role |
|------|------|
| `src/views/content/BannerListView.vue` | Filterable list with status toggle + delete |
| `src/views/content/BannerFormView.vue` | Create + edit (mode from `route.name`) |
| `src/services/banner.service.js` | 6 endpoint wrappers (multipart for create/update) |
| `src/components/common/AppTable.vue` | List table |
| `src/components/common/AppPagination.vue` | Pager |
| `src/components/common/AppButton.vue` | Action buttons |
| `src/components/common/AppInput.vue` | Text fields |
| `src/components/forms/ImageUpload.vue` | Banner image upload |
| `src/components/common/AppBackButton.vue` | Back-to-list link |

---

## 4. API Endpoints

| Method | Path | Body | Used by |
|--------|------|------|---------|
| `GET`    | `/banners?page&limit&status`      | —                       | List view |
| `GET`    | `/banners/{id}`                    | —                       | Form view (edit) |
| `POST`   | `/banners`                         | multipart `FormData`    | Form view (create) |
| `PUT`    | `/banners/{id}`                    | multipart `FormData`    | Form view (edit) |
| `DELETE` | `/banners/{id}`                    | —                       | List view (delete) |
| `PATCH`  | `/banners/{id}/status`             | `{ status: 'active' \| 'inactive' }` | List view (inline toggle) |

`FormData` fields: `title`, `caption?`, `image` (required on create; optional on edit), `status`.

Public endpoints (`/public/banner/...`) defined in swagger but not consumed by the admin.

---

## 5. Data Model

```ts
type Banner = {
  id: string
  title: string
  caption?: string | null
  image: string                  // filename served via /files/{image}
  image_url?: string             // sometimes pre-rendered
  status: 'active' | 'inactive'
  created_at?: string
  updated_at?: string
}
```

---

## 6. Behaviors

### 6.1 List (`BannerListView.vue`)

- Status filter dropdown: Semua / Aktif / Nonaktif.
- Table columns: Image (thumbnail 60x40) | Judul | Caption | Status (badge) | Aksi.
- Inline status toggle via dropdown menu → `PATCH /banners/{id}/status`.
- Delete via confirm dialog → `DELETE /banners/{id}`.

### 6.2 Form (`BannerFormView.vue`)

- Mode from `route.name`.
- Fields:
  - `title` — required.
  - `caption` — optional textarea.
  - `image` — required on create; on edit, shows current image with option to replace.
  - `status` — radio (Aktif / Nonaktif), default Aktif.
- Submit builds `FormData` and posts/puts.

---

## 7. UI Notes

- Form is single-column inside a constrained container with `AppBackButton` at the top.
- Image preview shows 16:9 thumbnail via `useImagePreview`.

---

## 8. Notes & Gotchas

- **`Content-Type` header** is set explicitly in `banner.service.js` for multipart — confirm the backend tolerates both `multipart/form-data; boundary=...` and the simpler string.
- **No ordering field** — banners have no explicit sort order. If the backend orders by `created_at`, the UI cannot reorder them.
- **No start/end date** — banners are simply active/inactive, no scheduled visibility.