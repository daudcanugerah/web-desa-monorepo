<template>
  <div>
    <div class="flex items-center gap-3 mb-6">
      <AppBackButton />
      <h1 class="text-2xl font-bold text-secondary-800 dark:text-secondary-100">
        {{ isEdit ? 'Edit Fasilitas' : 'Tambah Fasilitas' }}
      </h1>
    </div>

    <div class="max-w-4xl mx-auto space-y-6">
        <!-- Section: Info Dasar -->
        <section class="bg-white dark:bg-secondary-800 rounded-lg shadow-sm border border-secondary-200 dark:border-secondary-700">
          <header class="flex items-center gap-2 px-6 py-4 border-b border-secondary-200 dark:border-secondary-700 bg-secondary-50 dark:bg-secondary-800/60 rounded-t-lg">
            <svg class="w-5 h-5 text-secondary-500 dark:text-secondary-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
            </svg>
            <h2 class="text-sm font-semibold text-secondary-800 dark:text-secondary-100">Info Dasar</h2>
          </header>
          <div class="p-6 space-y-4">
            <AppInput v-model="form.name" label="Nama Fasilitas" :error="errors.name" placeholder="Nama fasilitas" />
            <AppCategoryPicker
              v-model="form.category"
              :suggestions="categorySuggestions"
              :service="facilityService"
              label="Kategori"
              :error="errors.category"
              placeholder="Pilih atau buat kategori..."
              @manage="showCategoryModal = true"
            />
            <div>
              <label class="block text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-1.5">Deskripsi</label>
              <textarea
                v-model="form.description"
                rows="3"
                placeholder="Deskripsi fasilitas"
                class="w-full px-3 py-2 border border-secondary-300 dark:border-secondary-600 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-primary-500 bg-white dark:bg-secondary-900 text-secondary-700 dark:text-secondary-200"
              />
            </div>
          </div>
        </section>

        <!-- Section: Lokasi -->
        <section class="bg-white dark:bg-secondary-800 rounded-lg shadow-sm border border-secondary-200 dark:border-secondary-700 overflow-hidden">
          <header class="flex items-center gap-2 px-6 py-4 border-b border-secondary-200 dark:border-secondary-700 bg-secondary-50 dark:bg-secondary-800/60">
            <svg class="w-5 h-5 text-secondary-500 dark:text-secondary-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z" />
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 11a3 3 0 11-6 0 3 3 0 016 0z" />
            </svg>
            <h2 class="text-sm font-semibold text-secondary-800 dark:text-secondary-100">Lokasi</h2>
            <span class="ml-auto text-xs text-secondary-500 dark:text-secondary-400">
              Klik peta atau isi koordinat
            </span>
          </header>
          <div class="p-6 space-y-4">
            <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <AppInput
                v-model="form.latitude"
                label="Latitude"
                :error="errors.latitude"
                placeholder="-6.200000"
                @input="onCoordInput"
              />
              <AppInput
                v-model="form.longitude"
                label="Longitude"
                :error="errors.longitude"
                placeholder="106.816000"
                @input="onCoordInput"
              />
            </div>
            <div class="relative rounded-lg overflow-hidden border border-secondary-200 dark:border-secondary-700">
              <MapPicker
                :lat="mapLat"
                :lng="mapLng"
                @update:lat="(v) => { form.latitude = String(v); errors.latitude = '' }"
                @update:lng="(v) => { form.longitude = String(v); errors.longitude = '' }"
              />
              <button
                type="button"
                class="absolute top-2 right-2 z-[500] px-2.5 py-1.5 rounded-lg bg-white dark:bg-secondary-800 text-xs font-medium text-secondary-700 dark:text-secondary-200 shadow hover:bg-secondary-50 dark:hover:bg-secondary-700 transition-colors"
                @click="showMapModal = true"
              >
                ⛶ Perbesar
              </button>
            </div>
          </div>
        </section>

        <!-- Section: Gambar -->
        <section class="bg-white dark:bg-secondary-800 rounded-lg shadow-sm border border-secondary-200 dark:border-secondary-700 overflow-hidden">
          <header class="flex items-center gap-2 px-6 py-4 border-b border-secondary-200 dark:border-secondary-700 bg-secondary-50 dark:bg-secondary-800/60">
            <svg class="w-5 h-5 text-secondary-500 dark:text-secondary-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
            </svg>
            <h2 class="text-sm font-semibold text-secondary-800 dark:text-secondary-100">Gambar</h2>
            <span class="ml-auto text-xs text-secondary-500 dark:text-secondary-400">Opsional</span>
          </header>
          <div class="p-6 space-y-3">
            <!-- Existing + new image thumbnails -->
            <div v-if="allImages.length" class="flex gap-2 flex-wrap">
              <button
                v-for="(img, idx) in allImages"
                :key="img.key"
                type="button"
                class="relative w-24 h-24 border border-secondary-300 dark:border-secondary-600 rounded-lg overflow-hidden group focus:outline-none focus:ring-2 focus:ring-primary-500"
                :title="`Lihat gambar ${idx + 1}`"
                @click="openLightbox(idx)"
              >
                <SafeImg
                  :src="img.src"
                  :alt="`Image ${idx + 1}`"
                  class="w-full h-full group-hover:scale-105 transition-transform duration-200"
                  :lazy="false"
                />
                <span class="absolute inset-0 bg-black/0 group-hover:bg-black/20 transition-colors flex items-center justify-center opacity-0 group-hover:opacity-100">
                  <svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0zm5-6a3 3 0 01-3 3m3-3a3 3 0 00-3 3m3-3v6m-9-9a3 3 0 013 3m-3-3a3 3 0 00-3 3m3-3h6" />
                  </svg>
                </span>
                <AppIconButton
                  type="button"
                  variant="danger"
                  size="sm"
                  class="absolute top-1 right-1"
                  :title="img.isNew ? 'Hapus gambar baru' : 'Hapus gambar'"
                  @click.stop="removeImage(img)"
                >
                  <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </AppIconButton>
              </button>
            </div>

            <!-- Drop zone -->
            <label
              class="flex flex-col items-center justify-center gap-1.5 border-2 border-dashed rounded-lg py-6 px-4 text-center cursor-pointer transition-colors"
              :class="isImageDragging
                ? 'border-primary-500 bg-primary-50 dark:bg-primary-900/20'
                : 'border-secondary-300 dark:border-secondary-600 hover:border-primary-500 hover:bg-primary-50 dark:hover:bg-primary-900/10'"
              @dragover.prevent="isImageDragging = true"
              @dragleave.prevent="isImageDragging = false"
              @drop.prevent="onImagesDrop"
            >
              <svg class="w-8 h-8 text-secondary-400 dark:text-secondary-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
              </svg>
              <span class="text-sm text-secondary-600 dark:text-secondary-300">
                Klik atau seret gambar ke sini
              </span>
              <span class="text-xs text-secondary-400 dark:text-secondary-500">JPEG, PNG, WebP — beberapa file</span>
              <input
                ref="imageInputRef"
                type="file"
                multiple
                accept="image/*"
                class="hidden"
                @change="onImagesChange"
              />
            </label>
          </div>
        </section>

        <!-- Actions -->
        <div class="flex gap-3 pt-2">
          <AppButton variant="primary" :loading="saving" @click="handleSubmit">
            {{ isEdit ? 'Simpan Perubahan' : 'Tambah Fasilitas' }}
          </AppButton>
          <AppButton variant="secondary" @click="$router.back()">Batal</AppButton>
        </div>
    </div>

    <!-- Fullscreen Map Modal -->
    <Teleport to="body">
      <div
        v-if="showMapModal"
        class="fixed inset-0 z-[200] flex items-stretch bg-black/70"
        @click.self="showMapModal = false"
      >
        <div class="relative w-full h-full bg-white dark:bg-secondary-800 flex flex-col">
          <div class="flex items-center justify-between px-6 py-4 border-b border-secondary-200 dark:border-secondary-700 shrink-0">
            <h3 class="text-lg font-semibold text-secondary-800 dark:text-secondary-200">Pilih Lokasi</h3>
            <AppIconButton variant="ghost" size="md" @click="showMapModal = false">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </AppIconButton>
          </div>
          <div class="flex-1 p-4">
            <div class="w-full h-full rounded-lg overflow-hidden">
              <MapPicker
                :lat="mapLat"
                :lng="mapLng"
                @update:lat="(v) => { form.latitude = String(v); errors.latitude = '' }"
                @update:lng="(v) => { form.longitude = String(v); errors.longitude = '' }"
              />
            </div>
          </div>
        </div>
      </div>
    </Teleport>

    <CategoryManagerModal
      :show="showCategoryModal"
      :service="facilityService"
      title="Kelola Kategori Fasilitas"
      @close="showCategoryModal = false"
      @selected="onCategorySelected"
    />

    <!-- Image lightbox -->
    <AppModal
      :show="showLightbox"
      :title="`Gambar ${lightboxIndex + 1} dari ${allImages.length}`"
      size="lg"
      @close="closeLightbox"
    >
      <div v-if="allImages[lightboxIndex]" class="flex flex-col items-center">
        <SafeImg
          :src="allImages[lightboxIndex].src"
          :alt="`Image ${lightboxIndex + 1}`"
          class="max-w-full max-h-[70vh] object-contain rounded-lg"
          :lazy="false"
        />
        <div v-if="allImages.length > 1" class="flex items-center gap-3 mt-4">
          <AppButton variant="secondary" size="sm" :disabled="lightboxIndex === 0" @click="lightboxIndex--">
            ← Sebelumnya
          </AppButton>
          <AppButton variant="secondary" size="sm" :disabled="lightboxIndex >= allImages.length - 1" @click="lightboxIndex++">
            Berikutnya →
          </AppButton>
        </div>
      </div>
    </AppModal>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import AppInput from '../../components/common/AppInput.vue'
