<template>
  <div class="min-h-screen bg-gray-50">
    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
      <div v-if="!selectedPPID">
        <PageHeader
          title="PPID"
          :subtitle="`Pejabat Pengelola Informasi dan Dokumentasi - Akses informasi publik dan dokumen resmi ${desaName}`"
        />

        <LoadingSpinner v-if="loading" />

        <div v-else-if="loadError" class="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg text-sm">
          {{ loadError }}
        </div>

        <div v-else-if="data && data.length > 0">
          <div class="bg-white rounded-xl shadow-sm border border-gray-200 p-4 mb-6">
            <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
              <div>
                <label class="block text-xs font-medium text-gray-700 mb-1.5">Cari Dokumen</label>
                <SearchInput
                  v-model="searchQuery"
                  placeholder="Judul atau deskripsi..."
                  @clear="resetFilters"
                />
              </div>
              <div>
                <label class="block text-xs font-medium text-gray-700 mb-1.5">Rentang Tanggal</label>
                <div class="relative">
                  <select v-model="dateFilter" class="w-full px-3 py-2 text-sm border border-gray-200 rounded-lg focus:ring-2 focus:ring-emerald-500 focus:border-emerald-500 bg-gray-50 appearance-none transition-colors">
                    <option value="all">Semua Tanggal</option>
                    <option value="last30days">30 Hari Terakhir</option>
                    <option value="last90days">90 Hari Terakhir</option>
                    <option value="thisYear">Tahun Ini</option>
                  </select>
                  <svg class="absolute right-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400 pointer-events-none" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" /></svg>
                </div>
              </div>
              <div>
                <label class="block text-xs font-medium text-gray-700 mb-1.5">Kategori</label>
                <div class="relative">
                  <select v-model="selectedCategory" class="w-full px-3 py-2 text-sm border border-gray-200 rounded-lg focus:ring-2 focus:ring-emerald-500 focus:border-emerald-500 bg-gray-50 appearance-none transition-colors">
                    <option v-for="category in categories" :key="category" :value="category">{{ category }}</option>
                  </select>
                  <svg class="absolute right-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400 pointer-events-none" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" /></svg>
                </div>
              </div>
            </div>

            <div class="mt-4 pt-3 border-t border-gray-200 flex items-center justify-between">
              <p class="text-xs text-gray-600">
                Menampilkan <span class="font-bold text-emerald-600">{{ startIndex + 1 }}-{{ Math.min(endIndex, filteredDocuments.length) }}</span> dari <span class="font-bold">{{ filteredDocuments.length }}</span> dokumen
                <span v-if="filteredDocuments.length !== data.length" class="text-gray-400">({{ data.length }} total)</span>
              </p>
              <button v-if="hasActiveFilter" @click="resetFilters" class="text-xs text-emerald-600 hover:text-emerald-700 font-medium flex items-center gap-1.5">
                <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" /></svg>
                Reset Filter
              </button>
            </div>
          </div>

          <div v-if="filteredDocuments.length === 0" class="py-4">
            <EmptyState message="Tidak ada dokumen yang cocok" description="Coba ubah filter atau kata kunci pencarian." action-label="Reset Filter" @action="resetFilters" />
          </div>

          <div v-else class="space-y-3 mb-8">
            <article
              v-for="doc in paginatedDocuments"
              :key="doc.id"
              class="bg-white rounded-xl shadow-sm hover:shadow-md transition-all border border-gray-200 p-4"
            >
              <div class="flex items-start gap-4">
                <div v-if="thumb(doc)" class="w-24 h-24 flex-shrink-0 rounded-lg overflow-hidden bg-gray-200">
                  <img
                    :src="thumb(doc)"
                    :alt="doc.title"
                    loading="lazy"
                    class="w-full h-full object-cover"
                    @error="handleImageError"
                  />
                </div>
                <div v-else class="w-24 h-24 flex-shrink-0 bg-emerald-100 rounded-lg flex items-center justify-center">
                  <svg class="w-12 h-12 text-emerald-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M7 21h10a2 2 0 002-2V9.414a1 1 0 00-.293-.707l-5.414-5.414A1 1 0 0012.586 3H7a2 2 0 00-2 2v14a2 2 0 002 2z" />
                  </svg>
                </div>

                <div class="flex-1 min-w-0">
                  <h2 class="text-base font-bold text-gray-900 mb-1.5 line-clamp-2">{{ doc.title }}</h2>
                  <div class="flex flex-wrap items-center gap-2 mb-2">
                    <span class="inline-block bg-emerald-100 text-emerald-800 text-xs px-2 py-0.5 rounded-full font-medium">{{ doc.category }}</span>
                    <span class="text-xs text-gray-500">{{ formatDate(doc.publication_at) }}</span>
                  </div>
                  <p class="text-gray-600 text-sm mb-3 line-clamp-2">{{ stripHtml(doc.description) || 'Tidak ada deskripsi' }}</p>
                  <button
                    @click="viewDetail(doc)"
                    class="bg-emerald-600 hover:bg-emerald-700 text-white px-4 py-1.5 rounded-lg text-xs font-medium transition-colors inline-flex items-center gap-1.5"
                  >
                    Lihat Detail
                    <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" /></svg>
                  </button>
                </div>
              </div>
            </article>
          </div>

          <Pagination
            v-if="filteredDocuments.length > 0"
            v-model:current-page="currentPage"
            :total-pages="totalPages"
          />
        </div>

        <EmptyState
          v-else-if="!loading && data && data.length === 0"
          message="Belum ada dokumen tersedia"
        />
      </div>

      <div v-else>
        <button @click="backToList" class="mb-6 text-emerald-600 hover:text-emerald-700 font-medium inline-flex items-center gap-2 group">
          <svg class="w-4 h-4 group-hover:-translate-x-1 transition-transform" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" /></svg>
          Kembali ke Daftar
        </button>

        <article class="bg-white rounded-xl shadow-sm border border-gray-200 p-6 md:p-8">
          <header class="mb-6 pb-6 border-b border-gray-200">
            <h1 class="text-2xl md:text-3xl font-bold text-gray-900 mb-3">{{ selectedPPID.title }}</h1>
            <div class="flex flex-wrap items-center gap-3 text-sm">
              <span class="inline-block bg-emerald-100 text-emerald-800 px-3 py-1 rounded-full font-medium">{{ selectedPPID.category }}</span>
              <span class="text-gray-600">{{ formatDate(selectedPPID.publication_at) }}</span>
            </div>
          </header>

          <section class="mb-8">
            <h2 class="text-base font-semibold text-gray-900 mb-3">Deskripsi</h2>
            <div class="prose prose-sm max-w-none text-gray-700" v-html="sanitizedDescription"></div>
          </section>

          <section class="mb-8">
            <h2 class="text-base font-semibold text-gray-900 mb-3">Informasi Dokumen</h2>
            <dl class="grid grid-cols-1 sm:grid-cols-2 gap-4 text-sm">
              <div>
                <dt class="text-gray-500 mb-1">Kategori</dt>
                <dd class="font-medium text-gray-900">{{ selectedPPID.category }}</dd>
              </div>
              <div>
                <dt class="text-gray-500 mb-1">Tanggal Publikasi</dt>
                <dd class="font-medium text-gray-900">{{ formatDate(selectedPPID.publication_at) }}</dd>
              </div>
              <div>
                <dt class="text-gray-500 mb-1">Dibuat</dt>
                <dd class="font-medium text-gray-900">{{ formatDate(selectedPPID.created_at) }}</dd>
              </div>
              <div>
                <dt class="text-gray-500 mb-1">Diperbarui</dt>
                <dd class="font-medium text-gray-900">{{ formatDate(selectedPPID.updated_at) }}</dd>
              </div>
            </dl>
          </section>

          <div class="flex flex-wrap gap-3">
            <a
              v-if="doc(selectedPPID)"
              :href="doc(selectedPPID)"
              target="_blank"
              rel="noopener"
              class="inline-flex items-center px-5 py-2.5 bg-emerald-600 hover:bg-emerald-700 text-white rounded-lg font-medium transition-colors"
            >
              <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" /></svg>
              Unduh Dokumen
            </a>
            <button
              @click="openRequestDialog"
              class="inline-flex items-center px-5 py-2.5 bg-white border-2 border-emerald-600 text-emerald-700 hover:bg-emerald-50 rounded-lg font-medium transition-colors"
            >
              <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" /></svg>
              Minta Dokumen
            </button>
          </div>
        </article>
      </div>

      <div v-if="showRequestDialog" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4 animate-fade-in" @click.self="closeRequestDialog">
        <div class="bg-white rounded-xl shadow-xl max-w-md w-full p-6 max-h-[90vh] overflow-y-auto">
          <div class="flex items-center justify-between mb-4">
            <h3 class="text-lg font-bold text-gray-900">Minta Dokumen</h3>
            <button @click="closeRequestDialog" :disabled="isSubmitting" class="text-gray-400 hover:text-gray-600 disabled:opacity-50 p-1 hover:bg-gray-100 rounded-full transition-colors" aria-label="Close">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" /></svg>
            </button>
          </div>

          <div class="mb-6 p-4 bg-gray-50 rounded-lg border border-gray-200">
            <p class="text-xs text-gray-500 mb-1">Anda meminta:</p>
            <h4 class="font-semibold text-gray-900 text-sm">{{ selectedPPID.title }}</h4>
            <p class="text-xs text-gray-600 mt-1">{{ selectedPPID.category }}</p>
          </div>

          <form @submit.prevent="submitRequest" class="space-y-4">
            <div>
              <label for="requester-name" class="block text-sm font-medium text-gray-700 mb-1.5">Nama Lengkap <span class="text-red-500">*</span></label>
              <input id="requester-name" v-model="requesterName" type="text" placeholder="Nama lengkap Anda" :disabled="isSubmitting" required class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500 focus:border-emerald-500 disabled:opacity-50 transition-colors" />
            </div>

            <div>
              <label for="requester-email" class="block text-sm font-medium text-gray-700 mb-1.5">Alamat Email <span class="text-red-500">*</span></label>
              <input id="requester-email" v-model="requesterEmail" type="email" placeholder="email@contoh.com" :disabled="isSubmitting" required class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500 focus:border-emerald-500 disabled:opacity-50 transition-colors" />
              <p class="text-xs text-gray-500 mt-1">Email konfirmasi akan dikirim ke alamat ini.</p>
            </div>

            <div>
              <label for="request-purpose" class="block text-sm font-medium text-gray-700 mb-1.5">Tujuan Penggunaan <span class="text-gray-400 font-normal">(opsional)</span></label>
              <textarea id="request-purpose" v-model="requestPurpose" placeholder="Jelaskan tujuan Anda meminta dokumen ini" :disabled="isSubmitting" rows="3" class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500 focus:border-emerald-500 disabled:opacity-50 resize-none transition-colors"></textarea>
            </div>

            <div v-if="requestError" class="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg text-sm">
              {{ requestError }}
            </div>

            <div class="flex gap-3 pt-2">
              <button type="button" @click="closeRequestDialog" :disabled="isSubmitting" class="flex-1 px-4 py-2 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 transition-colors disabled:opacity-50 font-medium">Batal</button>
              <button type="submit" :disabled="isSubmitting" class="flex-1 px-4 py-2 bg-emerald-600 text-white rounded-lg hover:bg-emerald-700 transition-colors disabled:opacity-50 font-medium inline-flex items-center justify-center gap-2">
                <svg v-if="isSubmitting" class="w-4 h-4 animate-spin" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" /></svg>
                {{ isSubmitting ? 'Mengirim...' : 'Kirim Permintaan' }}
              </button>
            </div>
          </form>
        </div>
      </div>

      <div v-if="showSuccessDialog" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4 animate-fade-in" @click.self="closeSuccessDialog">
        <div class="bg-white rounded-xl shadow-xl max-w-md w-full p-6 text-center">
          <div class="w-16 h-16 bg-emerald-100 rounded-full flex items-center justify-center mx-auto mb-4">
            <svg class="w-8 h-8 text-emerald-600" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" /></svg>
          </div>
          <h3 class="text-lg font-bold text-gray-900 mb-2">Permintaan Berhasil!</h3>
          <p class="text-gray-600 mb-6 text-sm">
            Kami akan mengirimkan email konfirmasi ke <span class="font-semibold text-gray-900">{{ requesterEmail }}</span> dalam beberapa menit.
          </p>
          <button @click="closeSuccessDialog" class="w-full px-4 py-2 bg-emerald-600 text-white rounded-lg hover:bg-emerald-700 transition-colors font-medium">Tutup</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, onMounted, computed, watch } from 'vue'
