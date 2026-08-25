<template>
  <div>
    <!-- Drop zone -->
    <div
      class="border-2 border-dashed rounded-lg p-6 text-center cursor-pointer transition-colors"
      :class="isDragging ? 'border-blue-400 bg-blue-50 dark:bg-blue-900/20' : 'border-secondary-300 dark:border-secondary-600 hover:border-secondary-400 dark:hover:border-secondary-500'"
      @click="fileInput.click()"
      @dragover.prevent="isDragging = true"
      @dragleave.prevent="isDragging = false"
      @drop.prevent="onDrop"
    >
      <div v-if="selectedFile" class="flex items-center justify-center gap-3">
        <svg class="w-8 h-8 text-blue-500 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5"
            d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
        </svg>
        <div class="text-left">
          <p class="text-sm font-medium text-secondary-800 dark:text-secondary-200 truncate max-w-xs">{{ selectedFile.name }}</p>
          <p class="text-xs text-secondary-500 dark:text-secondary-400">{{ formattedSize }}</p>
        </div>
        <AppButton
          variant="secondary"
          size="sm"
          title="Lihat file"
          @click.stop="openPreview"
        >
          Lihat
        </AppButton>
        <AppIconButton
          variant="ghost"
          size="sm"
          class="ml-2"
          title="Clear file"
          @click.stop="clearFile"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </AppIconButton>
      </div>

      <div v-else class="text-secondary-400 dark:text-secondary-500">
        <svg class="mx-auto w-10 h-10 mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5"
            d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
        </svg>
        <p class="text-sm">Klik atau seret file ke sini</p>
        <p class="text-xs mt-1">PDF, DOC, DOCX, XLS, XLSX — maks {{ maxMB }}MB</p>
      </div>
    </div>

    <input
      ref="fileInput"
      type="file"
      :accept="acceptAttr"
      class="hidden"
      @change="onFileChange"
    />

    <p v-if="existingUrl && !selectedFile" class="mt-1 text-xs text-secondary-500 dark:text-secondary-400">
      Dokumen saat ini:
      <a :href="existingUrl" target="_blank" rel="noopener" class="text-blue-600 hover:underline">Lihat dokumen</a>
      — file baru akan menggantikannya.
    </p>
    <p v-if="error" class="mt-1 text-xs text-red-600">{{ error }}</p>

    <!-- Preview modal for the newly selected file -->
    <AppModal :show="showPreview" :title="selectedFile?.name || 'Pratinjau File'" size="lg" @close="closePreview">
      <div v-if="previewUrl" class="space-y-4">
        <!-- PDF: render inline -->
        <iframe
          v-if="isPdf"
          :src="previewUrl"
          class="w-full h-[70vh] rounded-lg border border-secondary-200 dark:border-secondary-700 bg-white"
          title="Pratinjau dokumen PDF"
        />
        <!-- Image: render inline -->
        <img
          v-else-if="isImage"
          :src="previewUrl"
          :alt="selectedFile?.name"
          class="max-w-full max-h-[70vh] mx-auto rounded-lg border border-secondary-200 dark:border-secondary-700"
        />
        <!-- Other formats: info card only -->
        <div
          v-else
          class="flex flex-col items-center justify-center h-64 text-center border border-secondary-200 dark:border-secondary-700 rounded-lg bg-secondary-50 dark:bg-secondary-900/50"
        >
          <svg class="w-12 h-12 text-secondary-400 dark:text-secondary-500 mb-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5"
              d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
          </svg>
          <p class="text-sm text-secondary-600 dark:text-secondary-400 mb-2">
            Pratinjau tidak tersedia untuk format ini.
          </p>
          <a
            :href="previewUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="text-sm text-blue-600 hover:underline font-medium"
          >
            Buka di tab baru
          </a>
        </div>
      </div>
      <div v-else class="flex items-center justify-center h-40 text-sm text-secondary-500 dark:text-secondary-400">
        Membuat pratinjau...
      </div>
    </AppModal>
  </div>
</template>

<script setup>
import { ref, computed, onBeforeUnmount } from 'vue'
import AppIconButton from '../common/AppIconButton.vue'
import AppButton from '../common/AppButton.vue'
import AppModal from '../common/AppModal.vue'

const ALLOWED_TYPES = [
  'application/pdf',
  'application/msword',
  'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
  'application/vnd.ms-excel',
  'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
]

const ALLOWED_EXTENSIONS = ['.pdf', '.doc', '.docx', '.xls', '.xlsx']

const props = defineProps({
  maxMB: { type: Number, default: 50 },
  existingUrl: { type: String, default: null },
})

const emit = defineEmits(['change'])

const fileInput = ref(null)
const isDragging = ref(false)
const selectedFile = ref(null)
const error = ref('')
const showPreview = ref(false)
const previewUrl = ref('')

const acceptAttr = ALLOWED_EXTENSIONS.join(',')

const isPdf = computed(() => selectedFile.value?.type === 'application/pdf' || selectedFile.value?.name?.toLowerCase().endsWith('.pdf'))
const isImage = computed(() => selectedFile.value?.type?.startsWith('image/'))

function openPreview() {
  if (!selectedFile.value) return
  if (previewUrl.value) URL.revokeObjectURL(previewUrl.value)
  previewUrl.value = URL.createObjectURL(selectedFile.value)
  showPreview.value = true
}

function closePreview() {
  showPreview.value = false
  if (previewUrl.value) {
    URL.revokeObjectURL(previewUrl.value)
    previewUrl.value = ''
  }
}

const formattedSize = computed(() => {
  if (!selectedFile.value) return ''
  const bytes = selectedFile.value.size
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
})

function isValidType(file) {
  if (ALLOWED_TYPES.includes(file.type)) return true
  // Fallback: check extension for cases where MIME is generic
  const ext = '.' + file.name.split('.').pop().toLowerCase()
  return ALLOWED_EXTENSIONS.includes(ext)
}

function processFile(file) {
  error.value = ''

  if (!isValidType(file)) {
    error.value = 'Format file tidak didukung. Gunakan PDF, DOC, DOCX, XLS, atau XLSX.'
    return false
  }

  const maxBytes = props.maxMB * 1024 * 1024
  if (file.size > maxBytes) {
    error.value = `Ukuran file melebihi batas ${props.maxMB}MB.`
    return false
  }

  selectedFile.value = file
  emit('change', file)
  return true
}

function onFileChange(e) {
  const file = e.target.files[0]
  if (!file) return
  processFile(file)
  // Reset input so same file can be re-selected
  e.target.value = ''
}

function onDrop(e) {
  isDragging.value = false
  const file = e.dataTransfer.files[0]
  if (!file) return
  processFile(file)
}

function clearFile() {
  selectedFile.value = null
  error.value = ''
  if (previewUrl.value) {
    URL.revokeObjectURL(previewUrl.value)
    previewUrl.value = ''
  }
  showPreview.value = false
  emit('change', null)
}

onBeforeUnmount(() => {
  if (previewUrl.value) URL.revokeObjectURL(previewUrl.value)
})

defineExpose({ clearFile })
</script>
