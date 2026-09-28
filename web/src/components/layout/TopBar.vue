<template>
  <div class="bg-emerald-800 text-emerald-50 text-xs">
    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
      <div class="flex items-center justify-between h-9">
        <div class="hidden sm:flex items-center gap-4">
          <a href="https://kemendagri.go.id" target="_blank" rel="noopener" class="hover:text-white transition-colors inline-flex items-center gap-1">
            <Icon name="shield" class="w-3 h-3" />
            <span>Kemendagri</span>
          </a>
          <span class="text-emerald-600">|</span>
          <a href="https://kemendesa.go.id" target="_blank" rel="noopener" class="hover:text-white transition-colors">Kemendes PDTT</a>
        </div>

        <div class="flex items-center gap-4 ml-auto">
          <span class="hidden md:inline-flex items-center gap-1.5">
            <Icon name="calendar" class="w-3 h-3" />
            <span>{{ formattedDate }}</span>
          </span>
          <span class="text-emerald-600 hidden md:inline">|</span>
          <button class="hover:text-white transition-colors inline-flex items-center gap-1 font-medium" aria-label="Bahasa">
            <Icon name="globe" class="w-3 h-3" />
            <span>ID</span>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { computed } from 'vue'
import Icon from '../common/Icon.vue'

export default {
  name: 'TopBar',
  components: { Icon },
  setup() {
    const formattedDate = computed(() => {
      return new Date().toLocaleDateString('id-ID', {
        weekday: 'long',
        day: 'numeric',
        month: 'long',
        year: 'numeric'
      })
    })
    return { formattedDate }
  }
}
</script>