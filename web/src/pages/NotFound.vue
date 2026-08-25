<template>
  <div class="min-h-[60vh] flex items-center justify-center px-4 py-16">
    <div class="text-center max-w-md">
      <div class="inline-flex items-center justify-center w-20 h-20 rounded-full bg-emerald-50 mb-6">
        <svg class="w-10 h-10 text-emerald-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9.879 16.121A3 3 0 1012.015 11L11 14H9c0 .768.293 1.536.879 2.121z" />
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M3.124 7.5A8.969 8.969 0 015.292 3m13.416 0a8.969 8.969 0 012.168 4.5" />
        </svg>
      </div>
      <h1 class="text-6xl font-bold text-gray-900 mb-2">404</h1>
      <p class="text-xl text-gray-700 mb-2">Halaman tidak ditemukan</p>
      <p class="text-sm text-gray-500 mb-8">
        Halaman yang Anda cari mungkin telah dipindahkan atau tidak tersedia.
      </p>
      <RouterLink
        to="/"
        class="inline-flex items-center px-6 py-3 bg-emerald-600 text-white font-medium rounded-lg hover:bg-emerald-700 transition-colors"
      >
        <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6" />
        </svg>
        Kembali ke Beranda
      </RouterLink>
    </div>
  </div>
</template>

<script>
export default {
  name: 'NotFound'
}
</script>