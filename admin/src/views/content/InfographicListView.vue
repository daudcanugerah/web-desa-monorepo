<template>
  <div>
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-2xl font-bold text-secondary-800 dark:text-secondary-200">Infographic Dashboard</h1>
      <div class="flex gap-2">
        <AppButton variant="primary" @click="router.push('/infographic/create')">Tambah Dashboard</AppButton>
      </div>
    </div>

    <!-- Search and Filters -->
    <div class="bg-white dark:bg-secondary-800 rounded-lg shadow-sm border border-secondary-200 dark:border-secondary-700 p-4 mb-6">
      <div class="flex flex-col sm:flex-row gap-4">
        <div class="flex-1">
          <input
            v-model="searchQuery"
            type="text"
            placeholder="Cari dashboard..."
            class="w-full px-3 py-2 border border-secondary-300 dark:border-secondary-600 rounded text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 bg-white dark:bg-secondary-800 text-secondary-700 dark:text-secondary-200"
          />
        </div>
        <div class="sm:w-48">
          <AppSelect v-model="selectedType">
            <option value="">{{ t('filter.allTypes') }}</option>
            <option value="dashboard">Dashboard</option>
            <option value="question">Question</option>
          </AppSelect>
        </div>
        <div class="sm:w-32">
          <AppSelect v-model="selectedState">
            <option value="">{{ t('filter.allStatuses') }}</option>
            <option value="true">{{ t('filter.active') }}</option>
            <option value="false">{{ t('filter.inactive') }}</option>
          </AppSelect>
        </div>
      </div>
    </div>

    <!-- Table -->
    <div class="bg-white dark:bg-secondary-800 rounded-lg shadow-sm border border-secondary-200 dark:border-secondary-700">
      <AppTable
        :columns="columns"
        :data="infographics"
        :loading="loading"
        empty-message="Tidak ada dashboard ditemukan"
      >
        <template #section_name="{ row }">
          <span class="font-medium text-secondary-900 dark:text-secondary-100">{{ row.section_name }}</span>
        </template>
        
        <template #component_type="{ row }">
          <AppBadge :variant="row.component_type === 'dashboard' ? 'primary' : 'warning'">
            {{ row.component_type === 'dashboard' ? 'Dashboard' : 'Question' }}
          </AppBadge>
        </template>

        <template #component_id="{ row }">
          <span class="text-sm text-secondary-600 dark:text-secondary-400 font-mono">{{ row.component_id }}</span>
        </template>

        <template #state="{ row }">
          <AppBadge :variant="row.state ? 'success' : 'danger'">
            {{ row.state ? 'Aktif' : 'Tidak Aktif' }}
          </AppBadge>
        </template>

        <template #created_at="{ row }">
          <span class="text-sm text-secondary-600 dark:text-secondary-400">{{ formatDate(row.created_at) }}</span>
        </template>

        <template #actions="{ row }">
          <div class="flex gap-2">
            <AppButton size="sm" variant="primary" @click="previewDashboard(row)">
              Preview
            </AppButton>
            <AppButton size="sm" variant="secondary" @click="router.push(`/infographic/${row.id}/edit`)">
              Edit
            </AppButton>
            <AppButton size="sm" variant="danger" @click="handleDelete(row)">
              Hapus
            </AppButton>
          </div>
        </template>

        <template #empty>
          <AppEmptyState
            v-if="hasFilters"
            icon="search"
            title="Tidak ada dashboard yang cocok"
            :hint="emptyHint"
          >
            <AppButton variant="ghost" @click="resetFilters">Reset filter</AppButton>
          </AppEmptyState>
          <AppEmptyState
            v-else
            icon="📊"
            title="Belum ada dashboard"
            hint="Tambahkan dashboard Metabase pertama untuk mulai menampilkan data infografis desa."
          >
            <AppButton variant="primary" @click="router.push('/infographic/create')">Tambah Dashboard</AppButton>
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

    <!-- Preview Modal -->
    <AppModal :show="showPreview" title="Preview Dashboard" size="xl" @close="closePreview">
      <div v-if="previewLoading" class="flex items-center justify-center h-[70vh]">
        <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600"></div>
      </div>
      <div v-else-if="previewToken" class="border rounded overflow-hidden">
        <metabase-question
          v-if="previewComponentType === 'question'"
          :token="previewToken"
          with-title="true"
          with-downloads="true"
          class="w-full h-[70vh]"
        ></metabase-question>
        <metabase-dashboard
          v-else-if="previewComponentType === 'dashboard'"
          :token="previewToken"
          with-title="true"
          with-downloads="true"
          class="w-full h-[70vh]"
        ></metabase-dashboard>
      </div>
      <div v-else class="flex items-center justify-center h-[70vh] text-secondary-500 dark:text-secondary-400">
        Gagal memuat preview
      </div>
    </AppModal>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import AppButton from '../../components/common/AppButton.vue'
