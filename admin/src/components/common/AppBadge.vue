<template>
  <span :class="[baseClasses, variantClasses, sizeClasses]">
    <slot />
  </span>
</template>

<script setup>
const props = defineProps({
  variant: {
    type: String,
    default: 'primary',
    validator: (v) => ['primary', 'secondary', 'success', 'warning', 'danger'].includes(v),
  },
  size: {
    type: String,
    default: 'md',
    validator: (v) => ['sm', 'md', 'lg'].includes(v),
  },
})

const baseClasses = 'inline-flex items-center rounded-full font-medium'

const variantMap = {
  primary: 'bg-primary-100 text-primary-700 dark:bg-primary-900/30 dark:text-primary-300',
  secondary: 'bg-secondary-100 text-secondary-700 dark:bg-secondary-800 dark:text-secondary-200',
  success: 'bg-success-100 text-success-700 dark:bg-success-900/30 dark:text-success-300',
  warning: 'bg-warning-100 text-warning-700 dark:bg-warning-900/30 dark:text-warning-300',
  danger: 'bg-danger-100 text-danger-700 dark:bg-danger-900/30 dark:text-danger-300',
}

const sizeMap = {
  sm: 'px-2 py-0.5 text-xs',
  md: 'px-2.5 py-0.5 text-xs',
  lg: 'px-3 py-1 text-sm',
}

const variantClasses = variantMap[props.variant] ?? variantMap.primary
const sizeClasses = sizeMap[props.size] ?? sizeMap.md
</script>