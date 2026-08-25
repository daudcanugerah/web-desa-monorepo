<template>
  <div class="max-w-4xl mx-auto">
    <div class="flex items-center gap-3 mb-6">
      <AppBackButton />
      <h1 class="text-2xl font-bold text-secondary-800 dark:text-secondary-100">{{ isEdit ? 'Edit Dokumen PPID' : 'Tambah Dokumen PPID' }}</h1>
    </div>

    <div class="bg-white dark:bg-secondary-800 rounded-lg shadow-sm border border-secondary-200 dark:border-secondary-700 p-6 space-y-5">
      <AppInput
        v-model="form.title"
        label="Judul"
        :error="errors.title"
        placeholder="Judul dokumen"
      />

      <AppCategoryPicker
        v-model="form.category"
        :suggestions="categorySuggestions"
        :service="ppidService"
        label="Kategori"
        placeholder="Pilih atau buat kategori..."
        @manage="showCategoryModal = true"
      />

      <AppInput
        v-model="form.publication_at"
        type="date"
        label="Tanggal Publikasi (Opsional)"
      />

      <div>
        <label class="block text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-1">Deskripsi</label>
        <textarea
          v-model="form.description"
          rows="4"
          placeholder="Deskripsi dokumen (opsional)"
          class="w-full px-3 py-2 border border-secondary-300 dark:border-secondary-600 rounded text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 resize-y bg-white dark:bg-secondary-800 text-secondary-700 dark:text-secondary-200"
        />
      </div>

      <div>
        <label class="block text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-1">
          Thumbnail (Opsional)
        </label>
        <ImageUpload
          :existing-url="existingThumbnailUrl"
          @change="onThumbnailChange"
        />
      </div>

      <div>
        <label class="block text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-1">
          Dokumen <span v-if="!isEdit" class="text-red-500">*</span>
        </label>
        <FileUpload
          ref="fileUploadRef"
          :existing-url="existingDocUrl"
          @change="onFileChange"
        />
        <div v-if="errors.document" class="mt-1 text-xs text-red-600">{{ errors.document }}</div>

        <div v-if="isEdit && existingDocUrl" class="mt-2">
          <a
            :href="existingDocUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="inline-flex items-center gap-2 px-3 py-1.5 rounded-lg border-2 border-primary-600 text-primary-600 dark:text-primary-300 text-xs font-medium hover:bg-primary-50 dark:hover:bg-primary-900/30 transition-colors"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
            </svg>
            Buka Dokumen
          </a>
          <span v-if="form.document_name" class="ml-2 text-xs text-secondary-500 dark:text-secondary-400">{{ form.document_name }}</span>
        </div>
      </div>

      <div class="flex gap-3 pt-2">
        <AppButton variant="primary" :loading="saving" @click="handleSubmit">
          {{ isEdit ? 'Simpan' : 'Tambah' }}
        </AppButton>
        <AppButton variant="secondary" @click="$router.push('/ppid')">Batal</AppButton>
      </div>
    </div>

    <CategoryManagerModal
      :show="showCategoryModal"
      :service="ppidService"
      title="Kelola Kategori PPID"
      @close="showCategoryModal = false"
      @selected="onCategorySelected"
    />
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import AppInput from '../../components/common/AppInput.vue'
import AppButton from '../../components/common/AppButton.vue'
import FileUpload from '../../components/forms/FileUpload.vue'
import ImageUpload from '../../components/forms/ImageUpload.vue'
import CategoryManagerModal from '../../components/common/CategoryManagerModal.vue'
import AppCategoryPicker from '../../components/common/AppCategoryPicker.vue'
import { ppidService } from '../../services/ppid.service'
import { useNotificationStore } from '../../stores/notification'
import AppBackButton from '../../components/common/AppBackButton.vue'
import { getImageUrl } from '../../utils/imageUrl'

const route = useRoute()
const router = useRouter()
const notificationStore = useNotificationStore()

const isEdit = computed(() => !!route.params.id)
const existingDocUrl = ref(null)
const existingThumbnailUrl = ref(null)
const existingDocumentMediaId = ref('')
const existingThumbnailMediaId = ref('')
const fileUploadRef = ref(null)
const showCategoryModal = ref(false)
const categorySuggestions = ref([])

const form = ref({ title: '', category: { id: '', name: '' }, description: '', publication_at: '', document_name: '' })
const documentFile = ref(null)
const thumbnailFile = ref(null)
const errors = ref({})
const saving = ref(false)

