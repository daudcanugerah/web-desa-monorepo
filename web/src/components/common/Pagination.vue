<template>
  <nav v-if="totalPages > 1" class="flex items-center justify-center" aria-label="Pagination">
    <div class="flex items-center gap-1">
      <button
        @click="go(currentPage - 1)"
        :disabled="currentPage === 1"
        class="px-3 py-1.5 text-sm rounded-lg disabled:text-gray-300 disabled:cursor-not-allowed text-gray-700 hover:bg-gray-100 disabled:hover:bg-transparent transition-colors"
        aria-label="Previous page"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
        </svg>
      </button>

      <template v-for="(item, idx) in pageItems" :key="idx">
        <span
          v-if="item === '...'"
          class="px-2 py-1.5 text-sm text-gray-400 select-none"
        >…</span>
        <button
          v-else
          @click="go(item)"
          :aria-current="item === currentPage ? 'page' : undefined"
          :class="[
            'min-w-[2rem] px-2.5 py-1.5 text-sm rounded-lg transition-colors',
            item === currentPage
              ? 'bg-emerald-600 text-white font-medium'
              : 'text-gray-700 hover:bg-gray-100'
          ]"
        >
          {{ item }}
        </button>
      </template>

      <button
        @click="go(currentPage + 1)"
        :disabled="currentPage === totalPages"
        class="px-3 py-1.5 text-sm rounded-lg disabled:text-gray-300 disabled:cursor-not-allowed text-gray-700 hover:bg-gray-100 disabled:hover:bg-transparent transition-colors"
        aria-label="Next page"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
        </svg>
      </button>
    </div>
  </nav>
</template>

<script>
import { computed } from 'vue'

export default {
  name: 'Pagination',
  props: {
    currentPage: { type: Number, required: true },
    totalPages: { type: Number, required: true },
    siblingCount: { type: Number, default: 1 },
    edgeCount: { type: Number, default: 1 }
  },
  emits: ['update:currentPage'],
  setup(props, { emit }) {
    const DOTS = '...'

    const pageRange = (start, end) => {
      const length = end - start + 1
      return Array.from({ length }, (_, i) => start + i)
    }

    const pageItems = computed(() => {
      const total = props.totalPages
      const siblings = props.siblingCount
      const edges = props.edgeCount
      const current = props.currentPage

      const totalNumbers = siblings * 2 + edges * 2 + 3
      if (totalNumbers >= total) {
        return pageRange(1, total)
      }

      const leftSibling = Math.max(current - siblings, edges + 2)
      const rightSibling = Math.min(current + siblings, total - edges - 1)

      const showLeftDots = leftSibling > edges + 2
      const showRightDots = rightSibling < total - edges - 1

      const items = []

      items.push(...pageRange(1, edges))

      if (showLeftDots) {
        items.push(DOTS)
      } else {
        items.push(...pageRange(edges + 1, leftSibling - 1))
      }

      items.push(...pageRange(leftSibling, rightSibling))

      if (showRightDots) {
        items.push(DOTS)
      } else {
        items.push(...pageRange(rightSibling + 1, total - edges))
      }

      items.push(...pageRange(total - edges + 1, total))

      return items
    })

    const go = (page) => {
      if (page < 1 || page > props.totalPages || page === props.currentPage) return
      emit('update:currentPage', page)
      window.scrollTo({ top: 0, behavior: 'smooth' })
    }

    return { pageItems, go }
  }
}
</script>