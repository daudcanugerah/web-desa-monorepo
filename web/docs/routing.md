# Routing

`createWebHistory()` (HTML5 mode — needs server-side fallback in prod).

Single parent route renders `Layout.vue` (Navbar + `<RouterView />` + Footer).

```
/
└── /                 (Layout.vue)
    ├── ''              → Home
    ├── profil          → Profil
    ├── infografik      → Infografik
    ├── peta            → Peta
    ├── berita          → Berita
    ├── berita/:slug    → BeritaDetail
    ├── umkm            → UMKM
    └── ppid            → PPID
```

- No named routes.
- No `beforeEach` guards. All routes public.
- No route meta (`requiresAuth`, roles, etc.).
- No 404 catch-all — unmatched paths fall through silently.
- Filter/pagination state lives in page-local refs, not query params.

## Navbar Links (`components/layout/Navbar.vue`)

| Path | Label |
|---|---|
| `/` | Beranda |
| `/profil` | Profil |
| `/infografik` | Infografik |
| `/peta` | Peta |
| `/berita` | Berita |
| `/umkm` | UMKM |
| `/ppid` | PPID |

`isActive(path)` uses `route.path === path` for `/`, `startsWith(path)` otherwise. Mobile hamburger toggles `isMenuOpen`; click on a link closes the menu.

## Footer Quick Links

Profil Desa, Berita, PPID. Static hardcoded in `Footer.vue`.
