# 12 — Village Profile (Desa)

Single-record edit form for the village's identity. The simplest content page.

---

## 1. Purpose

Store the village's basic identity: name, address, contact info, vision & mission. A single record (no list, no create) — the backend holds at most one `desa` document.

---

## 2. Routes

| Path | Name | Breadcrumb | Guard |
|------|------|-----------|-------|
| `/profile` | `Profile` | Profil Desa | `requiresAuth` |

---

## 3. Files

| File | Role |
|------|------|
| `src/views/organization/ProfileView.vue` | View/edit toggle for the single record |
| `src/services/profile.service.js` | `GET /desa`, `PUT /desa` |
| `src/components/common/LoadingSpinner.vue` | Initial load spinner |

---

## 4. API Endpoints

| Method | Path | Body | Used by |
|--------|------|------|---------|
| `GET` | `/desa` | — | View (initial load) |
| `PUT` | `/desa` | `{ name, address?, phone?, email?, website?, vision_mission? }` | View (save) |

Public `/public/desa` is defined in swagger but not consumed by the admin.

---

## 5. Data Model

```ts
type Desa = {
  id?: string
  name: string                  // required
  address?: string
  phone?: string
  email?: string
  website?: string
  vision_mission?: string       // rich text or plain
  logo_url?: string             // server-rendered
  // any other backend-specific fields
}
```

---

## 6. Behaviors

- On mount → `profileService.get()` → populate form.
- **View mode**: read-only display of all fields. Logo thumbnail shown if present.
- **Edit mode**: fields become editable inputs (text, email, tel, url, textarea).
- Validation:
  - `name` required.
  - `email` (if present) matches email regex.
  - `website` (if present) starts with `http://` or `https://`.
- Save → `PUT /desa` → success toast + switch back to view mode.

---

## 7. UI Notes

- Single card with two sections: identity (name, address, phone, email, website) and vision/mission.
- A primary "Edit" button in the top-right; in edit mode it becomes "Save" + "Cancel".

---

## 8. Notes & Gotchas

- **No create flow** — if the backend returns 404 on `GET /desa`, the view may need to flip into "create" mode. Confirm by reading the view.
- **Vision/mission** is a plain textarea in the admin — the public site may render it differently (rich text, line breaks, etc.).
- **Logo upload** is not currently part of this view — the logo URL is server-rendered. If logo editing is needed, add an `ImageUpload` field.