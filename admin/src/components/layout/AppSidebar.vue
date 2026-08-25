<template>
  <!-- Mobile backdrop -->
  <div
    v-if="!uiStore.sidebarCollapsed"
    class="fixed inset-0 z-20 bg-black/50 md:hidden"
    aria-hidden="true"
    @click="uiStore.toggleSidebar"
  />

  <!-- Sidebar -->
  <aside
    class="fixed inset-y-0 left-0 z-30 bg-secondary-800 dark:bg-secondary-900 text-white dark:text-secondary-200 min-h-screen flex flex-col transition-all duration-300 md:static md:translate-x-0"
    :class="[
      uiStore.sidebarCollapsed ? '-translate-x-full' : 'translate-x-0',
      uiStore.sidebarIconOnly ? 'w-16' : 'w-64',
    ]"
  >
    <!-- Logo / Brand + desktop collapse toggle -->
    <div class="flex items-center border-b border-secondary-700 h-14 px-3 shrink-0"
         :class="uiStore.sidebarIconOnly ? 'justify-center' : 'justify-between px-4'">
      <span v-if="!uiStore.sidebarIconOnly" class="text-base font-bold tracking-wide truncate">Desa Admin</span>
      <!-- Desktop collapse button -->
      <button
        class="hidden md:flex items-center justify-center w-8 h-8 rounded-lg text-secondary-400 dark:text-secondary-500 hover:bg-secondary-700 hover:text-white transition-colors focus:outline-none focus:ring-2 focus:ring-primary-500"
        :aria-label="uiStore.sidebarIconOnly ? 'Expand sidebar' : 'Collapse sidebar'"
        @click="uiStore.toggleIconOnly"
      >
        <!-- Chevron left when expanded, right when collapsed -->
        <svg class="w-4 h-4 transition-transform" :class="uiStore.sidebarIconOnly ? 'rotate-180' : ''"
             fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
        </svg>
      </button>
    </div>

    <!-- Nav links -->
    <nav class="flex-1 overflow-y-auto py-3" aria-label="Main navigation">
      <template v-for="section in navSections" :key="section.label ?? 'main'">
        <!-- Section label (hidden in icon-only mode) -->
        <p
          v-if="section.label && !uiStore.sidebarIconOnly"
          class="mt-4 mb-1 px-4 text-xs font-semibold uppercase tracking-wider text-secondary-500 dark:text-secondary-400 select-none"
        >
          {{ section.label }}
        </p>
        <!-- Spacer in icon-only mode between sections -->
        <div v-else-if="section.label && uiStore.sidebarIconOnly" class="mt-3" />

        <ul class="space-y-0.5" :class="uiStore.sidebarIconOnly ? 'px-2' : 'px-3'">
          <li v-for="link in section.links" :key="link.to">
            <!-- Main link -->
            <router-link
              v-if="!link.submenu"
              :to="link.to"
              :title="uiStore.sidebarIconOnly ? link.label : undefined"
              class="flex items-center rounded-lg py-2 text-sm font-medium transition-colors hover:bg-secondary-700 group relative"
              :class="[
                uiStore.sidebarIconOnly ? 'justify-center px-2' : 'gap-3 px-3',
                isActive(link.to)
                  ? 'bg-secondary-700 text-white border-l-4 border-primary-400'
                  : 'text-secondary-300 border-l-4 border-transparent',
              ]"
              @click="closeSidebarOnMobile"
            >
              <span class="text-base shrink-0" aria-hidden="true">{{ link.icon }}</span>
              <span v-if="!uiStore.sidebarIconOnly" class="truncate">{{ link.label }}</span>
            </router-link>

            <!-- Link with submenu -->
            <div v-else>
              <button
                :title="uiStore.sidebarIconOnly ? link.label : undefined"
                class="w-full flex items-center rounded-lg py-2 text-sm font-medium transition-colors hover:bg-secondary-700 group relative"
                :class="[
                  uiStore.sidebarIconOnly ? 'justify-center px-2' : 'gap-3 px-3',
                  hasActiveSubmenu(link.submenu)
                    ? 'bg-secondary-700 text-white border-l-4 border-primary-400'
                    : 'text-secondary-300 border-l-4 border-transparent',
                ]"
                @click="toggleSubmenu(link.to)"
              >
                <span class="text-base shrink-0" aria-hidden="true">{{ link.icon }}</span>
                <span v-if="!uiStore.sidebarIconOnly" class="truncate flex-1 text-left">{{ link.label }}</span>
                <svg
                  v-if="!uiStore.sidebarIconOnly"
                  class="w-4 h-4 transition-transform"
                  :class="expandedMenus.includes(link.to) ? 'rotate-90' : ''"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
                </svg>
              </button>

              <!-- Submenu -->
              <ul
                v-if="!uiStore.sidebarIconOnly && expandedMenus.includes(link.to)"
                class="ml-6 mt-1 space-y-0.5"
              >
                <li v-for="sublink in link.submenu" :key="sublink.to">
                  <router-link
                    :to="sublink.to"
                    class="flex items-center gap-3 rounded-lg py-2 px-3 text-sm font-medium transition-colors hover:bg-secondary-700"
                    :class="
                      isActive(sublink.to)
                        ? 'bg-secondary-600 text-white'
                        : 'text-secondary-400'
                    "
                    @click="closeSidebarOnMobile"
                  >
                    <span class="text-sm shrink-0" aria-hidden="true">{{ sublink.icon }}</span>
                    <span class="truncate">{{ sublink.label }}</span>
                  </router-link>
                </li>
              </ul>
            </div>
          </li>
        </ul>
      </template>
    </nav>
  </aside>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useUiStore } from '../../stores/ui'

