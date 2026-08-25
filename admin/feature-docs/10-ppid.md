# 10 — PPID (Pejabat Pengelola Informasi dan Dokumentasi)

Public information documents + access requests. The only feature with an approval workflow.

---

## 1. Purpose

Manage information documents that the village is legally required to publish. Citizens can browse and request access to documents; admins upload documents, manage requests, and approve/revoke access. Approved requesters can download the document file.

---

## 2. Routes

| Path | Name | Breadcrumb | Guard |
|------|------|-----------|-------|
| `/ppid` | `PPID` | PPID | `requiresAuth` |
| `/ppid/create` | `PpidCreate` | Tambah Dokumen PPID | `requiresAuth` |
| `/ppid/:id/edit` | `PpidEdit` | Edit Dokumen PPID | `requiresAuth` |
| `/ppid-requests` | `PpidRequests` | PPID Requests | `requiresAuth` |

Public endpoints (`/public/ppid/...`) defined in swagger are consumed by the public-facing village website, not by the admin.

---

## 3. Files

| File | Role |
|------|------|
| `src/views/content/PpidListView.vue` | Document list with category filter + inline download |
| `src/views/content/PpidFormView.vue` | Create + edit document (upload file + thumbnail) |
| `src/views/content/PpidRequestsView.vue` | Request queue with approve/revoke |
| `src/services/ppid.service.js` | ~13 endpoint wrappers (public + admin) |
| `src/components/forms/FileUpload.vue` | Document file upload (PDF / DOC / XLS) |
| `src/components/forms/ImageUpload.vue` | Thumbnail upload |
| `src/components/common/AppTable.vue` | List + requests table |
| `src/components/common/AppPagination.vue` | Pager |
| `src/components/common/AppBadge.vue` | Status badges |
| `src/components/common/LoadingSpinner.vue` | Spinner |

---

## 4. API Endpoints

### Admin — documents

| Method | Path | Body | Used by |
|--------|------|------|---------|
| `GET`    | `/ppid?page&limit&q&category` | — | List view |
| `GET`    | `/ppid/{id}`                   | — | Form view (edit) |
| `POST`   | `/ppid`                        | multipart `FormData` | Form view (create) |
| `PUT`    | `/ppid/{id}`                   | multipart `FormData` | Form view (edit) |
| `DELETE` | `/ppid/{id}`                   | — | List view (delete) |

### Admin — categories

| Method | Path | Body | Used by |
|--------|------|------|---------|
| `GET`    | `/public/ppid/categories?page&limit&q` | — | List view (filter dropdown), Form view (datalist), CategoryManagerModal |
| `POST`   | `/ppid/categories`             | `{ name }` | CategoryManagerModal (create) |
| `DELETE` | `/ppid/categories/{id}`       | — | CategoryManagerModal (delete) |

### Admin — requests

| Method | Path | Body | Used by |
|--------|------|------|---------|
| `GET`    | `/ppid/requests?page&limit&status` | — | Requests view |
| `POST`   | `/ppid/requests/{id}/approve`      | — | Requests view (approve) |
| `POST`   | `/ppid/requests/{id}/revoke`       | — | Requests view (revoke) |
| `GET`    | `/ppid/document/{documentId}/download` | — | List view (download blob) |

### Public (not consumed by admin, declared in swagger)

| Method | Path |
|--------|------|
| `GET`    | `/public/ppid/list?page&limit&q&category` |
| `GET`    | `/public/ppid/{id}` |
| `POST`   | `/public/ppid/{id}/requests` |

### FormData fields (admin upload)

| Field | Type | Required |
|-------|------|----------|
| `title` | string | yes |
| `category` | UUID | yes (category id, resolved from datalist or auto-created on submit) |
| `description` | string | optional |
| `document` | File (PDF/DOC/XLS) | yes on create |
| `thumbnail` | File (JPEG/PNG/WebP) | optional |

---

## 5. Data Model

```ts
type PpidDocument = {
  id: string              // UUID
  title: string
  category: string
  description?: string
  document_url: string    // only available to approved requesters (server-controlled)
  publication_at?: string
  created_at?: string
  updated_at?: string
}

type PpidRequest = {
  id: string              // UUID
  ppid_id: string         // UUID
  requester_name: string
  requester_email: string
  purpose: string
  status: 'pending' | 'approved' | 'revoked'
  approved_at?: string
  approved_by?: string    // admin user id
  revoked_at?: string
  revoked_by?: string
  created_at?: string
}
```

---

## 6. Behaviors

### 6.1 Document List (`PpidListView.vue`)

- Category filter (populated from `GET /public/ppid/categories`).
- Search input (`?q=`).
- Table columns: Thumbnail | Judul | Kategori | Tanggal Publikasi | Aksi.
- **Download action**:
  - Calls `GET /ppid/document/{documentId}/download` (singular `document` — matches swagger).
  - Response is a `Blob`.
  - Creates an object URL + invisible `<a download>` + click to trigger download.
- Edit + delete as usual.

### 6.2 Document Form (`PpidFormView.vue`)

- Mode from `route.name`.
- Fields: title, category, description, publication_at, document file (`FileUpload`), thumbnail (`ImageUpload`).
- **Category field**: free-text `AppInput` with a `<datalist>` autocomplete sourced from the category list, plus a `Kelola Kategori` button that opens the reusable `CategoryManagerModal`.
- File validation: PDF / DOC / DOCX / XLS / XLSX (via `isValidDocumentType`).

### 6.3 Requests (`PpidRequestsView.vue`)

- Table columns: Pemohon | Email | Dokumen | Tujuan | Status | Aksi.
- Status badge: pending (warning), approved (success), revoked (danger).
- Actions:
  - **Approve** (only on `pending`) → confirm → `POST /ppid/requests/{id}/approve`.
  - **Revoke** (only on `approved`) → confirm → `POST /ppid/requests/{id}/revoke`.

---

## 7. UI Notes

- `FileUpload` is a drag-and-drop area with a click-to-select fallback. Shows filename + size after selection.
- Download icon is the standard "download" feather icon.
- Status badges use `AppBadge` with `warning` / `success` / `danger` variants.

---

## 8. Notes & Gotchas

- **Document URL visibility** is controlled server-side; the admin download endpoint presumably bypasses the approval check. Confirm intent.
- **No bulk approve / reject**.
- **No requester notification** is triggered from the admin frontend — the backend presumably sends an email.
- **Category manager** uses the reusable `CategoryManagerModal` component. The delete endpoint returns 409 when a category is still referenced by a document — surfaced as "Kategori masih digunakan dan tidak dapat dihapus".