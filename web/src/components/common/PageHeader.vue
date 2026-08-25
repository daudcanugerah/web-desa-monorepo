<template>
  <div class="mb-6">
    <h1 class="text-2xl sm:text-3xl font-bold text-gray-900 mb-1.5">{{ title }}</h1>
    <p v-if="subtitle" class="text-sm text-gray-600">{{ subtitle }}</p>
    <slot />
  </div>
</template>

<script>
export default {
  name: 'PageHeader',
  props: {
    title: { type: String, required: true },
    subtitle: { type: String, default: '' }
  }
}
</script>