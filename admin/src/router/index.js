import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import AppLayout from '../components/layout/AppLayout.vue'

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('../views/auth/LoginView.vue'),
  },
  {
    path: '/forgot-password',
    name: 'ForgotPassword',
    component: () => import('../views/auth/ForgotPasswordView.vue'),
  },
  {
    path: '/reset-password',
    name: 'ResetPassword',
    component: () => import('../views/auth/ResetPasswordView.vue'),
  },
  {
    path: '/',
    component: AppLayout,
    meta: { requiresAuth: true },
    children: [
      { path: '', name: 'Dashboard', component: () => import('../views/dashboard/DashboardView.vue'), meta: { breadcrumb: 'Dashboard' } },
      { path: 'users', name: 'Users', component: () => import('../views/users/UserListView.vue'), meta: { breadcrumb: 'Users', requiresAdmin: true } },
      { path: 'users/create', name: 'UserCreate', component: () => import('../views/users/UserFormView.vue'), meta: { breadcrumb: 'Tambah User', requiresAdmin: true } },
      { path: 'users/:id/edit', name: 'UserEdit', component: () => import('../views/users/UserFormView.vue'), meta: { breadcrumb: 'Edit User', requiresAdmin: true } },
      { path: 'roles', name: 'Roles', component: () => import('../views/users/RoleListView.vue'), meta: { breadcrumb: 'Roles' } },
      { path: 'banners', name: 'Banners', component: () => import('../views/content/BannerListView.vue'), meta: { breadcrumb: 'Banners' } },
      { path: 'banners/create', name: 'BannerCreate', component: () => import('../views/content/BannerFormView.vue'), meta: { breadcrumb: 'Tambah Banner' } },
      { path: 'banners/:id/edit', name: 'BannerEdit', component: () => import('../views/content/BannerFormView.vue'), meta: { breadcrumb: 'Edit Banner' } },
      { path: 'berita', name: 'Berita', component: () => import('../views/content/NewsListView.vue'), meta: { breadcrumb: 'Berita' } },
      { path: 'berita/create', name: 'BeritaCreate', component: () => import('../views/content/NewsFormView.vue'), meta: { breadcrumb: 'Tambah Berita' } },
      { path: 'berita/:id/edit', name: 'BeritaEdit', component: () => import('../views/content/NewsFormView.vue'), meta: { breadcrumb: 'Edit Berita' } },
      { path: 'umkm', name: 'UMKM', component: () => import('../views/content/UmkmListView.vue'), meta: { breadcrumb: 'UMKM' } },
      { path: 'umkm/create', name: 'UmkmCreate', component: () => import('../views/content/UmkmFormView.vue'), meta: { breadcrumb: 'Tambah UMKM' } },
      { path: 'umkm/:id/edit', name: 'UmkmEdit', component: () => import('../views/content/UmkmFormView.vue'), meta: { breadcrumb: 'Edit UMKM' } },
      { path: 'fasilitas', name: 'Fasilitas', component: () => import('../views/content/FacilityListView.vue'), meta: { breadcrumb: 'Fasilitas' } },
      { path: 'fasilitas/create', name: 'FasilitasCreate', component: () => import('../views/content/FacilityFormView.vue'), meta: { breadcrumb: 'Tambah Fasilitas' } },
      { path: 'fasilitas/:id/edit', name: 'FasilitasEdit', component: () => import('../views/content/FacilityFormView.vue'), meta: { breadcrumb: 'Edit Fasilitas' } },
      { path: 'ppid', name: 'PPID', component: () => import('../views/content/PpidListView.vue'), meta: { breadcrumb: 'PPID' } },
      { path: 'ppid/create', name: 'PpidCreate', component: () => import('../views/content/PpidFormView.vue'), meta: { breadcrumb: 'Tambah Dokumen PPID' } },
      { path: 'ppid/:id/edit', name: 'PpidEdit', component: () => import('../views/content/PpidFormView.vue'), meta: { breadcrumb: 'Edit Dokumen PPID' } },
      { path: 'ppid-requests', name: 'PpidRequests', component: () => import('../views/content/PpidRequestsView.vue'), meta: { breadcrumb: 'PPID Requests' } },
      { path: 'gallery', name: 'Gallery', component: () => import('../views/content/gallery/GalleryFolderListView.vue'), meta: { breadcrumb: 'Galeri' } },
      { path: 'gallery/folders/create', name: 'GalleryFolderCreate', component: () => import('../views/content/gallery/GalleryFolderFormView.vue'), meta: { breadcrumb: 'Tambah Folder Galeri' } },
      { path: 'gallery/folders/:id', name: 'GalleryFolderDetail', component: () => import('../views/content/gallery/GalleryFolderDetailView.vue'), meta: { breadcrumb: 'Detail Folder' } },
      { path: 'gallery/folders/:id/edit', name: 'GalleryFolderEdit', component: () => import('../views/content/gallery/GalleryFolderFormView.vue'), meta: { breadcrumb: 'Edit Folder Galeri' } },
      { path: 'gallery/media', name: 'GalleryMedia', component: () => import('../views/content/gallery/GalleryMediaListView.vue'), meta: { breadcrumb: 'Semua Media' } },
      { path: 'struktur', name: 'Struktur', component: () => import('../views/organization/StructureView.vue'), meta: { breadcrumb: 'Struktur Organisasi' } },
      { path: 'profile', name: 'Profile', component: () => import('../views/organization/ProfileView.vue'), meta: { breadcrumb: 'Profil Desa' } },
      { path: 'profile/sections', name: 'ProfileSections', component: () => import('../views/organization/ProfileSectionListView.vue'), meta: { breadcrumb: 'Bagian Profil' } },
      { path: 'profile/sections/create', name: 'ProfileSectionCreate', component: () => import('../views/organization/ProfileSectionFormView.vue'), meta: { breadcrumb: 'Tambah Bagian Profil' } },
      { path: 'profile/sections/:id/edit', name: 'ProfileSectionEdit', component: () => import('../views/organization/ProfileSectionFormView.vue'), meta: { breadcrumb: 'Edit Bagian Profil' } },
      { path: 'infographic', name: 'Infographic', component: () => import('../views/content/InfographicListView.vue'), meta: { breadcrumb: 'Infographic Dashboard' } },
      { path: 'infographic/create', name: 'InfographicCreate', component: () => import('../views/content/InfographicFormView.vue'), meta: { breadcrumb: 'Tambah Dashboard' } },
      { path: 'infographic/:id/edit', name: 'InfographicEdit', component: () => import('../views/content/InfographicFormView.vue'), meta: { breadcrumb: 'Edit Dashboard' } },
      { path: 'infographic/access-logs', name: 'InfographicAccessLogs', component: () => import('../views/content/InfographicAccessLogsView.vue'), meta: { breadcrumb: 'Access Log Infographic' } },
      { path: 'me', name: 'MyProfile', component: () => import('../views/users/MyProfileView.vue'), meta: { breadcrumb: 'Profil Saya' } },
    ],
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'NotFound',
    component: () => import('../views/NotFoundView.vue'),
  },
]

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
})

router.beforeEach(async (to) => {
  const authStore = useAuthStore()

  if (to.meta.requiresAuth && !authStore.isAuthenticated) {
    return { path: '/login' }
  }

  if (to.path === '/login' && authStore.isAuthenticated) {
    return { path: '/' }
  }

  if (to.meta.requiresAdmin && authStore.isAuthenticated) {
    // Ensure currentUser is loaded
    if (!authStore.currentUser) {
      try {
        await authStore.fetchCurrentUser()
      } catch {
        return { path: '/' }
      }
    }
    if (!authStore.currentUser?.roles?.includes('admin')) {
      return { path: '/' }
    }
  }
})

export default router
