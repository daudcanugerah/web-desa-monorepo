<template>
  <div>
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-2xl font-bold text-secondary-800 dark:text-secondary-200">Profil Desa</h1>
      <AppButton v-if="!editing" variant="primary" @click="startEdit">Edit</AppButton>
    </div>

    <LoadingSpinner v-if="loading" class="py-12" />

    <div v-else-if="!editing" class="bg-white dark:bg-secondary-800 rounded-lg shadow-sm border border-secondary-200 dark:border-secondary-700 p-6 space-y-5">
      <div class="grid grid-cols-1 sm:grid-cols-2 gap-5">
        <div>
          <p class="text-xs font-medium text-secondary-500 dark:text-secondary-400 uppercase tracking-wide mb-1">Nama Desa</p>
          <p class="text-secondary-800 dark:text-secondary-200">{{ profile.name || '—' }}</p>
        </div>
        <div>
          <p class="text-xs font-medium text-secondary-500 dark:text-secondary-400 uppercase tracking-wide mb-1">Alamat</p>
          <p class="text-secondary-800 dark:text-secondary-200">{{ profile.address || '—' }}</p>
        </div>
        <div>
          <p class="text-xs font-medium text-secondary-500 dark:text-secondary-400 uppercase tracking-wide mb-1">Telepon</p>
          <p class="text-secondary-800 dark:text-secondary-200">{{ profile.phone || '—' }}</p>
        </div>
        <div>
          <p class="text-xs font-medium text-secondary-500 dark:text-secondary-400 uppercase tracking-wide mb-1">Email</p>
          <p class="text-secondary-800 dark:text-secondary-200">{{ profile.email || '—' }}</p>
        </div>
        <div>
          <p class="text-xs font-medium text-secondary-500 dark:text-secondary-400 uppercase tracking-wide mb-1">Website</p>
          <p class="text-secondary-800 dark:text-secondary-200">{{ profile.website || '—' }}</p>
        </div>
      </div>
      <div>
        <p class="text-xs font-medium text-secondary-500 dark:text-secondary-400 uppercase tracking-wide mb-1">Visi &amp; Misi</p>
        <p class="text-secondary-800 dark:text-secondary-200 whitespace-pre-line">{{ profile.vision_mission || '—' }}</p>
      </div>

      <div class="grid grid-cols-1 sm:grid-cols-2 gap-5">
        <div>
          <p class="text-xs font-medium text-secondary-500 dark:text-secondary-400 uppercase tracking-wide mb-1">Kepala Desa</p>
          <p class="text-secondary-800 dark:text-secondary-200">{{ profile.kepala_desa || '—' }}</p>
        </div>
        <div>
          <p class="text-xs font-medium text-secondary-500 dark:text-secondary-400 uppercase tracking-wide mb-1">Motto</p>
          <p class="text-secondary-800 dark:text-secondary-200">{{ profile.motto || '—' }}</p>
        </div>
      </div>
      <div>
        <p class="text-xs font-medium text-secondary-500 dark:text-secondary-400 uppercase tracking-wide mb-2">Foto Kepala Desa</p>
        <SafeImg
          v-if="profile.kepala_desa_media && (profile.kepala_desa_media.url || profile.kepala_desa_media.thumbnail_url)"
          :src="profile.kepala_desa_media.thumbnail_url || profile.kepala_desa_media.url"
          alt="Foto Kepala Desa"
          class="w-40 rounded-lg border border-secondary-200 dark:border-secondary-700"
        />
        <p v-else class="text-sm text-secondary-400">Belum ada foto.</p>
      </div>
      <div>
        <p class="text-xs font-medium text-secondary-500 dark:text-secondary-400 uppercase tracking-wide mb-1">Sambutan Kepala Desa</p>
        <p class="text-secondary-800 dark:text-secondary-200 whitespace-pre-line">{{ profile.kepala_desa_message || '—' }}</p>
      </div>

      <div class="grid grid-cols-1 sm:grid-cols-3 gap-5">
        <div>
          <p class="text-xs font-medium text-secondary-500 dark:text-secondary-400 uppercase tracking-wide mb-1">Kecamatan</p>
          <p class="text-secondary-800 dark:text-secondary-200">{{ profile.kecamatan || '—' }}</p>
        </div>
        <div>
          <p class="text-xs font-medium text-secondary-500 dark:text-secondary-400 uppercase tracking-wide mb-1">Kabupaten</p>
          <p class="text-secondary-800 dark:text-secondary-200">{{ profile.kabupaten || '—' }}</p>
        </div>
        <div>
          <p class="text-xs font-medium text-secondary-500 dark:text-secondary-400 uppercase tracking-wide mb-1">Provinsi</p>
          <p class="text-secondary-800 dark:text-secondary-200">{{ profile.provinsi || '—' }}</p>
        </div>
      </div>

      <div>
        <p class="text-xs font-medium text-secondary-500 dark:text-secondary-400 uppercase tracking-wide mb-2">Statistik Desa</p>
        <div class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-4">
          <div v-for="s in statistikPreview" :key="s.label" class="bg-secondary-50 dark:bg-secondary-900/40 rounded-lg p-3 text-center">
            <p class="text-lg font-bold text-secondary-800 dark:text-secondary-200 tabular-nums">{{ s.value ?? '—' }}</p>
            <p class="text-[11px] text-secondary-500 dark:text-secondary-400">{{ s.label }}</p>
          </div>
        </div>
      </div>
    </div>

    <form v-else class="bg-white dark:bg-secondary-800 rounded-lg shadow-sm border border-secondary-200 dark:border-secondary-700 p-6 space-y-5" @submit.prevent="handleSave">
      <div class="grid grid-cols-1 sm:grid-cols-2 gap-5">
        <AppInput
          v-model="form.name"
          label="Nama Desa"
          required
          :error="formErrors.name"
          placeholder="Nama desa"
        />
        <AppInput
          v-model="form.address"
          label="Alamat"
          placeholder="Alamat desa"
        />
        <AppInput
          v-model="form.phone"
          label="Telepon"
          placeholder="08xxxxxxxxxx"
        />
        <AppInput
          v-model="form.email"
          label="Email"
          type="email"
          placeholder="email@desa.go.id"
        />
        <AppInput
          v-model="form.website"
          label="Website"
          placeholder="https://desa.go.id"
        />
      </div>
      <div>
        <label class="block text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-1">Visi &amp; Misi</label>
        <textarea
          v-model="form.vision_mission"
          rows="5"
          placeholder="Tuliskan visi dan misi desa..."
          class="w-full border border-secondary-300 dark:border-secondary-600 rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 bg-white dark:bg-secondary-800 text-secondary-700 dark:text-secondary-200"
        />
      </div>

      <div class="grid grid-cols-1 sm:grid-cols-2 gap-5">
        <AppInput
          v-model="form.kepala_desa"
          label="Kepala Desa"
          placeholder="Nama kepala desa"
        />
        <AppInput
          v-model="form.motto"
          label="Motto"
          placeholder="Motto desa"
        />
      </div>
      <div>
        <label class="block text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-1">Sambutan Kepala Desa</label>
        <textarea
          v-model="form.kepala_desa_message"
          rows="4"
          placeholder="Sambutan / pesan kepala desa untuk beranda..."
          class="w-full border border-secondary-300 dark:border-secondary-600 rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 bg-white dark:bg-secondary-800 text-secondary-700 dark:text-secondary-200"
        />
      </div>

      <div>
        <label class="block text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-1">Foto Kepala Desa</label>
        <ImageUpload
          ref="imageUploadRef"
          :existing-url="form.existingImageUrl"
          @change="onImageChange"
        />
      </div>

      <div class="grid grid-cols-1 sm:grid-cols-3 gap-5">
        <AppInput v-model="form.kecamatan" label="Kecamatan" placeholder="Kecamatan" />
        <AppInput v-model="form.kabupaten" label="Kabupaten" placeholder="Kabupaten" />
        <AppInput v-model="form.provinsi" label="Provinsi" placeholder="Provinsi" />
      </div>

      <div class="grid grid-cols-2 sm:grid-cols-3 gap-5">
        <AppInput v-model="form.jumlah_penduduk" label="Jumlah Penduduk" type="number" />
        <AppInput v-model="form.jumlah_kk" label="Kartu Keluarga" type="number" />
        <AppInput v-model="form.jumlah_dusun" label="Jumlah Dusun" type="number" />
        <AppInput v-model="form.jumlah_rt" label="Jumlah RT" type="number" />
        <AppInput v-model="form.jumlah_rw" label="Jumlah RW" type="number" />
        <AppInput v-model="form.jumlah_umkm" label="Jumlah UMKM" type="number" />
      </div>

      <div>
        <label class="block text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-2">Media Sosial</label>
        <div class="grid grid-cols-1 sm:grid-cols-2 gap-5">
          <AppInput v-model="form.social_facebook" label="Facebook" placeholder="https://facebook.com/..." />
          <AppInput v-model="form.social_instagram" label="Instagram" placeholder="https://instagram.com/..." />
          <AppInput v-model="form.social_youtube" label="YouTube" placeholder="https://youtube.com/@..." />
          <AppInput v-model="form.social_tiktok" label="TikTok" placeholder="https://tiktok.com/@..." />
          <AppInput v-model="form.social_whatsapp" label="WhatsApp" placeholder="https://wa.me/62..." />
          <AppInput v-model="form.social_twitter" label="X / Twitter" placeholder="https://x.com/..." />
        </div>
      </div>

      <div class="flex justify-end gap-3 pt-2">
        <AppButton variant="secondary" type="button" @click="cancelEdit">Batal</AppButton>
        <AppButton variant="primary" type="submit" :loading="saving">Simpan</AppButton>
      </div>
    </form>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import AppButton from '../../components/common/AppButton.vue'
