import { onMounted, onBeforeUnmount } from 'vue'

const isMac = typeof navigator !== 'undefined' && /Mac|iPhone|iPad/.test(navigator.platform)

export function useSaveShortcut(handler, { enabled = true } = {}) {
  function onKeyDown(e) {
    if (!enabled) return
    const isSaveCombo = (e.metaKey || e.ctrlKey) && e.key.toLowerCase() === 's'
    if (!isSaveCombo) return
    const target = e.target
    if (target && target instanceof HTMLElement) {
      const tag = target.tagName
      if (tag === 'TEXTAREA' || tag === 'SELECT') return
      if (target.isContentEditable) return
    }
    e.preventDefault()
    handler(e)
  }

  onMounted(() => {
    window.addEventListener('keydown', onKeyDown)
  })

  onBeforeUnmount(() => {
    window.removeEventListener('keydown', onKeyDown)
  })

  return {
    hint: `${isMac ? '⌘' : 'Ctrl'}+S`,
  }
}