<template>
  <button
    :type="type"
    :disabled="disabled || loading"
    :aria-disabled="disabled || loading"
    :class="[baseClasses, variantClasses, sizeClasses]"
    :title="title"
  >
    <span v-if="loading" class="inline-flex items-center gap-2">
      <svg class="animate-spin h-4 w-4" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
        <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
        <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
      </svg>
      <slot />
    </span>
    <slot v-else />
  </button>
</template>

<script setup>
const props = defineProps({
  variant: {
    type: String,
    default: 'ghost',
    validator: (v) => ['primary', 'secondary', 'danger', 'success', 'outline', 'ghost'].includes(v),
  },
  size: {
    type: String,
    default: 'md',
    validator: (v) => ['sm', 'md', 'lg', 'icon'].includes(v),
  },
  loading: {
    type: Boolean,
    default: false,
  },
  disabled: {
    type: Boolean,
    default: false,
  },
  type: {
    type: String,
    default: 'button',
  },
  title: {
    type: String,
    default: '',
  },
})

const baseClasses =
  'rounded font-medium transition-colors focus:outline-none focus:ring-2 focus:ring-primary-500 disabled:opacity-50 disabled:cursor-not-allowed inline-flex items-center justify-center'

const variantMap = {
  primary: 'bg-primary-600 hover:bg-primary-700 text-white dark:bg-primary-500 dark:hover:bg-primary-400',
  secondary: 'bg-secondary-200 hover:bg-secondary-300 text-secondary-800 dark:bg-secondary-700 dark:hover:bg-secondary-600 dark:text-secondary-100',
  danger: 'bg-danger-600 hover:bg-danger-700 text-white dark:bg-danger-500 dark:hover:bg-danger-400',
  success: 'bg-success-600 hover:bg-success-700 text-white dark:bg-success-500 dark:hover:bg-success-400',
  outline: 'border-2 border-primary-600 text-primary-600 hover:bg-primary-50 dark:border-primary-400 dark:text-primary-300 dark:hover:bg-primary-900/30',
  ghost: 'text-secondary-600 hover:bg-secondary-100 dark:text-secondary-300 dark:hover:bg-secondary-700',
}

const sizeMap = {
  sm: 'p-1 text-xs',
  md: 'p-1.5 text-sm',
  lg: 'p-2 text-base',
  icon: 'p-1.5 text-sm',
}

const variantClasses = variantMap[props.variant] ?? variantMap.ghost
const sizeClasses = sizeMap[props.size] ?? sizeMap.md
</script>