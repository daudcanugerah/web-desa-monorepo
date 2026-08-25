import { ref, watch, onMounted, onBeforeUnmount } from 'vue'

const DRAFT_PREFIX = 'admin_draft:'

export function useDraftAutosave(key, source, { debounceMs = 1500, ttlMs = 7 * 24 * 60 * 60 * 1000 } = {}) {
  const storageKey = `${DRAFT_PREFIX}${key}`
  const restored = ref(false)
  const lastSavedAt = ref(null)

  function read() {
    try {
      const raw = localStorage.getItem(storageKey)
      if (!raw) return null
      const parsed = JSON.parse(raw)
      if (Date.now() - parsed.savedAt > ttlMs) {
        localStorage.removeItem(storageKey)
        return null
      }
      return parsed.data
    } catch {
      return null
    }
  }

  function write() {
    try {
      lastSavedAt.value = Date.now()
      localStorage.setItem(
        storageKey,
        JSON.stringify({ savedAt: lastSavedAt.value, data: JSON.parse(JSON.stringify(source.value)) }),
      )
    } catch {
      // quota / privacy mode — silent
    }
  }

  function clear() {
    localStorage.removeItem(storageKey)
    lastSavedAt.value = null
  }

  let timer = null

  function scheduleSave() {
    if (timer) clearTimeout(timer)
    timer = setTimeout(() => {
      write()
      timer = null
    }, debounceMs)
  }

  const stop = watch(
    source,
    () => {
      if (restored.value) scheduleSave()
    },
    { deep: true },
  )

  function restore() {
    const data = read()
    if (data && typeof source.value === 'object' && source.value !== null) {
      Object.assign(source.value, data)
    }
    restored.value = true
  }

  onMounted(restore)
  onBeforeUnmount(() => {
    if (timer) clearTimeout(timer)
    stop()
  })

  return { restored, lastSavedAt, clear }
}