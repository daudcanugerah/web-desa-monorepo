import { defineStore } from 'pinia'

export const useUiStore = defineStore('ui', {
  state: () => ({
    // Mobile: hidden (true) / visible (false)
    sidebarCollapsed: true,
    // Desktop: icon-only mode (true) / full width (false)
    sidebarIconOnly: localStorage.getItem('sidebar_icon_only') === 'true',
    confirmDialog: {
      visible: false,
      title: '',
      message: '',
      itemName: '',
      resolve: null,
      loading: false,
    },
  }),

  actions: {
    // Mobile toggle (hamburger button in header)
    toggleSidebar() {
      this.sidebarCollapsed = !this.sidebarCollapsed
    },

    // Desktop icon-only toggle
    toggleIconOnly() {
      this.sidebarIconOnly = !this.sidebarIconOnly
      localStorage.setItem('sidebar_icon_only', String(this.sidebarIconOnly))
    },

    showConfirm({ title, message, itemName = '' }) {
      return new Promise((resolve) => {
        this.confirmDialog = {
          visible: true,
          title,
          message,
          itemName,
          resolve,
          loading: false,
        }
      })
    },

    setConfirmLoading(loading) {
      this.confirmDialog.loading = loading
    },

    closeConfirm(result) {
      if (this.confirmDialog.resolve) {
        this.confirmDialog.resolve(result)
      }
      this.confirmDialog = {
        visible: false,
        title: '',
        message: '',
        itemName: '',
        resolve: null,
        loading: false,
      }
    },
  },
})
