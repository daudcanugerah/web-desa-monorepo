<template>
  <button
    class="text-secondary-500 hover:text-secondary-700 p-1.5 rounded-lg hover:bg-secondary-100 transition-colors"
    @click="$router.back()"
  >
    <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
    </svg>
  </button>
</template>

<script setup>
defineOptions({
  inheritAttrs: false,
})
</script>