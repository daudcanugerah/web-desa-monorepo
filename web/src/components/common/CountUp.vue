<template>
  <span ref="el">{{ formatted }}</span>
</template>

<script>
import { ref, computed, onMounted, watch } from 'vue'

export default {
  name: 'CountUp',
  props: {
    value: { type: Number, required: true },
    duration: { type: Number, default: 1800 },
    decimals: { type: Number, default: 0 },
    prefix: { type: String, default: '' },
    suffix: { type: String, default: '' },
    separator: { type: String, default: '.' }
  },
  setup(props) {
    const display = ref(0)
    const started = ref(false)
    const el = ref(null)

    const easeOutCubic = (t) => 1 - Math.pow(1 - t, 3)

    const formatted = computed(() => {
      const n = Number(display.value.toFixed(props.decimals))
      const fixed = n.toFixed(props.decimals)
      if (props.separator && props.separator !== '.') {
        const [intPart, decPart] = fixed.split('.')
        const withSep = intPart.replace(/\B(?=(\d{3})+(?!\d))/g, props.separator)
        return `${props.prefix}${decPart ? withSep + '.' + decPart : withSep}${props.suffix}`
      }
      return `${props.prefix}${fixed}${props.suffix}`
    })

    const animate = () => {
      if (started.value) return
      started.value = true
      const start = performance.now()
      const tick = (now) => {
        const elapsed = now - start
        const progress = Math.min(elapsed / props.duration, 1)
        display.value = props.value * easeOutCubic(progress)
        if (progress < 1) requestAnimationFrame(tick)
        else display.value = props.value
      }
      requestAnimationFrame(tick)
    }

    onMounted(() => {
      if (!el.value || typeof IntersectionObserver === 'undefined') {
        display.value = props.value
        started.value = true
        return
      }
      const obs = new IntersectionObserver(([entry]) => {
        if (entry.isIntersecting) {
          animate()
          obs.disconnect()
        }
      }, { threshold: 0.2 })
      obs.observe(el.value)
    })

    watch(() => props.value, (v) => {
      if (started.value) display.value = v
    })

    return { el, formatted }
  }
}
</script>