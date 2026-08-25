<template>
  <div class="max-w-4xl mx-auto">
    <div class="flex items-center gap-3 mb-6">
      <AppBackButton />
      <h1 class="text-2xl font-bold text-secondary-800 dark:text-secondary-100">{{ isEdit ? 'Edit UMKM' : 'Tambah UMKM' }}</h1>
    </div>

    <div class="bg-white dark:bg-secondary-800 rounded-lg shadow-sm border border-secondary-200 dark:border-secondary-700 p-6 space-y-4">
      <AppInput v-model="form.name" label="Nama Usaha" :error="errors.name" placeholder="Nama usaha" />
      <AppInput v-model="form.owner" label="Pemilik" :error="errors.owner" placeholder="Nama pemilik" />
      <AppInput v-model="form.address" label="Alamat" :error="errors.address" placeholder="Alamat usaha" />
      <AppInput v-model="form.phone" label="Telepon" :error="errors.phone" placeholder="08xxxxxxxxxx" />
      <AppInput v-model="form.email" label="Email" type="email" placeholder="email@contoh.com" />
      <AppInput v-model="form.website" label="Website" placeholder="https://..." />

      <AppCategoryPicker
        v-model="form.category"
        :suggestions="categorySuggestions"
        :service="umkmService"
        label="Kategori (Opsional)"
        placeholder="Pilih atau buat kategori..."
        @manage="showCategoryModal = true"
      />

      <div>
        <label class="block text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-1">Deskripsi (Opsional)</label>
        <textarea
          v-model="form.description"
          rows="3"
          placeholder="Deskripsi singkat usaha"
          class="w-full px-3 py-2 border border-secondary-300 dark:border-secondary-600 rounded text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 bg-white dark:bg-secondary-800 text-secondary-700 dark:text-secondary-200"
        />
      </div>

      <div>
        <label class="block text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-2">Gambar (Opsional)</label>
        <AppMultiImageUpload v-model="imagesState" />
      </div>

      <div class="flex gap-3 pt-2">
        <AppButton variant="primary" :loading="saving" @click="handleSubmit">
          {{ isEdit ? 'Simpan' : 'Tambah' }}
        </AppButton>
        <AppButton variant="secondary" @click="$router.back()">Batal</AppButton>
      </div>
    </div>

    <CategoryManagerModal
      :show="showCategoryModal"
      :service="umkmService"
      title="Kelola Kategori UMKM"
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
import CategoryManagerModal from '../../components/common/CategoryManagerModal.vue'
import AppCategoryPicker from '../../components/common/AppCategoryPicker.vue'
import AppMultiImageUpload from '../../components/common/AppMultiImageUpload.vue'
import { umkmService } from '../../services/umkm.service'
import { useNotificationStore } from '../../stores/notification'

const route = useRoute()
const router = useRouter()
const notificationStore = useNotificationStore()

const isEdit = computed(() => !!route.params.id)
const form = ref({ name: '', owner: '', address: '', phone: '', email: '', website: '', category: { id: '', name: '' }, description: '', images: [] })
const imagesState = ref([...form.value.images])
const errors = ref({})
const saving = ref(false)
const showCategoryModal = ref(false)
const categorySuggestions = ref([])

function onCategorySelected(category) {
  form.value.category = { id: category.id, name: category.name }
}

async function loadCategorySuggestions() {
  try {
    const res = await umkmService.getCategories({ limit: 100 })
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
  if (!form.value.name.trim()) errors.value.name = 'Nama usaha wajib diisi'
  if (!form.value.owner.trim()) errors.value.owner = 'Nama pemilik wajib diisi'
  if (!form.value.address.trim()) errors.value.address = 'Alamat wajib diisi'
  if (!form.value.phone.trim()) errors.value.phone = 'Telepon wajib diisi'
  else if (!/^[0-9+\-\s()]{7,20}$/.test(form.value.phone)) errors.value.phone = 'Format telepon tidak valid'
  return Object.keys(errors.value).length === 0
}

async function handleSubmit() {
  if (!validate()) return
  saving.value = true
  try {
    const categoryId = form.value.category?.id || ''
    const fd = new FormData()
    fd.append('name', form.value.name)
    fd.append('owner', form.value.owner)
    fd.append('address', form.value.address)
    fd.append('phone', form.value.phone)
    if (form.value.email) fd.append('email', form.value.email)
    if (form.value.website) fd.append('website', form.value.website)
    if (categoryId) fd.append('category', categoryId)
    if (form.value.description) fd.append('description', form.value.description)

    imagesState.value.forEach(img => {
      if (img instanceof File) {
        fd.append('images', img)
      } else {
        fd.append('existing_images', img)
      }
    })

    if (isEdit.value) {
      await umkmService.update(route.params.id, fd)
      notificationStore.success('UMKM diperbarui')
    } else {
      await umkmService.create(fd)
      notificationStore.success('UMKM ditambahkan')
    }
    router.push('/umkm')
  } catch (err) {
    if (!err.response) return
    notificationStore.error(err.response?.data?.error || 'Gagal menyimpan UMKM')
  } finally {
    saving.value = false
  }
}

onMounted(async () => {
  await loadCategorySuggestions()
  if (isEdit.value) {
    try {
      const res = await umkmService.get(route.params.id)
      const d = res.data || {}
      form.value = {
        name: d.name || '',
        owner: d.owner || '',
        address: d.address || '',
        phone: d.phone || '',
        email: d.email || '',
        website: d.website || '',
        category: { id: d.category_id || d.category?.id || '', name: d.category?.name || '' },
        description: d.description || '',
        images: d.images || []
      }
    } catch {
      notificationStore.error('Gagal memuat data UMKM')
      router.push('/umkm')
    }
  }
})
</script>
