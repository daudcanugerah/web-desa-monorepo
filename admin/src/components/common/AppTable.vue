<template>
  <div class="w-full overflow-x-auto">
    <table class="w-full text-sm text-left border-collapse">
      <thead>
        <tr>
          <th
            v-if="selectable"
            :style="{ width: '40px' }"
            class="bg-secondary-100 dark:bg-secondary-800 px-3 py-3 border-b border-secondary-200 dark:border-secondary-700"
          >
            <input
              type="checkbox"
              class="rounded border-secondary-300 dark:border-secondary-600 text-primary-600 focus:ring-primary-500"
              :checked="allOnPageSelected"
              :indeterminate.prop="someOnPageSelected && !allOnPageSelected"
              aria-label="Pilih semua baris di halaman ini"
              @change="$emit('toggle-all', pageRowIds)"
            />
          </th>
          <th
            v-for="col in columns"
            :key="col.key"
            :style="col.width ? { width: col.width } : {}"
            class="bg-secondary-100 dark:bg-secondary-800 text-secondary-600 dark:text-secondary-400 uppercase text-xs font-semibold px-4 py-3 border-b border-secondary-200 dark:border-secondary-700"
          >
            {{ col.label }}
          </th>
        </tr>
      </thead>
      <tbody>
        <!-- Loading skeleton -->
        <template v-if="loading">
          <tr v-for="i in 5" :key="`skeleton-${i}`" class="border-b border-secondary-200 dark:border-secondary-700">
            <td v-if="selectable" class="px-3 py-3">
              <div class="h-4 w-4 bg-secondary-200 dark:bg-secondary-700 rounded animate-pulse" />
            </td>
            <td v-for="col in columns" :key="col.key" class="px-4 py-3">
              <div class="h-4 bg-secondary-200 dark:bg-secondary-700 rounded animate-pulse" />
            </td>
          </tr>
        </template>

        <!-- Empty state -->
        <template v-else-if="data.length === 0">
          <tr>
            <td :colspan="columns.length + (selectable ? 1 : 0)" class="px-4 py-8 text-center text-secondary-500 dark:text-secondary-400">
              <slot name="empty">No data found.</slot>
            </td>
          </tr>
        </template>

        <!-- Data rows -->
        <template v-else>
          <tr
            v-for="(row, rowIndex) in data"
            :key="rowIndex"
            :class="[
              'border-b border-secondary-200 dark:border-secondary-700 transition-colors',
              rowId(row) && isSelected(rowId(row)) ? 'bg-primary-50 dark:bg-primary-900/20' : 'hover:bg-secondary-50 dark:hover:bg-secondary-800/50',
            ]"
          >
            <td v-if="selectable" class="px-3 py-3">
              <input
                type="checkbox"
                class="rounded border-secondary-300 dark:border-secondary-600 text-primary-600 focus:ring-primary-500"
                :checked="rowId(row) && isSelected(rowId(row))"
                :aria-label="`Pilih ${rowLabel(row)}`"
                @change="$emit('toggle', rowId(row))"
              />
            </td>
            <td
              v-for="col in columns"
              :key="col.key"
              class="px-4 py-3 text-secondary-700 dark:text-secondary-200"
            >
              <slot :name="col.key" :row="row">
                {{ row[col.key] }}
              </slot>
            </td>
          </tr>
        </template>
      </tbody>
    </table>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  columns: {
    type: Array,
    required: true,
  },
  data: {
    type: Array,
    default: () => [],
  },
  loading: {
    type: Boolean,
    default: false,
  },
  selectable: {
    type: Boolean,
    default: false,
  },
  selected: {
    type: Set,
    default: () => new Set(),
  },
})

const emit = defineEmits(['toggle', 'toggle-all'])

function rowId(row) {
  return row?.id
}

function rowLabel(row) {
  return row?.name || row?.title || row?.email || row?.id || 'baris'
}

const pageRowIds = computed(() => props.data.map((r) => r?.id).filter((id) => id != null))

function isSelected(id) {
  if (id == null) return false
  return props.selected.has(id)
}

const allOnPageSelected = computed(() =>
  pageRowIds.value.length > 0 && pageRowIds.value.every((id) => props.selected.has(id)),
)

const someOnPageSelected = computed(() =>
  pageRowIds.value.some((id) => props.selected.has(id)),
)

defineExpose({ rowId, rowLabel, pageRowIds, isSelected })
</script>