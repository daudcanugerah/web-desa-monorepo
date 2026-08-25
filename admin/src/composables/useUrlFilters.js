import { ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

export function useUrlFilters(fields) {
  const route = useRoute()
  const router = useRouter()

  const refs = {}
  for (const field of fields) {
    refs[field] = ref(route.query[field] ?? '')
  }

  function syncToUrl() {
    const query = { ...route.query }
    let changed = false
    for (const field of fields) {
      const value = refs[field].value
      if (value === '' || value === null || value === undefined) {
        if (query[field] !== undefined) {
          delete query[field]
          changed = true
        }
      } else if (String(query[field] ?? '') !== String(value)) {
        query[field] = String(value)
        changed = true
      }
    }
    if (changed) {
      router.replace({ query })
    }
  }

  function readFromUrl() {
    for (const field of fields) {
      refs[field].value = route.query[field] ?? ''
    }
  }

  function reset() {
    for (const field of fields) {
      refs[field].value = ''
    }
    syncToUrl()
  }

  watch(
    () => route.query,
    () => readFromUrl(),
  )

  return { refs, syncToUrl, readFromUrl, reset }
}