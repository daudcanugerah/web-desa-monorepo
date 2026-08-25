<template>
  <div class="min-h-screen bg-gray-50">
    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
      <div v-if="!flags.galleryPublic" class="bg-white rounded-xl shadow-sm border border-gray-200 p-8 text-center max-w-xl mx-auto">
        <Icon name="grid" class="w-12 h-12 mx-auto mb-4 text-emerald-400" />
        <h2 class="text-lg font-bold text-gray-900 mb-2">Galeri sedang dalam pemeliharaan</h2>
        <p class="text-sm text-gray-600">Fitur galeri publik belum tersedia pada backend. Silakan kembali lagi nanti.</p>
      </div>

      <template v-else>
        <PageHeader
          :title="`Galeri ${desaName}`"
          :subtitle="`Dokumentasi kegiatan dan pemandangan ${desaName}`"
        />

      <LoadingSpinner v-if="loading" />

      <template v-else>
        <div class="bg-white rounded-xl shadow-sm border border-gray-200 p-4 mb-6">
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label class="block text-xs font-medium text-gray-700 mb-1.5">Cari Album</label>
              <SearchInput
                v-model="searchQuery"
                placeholder="Cari nama atau deskripsi album..."
                @clear="resetFilters"
              />
            </div>
            <div class="flex items-end">
              <p class="text-xs text-gray-600">
                Menampilkan
                <span class="font-bold text-emerald-600">{{ folders.length }}</span>
                album publik
              </p>
            </div>
          </div>

          <div v-if="hasActiveFilter" class="mt-4 pt-3 border-t border-gray-200 flex justify-end">
            <button
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

        <EmptyState
          v-if="folders.length === 0"
          :message="hasActiveFilter ? 'Album tidak ditemukan' : 'Belum ada album publik'"
          :description="hasActiveFilter ? 'Coba ubah kata kunci pencarian.' : 'Album publik akan muncul di sini setelah tersedia.'"
          :action-label="hasActiveFilter ? 'Reset Filter' : ''"
          @action="resetFilters"
        />

        <div v-else class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
          <RouterLink
            v-for="folder in folders"
            :key="folder.id"
            :to="`/galeri/${folder.id}`"
            class="group focus:outline-none"
          >
            <article class="bg-white rounded-xl overflow-hidden shadow-sm hover:shadow-lg transition-all border border-gray-200 h-full flex flex-col group-hover:-translate-y-1 duration-300">
              <div class="aspect-video bg-gradient-to-br from-emerald-100 to-teal-100 relative overflow-hidden">
                <img
                  v-if="folderCoverUrl(folder)"
                  :src="folderCoverUrl(folder)"
                  :alt="folder.name"
                  loading="lazy"
                  class="absolute inset-0 w-full h-full object-cover transition-transform duration-500 group-hover:scale-105"
                  @error="handleImageError"
                />
                <div v-else class="absolute inset-0 flex items-center justify-center text-emerald-300">
                  <svg class="w-16 h-16" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
                  </svg>
                </div>
                <div class="absolute top-3 right-3 bg-black/60 text-white text-xs px-2.5 py-1 rounded-full font-medium backdrop-blur-sm">
                  {{ folder.public_media_count || folder.media_count || 0 }} media
                </div>
              </div>
              <div class="p-5 flex flex-col flex-grow">
                <h2 class="text-base font-bold text-gray-900 mb-2 line-clamp-2 group-hover:text-emerald-600 transition-colors">
                  {{ folder.name }}
                </h2>
                <p v-if="folder.description" class="text-sm text-gray-600 line-clamp-3 flex-grow">
                  {{ folder.description }}
                </p>
                <div class="flex justify-between items-center text-xs text-gray-500 mt-auto pt-3 border-t border-gray-100">
                  <span>{{ formatDate(folder.created_at) }}</span>
                  <span class="text-emerald-600 font-medium group-hover:translate-x-1 transition-transform inline-flex items-center">
                    Buka
                    <svg class="w-3.5 h-3.5 ml-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
                    </svg>
                  </span>
                </div>
              </div>
            </article>
          </RouterLink>
        </div>
      </template>
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
import PageHeader from '../components/common/PageHeader.vue'
import Icon from '../components/common/Icon.vue'
import { useDesaInfo } from '../composables/useDesaInfo'
import { useFeatureFlags } from '../composables/useFeatureFlags'

export default {
  name: 'Galeri',
  components: { LoadingSpinner, EmptyState, SearchInput, PageHeader, Icon },
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

    const folders = computed(() => {
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
      folders,
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