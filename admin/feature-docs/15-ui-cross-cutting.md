# 15 — UI / Cross-Cutting

The shell every authenticated page renders into: layout, sidebar, header, breadcrumbs, toasts, confirm dialog, theme, locale, common components.

---

## 1. Purpose

Provide the chrome around every feature page — navigation, identity, notifications, dark mode, language switching, and the shared component library (`AppButton`, `AppModal`, `AppTable`, etc.).

---

## 2. Files

### 2.1 Layout

| File | Role |
|------|------|
| `src/components/layout/AppLayout.vue` | Sidebar + Header + Breadcrumb + `<router-view>`; calls `authStore.fetchCurrentUser()` on mount |
| `src/components/layout/AppSidebar.vue` | Sectioned nav, mobile + icon-only modes, "Profil Desa" submenu |
| `src/components/layout/AppHeader.vue` | Avatar dropdown, ThemeToggle, LanguageSwitcher, hamburger |
| `src/components/layout/AppBreadcrumb.vue` | Builds crumbs from `route.matched[].meta.breadcrumb` |

### 2.2 Common components

| File | Purpose |
|------|---------|
| `src/components/common/AppBackButton.vue` | Back link (uses `router.back()` or `to` prop) |
| `src/components/common/AppBadge.vue` | Pill badge — primary / secondary / success / warning / danger |
| `src/components/common/AppButton.vue` | Variants: primary / secondary / danger / success / outline / ghost |
| `src/components/common/AppIconButton.vue` | Icon-first button, same variants |
| `src/components/common/AppInput.vue` | Labeled text input with error state |
| `src/components/common/AppModal.vue` | `<Teleport>` modal with focus trap, Esc to close, sizes sm/md/lg |
| `src/components/common/AppNotification.vue` | Top-right toast renderer |
| `src/components/common/AppPagination.vue` | Numbered pager (max 5 visible pages) |
| `src/components/common/AppTable.vue` | Generic table with slot cells + loading skeleton |
| `src/components/common/ConfirmDialog.vue` | `AppModal` wrapper driven by `ui` store |
| `src/components/common/CategoryManagerModal.vue` | Reusable modal: list + add + delete categories via injected service (`getCategories` / `createCategory` / `deleteCategory`). Emits `selected` when user picks a category. Used by Berita, UMKM, PPID form views. |
| `src/components/common/LanguageSwitcher.vue` | Dropdown wired to `useLocale` |
| `src/components/common/LoadingSpinner.vue` | Spinner sizes sm / md / lg |
| `src/components/common/ThemeToggle.vue` | Sun/moon icon button → `themeStore.toggle()` |

### 2.3 Stores

| Store | State | Notes |
|-------|-------|-------|
| `auth` | tokens + `currentUser` | See `01-authentication.md` |
| `notification` | `toasts: []` | See `00-architecture.md` §13 |
| `theme` | `dark: boolean` | Persisted in `localStorage.darkMode` |
| `ui` | `sidebarCollapsed`, `sidebarIconOnly`, `confirmDialog` | Persisted: `sidebar_icon_only` |

### 2.4 Composables

| Composable | Returns |
|------------|---------|
| `useConfirm()` | `{ confirm(title, message, itemName) → Promise<boolean> }` |
| `useImagePreview({maxMB,warnMB})` | `{ previewUrl, fileSize, error, warning, selectedFile, processFile, reset }` |
| `useLocale()` | `{ currentLocale (writable), availableLocales, t }` |
| `useMetabase()` | `{ initializeMetabase(), isMetabaseInitialized }` |

### 2.5 Utilities

| File | Exports |
|------|---------|
| `src/utils/imageUrl.js` | `getImageUrl(nameOrUrl)` |
| `src/utils/storage.js` | `getAccessToken`, `getRefreshToken`, `setTokens`, `clearTokens` |
| `src/utils/validators.js` | `isValidImageType`, `isValidDocumentType`, `isWithinSize`, `formatFileSize` |

---

## 3. Sidebar Navigation

