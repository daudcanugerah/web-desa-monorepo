<template>
  <div>
    <!-- Drop zone -->
    <div
      class="border-2 border-dashed rounded-lg p-6 text-center cursor-pointer transition-colors"
      :class="isDragging ? 'border-blue-400 bg-blue-50 dark:bg-blue-900/20' : 'border-secondary-300 dark:border-secondary-600 hover:border-secondary-400 dark:hover:border-secondary-500'"
      @click="fileInput.click()"
      @dragover.prevent="isDragging = true"
      @dragleave.prevent="isDragging = false"
      @drop.prevent="onDrop"
    >
<!-- Preview -->
    <SafeImg
      v-if="displayImageUrl"
      :src="displayImageUrl"
      alt="Preview"
      class="mx-auto max-h-40 rounded mb-2 object-contain"
    />
      <div v-else class="text-secondary-400 dark:text-secondary-500">
        <svg class="mx-auto w-10 h-10 mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5"
            d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
        </svg>
        <p class="text-sm">Click or drag an image here</p>
        <p class="text-xs mt-1">JPEG, PNG, WebP — max {{ maxMB }}MB</p>
      </div>

      <p v-if="fileSize" class="text-xs text-secondary-500 dark:text-secondary-400 mt-2">{{ fileSize }}</p>
    </div>

    <input ref="fileInput" type="file" accept="image/jpeg,image/png,image/webp" class="hidden" @change="onFileChange" />

    <p v-if="warning" class="mt-1 text-xs text-yellow-600">⚠ {{ warning }}</p>
    <p v-if="error" class="mt-1 text-xs text-red-600">{{ error }}</p>
    <p v-if="getImageUrl(existingUrl) && !previewUrl" class="mt-1 text-xs text-secondary-500 dark:text-secondary-400">Current image will be kept if no new file is selected.</p>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useImagePreview } from '../../composables/useImagePreview'
import { getImageUrl } from '../../utils/imageUrl'
import SafeImg from '../common/SafeImg.vue'

const props = defineProps({
  maxMB: { type: Number, default: 10 },
  warnMB: { type: Number, default: 5 },
  existingUrl: { type: String, default: null },
})

const emit = defineEmits(['change'])

const fileInput = ref(null)
const isDragging = ref(false)

const { previewUrl, fileSize, error, warning, processFile, reset } = useImagePreview({
  maxMB: props.maxMB,
  warnMB: props.warnMB,
})

const displayImageUrl = computed(() => {
  return previewUrl.value || getImageUrl(props.existingUrl)
})

function onFileChange(e) {
  const file = e.target.files[0]
  if (!file) return
  if (processFile(file)) emit('change', file)
}

function onDrop(e) {
  isDragging.value = false
  const file = e.dataTransfer.files[0]
  if (!file) return
  if (processFile(file)) emit('change', file)
}

defineExpose({ reset })
</script>
