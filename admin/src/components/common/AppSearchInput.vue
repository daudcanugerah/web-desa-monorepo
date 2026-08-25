<template>
  <div class="relative">
    <span class="absolute inset-y-0 left-0 flex items-center pl-3 text-secondary-400 dark:text-secondary-500 pointer-events-none" aria-hidden="true">
      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-4.35-4.35M11 19a8 8 0 100-16 8 8 0 000 16z" />
      </svg>
    </span>
    <input
      :id="id"
      :value="modelValue"
      type="search"
      :placeholder="placeholder"
      :aria-label="ariaLabel || placeholder"
      class="w-full pl-9 pr-9 py-2 border rounded text-sm focus:outline-none focus:ring-2 focus:ring-primary-500 bg-white dark:bg-secondary-800 text-secondary-700 dark:text-secondary-200"
      :class="[
        widthClass,
        borderClass,
      ]"
      @input="onInput"
      @clear="clear"
    />
    <button
      v-if="modelValue"
      type="button"
      class="absolute inset-y-0 right-0 flex items-center pr-3 text-secondary-400 hover:text-secondary-600 dark:text-secondary-500 dark:hover:text-secondary-300"
      aria-label="Bersihkan pencarian"
      @click="clear"
    >
      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
      </svg>
    </button>
  </div>
</template>

<script setup>
import { computed, ref, onBeforeUnmount } from 'vue'

const props = defineProps({
  modelValue: { type: String, default: '' },
  placeholder: { type: String, default: 'Cari...' },
  ariaLabel: { type: String, default: '' },
  id: { type: String, default: '' },
  width: { type: String, default: 'sm' },
  debounce: { type: Number, default: 300 },
  immediate: { type: Boolean, default: false },
})

const emit = defineEmits(['update:modelValue', 'search'])

const widthClass = computed(() => ({
  xs: 'max-w-xs',
  sm: 'max-w-sm',
  md: 'max-w-md',
  lg: 'max-w-lg',
  full: 'w-full',
}[props.width] || 'max-w-sm'))

const borderClass = 'border-secondary-300 dark:border-secondary-600 hover:border-secondary-400 dark:hover:border-secondary-500 focus:border-primary-500'

let timer = null

function onInput(e) {
  const value = e.target.value
  emit('update:modelValue', value)

  if (timer) clearTimeout(timer)

  if (props.immediate) {
    emit('search', value)
    return
  }
  timer = setTimeout(() => {
    emit('search', value)
  }, props.debounce)
}

function clear() {
  if (timer) clearTimeout(timer)
  emit('update:modelValue', '')
  emit('search', '')
}

onBeforeUnmount(() => {
  if (timer) clearTimeout(timer)
})
</script>