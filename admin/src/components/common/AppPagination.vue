<template>
  <div class="flex items-center gap-1 mt-4">
    <AppButton
      variant="secondary"
      size="sm"
      :disabled="page === 1"
      @click="emit('change', page - 1)"
    >
      Previous
    </AppButton>

    <AppButton
      v-for="p in visiblePages"
      :key="p"
      variant="secondary"
      size="sm"
      :class="p === page ? 'bg-primary-600 text-white hover:bg-primary-700' : ''"
      @click="emit('change', p)"
    >
      {{ p }}
    </AppButton>

    <AppButton
      variant="secondary"
      size="sm"
      :disabled="page === totalPages"
      @click="emit('change', page + 1)"
    >
      Next
    </AppButton>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import AppButton from './AppButton.vue'

const props = defineProps({
  page: {
    type: Number,
    required: true,
  },
  totalPages: {
    type: Number,
    required: true,
  },
})

const emit = defineEmits(['change'])

const visiblePages = computed(() => {
  const total = props.totalPages
  const current = props.page
  const maxVisible = 5

  if (total <= maxVisible) {
    return Array.from({ length: total }, (_, i) => i + 1)
  }

  let start = Math.max(1, current - Math.floor(maxVisible / 2))
  let end = start + maxVisible - 1

  if (end > total) {
    end = total
    start = Math.max(1, end - maxVisible + 1)
  }

  return Array.from({ length: end - start + 1 }, (_, i) => start + i)
})
</script>