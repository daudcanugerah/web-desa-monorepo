<template>
  <div class="h-[calc(100dvh-116px)] flex overflow-hidden relative">
    <div class="flex-1 flex flex-col overflow-hidden bg-gray-50">
      <LoadingSpinner v-if="loading" class="flex-1 flex items-center justify-center" />

      <template v-else>
        <!-- Center header -->
        <div class="flex-shrink-0 px-4 md:px-6 py-3 bg-white border-b border-gray-200">
          <div class="max-w-4xl mx-auto flex items-center gap-3">
            <div class="min-w-0">
              <h1 class="text-sm font-semibold text-gray-900 truncate">
                {{ selectedSection || `Profil ${desaName}` }}
              </h1>
              <p class="text-[11px] text-gray-500 truncate">
                {{ activeGroupName || 'Informasi profil desa' }}
              </p>
            </div>
            <button
              @click="showSidebar = true"
              class="md:hidden ml-auto inline-flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium text-white bg-emerald-600 rounded-lg hover:bg-emerald-700 transition-colors flex-shrink-0"
            >
              <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" /></svg>
              Kategori
            </button>
          </div>
        </div>

        <div class="flex-1 overflow-y-auto p-4 md:p-6">
          <div v-if="selectedSection" class="max-w-4xl mx-auto">
            <!-- Perangkat Desa (officials) -->
            <div v-if="isPerangkat" class="bg-white rounded-lg shadow-sm border border-gray-200 p-4 md:p-6">
              <div class="flex items-center justify-between mb-4">
                <h2 class="text-base md:text-lg font-bold text-gray-900">Perangkat Desa</h2>
                <span class="text-[11px] text-gray-500 bg-gray-100 px-2 py-0.5 rounded-full">{{ officials.length }} orang</span>
              </div>
              <div v-if="paginatedOfficials.length > 0" class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
                <div v-for="official in paginatedOfficials" :key="official.id" class="bg-gray-50 rounded-lg p-4 text-center hover:bg-gray-100 transition-colors">
                  <div class="w-16 h-16 mx-auto mb-3 bg-gray-200 rounded-full overflow-hidden flex items-center justify-center">
                    <img
                      v-if="officialPhoto(official)"
                      :src="officialPhoto(official)"
                      :alt="official.name"
                      loading="lazy"
                      class="w-full h-full object-cover"
                      @error="handleImageError"
                    />
                    <svg v-else class="w-8 h-8 text-gray-400" fill="currentColor" viewBox="0 0 20 20">
                      <path fill-rule="evenodd" d="M10 9a3 3 0 100-6 3 3 0 000 6zm-7 9a7 7 0 1114 0H3z" clip-rule="evenodd" />
                    </svg>
                  </div>
                  <h3 class="text-sm font-bold text-gray-900 mb-0.5">{{ official.name }}</h3>
                  <p class="text-xs text-emerald-600 font-medium mb-1.5">{{ official.position }}</p>
                  <p v-if="official.phone" class="text-[11px] text-gray-500">{{ official.phone }}</p>
                  <p v-if="official.description" class="text-[11px] text-gray-500 mt-1.5 line-clamp-2">{{ official.description }}</p>
                </div>
              </div>
              <EmptyState v-else message="Data perangkat desa tidak tersedia" />

              <Pagination
                v-if="totalPages > 1"
                v-model:current-page="currentPage"
                :total-pages="totalPages"
                class="mt-5"
              />
            </div>

            <!-- Misi (numbered list) -->
            <div v-else-if="isMisi" class="bg-white rounded-lg shadow-sm border border-gray-200 p-4 md:p-6">
              <div class="flex items-center justify-between mb-3">
                <h2 class="text-base md:text-lg font-bold text-gray-900">{{ selectedSection }}</h2>
                <span class="text-[11px] text-gray-500 bg-gray-100 px-2 py-0.5 rounded-full">{{ activeSections.length }} poin</span>
              </div>
              <ol v-if="activeSections.length" class="list-decimal list-inside space-y-2.5 text-sm text-gray-700">
                <li v-for="m in activeSections" :key="m.id" class="leading-relaxed whitespace-pre-line">{{ stripHtml(m.content) }}</li>
              </ol>
              <EmptyState v-else message="Data misi tidak tersedia" />
            </div>

            <!-- Generic rich-text section -->
            <div v-else class="bg-white rounded-lg shadow-sm border border-gray-200 p-4 md:p-6">
              <div class="flex items-center justify-between mb-4">
                <h2 class="text-base md:text-lg font-bold text-gray-900">{{ selectedSection }}</h2>
              </div>
              <div v-if="activeSections.length" class="space-y-4">
                <div
                  v-for="s in activeSections"
                  :key="s.id"
                  class="text-sm text-gray-700 leading-relaxed whitespace-pre-line"
                  v-html="sanitize(s.content)"
                ></div>
              </div>
              <EmptyState v-else message="Data belum tersedia" />
            </div>
          </div>

          <div v-else class="flex items-center justify-center h-full">
            <div class="text-center max-w-md">
              <svg class="w-12 h-12 mx-auto mb-3 text-emerald-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              <h2 class="text-base font-bold text-gray-900 mb-1.5">Profil {{ desaName }}</h2>
              <p class="text-xs text-gray-600">Pilih salah satu kategori di sidebar untuk melihat informasinya.</p>
            </div>
          </div>
        </div>
      </template>
    </div>

    <!-- Desktop sidebar -->
    <div class="hidden md:flex w-80 bg-white border-l border-gray-200 flex-col overflow-hidden z-[50]">
      <div class="flex-shrink-0 px-4 py-3 border-b border-gray-200">
        <h2 class="text-sm font-semibold text-gray-900">Profil Desa</h2>
        <p class="text-[11px] text-gray-500 mt-0.5">Informasi {{ desaName }}</p>
      </div>

      <div class="flex-shrink-0 px-4 py-3 border-b border-gray-100">
        <SearchInput v-model="searchText" placeholder="Cari kategori..." size="sm" />
      </div>

      <div class="flex-1 overflow-y-auto">
        <div v-if="filteredGroups.length === 0" class="p-6 text-center text-xs text-gray-500">
          Tidak ada kategori
        </div>
        <div v-else class="px-3 py-3 space-y-3">
          <div v-for="group in filteredGroups" :key="group.name">
            <p class="px-1 mb-1.5 text-[10px] font-semibold uppercase tracking-wide text-gray-400">{{ group.name }}</p>
            <div class="space-y-1">
              <button
                v-for="item in group.items"
                :key="item.name"
                @click="selectSection(item.name)"
                :class="[
                  'w-full px-3 py-2 rounded-lg text-left transition-all border flex items-center gap-2',
                  selectedSection === item.name
                    ? 'bg-emerald-50 border-emerald-500'
                    : 'bg-gray-50 border-transparent hover:bg-gray-100'
                ]"
              >
                <span
                  class="w-1.5 h-1.5 rounded-full flex-shrink-0"
                  :class="selectedSection === item.name ? 'bg-emerald-500' : 'bg-gray-300'"
                ></span>
                <span class="text-xs font-medium text-gray-900 truncate flex-1">{{ item.name }}</span>
                <span v-if="item.count > 1" class="text-[10px] text-gray-400">{{ item.count }}</span>
              </button>
            </div>
          </div>
        </div>
      </div>

      <div class="flex-shrink-0 px-4 py-2.5 border-t border-gray-100 bg-gray-50">
        <p class="text-[11px] text-gray-500 text-center">{{ sectionCount }} kategori</p>
      </div>
    </div>

    <!-- Mobile sheet -->
    <Transition name="slide-up">
      <div v-if="showSidebar && isMobile" class="absolute inset-x-0 bottom-0 z-[600] bg-white rounded-t-2xl shadow-xl max-h-[70vh] flex flex-col md:hidden">
        <div class="px-4 py-3 border-b border-gray-200 flex items-center justify-between">
          <h2 class="text-sm font-semibold text-gray-900">Profil Desa</h2>
          <button @click="showSidebar = false" class="p-1.5 hover:bg-gray-100 rounded-full">
            <svg class="w-4 h-4 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
        <div class="px-4 py-3 border-b border-gray-100">
          <SearchInput v-model="searchText" placeholder="Cari kategori..." size="sm" />
        </div>
        <div class="px-3 py-3 space-y-3 overflow-y-auto">
          <div v-for="group in filteredGroups" :key="group.name">
            <p class="px-1 mb-1.5 text-[10px] font-semibold uppercase tracking-wide text-gray-400">{{ group.name }}</p>
            <div class="space-y-1">
              <button
                v-for="item in group.items"
                :key="item.name"
                @click="selectSection(item.name); showSidebar = false"
                :class="[
                  'w-full px-3 py-2 rounded-lg text-left transition-all border flex items-center gap-2',
                  selectedSection === item.name
                    ? 'bg-emerald-50 border-emerald-500'
                    : 'bg-gray-50 border-transparent hover:bg-gray-100'
                ]"
              >
                <span
                  class="w-1.5 h-1.5 rounded-full flex-shrink-0"
                  :class="selectedSection === item.name ? 'bg-emerald-500' : 'bg-gray-300'"
                ></span>
                <span class="text-xs font-medium text-gray-900 truncate flex-1">{{ item.name }}</span>
                <span v-if="item.count > 1" class="text-[10px] text-gray-400">{{ item.count }}</span>
              </button>
            </div>
          </div>
        </div>
      </div>
    </Transition>
  </div>
