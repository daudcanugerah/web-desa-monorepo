<template>
  <AppModal
    :show="uiStore.confirmDialog.visible"
    :title="uiStore.confirmDialog.title"
    size="sm"
    @close="uiStore.closeConfirm(false)"
  >
    <p class="text-secondary-700 dark:text-secondary-300">{{ uiStore.confirmDialog.message }}</p>
    <p v-if="uiStore.confirmDialog.itemName" class="mt-2 font-semibold text-secondary-800 dark:text-secondary-100">
      {{ uiStore.confirmDialog.itemName }}
    </p>

    <template #footer>
      <div class="flex justify-end gap-3">
        <AppButton variant="secondary" @click="uiStore.closeConfirm(false)">
          Batal
        </AppButton>
        <AppButton variant="danger" :loading="uiStore.confirmDialog.loading" @click="uiStore.closeConfirm(true)">
          Hapus
        </AppButton>
      </div>
    </template>
  </AppModal>
</template>

<script setup>
import { useUiStore } from '../../stores/ui'
import AppModal from './AppModal.vue'
import AppButton from './AppButton.vue'

const uiStore = useUiStore()
</script>