<template>
  <div class="flex flex-col gap-1">
    <label v-if="label" :for="id" class="text-sm font-medium text-secondary-700 dark:text-secondary-300">
      {{ label }}
      <span v-if="required" class="text-danger-500 ml-0.5" aria-hidden="true">*</span>
      <span v-else-if="hint" class="text-secondary-500 text-xs ml-1">({{ hint }})</span>
    </label>
    <input
      :id="id"
      :type="type"
      :value="modelValue"
      :placeholder="placeholder"
      :required="required"
      :disabled="disabled"
      :aria-describedby="error && errorId ? errorId : undefined"
      :class="[
        'w-full border rounded px-3 py-2 text-sm transition-colors focus:outline-none focus:ring-2 focus:ring-primary-500',
        disabled ? 'bg-secondary-100 dark:bg-secondary-800 cursor-not-allowed opacity-60' : 'bg-white dark:bg-secondary-800',
        error ? 'border-danger-500 focus:ring-danger-500 dark:border-danger-400' : 'border-secondary-300 dark:border-secondary-600 hover:border-secondary-400 dark:hover:border-secondary-500',
      ]"
      @input="$emit('update:modelValue', $event.target.value)"
    />
    <p
      v-if="error"
      :id="errorId"
      class="text-danger-500 dark:text-danger-400 text-xs mt-1"
      role="alert"
      aria-live="polite"
    >
      {{ error }}
    </p>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  modelValue: {
    type: String,
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
  type: {
    type: String,
    default: 'text',
  },
  error: {
    type: String,
    default: '',
  },
  placeholder: {
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