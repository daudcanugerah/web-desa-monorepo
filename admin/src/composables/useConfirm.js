import { useUiStore } from '../stores/ui'

export function useConfirm() {
  const uiStore = useUiStore()
  return {
    confirm: (title, message, itemName = '') => uiStore.showConfirm({ title, message, itemName }),
    confirmDelete: (itemLabel, itemName = '') =>
      uiStore.showConfirm({
        title: `Hapus ${itemLabel}?`,
        message: `Tindakan ini tidak dapat dibatalkan. ${itemLabel} yang dihapus akan hilang secara permanen.`,
        itemName,
      }),
  }
}
