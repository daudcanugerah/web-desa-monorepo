# Webdesa Docs

Public website frontend for **Desa Sukamaju** (sample Indonesian village).

Vue 3 SPA. Talks to separate REST API at `VITE_API_BASE_URL` (default `http://localhost:8080/api/v1`).

## Index

### Core
- [architecture.md](./architecture.md) — stack, file layout, conventions, data flow
- [routing.md](./routing.md) — routes, layout, nav
- [services.md](./services.md) — `apiClient` + `desaService` API reference
- [components.md](./components.md) — shared components

### Features
- [features/home.md](./features/home.md) — landing
- [features/profil.md](./features/profil.md) — village profile
- [features/infografik.md](./features/infografik.md) — statistics
- [features/peta.md](./features/peta.md) — Leaflet map
- [features/berita.md](./features/berita.md) — news list + detail
- [features/umkm.md](./features/umkm.md) — SME directory
- [features/ppid.md](./features/ppid.md) — public info disclosure

## Run

```bash
npm install
npm run dev      # http://localhost:5173
npm run build
npm run preview
```

## Backend Contract

Source of truth: [`../openapi.yaml`](../openapi.yaml). All `/public/*` endpoints are the ones consumed here.
