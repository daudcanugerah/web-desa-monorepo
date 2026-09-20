<template>
  <Teleport to="body">
    <Transition name="lightbox">
      <div
        v-if="show"
        class="fixed inset-0 z-50 flex items-center justify-center bg-black/90 backdrop-blur-sm"
        role="dialog"
        aria-modal="true"
        :aria-label="current?.original_filename || 'Pratinjau media'"
        @click.self="close"
      >
        <!-- Close -->
        <button
          type="button"
          class="absolute top-4 right-4 z-20 p-2 rounded-full text-white/80 hover:text-white hover:bg-white/10 transition-colors focus:outline-none focus:ring-2 focus:ring-white/60"
          aria-label="Tutup pratinjau"
          @click="close"
        >
          <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>

        <!-- Prev -->
        <button
          v-if="list.length > 1"
          type="button"
          class="absolute left-4 top-1/2 -translate-y-1/2 z-20 p-3 rounded-full text-white/80 hover:text-white hover:bg-white/10 transition-colors focus:outline-none focus:ring-2 focus:ring-white/60"
          aria-label="Sebelumnya"
          @click.stop="step(-1)"
        >
          <svg class="w-7 h-7" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
          </svg>
        </button>

        <!-- Next -->
        <button
          v-if="list.length > 1"
          type="button"
          class="absolute right-4 top-1/2 -translate-y-1/2 z-20 p-3 rounded-full text-white/80 hover:text-white hover:bg-white/10 transition-colors focus:outline-none focus:ring-2 focus:ring-white/60"
          aria-label="Berikutnya"
          @click.stop="step(1)"
        >
          <svg class="w-7 h-7" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
          </svg>
        </button>

        <!-- Content -->
        <div class="max-w-5xl w-full px-4 md:px-16 flex flex-col items-center" @click.self="close">
          <img
            v-if="current && current.media_type !== 'video' && srcFor(current)"
            :src="srcFor(current)"
            :alt="current.original_filename"
            class="max-h-[78vh] max-w-full object-contain rounded-lg shadow-2xl"
          />
          <video
            v-else-if="current && current.media_type === 'video'"
            :src="srcFor(current)"
            controls
            autoplay
            class="max-h-[78vh] max-w-full object-contain rounded-lg shadow-2xl"
          />
          <div
            v-else
            class="text-white/60 text-sm flex flex-col items-center gap-3 py-10"
          >
            <svg class="w-12 h-12" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M12 9v2m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            Media tidak tersedia
          </div>

          <!-- Meta bar -->
          <div class="mt-4 w-full max-w-2xl flex items-center justify-between gap-3 text-white">
            <div class="min-w-0">
              <p class="text-sm font-medium text-white truncate">
                {{ current?.original_filename || 'Tanpa nama' }}
              </p>
              <p class="text-xs text-white/60 mt-0.5">
                {{ index + 1 }}/{{ list.length }} · {{ current?.media_type }} · {{ formatBytes(current?.file_size) }}
              </p>
            </div>

            <div class="flex items-center gap-2 shrink-0">
              <button
                type="button"
                class="px-3 py-1.5 rounded-lg text-xs font-medium transition-colors"
                :class="current?.is_public
                  ? 'bg-green-500/20 text-green-300 hover:bg-green-500/30'
                  : 'bg-white/10 text-white/80 hover:bg-white/20'"
                @click="toggleVisibility"
              >
                {{ current?.is_public ? 'Publik' : 'Privat' }}
              </button>
              <button
                v-if="showDelete"
                type="button"
                class="px-3 py-1.5 rounded-lg text-xs font-medium bg-red-500/20 text-red-300 hover:bg-red-500/30 transition-colors"
                @click="remove"
              >
                Hapus
              </button>
            </div>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup>
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { resolveMediaUrl } from '../../utils/imageUrl'

const props = defineProps({
  show: { type: Boolean, default: false },
  list: { type: Array, default: () => [] },
  startIndex: { type: Number, default: 0 },
  showDelete: { type: Boolean, default: true },
})

const emit = defineEmits(['update:show', 'toggle-visibility', 'delete'])

const index = ref(0)

const current = computed(() => props.list[index.value] || null)

function srcFor(media) {
  if (!media) return ''
  return resolveMediaUrl(media.content_url || media.thumbnail_url || '')
}

function formatBytes(bytes) {
  if (!bytes) return '—'
  const units = ['B', 'KB', 'MB', 'GB']
  let i = 0
  let v = bytes
  while (v >= 1024 && i < units.length - 1) {
    v /= 1024
    i++
  }
  return `${v.toFixed(v >= 10 || i === 0 ? 0 : 1)} ${units[i]}`
}

function step(dir) {
  const n = props.list.length
  if (n < 2) return
  index.value = (index.value + dir + n) % n
}

function close() {
  emit('update:show', false)
}

function toggleVisibility() {
  if (current.value) emit('toggle-visibility', current.value)
}

function remove() {
  if (current.value) emit('delete', current.value)
}

function handleKeydown(e) {
  if (!props.show) return
  if (e.key === 'Escape') close()
  else if (e.key === 'ArrowLeft') step(-1)
  else if (e.key === 'ArrowRight') step(1)
}

watch(
  () => props.show,
  (val) => {
    if (val) index.value = props.startIndex
  },
)

watch(
  () => props.list.length,
  (n) => {
    if (props.show && n > 0) index.value = Math.min(index.value, n - 1)
  },
)

onMounted(() => document.addEventListener('keydown', handleKeydown))
onUnmounted(() => document.removeEventListener('keydown', handleKeydown))
</script>

<style scoped>
.lightbox-enter-active,
.lightbox-leave-active {
  transition: opacity 0.2s ease;
}
.lightbox-enter-from,
.lightbox-leave-to {
  opacity: 0;
}
</style>