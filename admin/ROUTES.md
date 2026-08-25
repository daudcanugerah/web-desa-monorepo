# Routes

## Public Routes (No Auth)

| Path | Name | Component |
|------|------|-----------|
| `/login` | Login | `LoginView.vue` |
| `/forgot-password` | ForgotPassword | `ForgotPasswordView.vue` |
| `/reset-password` | ResetPassword | `ResetPasswordView.vue` |
| `/:pathMatch(.*)*` | NotFound | `NotFoundView.vue` |

## Protected Routes (Authenticated)

All routes below require authentication and are wrapped in `AppLayout`.

### Dashboard

| Path | Name | Breadcrumb | Guard |
|------|------|------------|-------|
| `/` | Dashboard | Dashboard | - |

### Users

| Path | Name | Breadcrumb | Guard |
|------|------|------------|-------|
| `/users` | Users | Users | requiresAdmin |
| `/users/create` | UserCreate | Tambah User | requiresAdmin |
| `/users/:id/edit` | UserEdit | Edit User | requiresAdmin |
| `/me` | MyProfile | Profil Saya | - |

### Roles

| Path | Name | Breadcrumb | Guard |
|------|------|------------|-------|
| `/roles` | Roles | Roles | - |

### Banners

| Path | Name | Breadcrumb | Guard |
|------|------|------------|-------|
| `/banners` | Banners | Banners | - |
| `/banners/create` | BannerCreate | Tambah Banner | - |
| `/banners/:id/edit` | BannerEdit | Edit Banner | - |

### Berita (News)

| Path | Name | Breadcrumb | Guard |
|------|------|------------|-------|
| `/berita` | Berita | Berita | - |
| `/berita/create` | BeritaCreate | Tambah Berita | - |
| `/berita/:id/edit` | BeritaEdit | Edit Berita | - |

### UMKM

| Path | Name | Breadcrumb | Guard |
|------|------|------------|-------|
| `/umkm` | UMKM | UMKM | - |
| `/umkm/create` | UmkmCreate | Tambah UMKM | - |
| `/umkm/:id/edit` | UmkmEdit | Edit UMKM | - |

### Fasilitas (Facilities)

| Path | Name | Breadcrumb | Guard |
|------|------|------------|-------|
| `/fasilitas` | Fasilitas | Fasilitas | - |
| `/fasilitas/create` | FasilitasCreate | Tambah Fasilitas | - |
| `/fasilitas/:id/edit` | FasilitasEdit | Edit Fasilitas | - |

### PPID

| Path | Name | Breadcrumb | Guard |
|------|------|------------|-------|
| `/ppid` | PPID | PPID | - |
| `/ppid/create` | PpidCreate | Tambah Dokumen PPID | - |
| `/ppid/:id/edit` | PpidEdit | Edit Dokumen PPID | - |
| `/ppid-requests` | PpidRequests | PPID Requests | - |

### Organization

| Path | Name | Breadcrumb | Guard |
|------|------|------------|-------|
| `/struktur` | Struktur | Struktur Organisasi | - |

### Profile

| Path | Name | Breadcrumb | Guard |
|------|------|------------|-------|
| `/profile` | Profile | Profil Desa | - |
| `/profile/sections` | ProfileSections | Bagian Profil | - |
| `/profile/sections/create` | ProfileSectionCreate | Tambah Bagian Profil | - |
| `/profile/sections/:id/edit` | ProfileSectionEdit | Edit Bagian Profil | - |

### Infographic

| Path | Name | Breadcrumb | Guard |
|------|------|------------|-------|
| `/infographic` | Infographic | Infographic Dashboard | - |
| `/infographic/create` | InfographicCreate | Tambah Dashboard | - |
| `/infographic/:id/edit` | InfographicEdit | Edit Dashboard | - |

## Guard Logic

```
beforeEach:
  1. If route requiresAuth and not authenticated → redirect to /login
  2. If path is /login and already authenticated → redirect to /
  3. If route requiresAdmin:
     a. Fetch currentUser if not loaded
     b. If user lacks 'admin' role → redirect to /
```