<template>
  <div class="relative" ref="rootRef">
    <!-- Label -->
    <label v-if="label" :for="buttonId" class="block text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-1.5">
      {{ label }}
      <span v-if="required" class="text-danger-500 ml-0.5" aria-hidden="true">*</span>
    </label>

    <!-- Trigger -->
    <button
      :id="buttonId"
      type="button"
      class="w-full flex items-center justify-between gap-2 border rounded px-3 py-2 text-sm transition-colors focus:outline-none focus:ring-2 focus:ring-primary-500 bg-white dark:bg-secondary-800"
      :class="[
        error
          ? 'border-danger-500 focus:ring-danger-500 dark:border-danger-400'
          : 'border-secondary-300 dark:border-secondary-600 hover:border-secondary-400 dark:hover:border-secondary-500',
      ]"
      :aria-expanded="open"
      aria-haspopup="listbox"
      @click="toggle"
    >
      <!-- Selected value or placeholder -->
      <span
        v-if="selectedName"
        class="inline-flex items-center gap-1.5 min-w-0"
      >
        <AppBadge variant="primary" class="max-w-full">
          <span class="truncate">{{ selectedName }}</span>
          <button
            v-if="allowClear"
            type="button"
            class="ml-1 -mr-0.5 inline-flex items-center justify-center w-3.5 h-3.5 rounded-full bg-primary-700/30 hover:bg-primary-700/50 text-primary-100"
            :aria-label="`Hapus kategori ${selectedName}`"
            @click.stop="clear"
          >
            <svg class="w-2.5 h-2.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="3" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </AppBadge>
      </span>
      <span v-else class="text-secondary-400 dark:text-secondary-500 truncate">
        {{ placeholder }}
      </span>

      <span class="flex items-center gap-1 shrink-0">
        <button
          v-if="manageLabel"
          type="button"
          class="px-2 py-0.5 rounded text-xs font-medium text-secondary-500 dark:text-secondary-400 hover:text-primary-600 dark:hover:text-primary-400 hover:bg-secondary-100 dark:hover:bg-secondary-700 transition-colors"
          @click.stop="$emit('manage')"
        >
          {{ manageLabel }}
        </button>
        <svg
          class="w-4 h-4 text-secondary-400 dark:text-secondary-500 transition-transform"
          :class="open ? 'rotate-180' : ''"
          fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"
        >
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
        </svg>
      </span>
    </button>

    <!-- Dropdown -->
    <Transition
      enter-active-class="transition ease-out duration-100"
      enter-from-class="opacity-0 scale-95"
      enter-to-class="opacity-100 scale-100"
      leave-active-class="transition ease-in duration-75"
      leave-from-class="opacity-100 scale-100"
      leave-to-class="opacity-0 scale-95"
    >
      <div
        v-if="open"
        class="absolute z-30 mt-1 w-full bg-white dark:bg-secondary-800 rounded-lg shadow-lg border border-secondary-200 dark:border-secondary-700 overflow-hidden"
        role="listbox"
        :aria-label="label || 'Kategori'"
      >
        <!-- Search -->
        <div class="p-2 border-b border-secondary-200 dark:border-secondary-700">
          <AppSearchInput
            v-model="query"
            :placeholder="`Cari ${label.toLowerCase()}...`"
            width="full"
            :debounce="0"
            @search="onSearch"
          />
        </div>

        <!-- Loading -->
        <div v-if="loading" class="p-4 text-center text-sm text-secondary-400 dark:text-secondary-500">
          Memuat kategori...
        </div>

        <!-- Empty -->
        <div v-else-if="filtered.length === 0" class="p-4 text-center">
          <p class="text-sm text-secondary-500 dark:text-secondary-400">
            Tidak ada kategori yang cocok.
          </p>
          <button
            v-if="query.trim()"
            type="button"
            class="mt-2 inline-flex items-center gap-1 text-sm font-medium text-primary-600 dark:text-primary-400 hover:underline"
            @click="createNew"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
            </svg>
            Buat "{{ query.trim() }}"
          </button>
        </div>

        <!-- List -->
        <ul v-else class="max-h-56 overflow-y-auto p-1" @click="close">
          <li v-for="cat in filtered" :key="cat.id || cat.name">
            <button
              type="button"
              class="w-full text-left flex items-center justify-between gap-2 px-3 py-2 rounded text-sm transition-colors hover:bg-secondary-100 dark:hover:bg-secondary-700"
              :class="isSelected(cat) ? 'text-primary-600 dark:text-primary-400 font-medium' : 'text-secondary-700 dark:text-secondary-200'"
              role="option"
              :aria-selected="isSelected(cat)"
              @click="select(cat)"
            >
              <span class="truncate">{{ cat.name }}</span>
              <span v-if="isSelected(cat)" class="shrink-0">
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
                </svg>
              </span>
            </button>
          </li>
        </ul>

        <!-- Inline create (bottom) -->
        <div v-if="query.trim() && !filtered.some((c) => c.name?.toLowerCase() === query.trim().toLowerCase())" class="border-t border-secondary-200 dark:border-secondary-700 p-2">
          <button
            type="button"
            class="w-full flex items-center gap-2 px-3 py-2 rounded text-sm text-primary-600 dark:text-primary-400 font-medium hover:bg-primary-50 dark:hover:bg-primary-900/20 transition-colors"
            @click="createNew"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
            </svg>
            Buat "{{ query.trim() }}"
          </button>
        </div>
      </div>
    </Transition>

    <p v-if="error" class="text-danger-500 dark:text-danger-400 text-xs mt-1" role="alert">{{ error }}</p>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted, onBeforeUnmount } from 'vue'