```
Main
├── Dashboard               → /
Manajemen Pengguna
├── Pengguna                → /users         (admin only)
├── Peran                   → /roles
Konten
├── Banner                  → /banners
├── Berita                  → /berita
├── UMKM                    → /umkm
├── Fasilitas               → /fasilitas
├── PPID                    → /ppid
├── PPID Requests           → /ppid-requests
├── Infographic             → /infographic
Organisasi
├── Struktur Organisasi     → /struktur
└── Profil Desa             (submenu, auto-expanded on child route)
    ├── Info Umum           → /profile
    └── Informasi Detail    → /profile/sections
```

The avatar dropdown in the header offers:
- **Profil Saya** → `/me`
- **Logout** → `authStore.clearTokens()` + `router.push('/login')`

---

## 4. AppLayout Lifecycle

```vue
<script setup>
onMounted(async () => {
  if (!authStore.currentUser) {
    await authStore.fetchCurrentUser()
  }
})
</script>
```

This guarantees the header avatar / name / role pills are populated on every authenticated route.

---

## 5. Theming

- Tailwind dark-mode strategy: `class`.
- `themeStore.init()` runs in `main.js` **before** mount → no flash of wrong theme.
- `themeStore.dark` is persisted in `localStorage.darkMode` (string `"true"` / `"false"`).
- `applyTheme()` toggles `<html class="dark">`.

---

## 6. Locale (i18n)

- Default: `id`. Fallback: `id`. Supported: `id`, `en`.
- Persistence: `localStorage.locale`.
- Side effect: `document.documentElement.lang` is updated on change.
- **Coverage**: only `RoleListView` uses `$t()` end-to-end. Other views hardcode Indonesian strings. Adding translations to the other views is a low-effort improvement.
- Translation files:
  - `src/i18n/locales/id.json`
  - `src/i18n/locales/en.json`

---

## 7. Toast Notifications

- Top-right stack.
- `success(msg)` → auto-dismiss after 3 s.
- `error(msg)` → sticky (`duration: 0`); user dismisses manually.
- Triggered from views via `notificationStore.success(...)` / `.error(...)`.

---

## 8. Confirm Dialog

Pattern in a view:

```js
import { useConfirm } from '@/composables/useConfirm'
const { confirm } = useConfirm()

const onDelete = async (item) => {
  const ok = await confirm({
    title: 'Hapus Data',
    message: `Yakin ingin menghapus "${item.name}"?`,
  })
  if (!ok) return
  await service.remove(item.id)
}
```

The dialog is rendered once globally by `ConfirmDialog.vue`; views only flip the store.

---

## 9. Pagination Convention

- Services return `{ pagination: { page, limit, total, total_pages } }`.
- `AppPagination` shows up to 5 page numbers + prev/next.
- Most views use `pagination.total_pages`; `ProfileSectionListView` and `InfographicListView` use `Math.ceil(total / limit)` (see `00-architecture.md` §10).

---

## 10. Common Component Usage

### `AppButton`

```vue
<AppButton variant="primary" :loading="saving" @click="save">Simpan</AppButton>
<AppButton variant="danger" size="sm" @click="remove">Hapus</AppButton>
```

### `AppModal`

```vue
<AppModal :open="showModal" title="Edit" size="md" @close="showModal = false">
  <p>Body</p>
  <template #footer>
    <AppButton @click="showModal = false">Tutup</AppButton>
  </template>
</AppModal>
```

### `AppTable`

```vue
<AppTable :columns="columns" :rows="rows" :loading="loading" :empty-message="'Tidak ada data'">
  <template #cell-name="{ row }">{{ row.name }}</template>
  <template #cell-actions="{ row }">
    <AppIconButton icon="edit" @click="edit(row)" />
    <AppIconButton icon="trash" variant="danger" @click="remove(row)" />
  </template>
</AppTable>
```

### `ImageUpload`

```vue
<ImageUpload v-model="form.avatar" :max-mb="2" accept="image/*" />
```

(Internally uses `useImagePreview` for type/size validation + preview URL.)

### `RichTextEditor`

```vue
<RichTextEditor v-model="form.body" />
```

