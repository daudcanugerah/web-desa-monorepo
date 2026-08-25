<template>
  <div class="flex flex-col items-center justify-center py-20">
    <div class="text-6xl mb-4">🔍</div>
    <h1 class="text-2xl font-bold text-secondary-800 dark:text-secondary-200 mb-2">Halaman Tidak Ditemukan</h1>
    <p class="text-secondary-500 dark:text-secondary-400 mb-6">Maaf, halaman yang Anda cari tidak tersedia.</p>
    <AppButton variant="primary" @click="router.push('/')">Kembali ke Dashboard</AppButton>
  </div>
</template>

<script setup>
import { useRouter } from 'vue-router'
import AppButton from '../components/common/AppButton.vue'

const router = useRouter()
</script>