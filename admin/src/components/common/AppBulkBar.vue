<template>
  <div
    v-if="selected.size > 0"
    class="fixed bottom-6 left-1/2 -translate-x-1/2 z-40 bg-secondary-900 dark:bg-secondary-700 text-white rounded-xl shadow-2xl border border-secondary-700 dark:border-secondary-600 px-4 py-3 flex items-center gap-3"
    role="region"
    aria-live="polite"
  >
    <span class="text-sm font-medium">
      {{ selected.size }} dipilih
    </span>
    <div class="h-5 w-px bg-secondary-600" />
    <slot :selected="Array.from(selected)" :clear="clear" />
    <button
      type="button"
      class="ml-2 text-secondary-400 hover:text-white transition-colors"
      aria-label="Bersihkan pilihan"
      @click="clear"
    >
      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
      </svg>
    </button>
  </div>
</template>

<script setup>
const props = defineProps({
  selected: { type: Set, required: true },
})
const emit = defineEmits(['update:selected'])

function clear() {
  emit('update:selected', new Set())
}
</script>