const uiStore = useUiStore()
const route = useRoute()
const expandedMenus = ref([])

const navSections = [
  {
    label: null,
    links: [
      { to: '/', label: 'Dashboard', icon: '🏠' },
    ],
  },
  {
    label: 'Manajemen Pengguna',
    links: [
      { to: '/users', label: 'Users', icon: '👥' },
      { to: '/roles', label: 'Roles', icon: '🔑' },
    ],
  },
  {
    label: 'Konten',
    links: [
      { to: '/banners', label: 'Banners', icon: '🖼️' },
      { to: '/berita', label: 'Berita', icon: '📰' },
      { to: '/umkm', label: 'UMKM', icon: '🏪' },
      { to: '/fasilitas', label: 'Fasilitas', icon: '🏛️' },
      { to: '/ppid', label: 'PPID', icon: '📄' },
      { to: '/ppid-requests', label: 'PPID Requests', icon: '📋' },
      { to: '/infographic', label: 'Infographic', icon: '📊' },
      {
        to: '/gallery',
        label: 'Galeri',
        icon: '🖼️',
        submenu: [
          { to: '/gallery', label: 'Folder', icon: '📁' },
          { to: '/gallery/media', label: 'Semua Media', icon: '🖼️' },
        ],
      },
    ],
  },
  {
    label: 'Organisasi',
    links: [
      { to: '/struktur', label: 'Struktur', icon: '🏢' },
      { 
        to: '/profile', 
        label: 'Profil Desa', 
        icon: '🌿',
        submenu: [
          { to: '/profile', label: 'Info Umum', icon: '📋' },
          { to: '/profile/sections', label: 'Informasi Detail', icon: '📑' },
        ]
      },
    ],
  },
]

function isActive(to) {
  if (to === '/') return route.path === '/'
  // Exact match or match with trailing slash
  return route.path === to || route.path === to + '/'
}

function hasActiveSubmenu(submenu) {
  // Check if any submenu item is exactly active
  return submenu.some(sublink => isActive(sublink.to))
}

function toggleSubmenu(menuTo) {
  const index = expandedMenus.value.indexOf(menuTo)
  if (index > -1) {
    expandedMenus.value.splice(index, 1)
  } else {
    expandedMenus.value.push(menuTo)
  }
}

function closeSidebarOnMobile() {
  if (window.innerWidth < 768) {
    uiStore.sidebarCollapsed = true
  }
}

// Auto-expand submenu if current route is in submenu
onMounted(() => {
  navSections.forEach(section => {
    section.links.forEach(link => {
      if (link.submenu && hasActiveSubmenu(link.submenu)) {
        if (!expandedMenus.value.includes(link.to)) {
          expandedMenus.value.push(link.to)
        }
      }
    })
  })
})
</script>