import AppTable from '../../components/common/AppTable.vue'
import AppBadge from '../../components/common/AppBadge.vue'
import AppSelect from '../../components/common/AppSelect.vue'
import AppPagination from '../../components/common/AppPagination.vue'
import AppEmptyState from '../../components/common/AppEmptyState.vue'
import AppModal from '../../components/common/AppModal.vue'
import { infographicService } from '../../services/infographic.service'
import { useNotificationStore } from '../../stores/notification'
import { useConfirm } from '../../composables/useConfirm'
import { useMetabase } from '../../composables/useMetabase'
import { useUrlFilters } from '../../composables/useUrlFilters'
import { formatDate as formatDateUtil } from '../../utils/dateFormat'

const router = useRouter()
const { t } = useI18n()
const notificationStore = useNotificationStore()
const { confirmDelete } = useConfirm()
const { initializeMetabase } = useMetabase()

const infographics = ref([])
const loading = ref(false)
const pagination = ref({ page: 1, limit: 10, total: 0 })

const { refs: urlRefs, syncToUrl, reset: resetUrlFilters } = useUrlFilters(['q', 'type', 'state'])
const searchQuery = urlRefs.q
const selectedType = urlRefs.type
const selectedState = urlRefs.state

watch([searchQuery, selectedType, selectedState], () => {
  syncToUrl()
  clearTimeout(searchTimeout)
  searchTimeout = setTimeout(() => {
    pagination.value.page = 1
    fetchInfographics()
  }, 300)
})

const hasFilters = computed(() => !!(searchQuery.value || selectedType.value || selectedState.value))
const emptyHint = computed(() => {
  const parts = []
  if (searchQuery.value) parts.push(`pencarian "${searchQuery.value}"`)
  if (selectedType.value) parts.push(`tipe "${selectedType.value}"`)
  if (selectedState.value) parts.push(`status "${selectedState.value === 'true' ? 'aktif' : 'tidak aktif'}"`)
  return `Tidak ada hasil untuk ${parts.join(', ')}.`
})

function resetFilters() {
  resetUrlFilters()
  pagination.value.page = 1
  fetchInfographics()
}

// Preview modal
const showPreview = ref(false)
const previewToken = ref('')
const previewComponentType = ref('')
const previewLoading = ref(false)

const columns = [
  { key: 'section_name', label: 'Nama Bagian', sortable: true },
  { key: 'component_type', label: 'Tipe' },
  { key: 'component_id', label: 'Component ID' },
  { key: 'state', label: 'Status' },
  { key: 'created_at', label: 'Dibuat' },
  { key: 'actions', label: 'Aksi', width: '180px' },
]

const searchParams = computed(() => ({
  page: pagination.value.page,
  limit: pagination.value.limit,
  ...(searchQuery.value && { search: searchQuery.value }),
  ...(selectedType.value && { component_type: selectedType.value }),
  ...(selectedState.value && { state: selectedState.value === 'true' }),
}))

async function fetchInfographics() {
  loading.value = true
  try {
    const res = await infographicService.list(searchParams.value)
    infographics.value = res.data.infographic || []
    pagination.value = res.data.pagination
  } catch (err) {
    notificationStore.error(err.response?.data?.error || 'Gagal memuat dashboard')
  } finally {
    loading.value = false
  }
}

async function previewDashboard(infographic) {
  showPreview.value = true
  previewLoading.value = true
  previewToken.value = ''
  previewComponentType.value = infographic.component_type

  try {
    // Initialize Metabase if not already done
    await initializeMetabase()

    const res = await infographicService.generatePreviewToken({
      component_id: infographic.component_id,
      component_type: infographic.component_type,
    })

    previewToken.value = res.data.token
  } catch (err) {
    notificationStore.error(err.response?.data?.error || 'Gagal membuat preview')
    closePreview()
  } finally {
    previewLoading.value = false
  }
}

function closePreview() {
  showPreview.value = false
  previewToken.value = ''
  previewComponentType.value = ''
  previewLoading.value = false
}

async function handleDelete(infographic) {
  const confirmed = await confirmDelete('Dashboard', infographic.section_name)
  
  if (confirmed) {
    try {
      await infographicService.delete(infographic.id)
      notificationStore.success('Dashboard berhasil dihapus')
      fetchInfographics()
    } catch (err) {
      notificationStore.error(err.response?.data?.error || 'Gagal menghapus dashboard')
    }
  }
}

function handlePageChange(page) {
  pagination.value.page = page
  fetchInfographics()
}

function formatDate(dateStr) {
  return formatDateUtil(dateStr)
}

// Debounced search
let searchTimeout

onMounted(async () => {
  // Initialize Metabase
  try {
    await initializeMetabase()
  } catch (err) {
    console.error('Failed to initialize Metabase:', err)
  }
  
  fetchInfographics()
})
</script>