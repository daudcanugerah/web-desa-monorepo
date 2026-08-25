<template>
  <div :class="['flex items-center justify-center', $attrs.class]">
    <div class="relative">
      <div class="w-10 h-10 rounded-full border-2 border-gray-200"></div>
      <div class="absolute inset-0 w-10 h-10 rounded-full border-2 border-emerald-600 border-t-transparent animate-spin"></div>
    </div>
    <span v-if="label" class="ml-3 text-sm text-gray-600">{{ label }}</span>
  </div>
</template>

<script>
export default {
  name: 'LoadingSpinner',
  inheritAttrs: false,
  props: {
    label: { type: String, default: '' }
  }
}
</script>