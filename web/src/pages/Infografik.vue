<template>
  <div class="h-[calc(100dvh-116px)] flex overflow-hidden relative">
    <div class="flex-1 flex flex-col overflow-hidden bg-gray-50">
      <LoadingSpinner v-if="loading" class="flex-1 flex items-center justify-center" />

      <template v-else>
        <div
          v-if="!metabaseConfigured"
          class="m-4 md:m-6 bg-amber-50 border border-amber-200 text-amber-800 px-4 py-3 rounded-lg text-sm"
        >
          <strong>Konfigurasi belum lengkap.</strong> Set <code class="bg-amber-100 px-1.5 py-0.5 rounded">VITE_METABASE_URL</code> di <code class="bg-amber-100 px-1.5 py-0.5 rounded">.env</code> untuk mengaktifkan embed.
        </div>

        <div
          v-if="rateLimited"
          class="mx-4 md:mx-6 mt-4 bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg text-sm"
        >
          Batas permintaan token tercapai (maks 30 per komponen per jam per IP). Coba lagi nanti.
        </div>

        <div class="md:hidden px-3 pt-3">
          <div class="flex gap-1.5 overflow-x-auto pb-1 -mx-3 px-3">
            <button
              v-for="sec in filteredSections"
              :key="sec"
              @click="openSection(sec)"
              :class="[
                'flex-shrink-0 px-3 py-1.5 rounded-full text-xs font-medium transition-colors',
                selectedSection === sec
                  ? 'bg-emerald-600 text-white'
                  : 'bg-white text-gray-700 border border-gray-200'
              ]"
            >
              {{ sec }}
            </button>
          </div>
        </div>

        <template v-if="selectedSection && activeSectionComponents.length > 0">
          <div class="px-3 md:px-4 pt-3">
            <div class="bg-white rounded-t-lg border border-b-0 border-gray-200 px-3 py-2.5 flex flex-wrap items-center justify-between gap-2">
              <div class="flex-1 min-w-0">
                <h2 class="text-sm md:text-base font-bold text-gray-900 truncate">{{ selectedSection }}</h2>
                <p v-if="activeComponent" class="text-[11px] text-gray-500 capitalize">
                  {{ activeComponent.component_type }} · id #{{ activeComponent.component_id }}
                  <span v-if="tokenExpiry"> · token {{ formatRelative(tokenExpiry.expiry) }}</span>
                </p>
              </div>
              <button
                v-if="activeComponent"
                @click="refreshToken"
                :disabled="tokenLoading"
                class="text-[11px] text-emerald-600 hover:text-emerald-700 font-medium px-2.5 py-1 rounded border border-emerald-200 hover:border-emerald-400 disabled:opacity-50 flex-shrink-0"
                title="Refresh Metabase token (5 min expiry)"
              >
                <span v-if="tokenLoading">Refreshing…</span>
                <span v-else>Refresh</span>
              </button>
            </div>

            <div
              v-if="activeSectionComponents.length > 1"
              class="bg-white border-x border-gray-200 px-3 py-1.5 overflow-x-auto"
            >
              <div class="flex gap-1.5 min-w-max">
                <button
                  v-for="(comp, idx) in activeSectionComponents"
                  :key="comp.id"
                  @click="switchToIndex(idx)"
                  :class="[
                    'px-2.5 py-1 text-[11px] rounded-full transition-colors',
                    activeComponentIndex === idx
                      ? 'bg-emerald-600 text-white'
                      : 'bg-white text-gray-700 border border-gray-200 hover:bg-gray-100'
                  ]"
                  :disabled="comp.state === false"
                >
                  {{ comp.section_endpoint || 'Komponen #' + comp.component_id }}
                  <span v-if="comp.state === false" class="ml-1 text-red-300">(off)</span>
                </button>
              </div>
            </div>
          </div>

          <div class="flex-1 relative overflow-hidden bg-white border border-gray-200 rounded-b-lg mx-3 md:mx-4 mb-3 md:mb-4">
            <LoadingSpinner v-if="iframeLoading" class="absolute inset-0 m-auto" />
            <div v-if="iframeError" class="flex items-center justify-center h-full text-red-600 text-sm p-8 text-center">
              {{ iframeError }}
            </div>
            <iframe
              v-if="iframeSrc"
              :src="iframeSrc"
              :title="activeComponent?.section_name || selectedSection"
              sandbox="allow-scripts allow-same-origin"
              class="w-full h-full"
              @load="onIframeLoad"
              @error="onIframeError"
            />
            <div v-else-if="activeComponent && activeComponent.state === false" class="flex items-center justify-center h-full text-gray-500 text-sm p-8 text-center">
              Komponen ini dinonaktifkan.
            </div>
            <div v-else class="flex items-center justify-center h-full text-gray-500 text-sm p-8 text-center">
              Token belum tersedia.
            </div>
          </div>
        </template>

        <div v-else-if="!loading" class="flex-1 flex items-center justify-center px-6">
          <div class="text-center max-w-md">
            <svg class="w-12 h-12 mx-auto mb-3 text-emerald-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 19V6l12-3v13M9 19c0 1.105-1.343 2-3 2s-3-.895-3-2 1.343-2 3-2 3 .895 3 2zm12-3c0 1.105-1.343 2-3 2s-3-.895-3-2 1.343-2 3-2 3 .895 3 2zM9 10l12-3" />
            </svg>
            <h2 class="text-base font-bold text-gray-900 mb-1.5">Pilih Bagian Infografik</h2>
            <p class="text-xs text-gray-600">
              <span v-if="sections.length === 0">Belum ada bagian infografik yang tersedia.</span>
              <span v-else>Pilih salah satu bagian di sidebar untuk membuka dashboard Metabase interaktif.</span>
            </p>
          </div>
        </div>
      </template>
    </div>

    <div class="hidden md:flex w-80 bg-white border-l border-gray-200 flex-col overflow-hidden z-[50]">
      <div class="flex-shrink-0 px-4 py-3 border-b border-gray-200">
        <h2 class="text-sm font-semibold text-gray-900">Infografik</h2>
        <p class="text-[11px] text-gray-500 mt-0.5">Pilih bagian statistik {{ desaName }}</p>
      </div>

      <div class="flex-shrink-0 px-4 py-3 border-b border-gray-100 space-y-2">
        <div class="relative">
          <svg class="absolute left-2.5 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-gray-400 pointer-events-none" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
          </svg>
          <input
            v-model="searchText"
            type="text"
            placeholder="Cari bagian..."
            class="w-full pl-8 pr-3 py-1.5 text-xs border border-gray-200 rounded-lg focus:ring-2 focus:ring-emerald-500 focus:border-emerald-500 bg-gray-50"
          />
        </div>
        <div>
          <label class="block text-[11px] font-medium text-gray-700 mb-1">Kategori</label>
          <div class="relative">
            <select
              v-model="selectedCategory"
              class="w-full px-2.5 py-1.5 text-xs border border-gray-200 rounded-lg focus:ring-2 focus:ring-emerald-500 focus:border-emerald-500 bg-gray-50 appearance-none transition-colors"
            >
              <option v-for="cat in availableCategories" :key="cat" :value="cat">{{ cat }}</option>
            </select>
            <svg class="absolute right-2.5 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-gray-400 pointer-events-none" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
            </svg>
          </div>
        </div>
      </div>

      <div class="flex-1 overflow-y-auto">
        <div class="px-3 py-3 space-y-1.5">
          <button
            v-for="sec in filteredSections"
            :key="sec"
            @click="openSection(sec)"
            :class="[
              'w-full px-3 py-2 rounded-lg text-left transition-all border',
              selectedSection === sec
                ? 'bg-emerald-50 border-emerald-500'
                : 'bg-gray-50 border-transparent hover:bg-gray-100'
            ]"
          >
            <h4 class="text-xs font-medium text-gray-900 truncate">{{ sec }}</h4>
            <p class="text-[11px] text-gray-500">{{ getSectionCount(sec) }} komponen</p>
          </button>
        </div>
      </div>

      <div class="flex-shrink-0 px-4 py-2.5 border-t border-gray-100 bg-gray-50">
        <p class="text-[11px] text-gray-500 text-center">{{ filteredSections.length }} dari {{ sections.length }} bagian</p>
      </div>
    </div>

    <button
      v-if="!loading"
      @click="showSidebar = !showSidebar"
      class="absolute bottom-6 right-4 z-[500] bg-emerald-600 text-white rounded-full shadow-lg p-3 hover:bg-emerald-700 transition-colors md:hidden"
      title="Bagian Infografik"
    >
      <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
      </svg>
    </button>

    <Transition name="slide-up">
      <div v-if="showSidebar && isMobile" class="absolute inset-x-0 bottom-0 z-[600] bg-white rounded-t-2xl shadow-xl max-h-[70vh] flex flex-col md:hidden">
        <div class="px-4 py-3 border-b border-gray-200 flex items-center justify-between">
          <h2 class="text-sm font-semibold text-gray-900">Infografik</h2>
          <button @click="showSidebar = false" class="p-1.5 hover:bg-gray-100 rounded-full">
            <svg class="w-4 h-4 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
        <div class="px-4 py-3 border-b border-gray-100 space-y-2">
          <div class="relative">
            <svg class="absolute left-2.5 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-gray-400 pointer-events-none" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
            </svg>
            <input
              v-model="searchText"
              type="text"
              placeholder="Cari bagian..."
              class="w-full pl-8 pr-3 py-1.5 text-xs border border-gray-200 rounded-lg focus:ring-2 focus:ring-emerald-500 focus:border-emerald-500 bg-gray-50"
            />
          </div>
          <div>
            <label class="block text-[11px] font-medium text-gray-700 mb-1">Kategori</label>
            <select
              v-model="selectedCategory"
              class="w-full px-2.5 py-1.5 text-xs border border-gray-200 rounded-lg focus:ring-2 focus:ring-emerald-500 focus:border-emerald-500 bg-gray-50"
            >
              <option v-for="cat in availableCategories" :key="cat" :value="cat">{{ cat }}</option>
            </select>
          </div>
        </div>
        <div class="px-3 py-3 space-y-1.5 overflow-y-auto">
          <button
            v-for="sec in filteredSections"
            :key="sec"
            @click="openSection(sec); showSidebar = false"
            :class="[
              'w-full px-3 py-2 rounded-lg text-left transition-all border',
              selectedSection === sec
                ? 'bg-emerald-50 border-emerald-500'
                : 'bg-gray-50 border-transparent hover:bg-gray-100'
            ]"
          >
            <h4 class="text-xs font-medium text-gray-900 truncate">{{ sec }}</h4>
            <p class="text-[11px] text-gray-500">{{ getSectionCount(sec) }} komponen</p>
          </button>
        </div>
      </div>
    </Transition>
  </div>