import AppInput from '../../components/common/AppInput.vue'
import LoadingSpinner from '../../components/common/LoadingSpinner.vue'
import SafeImg from '../../components/common/SafeImg.vue'
import ImageUpload from '../../components/forms/ImageUpload.vue'
import { profileService } from '../../services/profile.service'
import { useNotificationStore } from '../../stores/notification'

const notificationStore = useNotificationStore()

const loading = ref(false)
const saving = ref(false)
const editing = ref(false)
const profile = ref({})
const form = ref({})
const formErrors = ref({})
const imageUploadRef = ref(null)
const selectedImage = ref(null)

function onImageChange(file) {
  selectedImage.value = file
}

// socialUrl returns the stored URL for one platform slug (or '').
function socialUrl(profile, platform) {
  const link = (profile.social_media || []).find((x) => x && x.platform === platform)
  return link?.url || ''
}

// collectSocialLinks turns the six fixed inputs into a [{platform,url}] list,
// dropping blanks so the backend sanitizer has nothing to strip.
function collectSocialLinks(form) {
  return [
    ['facebook', form.social_facebook],
    ['instagram', form.social_instagram],
    ['youtube', form.social_youtube],
    ['tiktok', form.social_tiktok],
    ['whatsapp', form.social_whatsapp],
    ['twitter', form.social_twitter],
  ]
    .filter(([, url]) => url && String(url).trim())
    .map(([platform, url]) => ({ platform, url: String(url).trim() }))
}

