# Architecture

## Stack

| Layer | Choice |
|---|---|
| Framework | Vue 3 (Composition API via `setup()`, **not** `<script setup>`) |
| Build | Vite 4 |
| Router | vue-router 4 (HTML5 history) |
| Styling | TailwindCSS 3 + PostCSS + Autoprefixer |
| Maps | Leaflet + leaflet-minimap + leaflet.markercluster |
| Lang | JavaScript ES2022+, no TypeScript |
| Node | 16+ |

No tests, no linter, no formatter, no CI, no Docker. Three npm scripts: `dev`, `build`, `preview`.

## File Layout

```
src/
  App.vue                 # <RouterView />
  main.js                 # createApp + router
  style.css               # Tailwind + Leaflet CSS + custom @layer
  router/index.js         # routes
  pages/<Name>.vue        # route components (PascalCase)
  components/
    layout/               # Layout, Navbar, Footer
    common/               # LoadingSpinner, Card, EmptyState
  services/
    apiClient.js          # HTTP client singleton
    desaService.js        # all data access
public/
  area-desa-poly.json     # village polygon ([lng,lat] pairs) loaded by Peta.vue
openapi.yaml              # backend contract (2152 lines)
```

## Component Style

Every page uses the **Options API shell + Composition `setup()` returning refs**. No `<script setup>`.

```js
export default {
  name: 'PageName',
  components: { LoadingSpinner, EmptyState },
  setup() {
    const data = ref(null)
    const loading = ref(true)
    onMounted(async () => {
      try { data.value = await fetchSomething() }
      catch (e) { console.error(...) }
      finally { loading.value = false }
    })
    return { data, loading }
  }
}
```

## Naming

- Files: `PascalCase.vue`
- Refs/computed/methods: `camelCase`
- API functions: `getPublic<Resource>List` / `getPublic<Resource>ById` / `create<Resource>...`
- UI copy: Indonesian. Keys/props: English.

## State

No Pinia/Vuex. Local `ref()` per page. No shared composables. `apiClient` is module-level singleton.

## Styling

- Tailwind utility-only. No CSS Modules.
- Global CSS in `style.css` — Tailwind directives + custom `@layer components` (`.animate-fade-in`, `.line-clamp-2`, etc.).
- Brand color: `emerald-600`. Container: `max-w-7xl mx-auto px-4 sm:px-6 lg:px-8`.
- Mobile-first. Single breakpoint `md:` (768px). Mobile detected via `window.innerWidth < 768` + `resize` listener cleaned in `onBeforeUnmount`.
- Mobile uses **bottom-sheet** pattern; desktop uses **right sidebar**.

## Pagination Pattern (9 per page)

Used in Berita, UMKM, PPID; 6 per page in Profil (Perangkat).

```js
const itemsPerPage = 9
const totalPages = computed(() => Math.ceil(filtered.value.length / itemsPerPage))
const startIndex = computed(() => (currentPage.value - 1) * itemsPerPage)
const endIndex   = computed(() => startIndex.value + itemsPerPage)
const paginated  = computed(() => filtered.value.slice(startIndex.value, endIndex.value))
```

## Filter Pattern

```js
const searchQuery = ref('')
const [dateFilter|selectedCategory] = ref('all'|'Semua')
const filtered = computed(() => /* match if query ⊂ field, lowercased */)
const resetFilters = () => { searchQuery.value = ''; ...; currentPage.value = 1 }
```

## Date Format Helper (inlined per page)

```js
const formatDate = (s) => new Date(s).toLocaleDateString('id-ID', {
  day: 'numeric', month: 'short', year: 'numeric'
})
```

## Known Gotchas

- `bg-${color}-100` template strings (`Home.vue`, `Infografik.vue`) get purged by Tailwind. Either safelist in `tailwind.config.js` or hardcode color maps.
- `recharts` dep unused.
- `Infografik.vue` uses real Metabase embeds with JWT tokens (no mock data).
- `public/area-desa-poly.json` is loaded at runtime by `Peta.vue` (polygon is no longer inlined).
- `.env` is committed. `VITE_API_BASE_URL` may omit `/api/v1` — `apiClient.normalizeBaseUrl()` appends it.

## Env

```ini
# .env / .env.local
VITE_API_BASE_URL=http://localhost:8081/api/v1
VITE_METABASE_URL=http://bi-embed.desapalasari.my.id
VITE_FEATURE_GALLERY_PUBLIC=false
```
