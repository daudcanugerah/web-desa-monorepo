<template>
  <div class="max-w-screen-2xl mx-auto">
    <div class="flex items-center gap-3 mb-6">
      <AppBackButton />
      <h1 class="text-2xl font-bold text-secondary-800 dark:text-secondary-100">{{ isEdit ? 'Edit Dashboard' : 'Tambah Dashboard' }}</h1>
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-2 xl:grid-cols-5 gap-6">
      <!-- Form -->
      <div class="lg:col-span-1 xl:col-span-2 bg-white dark:bg-secondary-800 rounded-lg shadow-sm border border-secondary-200 dark:border-secondary-700 p-6 space-y-6">
        <div>
          <label class="block text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-1">Nama Bagian <span class="text-red-500">*</span></label>
          <input
            v-model="form.section_name"
            type="text"
            placeholder="Contoh: Dashboard Keuangan"
            class="w-full px-3 py-2 border rounded text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 bg-white dark:bg-secondary-800 text-secondary-700 dark:text-secondary-200"
            :class="errors.section_name ? 'border-red-500' : 'border-secondary-300 dark:border-secondary-600'"
          />
          <p v-if="errors.section_name" class="text-red-500 dark:text-red-400 text-xs mt-1">{{ errors.section_name }}</p>
        </div>

        <div>
          <label class="block text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-1">
            Endpoint <span class="text-secondary-500 text-xs">(otomatis dari nama bagian)</span>
          </label>
          <div
            class="w-full px-3 py-2 border border-secondary-300 dark:border-secondary-600 rounded text-sm bg-secondary-50 dark:bg-secondary-900 text-secondary-700 dark:text-secondary-300 font-mono"
          >
            {{ form.section_endpoint || '—' }}
          </div>
        </div>

        <div>
          <AppSelect
            v-model="form.component_type"
            label="Tipe Component"
            :required="true"
            :error="errors.component_type"
          >
            <option value="">{{ t('filter.pickType') }}</option>
            <option value="dashboard">Dashboard</option>
            <option value="question">Question</option>
          </AppSelect>
        </div>

        <div>
          <label class="block text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-1">Component ID <span class="text-red-500">*</span></label>
          <input
            :value="form.component_id"
            @input="form.component_id = $event.target.value ? parseInt($event.target.value) : ''"
            type="number"
            placeholder="Contoh: 12"
            class="w-full px-3 py-2 border rounded text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 bg-white dark:bg-secondary-800 text-secondary-700 dark:text-secondary-200"
            :class="errors.component_id ? 'border-red-500' : 'border-secondary-300 dark:border-secondary-600'"
          />
          <p v-if="errors.component_id" class="text-red-500 text-xs mt-1">{{ errors.component_id }}</p>
          <p class="text-xs text-secondary-500 dark:text-secondary-400 mt-1">ID dari dashboard atau question di Metabase (integer)</p>
        </div>

        <div>
          <label class="block text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-1">Status</label>
          <div class="flex items-center">
            <input
              v-model="form.state"
              type="checkbox"
              class="h-4 w-4 text-blue-600 focus:ring-blue-500 border-secondary-300 dark:border-secondary-600 rounded"
            />
            <label class="ml-2 text-sm text-secondary-700 dark:text-secondary-300">Aktif</label>
          </div>
        </div>

        <div>
          <AppCategoryPicker
            v-model="form.category"
            :suggestions="categoryInput.suggestions.value"
            :service="infographicService"
            label="Kategori (Opsional)"
            placeholder="Pilih atau buat kategori..."
            @manage="showCategoryModal = true"
          />
        </div>

        <div class="flex gap-3 pt-4">
          <AppButton
            type="button"
            variant="primary"
            :loading="saving"
            :disabled="saving"
            @click="handleSubmit"
          >
            {{ isEdit ? 'Simpan Perubahan' : 'Tambah Dashboard' }}
          </AppButton>

          <AppButton
            type="button"
            variant="success"
            :loading="previewLoading"
            :disabled="!canPreview || previewLoading"
            @click="generatePreview"
          >
            Preview
          </AppButton>

          <AppButton
            type="button"
            variant="secondary"
            @click="$router.back()"
          >
            Batal
          </AppButton>
        </div>
      </div>

      <!-- Preview -->
      <div class="lg:col-span-1 xl:col-span-3 bg-white dark:bg-secondary-800 rounded-lg shadow-sm border border-secondary-200 dark:border-secondary-700 p-6">
        <h3 class="text-lg font-semibold mb-4">Preview Dashboard</h3>
        
        <div v-if="previewLoading" class="flex items-center justify-center h-96">
          <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
        </div>
        
        <div v-else-if="previewUrl" class="border rounded overflow-hidden">
          <metabase-question
            v-if="form.component_type === 'question'"
            :token="previewToken"
            with-title="true"
            with-downloads="true"
            class="w-full h-96"
          ></metabase-question>
          <metabase-dashboard
            v-else-if="form.component_type === 'dashboard'"
            :token="previewToken"
            with-title="true"
            with-downloads="true"
            class="w-full h-96"
          ></metabase-dashboard>
        </div>
        
        <div v-else class="flex flex-col items-center justify-center h-96 text-secondary-500 dark:text-secondary-400">
          <svg class="w-16 h-16 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z" />
          </svg>
          <p>Klik tombol Preview untuk melihat dashboard</p>
        </div>
      </div>
    </div>

    <CategoryManagerModal
      :show="showCategoryModal"
      :service="infographicService"
      title="Kelola Kategori Infographic"
      @close="showCategoryModal = false"
    />
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { infographicService } from '../../services/infographic.service'
import { useNotificationStore } from '../../stores/notification'
import AppButton from '../../components/common/AppButton.vue'
import AppSelect from '../../components/common/AppSelect.vue'
import AppBackButton from '../../components/common/AppBackButton.vue'
import AppCategoryPicker from '../../components/common/AppCategoryPicker.vue'
import CategoryManagerModal from '../../components/common/CategoryManagerModal.vue'
import { useCategoryInput } from '../../composables/useCategoryInput'
import { useMetabase } from '../../composables/useMetabase'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const notificationStore = useNotificationStore()
const { initializeMetabase } = useMetabase()

