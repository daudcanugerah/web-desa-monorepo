import { ref, computed } from 'vue'

export function useBulkSelect() {
  const selected = ref(new Set())

  function isSelected(id) {
    return selected.value.has(id)
  }

  function toggle(id) {
    const next = new Set(selected.value)
    if (next.has(id)) next.delete(id)
    else next.add(id)
    selected.value = next
  }

  function toggleAll(ids) {
    const allSelected = ids.every((id) => selected.value.has(id))
    const next = new Set(selected.value)
    if (allSelected) ids.forEach((id) => next.delete(id))
    else ids.forEach((id) => next.add(id))
    selected.value = next
  }

  function clear() {
    selected.value = new Set()
  }

  const allSelected = computed(() => false)

  return { selected, isSelected, toggle, toggleAll, clear, allSelected }
}