# 02 — Dashboard

Single landing page after login. Shows 4 stat cards that summarise the size of each major content resource.

---

## 1. Purpose

Provide at-a-glance totals for the four most-edited content areas — users, berita (news), UMKM, and fasilitas — so an admin can spot unusual growth or stale data immediately.

---

## 2. Routes

| Path | Name | Breadcrumb | Guard |
|------|------|-----------|-------|
| `/` | `Dashboard` | Dashboard | `requiresAuth` |

---

## 3. Files

| File | Role |
|------|------|
| `src/views/dashboard/DashboardView.vue` | Page; runs 4 parallel `limit=1` list calls and reads `pagination.total` |
| `src/components/common/LoadingSpinner.vue` | Used during initial fetch |

No dedicated service — the view calls the existing domain services directly.

---

## 4. API Endpoints

The dashboard issues 4 list requests in parallel, each with `limit=1` to minimise payload:

| Method | Path | Query | Purpose |
|--------|------|-------|---------|
| `GET` | `/users` | `?page=1&limit=1` | Total user count |
| `GET` | `/berita` | `?page=1&limit=1` | Total berita count |
| `GET` | `/umkm` | `?page=1&limit=1` | Total UMKM count |
| `GET` | `/fasilitas` | `?page=1&limit=1` | Total fasilitas count |

Each response provides `pagination.total`, which is the only field the view consumes.

---

## 5. Data Model (displayed)

No persistent model — the view derives counts from each list endpoint's `pagination.total`:

| Card | Title | Source |
|------|-------|--------|
| Users | "Pengguna" | `GET /users` |
| Berita | "Berita" | `GET /berita` |
| UMKM | "UMKM" | `GET /umkm` |
| Fasilitas | "Fasilitas" | `GET /fasilitas` |

---

## 6. Behaviors

- On mount, runs all 4 requests in parallel (`Promise.all`).
- Renders `LoadingSpinner` while any of the 4 is pending.
- On any error → `notification.error(...)` (sticky).
- The four cards are clickable and navigate to the corresponding list page (`/users`, `/berita`, `/umkm`, `/fasilitas`) — though this is not enforced in the current view; the cards display counts only.

---

## 7. UI Notes

- 2x2 grid on desktop, stacked single column on mobile.
- Each card shows the icon, label, and the count in large type.
- Uses Tailwind dark-mode-aware colors (`bg-white dark:bg-gray-800`, etc.).

---

## 8. Notes & Gotchas

- **No caching** — every visit re-issues all 4 requests. Could be optimised with a short TTL cache or `stale-while-revalidate`.
- **No trend/history** — the cards only show the current total; there is no time-series.
- **No banner / PPID / struktur counts** — could be added to surface all content areas.
- **The `pagination.total` field is assumed** — if the backend omits it for some endpoints the card will show `0` or `undefined`.