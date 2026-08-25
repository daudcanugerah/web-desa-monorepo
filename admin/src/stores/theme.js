import { defineStore } from 'pinia'

const STORAGE_KEY = 'theme'

function readStored() {
  try {
    const v = localStorage.getItem(STORAGE_KEY)
    return v === 'light' || v === 'dark' || v === 'auto' ? v : 'auto'
  } catch {
    return 'auto'
  }
}

function systemPrefersDark() {
  return typeof window !== 'undefined'
    && typeof window.matchMedia === 'function'
    && window.matchMedia('(prefers-color-scheme: dark)').matches
}

function resolve(mode) {
  return mode === 'auto' ? (systemPrefersDark() ? 'dark' : 'light') : mode
}

export const useThemeStore = defineStore('theme', {
  state: () => ({
    mode: readStored(),
  }),
  getters: {
    isDark: (state) => resolve(state.mode) === 'dark',
  },
  actions: {
    init() {
      this.apply()
      if (typeof window !== 'undefined' && typeof window.matchMedia === 'function') {
        const mq = window.matchMedia('(prefers-color-scheme: dark)')
        const handler = () => {
          if (this.mode === 'auto') this.apply()
        }
        if (mq.addEventListener) mq.addEventListener('change', handler)
        else if (mq.addListener) mq.addListener(handler)
      }
    },
    set(mode) {
      if (mode !== 'light' && mode !== 'dark' && mode !== 'auto') return
      this.mode = mode
      try { localStorage.setItem(STORAGE_KEY, mode) } catch { /* ignore */ }
      this.apply()
    },
    cycle() {
      this.set(this.mode === 'light' ? 'dark' : this.mode === 'dark' ? 'auto' : 'light')
    },
    apply() {
      const dark = this.isDark
      document.documentElement.classList.toggle('dark', dark)
    },
  },
})