import AppButton from '../../components/common/AppButton.vue'
import AppBackButton from '../../components/common/AppBackButton.vue'
import AppIconButton from '../../components/common/AppIconButton.vue'
import AppModal from '../../components/common/AppModal.vue'
import SafeImg from '../../components/common/SafeImg.vue'
import MapPicker from '../../components/forms/MapPicker.vue'
import CategoryManagerModal from '../../components/common/CategoryManagerModal.vue'
import AppCategoryPicker from '../../components/common/AppCategoryPicker.vue'
import { getImageUrl } from '../../utils/imageUrl'
import { useNotificationStore } from '../../stores/notification'
import { facilityService } from '../../services/facility.service'

const route = useRoute()
const router = useRouter()
const notificationStore = useNotificationStore()

const isEdit = computed(() => !!route.params.id)
const form = ref({ name: '', category: { id: '', name: '' }, description: '', latitude: '', longitude: '' })
const existingImages = ref([])
const showCategoryModal = ref(false)
const categorySuggestions = ref([])

function onCategorySelected(category) {
  form.value.category = { id: category.id, name: category.name }
}

async function loadCategorySuggestions() {
  try {
    const res = await facilityService.getCategories({ limit: 100 })
    const list = res.data?.categories || res.data || []
    categorySuggestions.value = (Array.isArray(list) ? list : [])
      .filter((c) => c && (c.id || c.name))
      .map((c) => ({ id: c.id, name: c.name }))
  } catch {
    categorySuggestions.value = []
  }
}
const newImages = ref([])
const newImageUrls = new Map() // File -> objectURL, revoked on remove/unmount
const imageInputRef = ref(null)
const errors = ref({})
const saving = ref(false)
const showMapModal = ref(false)
const showLightbox = ref(false)
const lightboxIndex = ref(0)
const isImageDragging = ref(false)