</template>

<script>
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { getPublicInfographicList, getPublicInfographicWithToken, getInfographicEmbedUrl, decodeJwtExp } from '../services/desaService'
import LoadingSpinner from '../components/common/LoadingSpinner.vue'
import { useDesaInfo } from '../composables/useDesaInfo'

const METABASE_URL = import.meta.env.VITE_METABASE_URL || ''

const { desaInfo } = useDesaInfo()

export default {
  name: 'Infografik',
  components: { LoadingSpinner },
  setup() {
    const loading = ref(true)
    const desaName = computed(() => desaInfo.value?.name || 'Desa')
    const components = ref([])
    const searchText = ref('')
    const selectedCategory = ref('Semua')
    const selectedSection = ref('')
    const showSidebar = ref(false)
    const isMobile = ref(false)

    const activeComponentIndex = ref(0)
    const iframeLoading = ref(true)
    const iframeError = ref('')
    const tokenLoading = ref(false)
    const tokenExpiry = ref(null)
    const rateLimited = ref(false)
    const autoRefreshTimer = ref(null)

    const AUTO_REFRESH_MARGIN_MS = 60 * 1000

    const metabaseConfigured = computed(() => Boolean(METABASE_URL))

    const sections = computed(() => {
      const names = new Set()
      for (const c of components.value) {
        if (c.section_name) names.add(c.section_name)
      }
      return Array.from(names).sort()
    })

    // Category labels present in the loaded components (plus "Semua").
    const availableCategories = computed(() => {
      const set = new Set()
      let hasUncategorized = false
      for (const c of components.value) {
        const name = typeof c.category === 'object' && c.category ? c.category.name : c.category
        if (name) set.add(name)
        else hasUncategorized = true
      }
      const list = [...set].sort()
      if (hasUncategorized) list.push('Tanpa Kategori')
      return ['Semua', ...list]
    })

    // A section qualifies when any of its components matches the active
    // category filter; the section list is then narrowed by the search text.
    const filteredSections = computed(() => {
      let list = sections.value
      if (selectedCategory.value !== 'Semua') {
        const matchingSections = new Set()
        for (const c of components.value) {
          const name = typeof c.category === 'object' && c.category ? c.category.name : c.category
          const label = name || 'Tanpa Kategori'
          if (label === selectedCategory.value && c.section_name) matchingSections.add(c.section_name)
        }
        list = list.filter(s => matchingSections.has(s))
      }
      if (searchText.value) {
        const query = searchText.value.toLowerCase()
        list = list.filter(s => s.toLowerCase().includes(query))
      }
      return list
    })

    const activeSectionComponents = computed(() => {
      if (!selectedSection.value) return []
      return components.value.filter(c => c.section_name === selectedSection.value)
    })

    const activeComponent = computed(() => {
      const list = activeSectionComponents.value
      if (list.length === 0) return null
      const idx = activeComponentIndex.value >= list.length ? 0 : activeComponentIndex.value
      return list[idx] || null
    })

    const iframeSrc = computed(() => {
      const comp = activeComponent.value
      if (!comp || !tokenExpiry.value) return null
      return getInfographicEmbedUrl(METABASE_URL, comp, tokenExpiry.value.token)
    })

    const getSectionCount = (name) => {
      return components.value.filter(c => c.section_name === name).length
    }

    const formatRelative = (date) => {
      if (!date || typeof date.getTime !== 'function') return ''
      const diffMs = date.getTime() - Date.now()
      const mins = Math.round(diffMs / 60000)
      if (mins <= 0) return 'kadaluarsa'
      return `refresh dalam ${mins} mnt`
    }

    // If the active selection drops out of the filtered list, reset it.
    watch(filteredSections, (list) => {
      if (selectedSection.value && !list.includes(selectedSection.value)) {
        selectedSection.value = ''
      }
    })

    const checkMobile = () => {
      isMobile.value = window.innerWidth < 768
    }

    const fetchList = async () => {
      loading.value = true
      try {
        const result = await getPublicInfographicList({
          page: 1,
          limit: 50
        })
        components.value = result.infographics
      } catch (error) {
        console.error('Error fetching infographic list:', error)
        components.value = []
      } finally {
        loading.value = false
      }
    }

    const requestEmbedToken = async (componentId) => {
      try {
        return await getPublicInfographicWithToken(componentId)
      } catch (error) {
        if (error?.status === 429) {
          rateLimited.value = true
          return { rateLimited: true }
        }
        if (error?.status === 404) {
          return { notFound: true }
        }
        return null
      }
    }

    const clearAutoRefresh = () => {
      if (autoRefreshTimer.value !== null) {
        clearTimeout(autoRefreshTimer.value)
        autoRefreshTimer.value = null
      }
    }

    const scheduleAutoRefresh = (expiry) => {
      clearAutoRefresh()
      if (!expiry) return
      const delay = Math.max(0, expiry.getTime() - Date.now() - AUTO_REFRESH_MARGIN_MS)
      autoRefreshTimer.value = setTimeout(() => {
        autoRefreshTimer.value = null
        refreshToken({ silent: true })
      }, delay)
    }

    const loadTokenFor = async (idx) => {
      clearAutoRefresh()
      const list = activeSectionComponents.value
      if (list.length === 0) return
      const comp = list[idx] || list[0]
      activeComponentIndex.value = list.indexOf(comp)
      if (comp.state === false) {
        tokenExpiry.value = null
        iframeLoading.value = false
        iframeError.value = ''
        return
      }
      iframeError.value = ''
      iframeLoading.value = true
      tokenExpiry.value = null
      const result = await requestEmbedToken(comp.id)
      if (!result || result.rateLimited || result.notFound) {
        iframeLoading.value = false
        if (result?.notFound) iframeError.value = 'Komponen tidak ditemukan atau dinonaktifkan.'
        else if (result?.rateLimited) iframeError.value = 'Batas permintaan token tercapai. Coba lagi nanti.'
        else iframeError.value = 'Gagal memuat token.'
        return
      }
      const expMs = decodeJwtExp(result.token)
      const expiry = new Date(expMs || Date.now() + 5 * 60 * 1000)
      tokenExpiry.value = { token: result.token, expiry }
      iframeLoading.value = false
      scheduleAutoRefresh(expiry)
    }

    const openSection = async (name) => {
      if (!metabaseConfigured.value) return
      rateLimited.value = false
      selectedSection.value = name
      const list = components.value.filter(c => c.section_name === name)
      const firstActive = list.findIndex(c => c.state !== false)
      await loadTokenFor(firstActive >= 0 ? firstActive : 0)
    }

    const switchToIndex = async (idx) => {
      await loadTokenFor(idx)
    }

    const refreshToken = async ({ silent = false } = {}) => {
      const comp = activeComponent.value
      if (!comp) return
      clearAutoRefresh()
      if (!silent) tokenLoading.value = true
      iframeLoading.value = true
      iframeError.value = ''
      tokenExpiry.value = null
      const result = await requestEmbedToken(comp.id)
      if (!silent) tokenLoading.value = false
      if (!result || result.rateLimited) {
        iframeLoading.value = false
        if (result?.rateLimited) iframeError.value = 'Batas permintaan token tercapai. Coba lagi nanti.'
        else iframeError.value = 'Gagal memuat token.'
        return
      }
      const expMs = decodeJwtExp(result.token)
      const expiry = new Date(expMs || Date.now() + 5 * 60 * 1000)
      tokenExpiry.value = { token: result.token, expiry }
      iframeLoading.value = false
      scheduleAutoRefresh(expiry)
    }

    const onIframeLoad = () => {
      iframeLoading.value = false
    }

    const onIframeError = () => {
      iframeLoading.value = false
      iframeError.value = 'Gagal memuat dashboard. Coba refresh token.'
    }

    onMounted(() => {
      checkMobile()
      window.addEventListener('resize', checkMobile)
      fetchList()
    })

    onUnmounted(() => {
      window.removeEventListener('resize', checkMobile)
      clearAutoRefresh()
    })

    return {
      loading,
      desaName,
      searchText,
      selectedCategory,
      selectedSection,
      showSidebar,
      isMobile,
      sections,
      availableCategories,
      filteredSections,
      activeSectionComponents,
      activeComponent,
      activeComponentIndex,
      metabaseConfigured,
      iframeSrc,
      iframeLoading,
      iframeError,
      tokenLoading,
      tokenExpiry,
      rateLimited,
      formatRelative,
      getSectionCount,
      openSection,
      switchToIndex,
      refreshToken,
      onIframeLoad,
      onIframeError
    }
  }
}
</script>

<style scoped>
.slide-up-enter-active,
.slide-up-leave-active {
  transition: transform 0.3s ease-in-out;
}
.slide-up-enter-from,
.slide-up-leave-to {
  transform: translateY(100%);
}
</style>
