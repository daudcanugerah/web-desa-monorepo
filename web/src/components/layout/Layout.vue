<template>
  <div class="flex flex-col min-h-screen">
    <TopBar />
    <Navbar />
    <main class="flex-1 overflow-hidden">
      <RouterView />
    </main>
    <Footer />
  </div>
</template>

<script>
import { onMounted, watch } from 'vue'
import TopBar from './TopBar.vue'
import Navbar from './Navbar.vue'
import Footer from './Footer.vue'
import { useDesaInfo } from '../../composables/useDesaInfo'

export default {
  name: 'Layout',
  components: {
    TopBar,
    Navbar,
    Footer
  },
  setup() {
    const { desaInfo, loadDesa } = useDesaInfo()

    onMounted(() => {
      loadDesa()
    })

    watch(desaInfo, (val) => {
      if (val?.name) {
        document.title = `${val.name} - Website Resmi`
      }
    }, { immediate: true })

    return {}
  }
}
</script>
