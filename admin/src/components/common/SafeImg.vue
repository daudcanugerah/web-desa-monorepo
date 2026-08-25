<template>
  <div
    class="relative inline-flex items-center justify-center overflow-hidden bg-secondary-100 dark:bg-secondary-800"
    :class="containerClass"
  >
    <img
      v-if="resolvedSrc && !errored"
      :src="resolvedSrc"
      :alt="alt"
      :loading="lazy ? 'lazy' : 'eager'"
      :decoding="lazy ? 'async' : 'auto'"
      :class="['max-w-full max-h-full', objectClass]"
      @error="onError"
    />
    <div
      v-else-if="!resolvedSrc"
      class="flex flex-col items-center justify-center text-secondary-400 dark:text-secondary-500 w-full h-full"
      aria-hidden="true"
    >
      <svg class="w-1/2 h-1/2 max-w-8 max-h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path
          stroke-linecap="round"
          stroke-linejoin="round"
          stroke-width="1.5"
          d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z"
        />
      </svg>
    </div>
    <div
      v-else
      class="flex flex-col items-center justify-center text-secondary-400 dark:text-secondary-500 w-full h-full text-xs"
      aria-hidden="true"
    >
      <svg class="w-1/2 h-1/2 max-w-8 max-h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path
          stroke-linecap="round"
          stroke-linejoin="round"
          stroke-width="1.5"
          d="M12 9v2m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
        />
      </svg>
    </div>
  </div>
</template>

<script setup>
import { computed, ref } from 'vue'
import { getImageUrl } from '../../utils/imageUrl'

const props = defineProps({
  src: { type: String, default: '' },
  alt: { type: String, default: '' },
  lazy: { type: Boolean, default: true },
  rounded: { type: Boolean, default: false },
  circle: { type: Boolean, default: false },
  objectCover: { type: Boolean, default: true },
  fixedHeight: { type: String, default: null },
})

const errored = ref(false)

const resolvedSrc = computed(() => {
  const src = props.src
  if (!src) return null
  if (/^(https?:|blob:|data:)/i.test(src)) return src
  return getImageUrl(src)
})

const containerClass = computed(() => ({
  rounded: props.rounded && !props.circle,
  'rounded-full': props.circle,
  'object-cover': props.objectCover,
}))

const objectClass = computed(() => (props.objectCover ? 'object-cover' : 'object-contain'))

function onError() {
  errored.value = true
}
</script>