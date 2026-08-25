<template>
  <nav class="bg-white border-b border-gray-200 sticky top-0 z-50">
    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
      <div class="flex justify-between items-center h-20">
        <RouterLink to="/" class="flex items-center space-x-3 min-w-0" @click="closeMenu">
          <div class="w-12 h-12 bg-emerald-600 rounded flex items-center justify-center flex-shrink-0">
            <span class="text-white font-bold text-lg">{{ brandInitials }}</span>
          </div>
          <div class="flex flex-col min-w-0">
            <span class="font-bold text-lg text-gray-900 leading-tight truncate">{{ brandName }}</span>
            <span class="text-xs text-gray-500 truncate">{{ brandTagline }}</span>
          </div>
        </RouterLink>

        <div class="hidden md:flex space-x-1">
          <RouterLink
            v-for="link in navLinks"
            :key="link.path"
            :to="link.path"
            :class="[
              'px-4 py-2 text-sm font-medium transition-colors',
              isActive(link.path)
                ? 'text-emerald-600 border-b-2 border-emerald-600'
                : 'text-gray-600 hover:text-emerald-600'
            ]"
          >
            {{ link.label }}
          </RouterLink>
        </div>

        <button
          @click="toggleMenu"
          class="md:hidden p-2 rounded-md text-gray-700 hover:bg-gray-100 focus:outline-none"
          aria-label="Toggle menu"
        >
          <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path
              v-if="isMenuOpen"
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M6 18L18 6M6 6l12 12"
            />
            <path
              v-else
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M4 6h16M4 12h16M4 18h16"
            />
          </svg>
        </button>
      </div>

      <div v-if="isMenuOpen" class="md:hidden py-4 border-t border-gray-200 bg-white">
        <div class="flex flex-col space-y-1">
          <RouterLink
            v-for="link in navLinks"
            :key="link.path"
            :to="link.path"
            @click="closeMenu"
            :class="[
              'px-4 py-3 text-sm font-medium transition-colors',
              isActive(link.path)
                ? 'text-emerald-600 bg-emerald-50'
                : 'text-gray-700 hover:bg-gray-50'
            ]"
          >
            {{ link.label }}
          </RouterLink>
        </div>
      </div>
    </div>
  </nav>
</template>

<script>
import { ref, computed } from 'vue'
import { useRoute } from 'vue-router'
import { useDesaInfo } from '../../composables/useDesaInfo'
import { useFeatureFlags } from '../../composables/useFeatureFlags'

export default {
  name: 'Navbar',
  setup() {
    const isMenuOpen = ref(false)
    const route = useRoute()
    const { desaInfo, initials } = useDesaInfo()
    const flags = useFeatureFlags()

    const brandName = computed(() => desaInfo.value?.name || 'Desa')
    const brandInitials = computed(() => initials(desaInfo.value?.name))
    const brandTagline = computed(() => {
      const name = desaInfo.value?.name
      if (!name) return 'Website Resmi'
      return 'Website Resmi'
    })

    const navLinks = computed(() => {
      const links = [
        { path: '/', label: 'Beranda' },
        { path: '/profil', label: 'Profil' },
        { path: '/infografik', label: 'Infografik' },
        { path: '/peta', label: 'Peta' },
        { path: '/berita', label: 'Berita' },
        { path: '/umkm', label: 'UMKM' },
        { path: '/ppid', label: 'PPID' }
      ]
      if (flags.value.galleryPublic) {
        links.push({ path: '/galeri', label: 'Galeri' })
      }
      return links
    })

    const isActive = (path) => {
      if (path === '/') return route.path === '/'
      return route.path.startsWith(path)
    }

    const toggleMenu = () => { isMenuOpen.value = !isMenuOpen.value }
    const closeMenu = () => { isMenuOpen.value = false }

    return {
      isMenuOpen,
      navLinks,
      brandName,
      brandInitials,
      brandTagline,
      isActive,
      toggleMenu,
      closeMenu,
      flags
    }
  }
}
</script>