import { getPublicPPIDList, getPublicPPIDById, createPPIDRequest, thumbnailUrl, documentUrl } from '../services/desaService'
import LoadingSpinner from '../components/common/LoadingSpinner.vue'
import EmptyState from '../components/common/EmptyState.vue'
import SearchInput from '../components/common/SearchInput.vue'
import PageHeader from '../components/common/PageHeader.vue'
import Pagination from '../components/common/Pagination.vue'
import { useDesaInfo } from '../composables/useDesaInfo'

const ALLOWED_TAGS = new Set(['p', 'br', 'strong', 'em', 'u', 'ul', 'ol', 'li', 'a', 'h1', 'h2', 'h3', 'h4', 'blockquote', 'code', 'pre'])
const ALLOWED_ATTRS = new Set(['href', 'title', 'target', 'rel'])

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
          if (!ALLOWED_ATTRS.has(attr.name) || (attr.name === 'href' && !/^(https?:|mailto:|#|\/)/i.test(attr.value))) {
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
  name: 'PPID',
  components: {
    LoadingSpinner,
    EmptyState,
    SearchInput,
    PageHeader,
    Pagination
  },
  setup() {
    const { desaInfo } = useDesaInfo()
    const desaName = computed(() => desaInfo.value?.name || 'Desa')
    const data = ref(null)
    const loading = ref(true)
    const loadError = ref('')
    const searchQuery = ref('')
    const dateFilter = ref('all')
    const selectedCategory = ref('Semua')
    const currentPage = ref(1)
    const itemsPerPage = 9
    const selectedPPID = ref(null)
    const showRequestDialog = ref(false)
    const showSuccessDialog = ref(false)
    const requesterName = ref('')
    const requesterEmail = ref('')
    const requestPurpose = ref('')
    const isSubmitting = ref(false)
    const requestError = ref('')

    const formatDate = (dateString) => {
      if (!dateString) return '-'
      try {
        return new Date(dateString).toLocaleDateString('id-ID', {
          year: 'numeric', month: 'short', day: 'numeric'
        })
      } catch { return '-' }
    }

    const handleImageError = (event) => { event.target.style.display = 'none' }

    const thumb = (item) => thumbnailUrl(item)
    const doc = (item) => documentUrl(item)

    const stripHtml = (html) => {
      if (!html) return ''
      const tmp = document.createElement('div')
      tmp.innerHTML = html
      return tmp.textContent || tmp.innerText || ''
    }

    const categories = computed(() => {
      if (!data.value) return ['Semua']
      const cats = [...new Set(data.value.map(doc => doc.category).filter(Boolean))]
      return ['Semua', ...cats]
    })

    const filteredDocuments = computed(() => {
      if (!data.value) return []
      let filtered = data.value

      if (selectedCategory.value !== 'Semua') {
        filtered = filtered.filter(doc => doc.category === selectedCategory.value)
      }

      if (searchQuery.value) {
        const query = searchQuery.value.toLowerCase().trim()
        if (query) {
          filtered = filtered.filter(doc =>
            doc.title.toLowerCase().includes(query) ||
            stripHtml(doc.description).toLowerCase().includes(query)
          )
        }
      }

      if (dateFilter.value !== 'all') {
        const now = new Date()
        let filterDate = new Date()
        switch (dateFilter.value) {
          case 'last30days': filterDate.setDate(now.getDate() - 30); break
          case 'last90days': filterDate.setDate(now.getDate() - 90); break
          case 'thisYear': filterDate = new Date(now.getFullYear(), 0, 1); break
        }
        filtered = filtered.filter(doc => {
          const docDate = doc.publication_at ? new Date(doc.publication_at) : null
          return docDate && docDate >= filterDate
        })
      }

      return filtered
    })

    const totalPages = computed(() => Math.max(1, Math.ceil(filteredDocuments.value.length / itemsPerPage)))
    const startIndex = computed(() => (currentPage.value - 1) * itemsPerPage)
    const endIndex = computed(() => startIndex.value + itemsPerPage)
    const paginatedDocuments = computed(() => filteredDocuments.value.slice(startIndex.value, endIndex.value))

    const hasActiveFilter = computed(() => Boolean(searchQuery.value) || dateFilter.value !== 'all' || selectedCategory.value !== 'Semua')

    const sanitizedDescription = computed(() => sanitizeHtml(selectedPPID.value?.description || ''))

    const viewDetail = (doc) => {
      selectedPPID.value = doc
      window.scrollTo({ top: 0, behavior: 'smooth' })
      fetchFullDetail(doc.id)
    }

    const fetchFullDetail = async (id) => {
      try {
        const result = await getPublicPPIDById(id)
        if (result) selectedPPID.value = result
      } catch (error) {
        console.error('Error fetching full PPID detail:', error)
      }
    }

    const backToList = () => {
      selectedPPID.value = null
      closeRequestDialog()
    }

    const openRequestDialog = () => {
      requesterName.value = ''
      requesterEmail.value = ''
      requestPurpose.value = ''
      requestError.value = ''
      showRequestDialog.value = true
    }

    const closeRequestDialog = () => {
      showRequestDialog.value = false
      requestError.value = ''
    }

    const closeSuccessDialog = () => {
      showSuccessDialog.value = false
      backToList()
    }

    const submitRequest = async () => {
      requestError.value = ''

      if (!requesterName.value.trim()) { requestError.value = 'Nama lengkap harus diisi'; return }
      if (!requesterEmail.value.trim()) { requestError.value = 'Alamat email harus diisi'; return }
      const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
      if (!emailRegex.test(requesterEmail.value)) { requestError.value = 'Format email tidak valid'; return }

      isSubmitting.value = true
      try {
        const result = await createPPIDRequest(selectedPPID.value.id, {
          requester_name: requesterName.value,
          requester_email: requesterEmail.value,
          purpose: requestPurpose.value || undefined
        })
        if (result) {
          showRequestDialog.value = false
          showSuccessDialog.value = true
        } else {
          requestError.value = 'Gagal mengirim permintaan. Silakan coba lagi.'
        }
      } catch (error) {
        console.error('Error submitting request:', error)
        requestError.value = error?.data?.error || error?.message || 'Terjadi kesalahan. Silakan coba lagi.'
      } finally {
        isSubmitting.value = false
      }
    }

    const resetFilters = () => {
      searchQuery.value = ''
      dateFilter.value = 'all'
      selectedCategory.value = 'Semua'
      currentPage.value = 1
    }

    watch([searchQuery, dateFilter, selectedCategory], () => {
      currentPage.value = 1
    })

    onMounted(async () => {
      loading.value = true
      loadError.value = ''
      try {
        data.value = await getPublicPPIDList()
      } catch (error) {
        console.error('Error fetching PPID data:', error)
        data.value = []
        loadError.value = error?.data?.error || 'Gagal memuat daftar dokumen.'
      } finally {
        loading.value = false
      }
    })

    return {
      desaName,
      data,
      loading,
      loadError,
      searchQuery,
      dateFilter,
      selectedCategory,
      currentPage,
      categories,
      filteredDocuments,
      totalPages,
      startIndex,
      endIndex,
      paginatedDocuments,
      hasActiveFilter,
      selectedPPID,
      showRequestDialog,
      showSuccessDialog,
      requesterName,
      requesterEmail,
      requestPurpose,
      isSubmitting,
      requestError,
      formatDate,
      handleImageError,
      thumb,
      doc,
      stripHtml,
      sanitizedDescription,
      viewDetail,
      backToList,
      openRequestDialog,
      closeRequestDialog,
      closeSuccessDialog,
      submitRequest,
      resetFilters
    }
  }
}
</script>