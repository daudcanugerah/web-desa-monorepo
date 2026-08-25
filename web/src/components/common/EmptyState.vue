<template>
  <div class="flex flex-col items-center justify-center py-16 px-4 text-center">
    <div class="inline-flex items-center justify-center w-16 h-16 rounded-full bg-gray-100 mb-4">
      <svg class="w-8 h-8 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4" />
      </svg>
    </div>
    <h3 class="text-base font-semibold text-gray-900 mb-1">{{ message }}</h3>
    <p v-if="description" class="text-sm text-gray-500 mb-4 max-w-sm">{{ description }}</p>
    <button
      v-if="actionLabel"
      @click="$emit('action')"
      class="inline-flex items-center px-4 py-2 bg-emerald-600 text-white text-sm font-medium rounded-lg hover:bg-emerald-700 transition-colors"
    >
      {{ actionLabel }}
    </button>
  </div>
</template>

<script>
export default {
  name: 'EmptyState',
  props: {
    message: { type: String, default: 'Data tidak tersedia' },
    description: { type: String, default: '' },
    actionLabel: { type: String, default: '' }
  },
  emits: ['action']
}
</script>