const statistikPreview = computed(() => [
  { label: 'Penduduk', value: profile.value.jumlah_penduduk },
  { label: 'Kartu Keluarga', value: profile.value.jumlah_kk },
  { label: 'Dusun', value: profile.value.jumlah_dusun },
  { label: 'RT', value: profile.value.jumlah_rt },
  { label: 'RW', value: profile.value.jumlah_rw },
  { label: 'UMKM', value: profile.value.jumlah_umkm },
])

async function fetchProfile() {
  loading.value = true
  try {
    const res = await profileService.get()
    profile.value = res.data
  } catch (err) {
    notificationStore.error(err.response?.data?.error || 'Gagal memuat profil desa')
  } finally {
    loading.value = false
  }
}

function startEdit() {
  const p = profile.value
  form.value = {
    name: p.name || '',
    address: p.address || '',
    phone: p.phone || '',
    email: p.email || '',
    website: p.website || '',
    vision_mission: p.vision_mission || '',
    kepala_desa: p.kepala_desa || '',
    kepala_desa_message: p.kepala_desa_message || '',
    kepala_desa_media_id: p.kepala_desa_media_id || null,
    existingImageUrl: p.kepala_desa_media?.thumbnail_url || p.kepala_desa_media?.url || null,
    motto: p.motto || '',
    kecamatan: p.kecamatan || '',
    kabupaten: p.kabupaten || '',
    provinsi: p.provinsi || '',
    jumlah_penduduk: p.jumlah_penduduk ?? null,
    jumlah_kk: p.jumlah_kk ?? null,
    jumlah_dusun: p.jumlah_dusun ?? null,
    jumlah_rt: p.jumlah_rt ?? null,
    jumlah_rw: p.jumlah_rw ?? null,
    jumlah_umkm: p.jumlah_umkm ?? null,
    social_facebook: socialUrl(p, 'facebook'),
    social_instagram: socialUrl(p, 'instagram'),
    social_youtube: socialUrl(p, 'youtube'),
    social_tiktok: socialUrl(p, 'tiktok'),
    social_whatsapp: socialUrl(p, 'whatsapp'),
    social_twitter: socialUrl(p, 'twitter'),
  }
  formErrors.value = {}
  editing.value = true
}

