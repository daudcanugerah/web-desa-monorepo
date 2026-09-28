<template>
  <div class="max-w-6xl mx-auto">
    <div class="flex items-center gap-3 mb-6">
      <AppBackButton />
      <h1 class="text-2xl font-bold text-secondary-800 dark:text-secondary-100">
        {{ isEdit ? 'Edit Banner' : 'Tambah Banner' }}
      </h1>
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-5 gap-6">
      <!-- Form column -->
      <div class="lg:col-span-3 bg-white dark:bg-secondary-800 rounded-lg shadow-sm border border-secondary-200 dark:border-secondary-700 p-6 space-y-5">
        <AppInput v-model="form.title" label="Title" :error="errors.title" placeholder="Banner title" />

        <AppInput v-model="form.link" label="Link (Optional)" placeholder="https://example.com" />

        <AppCategoryPicker
          v-model="form.category"
          :suggestions="categoryInput.suggestions.value"
          :service="bannerService"
          label="Kategori (Opsional)"
          placeholder="Pilih atau buat kategori..."
          @manage="showCategoryModal = true"
        />

        <div class="flex items-center gap-3">
          <span class="text-sm font-medium text-secondary-700 dark:text-secondary-300">Status</span>
          <AppToggle
            :model-value="form.status === 'active'"
            :disabled="updatingStatus"
            :label="form.status === 'active' ? 'Aktif' : 'Nonaktif'"
            aria-label="Toggle banner status"
            @update:model-value="toggleStatus"
          />
        </div>

        <div class="flex gap-3 pt-2">
          <AppButton variant="primary" :loading="saving" :loading-label="isEdit ? 'Menyimpan...' : 'Membuat...'" @click="handleSubmit">
            {{ isEdit ? 'Simpan Perubahan' : 'Buat Banner' }}
          </AppButton>
          <AppButton variant="secondary" @click="$router.back()">Batal</AppButton>
        </div>
      </div>

      <!-- Image selection column -->
      <div class="lg:col-span-2 space-y-3">
        <div class="flex items-center justify-between">
          <h3 class="text-sm font-semibold text-secondary-800 dark:text-secondary-200">Gambar Banner</h3>
          <span
            v-if="hasUnsavedImage"
            class="inline-flex items-center gap-1 px-2 py-0.5 text-xs font-medium rounded-full bg-primary-100 text-primary-700 dark:bg-primary-900/30 dark:text-primary-300"
          >
            Gambar baru dipilih
          </span>
        </div>

        <!-- Image box -->
        <div
          class="relative aspect-video rounded-lg overflow-hidden border-2 border-dashed transition-colors"
          :class="boxClasses"
          @click="triggerFileInput"
          @dragover.prevent="isDragging = true"
          @dragleave.prevent="isDragging = false"
          @drop.prevent="onDrop"
        >
          <!-- New uploaded file (pre-save preview) -->
          <template v-if="imageFile">
            <SafeImg
              :src="previewUrl"
              alt="Pratinjau gambar baru"
              class="w-full h-full object-cover"
            />
            <div class="absolute top-2 left-2 px-2 py-0.5 text-xs font-medium rounded bg-primary-500/90 text-white">
              Gambar baru
            </div>
            <div class="absolute inset-0 bg-black/0 group-hover:bg-black/40 transition-colors flex items-center justify-center gap-2 opacity-0 hover:opacity-100">
              <button
                type="button"
                class="px-3 py-1.5 rounded bg-white text-secondary-800 text-xs font-medium shadow hover:bg-secondary-100"
                title="Edit & crop gambar"
                @click.stop="openImageEditor"
              >
                ✂️ Edit
              </button>
              <button
                type="button"
                class="px-3 py-1.5 rounded bg-white text-secondary-800 text-xs font-medium shadow hover:bg-secondary-100"
                @click.stop="triggerFileInput"
              >
                Ganti
              </button>
              <button
                type="button"
                class="px-3 py-1.5 rounded bg-danger-600 text-white text-xs font-medium shadow hover:bg-danger-700"
                @click.stop="clearNewImage"
              >
                Batal
              </button>
            </div>
          </template>

          <!-- Existing image (server) -->
          <template v-else-if="displayImageUrl">
            <SafeImg
              :src="displayImageUrl"
              :alt="form.title || 'Banner'"
              class="w-full h-full object-cover"
            />
            <div v-if="selectedMediaId" class="absolute top-2 left-2 px-2 py-0.5 text-xs font-medium rounded bg-secondary-700/90 text-white">
              Dari galeri
            </div>
            <div class="absolute inset-0 bg-black/0 group-hover:bg-black/40 transition-colors flex items-center justify-center gap-2 opacity-0 hover:opacity-100">
              <button
                type="button"
                class="px-3 py-1.5 rounded bg-white text-secondary-800 text-xs font-medium shadow hover:bg-secondary-100"
                title="Unggah file baru"
                @click.stop="triggerFileInput"
              >
                Unggah baru
              </button>
              <button
                type="button"
                class="px-3 py-1.5 rounded bg-white text-secondary-800 text-xs font-medium shadow hover:bg-secondary-100"
                title="Pilih dari galeri"
                @click.stop="openGalleryPicker"
              >
                Pilih dari galeri
              </button>
            </div>
          </template>

          <!-- Empty state -->
          <template v-else>
            <div class="absolute inset-0 flex flex-col items-center justify-center text-secondary-500 dark:text-secondary-400">
              <svg class="w-12 h-12 mb-3" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
              </svg>
              <p class="text-sm font-medium">Pilih gambar banner</p>
              <p class="text-xs mt-1 text-secondary-400 dark:text-secondary-500">Unggah file baru atau pilih dari galeri</p>
            </div>
          </template>
        </div>

        <!-- Action row when empty -->
        <div v-if="!displayImageUrl && !imageFile" class="flex flex-col sm:flex-row gap-2">
          <AppButton variant="primary" size="sm" @click="triggerFileInput">
            Unggah File
          </AppButton>
          <AppButton variant="secondary" size="sm" @click="openGalleryPicker">
            Pilih dari Galeri
          </AppButton>
        </div>

        <!-- Reset / pick-again row when has existing image -->
        <div v-else-if="displayImageUrl && !imageFile" class="flex flex-col sm:flex-row gap-2">
          <AppButton variant="secondary" size="sm" @click="triggerFileInput">
            Unggah File Baru
          </AppButton>
          <AppButton variant="ghost" size="sm" @click="openGalleryPicker">
            Ganti dari Galeri
          </AppButton>
        </div>

        <input
          ref="fileInput"
          type="file"
          accept="image/jpeg,image/png,image/webp"
          class="hidden"
          @change="onFileChange"
        />

        <!-- Warnings / errors -->
        <p v-if="warning" class="text-xs text-warning-700 dark:text-warning-300">
          ⚠ {{ warning }}
        </p>
        <p v-if="errors.image" class="text-xs text-danger-600 dark:text-danger-400">
          {{ errors.image }}
        </p>

        <!-- Format info -->
        <div class="bg-secondary-50 dark:bg-secondary-900/50 rounded-lg p-3 border border-secondary-200 dark:border-secondary-700">
          <h4 class="text-xs font-semibold text-secondary-700 dark:text-secondary-300 mb-1">Spesifikasi</h4>
          <ul class="text-xs text-secondary-600 dark:text-secondary-400 space-y-0.5">
            <li>• Rasio rekomendasi: 16:9</li>
            <li>• Minimal 1280×720 piksel</li>
            <li>• Format: JPEG, PNG, atau WebP</li>
            <li>• Maksimal 10MB</li>
          </ul>
        </div>
      </div>
    </div>

    <CategoryManagerModal
      :show="showCategoryModal"
      :service="bannerService"
      title="Kelola Kategori Banner"
      @close="showCategoryModal = false"
      @selected="onCategorySelected"
    />

    <GalleryMediaPickerModal
      :show="showGalleryPicker"
      @close="showGalleryPicker = false"
      @select="onMediaPicked"
    />

    <AppImageEditor
      :show="showImageEditor"
      :src="editorSource"
      :original-name="imageFile?.name || 'banner'"
      :aspect-ratio="16 / 9"
      title="Edit Gambar Banner"
      hint="Rasio 16:9 terkunci. Seret untuk memindah area, gunakan tombol untuk rotasi dan zoom."
      @update:show="showImageEditor = $event"
      @crop="onImageCropped"
    />
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter, onBeforeRouteLeave } from 'vue-router'
import AppInput from '../../components/common/AppInput.vue'
import AppButton from '../../components/common/AppButton.vue'
import AppBackButton from '../../components/common/AppBackButton.vue'
import AppToggle from '../../components/common/AppToggle.vue'
import AppCategoryPicker from '../../components/common/AppCategoryPicker.vue'
import SafeImg from '../../components/common/SafeImg.vue'
import CategoryManagerModal from '../../components/common/CategoryManagerModal.vue'
import GalleryMediaPickerModal from '../../components/common/GalleryMediaPickerModal.vue'
import AppImageEditor from '../../components/common/AppImageEditor.vue'
import { useCategoryInput } from '../../composables/useCategoryInput'
import { useImagePreview } from '../../composables/useImagePreview'
import { bannerService } from '../../services/banner.service'
import { useNotificationStore } from '../../stores/notification'

