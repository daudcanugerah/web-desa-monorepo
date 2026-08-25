<template>
  <div class="relative">
    <svg
      class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400 pointer-events-none"
      fill="none" stroke="currentColor" viewBox="0 0 24 24"
    >
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
    </svg>
    <input
      :value="modelValue"
      @input="onInput"
      :type="type"
      :placeholder="placeholder"
      :aria-label="placeholder"
      class="w-full pl-10 pr-10 py-2 text-sm border border-gray-200 rounded-lg focus:ring-2 focus:ring-emerald-500 focus:border-emerald-500 bg-gray-50 placeholder-gray-400 transition-colors"
    />
    <button
      v-if="modelValue"
      @click="clear"
      type="button"
      class="absolute right-2 top-1/2 -translate-y-1/2 p-1 text-gray-400 hover:text-gray-600 hover:bg-gray-100 rounded-full transition-colors"
      aria-label="Clear search"
      tabindex="-1"
    >
      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
      </svg>
    </button>
  </div>
</template>

<script>
export default {
  name: 'SearchInput',
  props: {
    modelValue: { type: String, default: '' },
    placeholder: { type: String, default: 'Cari...' },
    type: { type: String, default: 'text' }
  },
  emits: ['update:modelValue', 'clear'],
  setup(props, { emit }) {
    const onInput = (e) => emit('update:modelValue', e.target.value)
    const clear = () => emit('clear')
    return { onInput, clear }
  }
}
</script>