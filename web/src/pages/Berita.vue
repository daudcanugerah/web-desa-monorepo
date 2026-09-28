<template>
  <div class="min-h-screen bg-gray-50">
    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <!-- Header -->
      <div class="mb-5">
        <h1 class="text-xl sm:text-2xl font-bold text-gray-900">Berita {{ desaName }}</h1>
        <p class="text-xs text-gray-500 mt-0.5">Informasi dan berita terkini dari {{ desaName }}</p>
      </div>

      <LoadingSpinner v-if="loading" />

      <div v-else-if="data && data.length > 0">
        <!-- Filter bar -->
        <div class="bg-white rounded-lg shadow-sm border border-gray-200 p-3 mb-5">
          <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
            <div>
              <label class="block text-[11px] font-medium text-gray-700 mb-1">Cari Berita</label>
              <SearchInput
                v-model="searchQuery"
                placeholder="Cari judul atau kategori..."
                size="sm"
                @clear="resetFilters"
              />
            </div>
            <div>
              <label class="block text-[11px] font-medium text-gray-700 mb-1">Kategori</label>
              <div class="relative">
                <select
                  v-model="selectedCategory"
                  class="w-full px-2.5 py-1.5 text-xs border border-gray-200 rounded-lg focus:ring-2 focus:ring-emerald-500 focus:border-emerald-500 bg-gray-50 appearance-none transition-colors"
                >
                  <option v-for="cat in categories" :key="cat" :value="cat">{{ cat }}</option>
                </select>
                <svg class="absolute right-2.5 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-gray-400 pointer-events-none" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
                </svg>
              </div>
            </div>
            <div>
              <label class="block text-[11px] font-medium text-gray-700 mb-1">Rentang Tanggal</label>
              <div class="relative">
                <select
                  v-model="dateFilter"
                  class="w-full px-2.5 py-1.5 text-xs border border-gray-200 rounded-lg focus:ring-2 focus:ring-emerald-500 focus:border-emerald-500 bg-gray-50 appearance-none transition-colors"
                >
                  <option value="all">Semua Tanggal</option>
                  <option value="last30days">30 Hari Terakhir</option>
                  <option value="last90days">90 Hari Terakhir</option>
                  <option value="thisYear">Tahun Ini</option>
                </select>
                <svg class="absolute right-2.5 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-gray-400 pointer-events-none" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
                </svg>
              </div>
            </div>
          </div>

          <div class="mt-3 pt-2.5 border-t border-gray-200 flex items-center justify-between">
            <p class="text-[11px] text-gray-600">
              Menampilkan <span class="font-bold text-emerald-600">{{ startIndex + 1 }}-{{ Math.min(endIndex, filteredBerita.length) }}</span> dari <span class="font-bold">{{ filteredBerita.length }}</span> berita
              <span v-if="filteredBerita.length !== data.length" class="text-gray-400">({{ data.length }} total)</span>
            </p>
            <button
              v-if="hasActiveFilter"
              @click="resetFilters"
              class="text-[11px] text-emerald-600 hover:text-emerald-700 font-medium flex items-center gap-1"
            >
              <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
              Reset Filter
            </button>
          </div>
        </div>

        <div v-if="filteredBerita.length === 0" class="py-4">
          <EmptyState
            message="Tidak ada berita yang cocok"
            description="Coba ubah kata kunci pencarian atau rentang tanggal."
            action-label="Reset Filter"
            @action="resetFilters"
          />
        </div>

        <template v-else>
          <!-- Grid — Google-News-style cards -->
          <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
            <RouterLink
              v-for="news in paginatedBerita"
              :key="news.id"
              :to="`/berita/${news.id}`"
              class="group focus:outline-none"
            >
              <article class="h-full flex flex-col">
                <div class="aspect-[16/10] bg-gray-100 rounded-xl overflow-hidden">
                  <img
                    v-if="newsMediaUrl(news)"
                    :src="newsMediaUrl(news)"
                    :alt="news.title"
                    loading="lazy"
                    class="w-full h-full object-cover transition-transform duration-500 group-hover:scale-105"
                    @error="handleImageError"
                  />
                </div>
                <h2 class="mt-2.5 text-base font-bold text-gray-900 leading-snug line-clamp-2 group-hover:text-emerald-600 transition-colors">
                  {{ news.title }}
                </h2>
                <p class="mt-1.5 text-xs text-gray-500">{{ formatRelative(getDate(news)) }}</p>
              </article>
            </RouterLink>
          </div>

          <Pagination
            v-model:current-page="currentPage"
            :total-pages="totalPages"
          />
        </template>
      </div>

      <EmptyState
        v-else-if="!loading && data && data.length === 0"
        message="Belum ada berita tersedia"
      />
    </div>
  </div>
</template>

