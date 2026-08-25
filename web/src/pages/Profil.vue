<template>
  <div class="h-[calc(100vh-80px)] flex overflow-hidden relative">
    <div class="flex-1 overflow-y-auto bg-gray-50 p-4 md:p-6">
      <LoadingSpinner v-if="loading" class="flex justify-center py-12" />
      
      <template v-else>
        <div class="md:hidden mb-4 overflow-x-auto -mx-4 px-4">
          <div class="flex gap-2 pb-2">
            <button
              v-for="cat in categories"
              :key="cat.id"
              @click="selectCategory(cat.id)"
              :class="[
                'flex-shrink-0 px-4 py-2 rounded-full text-sm font-medium transition-colors',
                selectedCategory === cat.id
                  ? 'bg-emerald-600 text-white'
                  : 'bg-white text-gray-700 border border-gray-200'
              ]"
            >
              {{ cat.name }}
            </button>
          </div>
        </div>
        
        <div v-if="selectedCategory" class="max-w-4xl mx-auto">
          
          <!-- Sejarah Desa -->
          <div v-if="selectedCategory === 'Sejarah'" class="bg-white rounded-xl shadow-sm border border-gray-200 p-6 md:p-8">
            <div class="flex items-center justify-between mb-6">
              <h2 class="text-xl md:text-2xl font-bold text-gray-900">Sejarah {{ desaName }}</h2>
              <span v-if="sejarahProfile?.section_name" class="text-xs text-gray-500 bg-gray-100 px-2.5 py-1 rounded-full">
                {{ sejarahProfile.section_endpoint }}
              </span>
            </div>
            <div v-if="sejarahProfile?.content" class="text-gray-700 leading-relaxed whitespace-pre-line">{{ sejarahProfile.content }}</div>
            <EmptyState v-else message="Data sejarah tidak tersedia" />
          </div>

          <!-- Visi Misi -->
          <div v-else-if="selectedCategory === 'Visi Misi'" class="space-y-6">
            <div class="bg-white rounded-xl shadow-sm border border-gray-200 p-6 md:p-8">
              <div class="flex items-center justify-between mb-4">
                <h2 class="text-xl font-bold text-gray-900">Visi</h2>
                <span v-if="visionProfile?.section_name" class="text-xs text-gray-500 bg-gray-100 px-2.5 py-1 rounded-full">
                  {{ visionProfile.section_endpoint }}
                </span>
              </div>
              <div v-if="visionProfile?.content" class="text-gray-700 leading-relaxed italic whitespace-pre-line">{{ visionProfile.content }}</div>
              <EmptyState v-else message="Data visi tidak tersedia" />
            </div>
            <div v-if="missionProfiles.length > 0" class="bg-white rounded-xl shadow-sm border border-gray-200 p-6 md:p-8">
              <div class="flex items-center justify-between mb-4">
                <h2 class="text-xl font-bold text-gray-900">Misi</h2>
                <span class="text-xs text-gray-500 bg-gray-100 px-2.5 py-1 rounded-full">{{ missionProfiles.length }} poin</span>
              </div>
              <ol class="list-decimal list-inside space-y-3 text-gray-700">
                <li v-for="m in missionProfiles" :key="m.id" class="leading-relaxed whitespace-pre-line">{{ m.content }}</li>
              </ol>
            </div>
          </div>

          <!-- Struktur Organisasi -->
          <div v-else-if="selectedCategory === 'Struktur'" class="bg-white rounded-xl shadow-sm border border-gray-200 p-6 md:p-8">
            <div class="flex items-center justify-between mb-6">
              <h2 class="text-xl md:text-2xl font-bold text-gray-900">Struktur Organisasi</h2>
              <span v-if="strukturProfile?.section_name" class="text-xs text-gray-500 bg-gray-100 px-2.5 py-1 rounded-full">
                {{ strukturProfile.section_endpoint }}
              </span>
            </div>
            <div v-if="strukturProfile?.content" class="text-gray-700 leading-relaxed whitespace-pre-line">{{ strukturProfile.content }}</div>
            <EmptyState v-else message="Data struktur tidak tersedia" />
          </div>

          <!-- Perangkat Desa -->
          <div v-else-if="selectedCategory === 'Perangkat'" class="bg-white rounded-xl shadow-sm border border-gray-200 p-6 md:p-8">
            <h2 class="text-xl font-bold text-gray-900 mb-6">Perangkat Desa</h2>
            <div v-if="paginatedOfficials.length > 0" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
              <div v-for="official in paginatedOfficials" :key="official.id" class="bg-gray-50 rounded-xl p-6 text-center hover:bg-gray-100 transition-colors">
                <div class="w-20 h-20 mx-auto mb-4 bg-gray-200 rounded-full overflow-hidden flex items-center justify-center">
                  <img
                    v-if="officialPhoto(official)"
                    :src="officialPhoto(official)"
                    :alt="official.name"
                    loading="lazy"
                    class="w-full h-full object-cover"
                    @error="handleImageError"
                  />
                  <svg v-else class="w-10 h-10 text-gray-400" fill="currentColor" viewBox="0 0 20 20">
                    <path fill-rule="evenodd" d="M10 9a3 3 0 100-6 3 3 0 000 6zm-7 9a7 7 0 1114 0H3z" clip-rule="evenodd" />
                  </svg>
                </div>
                <h3 class="text-lg font-bold text-gray-900 mb-1">{{ official.name }}</h3>
                <p class="text-emerald-600 font-medium mb-2">{{ official.position }}</p>
                <p v-if="official.phone" class="text-gray-500 text-sm">{{ official.phone }}</p>
                <p v-if="official.description" class="text-gray-500 text-xs mt-2 line-clamp-2">{{ official.description }}</p>
              </div>
            </div>
            <EmptyState v-else message="Data perangkat desa tidak tersedia" />

            <Pagination
              v-if="totalPages > 1"
              v-model:current-page="currentPage"
              :total-pages="totalPages"
              class="mt-6"
            />
          </div>

          <div v-else class="text-center text-gray-500 py-8">
            Pilih kategori dari sidebar
          </div>
        </div>

        <div v-else class="flex items-center justify-center h-full">
          <p class="text-gray-500">Pilih kategori dari sidebar</p>
        </div>
      </template>
    </div>

    <div class="hidden md:flex w-80 bg-white border-l border-gray-200 flex-col overflow-hidden z-[50]">
      <div class="p-5 border-b border-gray-200">
        <h2 class="text-base font-semibold text-gray-900">Profil Desa</h2>
        <p class="text-xs text-gray-500 mt-1">Informasi {{ desaName }}</p>
      </div>
      
      <div class="p-4 space-y-3 border-b border-gray-100">
        <SearchInput v-model="searchText" placeholder="Cari kategori..." />
        <select
          v-model="selectedCategory"
          class="w-full px-3 py-2.5 text-sm border border-gray-200 rounded-lg focus:ring-2 focus:ring-emerald-500 focus:border-emerald-500 bg-gray-50"
        >
          <option value="">Semua Kategori</option>
          <option v-for="cat in categories" :key="cat.id" :value="cat.id">{{ cat.name }}</option>
        </select>
      </div>

      <div class="flex-1 overflow-y-auto">
        <div class="p-4 space-y-2">
          <button
            v-for="cat in filteredCategories"
            :key="cat.id"
            @click="selectCategory(cat.id)"
            :class="[
              'w-full p-4 rounded-xl text-left transition-all flex items-center gap-3',
              selectedCategory === cat.id
                ? 'bg-emerald-50 border-2 border-emerald-500'
                : 'bg-gray-50 border-2 border-transparent hover:bg-gray-100'
            ]"
          >
            <div :class="['w-10 h-10 rounded-lg flex items-center justify-center', cat.bgColor]">
              <svg :class="['w-5 h-5', cat.iconColor]" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" :d="cat.icon" />
              </svg>
            </div>
            <div>
              <h4 class="text-sm font-medium text-gray-900">{{ cat.name }}</h4>
              <p class="text-xs text-gray-500">{{ cat.description }}</p>
            </div>
          </button>
        </div>
      </div>

      <div class="p-4 border-t border-gray-100 bg-gray-50">
        <p class="text-xs text-gray-500 text-center">{{ filteredCategories.length }} kategori</p>
      </div>
    </div>

    <button
      v-if="!loading"
      @click="showSidebar = !showSidebar"
      class="absolute bottom-6 right-4 z-[500] bg-emerald-600 text-white rounded-full shadow-lg p-3 hover:bg-emerald-700 transition-colors md:hidden"
      title="Kategori Profil"
    >
      <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
      </svg>
    </button>

    <Transition name="slide-up">
      <div v-if="showSidebar && isMobile" class="absolute inset-x-0 bottom-0 z-[600] bg-white rounded-t-2xl shadow-xl max-h-[70vh] flex flex-col md:hidden">
        <div class="p-4 border-b border-gray-200 flex items-center justify-between">
          <h2 class="text-base font-semibold text-gray-900">Profil Desa</h2>
          <button @click="showSidebar = false" class="p-2 hover:bg-gray-100 rounded-full">
            <svg class="w-5 h-5 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
        <div class="p-4 space-y-3 border-b border-gray-100">
          <SearchInput v-model="searchText" placeholder="Cari kategori..." />
          <select
            v-model="selectedCategory"
            class="w-full px-3 py-2.5 text-sm border border-gray-200 rounded-lg focus:ring-2 focus:ring-emerald-500 focus:border-emerald-500 bg-gray-50"
          >
            <option value="">Semua Kategori</option>
            <option v-for="cat in categories" :key="cat.id" :value="cat.id">{{ cat.name }}</option>
          </select>
        </div>
        <div class="p-4 space-y-2 overflow-y-auto">
          <button
            v-for="cat in filteredCategories"
            :key="cat.id"
            @click="selectCategory(cat.id); showSidebar = false"
            :class="[
              'w-full p-4 rounded-xl text-left transition-all flex items-center gap-3',
              selectedCategory === cat.id
                ? 'bg-emerald-50 border-2 border-emerald-500'
                : 'bg-gray-50 border-2 border-transparent hover:bg-gray-100'
            ]"
          >
            <div :class="['w-10 h-10 rounded-lg flex items-center justify-center', cat.bgColor]">
              <svg :class="['w-5 h-5', cat.iconColor]" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" :d="cat.icon" />
              </svg>
            </div>
            <div>
              <h4 class="text-sm font-medium text-gray-900">{{ cat.name }}</h4>
              <p class="text-xs text-gray-500">{{ cat.description }}</p>
            </div>
          </button>
        </div>
      </div>
    </Transition>
  </div>
