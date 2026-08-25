# 08 — UMKM

Local micro-business (Usaha Mikro, Kecil, dan Menengah) listings with multi-image galleries.

---

## 1. Purpose

CRUD for village UMKM listings. Each entry has owner name, business name, description, contact info, address, and a gallery of images.

---

## 2. Routes

| Path | Name | Breadcrumb | Guard |
|------|------|-----------|-------|
| `/umkm` | `UMKM` | UMKM | `requiresAuth` |
| `/umkm/create` | `UmkmCreate` | Tambah UMKM | `requiresAuth` |
| `/umkm/:id/edit` | `UmkmEdit` | Edit UMKM | `requiresAuth` |

---

## 3. Files

| File | Role |
|------|------|
| `src/views/content/UmkmListView.vue` | Searchable, paginated list |
| `src/views/content/UmkmFormView.vue` | Create + edit (mode from `route.name`) |
| `src/services/umkm.service.js` | 5 endpoint wrappers (multipart) |
| `src/components/forms/ImageUpload.vue` | Per-image upload (used in a v-for gallery) |
| `src/components/common/AppTable.vue` | List table |
| `src/components/common/AppPagination.vue` | Pager |
| `src/components/common/AppBackButton.vue` | Back-to-list |

No dedicated composable — uses `useImagePreview` per image slot.

---

## 4. API Endpoints

| Method | Path | Body | Used by |
|--------|------|------|---------|
| `GET`    | `/umkm?page&limit&search` | —                  | List view |
| `GET`    | `/umkm/{id}`              | —                  | Form view (edit) |
| `POST`   | `/umkm`                   | multipart `FormData` | Form view (create) |
| `PUT`    | `/umkm/{id}`              | multipart `FormData` | Form view (edit) |
| `DELETE` | `/umkm/{id}`              | —                  | List view (delete) |
| `GET`    | `/public/umkm/categories?page&limit&q` | —                  | Form view (datalist) + CategoryManagerModal |
| `POST`   | `/umkm/categories`        | `{ name }`         | CategoryManagerModal (create) |
| `DELETE` | `/umkm/categories/{id}`   | —                  | CategoryManagerModal (delete) |

Public endpoints (`/public/umkm/...`) and `/umkm/categories` are defined in swagger but not used by the admin frontend.

### FormData fields

| Field | Type | Notes |
|-------|------|-------|
| `name` | string | business name |
| `owner` | string | |
| `description` | string | |
| `category` | UUID | category id, resolved from datalist or auto-created on submit |
| `phone` | string | |
| `address` | string | |
| `images` | File[] | one or more — sent as `images[]` / repeated field |
| `existing_images` | string[] | filenames already on the server (preserved on edit) |

---

## 5. Data Model

```ts
type Umkm = {
  id: string
  name: string
  owner_name: string
  description: string
  category: string
  phone?: string
  address?: string
  images: string[]            // filenames
  created_at?: string
  updated_at?: string
}
```

---

## 6. Behaviors

### 6.1 List (`UmkmListView.vue`)

- Search input (debounced 300 ms → `?search=`).
- Table columns: Thumbnail (first image) | Nama | Pemilik | Kategori | Aksi.
- Delete with confirm.

### 6.2 Form (`UmkmFormView.vue`)

- Mode from `route.name`.
- Fields:
  - `name` — required.
  - `owner` — required.
  - `category` — free-text `AppInput` with a `<datalist>` autocomplete sourced from `/public/umkm/categories`, plus a `Kelola Kategori` button that opens the reusable `CategoryManagerModal`.
  - `phone` — required (simple format check).
  - `address` — required.
  - `description` — textarea.
- **Images gallery**:
  - On create: array of `useImagePreview` slots, each with drag-and-drop.
  - On edit: existing images shown as thumbnails; each has a remove (×) button. New images can be added.
  - On submit: **re-fetches existing images** via `fetch(getImageUrl(img)).then(r => r.blob())` and re-attaches them as `File` objects in `FormData`, alongside the new files. This is a workaround for a backend that does not have a separate "preserve" path.
- Remove button only removes from the local preview state — final removal happens implicitly because the file is no longer in the submission.

---

## 7. UI Notes

- Form uses a 2-column layout on desktop (form left, image gallery right), single column on mobile.
- Gallery is a responsive grid of square thumbnails.

---

## 8. Notes & Gotchas

- **Image re-fetch quirk**: see §6.2. If `/files/...` is not CORS-enabled, edit-mode image submission will fail. A cleaner fix would be a `keep_images[]` field on the backend that preserves by filename.
- **No image order / cover flag**: there is no way to mark a primary image or reorder.
- **Category autocomplete** is missing — could improve UX by adding a free-text-with-suggestions component.
- **FormData `Content-Type` header** is not set explicitly here — Axios auto-generates the boundary. Confirm backend compatibility.