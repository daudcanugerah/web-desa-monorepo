<template>
  <div class="min-h-screen bg-gray-50">
    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <!-- Header -->
      <div class="mb-5">
        <h1 class="text-xl sm:text-2xl font-bold text-gray-900">UMKM {{ desaName }}</h1>
        <p class="text-xs text-gray-500 mt-0.5">Dukung usaha mikro kecil menengah di desa kita</p>
      </div>

      <LoadingSpinner v-if="loading" />

      <div v-else-if="data && data.length > 0">
        <!-- Filter bar -->
        <div class="bg-white rounded-lg shadow-sm border border-gray-200 p-3 mb-5">
          <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
            <div>
              <label class="block text-[11px] font-medium text-gray-700 mb-1">Cari UMKM</label>
              <SearchInput
                v-model="searchQuery"
                placeholder="Nama, pemilik, atau alamat..."
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
                  <option v-for="category in categories" :key="category" :value="category">{{ category }}</option>
                </select>
                <svg class="absolute right-2.5 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-gray-400 pointer-events-none" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
                </svg>
              </div>
            </div>
          </div>

          <div class="mt-3 pt-2.5 border-t border-gray-200 flex items-center justify-between">
            <p class="text-[11px] text-gray-600">
              Menampilkan <span class="font-bold text-emerald-600">{{ startIndex + 1 }}-{{ Math.min(endIndex, filteredUMKM.length) }}</span> dari <span class="font-bold">{{ filteredUMKM.length }}</span> UMKM
              <span v-if="filteredUMKM.length !== data.length" class="text-gray-400">({{ data.length }} total)</span>
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

        <div v-if="filteredUMKM.length === 0" class="py-4">
          <EmptyState
            message="Tidak ada UMKM yang cocok"
            description="Coba ubah kata kunci pencarian atau pilih kategori lain."
            action-label="Reset Filter"
            @action="resetFilters"
          />
        </div>

        <!-- Grid — Google-News-style cards -->
        <div v-else class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
          <article
            v-for="business in paginatedUMKM"
            :key="business.id"
            class="h-full flex flex-col bg-white rounded-xl border border-gray-200 overflow-hidden transition-all hover:shadow-md hover:-translate-y-0.5 duration-300"
          >
            <div class="aspect-[16/10] bg-gray-100 overflow-hidden">
              <img
                v-if="businessMediaUrl(business)"
                :src="businessMediaUrl(business)"
                :alt="business.name"
                loading="lazy"
                class="w-full h-full object-cover transition-transform duration-500 hover:scale-105"
                @error="handleImageError"
              />
              <div v-else class="w-full h-full flex items-center justify-center">
                <svg class="w-12 h-12 text-emerald-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
                </svg>
              </div>
            </div>
            <div class="p-3.5 flex flex-col flex-grow">
              <span class="inline-block bg-emerald-100 text-emerald-800 text-[11px] px-2 py-0.5 rounded-full font-medium mb-2 self-start">
                {{ business.category || 'Umum' }}
              </span>
              <h2 class="text-base font-bold text-gray-900 leading-snug mb-1.5 line-clamp-2">{{ business.name }}</h2>
              <p class="text-gray-600 text-xs mb-3 line-clamp-2 flex-grow">{{ business.description }}</p>

              <div class="space-y-1.5 pt-2.5 border-t border-gray-100">
                <div v-if="business.owner" class="flex items-center text-[11px] text-gray-700">
                  <Icon name="user" class="text-gray-400 mr-1.5" />
                  <span class="truncate">{{ business.owner }}</span>
                </div>
                <div v-if="business.phone" class="flex items-center text-[11px] text-gray-700">
                  <Icon name="phone" class="text-gray-400 mr-1.5" />
                  <a :href="`tel:${business.phone}`" class="hover:text-emerald-600 truncate">{{ business.phone }}</a>
                </div>
                <div v-if="business.email" class="flex items-center text-[11px] text-gray-700">
                  <Icon name="mail" class="text-gray-400 mr-1.5" />
                  <a :href="`mailto:${business.email}`" class="hover:text-emerald-600 truncate">{{ business.email }}</a>
                </div>
                <div v-if="business.website" class="flex items-center text-[11px] text-gray-700">
                  <Icon name="globe" class="text-gray-400 mr-1.5" />
                  <a :href="business.website" target="_blank" rel="noopener" class="hover:text-emerald-600 truncate">{{ business.website.replace(/^https?:\/\//, '') }}</a>
                </div>
                <div v-if="business.address" class="flex items-start text-[11px] text-gray-700">
                  <Icon name="pin" class="text-gray-400 mr-1.5 mt-0.5 flex-shrink-0" />
                  <span class="line-clamp-2">{{ business.address }}</span>
                </div>
              </div>
            </div>
          </article>
        </div>

        <Pagination
          v-if="filteredUMKM.length > 0"
          v-model:current-page="currentPage"
          :total-pages="totalPages"
        />
      </div>

      <EmptyState
        v-else-if="!loading && data && data.length === 0"
        message="Belum ada UMKM tersedia"
      />
    </div>
  </div>
</template>

<script>
import { ref, onMounted, computed, watch } from 'vue'
import { getPublicUMKMList, mediaUrl } from '../services/desaService'
import LoadingSpinner from '../components/common/LoadingSpinner.vue'
import EmptyState from '../components/common/EmptyState.vue'
import SearchInput from '../components/common/SearchInput.vue'
import Pagination from '../components/common/Pagination.vue'
import Icon from '../components/common/Icon.vue'
import { useDesaInfo } from '../composables/useDesaInfo'

export default {
  name: 'UMKM',
  components: {
    LoadingSpinner,
    EmptyState,
    SearchInput,
    Pagination,
    Icon
  },
  setup() {
    const data = ref(null)
    const loading = ref(true)
    const searchQuery = ref('')
    const selectedCategory = ref('Semua')
    const currentPage = ref(1)
    const itemsPerPage = 9

    const { desaInfo } = useDesaInfo()
    const desaName = computed(() => desaInfo.value?.name || 'Desa')

    const categories = computed(() => {
      if (!data.value) return ['Semua']
      const cats = [...new Set(data.value.map(u => u.category).filter(Boolean))]
      return ['Semua', ...cats]
    })

    const filteredUMKM = computed(() => {
      if (!data.value) return []
      let filtered = data.value

      if (selectedCategory.value !== 'Semua') {
        filtered = filtered.filter(u => u.category === selectedCategory.value)
      }

      if (searchQuery.value) {
        const query = searchQuery.value.toLowerCase().trim()
        if (query) {
          filtered = filtered.filter(u =>
            u.name?.toLowerCase().includes(query) ||
            u.description?.toLowerCase().includes(query) ||
            (typeof u.category === 'string' && u.category.toLowerCase().includes(query)) ||
            u.owner?.toLowerCase().includes(query) ||
            u.address?.toLowerCase().includes(query)
          )
        }
      }

      return filtered
    })

    const totalPages = computed(() => Math.max(1, Math.ceil(filteredUMKM.value.length / itemsPerPage)))
    const startIndex = computed(() => (currentPage.value - 1) * itemsPerPage)
    const endIndex = computed(() => startIndex.value + itemsPerPage)
    const paginatedUMKM = computed(() => filteredUMKM.value.slice(startIndex.value, endIndex.value))

    const hasActiveFilter = computed(() => Boolean(searchQuery.value) || selectedCategory.value !== 'Semua')

    const resetFilters = () => {
      searchQuery.value = ''
      selectedCategory.value = 'Semua'
      currentPage.value = 1
    }

    const handleImageError = (event) => {
      event.target.style.display = 'none'
    }

    const businessMediaUrl = (item) => mediaUrl(item, 'array')

    watch([searchQuery, selectedCategory], () => {
      currentPage.value = 1
    })

    onMounted(async () => {
      try {
        const result = await getPublicUMKMList()
        data.value = result
      } catch (error) {
        console.error('Error fetching UMKM data:', error)
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
      selectedCategory,
      currentPage,
      categories,
      filteredUMKM,
      totalPages,
      startIndex,
      endIndex,
      paginatedUMKM,
      hasActiveFilter,
      resetFilters,
      handleImageError,
      businessMediaUrl
    }
  }
}
</script>