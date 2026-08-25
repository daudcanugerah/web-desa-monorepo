<template>
  <div :class="['bg-white shadow-md rounded-lg p-6', className]">
    <slot></slot>
  </div>
</template>

<script>
export default {
  name: 'Card',
  props: {
    className: {
      type: String,
      default: ''
    }
  }
}
</script>
