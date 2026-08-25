<template>
  <div>
    <div class="flex flex-wrap gap-2 mb-2">
      <div
        v-for="(img, idx) in images"
        :key="`existing-${idx}`"
        class="relative w-20 h-20 border border-secondary-300 dark:border-secondary-600 rounded overflow-hidden group"
      >
        <SafeImg
          :src="typeof img === 'string' ? img : URL.createObjectURL(img)"
          :alt="`Image ${idx + 1}`"
          class="w-full h-full"
          :lazy="false"
        />
        <button
          type="button"
          class="absolute top-0 right-0 bg-danger-600 text-white rounded-bl px-1.5 py-0.5 text-xs opacity-90 hover:opacity-100"
          :aria-label="`Hapus gambar ${idx + 1}`"
          @click="removeAt(idx)"
        >
          ×
        </button>
      </div>
      <label
        class="w-20 h-20 border-2 border-dashed border-secondary-300 dark:border-secondary-600 rounded flex items-center justify-center cursor-pointer hover:border-primary-500 hover:bg-primary-50 dark:hover:bg-primary-900/10 transition-colors"
        :class="{ 'opacity-50 cursor-not-allowed': disabled }"
      >
        <svg class="w-6 h-6 text-secondary-400 dark:text-secondary-500" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
        </svg>
        <input
          ref="inputRef"
          type="file"
          multiple
          accept="image/*"
          class="hidden"
          :disabled="disabled"
          @change="onPick"
        />
      </label>
    </div>
    <p v-if="error" class="mt-1 text-xs text-red-600">{{ error }}</p>
    <p v-else class="text-xs text-secondary-500 dark:text-secondary-400">JPEG, PNG, atau WebP. Maksimal {{ maxMB }}MB per gambar.</p>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'
import SafeImg from './SafeImg.vue'
import { isValidImageType, isWithinSize } from '../../utils/validators'

const props = defineProps({
  modelValue: {
    type: Array,
    default: () => [],
  },
  maxMB: { type: Number, default: 10 },
  disabled: { type: Boolean, default: false },
})

const emit = defineEmits(['update:modelValue'])

const images = ref([...props.modelValue])
const inputRef = ref(null)
const error = ref('')

watch(() => props.modelValue, (val) => {
  images.value = [...val]
}, { deep: true })

function onPick(e) {
  error.value = ''
  const files = Array.from(e.target.files || [])
  for (const file of files) {
    if (!isValidImageType(file)) {
      error.value = 'Hanya file JPEG, PNG, atau WebP yang diperbolehkan.'
      continue
    }
    if (!isWithinSize(file, props.maxMB)) {
      error.value = `Ukuran file melebihi ${props.maxMB}MB.`
      continue
    }
    images.value.push(file)
  }
  if (inputRef.value) inputRef.value.value = ''
  emit('update:modelValue', images.value)
}

function removeAt(idx) {
  images.value.splice(idx, 1)
  emit('update:modelValue', images.value)
}
</script>