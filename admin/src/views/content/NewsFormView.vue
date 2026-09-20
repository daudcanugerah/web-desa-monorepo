<template>
  <div class="max-w-5xl mx-auto">
    <div class="flex items-center gap-3 mb-6">
      <AppBackButton />
      <h1 class="text-2xl font-bold text-secondary-800 dark:text-secondary-100">{{ isEdit ? 'Edit Berita' : 'Tambah Berita' }}</h1>
    </div>

    <!-- Draft restored notice -->
    <div
      v-if="draftRestored && form.title"
      class="mb-4 flex items-center justify-between gap-3 px-4 py-3 rounded-lg border border-blue-200 bg-blue-50 dark:bg-blue-900/20 dark:border-blue-800"
    >
      <p class="text-sm text-blue-700 dark:text-blue-300">
        💾 Draft otomatis dipulihkan dari sesi sebelumnya.
        <span v-if="draftSavedAt" class="text-xs opacity-75">
          ({{ formatRelative(draftSavedAt) }})
        </span>
      </p>
      <button
        type="button"
        class="text-xs text-blue-700 dark:text-blue-300 hover:underline font-medium"
        @click="clearDraft"
      >
        Buang draft
      </button>
    </div>

    <div class="bg-white dark:bg-secondary-800 rounded-lg shadow-sm border border-secondary-200 dark:border-secondary-700 p-6 space-y-5">
      <AppInput v-model="form.title" label="Judul" :error="errors.title" placeholder="Judul berita" />

      <AppCategoryPicker
        v-model="form.category"
        :suggestions="categorySuggestions"
        :service="newsService"
        label="Kategori (Opsional)"
        placeholder="Pilih atau buat kategori..."
        @manage="showCategoryModal = true"
      />

      <div>
        <label class="block text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-1">Konten</label>
        <RichTextEditor v-model="form.content" />
        <p v-if="errors.content" class="mt-1 text-xs text-red-600">{{ errors.content }}</p>
      </div>

      <div>
        <label class="block text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-1">Gambar</label>
        <ImageUpload :existing-url="existingImageUrl" @change="onImageChange" />
      </div>

      <div class="flex gap-3 pt-2">
        <AppButton variant="primary" :loading="saving" @click="handleSubmit">
          {{ isEdit ? 'Simpan' : 'Publikasikan' }}
        </AppButton>
        <AppButton variant="secondary" @click="$router.back()">Batal</AppButton>
      </div>
    </div>

    <CategoryManagerModal
      :show="showCategoryModal"
      :service="newsService"
      title="Kelola Kategori Berita"
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
import AppBackButton from '../../components/common/AppBackButton.vue'
import AppCategoryPicker from '../../components/common/AppCategoryPicker.vue'
import ImageUpload from '../../components/forms/ImageUpload.vue'
import RichTextEditor from '../../components/forms/RichTextEditor.vue'
import CategoryManagerModal from '../../components/common/CategoryManagerModal.vue'
import { useDraftAutosave } from '../../composables/useDraftAutosave'
import { newsService } from '../../services/news.service'
import { useNotificationStore } from '../../stores/notification'
import { formatRelative } from '../../utils/dateFormat'

const route = useRoute()
const router = useRouter()
const notificationStore = useNotificationStore()

const isEdit = computed(() => !!route.params.id)
const existingImageUrl = ref(null)
const showCategoryModal = ref(false)
const categorySuggestions = ref([])

const form = ref({ title: '', content: '', category: { id: '', name: '' } })
const imageFile = ref(null)
const errors = ref({})
const saving = ref(false)

const { restored: draftRestored, lastSavedAt: draftSavedAt, clear: clearDraft } = useDraftAutosave(
  isEdit.value ? `news:${route.params.id}` : 'news:new',
  form,
)

function onImageChange(file) {
  imageFile.value = file
}

function onCategorySelected(category) {
  form.value.category = { id: category.id, name: category.name }
}

async function loadCategorySuggestions() {
  try {
    const res = await newsService.getCategories({ limit: 100 })
    const list = res.data?.categories || res.data || []
    categorySuggestions.value = (Array.isArray(list) ? list : [])
      .filter((c) => c && (c.id || c.name))
      .map((c) => ({ id: c.id, name: c.name }))
  } catch {
    categorySuggestions.value = []
  }
}

function validate() {
  errors.value = {}
  if (!form.value.title.trim()) errors.value.title = 'Judul wajib diisi'
  if (!form.value.content || form.value.content === '<p></p>') errors.value.content = 'Konten wajib diisi'
  return Object.keys(errors.value).length === 0
}

async function handleSubmit() {
  if (!validate()) return
  saving.value = true
  try {
    const categoryId = form.value.category?.id || ''
    const fd = new FormData()
    fd.append('title', form.value.title)
    fd.append('content', form.value.content)
    if (categoryId) fd.append('category', categoryId)
    if (imageFile.value) fd.append('image', imageFile.value)

    if (isEdit.value) {
      await newsService.update(route.params.id, fd)
      notificationStore.success('Berita diperbarui')
    } else {
      await newsService.create(fd)
      notificationStore.success('Berita dipublikasikan')
    }
    clearDraft()
    router.push('/berita')
  } catch (err) {
    if (!err.response) return
    notificationStore.error(err.response?.data?.error || 'Gagal menyimpan berita')
  } finally {
    saving.value = false
  }
}

onMounted(async () => {
  await loadCategorySuggestions()
  if (isEdit.value) {
    try {
      const res = await newsService.get(route.params.id)
      const article = res.data || {}
      form.value.title = article.title || ''
      form.value.content = article.content || ''
      const catId = article.category_id || article.category?.id || ''
      const catName = article.category?.name || ''
      form.value.category = { id: catId, name: catName }
      existingImageUrl.value = article.image_url
    } catch {
      notificationStore.error('Gagal memuat berita')
      router.push('/berita')
    }
  }
})
</script>