<script>
import { ref, onMounted, computed, watch } from 'vue'
import { getPublicBeritaList, mediaUrl } from '../services/desaService'
import LoadingSpinner from '../components/common/LoadingSpinner.vue'
import EmptyState from '../components/common/EmptyState.vue'
import SearchInput from '../components/common/SearchInput.vue'
import Pagination from '../components/common/Pagination.vue'
import { useDesaInfo } from '../composables/useDesaInfo'

export default {
  name: 'Berita',
  components: {
    LoadingSpinner,
    EmptyState,
    SearchInput,
    Pagination
  },
  setup() {
    const data = ref(null)
    const loading = ref(true)
    const searchQuery = ref('')
    const dateFilter = ref('all')
    const selectedCategory = ref('Semua')
    const currentPage = ref(1)
    const itemsPerPage = 9

    const { desaInfo } = useDesaInfo()
    const desaName = computed(() => desaInfo.value?.name || 'Desa')

    const formatDate = (dateString) => {
      if (!dateString) return '-'
      try {
        return new Date(dateString).toLocaleDateString('id-ID', {
          day: 'numeric',
          month: 'short',
          year: 'numeric'
        })
      } catch {
        return '-'
      }
    }

    // Google-News-style relative time: "3 jam lalu", "2 hari lalu", …
    const formatRelative = (dateString) => {
      if (!dateString) return ''
      const then = new Date(dateString).getTime()
      if (Number.isNaN(then)) return ''
      const diffMs = Date.now() - then
      const mins = Math.floor(diffMs / 60000)
      if (mins < 1) return 'Baru saja'
      if (mins < 60) return `${mins} menit lalu`
      const hours = Math.floor(mins / 60)
      if (hours < 24) return `${hours} jam lalu`
      const days = Math.floor(hours / 24)
      if (days < 7) return `${days} hari lalu`
      const weeks = Math.floor(days / 7)
      if (weeks < 5) return `${weeks} minggu lalu`
      return formatDate(dateString)
    }

    const getDate = (berita) => berita?.created_at || null

    const newsMediaUrl = (item) => mediaUrl(item, 'single')

    const handleImageError = (event) => {
      event.target.style.display = 'none'
    }

    const categories = computed(() => {
      if (!data.value) return ['Semua']
      const cats = [...new Set(data.value.map(b => b.category).filter(Boolean))]
      cats.sort((a, b) => a.localeCompare(b))
      return ['Semua', ...cats]
    })

    const filteredBerita = computed(() => {
      if (!data.value) return []
      let filtered = data.value

      if (selectedCategory.value !== 'Semua') {
        filtered = filtered.filter(berita => berita.category === selectedCategory.value)
      }

      if (searchQuery.value) {
        const query = searchQuery.value.toLowerCase().trim()
        if (query) {
          filtered = filtered.filter(berita =>
            berita.title?.toLowerCase().includes(query) ||
            (typeof berita.category === 'string' && berita.category.toLowerCase().includes(query))
          )
        }
      }

      if (dateFilter.value !== 'all') {
        const now = new Date()
        let filterDate = new Date()
        switch (dateFilter.value) {
          case 'last30days': filterDate.setDate(now.getDate() - 30); break
          case 'last90days': filterDate.setDate(now.getDate() - 90); break
          case 'thisYear': filterDate = new Date(now.getFullYear(), 0, 1); break
        }
        filtered = filtered.filter(berita => {
          const date = getDate(berita)
          return date && new Date(date) >= filterDate
        })
      }

      return filtered
    })

    const totalPages = computed(() => Math.max(1, Math.ceil(filteredBerita.value.length / itemsPerPage)))
    const startIndex = computed(() => (currentPage.value - 1) * itemsPerPage)
    const endIndex = computed(() => startIndex.value + itemsPerPage)
    const paginatedBerita = computed(() => filteredBerita.value.slice(startIndex.value, endIndex.value))

    const hasActiveFilter = computed(() =>
      Boolean(searchQuery.value) || dateFilter.value !== 'all' || selectedCategory.value !== 'Semua'
    )

    const resetFilters = () => {
      searchQuery.value = ''
      dateFilter.value = 'all'
      selectedCategory.value = 'Semua'
      currentPage.value = 1
    }

    watch([searchQuery, dateFilter, selectedCategory], () => {
      currentPage.value = 1
    })

    onMounted(async () => {
      try {
        const result = await getPublicBeritaList()
        data.value = result
      } catch (error) {
        console.error('Error fetching berita data:', error)
        data.value = []
      } finally {
        loading.value = false
      }
    })

    return {
      data,
      loading,
      desaName,
      searchQuery,
      dateFilter,
      selectedCategory,
      categories,
      currentPage,
      filteredBerita,
      totalPages,
      startIndex,
      endIndex,
      paginatedBerita,
      hasActiveFilter,
      formatDate,
      formatRelative,
      getDate,
      handleImageError,
      resetFilters,
      newsMediaUrl
    }
  }
}
</script>