function onCategorySelected(category) {
  form.value.category = { id: category.id, name: category.name }
}

async function loadCategorySuggestions() {
  try {
    const res = await ppidService.getCategories({ limit: 100 })
    const list = res.data?.categories || res.data || []
    categorySuggestions.value = (Array.isArray(list) ? list : [])
      .filter((c) => c && (c.id || c.name))
      .map((c) => ({ id: c.id, name: c.name }))
  } catch {
    categorySuggestions.value = []
  }
}

function onFileChange(file) {
  documentFile.value = file
  errors.value.document = ''
}

function onThumbnailChange(file) {
  thumbnailFile.value = file
}

function validate() {
  errors.value = {}
  if (!form.value.title.trim()) errors.value.title = 'Judul wajib diisi'
  if (!isEdit.value && !documentFile.value) errors.value.document = 'Dokumen wajib diunggah'
  return Object.keys(errors.value).length === 0
}

async function handleSubmit() {
  if (!validate()) return
  saving.value = true
  try {
    const categoryId = form.value.category?.id || ''
    const publication_at = form.value.publication_at
      ? new Date(form.value.publication_at + 'T00:00:00Z').toISOString()
      : ''

    // Upload new files via the combined endpoint to get media ids first.
    let documentMediaId = existingDocumentMediaId.value
    let thumbnailMediaId = existingThumbnailMediaId.value
    if (documentFile.value || thumbnailFile.value) {
      const up = await ppidService.upload(documentFile.value, thumbnailFile.value)
      const media = up.data?.data || up.data || {}
      if (media.document?.media_id) documentMediaId = media.document.media_id
      if (media.thumbnail?.media_id) thumbnailMediaId = media.thumbnail.media_id
    }

    if (isEdit.value) {
      // PUT /ppid/{id} takes a JSON body.
      const payload = { title: form.value.title }
      if (categoryId) payload.category = categoryId
      if (form.value.description) payload.description = form.value.description
      if (publication_at) payload.publication_at = publication_at
      if (documentMediaId) payload.document_media_id = documentMediaId
      if (thumbnailMediaId) payload.thumbnail_media_id = thumbnailMediaId
      await ppidService.update(route.params.id, payload)
      notificationStore.success('Dokumen diperbarui')
    } else {
      // POST /ppid is multipart; attach the uploaded media ids.
      const fd = new FormData()
      fd.append('title', form.value.title)
      if (categoryId) fd.append('category', categoryId)
      if (form.value.description) fd.append('description', form.value.description)
      if (publication_at) fd.append('publication_at', publication_at)
      if (documentMediaId) fd.append('document_media_id', documentMediaId)
      if (thumbnailMediaId) fd.append('thumbnail_media_id', thumbnailMediaId)
      await ppidService.create(fd)
      notificationStore.success('Dokumen ditambahkan')
    }
    router.push('/ppid')
  } catch (err) {
    if (!err.response) return
    notificationStore.error(err.response?.data?.error || 'Gagal menyimpan dokumen')
  } finally {
    saving.value = false
  }
}

onMounted(async () => {
  await loadCategorySuggestions()
  if (isEdit.value) {
    try {
      const res = await ppidService.adminGet(route.params.id)
      const d = res.data || {}
      form.value = {
        title: d.title || '',
        category: { id: d.category_id || d.category?.id || '', name: d.category?.name || '' },
        description: d.description || '',
        publication_at: d.publication_at ? d.publication_at.split('T')[0] : '',
        document_name: d.document?.filename || '',
      }
      // New schema: document/thumbnail are MediaInfo {media_id, url, thumbnail_url}.
      // getImageUrl prefixes the API host so relative signed URLs resolve
      // against the backend, not the SPA origin.
      existingDocumentMediaId.value = d.document?.media_id || ''
      existingThumbnailMediaId.value = d.thumbnail?.media_id || ''
      existingDocUrl.value = getImageUrl(d.document?.url || d.document?.thumbnail_url || '') || null
      existingThumbnailUrl.value = d.thumbnail?.thumbnail_url || d.thumbnail?.url || null
      if (!existingDocUrl.value && d.document_url) {
        existingDocUrl.value = getImageUrl(d.document_url)
      }
      if (!existingThumbnailUrl.value && d.thumbnail_url) {
        existingThumbnailUrl.value = getImageUrl(d.thumbnail_url)
      }
    } catch {
      notificationStore.error('Gagal memuat dokumen')
      router.push('/ppid')
    }
  }
})
</script>