<template>
  <div class="h-[calc(100dvh-116px)] flex overflow-hidden relative bg-gray-50">
    <!-- Center: document list / detail -->
    <div class="flex-1 flex flex-col overflow-hidden">
      <!-- ===== DETAIL VIEW ===== -->
      <template v-if="selectedPPID">
        <div class="flex-shrink-0 flex items-center gap-3 px-4 sm:px-6 py-3 bg-white border-b border-gray-200">
          <button
            @click="backToList"
            class="inline-flex items-center gap-1.5 text-sm font-medium text-emerald-600 hover:text-emerald-700 group flex-shrink-0"
          >
            <svg class="w-4 h-4 group-hover:-translate-x-1 transition-transform" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" /></svg>
            Kembali
          </button>
          <span class="h-4 w-px bg-gray-200 flex-shrink-0"></span>
          <h1 class="text-sm font-semibold text-gray-900 truncate">{{ selectedPPID.title }}</h1>
        </div>

        <div ref="scrollRef" class="flex-1 overflow-y-auto">
          <article class="max-w-3xl mx-auto bg-white border-b border-gray-200">
            <div v-if="detailThumb" class="w-full bg-gray-100">
              <img :src="detailThumb" :alt="selectedPPID.title" class="w-full max-h-[360px] object-contain" @error="handleImageError" />
            </div>

            <div class="p-5 md:p-6">
            <header class="mb-5 pb-5 border-b border-gray-200">
              <h1 class="text-xl md:text-2xl font-bold text-gray-900 mb-2.5">{{ selectedPPID.title }}</h1>
              <div class="flex flex-wrap items-center gap-2.5 text-xs">
                <span class="inline-block bg-emerald-100 text-emerald-800 px-2.5 py-0.5 rounded-full font-medium">{{ selectedPPID.category }}</span>
                <span class="text-gray-600">{{ formatDate(selectedPPID.publication_at) }}</span>
              </div>
            </header>

            <section class="mb-6">
              <h2 class="text-sm font-semibold text-gray-900 mb-2">Deskripsi</h2>
              <div class="text-sm text-gray-700 leading-relaxed" v-html="sanitizedDescription"></div>
            </section>

            <section class="mb-6">
              <h2 class="text-sm font-semibold text-gray-900 mb-2">Informasi Dokumen</h2>
              <dl class="grid grid-cols-1 sm:grid-cols-2 gap-3 text-xs">
                <div>
                  <dt class="text-gray-500 mb-0.5">Kategori</dt>
                  <dd class="font-medium text-gray-900">{{ selectedPPID.category }}</dd>
                </div>
                <div>
                  <dt class="text-gray-500 mb-0.5">Tanggal Publikasi</dt>
                  <dd class="font-medium text-gray-900">{{ formatDate(selectedPPID.publication_at) }}</dd>
                </div>
                <div>
                  <dt class="text-gray-500 mb-0.5">Dibuat</dt>
                  <dd class="font-medium text-gray-900">{{ formatDate(selectedPPID.created_at) }}</dd>
                </div>
                <div>
                  <dt class="text-gray-500 mb-0.5">Diperbarui</dt>
                  <dd class="font-medium text-gray-900">{{ formatDate(selectedPPID.updated_at) }}</dd>
                </div>
              </dl>
            </section>

            <button
              @click="openRequestDialog"
              class="inline-flex items-center px-4 py-2 text-sm bg-emerald-600 hover:bg-emerald-700 text-white rounded-lg font-medium transition-colors"
            >
              <svg class="w-4 h-4 mr-1.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" /></svg>
              Minta Dokumen
            </button>
            </div>
          </article>
        </div>
      </template>

      <!-- ===== LIST VIEW ===== -->
      <template v-else>
        <div class="flex-shrink-0 flex items-center gap-3 px-4 sm:px-6 py-3 bg-white border-b border-gray-200">
          <div class="min-w-0">
            <h1 class="text-sm font-semibold text-gray-900">PPID</h1>
            <p class="text-[11px] text-gray-500 truncate">Informasi &amp; dokumen publik {{ desaName }}</p>
          </div>
          <button
            @click="showFilters = true"
            class="md:hidden ml-auto inline-flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium text-white bg-emerald-600 rounded-lg hover:bg-emerald-700 transition-colors"
          >
            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 4a1 1 0 011-1h16a1 1 0 011 1v2a1 1 0 01-.293.707L14 13.414V19a1 1 0 01-1.447.894l-4-2A1 1 0 018 17v-3.586L3.293 6.707A1 1 0 013 6V4z" /></svg>
            Filter
            <span v-if="hasActiveFilter" class="w-1.5 h-1.5 rounded-full bg-white"></span>
          </button>
        </div>

        <div ref="scrollRef" class="flex-1 overflow-y-auto">
          <LoadingSpinner v-if="loading" class="py-10" />

          <div v-else-if="loadError" class="m-4 bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg text-sm">
            {{ loadError }}
          </div>

          <EmptyState
            v-else-if="!data || data.length === 0"
            class="py-16"
            message="Belum ada dokumen tersedia"
          />

          <template v-else>
            <div v-if="filteredDocuments.length === 0" class="py-16">
              <EmptyState message="Tidak ada dokumen yang cocok" description="Coba ubah filter atau kata kunci pencarian." action-label="Reset Filter" @action="resetFilters" />
            </div>

            <div v-else class="p-4 sm:p-6 space-y-2.5">
              <article
                v-for="doc in paginatedDocuments"
                :key="doc.id"
                class="bg-white rounded-lg shadow-sm hover:shadow-md transition-all border border-gray-200 p-3"
              >
                <div class="flex items-start gap-3">
                  <div v-if="thumb(doc)" class="w-20 h-20 flex-shrink-0 rounded-lg overflow-hidden bg-gray-200">
                    <img :src="thumb(doc)" :alt="doc.title" loading="lazy" class="w-full h-full object-cover" @error="handleImageError" />
                  </div>
                  <div v-else class="w-20 h-20 flex-shrink-0 bg-emerald-100 rounded-lg flex items-center justify-center">
                    <svg class="w-9 h-9 text-emerald-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M7 21h10a2 2 0 002-2V9.414a1 1 0 00-.293-.707l-5.414-5.414A1 1 0 0012.586 3H7a2 2 0 00-2 2v14a2 2 0 002 2z" />
                    </svg>
                  </div>

                  <div class="flex-1 min-w-0">
                    <h2 class="text-sm font-bold text-gray-900 mb-1 line-clamp-2">{{ doc.title }}</h2>
                    <div class="flex flex-wrap items-center gap-2 mb-1.5">
                      <span class="inline-block bg-emerald-100 text-emerald-800 text-[11px] px-2 py-0.5 rounded-full font-medium">{{ doc.category }}</span>
                      <span class="text-[11px] text-gray-500">{{ formatDate(doc.publication_at) }}</span>
                    </div>
                    <p class="text-gray-600 text-xs mb-2 line-clamp-2">{{ stripHtml(doc.description) || 'Tidak ada deskripsi' }}</p>
                    <button
                      @click="viewDetail(doc)"
                      class="bg-emerald-600 hover:bg-emerald-700 text-white px-3 py-1 rounded-lg text-[11px] font-medium transition-colors inline-flex items-center gap-1"
                    >
                      Lihat Detail
                      <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" /></svg>
                    </button>
                  </div>
                </div>
              </article>
            </div>
          </template>
        </div>

        <div v-if="!loading && filteredDocuments.length > 0" class="flex-shrink-0 border-t border-gray-200 bg-white py-2.5 px-4">
          <Pagination
            v-model:current-page="currentPage"
            :total-pages="totalPages"
          />
        </div>
      </template>
    </div>

    <!-- Desktop sidebar: controls (hidden while a document detail is open) -->
    <aside v-if="!selectedPPID" class="hidden md:flex w-80 bg-white border-l border-gray-200 flex-col overflow-hidden z-[50]">
      <div class="flex-shrink-0 px-4 py-3 border-b border-gray-200">
        <h2 class="text-sm font-semibold text-gray-900 mb-3">Filter Dokumen</h2>
        <div class="space-y-3">
          <div>
            <label class="block text-[11px] font-medium text-gray-700 mb-1">Cari Dokumen</label>
            <SearchInput
              v-model="searchQuery"
              placeholder="Judul atau deskripsi..."
              size="sm"
              @clear="resetFilters"
            />
          </div>
          <div>
            <label class="block text-[11px] font-medium text-gray-700 mb-1">Rentang Tanggal</label>
            <div class="relative">
              <select v-model="dateFilter" class="w-full px-2.5 py-1.5 text-xs border border-gray-200 rounded-lg focus:ring-2 focus:ring-emerald-500 focus:border-emerald-500 bg-gray-50 appearance-none transition-colors">
                <option value="all">Semua Tanggal</option>
                <option value="last30days">30 Hari Terakhir</option>
                <option value="last90days">90 Hari Terakhir</option>
                <option value="thisYear">Tahun Ini</option>
              </select>
              <svg class="absolute right-2.5 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-gray-400 pointer-events-none" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" /></svg>
            </div>
          </div>
          <div>
            <label class="block text-[11px] font-medium text-gray-700 mb-1">Kategori</label>
            <div class="relative">
              <select v-model="selectedCategory" class="w-full px-2.5 py-1.5 text-xs border border-gray-200 rounded-lg focus:ring-2 focus:ring-emerald-500 focus:border-emerald-500 bg-gray-50 appearance-none transition-colors">
                <option v-for="category in categories" :key="category" :value="category">{{ category }}</option>
              </select>
              <svg class="absolute right-2.5 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-gray-400 pointer-events-none" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" /></svg>
            </div>
          </div>
        </div>
      </div>

      <div class="flex-1"></div>

      <div v-if="!loading && data && data.length > 0" class="flex-shrink-0 px-4 py-3 border-t border-gray-100 bg-gray-50">
        <p class="text-[11px] text-gray-600">
          Menampilkan <span class="font-bold text-emerald-600">{{ startIndex + 1 }}-{{ Math.min(endIndex, filteredDocuments.length) }}</span> dari <span class="font-bold">{{ filteredDocuments.length }}</span> dokumen
          <span v-if="filteredDocuments.length !== data.length" class="text-gray-400">({{ data.length }} total)</span>
        </p>
        <button
          v-if="hasActiveFilter"
          @click="resetFilters"
          class="mt-1.5 text-[11px] text-emerald-600 hover:text-emerald-700 font-medium inline-flex items-center gap-1"
        >
          <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" /></svg>
          Reset Filter
        </button>
      </div>
    </aside>

    <!-- Mobile filter bottom sheet -->
    <Transition name="slide-up">
      <div v-if="showFilters" class="absolute inset-x-0 bottom-0 z-[600] bg-white rounded-t-2xl shadow-xl max-h-[70vh] flex flex-col md:hidden">
        <div class="px-4 py-3 border-b border-gray-200 flex items-center justify-between">
          <h2 class="text-sm font-semibold text-gray-900">Filter Dokumen</h2>
          <button @click="showFilters = false" class="p-1.5 hover:bg-gray-100 rounded-full">
            <svg class="w-4 h-4 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" /></svg>
          </button>
        </div>
        <div class="px-4 py-3 space-y-3 overflow-y-auto">
          <div>
            <label class="block text-[11px] font-medium text-gray-700 mb-1">Cari Dokumen</label>
            <SearchInput v-model="searchQuery" placeholder="Judul atau deskripsi..." size="sm" @clear="resetFilters" />
          </div>
          <div>
            <label class="block text-[11px] font-medium text-gray-700 mb-1">Rentang Tanggal</label>
            <select v-model="dateFilter" class="w-full px-2.5 py-1.5 text-xs border border-gray-200 rounded-lg focus:ring-2 focus:ring-emerald-500 focus:border-emerald-500 bg-gray-50">
              <option value="all">Semua Tanggal</option>
              <option value="last30days">30 Hari Terakhir</option>
              <option value="last90days">90 Hari Terakhir</option>
              <option value="thisYear">Tahun Ini</option>
            </select>
          </div>
          <div>
            <label class="block text-[11px] font-medium text-gray-700 mb-1">Kategori</label>
            <select v-model="selectedCategory" class="w-full px-2.5 py-1.5 text-xs border border-gray-200 rounded-lg focus:ring-2 focus:ring-emerald-500 focus:border-emerald-500 bg-gray-50">
              <option v-for="category in categories" :key="category" :value="category">{{ category }}</option>
            </select>
          </div>
          <div class="flex items-center justify-between pt-1">
            <button v-if="hasActiveFilter" @click="resetFilters" class="text-[11px] text-emerald-600 hover:text-emerald-700 font-medium">Reset Filter</button>
            <button @click="showFilters = false" class="ml-auto px-4 py-1.5 text-xs font-medium text-white bg-emerald-600 rounded-lg hover:bg-emerald-700">Terapkan</button>
          </div>
        </div>
      </div>
    </Transition>

    <!-- Request dialog -->
    <div v-if="showRequestDialog" class="fixed inset-0 bg-black/50 flex items-center justify-center z-[700] p-4 animate-fade-in" @click.self="closeRequestDialog">
      <div class="bg-white rounded-lg shadow-xl max-w-md w-full p-5 max-h-[90vh] overflow-y-auto">
        <div class="flex items-center justify-between mb-3">
          <h3 class="text-base font-bold text-gray-900">Minta Dokumen</h3>
          <button @click="closeRequestDialog" :disabled="isSubmitting" class="text-gray-400 hover:text-gray-600 disabled:opacity-50 p-1 hover:bg-gray-100 rounded-full transition-colors" aria-label="Close">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" /></svg>
          </button>
        </div>

        <div class="mb-4 p-3 bg-gray-50 rounded-lg border border-gray-200">
          <p class="text-[11px] text-gray-500 mb-0.5">Anda meminta:</p>
          <h4 class="font-semibold text-gray-900 text-xs">{{ selectedPPID.title }}</h4>
          <p class="text-[11px] text-gray-600 mt-0.5">{{ selectedPPID.category }}</p>
        </div>

        <form @submit.prevent="submitRequest" class="space-y-3">
          <div>
            <label for="requester-name" class="block text-xs font-medium text-gray-700 mb-1">Nama Lengkap <span class="text-red-500">*</span></label>
            <input id="requester-name" v-model="requesterName" type="text" placeholder="Nama lengkap Anda" :disabled="isSubmitting" required class="w-full px-3 py-1.5 text-sm border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500 focus:border-emerald-500 disabled:opacity-50 transition-colors" />
          </div>

          <div>
            <label for="requester-email" class="block text-xs font-medium text-gray-700 mb-1">Alamat Email <span class="text-red-500">*</span></label>
            <input id="requester-email" v-model="requesterEmail" type="email" placeholder="email@contoh.com" :disabled="isSubmitting" required class="w-full px-3 py-1.5 text-sm border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500 focus:border-emerald-500 disabled:opacity-50 transition-colors" />
            <p class="text-[11px] text-gray-500 mt-1">Email konfirmasi akan dikirim ke alamat ini.</p>
          </div>

          <div>
            <label for="request-purpose" class="block text-xs font-medium text-gray-700 mb-1">Tujuan Penggunaan <span class="text-gray-400 font-normal">(opsional)</span></label>
            <textarea id="request-purpose" v-model="requestPurpose" placeholder="Jelaskan tujuan Anda meminta dokumen ini" :disabled="isSubmitting" rows="3" class="w-full px-3 py-1.5 text-sm border border-gray-300 rounded-lg focus:ring-2 focus:ring-emerald-500 focus:border-emerald-500 disabled:opacity-50 resize-none transition-colors"></textarea>
          </div>

          <div v-if="requestError" class="bg-red-50 border border-red-200 text-red-700 px-3 py-2 rounded-lg text-xs">
            {{ requestError }}
          </div>

          <div class="flex gap-2.5 pt-1">
            <button type="button" @click="closeRequestDialog" :disabled="isSubmitting" class="flex-1 px-3 py-2 text-sm border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 transition-colors disabled:opacity-50 font-medium">Batal</button>
            <button type="submit" :disabled="isSubmitting" class="flex-1 px-3 py-2 text-sm bg-emerald-600 text-white rounded-lg hover:bg-emerald-700 transition-colors disabled:opacity-50 font-medium inline-flex items-center justify-center gap-1.5">
              <svg v-if="isSubmitting" class="w-4 h-4 animate-spin" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" /></svg>
              {{ isSubmitting ? 'Mengirim...' : 'Kirim Permintaan' }}
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- Success dialog -->
    <div v-if="showSuccessDialog" class="fixed inset-0 bg-black/50 flex items-center justify-center z-[700] p-4 animate-fade-in" @click.self="closeSuccessDialog">
      <div class="bg-white rounded-lg shadow-xl max-w-md w-full p-5 text-center">
        <div class="w-12 h-12 bg-emerald-100 rounded-full flex items-center justify-center mx-auto mb-3">
          <svg class="w-6 h-6 text-emerald-600" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" /></svg>
        </div>
        <h3 class="text-base font-bold text-gray-900 mb-1.5">Permintaan Berhasil!</h3>
        <p class="text-gray-600 mb-4 text-xs">
          Kami akan mengirimkan email konfirmasi ke <span class="font-semibold text-gray-900">{{ requesterEmail }}</span> dalam beberapa menit.
        </p>
        <button @click="closeSuccessDialog" class="w-full px-3 py-2 text-sm bg-emerald-600 text-white rounded-lg hover:bg-emerald-700 transition-colors font-medium">Tutup</button>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, onMounted, computed, watch, nextTick } from 'vue'
import { getPublicPPIDList, getPublicPPIDById, createPPIDRequest, thumbnailUrl } from '../services/desaService'
import LoadingSpinner from '../components/common/LoadingSpinner.vue'
import EmptyState from '../components/common/EmptyState.vue'
import SearchInput from '../components/common/SearchInput.vue'
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
    const showFilters = ref(false)
    const scrollRef = ref(null)
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

    // Full-size document image shown at the top of the detail view.
    const detailThumb = computed(() => selectedPPID.value ? thumbnailUrl(selectedPPID.value) : '')

    const scrollToTop = () => {
      nextTick(() => {
        if (scrollRef.value) scrollRef.value.scrollTop = 0
      })
    }

    const viewDetail = (doc) => {
      selectedPPID.value = doc
      scrollToTop()
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

    watch(currentPage, () => {
      scrollToTop()
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
      showFilters,
      scrollRef,
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
      stripHtml,
      sanitizedDescription,
      detailThumb,
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

<style>
.slide-up-enter-active,
.slide-up-leave-active {
  transition: transform 0.3s ease-in-out;
}
.slide-up-enter-from,
.slide-up-leave-to {
  transform: translateY(100%);
}
</style>
