<template>
  <div
    v-for="i in count"
    :key="i"
    class="bg-white rounded-xl overflow-hidden shadow-sm border border-gray-200 flex flex-col"
  >
    <div class="aspect-video skeleton rounded-none"></div>
    <div class="p-6 space-y-3 flex-1">
      <div class="skeleton h-4 w-20"></div>
      <div class="skeleton h-5 w-full"></div>
      <div class="skeleton h-5 w-3/4"></div>
      <div class="skeleton h-3 w-1/2 mt-4"></div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'CardSkeletonGrid',
  props: {
    count: { type: Number, default: 6 }
  }
}
</script>