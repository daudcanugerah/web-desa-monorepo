<template>
  <div class="min-h-screen bg-gray-50">
    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <div v-if="!flags.galleryPublic" class="bg-white rounded-lg shadow-sm border border-gray-200 p-8 text-center max-w-xl mx-auto">
        <Icon name="grid" class="w-12 h-12 mx-auto mb-4 text-emerald-400" />
        <h2 class="text-lg font-bold text-gray-900 mb-2">Galeri sedang dalam pemeliharaan</h2>
        <p class="text-sm text-gray-600">Fitur galeri publik belum tersedia pada backend. Silakan kembali lagi nanti.</p>
      </div>

      <template v-else>
        <!-- Header -->
        <div class="mb-5">
          <h1 class="text-xl sm:text-2xl font-bold text-gray-900">Galeri {{ desaName }}</h1>
          <p class="text-xs text-gray-500 mt-0.5">Dokumentasi kegiatan dan pemandangan {{ desaName }}</p>
        </div>

        <LoadingSpinner v-if="loading" />

        <div v-else-if="allFolders.length > 0">
          <!-- Filter bar -->
          <div class="bg-white rounded-lg shadow-sm border border-gray-200 p-3 mb-5">
            <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
              <div>
                <label class="block text-[11px] font-medium text-gray-700 mb-1">Cari Album</label>
                <SearchInput
                  v-model="searchQuery"
                  placeholder="Nama atau deskripsi album..."
                  size="sm"
                  @clear="resetFilters"
                />
              </div>
              <div class="flex items-end">
                <p class="text-[11px] text-gray-600">
                  Menampilkan <span class="font-bold text-emerald-600">{{ filteredFolders.length }}</span> dari <span class="font-bold">{{ allFolders.length }}</span> album publik
                </p>
              </div>
            </div>

            <div v-if="hasActiveFilter" class="mt-3 pt-2.5 border-t border-gray-200 flex justify-end">
              <button
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

          <div v-if="filteredFolders.length === 0" class="py-4">
            <EmptyState
              message="Album tidak ditemukan"
              description="Coba ubah kata kunci pencarian."
              action-label="Reset Filter"
              @action="resetFilters"
            />
          </div>

          <!-- Grid -->
          <div v-else class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-4 mb-6">
            <RouterLink
              v-for="folder in filteredFolders"
              :key="folder.id"
              :to="`/galeri/${folder.id}`"
              class="group focus:outline-none"
            >
              <article class="h-full flex flex-col bg-white rounded-xl border border-gray-200 overflow-hidden transition-all hover:shadow-md hover:-translate-y-0.5 duration-300">
                <div class="aspect-[16/10] bg-gradient-to-br from-emerald-100 to-teal-100 relative overflow-hidden">
                  <img
                    v-if="folderCoverUrl(folder)"
                    :src="folderCoverUrl(folder)"
                    :alt="folder.name"
                    loading="lazy"
                    class="w-full h-full object-cover transition-transform duration-500 group-hover:scale-105"
                    @error="handleImageError"
                  />
                  <div v-else class="w-full h-full flex items-center justify-center text-emerald-300">
                    <svg class="w-12 h-12" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
                    </svg>
                  </div>
                  <div class="absolute top-2 right-2 bg-black/60 text-white text-[11px] px-2 py-0.5 rounded-full font-medium backdrop-blur-sm">
                    {{ folder.media_count || 0 }} media
                  </div>
                </div>
                <div class="p-3.5 flex flex-col flex-grow">
                  <h2 class="text-sm font-bold text-gray-900 leading-snug mb-1.5 line-clamp-2 group-hover:text-emerald-600 transition-colors">
                    {{ folder.name }}
                  </h2>
                  <p v-if="folder.description" class="text-gray-600 text-xs line-clamp-2 flex-grow">
                    {{ folder.description }}
                  </p>
                  <div class="flex justify-between items-center text-[11px] text-gray-500 mt-2.5 pt-2.5 border-t border-gray-100">
                    <span>{{ formatDate(folder.created_at) }}</span>
                    <span class="text-emerald-600 font-medium group-hover:translate-x-1 transition-transform inline-flex items-center">
                      Buka
                      <svg class="w-3 h-3 ml-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
                      </svg>
                    </span>
                  </div>
                </div>
              </article>
            </RouterLink>
          </div>
        </div>

        <EmptyState
          v-else-if="!loading"
          message="Belum ada album publik"
          description="Album publik akan muncul di sini setelah tersedia."
        />
      </template>
    </div>
  </div>
</template>

<script>
import { ref, computed, onMounted, watch } from 'vue'
import { getPublicGalleryFolders, formatMediaDate, resolveFolderCoverUrl } from '../services/desaService'
import LoadingSpinner from '../components/common/LoadingSpinner.vue'
import EmptyState from '../components/common/EmptyState.vue'
import SearchInput from '../components/common/SearchInput.vue'
import Icon from '../components/common/Icon.vue'
import { useDesaInfo } from '../composables/useDesaInfo'
import { useFeatureFlags } from '../composables/useFeatureFlags'

export default {
  name: 'Galeri',
  components: { LoadingSpinner, EmptyState, SearchInput, Icon },
  setup() {
    const allFolders = ref([])
    const loading = ref(true)
    const searchQuery = ref('')

    const { desaInfo } = useDesaInfo()
    const flags = useFeatureFlags()
    const desaName = computed(() => desaInfo.value?.name || 'Desa')

    const formatDate = formatMediaDate

    const folderCoverUrl = (folder) => resolveFolderCoverUrl(folder)

    const handleImageError = (event) => {
      event.target.style.display = 'none'
    }

    const filteredFolders = computed(() => {
      if (!searchQuery.value) return allFolders.value
      const q = searchQuery.value.toLowerCase().trim()
      return allFolders.value.filter(folder =>
        folder.name?.toLowerCase().includes(q) ||
        folder.description?.toLowerCase().includes(q)
      )
    })

    const hasActiveFilter = computed(() => Boolean(searchQuery.value))

    const resetFilters = () => {
      searchQuery.value = ''
    }

    const fetchFolders = async () => {
      loading.value = true
      try {
        const result = await getPublicGalleryFolders({ limit: 100 })
        allFolders.value = result.folders || []
      } catch (error) {
        console.error('Error fetching gallery folders:', error)
        allFolders.value = []
      } finally {
        loading.value = false
      }
    }

    onMounted(fetchFolders)

    return {
      allFolders,
      filteredFolders,
      loading,
      desaName,
      searchQuery,
      hasActiveFilter,
      formatDate,
      handleImageError,
      resetFilters,
      folderCoverUrl,
      flags
    }
  }
}
</script>