(Internally wires Quill's image/video handlers to `newsService.uploadMedia`.)

### `MapPicker`

```vue
<MapPicker v-model="form.coordinates" />
```

(Internally loads `area-desa-poly.json`, constrains the marker to the polygon, and supports fullscreen mode.)

### `CategoryManagerModal`

```vue
<CategoryManagerModal
  :show="showCategoryModal"
  :service="newsService"
  title="Kelola Kategori Berita"
  @close="showCategoryModal = false"
  @selected="(category) => form.category = { id: category.id, name: category.name }"
/>
```

(Props: `show`, `service` (must expose `getCategories`, `createCategory`, `deleteCategory`), `title`, optional `allowSelect`. Emits: `close`, `selected({id, name})`. Loads list on open; supports free-text add, click-to-select, and per-row delete. Shows a friendly error when the backend returns 409 — "Kategori masih digunakan dan tidak dapat dihapus".)

### Category form binding pattern (UUID-based)

Per the swagger v3 spec, every POST/PUT body that accepts a `category` field expects a UUID string, not the human name. The form views store category as `{ id, name }` so they can both bind `name` to the input and submit `id` to the backend.

```vue
<AppInput
  v-model="categoryNameInput"
  list="my-resource-suggestions"
  @change="onCategoryNameInput"
/>
<datalist id="my-resource-suggestions">
  <option v-for="cat in categorySuggestions" :key="cat.id" :value="cat.name" />
</datalist>

<script setup>
const form = ref({ category: { id: '', name: '' } })
const categoryNameInput = computed({
  get: () => form.value.category.name,
  set: (val) => { form.value.category.name = val },
})

function onCategoryNameInput() {
  const match = categorySuggestions.value.find(
    (c) => c.name?.toLowerCase() === form.value.category.name.trim().toLowerCase()
  )
  form.value.category.id = match?.id || ''
}

async function resolveCategoryId() {
  const name = form.value.category.name.trim()
  if (!name) return ''
  if (form.value.category.id) {
    const match = categorySuggestions.value.find((c) => c.id === form.value.category.id)
    if (match && match.name?.toLowerCase() === name.toLowerCase()) return match.id
  }
  const localMatch = categorySuggestions.value.find(
    (c) => c.name?.toLowerCase() === name.toLowerCase()
  )
  if (localMatch) return localMatch.id
  const res = await myService.createCategory({ name })
  const id = res.data.data?.id || res.data?.id
  if (!id) throw new Error('Backend did not return category id')
  categorySuggestions.value.unshift({ id, name })
  return id
}

// On submit:
const categoryId = await resolveCategoryId()
fd.append('category', categoryId)
```

(Free-text still works: if the user types a name that doesn't match any existing category, `resolveCategoryId()` POSTs a new one and uses the returned UUID. The user never has to leave the form to manage categories.)

### Datalist autocomplete for category fields

```vue
<AppInput v-model="form.category" list="my-resource-suggestions" />
<datalist id="my-resource-suggestions">
  <option v-for="cat in categorySuggestions" :key="cat" :value="cat" />
</datalist>
```

(`categorySuggestions` is populated on mount from the public `getCategories` endpoint. Free text is still allowed — the datalist only suggests.)

---

## 11. Notes & Gotchas

- **`theme.js`** has stray `console.log` calls (debug remnants).
- **Sidebar icon-only state** is persisted (`localStorage.sidebar_icon_only`); sidebar mobile-collapsed state is **not** persisted (always hidden on mobile by default).
- **LanguageSwitcher** is only rendered in the header, which is part of `AppLayout`. Auth views (`/login`, `/forgot-password`, `/reset-password`) do not show the header, so the locale cannot be changed there.
- **Breadcrumbs** rely on every route declaring `meta.breadcrumb` — if a route is added without it, the breadcrumb will be empty for that segment.
- **No global error boundary** — JS errors bubble up to the browser; views handle API errors only.
- **`AppNotification` + `ConfirmDialog`** are mounted at the root in `App.vue` so they persist across route changes.