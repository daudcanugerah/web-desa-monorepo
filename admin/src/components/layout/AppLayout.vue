<template>
  <a
    href="#main-content"
    class="sr-only focus:not-sr-only focus:absolute focus:top-2 focus:left-2 focus:z-50 focus:bg-primary-600 focus:text-white focus:px-4 focus:py-2 focus:rounded-lg focus:shadow-lg focus:outline-none focus:ring-2 focus:ring-primary-300"
  >
    Lewati ke konten utama
  </a>
  <div class="flex h-screen overflow-hidden">
    <!-- Sidebar -->
    <AppSidebar />

    <!-- Main area — shrinks/grows based on sidebar width -->
    <div class="flex-1 flex flex-col overflow-auto min-w-0">
      <AppHeader />
      <main id="main-content" tabindex="-1" class="flex-1 p-6 bg-secondary-50 dark:bg-secondary-900 focus:outline-none">
        <AppBreadcrumb class="mb-4" />
        <router-view />
      </main>
    </div>
  </div>
</template>

<script setup>
import { onMounted } from 'vue'
import { useAuthStore } from '../../stores/auth'
import AppSidebar from './AppSidebar.vue'
import AppHeader from './AppHeader.vue'
import AppBreadcrumb from './AppBreadcrumb.vue'

const authStore = useAuthStore()

onMounted(() => {
  authStore.fetchCurrentUser()
})
</script>
