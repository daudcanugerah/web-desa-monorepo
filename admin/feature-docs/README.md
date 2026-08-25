# Webdesa Admin — Feature Documentation

Comprehensive documentation for the **Webdesa Admin Dashboard** frontend (Vue 3 + Vite + Pinia + Tailwind).

This directory is the canonical reference for every feature area, the data they touch, the API endpoints they call, and the components / stores / composables that power them.

---

## Table of Contents

| # | Document | Scope |
|---|----------|-------|
| 00 | [Architecture & Cross-Cutting Concerns](./00-architecture.md) | Tech stack, layering, auth flow, axios 401 refresh, route guards, service pattern, error handling, i18n, theme |
| 01 | [Authentication](./01-authentication.md) | Login, forgot password, reset password, JWT lifecycle |
| 02 | [Dashboard](./02-dashboard.md) | Stat cards (users / berita / umkm / fasilitas) |
| 03 | [User Management](./03-user-management.md) | CRUD users, role assignment, password reset (admin only) |
| 04 | [My Profile](./04-my-profile.md) | Self-service profile + password change |
| 05 | [Role & Permission Management](./05-role-permission.md) | Role CRUD, permission matrix, diff-and-sync |
| 06 | [Banner Management](./06-banner.md) | Hero/carousel banners with image + status toggle |
| 07 | [News / Berita](./07-news.md) | Rich-text articles with Quill, categories, inline media |
| 08 | [UMKM](./08-umkm.md) | Local micro-business listings with multi-image |
| 09 | [Facility / Fasilitas](./09-facility.md) | Village facilities with map picker + image gallery |
| 10 | [PPID](./10-ppid.md) | Public information documents, requests, approve/revoke |
| 11 | [Organization Structure](./11-structure.md) | Village officials with photos + positions |
| 12 | [Village Profile](./12-village-profile.md) | Single-record village identity (name, address, vision/mission) |
| 13 | [Profile Sections](./13-profile-sections.md) | Sub-pages of the public profile with rich text |
| 14 | [Infographic Dashboard](./14-infographic.md) | Metabase dashboard/question embeds with preview tokens |
| 15 | [UI / Cross-Cutting](./15-ui-cross-cutting.md) | Layout, sidebar, header, toast, confirm dialog, theme, locale |

---

## Document Conventions

Every feature document follows the same structure:

1. **Purpose** — what the feature does in plain language.
2. **Routes** — every URL the feature owns (path / name / breadcrumb / guard).
3. **Files** — views, services, stores, components, composables.
4. **API endpoints** — full HTTP method + path + payload shape.
5. **Data model** — TypeScript-ish field listing for the resource(s).
6. **Behaviors** — list of capabilities (CRUD, filters, validations, edge cases).
7. **UI notes** — table columns, form fields, modals, special widgets.
8. **Notes & gotchas** — known quirks, inconsistencies, TODOs.

---

## Quick Stack Snapshot

| Layer | Technology |
|-------|-----------|
| Framework | Vue 3.4 (Composition API + `<script setup>`) |
| Build | Vite |
| State | Pinia (options API) |
| Router | Vue Router 4 (lazy-loaded, HTML5 history) |
| HTTP | Axios + custom interceptors |
| Styling | Tailwind CSS 3.4 (with dark mode) |
| Validation | Hand-rolled (VeeValidate + Yup declared but unused) |
| i18n | Vue-i18n 9 (`id` default, `en`) |
| Rich text | Quill 2 + quill-better-table |
| Maps | Leaflet 1.9 + leaflet-minimap |
| Analytics | Metabase embed.js |

Base URL: `VITE_API_BASE_URL` (default `http://localhost:8080/api/v1`).
Metabase URL: `VITE_METABASE_URL` (default `http://localhost:3000`).

---

## Related Documentation

- `ARCHITECTURE.md` — short top-level architecture overview (repo root)
- `ROUTES.md` — exhaustive route table
- `FEATURES.md` — original single-file feature overview (preserved for reference)
- `SPEC.md` — one-page spec / quick reference
- `swagger.yaml` — OpenAPI 3 spec for the backend