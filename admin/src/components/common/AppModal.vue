<template>
  <Teleport to="body">
    <Transition name="modal">
      <div
        v-if="show"
        class="fixed inset-0 z-50 flex items-center justify-center p-4"
        @click.self="emit('close')"
      >
        <!-- Backdrop -->
        <div class="absolute inset-0 bg-black/50" @click="emit('close')" />

        <!-- Modal card -->
        <div
          ref="modalRef"
          role="dialog"
          aria-modal="true"
          :aria-labelledby="titleId"
          class="relative bg-white dark:bg-secondary-800 rounded-xl shadow-xl w-full z-10 flex flex-col max-h-[90vh]"
          :class="sizeClass"
        >
          <!-- Header -->
          <div class="flex items-center justify-between px-6 py-4 border-b border-secondary-200 dark:border-secondary-700">
            <h2 :id="titleId" class="text-lg font-semibold text-secondary-800 dark:text-secondary-200">
              {{ title }}
            </h2>
            <AppIconButton
              variant="ghost"
              size="md"
              aria-label="Close modal"
              @click="emit('close')"
            >
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </AppIconButton>
          </div>

          <!-- Body -->
          <div class="px-6 py-4 overflow-y-auto flex-1">
            <slot />
          </div>

          <!-- Footer -->
          <div v-if="$slots.footer" class="px-6 py-4 border-t border-secondary-200 dark:border-secondary-700 bg-secondary-50 dark:bg-secondary-900 rounded-b-xl">
            <slot name="footer" />
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup>
import { ref, computed, watch, onMounted, onUnmounted, nextTick } from 'vue'
import AppIconButton from './AppIconButton.vue'

const props = defineProps({
  show: {
    type: Boolean,
    default: false,
  },
  title: {
    type: String,
    default: '',
  },
  size: {
    type: String,
    default: 'md',
    validator: (v) => ['sm', 'md', 'lg'].includes(v),
  },
})

const emit = defineEmits(['close'])

const modalRef = ref(null)
const titleId = `modal-title-${Math.random().toString(36).slice(2, 9)}`

const sizeClass = computed(() => ({
  'max-w-sm': props.size === 'sm',
  'max-w-lg': props.size === 'md',
  'max-w-2xl': props.size === 'lg',
}))

function handleKeydown(e) {
  if (e.key === 'Escape' && props.show) {
    emit('close')
  }
}

function focusFirstElement() {
  if (!modalRef.value) return
  const focusable = modalRef.value.querySelectorAll(
    'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
  )
  if (focusable.length) focusable[0].focus()
}

watch(
  () => props.show,
  async (val) => {
    if (val) {
      await nextTick()
      focusFirstElement()
    }
  }
)

onMounted(() => {
  document.addEventListener('keydown', handleKeydown)
})

onUnmounted(() => {
  document.removeEventListener('keydown', handleKeydown)
})
</script>

<style scoped>
.modal-enter-active,
.modal-leave-active {
  transition: opacity 0.2s ease;
}
.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}
</style>
