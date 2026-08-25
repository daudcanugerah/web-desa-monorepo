<template>
  <div>
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-2xl font-bold text-secondary-800 dark:text-secondary-200">{{ $t('profileSection.title') }}</h1>
      <router-link
        to="/profile/sections/create"
        class="inline-flex items-center px-4 py-2 rounded-lg font-medium transition-colors bg-primary-600 text-white hover:bg-primary-700"
      >
        {{ $t('profileSection.addNew') }}
      </router-link>
    </div>

    <!-- Search and Filters -->
    <div class="bg-white dark:bg-secondary-800 rounded-lg shadow-sm border border-secondary-200 dark:border-secondary-700 p-4 mb-6">
      <div class="flex flex-col sm:flex-row gap-4">
        <div class="flex-1">
          <input
            v-model="searchQuery"
            type="text"
            :placeholder="$t('profileSection.searchPlaceholder')"
            class="w-full px-3 py-2 border border-secondary-300 dark:border-secondary-600 rounded text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 bg-white dark:bg-secondary-800 text-secondary-700 dark:text-secondary-200"
          />
        </div>
        <div class="sm:w-48">
          <AppSelect v-model="selectedSection">
            <option value="">{{ t('profileSection.allSections') }}</option>
            <option v-for="section in sectionNames" :key="section" :value="section">
              {{ section }}
            </option>
          </AppSelect>
        </div>
        <div class="sm:w-32">
          <AppSelect v-model="selectedState">
            <option value="">{{ t('profileSection.allStatus') }}</option>
            <option value="true">{{ t('common.active') }}</option>
            <option value="false">{{ t('common.inactive') }}</option>
          </AppSelect>
        </div>
      </div>
    </div>

    <!-- Table -->
    <div class="bg-white dark:bg-secondary-800 rounded-lg shadow-sm border border-secondary-200 dark:border-secondary-700">
      <AppTable
        :columns="columns"
        :data="profileSections"
        :loading="loading"
        :empty-message="$t('profileSection.noResults')"
      >
        <template #section_name="{ row }">
          <span class="font-medium text-secondary-900 dark:text-secondary-100">{{ row.section_name }}</span>
        </template>
        
        <template #content="{ row }">
          <div class="max-w-xs text-sm text-secondary-600 dark:text-secondary-400 line-clamp-2">
            {{ row.content.length > 100 ? row.content.slice(0, 100).replace(/<[^>]*>/g, '') + '…' : row.content.replace(/<[^>]*>/g, '') }}
          </div>
        </template>

        <template #state="{ row }">
          <AppBadge :variant="row.state ? 'success' : 'danger'">
            {{ row.state ? $t('common.active') : $t('common.inactive') }}
          </AppBadge>
        </template>

        <template #created_at="{ row }">
          <span class="text-sm text-secondary-600 dark:text-secondary-400">{{ formatDate(row.created_at) }}</span>
        </template>

        <template #actions="{ row }">
          <div class="flex items-center gap-2">
            <AppButton size="sm" variant="primary" @click="router.push(`/profile/sections/${row.id}/edit`)">
              {{ $t('common.edit') }}
            </AppButton>
            <AppButton size="sm" variant="danger" @click="handleDelete(row)">
              {{ $t('common.delete') }}
            </AppButton>
          </div>
        </template>

        <template #empty>
          <AppEmptyState
            v-if="hasFilters"
            icon="search"
            title="Tidak ada bagian yang cocok"
            :hint="emptyHint"
          >
            <AppButton variant="ghost" @click="resetFilters">Reset filter</AppButton>
          </AppEmptyState>
          <AppEmptyState
            v-else
            icon="document"
            title="Belum ada bagian profil"
            hint="Tambahkan bagian profil pertama untuk menampilkan informasi detail desa."
          >
            <AppButton variant="primary" @click="router.push('/profile/sections/create')">{{ $t('profileSection.addNew') }}</AppButton>
          </AppEmptyState>
        </template>
      </AppTable>

      <!-- Pagination -->
      <div v-if="pagination.total > 0" class="border-t border-secondary-200 dark:border-secondary-700 px-4 py-3">
        <AppPagination
          :page="pagination.page"
          :total-pages="Math.ceil(pagination.total / pagination.limit)"
          @change="handlePageChange"
        />
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import AppTable from '../../components/common/AppTable.vue'
import AppBadge from '../../components/common/AppBadge.vue'
import AppSelect from '../../components/common/AppSelect.vue'
import AppPagination from '../../components/common/AppPagination.vue'
import AppEmptyState from '../../components/common/AppEmptyState.vue'
import { profileSectionService } from '../../services/profile-section.service'
import { useNotificationStore } from '../../stores/notification'
import { useConfirm } from '../../composables/useConfirm'
import { useUrlFilters } from '../../composables/useUrlFilters'
import { formatDate as formatDateUtil } from '../../utils/dateFormat'

