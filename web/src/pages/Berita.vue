<template>
  <div class="min-h-screen bg-gray-50">
    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
      <PageHeader :title="`Berita ${desaName}`" :subtitle="`Informasi dan berita terkini dari ${desaName}`" />

      <LoadingSpinner v-if="loading" />

      <div v-else-if="data && data.length > 0">
        <div class="bg-white rounded-xl shadow-sm border border-gray-200 p-4 mb-6">
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label class="block text-xs font-medium text-gray-700 mb-1.5">Cari Berita</label>
              <SearchInput
                v-model="searchQuery"
                placeholder="Cari judul atau kategori..."
                @clear="resetFilters"
              />
            </div>
            <div>
              <label class="block text-xs font-medium text-gray-700 mb-1.5">Rentang Tanggal</label>
              <div class="relative">
                <select
                  v-model="dateFilter"
                  class="w-full px-3 py-2 text-sm border border-gray-200 rounded-lg focus:ring-2 focus:ring-emerald-500 focus:border-emerald-500 bg-gray-50 appearance-none transition-colors"
                >
                  <option value="all">Semua Tanggal</option>
                  <option value="last30days">30 Hari Terakhir</option>
                  <option value="last90days">90 Hari Terakhir</option>
                  <option value="thisYear">Tahun Ini</option>
                </select>
                <svg class="absolute right-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400 pointer-events-none" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
                </svg>
              </div>
            </div>
          </div>

          <div class="mt-4 pt-3 border-t border-gray-200 flex items-center justify-between">
            <p class="text-xs text-gray-600">
              Menampilkan <span class="font-bold text-emerald-600">{{ startIndex + 1 }}-{{ Math.min(endIndex, filteredBerita.length) }}</span> dari <span class="font-bold">{{ filteredBerita.length }}</span> berita
              <span v-if="filteredBerita.length !== data.length" class="text-gray-400">({{ data.length }} total)</span>
            </p>
            <button
              v-if="hasActiveFilter"
              @click="resetFilters"
              class="text-xs text-emerald-600 hover:text-emerald-700 font-medium flex items-center gap-1.5"
            >
              <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
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

        <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 mb-8">
          <RouterLink
            v-for="news in paginatedBerita"
            :key="news.id"
            :to="`/berita/${news.id}`"
            class="group focus:outline-none"
          >
            <article class="bg-white rounded-xl overflow-hidden shadow-sm hover:shadow-lg transition-all border border-gray-200 h-full flex flex-col group-hover:-translate-y-1 duration-300">
              <div class="aspect-video bg-gradient-to-br from-emerald-100 to-teal-100 relative overflow-hidden">
                <img
                  v-if="newsMediaUrl(news)"
                  :src="newsMediaUrl(news)"
                  :alt="news.title"
                  loading="lazy"
                  class="absolute inset-0 w-full h-full object-cover transition-transform duration-500 group-hover:scale-105"
                  @error="handleImageError"
                />
              </div>
              <div class="p-6 flex flex-col flex-grow">
                <span class="inline-block bg-emerald-100 text-emerald-800 text-xs px-2.5 py-0.5 rounded-full font-medium mb-3 self-start">
                  {{ news.category || 'Umum' }}
                </span>
                <h2 class="text-lg font-bold text-gray-900 mb-3 line-clamp-2 group-hover:text-emerald-600 transition-colors flex-grow">
                  {{ news.title }}
                </h2>
                <div class="flex justify-between items-center text-xs text-gray-500 mt-auto pt-2 border-t border-gray-100">
                  <span>{{ formatDate(getDate(news)) }}</span>
                  <span class="text-emerald-600 font-medium group-hover:translate-x-1 transition-transform inline-flex items-center">
                    Baca
                    <svg class="w-3.5 h-3.5 ml-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
                    </svg>
                  </span>
                </div>
              </div>
            </article>
          </RouterLink>
        </div>

        <Pagination
          v-if="filteredBerita.length > 0"
          v-model:current-page="currentPage"
          :total-pages="totalPages"
        />
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
import PageHeader from '../components/common/PageHeader.vue'
import Pagination from '../components/common/Pagination.vue'
import { useDesaInfo } from '../composables/useDesaInfo'

export default {
  name: 'Berita',
  components: {
    LoadingSpinner,
    EmptyState,
    SearchInput,
    PageHeader,
    Pagination
  },
  setup() {
    const data = ref(null)
    const loading = ref(true)
    const searchQuery = ref('')
    const dateFilter = ref('all')
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

    const getDate = (berita) => berita?.created_at || null

    const newsMediaUrl = (item) => mediaUrl(item, 'single')

    const handleImageError = (event) => {
      event.target.style.display = 'none'
    }

    const filteredBerita = computed(() => {
      if (!data.value) return []
      let filtered = data.value

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

    const hasActiveFilter = computed(() => Boolean(searchQuery.value) || dateFilter.value !== 'all')

    const resetFilters = () => {
      searchQuery.value = ''
      dateFilter.value = 'all'
      currentPage.value = 1
    }

    watch([searchQuery, dateFilter], () => {
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
      currentPage,
      filteredBerita,
      totalPages,
      startIndex,
      endIndex,
      paginatedBerita,
      hasActiveFilter,
      formatDate,
      getDate,
      handleImageError,
      resetFilters,
      newsMediaUrl
    }
  }
}
</script>