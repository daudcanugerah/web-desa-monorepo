<template>
  <div class="flex items-center gap-2">
    <AppSelect :model-value="currentLocale" @update:model-value="(v) => (currentLocale = v)">
      <option v-for="locale in availableLocales" :key="locale.code" :value="locale.code">
        {{ locale.name }}
      </option>
    </AppSelect>
  </div>
</template>

<script setup>
import AppSelect from './AppSelect.vue'
import { useLocale } from '../../composables/useLocale'

const { currentLocale, availableLocales } = useLocale()
</script>