const route = useRoute()
const router = useRouter()
const notificationStore = useNotificationStore()

const isEdit = computed(() => !!route.params.id)
const existingImageUrl = ref(null)
const existingMediaId = ref(null)
const imageFile = ref(null)
const previewUrl = ref('')
const fileInput = ref(null)
const isDragging = ref(false)
const showCategoryModal = ref(false)
const showGalleryPicker = ref(false)
const selectedMediaId = ref(null)
const showImageEditor = ref(false)
const editorSource = ref('')
const pendingFile = ref(null)

const form = ref({ title: '', link: '', status: 'inactive', category: { id: '', name: '' } })
const errors = ref({})
const saving = ref(false)
const updatingStatus = ref(false)

const categoryInput = useCategoryInput(bannerService)
const { warning, processFile, reset } = useImagePreview({ maxMB: 10, warnMB: 5 })

const displayImageUrl = computed(() => {
  if (imageFile.value) return previewUrl.value
  if (selectedMediaId.value) return existingImageUrl.value
  return existingImageUrl.value
})

const hasUnsavedImage = computed(() => !!imageFile.value || (selectedMediaId.value && selectedMediaId.value !== existingMediaId.value))

const boxClasses = computed(() => {
  if (isDragging.value) {
    return 'border-primary-500 bg-primary-50 dark:bg-primary-900/20 cursor-pointer'
  }
  if (imageFile.value || displayImageUrl.value) {
    return 'border-transparent cursor-pointer group'
  }
  return 'border-secondary-300 dark:border-secondary-600 hover:border-secondary-400 dark:hover:border-secondary-500 cursor-pointer bg-secondary-50 dark:bg-secondary-900/50'
})

