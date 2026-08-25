<template>
  <div
    class="flex items-center justify-center bg-secondary-100 dark:bg-secondary-800 text-secondary-400 dark:text-secondary-500 overflow-hidden"
    :class="[sizeClass, roundedClass]"
  >
    <svg v-if="variant === 'image'" class="w-1/2 h-1/2 max-w-8 max-h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
    </svg>
    <svg v-else-if="variant === 'document'" class="w-1/2 h-1/2 max-w-8 max-h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
    </svg>
    <svg v-else-if="variant === 'user'" class="w-1/2 h-1/2 max-w-8 max-h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
    </svg>
    <span v-else-if="text" class="text-xs font-medium">{{ text }}</span>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  variant: {
    type: String,
    default: 'image',
    validator: (v) => ['image', 'document', 'user', 'initials'].includes(v),
  },
  size: {
    type: String,
    default: 'md',
    validator: (v) => ['xs', 'sm', 'md', 'lg', 'xl', 'full'].includes(v),
  },
  rounded: {
    type: String,
    default: 'md',
    validator: (v) => ['none', 'sm', 'md', 'lg', 'full'].includes(v),
  },
  text: { type: String, default: '' },
})

const sizeClass = computed(() => ({
  xs: 'w-8 h-8',
  sm: 'w-10 h-10',
  md: 'w-12 h-12',
  lg: 'w-16 h-16',
  xl: 'w-24 h-24',
  full: 'w-full h-full',
}[props.size] || 'w-12 h-12'))

const roundedClass = computed(() => ({
  none: '',
  sm: 'rounded-sm',
  md: 'rounded',
  lg: 'rounded-lg',
  full: 'rounded-full',
}[props.rounded] || 'rounded'))
</script>