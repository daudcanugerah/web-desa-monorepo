import { defineStore } from 'pinia'

let toastIdCounter = 0

export const useNotificationStore = defineStore('notification', {
  state: () => ({
    toasts: [],
  }),

  actions: {
    add({ type, message, duration = 3000, action = null, actionLabel = '' }) {
      const id = ++toastIdCounter
      this.toasts.push({ id, type, message, duration, action, actionLabel })

      if (duration > 0) {
        setTimeout(() => {
          this.remove(id)
        }, duration)
      }

      return id
    },

    remove(id) {
      this.toasts = this.toasts.filter((t) => t.id !== id)
    },

    success(message, opts = {}) {
      return this.add({ type: 'success', message, duration: 3000, ...opts })
    },

    error(message, opts = {}) {
      return this.add({ type: 'error', message, duration: 0, ...opts })
    },

    undo(message, action, { duration = 5000, actionLabel = 'Batalkan' } = {}) {
      return this.add({
        type: 'undo',
        message,
        duration,
        action,
        actionLabel,
      })
    },
  },
})