# Webdesa — Public Village Website

Responsive public website for an Indonesian village (sample: **Desa Palasari**).
Vue 3 SPA that consumes the separate Go REST API (`webdesa/api`).

## Features

- **Home** — banner slider, stats, latest news, village sambutan
- **Profil** — history, vision/mission sections, village officials
- **Infografik** — Metabase dashboard/question embeds (JWT tokens)
- **Peta** — Leaflet map of the village polygon + facility markers, base-layer
  switcher, Google-Maps-style list ⇄ place-detail panel
- **Berita** — news list + article detail
- **UMKM** — local business directory
- **PPID** — public information documents + request submission
- **Galeri** — public gallery (public folders + public media only)

## Tech stack

- **Framework**: Vue 3 (Options API shell + Composition `setup()`, no `<script setup>`)
- **Build**: Vite 4
- **Styling**: TailwindCSS 3
- **Routing**: Vue Router 4
- **Mapping**: Leaflet + leaflet-minimap + leaflet.markercluster
- **Language**: JavaScript

## Project structure

```
src/
├── components/
│   ├── layout/         # Layout, TopBar, Navbar, Footer
│   └── common/         # Icon, Pagination, SearchInput, PlaceDetailPanel, …
├── pages/              # one file per route
├── composables/        # useDesaInfo, useFeatureFlags
├── services/
│   ├── apiClient.js    # fetch-based HTTP client singleton
│   └── desaService.js  # all public API access
├── router/index.js
├── App.vue
├── main.js
└── style.css
public/
└── area-desa-poly.json # village polygon ([lng, lat] pairs)
```

## Routes

| Path | Page |
|---|---|
| `/` | Home |
| `/profil` | Profil |
| `/infografik` | Infografik |
| `/peta` | Peta |
| `/berita` | Berita |
| `/berita/:id` | BeritaDetail |
| `/umkm` | UMKM |
| `/ppid` | PPID |
| `/galeri` | Galeri |
| `/galeri/:id` | GaleriDetail |
| `*` | NotFound |

## Getting started

```bash
npm install
npm run dev       # dev server on :5173
npm run build     # production build
npm run preview   # preview the build
```

Requires the API running. Configure via `.env`:

```ini
VITE_API_BASE_URL=http://localhost:8081/api/v1
VITE_METABASE_URL=http://bi-embed.desapalasari.my.id
VITE_FEATURE_GALLERY_PUBLIC=false
```

`VITE_API_BASE_URL` may be given with or without the `/api/v1` suffix — it is
appended automatically if missing.

## Data layer

All data access lives in `src/services/desaService.js` (see
[`docs/services.md`](./docs/services.md)). It talks **only** to the backend —
there are no mocks. Responses are `{ success, data }`; helpers (`unwrapData`,
`unwrapList`, `unwrapPaginated`) normalize them and flatten `category` objects
to plain names.

Media URLs returned by the API are signed, relative paths
(`/api/v1/media/{id}/...?jwt=`); resolve them with `resolveGalleryAssetUrl()`.

## Docs

See [`docs/`](./docs/README.md) for architecture, routing, services, components,
and per-feature notes.