import AppBadge from './AppBadge.vue'
import AppSearchInput from './AppSearchInput.vue'
import { useNotificationStore } from '../../stores/notification'

const props = defineProps({
  modelValue: {
    type: Object,
    default: () => ({ id: '', name: '' }),
  },
  suggestions: {
    type: Array,
    default: () => [],
  },
  service: { type: Object, required: true },
  label: { type: String, default: 'Kategori' },
  placeholder: { type: String, default: 'Pilih atau buat kategori...' },
  required: { type: Boolean, default: false },
  error: { type: String, default: '' },
  manageLabel: { type: String, default: 'Kelola' },
  allowClear: { type: Boolean, default: true },
  id: { type: String, default: 'app-category-picker' },
})

const emit = defineEmits(['update:modelValue', 'manage', 'created'])

const notificationStore = useNotificationStore()
const rootRef = ref(null)
const open = ref(false)
const query = ref('')
const loading = ref(false)
const creating = ref(false)

const buttonId = `${props.id}-trigger`

const selectedName = computed(() => props.modelValue?.name || '')
const normalizedQuery = computed(() => query.value.trim().toLowerCase())

const filtered = computed(() => {
  const list = props.suggestions || []
  if (!normalizedQuery.value) return list
  return list.filter((c) => (c.name || '').toLowerCase().includes(normalizedQuery.value))
})

function isSelected(cat) {
  return props.modelValue?.id && props.modelValue.id === cat.id
}

function toggle() {
  open.value = !open.value
  if (open.value) {
    query.value = props.modelValue?.name || ''
  }
}

function close() {
  open.value = false
}

function clear() {
  emit('update:modelValue', { id: '', name: '' })
}

function select(cat) {
  emit('update:modelValue', { id: cat.id, name: cat.name })
  close()
}

function onSearch() {
  // keep dropdown open; filtering is reactive via `filtered`
}

async function createNew() {
  const name = query.value.trim()
  if (!name || creating.value) return
  creating.value = true
  try {
    const res = await props.service.createCategory({ name })
    const created = res?.data || { name }
    const entry = { id: created.id || `local-${Date.now()}`, name: created.name || name }
    emit('update:modelValue', entry)
    emit('created', entry)
    close()
    notificationStore.success(`Kategori "${name}" dibuat`)
  } catch (err) {
    if (err.response?.status === 409) {
      // Exists already — select the matching existing one if present.
      const match = (props.suggestions || []).find((c) => c.name?.toLowerCase() === name.toLowerCase())
      if (match) {
        select(match)
      } else {
        notificationStore.error('Kategori sudah ada')
      }
    } else {
      notificationStore.error(err.response?.data?.error || 'Gagal membuat kategori')
    }
  } finally {
    creating.value = false
  }
}

function onClickOutside(e) {
  if (rootRef.value && !rootRef.value.contains(e.target)) {
    close()
  }
}

watch(open, (val) => {
  if (val) {
    document.addEventListener('click', onClickOutside, true)
  } else {
    document.removeEventListener('click', onClickOutside, true)
    query.value = ''
  }
})

onBeforeUnmount(() => {
  document.removeEventListener('click', onClickOutside, true)
})

// Reset search when suggestions change (e.g. after manage modal adds one)
watch(
  () => props.suggestions.length,
  () => {
    if (open.value) query.value = props.modelValue?.name || ''
  },
)

defineExpose({ open: () => { open.value = true } })
</script>