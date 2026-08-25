<template>
  <button
    type="button"
    class="p-2 rounded-lg text-secondary-600 hover:bg-secondary-100 dark:text-secondary-300 dark:hover:bg-secondary-700 transition-colors focus:outline-none focus:ring-2 focus:ring-primary-500"
    :title="title"
    :aria-label="title"
    @click="themeStore.cycle"
  >
    <!-- Light mode → show sun (click to go dark) -->
    <svg v-if="themeStore.isDark" class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z" />
    </svg>
    <!-- Dark mode → show moon (click to go auto) -->
    <svg v-else-if="themeStore.mode === 'auto'" class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z" />
    </svg>
    <!-- Auto mode → show half-tone circle (click to go light) -->
    <svg v-else class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
      <circle cx="12" cy="12" r="9" stroke-width="2" />
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 3v18M12 3a9 9 0 010 18" fill="currentColor" />
    </svg>
    <span
      v-if="themeStore.mode === 'auto'"
      class="absolute -top-0.5 -right-0.5 w-2 h-2 rounded-full bg-primary-500 ring-2 ring-white dark:ring-secondary-800"
      aria-hidden="true"
    />
  </button>
</template>

<script setup>
import { computed } from 'vue'
import { useThemeStore } from '../../stores/theme'

const themeStore = useThemeStore()

const title = computed(() => {
  if (themeStore.mode === 'light') return 'Mode terang (klik untuk gelap)'
  if (themeStore.mode === 'dark') return 'Mode gelap (klik untuk otomatis)'
  return 'Otomatis (klik untuk terang)'
})
</script>