const router = useRouter()
const { t } = useI18n()
const notificationStore = useNotificationStore()
const { confirm } = useConfirm()

const profileSections = ref([])
const sectionNames = ref([])
const loading = ref(false)
const pagination = ref({ page: 1, limit: 10, total: 0 })

const { refs: urlRefs, syncToUrl, reset: resetUrlFilters } = useUrlFilters(['q', 'section', 'state'])
const searchQuery = urlRefs.q
const selectedSection = urlRefs.section
const selectedState = urlRefs.state

watch([searchQuery, selectedSection, selectedState], () => {
  syncToUrl()
  clearTimeout(searchTimeout)
  searchTimeout = setTimeout(() => {
    pagination.value.page = 1
    fetchProfileSections()
  }, 300)
})

const columns = computed(() => [
  { key: 'section_name', label: t('profileSection.sectionName'), sortable: true },
  { key: 'content', label: t('profileSection.content') },
  { key: 'state', label: t('common.status') },
  { key: 'created_at', label: t('profileSection.created') },
  { key: 'actions', label: t('common.actions'), width: '120px' },
])

const searchParams = computed(() => ({
  page: pagination.value.page,
  limit: pagination.value.limit,
  ...(searchQuery.value && { search: searchQuery.value }),
  ...(selectedSection.value && { section_name: selectedSection.value }),
  ...(selectedState.value && { state: selectedState.value === 'true' }),
}))

const hasFilters = computed(() => !!(searchQuery.value || selectedSection.value || selectedState.value))
const emptyHint = computed(() => {
  const parts = []
  if (searchQuery.value) parts.push(`pencarian "${searchQuery.value}"`)
  if (selectedSection.value) parts.push(`bagian "${selectedSection.value}"`)
  if (selectedState.value) parts.push(`status "${selectedState.value === 'true' ? 'aktif' : 'nonaktif'}"`)
  return `Tidak ada hasil untuk ${parts.join(', ')}.`
})

function resetFilters() {
  resetUrlFilters()
  pagination.value.page = 1
  fetchProfileSections()
}

async function fetchProfileSections() {
  loading.value = true
  try {
    const res = await profileSectionService.list(searchParams.value)
    profileSections.value = res.data.profile || []
    pagination.value = res.data.pagination
  } catch (err) {
    notificationStore.error(err.response?.data?.error || t('profileSection.loadFailed'))
  } finally {
    loading.value = false
  }
}

async function fetchSectionNames() {
  try {
    const res = await profileSectionService.getSectionNames()
    sectionNames.value = res.data.section_names || []
  } catch (err) {
    console.error('Failed to fetch section names:', err)
  }
}

async function handleDelete(profileSection) {
  const confirmed = await confirm(
    t('common.confirm'),
    t('profileSection.deleteConfirm', { name: profileSection.section_name }),
    profileSection.section_name
  )
  
  if (!confirmed) return
  try {
    await profileSectionService.delete(profileSection.id)
    notificationStore.success(t('profileSection.deleteSuccess'))
    fetchProfileSections()
  } catch (err) {
    notificationStore.error(err.response?.data?.error || t('profileSection.deleteFailed'))
  }
}

function handlePageChange(page) {
  pagination.value.page = page
  fetchProfileSections()
}

function formatDate(dateStr) {
  return formatDateUtil(dateStr)
}

// Debounced search
let searchTimeout

onMounted(() => {
  fetchProfileSections()
  fetchSectionNames()
})
</script>