</template>

<script>
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { getPublicProfileList, getPublicStrukturList, mediaUrl } from '../services/desaService'
import LoadingSpinner from '../components/common/LoadingSpinner.vue'
import EmptyState from '../components/common/EmptyState.vue'
import SearchInput from '../components/common/SearchInput.vue'
import Pagination from '../components/common/Pagination.vue'
import { useDesaInfo } from '../composables/useDesaInfo'

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
    const selectedCategory = ref('')
    const showSidebar = ref(false)
    const isMobile = ref(false)
    const searchText = ref('')
    const currentPage = ref(1)
    const itemsPerPage = 6

    const { desaInfo } = useDesaInfo()
    const desaName = computed(() => desaInfo.value?.name || 'Desa')

    const profiles = ref([])
    const officials = ref([])

    const findProfile = (endpoint, names = []) => {
      return profiles.value.find(p =>
        p.state !== false && (
          p.section_endpoint === endpoint ||
          (p.section_name && names.includes(p.section_name))
        )
      ) || null
    }

    const findProfilesByPrefix = (prefix) => {
      return profiles.value
        .filter(p => p.state !== false && p.section_endpoint && p.section_endpoint.startsWith(prefix))
        .sort((a, b) => a.section_endpoint.localeCompare(b.section_endpoint))
    }

    const sejarahProfile = computed(() => findProfile('/sejarah', ['Sejarah', 'Sejarah Desa']))
    const strukturProfile = computed(() => findProfile('/struktur', ['Struktur', 'Struktur Organisasi']))
    const visionProfile = computed(() => findProfile('/visi', ['Visi']))
    const missionProfiles = computed(() => findProfilesByPrefix('/misi'))

    const categories = [
      { 
        id: 'Sejarah', 
        name: 'Sejarah', 
        description: 'Sejarah desa',
        icon: 'M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z',
        bgColor: 'bg-blue-100',
        iconColor: 'text-blue-600'
      },
      { 
        id: 'Visi Misi', 
        name: 'Visi Misi', 
        description: 'Visi dan misi',
        icon: 'M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z',
        bgColor: 'bg-emerald-100',
        iconColor: 'text-emerald-600'
      },
      { 
        id: 'Struktur', 
        name: 'Struktur', 
        description: 'Struktur organisasi',
        icon: 'M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-3 7h3m-3 4h3m-6-4h.01M9 16h.01',
        bgColor: 'bg-purple-100',
        iconColor: 'text-purple-600'
      },
      { 
        id: 'Perangkat', 
        name: 'Perangkat', 
        description: 'Perangkat desa',
        icon: 'M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z',
        bgColor: 'bg-amber-100',
        iconColor: 'text-amber-600'
      }
    ]

    const filteredCategories = computed(() => {
      let filtered = categories
      if (searchText.value) {
        const query = searchText.value.toLowerCase()
        filtered = filtered.filter(cat => 
          cat.name.toLowerCase().includes(query) ||
          cat.description.toLowerCase().includes(query)
        )
      }
      if (selectedCategory.value) {
        filtered = filtered.filter(cat => cat.id === selectedCategory.value)
      }
      return filtered
    })

    const totalPages = computed(() => {
      return Math.ceil(officials.value.length / itemsPerPage)
    })

    const startIndex = computed(() => (currentPage.value - 1) * itemsPerPage)
    const endIndex = computed(() => startIndex.value + itemsPerPage)

    const paginatedOfficials = computed(() => {
      return officials.value.slice(startIndex.value, endIndex.value)
    })

    const selectCategory = (id) => {
      selectedCategory.value = id
      currentPage.value = 1
    }

    const checkMobile = () => {
      isMobile.value = window.innerWidth < 768
    }

    onMounted(async () => {
      checkMobile()
      window.addEventListener('resize', checkMobile)

      try {
        const [profileResult, strukturResult] = await Promise.all([
          getPublicProfileList({ page: 1, limit: 50 }),
          getPublicStrukturList().catch(() => [])
        ])

        const list = Array.isArray(profileResult) ? profileResult : []
        profiles.value = list

        officials.value = Array.isArray(strukturResult) ? strukturResult : []
      } catch (error) {
        console.error('Error fetching profil data:', error)
        profiles.value = []
        officials.value = []
      } finally {
        loading.value = false
      }
    })

    onBeforeUnmount(() => {
      window.removeEventListener('resize', checkMobile)
    })

    const handleImageError = (event) => {
      event.target.style.display = 'none'
    }

    const officialPhoto = (item) => mediaUrl(item, 'single')

    return {
      loading,
      desaName,
      categories,
      selectedCategory,
      showSidebar,
      isMobile,
      searchText,
      currentPage,
      totalPages,
      paginatedOfficials,
      filteredCategories,
      sejarahProfile,
      strukturProfile,
      visionProfile,
      missionProfiles,
      selectCategory,
      handleImageError,
      officialPhoto
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