<template>
  <div class="fixed top-4 right-4 z-50 flex flex-col gap-2">
    <TransitionGroup
      name="toast"
      tag="div"
      class="flex flex-col gap-2"
    >
      <div
        v-for="toast in notificationStore.toasts"
        :key="toast.id"
        class="flex items-start gap-3 rounded-lg border px-4 py-3 shadow-md min-w-64 max-w-sm backdrop-blur-sm"
        :class="toastClasses(toast.type)"
        role="alert"
        aria-live="polite"
      >
        <!-- Icon -->
        <span class="mt-0.5 shrink-0 text-lg" aria-hidden="true">
          {{ toastIcons(toast.type) }}
        </span>

        <!-- Message -->
        <div class="flex-1 min-w-0">
          <p class="text-sm font-medium">{{ toast.message }}</p>
          <button
            v-if="toast.action"
            type="button"
            class="mt-1 text-xs font-semibold underline underline-offset-2 hover:no-underline"
            @click="runUndo(toast)"
          >
            {{ toast.actionLabel || 'Batalkan' }}
          </button>
        </div>

        <!-- Close button -->
        <AppIconButton
          variant="ghost"
          size="sm"
          :aria-label="`Dismiss ${toast.type} notification`"
          class="opacity-80 hover:opacity-100"
          @click="notificationStore.remove(toast.id)"
        >
          <svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </AppIconButton>
      </div>
    </TransitionGroup>
  </div>
</template>

<script setup>
import { useNotificationStore } from '../../stores/notification'
import AppIconButton from './AppIconButton.vue'

const notificationStore = useNotificationStore()

function toastClasses(type) {
  if (type === 'success') return 'bg-green-50 border-green-400 text-green-800 dark:bg-green-900/30 dark:border-green-700 dark:text-green-200'
  if (type === 'undo') return 'bg-blue-50 border-blue-400 text-blue-800 dark:bg-blue-900/30 dark:border-blue-700 dark:text-blue-200'
  return 'bg-red-50 border-red-400 text-red-800 dark:bg-red-900/30 dark:border-red-700 dark:text-red-200'
}

function toastIcons(type) {
  if (type === 'success') return '✓'
  if (type === 'undo') return '↶'
  return '✕'
}

function runUndo(toast) {
  if (typeof toast.action === 'function') {
    toast.action()
  }
  notificationStore.remove(toast.id)
}
</script>

<style scoped>
.toast-enter-active,
.toast-leave-active {
  transition: all 0.3s ease;
}
.toast-enter-from {
  opacity: 0;
  transform: translateX(100%);
}
.toast-leave-to {
  opacity: 0;
  transform: translateX(100%);
}
</style>