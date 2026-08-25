# 09 — Facility / Fasilitas

Village facilities (public places, services) with map coordinates and an image gallery. The only feature that uses the `MapPicker` component for location picking.

---

## 1. Purpose

CRUD for village facilities (puskesmas, balai desa, schools, etc.). Each facility has a name, category, description, address, geographic coordinates, and a gallery of images.

---

## 2. Routes

| Path | Name | Breadcrumb | Guard |
|------|------|-----------|-------|
| `/fasilitas` | `Fasilitas` | Fasilitas | `requiresAuth` |
| `/fasilitas/create` | `FasilitasCreate` | Tambah Fasilitas | `requiresAuth` |
| `/fasilitas/:id/edit` | `FasilitasEdit` | Edit Fasilitas | `requiresAuth` |

---

## 3. Files

| File | Role |
|------|------|
| `src/views/content/FacilityListView.vue` | Searchable, paginated list |
| `src/views/content/FacilityFormView.vue` | Create + edit, split layout (form left, map right) + fullscreen map modal |
| `src/services/facility.service.js` | 5 endpoint wrappers + image-removal support |
| `src/components/forms/MapPicker.vue` | Leaflet map with village polygon + draggable marker + fullscreen toggle |
| `src/components/forms/ImageUpload.vue` | Per-image upload |
| `src/components/common/AppTable.vue` | List table |
| `src/components/common/AppPagination.vue` | Pager |
| `src/components/common/AppBackButton.vue` | Back-to-list |

---

## 4. API Endpoints

| Method | Path | Body | Used by |
|--------|------|------|---------|
| `GET`    | `/fasilitas?page&limit`        | —                  | List view |
| `GET`    | `/fasilitas/{id}`              | —                  | Form view (edit) |
| `POST`   | `/fasilitas`                   | multipart `FormData` | Form view (create) |
| `PUT`    | `/fasilitas/{id}`              | multipart `FormData` | Form view (edit) |
| `DELETE` | `/fasilitas/{id}`              | —                  | List view (delete) |
| `DELETE` | `/fasilitas/{id}/images/{imageIndex}` | —           | Form view (remove single image) |

Public endpoints (`/public/fasilitas/...`) and `/fasilitas/{id}/images/{imageIndex}` (single-image delete) are defined in swagger but the admin edit flow uses `removed_images` instead.

### FormData fields

| Field | Type | Notes |
|-------|------|-------|
| `name` | string | required |
| `category` | string | required |
| `description` | string | required |
| `address` | string | required |
| `latitude` | number | -90 to 90 |
| `longitude` | number | -180 to 180 |
| `images` | File[] | one or more |
| `removed_images` | JSON string array | filenames of existing images to delete |

---

## 5. Data Model

```ts
type Facility = {
  id: string
  name: string
  category: string
  description: string
  address: string
  latitude: number
  longitude: number
  images: string[]           // filenames
  created_at?: string
  updated_at?: string
}
```

---

## 6. Behaviors

### 6.1 List (`FacilityListView.vue`)

- Pagination only (no search).
- Table columns: Thumbnail | Nama | Kategori | Alamat | Aksi.

### 6.2 Form (`FacilityFormView.vue`)

- Split layout: form left, map right.
- Form fields: name, type, description (textarea), latitude, longitude.
- **MapPicker**:
  - Renders a Leaflet map with the village polygon overlay (`area-desa-poly.json`).
  - Has a draggable marker; click on map also moves it.
  - Constrains the marker to the polygon bounds — drops outside snap back to the polygon centroid.
  - Includes a leaflet-minimap in the corner.
  - Fullscreen modal variant for precise picking (button at top-right of map).
- **Images**:
  - Existing images shown as thumbnails with a remove (×) icon each.
  - Removing an existing image calls `DELETE /fasilitas/{id}/images/{imageIndex}` immediately (optimistic UI; rolls back on failure).
  - New images can be added via `<input type="file" multiple>` and are submitted in `FormData` as `images`.
- **Validation**:
  - Latitude: `-90 ≤ v ≤ 90`.
  - Longitude: `-180 ≤ v ≤ 180`.
- **Submit**:
  - Builds `FormData` with all fields.
  - Image removal is handled out-of-band via the DELETE endpoint above — `FormData` no longer carries a `removed_images` field.

---

## 7. UI Notes

- The map is interactive — drag, click, zoom.
- The fullscreen map modal is the only place where the entire screen is dedicated to picking a point (useful for precise GPS coordinates).
- The polygon overlay gives immediate visual feedback if the marker is dragged outside the village area.

---

## 8. Notes & Gotchas

- **Image removal** uses `DELETE /fasilitas/{id}/images/{imageIndex}` (per-image, optimistic UI). The view rolls back the local removal if the request fails.
- **Polygon source** is a static asset (`src/assets/maps/area-desa-poly.json`). If the village boundary changes, the asset must be updated and the app redeployed. Not ideal — consider serving from backend.
- **No geocoding** — the admin must manually pick coordinates or type them.
- **No bulk delete** — images must be removed one-by-one.
- **The legacy `removed_images` FormData field is no longer sent** — if the backend still parses it for backwards compatibility, that's harmless; if the new DELETE endpoint isn't implemented on the backend yet, image removal will 404.