import { ref, readonly } from 'vue'
import { getPublicDesa } from '../services/desaService'

const desaInfo = ref({})
const loaded = ref(false)
const loading = ref(false)
let inflight = null

export const useDesaInfo = () => {
  const loadDesa = async () => {
    if (loaded.value || loading.value) return desaInfo.value
    if (inflight) return inflight

    loading.value = true
    inflight = (async () => {
      try {
        const data = await getPublicDesa()
        desaInfo.value = data || {}
      } catch (error) {
        console.error('Error loading desa info:', error)
        desaInfo.value = {}
      } finally {
        loaded.value = true
        loading.value = false
        inflight = null
      }
      return desaInfo.value
    })()
    return inflight
  }

  const initials = (name) => {
    if (!name) return 'DS'
    const parts = name.trim().split(/\s+/).filter(Boolean)
    if (parts.length === 0) return 'DS'
    if (parts.length === 1) return parts[0].slice(0, 2).toUpperCase()
    return (parts[0][0] + parts[1][0]).toUpperCase()
  }

  return {
    desaInfo: readonly(desaInfo),
    loaded: readonly(loaded),
    loading: readonly(loading),
    loadDesa,
    initials
  }
}