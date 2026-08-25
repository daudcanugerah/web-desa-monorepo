import { ref } from 'vue'
import { useNotificationStore } from '../stores/notification'

export function useCategoryInput(service) {
  const suggestions = ref([])
  const loading = ref(false)
  const notificationStore = useNotificationStore()

  function extract(payload) {
    if (Array.isArray(payload)) return payload
    if (Array.isArray(payload?.data)) return payload.data
    if (Array.isArray(payload?.categories)) return payload.categories
    if (Array.isArray(payload?.data?.categories)) return payload.data.categories
    if (Array.isArray(payload?.banner_categories)) return payload.banner_categories
    if (Array.isArray(payload?.data?.banner_categories)) return payload.data.banner_categories
    return []
  }

  async function load() {
    loading.value = true
    try {
      const res = await service.getCategories({ limit: 100 })
      const list = extract(res.data)
      suggestions.value = (Array.isArray(list) ? list : [])
        .filter((c) => c && (c.id || c.name))
        .map((c) => ({ id: c.id, name: c.name }))
    } catch {
      suggestions.value = []
    } finally {
      loading.value = false
    }
  }

  async function resolve(category) {
    const name = (category?.name || '').trim()
    if (!name) return ''
    if (category?.id) {
      const match = suggestions.value.find((c) => c.id === category.id)
      if (match && match.name?.toLowerCase() === name.toLowerCase()) return match.id
    }
    const localMatch = suggestions.value.find((c) => c.name?.toLowerCase() === name.toLowerCase())
    if (localMatch) return localMatch.id
    try {
      const res = await service.createCategory({ name })
      const created = res?.data || {}
      if (!created.id) throw new Error('Backend did not return category id')
      suggestions.value.unshift({ id: created.id, name })
      return created.id
    } catch (err) {
      notificationStore.error(err.response?.data?.error || 'Gagal membuat kategori baru')
      throw err
    }
  }

  function onSelected(category) {
    return { id: category.id, name: category.name }
  }

  return { suggestions, loading, load, resolve, onSelected }
}