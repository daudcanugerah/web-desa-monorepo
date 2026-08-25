<template>
  <div class="min-h-screen bg-gray-50">
    <div class="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
      <RouterLink to="/berita" class="inline-flex items-center text-emerald-600 hover:text-emerald-700 font-medium mb-8 group">
        <svg class="w-5 h-5 mr-2 group-hover:-translate-x-1 transition-transform" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
        </svg>
        Kembali ke Berita
      </RouterLink>

      <LoadingSpinner v-if="loading" />
      <article v-else-if="data" class="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden">
        <div class="h-64 sm:h-96 bg-gradient-to-br from-emerald-100 to-teal-100 relative overflow-hidden">
          <img
            v-if="coverUrl"
            :src="coverUrl"
            :alt="data.title"
            loading="eager"
            class="absolute inset-0 w-full h-full object-cover"
            @error="handleImageError"
          />
          <div class="absolute inset-0 bg-gradient-to-t from-black/40 to-transparent"></div>
        </div>
        <div class="p-6 sm:p-10 md:p-12">
          <div class="flex flex-wrap items-center gap-3 mb-4">
            <span class="inline-block bg-emerald-100 text-emerald-700 text-sm px-4 py-1 rounded-full font-medium">
              {{ data.category || 'Umum' }}
            </span>
            <span class="text-sm text-gray-500">{{ formatDate(getDate(data)) }}</span>
          </div>
          <h1 class="text-3xl sm:text-4xl font-bold text-gray-900 mb-6 leading-tight">{{ data.title }}</h1>
          <div class="prose prose-lg max-w-none text-gray-700 leading-relaxed" v-html="sanitizedContent"></div>
        </div>
      </article>
      <EmptyState v-else message="Berita tidak ditemukan" />
    </div>
  </div>
</template>

<script>
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { getPublicBeritaById, mediaUrl } from '../services/desaService'
import LoadingSpinner from '../components/common/LoadingSpinner.vue'
import EmptyState from '../components/common/EmptyState.vue'

const ALLOWED_TAGS = new Set(['p', 'br', 'strong', 'em', 'u', 'ul', 'ol', 'li', 'a', 'h1', 'h2', 'h3', 'h4', 'blockquote', 'img', 'figure', 'figcaption', 'code', 'pre'])
const ALLOWED_ATTRS = new Set(['href', 'title', 'target', 'rel', 'src', 'alt'])

const sanitizeHtml = (html) => {
  if (!html || typeof html !== 'string') return ''
  const doc = new DOMParser().parseFromString(html, 'text/html')
  const walk = (node) => {
    const children = Array.from(node.childNodes)
    for (const child of children) {
      if (child.nodeType === 1) {
        if (!ALLOWED_TAGS.has(child.tagName.toLowerCase())) {
          child.replaceWith(...Array.from(child.childNodes))
          continue
        }
        for (const attr of Array.from(child.attributes)) {
          const isSafeHref = attr.name === 'href' && /^(https?:|mailto:|#|\/)/i.test(attr.value)
          const isSafeSrc = attr.name === 'src' && /^(https?:|\/)/i.test(attr.value)
          if (!ALLOWED_ATTRS.has(attr.name) || ((attr.name === 'href' && !isSafeHref) || (attr.name === 'src' && !isSafeSrc))) {
            child.removeAttribute(attr.name)
          }
        }
        walk(child)
      } else if (child.nodeType !== 3) {
        child.remove()
      }
    }
  }
  walk(doc.body)
  return doc.body.innerHTML
}

export default {
  name: 'BeritaDetail',
  components: {
    LoadingSpinner,
    EmptyState
  },
  setup() {
    const route = useRoute()
    const data = ref(null)
    const loading = ref(true)

    const formatDate = (dateString) => {
      if (!dateString) return '-'
      try {
        return new Date(dateString).toLocaleDateString('id-ID', {
          day: 'numeric',
          month: 'long',
          year: 'numeric'
        })
      } catch {
        return '-'
      }
    }

    const getDate = (article) => article?.created_at || null

    const coverUrl = computed(() => mediaUrl(data.value, 'single'))

    const handleImageError = (event) => { event.target.style.display = 'none' }

    const sanitizedContent = computed(() => {
      const c = data.value?.content
      if (!c) return ''
      if (typeof c !== 'string') return ''
      return sanitizeHtml(c)
    })

    onMounted(async () => {
      try {
        const id = route.params.id
        const result = await getPublicBeritaById(id)
        data.value = result
      } catch (error) {
        console.error('Error fetching berita detail:', error)
        data.value = null
      } finally {
        loading.value = false
      }
    })

    return {
      data,
      loading,
      formatDate,
      getDate,
      handleImageError,
      sanitizedContent,
      coverUrl
    }
  }
}
</script>
