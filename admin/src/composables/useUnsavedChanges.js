import { computed, watch, onBeforeUnmount, getCurrentInstance } from 'vue'
import { onBeforeRouteLeave } from 'vue-router'

function snapshot(obj) {
  try {
    return JSON.stringify(obj)
  } catch {
    return null
  }
}

export function useUnsavedChanges(formRef, { message = 'Perubahan belum disimpan. Yakin ingin meninggalkan halaman ini?' } = {}) {
  const initial = snapshot(formRef)
  const dirty = computed(() => snapshot(formRef) !== initial)

  if (typeof window !== 'undefined') {
    const onBeforeUnload = (e) => {
      if (!dirty.value) return
      e.preventDefault()
      e.returnValue = message
      return message
    }
    window.addEventListener('beforeunload', onBeforeUnload)

    const instance = getCurrentInstance()
    if (instance) {
      onBeforeUnmount(() => {
        window.removeEventListener('beforeunload', onBeforeUnload)
      })
    }
  }

  onBeforeRouteLeave(() => {
    if (!dirty.value) return true
    return window.confirm(message)
  })

  function reset() {
    const snap = snapshot(formRef)
    watch(
      () => snapshot(formRef),
      () => {},
      { once: true },
    )
    return snap === initial
  }

  return { dirty }
}