</template>

<script>
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { getPublicProfileAll, getPublicProfileCategories, getPublicStrukturAll, mediaUrl } from '../services/desaService'
import LoadingSpinner from '../components/common/LoadingSpinner.vue'
import EmptyState from '../components/common/EmptyState.vue'
import SearchInput from '../components/common/SearchInput.vue'
import Pagination from '../components/common/Pagination.vue'
import { useDesaInfo } from '../composables/useDesaInfo'

const PERANGKAT = 'Perangkat'
// Sections intentionally hidden from the public nav.
const EXCLUDED_SECTIONS = ['Struktur Organisasi']

const ALLOWED_TAGS = new Set(['p', 'br', 'strong', 'em', 'u', 'ul', 'ol', 'li', 'a', 'h1', 'h2', 'h3', 'h4', 'blockquote', 'code', 'pre'])
const ALLOWED_ATTRS = new Set(['href', 'title', 'target', 'rel'])

const sanitizeHtml = (html) => {
  if (!html || typeof html !== 'string') return ''
  if (!/<[a-z][\s\S]*>/i.test(html)) return html
  const doc = new DOMParser().parseFromString(html, 'text/html')
  const walk = (node) => {
    for (const child of Array.from(node.childNodes)) {
      if (child.nodeType === 1) {
        if (!ALLOWED_TAGS.has(child.tagName.toLowerCase())) {
          child.replaceWith(...Array.from(child.childNodes))
          continue
        }
        for (const attr of Array.from(child.attributes)) {
          if (!ALLOWED_ATTRS.has(attr.name) || (attr.name === 'href' && !/^(https?:|mailto:|#|\/)/i.test(attr.value))) {
            child.removeAttribute(attr.name)
          }
        }
        walk(child)
      } else if (child.nodeType !== 3) {
        child.remove()
      }
    }
  }
  walk(doc.body)
  return doc.body.innerHTML
}

const stripHtml = (html) => {
  if (!html) return ''
  const tmp = document.createElement('div')
  tmp.innerHTML = html
  return tmp.textContent || tmp.innerText || ''
}

export default {
  name: 'Profil',
  components: {
    LoadingSpinner,
    EmptyState,
    SearchInput,
    Pagination
  },
  setup() {
    const loading = ref(true)
    const selectedSection = ref('')
    const showSidebar = ref(false)
    const isMobile = ref(false)
    const searchText = ref('')
    const currentPage = ref(1)
    const itemsPerPage = 6

    const { desaInfo } = useDesaInfo()
    const desaName = computed(() => desaInfo.value?.name || 'Desa')

    const profiles = ref([])
    const officials = ref([])
    const apiCategories = ref([])

    const isPerangkat = computed(() => selectedSection.value === PERANGKAT)
    const isMisi = computed(() => selectedSection.value === 'Misi')

    const activeSections = computed(() =>
      profiles.value.filter(p => p.state !== false && p.section_name === selectedSection.value)
    )

    // Distinct, non-excluded profile section names from the API.
    const apiSectionNames = computed(() => {
      const set = new Set()
      for (const p of profiles.value) {
        if (p.state === false || !p.section_name) continue
        if (EXCLUDED_SECTIONS.includes(p.section_name)) continue
        set.add(p.section_name)
      }
      return [...set]
    })

    const sectionCountFor = (name) =>
      name === PERANGKAT
        ? officials.value.length
        : profiles.value.filter(p => p.state !== false && p.section_name === name).length

    // Category display order comes from the API (`sort_order` ascending).
    // Unknown group names sort after known ones, ordered by the API category
    // list when available, then alphabetically.
    const categoryOrder = computed(() => {
      const list = [...apiCategories.value]
      list.sort((a, b) => (a.sort_order ?? 0) - (b.sort_order ?? 0) || a.name.localeCompare(b.name))
      return list
    })

    const categoryRank = computed(() => {
      const rank = new Map()
      categoryOrder.value.forEach((c, i) => rank.set(c.name, i))
      return rank
    })

    // Group label for a section: the category the API attached to the row.
    // No frontend inference — when a section has no category it stands alone
    // (grouped under its own name) instead of being merged into a synthetic
    // group. Visi and Misi therefore remain independent use cases.
    const categoryNameFor = (sectionName) => {
      const row = profiles.value.find(
        p => p.state !== false && p.section_name === sectionName && p.category
      )
      const cat = row?.category
      const name = typeof cat === 'object' && cat ? cat.name : cat
      return name || sectionName
    }

    const groupRank = (name) => {
      const i = categoryRank.value.get(name)
      return i === undefined ? categoryOrder.value.length : i
    }

    // Grouped nav: API sections + the Perangkat entry, grouped strictly by the
    // API-provided category and ordered by the API `sort_order`.
    // Perangkat is not a profile row — it is backed by the struktur API, so the
    // menu entry only appears when struktur data actually exists.
    const hasPerangkat = computed(() => officials.value.length > 0)

    const navGroups = computed(() => {
      const names = new Set(apiSectionNames.value)
      if (hasPerangkat.value) names.add(PERANGKAT)

      const buckets = new Map()
      for (const name of names) {
        const group = categoryNameFor(name)
        if (!buckets.has(group)) buckets.set(group, [])
        buckets.get(group).push({ name, count: sectionCountFor(name) })
      }

      return [...buckets.entries()]
        .sort((a, b) => groupRank(a[0]) - groupRank(b[0]) || a[0].localeCompare(b[0]))
        .map(([name, items]) => ({
          name,
          items: items.sort((a, b) => a.name.localeCompare(b.name))
        }))
    })

    const filteredGroups = computed(() => {
      const query = searchText.value.trim().toLowerCase()
      if (!query) return navGroups.value
      return navGroups.value
        .map(g => ({
          ...g,
          items: g.items.filter(i =>
            i.name.toLowerCase().includes(query) || g.name.toLowerCase().includes(query)
          )
        }))
        .filter(g => g.items.length > 0)
    })

    const sectionCount = computed(() =>
      navGroups.value.reduce((n, g) => n + g.items.length, 0)
    )

    const activeGroupName = computed(() => {
      for (const g of navGroups.value) {
        if (g.items.some(i => i.name === selectedSection.value)) return g.name
      }
      return ''
    })

    const totalPages = computed(() => Math.ceil(officials.value.length / itemsPerPage))
    const startIndex = computed(() => (currentPage.value - 1) * itemsPerPage)
    const paginatedOfficials = computed(() =>
      officials.value.slice(startIndex.value, startIndex.value + itemsPerPage)
    )

    const selectSection = (name) => {
      selectedSection.value = name
      currentPage.value = 1
    }

    const checkMobile = () => {
      isMobile.value = window.innerWidth < 768
    }

    const handleImageError = (event) => { event.target.style.display = 'none' }
    const officialPhoto = (item) => mediaUrl(item, 'single')

    onMounted(async () => {
      checkMobile()
      window.addEventListener('resize', checkMobile)

      try {
        const [profileResult, strukturResult, categoryResult] = await Promise.all([
          getPublicProfileAll(),
          getPublicStrukturAll().catch(() => []),
          getPublicProfileCategories().catch(() => [])
        ])
        profiles.value = Array.isArray(profileResult) ? profileResult : []
        officials.value = Array.isArray(strukturResult) ? strukturResult : []
        apiCategories.value = Array.isArray(categoryResult) ? categoryResult : []
      } catch (error) {
        console.error('Error fetching profil data:', error)
        profiles.value = []
        officials.value = []
      } finally {
        loading.value = false
      }

      // Default to the first available nav entry.
      const first = navGroups.value[0]?.items[0]?.name
      if (first) selectedSection.value = first
    })

    onBeforeUnmount(() => {
      window.removeEventListener('resize', checkMobile)
    })

    return {
      loading,
      desaName,
      searchText,
      selectedSection,
      showSidebar,
      isMobile,
      currentPage,
      totalPages,
      officials,
      paginatedOfficials,
      navGroups,
      filteredGroups,
      sectionCount,
      activeGroupName,
      activeSections,
      isPerangkat,
      isMisi,
      selectSection,
      handleImageError,
      officialPhoto,
      sanitize: sanitizeHtml,
      stripHtml
    }
  }
}
</script>

<style>
.slide-up-enter-active,
.slide-up-leave-active {
  transition: transform 0.3s ease-in-out;
}
.slide-up-enter-from,
.slide-up-leave-to {
  transform: translateY(100%);
}
</style>