function triggerFileInput() {
  fileInput.value?.click()
}

function onFileChange(e) {
  const file = e.target.files?.[0]
  if (!file) return
  acceptFile(file)
  e.target.value = ''
}

function onDrop(e) {
  isDragging.value = false
  const file = e.dataTransfer.files?.[0]
  if (!file) return
  acceptFile(file)
}

function acceptFile(file) {
  if (!processFile(file)) return
  // Open the crop editor so the user can adjust before committing.
  pendingFile.value = file
  editorSource.value = URL.createObjectURL(file)
  showImageEditor.value = true
}

function openImageEditor() {
  if (!imageFile.value) return
  pendingFile.value = imageFile.value
  if (previewUrl.value) editorSource.value = previewUrl.value
  else editorSource.value = URL.createObjectURL(imageFile.value)
  showImageEditor.value = true
}

function onCategorySelected(category) {
  form.value.category = { id: category.id, name: category.name }
}

function onImageCropped(croppedFile) {
  // Replace staged file with the cropped version
  if (previewUrl.value) URL.revokeObjectURL(previewUrl.value)
  if (editorSource.value) URL.revokeObjectURL(editorSource.value)
  previewUrl.value = URL.createObjectURL(croppedFile)
  imageFile.value = croppedFile
  selectedMediaId.value = null
  pendingFile.value = null
  errors.value.image = ''
  notificationStore.success('Gambar banner diperbarui')
}

