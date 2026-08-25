<template>
  <AppModal
    :show="show"
    :title="title"
    size="lg"
    @close="cancel"
  >
    <div class="space-y-4">
      <!-- Cropper -->
      <div class="relative bg-secondary-900 rounded-lg overflow-hidden" style="height: 420px;">
        <img
          v-if="src"
          ref="imageRef"
          :src="src"
          alt="Pratinjau gambar"
          class="max-w-full max-h-full"
        />
      </div>

      <!-- Controls -->
      <div class="flex flex-wrap items-center justify-center gap-2">
        <AppButton variant="secondary" size="sm" @click="rotate(-90)">⟲ Rotasi Kiri</AppButton>
        <AppButton variant="secondary" size="sm" @click="rotate(90)">⟳ Rotasi Kanan</AppButton>
        <AppButton variant="secondary" size="sm" @click="zoomIn">🔍 Perbesar</AppButton>
        <AppButton variant="secondary" size="sm" @click="zoomOut">🔎 Perkecil</AppButton>
        <AppButton variant="ghost" size="sm" @click="reset">↺ Reset</AppButton>
      </div>

      <p v-if="hint" class="text-xs text-secondary-500 dark:text-secondary-400 text-center">
        {{ hint }}
      </p>
    </div>

    <template #footer>
      <div class="flex justify-end gap-3">
        <AppButton variant="secondary" @click="cancel">Batal</AppButton>
        <AppButton variant="primary" :loading="processing" @click="applyCrop">
          {{ applyLabel }}
        </AppButton>
      </div>
    </template>
  </AppModal>
</template>

<script setup>
import { ref, watch, nextTick, onBeforeUnmount } from 'vue'
import Cropper from 'cropperjs'
import 'cropperjs/dist/cropper.css'
import AppModal from './AppModal.vue'
import AppButton from './AppButton.vue'

const props = defineProps({
  show: { type: Boolean, default: false },
  src: { type: String, default: '' },
  originalName: { type: String, default: 'image' },
  aspectRatio: { type: Number, default: 16 / 9 },
  title: { type: String, default: 'Edit Gambar' },
  applyLabel: { type: String, default: 'Terapkan' },
  hint: { type: String, default: 'Seret untuk mengatur area, gunakan tombol untuk rotasi dan zoom.' },
  // Output resolution cap
  maxWidth: { type: Number, default: 1920 },
  maxHeight: { type: Number, default: 1080 },
  outputType: { type: String, default: 'image/webp' },
  quality: { type: Number, default: 0.92 },
})

const emit = defineEmits(['update:show', 'crop'])

const imageRef = ref(null)
let cropper = null
const processing = ref(false)

function initCropper() {
  if (!imageRef.value) return
  const img = imageRef.value
  // Wait for the image to finish loading before initialising the cropper,
  // otherwise cropperjs can mis-measure natural dimensions.
  if (img.complete) {
    doInit()
    return
  }
  img.onload = doInit
  img.onerror = () => {
    console.error('[AppImageEditor] failed to load image source')
  }
}

function doInit() {
  if (!imageRef.value) return
  cropper = new Cropper(imageRef.value, {
    aspectRatio: props.aspectRatio,
    viewMode: 1,
    autoCropArea: 1,
    responsive: true,
    background: false,
    dragMode: 'move',
    minCropBoxWidth: 320,
    minCropBoxHeight: 180,
  })
}

function destroyCropper() {
  if (cropper) {
    cropper.destroy()
    cropper = null
  }
}

function rotate(deg) {
  cropper?.rotate(deg)
}

function zoomIn() {
  cropper?.zoom(0.1)
}

function zoomOut() {
  cropper?.zoom(-0.1)
}

function reset() {
  cropper?.reset()
}

async function applyCrop() {
  if (!cropper) return
  processing.value = true
  try {
    // Compute output size preserving aspect ratio, capped at maxWidth/maxHeight
    const { width, height } = cropper.getData()
    const ratio = width > 0 && height > 0 ? width / height : props.aspectRatio
    let outW = props.maxWidth
    let outH = Math.round(outW / ratio)
    if (outH > props.maxHeight) {
      outH = props.maxHeight
      outW = Math.round(outH * ratio)
    }

    const canvas = cropper.getCroppedCanvas({
      width: outW,
      height: outH,
      imageSmoothingEnabled: true,
      imageSmoothingQuality: 'high',
    })

    const blob = await new Promise((resolve, reject) => {
      canvas.toBlob((b) => (b ? resolve(b) : reject(new Error('Konversi gambar gagal'))), props.outputType, props.quality)
    })

    // Derive extension from MIME type
    const ext = props.outputType.split('/')[1].split('+')[0] || 'png'
    const name = props.originalName.replace(/\.[^.]+$/, '') + '.' + ext
    const file = new File([blob], name, { type: props.outputType || blob.type })

    emit('crop', file)
    emit('update:show', false)
  } catch (err) {
    // Surface via console + caller handles UI. Keep modal open.
    console.error('[AppImageEditor] crop failed:', err)
  } finally {
    processing.value = false
  }
}

function cancel() {
  emit('update:show', false)
}

watch(
  () => props.show,
  (val) => {
    if (val) {
      destroyCropper()
      nextTick(initCropper)
    } else {
      destroyCropper()
    }
  },
)

onBeforeUnmount(destroyCropper)
</script>

<style>
/* The cropper wraps the <img> in a .cropper-container that must fill
   its parent (the fixed-height box). The <img> itself must not impose
   its own layout size — cropper handles positioning. */
.cropper-container,
.cropper-container img {
  width: 100%;
  height: 100%;
}

.cropper-container .cropper-crop-box,
.cropper-container .cropper-view-box {
  border-radius: 4px;
}
</style>