const isEdit = computed(() => !!route.params.id)
const form = ref({
  section_name: '',
  section_endpoint: '',
  component_id: null,
  component_type: '',
  state: true,
  category: { id: '', name: '' },
})
const errors = ref({})
const saving = ref(false)
const showCategoryModal = ref(false)
const categoryInput = useCategoryInput(infographicService)

function slugify(input) {
  const base = (input || '')
    .toString()
    .normalize('NFKD')
    .replace(/[\u0300-\u036f]/g, '')
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '')
    .slice(0, 48) || 'bagian'
  const suffix = Math.random().toString(36).slice(2, 8)
  return `${base}-${suffix}`
}

watch(
  () => form.value.section_name,
  (name) => {
    form.value.section_endpoint = `/infographic/${slugify(name)}`
  },
)

// Preview
const previewUrl = ref('')
const previewToken = ref('')
const previewLoading = ref(false)

const canPreview = computed(() => {
  return form.value.component_id && form.value.component_type
})

function validate() {
  errors.value = {}

  if (!form.value.section_name.trim()) {
    errors.value.section_name = 'Nama bagian wajib diisi'
  }

  if (!form.value.component_id) {
    errors.value.component_id = 'Component ID wajib diisi'
  } else if (!Number.isInteger(form.value.component_id) || form.value.component_id <= 0) {
    errors.value.component_id = 'Component ID harus berupa angka positif'
  }
  
  if (!form.value.component_type) {
    errors.value.component_type = 'Tipe component wajib dipilih'
  }
  
  return Object.keys(errors.value).length === 0
}

async function generatePreview() {
  if (!canPreview.value) return
  
  previewLoading.value = true
  previewUrl.value = ''
  previewToken.value = ''
  
  try {
    // Initialize Metabase if not already done
    await initializeMetabase()
    
    const res = await infographicService.generatePreviewToken({
      component_id: form.value.component_id,
      component_type: form.value.component_type,
    })
    
    previewToken.value = res.data.token
    previewUrl.value = 'loaded' // Just to indicate preview is ready
  } catch (err) {
    notificationStore.error(err.response?.data?.error || 'Gagal membuat preview')
  } finally {
    previewLoading.value = false
  }
}

async function handleSubmit() {
  if (!validate()) return
  
  saving.value = true
  try {
    const categoryId = form.value.category?.id || ''
    const data = {
      section_name: form.value.section_name,
      section_endpoint: form.value.section_endpoint,
      component_id: form.value.component_id,
      component_type: form.value.component_type,
      state: form.value.state,
    }
    if (categoryId) data.category_id = categoryId

    if (isEdit.value) {
      await infographicService.update(route.params.id, data)
      notificationStore.success('Dashboard berhasil diperbarui')
    } else {
      await infographicService.create(data)
      notificationStore.success('Dashboard berhasil ditambahkan')
    }
    
    router.push('/infographic')
  } catch (err) {
    notificationStore.error(err.response?.data?.error || 'Gagal menyimpan dashboard')
  } finally {
    saving.value = false
  }
}

// Auto-generate preview when component info changes
watch([() => form.value.component_id, () => form.value.component_type], () => {
  if (canPreview.value && isEdit.value) {
    generatePreview()
  }
})

onMounted(async () => {
  // Initialize Metabase
  try {
    await initializeMetabase()
  } catch (err) {
    console.error('Failed to initialize Metabase:', err)
  }

  await categoryInput.load()

  if (isEdit.value) {
    try {
      const res = await infographicService.get(route.params.id)
      const data = res.data
      form.value = {
        section_name: data.section_name,
        section_endpoint: data.section_endpoint,
        component_id: parseInt(data.component_id),
        component_type: data.component_type,
        state: data.state,
        category: {
          id: data.category_id || data.category?.id || '',
          name: data.category?.name || '',
        },
      }

      // Auto-load preview for existing infographic
      if (canPreview.value) {
        generatePreview()
      }
    } catch (err) {
      notificationStore.error('Gagal memuat data dashboard')
      router.push('/infographic')
    }
  }
})
</script>