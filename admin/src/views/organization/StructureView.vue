<template>
  <div>
    <!-- Header -->
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-2xl font-bold text-secondary-800 dark:text-secondary-200">Struktur Organisasi</h1>
      <AppButton variant="primary" @click="openCreate">Tambah Anggota</AppButton>
    </div>

    <!-- Search -->
    <div class="mb-4">
      <AppSearchInput
        v-model="search"
        placeholder="Cari nama atau jabatan..."
        width="xs"
        @search="onSearch"
      />
    </div>

    <!-- Table -->
    <div class="bg-white dark:bg-secondary-800 rounded-lg shadow-sm border border-secondary-200 dark:border-secondary-700 overflow-hidden">
      <AppTable :columns="columns" :data="members" :loading="loading">
        <template #photo="{ row }">
          <SafeImg
            v-if="row.profile_image_url"
            :src="resolveMediaUrl(row.profile_image_url)"
            :alt="row.name"
            circle
            class="w-10 h-10"
          />
          <div
            v-else
            class="w-10 h-10 rounded-full bg-secondary-200 dark:bg-secondary-700 flex items-center justify-center text-secondary-400 dark:text-secondary-500 text-sm font-medium"
          >
            {{ row.name?.charAt(0)?.toUpperCase() }}
          </div>
        </template>

        <template #actions="{ row }">
          <div class="flex gap-2">
            <AppButton size="sm" variant="secondary" @click="openEdit(row)">Edit</AppButton>
            <AppButton size="sm" variant="danger" @click="handleDelete(row)">Hapus</AppButton>
          </div>
        </template>

        <template #empty>
          <AppEmptyState
            v-if="search"
            icon="search"
            title="Tidak ada anggota yang cocok"
            :hint="emptyHint"
          >
            <AppButton variant="ghost" @click="resetFilters">Reset pencarian</AppButton>
          </AppEmptyState>
          <AppEmptyState
            v-else
            icon="users"
            title="Belum ada anggota struktur"
            hint="Tambahkan anggota pertama untuk mulai menampilkan struktur organisasi desa."
          >
            <AppButton variant="primary" @click="openCreate">Tambah Anggota</AppButton>
          </AppEmptyState>
        </template>
      </AppTable>
    </div>

    <AppPagination
      v-if="pagination.total_pages > 1"
      :page="pagination.page"
      :total-pages="pagination.total_pages"
      class="mt-4"
      @change="loadPage"
    />

    <!-- Create / Edit Modal -->
    <AppModal
      :show="showModal"
      :title="editingId ? 'Edit Anggota' : 'Tambah Anggota'"
      size="lg"
      @close="closeModal"
    >
      <form @submit.prevent="handleSubmit">
        <div class="space-y-4">
          <AppInput
            v-model="form.name"
            label="Nama"
            required
            :error="formErrors.name"
            placeholder="Nama lengkap"
          />
          <AppInput
            v-model="form.position"
            label="Jabatan"
            placeholder="Jabatan / posisi"
          />
          <AppInput
            v-model="form.email"
            label="Email"
            type="email"
            placeholder="email@contoh.com"
          />
          <AppInput
            v-model="form.phone"
            label="Telepon"
            placeholder="08xxxxxxxxxx"
          />
          <div>
            <label class="block text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-1">Deskripsi</label>
            <textarea
              v-model="form.description"
              rows="3"
              placeholder="Deskripsi singkat..."
              class="w-full border border-secondary-300 dark:border-secondary-600 rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 bg-white dark:bg-secondary-800 text-secondary-700 dark:text-secondary-200"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-1">Foto Profil</label>
            <ImageUpload
              ref="imageUploadRef"
              :existing-url="editingId ? form.existingImageUrl : null"
              @change="onImageChange"
            />
          </div>
        </div>
      </form>

      <template #footer>
        <div class="flex justify-end gap-3">
          <AppButton variant="secondary" @click="closeModal">Batal</AppButton>
          <AppButton variant="primary" :loading="saving" @click="handleSubmit">
            {{ editingId ? 'Simpan Perubahan' : 'Tambah' }}
          </AppButton>
        </div>
      </template>
    </AppModal>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { resolveMediaUrl } from '../../utils/imageUrl'
