# 11 — Organization Structure (Struktur Organisasi)

Single-page CRUD for village officials (kepala desa, sekretaris, kaur, etc.).

---

## 1. Purpose

Maintain the list of village government officials — name, position, photo, optional contact info. Rendered on the public village site.

---

## 2. Routes

| Path | Name | Breadcrumb | Guard |
|------|------|-----------|-------|
| `/struktur` | `Struktur` | Struktur Organisasi | `requiresAuth` |

There is **no separate create/edit route** — both actions happen inside modals on the same page.

---

## 3. Files

| File | Role |
|------|------|
| `src/views/organization/StructureView.vue` | List + create/edit/delete via `AppModal` |
| `src/services/structure.service.js` | 5 endpoint wrappers (multipart) |
| `src/components/common/AppModal.vue` | Create / edit modal |
| `src/components/common/AppTable.vue` | Officials table |
| `src/components/common/AppInput.vue` | Text fields |
| `src/components/forms/ImageUpload.vue` | Profile photo |

---

## 4. API Endpoints

| Method | Path | Body | Used by |
|--------|------|------|---------|
| `GET`    | `/struktur?page&limit&search` | —                  | List |
| `GET`    | `/struktur/{id}`              | —                  | Edit modal (load) |
| `POST`   | `/struktur`                   | multipart `FormData` | Create modal |
| `PUT`    | `/struktur/{id}`              | multipart `FormData` | Edit modal |
| `DELETE` | `/struktur/{id}`              | —                  | Delete |

Public endpoints (`/public/struktur/...`) defined in swagger but not consumed by the admin.

### FormData fields

| Field | Type | Required |
|-------|------|----------|
| `name` | string | yes |
| `position` | string | yes |
| `photo` | File | optional |
| `phone` | string | optional |
| `email` | string | optional |
| `order` | number | optional (display order) |

---

## 5. Data Model

```ts
type Official = {
  id: string
  name: string
  position: string
  photo?: string              // filename
  photo_url?: string
  phone?: string
  email?: string
  order?: number
  created_at?: string
  updated_at?: string
}
```

---

## 6. Behaviors

### 6.1 List

- Search input (debounced 300 ms → `?search=`).
- Table columns: Photo (round) | Nama | Jabatan | Kontak | Aksi.
- Pagination via `pagination.total_pages`.

### 6.2 Create / Edit (Modal)

- "Tambah" button → opens modal in create mode (empty form).
- Row "Edit" → opens modal in edit mode (pre-filled).
- Modal contains: name, position, photo, phone, email.
- Photo uses `useImagePreview` for preview + validation.
- Submit → service call → reload list.

### 6.3 Delete

- Row "Delete" → confirm dialog → `DELETE /struktur/{id}`.

---

## 7. UI Notes

- Single page with a table; the modal appears on top of it.
- The "Tambah" button is at the top-right of the table card.
- Round photo thumbnail in the table.

---

## 8. Notes & Gotchas

- **No drag-to-reorder** despite an `order` field — the order is presumably set via the form input.
- **No nested hierarchy** — there is no way to express "Kaur Umum reports to Sekdes". A tree/parent_id field would be needed.
- **No separate create/edit routes** — everything is on `/struktur`. This is intentional (low-cardinality CRUD), but if the form grows it may need to be split.