// Separate reactive coords for MapPicker to avoid NaN on empty string
const mapLat = computed(() => parseFloat(form.value.latitude) || -6.2)
const mapLng = computed(() => parseFloat(form.value.longitude) || 106.816)

// Combined view of existing (server) + new (local File) images for the grid/lightbox.
const allImages = computed(() => {
  const existing = existingImages.value.map((img) => ({
    key: 'e-' + (typeof img === 'string' ? img : img.url || img.id),
    src: typeof img === 'string' ? getImageUrl(img) : getImageUrl(img.url || img.thumbnail_url || ''),
    isNew: false,
    ref: img,
  }))
  const fresh = newImages.value.map((file) => ({
    key: 'n-' + (newImageUrls.get(file) || file.name),
    src: newImageUrls.get(file) || '',
    isNew: true,
    ref: file,
  }))
  return [...existing, ...fresh]
})

function getFileUrl(file) {
  if (!newImageUrls.has(file)) {
    newImageUrls.set(file, URL.createObjectURL(file))
  }
  return newImageUrls.get(file)
}

function onCoordInput() {
  errors.value.latitude = ''
  errors.value.longitude = ''
}

function addFiles(files) {
  const imgs = Array.from(files || []).filter((f) => f.type.startsWith('image/'))
  if (!imgs.length) return
  imgs.forEach((f) => getFileUrl(f))
  newImages.value.push(...imgs)
}