import AppTable from '../../components/common/AppTable.vue'
import AppPagination from '../../components/common/AppPagination.vue'
import AppButton from '../../components/common/AppButton.vue'
import AppInput from '../../components/common/AppInput.vue'
import AppModal from '../../components/common/AppModal.vue'
import AppSearchInput from '../../components/common/AppSearchInput.vue'
import AppEmptyState from '../../components/common/AppEmptyState.vue'
import ImageUpload from '../../components/forms/ImageUpload.vue'
import SafeImg from '../../components/common/SafeImg.vue'
import { structureService } from '../../services/structure.service'
import { useNotificationStore } from '../../stores/notification'
import { useConfirm } from '../../composables/useConfirm'

const notificationStore = useNotificationStore()
const { confirm } = useConfirm()

const columns = [
  { key: 'photo', label: 'Foto', width: '70px' },
  { key: 'name', label: 'Nama' },
  { key: 'position', label: 'Jabatan' },
  { key: 'actions', label: '', width: '140px' },
]

const members = ref([])
const loading = ref(false)
const pagination = ref({ page: 1, total_pages: 1 })
const search = ref('')

const showModal = ref(false)
const editingId = ref(null)
const saving = ref(false)
const imageUploadRef = ref(null)
const selectedImage = ref(null)

const emptyForm = () => ({
  name: '',
  position: '',
  email: '',
  phone: '',
  description: '',
  existingImageUrl: null,
})

const form = ref(emptyForm())
const formErrors = ref({})

function onSearch() {
  fetchMembers(1)
}

async function fetchMembers(page = 1) {
  loading.value = true
  try {
    const params = { page, limit: 10 }
    if (search.value) params.search = search.value
    const res = await structureService.list(params)
    members.value = res.data.struktur
    pagination.value = res.data.pagination
  } catch (err) {
    notificationStore.error(err.response?.data?.error || 'Gagal memuat data struktur')
  } finally {
    loading.value = false
  }
}

function loadPage(page) {
  fetchMembers(page)
}

function openCreate() {
  editingId.value = null
  form.value = emptyForm()
  formErrors.value = {}
  selectedImage.value = null
  showModal.value = true
}

async function openEdit(member) {
  editingId.value = member.id
  try {
    // Fetch full member data to ensure we have all fields including description and image
    const res = await structureService.get(member.id)
    const fullMember = res.data
    form.value = {
      name: fullMember.name || '',
      position: fullMember.position || '',
      email: fullMember.email || '',
      phone: fullMember.phone || '',
      description: fullMember.description || '',
      existingImageUrl: fullMember.profile_image_url || null,
    }
  } catch (err) {
    notificationStore.error('Gagal memuat data anggota')
    return
  }
  formErrors.value = {}
  selectedImage.value = null
  showModal.value = true
}

function closeModal() {
  showModal.value = false
  imageUploadRef.value?.reset()
}

function onImageChange(file) {
  selectedImage.value = file
}

function validate() {
  formErrors.value = {}
  if (!form.value.name.trim()) {
    formErrors.value.name = 'Nama wajib diisi'
  }
  return Object.keys(formErrors.value).length === 0
}

async function handleSubmit() {
  if (!validate()) return

  saving.value = true
  try {
    const fd = new FormData()
    fd.append('name', form.value.name.trim())
    if (form.value.position) fd.append('position', form.value.position)
    if (form.value.email) fd.append('email', form.value.email)
    if (form.value.phone) fd.append('phone', form.value.phone)
    if (form.value.description) fd.append('description', form.value.description)
    if (selectedImage.value) fd.append('profile_image', selectedImage.value)

    if (editingId.value) {
      await structureService.update(editingId.value, fd)
      notificationStore.success('Data anggota berhasil diperbarui')
    } else {
      await structureService.create(fd)
      notificationStore.success('Anggota berhasil ditambahkan')
    }

    closeModal()
    fetchMembers(pagination.value.page)
  } catch (err) {
    notificationStore.error(err.response?.data?.error || 'Gagal menyimpan data')
  } finally {
    saving.value = false
  }
}

async function handleDelete(member) {
  const confirmed = await confirm(
    'Hapus Anggota',
    `Hapus "${member.name}"? Tindakan ini tidak dapat dibatalkan.`,
    member.name,
  )
  if (!confirmed) return
  try {
    await structureService.delete(member.id)
    notificationStore.success('Anggota berhasil dihapus')
    fetchMembers(pagination.value.page)
  } catch (err) {
    notificationStore.error(err.response?.data?.error || 'Gagal menghapus anggota')
  }
}

onMounted(() => fetchMembers())
</script>
