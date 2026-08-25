<template>
  <AppNotification />
  <ConfirmDialog />
  <router-view />
</template>

<script setup>
import AppNotification from './components/common/AppNotification.vue'
import ConfirmDialog from './components/common/ConfirmDialog.vue'
</script>