function cancelEdit() {
  editing.value = false
  formErrors.value = {}
  selectedImage.value = null
  imageUploadRef.value?.reset()
}

function validate() {
  formErrors.value = {}
  if (!form.value.name?.trim()) {
    formErrors.value.name = 'Nama desa wajib diisi'
  }
  return Object.keys(formErrors.value).length === 0
}

async function handleSave() {
  if (!validate()) return
  saving.value = true
  try {
    // Upload a newly chosen kepala desa photo first, then persist its media id.
    if (selectedImage.value) {
      const uploadRes = await profileService.uploadMedia(selectedImage.value)
      form.value.kepala_desa_media_id = uploadRes.data.media_id
    }

    const payload = { ...form.value }
    delete payload.existingImageUrl

    // Collapse the six social inputs into the API's social_media array.
    payload.social_media = collectSocialLinks(payload)
    for (const key of ['social_facebook', 'social_instagram', 'social_youtube',
                       'social_tiktok', 'social_whatsapp', 'social_twitter']) {
      delete payload[key]
    }

    // Number inputs emit strings; coerce to int, and blank -> null (omitted).
    for (const key of ['jumlah_penduduk', 'jumlah_kk', 'jumlah_dusun', 'jumlah_rt', 'jumlah_rw', 'jumlah_umkm']) {
      const raw = payload[key]
      if (raw === '' || raw === null || raw === undefined) {
        payload[key] = null
      } else {
        const n = Number(raw)
        payload[key] = Number.isNaN(n) ? null : Math.max(0, Math.trunc(n))
      }
    }
    const res = await profileService.update(payload)
    profile.value = res.data
    selectedImage.value = null
    imageUploadRef.value?.reset()
    notificationStore.success('Profil desa berhasil diperbarui')
    editing.value = false
  } catch (err) {
    notificationStore.error(err.response?.data?.error || 'Gagal menyimpan profil desa')
  } finally {
    saving.value = false
  }
}

onMounted(fetchProfile)
</script>