function clearNewImage() {
  if (previewUrl.value) URL.revokeObjectURL(previewUrl.value)
  if (editorSource.value) URL.revokeObjectURL(editorSource.value)
  previewUrl.value = ''
  editorSource.value = ''
  imageFile.value = null
  pendingFile.value = null
  selectedMediaId.value = existingMediaId.value
  reset()
}

function openGalleryPicker() {
  showGalleryPicker.value = true
}

function onMediaPicked(media) {
  selectedMediaId.value = media.id
  existingImageUrl.value = media.thumbnail_url || media.content_url
  if (imageFile.value) {
    if (previewUrl.value) URL.revokeObjectURL(previewUrl.value)
    if (editorSource.value) URL.revokeObjectURL(editorSource.value)
    previewUrl.value = ''
    editorSource.value = ''
    imageFile.value = null
    pendingFile.value = null
    reset()
  }
  errors.value.image = ''
  notificationStore.success('Media dipilih dari galeri')
}

function validate() {
  errors.value = {}
  if (!form.value.title.trim()) errors.value.title = 'Title is required'
  if (!isEdit.value && !imageFile.value && !selectedMediaId.value) {
    errors.value.image = 'Image is required'
  }
  return Object.keys(errors.value).length === 0
}

async function toggleStatus() {
  const newStatus = form.value.status === 'active' ? 'inactive' : 'active'

  // On create there is no id yet — hold the choice locally and send it with
  // the create request. On edit, persist immediately via PATCH /status.
  if (!isEdit.value) {
    form.value.status = newStatus
    return
  }

  updatingStatus.value = true
  try {
    await bannerService.updateStatus(route.params.id, newStatus)
    form.value.status = newStatus
    notificationStore.success(`Banner ${newStatus === 'active' ? 'diaktifkan' : 'dinonaktifkan'}`)
  } catch (err) {
    notificationStore.error(err.response?.data?.error || 'Gagal memperbarui status')
  } finally {
    updatingStatus.value = false
  }
}

async function handleSubmit() {
  if (!validate()) return
  saving.value = true
  try {
    const categoryId = form.value.category?.id || ''
    const fd = new FormData()
    fd.append('title', form.value.title)
    if (form.value.link) fd.append('link', form.value.link)
    if (categoryId) fd.append('category', categoryId)
    if (imageFile.value) {
      fd.append('image', imageFile.value)
    } else if (selectedMediaId.value) {
      fd.append('image_media_id', selectedMediaId.value)
    }
    // Status is settable on create; on edit it changes via PATCH /status.
    if (!isEdit.value) {
      fd.append('status', form.value.status || 'inactive')
    }
    // Note: status is NOT part of PUT /banners/{id} per swagger ("Status is NOT
    // changed here — use PATCH /status"). Status changes go through the toggle.

    if (isEdit.value) {
      await bannerService.update(route.params.id, fd)
      notificationStore.success('Banner berhasil diperbarui')
    } else {
      await bannerService.create(fd)
      notificationStore.success('Banner berhasil dibuat')
    }
    router.push('/banners')
  } catch (err) {
    if (!err.response) return
    notificationStore.error(err.response?.data?.error || 'Gagal menyimpan banner')
  } finally {
    saving.value = false
  }
}

onBeforeRouteLeave(() => {
  if (previewUrl.value) URL.revokeObjectURL(previewUrl.value)
  if (editorSource.value) URL.revokeObjectURL(editorSource.value)
})

onMounted(async () => {
  await categoryInput.load()
  if (isEdit.value) {
    try {
      const res = await bannerService.get(route.params.id)
      const banner = res.data
      form.value.title = banner.title
      form.value.link = banner.link || ''
      form.value.status = banner.status
      form.value.category = { id: banner.category_id || banner.category?.id || '', name: banner.category?.name || '' }
      if (banner.media?.url) {
        existingImageUrl.value = banner.media.thumbnail_url || banner.media.url
        existingMediaId.value = banner.media.media_id || null
        selectedMediaId.value = existingMediaId.value
      } else if (banner.image_url) {
        existingImageUrl.value = banner.image_url
      }
    } catch (err) {
      notificationStore.error('Gagal memuat banner')
      router.push('/banners')
    }
  }
})
</script>