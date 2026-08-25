<template>
  <div class="flex flex-col gap-1">
    <label v-if="label" :for="id" class="text-sm font-medium text-secondary-700 dark:text-secondary-300">
      {{ label }}
      <span v-if="required" class="text-danger-500 ml-0.5" aria-hidden="true">*</span>
      <span v-else-if="hint" class="text-secondary-500 text-xs ml-1">({{ hint }})</span>
    </label>
    <div class="relative">
      <select
        :id="id"
        :value="modelValue"
        :required="required"
        :disabled="disabled"
        :aria-describedby="error && errorId ? errorId : undefined"
        :class="[
          'w-full appearance-none border rounded pl-3 pr-9 py-2 text-sm transition-colors focus:outline-none focus:ring-2 focus:ring-primary-500 text-secondary-800 dark:text-secondary-100',
          disabled ? 'bg-secondary-100 dark:bg-secondary-800 cursor-not-allowed opacity-60' : 'bg-white dark:bg-secondary-800',
          error ? 'border-danger-500 focus:ring-danger-500 dark:border-danger-400' : 'border-secondary-300 dark:border-secondary-600 hover:border-secondary-400 dark:hover:border-secondary-500',
        ]"
        @change="$emit('update:modelValue', $event.target.value)"
      >
        <slot />
      </select>
      <svg
        class="pointer-events-none absolute right-3 top-1/2 -translate-y-1/2 h-4 w-4 text-secondary-500 dark:text-secondary-400"
        fill="none"
        viewBox="0 0 24 24"
        stroke="currentColor"
        aria-hidden="true"
      >
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
      </svg>
    </div>
    <p v-if="error" :id="errorId" class="text-danger-500 dark:text-danger-400 text-xs mt-1">{{ error }}</p>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  modelValue: {
    type: [String, Number, Boolean],
    default: '',
  },
  label: {
    type: String,
    default: '',
  },
  hint: {
    type: String,
    default: '',
  },
  error: {
    type: String,
    default: '',
  },
  required: {
    type: Boolean,
    default: false,
  },
  disabled: {
    type: Boolean,
    default: false,
  },
  id: {
    type: String,
    default: '',
  },
})

defineEmits(['update:modelValue'])

const errorId = computed(() => (props.id ? `${props.id}-error` : undefined))
</script>