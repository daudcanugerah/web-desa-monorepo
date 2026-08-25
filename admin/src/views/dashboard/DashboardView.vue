<template>
  <div>
    <h1 class="text-2xl font-bold text-secondary-800 dark:text-secondary-200 mb-6">Dashboard</h1>

    <!-- Loading state -->
    <div v-if="loading" class="flex justify-center items-center py-20">
      <LoadingSpinner size="lg" />
    </div>

    <!-- Error state -->
    <div v-else-if="error" class="flex flex-col items-center justify-center py-12 px-4 bg-white dark:bg-secondary-800 rounded-lg border border-secondary-200 dark:border-secondary-700">
      <span class="text-4xl mb-3" aria-hidden="true">⚠️</span>
      <p class="text-secondary-700 dark:text-secondary-300 mb-4">{{ error }}</p>
      <AppButton variant="primary" @click="loadStats">Coba lagi</AppButton>
    </div>

    <!-- Stats grid -->
    <div v-else class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
      <router-link
        v-for="card in statCards"
        :key="card.label"
        :to="card.link"
        class="bg-white dark:bg-secondary-800 rounded-lg shadow p-6 flex items-center gap-4 border border-secondary-200 dark:border-secondary-700 hover:shadow-md hover:border-primary-300 dark:hover:border-primary-600 transition-all"
      >
        <span class="text-4xl shrink-0" aria-hidden="true">{{ card.icon }}</span>
        <div class="min-w-0">
          <p class="text-sm text-secondary-500 dark:text-secondary-400">{{ card.label }}</p>
          <p class="text-3xl font-bold text-secondary-800 dark:text-secondary-200">{{ card.count }}</p>
        </div>
        <svg class="ml-auto w-4 h-4 text-secondary-400" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
        </svg>
      </router-link>
    </div>

    <!-- Last updated -->
    <p v-if="!loading && !error" class="mt-4 text-xs text-secondary-400 dark:text-secondary-500 text-center">
      Diperbarui {{ formatRelative(updatedAt) }}
    </p>

    <!-- API status -->
    <div class="mt-6 flex justify-center">
      <button
        type="button"
        class="inline-flex items-center gap-2 px-3 py-1.5 rounded-full text-xs border transition-colors"
        :class="apiStatus === 'ok'
          ? 'bg-green-50 dark:bg-green-900/20 text-green-700 dark:text-green-300 border-green-200 dark:border-green-800'
          : apiStatus === 'down'
            ? 'bg-red-50 dark:bg-red-900/20 text-red-700 dark:text-red-300 border-red-200 dark:border-red-800'
            : 'bg-secondary-50 dark:bg-secondary-800 text-secondary-600 dark:text-secondary-300 border-secondary-200 dark:border-secondary-700'"
        :disabled="apiStatus === 'checking'"
        @click="checkHealth"
      >
        <span
          class="w-2 h-2 rounded-full"
          :class="apiStatus === 'ok' ? 'bg-green-500'
            : apiStatus === 'down' ? 'bg-red-500'
              : 'bg-secondary-400'"
        />
        <span v-if="apiStatus === 'checking'">Memeriksa API...</span>
        <span v-else-if="apiStatus === 'ok'">API: {{ apiMessage }}</span>
        <span v-else-if="apiStatus === 'down'">API tidak merespons</span>
        <span v-else>Periksa API</span>
      </button>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import api from '../../services/api'
import { healthService } from '../../services/health.service'
import AppButton from '../../components/common/AppButton.vue'
import LoadingSpinner from '../../components/common/LoadingSpinner.vue'
import { useNotificationStore } from '../../stores/notification'
import { formatRelative } from '../../utils/dateFormat'

const notification = useNotificationStore()
const loading = ref(true)
const error = ref('')
const updatedAt = ref(null)
const apiStatus = ref('idle') // idle | checking | ok | down
const apiMessage = ref('')

const statCards = ref([
  { icon: '👥', label: 'Users', count: 0, link: '/users' },
  { icon: '📰', label: 'Berita', count: 0, link: '/berita' },
  { icon: '🏪', label: 'UMKM', count: 0, link: '/umkm' },
  { icon: '🏛️', label: 'Fasilitas', count: 0, link: '/fasilitas' },
])

async function loadStats() {
  loading.value = true
  error.value = ''
  try {
    const [users, berita, umkm, fasilitas] = await Promise.all([
      api.get('/users?limit=1'),
      api.get('/berita?limit=1'),
      api.get('/umkm?limit=1'),
      api.get('/fasilitas?limit=1'),
    ])

    statCards.value[0].count = users.data.pagination.total
    statCards.value[1].count = berita.data.pagination.total
    statCards.value[2].count = umkm.data.pagination.total
    statCards.value[3].count = fasilitas.data.pagination.total
    updatedAt.value = new Date()
  } catch (err) {
    error.value = 'Gagal memuat statistik dashboard.'
    notification.error(err.response?.data?.error || error.value)
  } finally {
    loading.value = false
  }
}

async function checkHealth() {
  apiStatus.value = 'checking'
  try {
    const res = await healthService.ping()
    apiMessage.value = res.data?.status || 'ok'
    apiStatus.value = 'ok'
  } catch {
    apiStatus.value = 'down'
  }
}

onMounted(() => {
  loadStats()
  checkHealth()
})
</script>