function onImagesChange(e) {
  addFiles(e.target.files)
  // Reset input so same files can be re-selected
  e.target.value = ''
}

function onImagesDrop(e) {
  isImageDragging.value = false
  addFiles(e.dataTransfer.files)
}

function openLightbox(idx) {
  lightboxIndex.value = idx
  showLightbox.value = true
}

function closeLightbox() {
  showLightbox.value = false
}

function removeImage(img) {
  if (img.isNew) {
    const idx = newImages.value.indexOf(img.ref)
    if (idx > -1) {
      newImages.value.splice(idx, 1)
      const url = newImageUrls.get(img.ref)
      if (url) {
        URL.revokeObjectURL(url)
        newImageUrls.delete(img.ref)
      }
    }
    return
  }
  // Existing image: remove from local list (index maps to existingImages order).
  const idx = existingImages.value.findIndex((x) => x === img.ref)
  if (idx === -1) return
  removeExistingImage(idx)
}

async function removeExistingImage(idx) {
  if (!isEdit.value) {
    existingImages.value.splice(idx, 1)
    return
  }
  const removed = existingImages.value[idx]
  existingImages.value.splice(idx, 1)
  try {
    await facilityService.deleteImage(route.params.id, idx)
    notificationStore.success('Gambar dihapus')
  } catch (err) {
    existingImages.value.splice(idx, 0, removed)
    notificationStore.error(err.response?.data?.error || 'Gagal menghapus gambar')
  }
}

function validate() {
  errors.value = {}
  if (!form.value.name.trim()) errors.value.name = 'Nama wajib diisi'
  const lat = parseFloat(form.value.latitude)
  const lng = parseFloat(form.value.longitude)
  if (isNaN(lat) || lat < -90 || lat > 90) errors.value.latitude = 'Latitude harus antara -90 dan 90'
  if (isNaN(lng) || lng < -180 || lng > 180) errors.value.longitude = 'Longitude harus antara -180 dan 180'
  return Object.keys(errors.value).length === 0
}

async function handleSubmit() {
  if (!validate()) return
  saving.value = true
  try {
    const categoryId = form.value.category?.id || ''

    // Upload new image files first (single-shot endpoint), collecting the
    // returned media ids. Existing images are referenced by their media_id.
    const uploadedIds = []
    for (const file of newImages.value) {
      const res = await facilityService.uploadMedia(file)
      const mediaId = res.data?.media_id
      if (mediaId) uploadedIds.push(mediaId)
    }

    const existingIds = existingImages.value
      .map((img) => (typeof img === 'string' ? null : img.media_id || img.id))
      .filter(Boolean)

    const fd = new FormData()
    fd.append('name', form.value.name)
    if (categoryId) fd.append('category', categoryId)
    if (form.value.description) fd.append('description', form.value.description)
    fd.append('latitude', parseFloat(form.value.latitude))
    fd.append('longitude', parseFloat(form.value.longitude))

    // The backend replaces the image set on update, so send every retained
    // image (existing + newly uploaded) as images_media_ids.
    ;[...existingIds, ...uploadedIds].forEach((id) => fd.append('images_media_ids', id))

    if (isEdit.value) {
      await facilityService.update(route.params.id, fd)
      notificationStore.success('Fasilitas diperbarui')
    } else {
      await facilityService.create(fd)
      notificationStore.success('Fasilitas ditambahkan')
    }
    router.push('/fasilitas')
  } catch (err) {
    if (!err.response) return
    notificationStore.error(err.response?.data?.error || 'Gagal menyimpan fasilitas')
  } finally {
    saving.value = false
  }
}

onMounted(async () => {
  await loadCategorySuggestions()
  if (isEdit.value) {
    try {
      const res = await facilityService.get(route.params.id)
      const d = res.data
      form.value = {
        name: d.name || '',
        category: { id: d.category_id || d.category?.id || '', name: d.category?.name || '' },
        description: d.description || '',
        latitude: String(d.latitude),
        longitude: String(d.longitude),
      }
      existingImages.value = d.media || d.images || []
    } catch {
      notificationStore.error('Gagal memuat fasilitas')
      router.push('/fasilitas')
    }
  }
})

onBeforeUnmount(() => {
  newImageUrls.forEach((url) => URL.revokeObjectURL(url))
  newImageUrls.clear()
})
</script>
