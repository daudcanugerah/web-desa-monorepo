<template>
  <div class="min-h-screen bg-gray-50">
    <!-- Pengumuman Ticker -->
    <section class="bg-gradient-to-r from-red-50 to-amber-50 border-b border-red-100 overflow-hidden">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div class="flex items-stretch">
          <div class="bg-red-600 text-white px-4 py-2.5 flex items-center gap-2 flex-shrink-0">
            <Icon name="speaker" class="w-4 h-4" />
            <span class="text-xs font-bold uppercase tracking-wider">Pengumuman</span>
          </div>
          <div class="flex-1 overflow-hidden relative">
            <div class="flex items-center h-full px-4">
              <div class="ticker-track flex items-center gap-12 whitespace-nowrap">
                <a
                  v-for="(p, idx) in [...pengumuman, ...pengumuman]"
                  :key="idx"
                  href="#"
                  class="text-sm text-gray-800 inline-flex items-center gap-2 hover:text-red-700 transition-colors"
                >
                  <span class="w-1.5 h-1.5 bg-red-600 rounded-full flex-shrink-0"></span>
                  <span class="font-medium">{{ p.title }}</span>
                  <span class="text-gray-500">— {{ p.date }}</span>
                </a>
              </div>
            </div>
          </div>
          <button class="hidden sm:flex items-center gap-1 px-3 text-xs text-red-700 hover:text-red-800 font-medium flex-shrink-0 border-l border-red-100">
            Semua
            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" /></svg>
          </button>
        </div>
      </div>
    </section>

    <!-- Hero Banner Slider -->
    <section class="relative h-[560px] overflow-hidden bg-gray-900">
      <template v-for="(slide, index) in slides" :key="index">
        <div
          :class="[
            'absolute inset-0 transition-opacity duration-1000',
            index === currentSlide ? 'opacity-100' : 'opacity-0'
          ]"
          aria-hidden="true"
        >
          <img :src="slide.image" :alt="slide.title" loading="eager" class="absolute inset-0 w-full h-full object-cover" />
          <div :class="`absolute inset-0 bg-gradient-to-br ${slide.gradient} opacity-70`"></div>
          <div class="absolute inset-0 bg-black opacity-30"></div>
        </div>

        <div
          :class="[
            'absolute inset-0 flex items-center transition-opacity duration-1000 pointer-events-none',
            index === currentSlide ? 'opacity-100' : 'opacity-0'
          ]"
          aria-hidden="true"
        >
          <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 w-full">
            <div class="text-center text-white pointer-events-auto">
              <div class="inline-flex items-center gap-2 bg-white/10 backdrop-blur-sm border border-white/20 px-4 py-1.5 rounded-full mb-6 animate-fade-in">
                <Icon name="garuda" class="w-4 h-4" />
                <span class="text-white text-xs font-medium tracking-wide">PEMERINTAH DESA</span>
              </div>
              <h1 class="text-4xl md:text-6xl font-bold mb-4 leading-tight animate-slide-up">
                {{ slide.title }}
              </h1>
              <p class="text-lg md:text-xl text-white/90 mb-3 max-w-3xl mx-auto animate-slide-up-delay">
                {{ slide.subtitle }}
              </p>
              <p class="text-sm text-emerald-100/90 mb-8 inline-flex items-center gap-1.5 animate-fade-in-delay">
                <Icon name="pin" class="w-3.5 h-3.5" />
                {{ wilayah }}
              </p>
              <div class="flex flex-col sm:flex-row gap-3 justify-center animate-fade-in-delay">
                <RouterLink to="/profil" class="inline-flex items-center justify-center px-7 py-3 bg-white text-emerald-700 font-semibold rounded-lg hover:bg-emerald-50 transition-all shadow-lg">
                  <Icon name="building" class="w-4 h-4 mr-2" />
                  Profil Desa
                </RouterLink>
                <RouterLink to="/ppid" class="inline-flex items-center justify-center px-7 py-3 bg-emerald-700/80 backdrop-blur-sm text-white font-semibold rounded-lg hover:bg-emerald-700 transition-all border border-white/20">
                  <Icon name="scroll" class="w-4 h-4 mr-2" />
                  Layanan Publik
                </RouterLink>
                <RouterLink to="/berita" class="inline-flex items-center justify-center px-7 py-3 bg-white/10 backdrop-blur-sm text-white font-semibold rounded-lg hover:bg-white/20 transition-all border border-white/30">
                  <Icon name="speaker" class="w-4 h-4 mr-2" />
                  Berita
                </RouterLink>
              </div>
            </div>
          </div>
        </div>
      </template>

      <div class="absolute bottom-6 left-1/2 -translate-x-1/2 flex gap-2 z-10">
        <button
          v-for="(_, index) in slides"
          :key="index"
          @click="currentSlide = index"
          :class="[
            'h-2 transition-all rounded-full',
            index === currentSlide ? 'w-10 bg-white' : 'w-2 bg-white/40 hover:bg-white/70'
          ]"
          :aria-label="`Slide ${index + 1}`"
        />
      </div>

      <button @click="previousSlide" class="absolute left-4 top-1/2 -translate-y-1/2 w-11 h-11 bg-white/10 backdrop-blur-sm rounded-full flex items-center justify-center text-white hover:bg-white/20 transition-all z-10" aria-label="Slide sebelumnya">
        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" /></svg>
      </button>
      <button @click="nextSlide" class="absolute right-4 top-1/2 -translate-y-1/2 w-11 h-11 bg-white/10 backdrop-blur-sm rounded-full flex items-center justify-center text-white hover:bg-white/20 transition-all z-10" aria-label="Slide berikutnya">
        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" /></svg>
      </button>

      <div class="absolute bottom-0 left-0 right-0 h-1 bg-gradient-to-r from-red-600 via-white to-red-600 z-10"></div>
    </section>

    <!-- Motto Ribbon -->
    <section class="bg-emerald-700 text-white py-3">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 flex items-center justify-center gap-3 text-center">
        <Icon name="flag" class="w-4 h-4 text-yellow-300 flex-shrink-0" />
        <p class="text-sm font-medium tracking-wide italic">"{{ motto }}"</p>
        <Icon name="flag" class="w-4 h-4 text-yellow-300 flex-shrink-0" />
      </div>
    </section>

    <!-- Sambutan Kepala Desa -->
    <section class="py-16 bg-white">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div class="grid md:grid-cols-12 gap-8 items-center">
          <div class="md:col-span-4">
            <div class="relative inline-block">
              <div class="absolute -inset-1 bg-gradient-to-br from-emerald-600 to-red-600 rounded-2xl blur opacity-30"></div>
              <div class="relative w-48 h-48 md:w-full md:h-auto md:aspect-square bg-gradient-to-br from-emerald-100 to-emerald-200 rounded-2xl flex items-center justify-center overflow-hidden shadow-lg">
                <div class="text-center">
                  <Icon name="user" class="w-20 h-20 text-emerald-600 mx-auto mb-2" />
                  <p class="text-emerald-800 font-bold text-lg">{{ kepalaDesa.initial }}</p>
                </div>
              </div>
            </div>
          </div>

          <div class="md:col-span-8">
            <div class="inline-flex items-center gap-2 bg-red-50 text-red-700 px-3 py-1 rounded-full text-xs font-medium mb-3 border border-red-100">
              <Icon name="award" class="w-3.5 h-3.5" />
              <span>Sambutan Kepala Desa</span>
            </div>
            <h2 class="text-3xl md:text-4xl font-bold text-gray-900 mb-2 leading-tight">{{ kepalaDesa.name }}</h2>
            <p class="text-emerald-700 font-medium mb-5">{{ kepalaDesa.position }}</p>
            <p class="text-gray-700 leading-relaxed mb-4 whitespace-pre-line">{{ kepalaDesa.message }}</p>
            <RouterLink to="/profil" class="inline-flex items-center text-emerald-700 hover:text-emerald-800 font-medium text-sm">
              Lihat Sambutan Lengkap
              <svg class="w-4 h-4 ml-1" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" /></svg>
            </RouterLink>
          </div>
        </div>
      </div>
    </section>

    <!-- Statistik Penduduk -->
    <section class="py-16 bg-gradient-to-br from-emerald-700 via-emerald-800 to-teal-900 text-white relative overflow-hidden">
      <div class="absolute inset-0 opacity-10">
        <svg class="w-full h-full" viewBox="0 0 100 100" preserveAspectRatio="none">
          <pattern id="dots" x="0" y="0" width="10" height="10" patternUnits="userSpaceOnUse">
            <circle cx="2" cy="2" r="1" fill="currentColor" />
          </pattern>
          <rect width="100" height="100" fill="url(#dots)" />
        </svg>
      </div>
      <div class="relative max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div class="text-center mb-12">
          <div class="inline-flex items-center gap-2 bg-white/10 backdrop-blur-sm border border-white/20 px-3 py-1 rounded-full text-xs font-medium mb-3">
            <Icon name="chart" class="w-3.5 h-3.5" />
            <span>Data Statistik</span>
          </div>
          <h2 class="text-3xl md:text-4xl font-bold mb-2">Data Desa {{ desaInfo.name || 'Sukamaju' }}</h2>
          <p class="text-emerald-100/90 text-sm">Data agregat kependudukan berdasarkan registrasi terkini</p>
        </div>

        <div class="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-4">
          <div v-for="stat in statistik" :key="stat.label" class="bg-white/10 backdrop-blur-sm border border-white/15 rounded-xl p-5 hover:bg-white/15 transition-colors group">
            <div class="w-11 h-11 rounded-lg bg-white/10 flex items-center justify-center mb-3 group-hover:scale-110 transition-transform">
              <Icon :name="stat.icon" class="w-5 h-5 text-yellow-300" />
            </div>
            <p class="text-2xl md:text-3xl font-bold mb-0.5 tabular-nums">
              <CountUp :value="stat.value" :separator="'.'" />
            </p>
            <p class="text-sm text-emerald-100">{{ stat.label }}</p>
          </div>
        </div>
      </div>
    </section>

    <!-- Berita + Sidebar -->
    <section class="py-16 bg-white">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div class="grid md:grid-cols-3 gap-8">
          <div class="md:col-span-2">
            <div class="flex items-end justify-between mb-6">
              <div>
                <h2 class="text-2xl font-bold text-gray-900 mb-1">Berita & Informasi</h2>
                <p class="text-sm text-gray-600">Kabar terbaru dari {{ desaInfo.name || 'desa kami' }}</p>
              </div>
              <RouterLink to="/berita" class="hidden md:inline-flex items-center text-emerald-700 hover:text-emerald-800 font-medium text-sm">
                Semua Berita
                <svg class="w-4 h-4 ml-1" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" /></svg>
              </RouterLink>
            </div>

            <LoadingSpinner v-if="newsLoading" />
            <div v-else-if="latestNews && latestNews.length > 0" class="space-y-4">
              <RouterLink
                v-for="(news, idx) in latestNews.slice(0, 4)"
                :key="news.id"
                :to="`/berita/${news.id}`"
                class="group block"
              >
                <article class="flex gap-4 bg-white rounded-xl overflow-hidden border border-gray-200 hover:border-emerald-200 hover:shadow-md transition-all">
                  <div :class="['relative overflow-hidden bg-gradient-to-br from-emerald-100 to-teal-100 flex-shrink-0', idx === 0 ? 'md:w-64 h-40 md:h-auto' : 'w-28 h-28 md:w-36 md:h-28']">
                    <img
                      v-if="newsMediaUrl(news)"
                      :src="newsMediaUrl(news)"
                      :alt="news.title"
                      loading="lazy"
                      class="absolute inset-0 w-full h-full object-cover transition-transform duration-500 group-hover:scale-105"
                      @error="handleImageError"
                    />
                  </div>
                  <div class="flex-1 p-4 min-w-0">
                    <div class="flex items-center gap-2 mb-2 flex-wrap">
                      <span class="inline-block bg-emerald-100 text-emerald-700 text-xs px-2.5 py-0.5 rounded-full font-medium">
                        {{ news.category || 'Umum' }}
                      </span>
                      <span class="text-xs text-gray-500">{{ formatDate(news.created_at) }}</span>
                    </div>
                    <h3 :class="['font-bold text-gray-900 mb-2 group-hover:text-emerald-700 transition-colors line-clamp-2', idx === 0 ? 'text-lg' : 'text-base']">
                      {{ news.title }}
                    </h3>
                    <span class="text-emerald-700 text-xs font-medium inline-flex items-center group-hover:gap-2 transition-all">
                      Baca Selengkapnya
                      <svg class="w-3.5 h-3.5 ml-1" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" /></svg>
                    </span>
                  </div>
                </article>
              </RouterLink>
            </div>
            <EmptyState v-else message="Belum ada berita tersedia" />
          </div>

          <aside class="md:col-span-1">
            <RouterLink
              to="/infografik"
              class="block bg-emerald-50 rounded-xl p-5 border border-emerald-100 hover:border-emerald-300 hover:bg-emerald-100/50 hover:shadow-md transition-all group"
            >
              <div class="flex items-center justify-between mb-3">
                <div class="flex items-center gap-2">
                  <Icon name="wallet" class="w-4 h-4 text-emerald-700" />
                  <h3 class="text-sm font-bold text-gray-900 uppercase tracking-wide">APBDesa {{ currentYear }}</h3>
                </div>
                <span class="inline-flex items-center gap-1 text-xs text-emerald-700 font-medium group-hover:gap-1.5 transition-all">
                  Detail
                  <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" /></svg>
                </span>
              </div>
              <div class="space-y-2 text-sm">
                <div class="flex justify-between">
                  <span class="text-gray-600">Pendapatan</span>
                  <span class="font-semibold text-gray-900">Rp 2,4 M</span>
                </div>
                <div class="flex justify-between">
                  <span class="text-gray-600">Belanja</span>
                  <span class="font-semibold text-gray-900">Rp 2,3 M</span>
                </div>
                <div class="mt-3 h-2 bg-emerald-100 rounded-full overflow-hidden">
                  <div class="h-full bg-emerald-600 rounded-full" style="width: 96%"></div>
                </div>
                <p class="text-xs text-gray-500 text-center mt-1">96% terealisasi · lihat detail di Infografis</p>
              </div>
            </RouterLink>
          </aside>
        </div>
      </div>
    </section>

    <!-- Mini Peta Preview -->
    <section class="py-16 bg-gray-50">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div class="text-center mb-12">
          <div class="inline-flex items-center gap-2 bg-emerald-100 text-emerald-700 px-3 py-1 rounded-full text-xs font-medium mb-3">
            <Icon name="pin" class="w-3.5 h-3.5" />
            <span>Peta Wilayah</span>
          </div>
          <h2 class="text-3xl font-bold text-gray-900 mb-2">Wilayah {{ desaInfo.name || 'Desa' }}</h2>
          <p class="text-gray-600 text-sm">Batas administrasi dan fasilitas umum di {{ desaInfo.name || 'desa kami' }}</p>
        </div>

        <div class="grid md:grid-cols-12 gap-6">
          <div class="md:col-span-8">
            <RouterLink to="/peta" class="block relative group overflow-hidden rounded-xl border border-gray-200 bg-white shadow-sm hover:shadow-lg transition-all">
              <div class="aspect-[16/9] relative bg-emerald-50">
                <svg viewBox="0 0 100 70" class="w-full h-full">
                  <rect width="100" height="70" fill="#ECFDF5" />
                  <g stroke="#A7F3D0" stroke-width="0.15" fill="none">
                    <line x1="0" y1="10" x2="100" y2="10" />
                    <line x1="0" y1="20" x2="100" y2="20" />
                    <line x1="0" y1="30" x2="100" y2="30" />
                    <line x1="0" y1="40" x2="100" y2="40" />
                    <line x1="0" y1="50" x2="100" y2="50" />
                    <line x1="0" y1="60" x2="100" y2="60" />
                    <line x1="10" y1="0" x2="10" y2="70" />
                    <line x1="20" y1="0" x2="20" y2="70" />
                    <line x1="30" y1="0" x2="30" y2="70" />
                    <line x1="40" y1="0" x2="40" y2="70" />
                    <line x1="50" y1="0" x2="50" y2="70" />
                    <line x1="60" y1="0" x2="60" y2="70" />
                    <line x1="70" y1="0" x2="70" y2="70" />
                    <line x1="80" y1="0" x2="80" y2="70" />
                    <line x1="90" y1="0" x2="90" y2="70" />
                  </g>
                  <polygon points="15,20 25,15 35,18 50,22 65,18 75,25 85,30 82,42 70,50 55,52 40,48 25,42 18,32" fill="#10B981" fill-opacity="0.2" stroke="#059669" stroke-width="0.5" />
                  <circle cx="25" cy="28" r="1" fill="#10B981" />
                  <circle cx="42" cy="35" r="1" fill="#3B82F6" />
                  <circle cx="58" cy="32" r="1" fill="#EF4444" />
                  <circle cx="70" cy="40" r="1" fill="#8B5CF6" />
                  <circle cx="48" cy="22" r="1" fill="#F59E0B" />
                  <text x="50" y="68" text-anchor="middle" font-size="3" fill="#047857" font-weight="bold">{{ desaInfo.name || 'Desa Sukamaju' }}</text>
                </svg>
                <div class="absolute inset-0 bg-emerald-900/0 group-hover:bg-emerald-900/10 transition-colors flex items-center justify-center">
                  <span class="opacity-0 group-hover:opacity-100 transition-opacity bg-white px-5 py-2 rounded-full text-sm font-medium text-emerald-700 shadow-lg">
                    Buka Peta Lengkap →
                  </span>
                </div>
              </div>
            </RouterLink>
          </div>

          <div class="md:col-span-4 space-y-3">
            <div class="bg-white rounded-xl p-4 border border-gray-200">
              <p class="text-xs text-gray-500 mb-1">Luas Wilayah</p>
              <p class="text-2xl font-bold text-gray-900 tabular-nums">12,5 <span class="text-sm font-normal text-gray-500">km²</span></p>
            </div>
            <div class="bg-white rounded-xl p-4 border border-gray-200">
              <p class="text-xs text-gray-500 mb-1">Jumlah Dusun</p>
              <p class="text-2xl font-bold text-gray-900 tabular-nums">8 <span class="text-sm font-normal text-gray-500">dusun</span></p>
            </div>
            <div class="bg-white rounded-xl p-4 border border-gray-200">
              <p class="text-xs text-gray-500 mb-1">Koordinat Pusat</p>
              <p class="text-sm font-mono text-gray-900">-6.7128°, 107.6925°</p>
            </div>
            <RouterLink to="/peta" class="block bg-emerald-700 hover:bg-emerald-800 text-white text-center font-medium py-3 rounded-xl transition-colors">
              Jelajahi Peta Interaktif
            </RouterLink>
          </div>
        </div>
      </div>
    </section>

    <!-- Jelajahi Desa -->
    <section class="py-16 bg-white">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div class="text-center mb-12">
          <div class="inline-flex items-center gap-2 bg-emerald-100 text-emerald-700 px-3 py-1 rounded-full text-xs font-medium mb-3">
            <Icon name="grid" class="w-3.5 h-3.5" />
            <span>Jelajahi</span>
          </div>
          <h2 class="text-3xl font-bold text-gray-900 mb-2">Jelajahi {{ desaInfo.name || 'Desa' }}</h2>
          <p class="text-gray-600 text-sm">Informasi lengkap tentang desa dalam satu tempat</p>
        </div>

        <div class="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-4">
          <RouterLink v-for="item in exploreItems" :key="item.link" :to="item.link" class="group relative bg-white rounded-xl p-5 border-2 border-gray-100 hover:border-transparent hover:shadow-lg transition-all overflow-hidden text-center">
            <div :class="`absolute inset-0 bg-gradient-to-br ${item.gradient} opacity-0 group-hover:opacity-5 transition-opacity`"></div>
            <div class="relative">
              <div :class="`inline-flex items-center justify-center w-12 h-12 rounded-lg bg-gradient-to-br ${item.gradient} mb-3 group-hover:scale-110 transition-transform shadow-md`">
                <Icon :name="item.icon" class="w-5 h-5 text-white" />
              </div>
              <p class="text-sm font-bold text-gray-900 mb-1">{{ item.title }}</p>
              <p class="text-xs text-gray-500 leading-tight line-clamp-2">{{ item.desc }}</p>
            </div>
          </RouterLink>
        </div>
      </div>
    </section>

    <!-- Contact CTA -->
    <section class="py-16 bg-gradient-to-br from-emerald-700 via-emerald-800 to-teal-900 text-white relative overflow-hidden">
      <div class="absolute top-0 left-0 right-0 h-1 bg-gradient-to-r from-red-600 via-white to-red-600"></div>
      <div class="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 text-center relative">
        <Icon name="shield" class="w-12 h-12 mx-auto mb-4 text-yellow-300" />
        <h2 class="text-3xl md:text-4xl font-bold mb-3">Layanan Aspirasi & Pengaduan</h2>
        <p class="text-emerald-100 mb-8 max-w-2xl mx-auto">
          Sampaikan aspirasi, saran, dan pengaduan Anda untuk kemajuan {{ desaInfo.name || 'desa kami' }}. Kami siap melayani dengan transparan.
        </p>
        <div class="flex flex-col sm:flex-row gap-3 justify-center">
          <a v-if="desaInfo.phone" :href="`tel:${desaInfo.phone}`" class="inline-flex items-center justify-center px-7 py-3 bg-white text-emerald-700 font-semibold rounded-lg hover:bg-emerald-50 transition-colors shadow-lg">
            <Icon name="phone" class="w-4 h-4 mr-2" />
            Hubungi {{ desaInfo.phone }}
          </a>
          <a v-if="desaInfo.email" :href="`mailto:${desaInfo.email}`" class="inline-flex items-center justify-center px-7 py-3 bg-emerald-900/50 backdrop-blur-sm text-white font-semibold rounded-lg hover:bg-emerald-900/70 transition-colors border border-white/20">
            <Icon name="mail" class="w-4 h-4 mr-2" />
            Kirim Email
          </a>
          <RouterLink to="/ppid" class="inline-flex items-center justify-center px-7 py-3 bg-emerald-900/50 backdrop-blur-sm text-white font-semibold rounded-lg hover:bg-emerald-900/70 transition-colors border border-white/20">
            <Icon name="scroll" class="w-4 h-4 mr-2" />
            Layanan Publik
          </RouterLink>
        </div>
      </div>
    </section>
  </div>
</template>

<script>
import { ref, onMounted, onUnmounted, watch, computed } from 'vue'
import { getPublicBeritaList, getActiveBanners, mediaUrl } from '../services/desaService'
import LoadingSpinner from '../components/common/LoadingSpinner.vue'
import EmptyState from '../components/common/EmptyState.vue'
import Icon from '../components/common/Icon.vue'
import CountUp from '../components/common/CountUp.vue'
import { useDesaInfo } from '../composables/useDesaInfo'

const FAKE_KepalaDesa = Object.freeze({
  name: 'H. Budi Santoso, S.Sos',
  position: 'Kepala Desa Sukamaju · Periode 2024–2030',
  initial: 'BS',
  message: `Assalamualaikum Warahmatullahi Wabarakatuh.

Puji syukur kehadirat Allah SWT atas segala rahmat dan karunia-Nya. Selamat datang di Website Resmi Pemerintah Desa Sukamaju. Website ini hadir sebagai wujud transparansi dan akuntabilitas pemerintahan desa dalam memberikan informasi dan layanan kepada masyarakat.

Kami berkomitmen untuk membangun desa yang maju, sejahtera, dan berbudaya dengan melibatkan seluruh elemen masyarakat. Semoga melalui media ini, hubungan antara pemerintah desa dan masyarakat semakin erat, dan pelayanan publik dapat berlangsung secara terbuka, cepat, dan tepat sasaran.

Mari bersama-sama wujudkan Desa Sukamaju yang lebih baik untuk generasi mendatang.`
})

const FAKE_MOTTO = 'Bersama Membangun Desa yang Maju, Sejahtera, dan Berbudaya'

const FAKE_STATISTIK = Object.freeze([
  { label: 'Penduduk', value: 5420, icon: 'users' },
  { label: 'Kartu Keluarga', value: 1450, icon: 'family' },
  { label: 'Dusun', value: 8, icon: 'home' },
  { label: 'RT', value: 32, icon: 'pin' },
  { label: 'RW', value: 8, icon: 'pin' },
  { label: 'UMKM', value: 45, icon: 'store' }
])

const FAKE_PENGUMUMAN = Object.freeze([
  { title: 'Pendaftaran BLT-DD Tahun Anggaran 2026 Tahap I', date: '12 Juli 2026' },
  { title: 'Jadwal Posyandu Balita Bulan Agustus 2026', date: '10 Juli 2026' },
  { title: 'Musyawarah Desa Penetapan APBDesa Perubahan 2026', date: '5 Juli 2026' },
  { title: 'Pemberitahuan Pelayanan Kantor Desa Selama Libur Nasional', date: '1 Juli 2026' },
  { title: 'Rekrutmen Perangkat Desa Formasi Kaur Keuangan 2026', date: '28 Juni 2026' }
])

export default {
  name: 'Home',
  components: {
    LoadingSpinner,
    EmptyState,
    Icon,
    CountUp
  },
  setup() {
    const latestNews = ref([])
    const { desaInfo: sharedDesaInfo, loading: desaLoading } = useDesaInfo()
    const newsLoading = ref(true)
    const currentSlide = ref(0)
    let slideTimer = null

    const currentYear = new Date().getFullYear()

    const kepalaDesa = computed(() => ({
      ...FAKE_KepalaDesa,
      name: sharedDesaInfo.value?.kepala_desa || FAKE_KepalaDesa.name,
      message: sharedDesaInfo.value?.kepala_desa_message || FAKE_KepalaDesa.message
    }))

    const motto = computed(() => sharedDesaInfo.value?.motto || FAKE_MOTTO)

    const wilayah = computed(() => {
      const kec = sharedDesaInfo.value?.kecamatan || 'Sukamaju'
      const kab = sharedDesaInfo.value?.kabupaten || 'Cianjur'
      const prov = sharedDesaInfo.value?.provinsi || 'Jawa Barat'
      return `Kec. ${kec}, Kab. ${kab}, Prov. ${prov}`
    })

    const statistik = computed(() => {
      const d = sharedDesaInfo.value || {}
      return [
        { label: 'Penduduk', value: d.jumlah_penduduk ?? FAKE_STATISTIK[0].value, icon: 'users' },
        { label: 'Kartu Keluarga', value: d.jumlah_kk ?? FAKE_STATISTIK[1].value, icon: 'family' },
        { label: 'Dusun', value: d.jumlah_dusun ?? FAKE_STATISTIK[2].value, icon: 'home' },
        { label: 'RT', value: d.jumlah_rt ?? FAKE_STATISTIK[3].value, icon: 'pin' },
        { label: 'RW', value: d.jumlah_rw ?? FAKE_STATISTIK[4].value, icon: 'pin' },
        { label: 'UMKM', value: d.jumlah_umkm ?? FAKE_STATISTIK[5].value, icon: 'store' }
      ]
    })

    const pengumuman = FAKE_PENGUMUMAN

    const defaultSlides = [
      {
        title: 'Selamat Datang',
        subtitle: 'Desa yang Maju, Sejahtera, dan Berbudaya',
        gradient: 'from-emerald-700 via-emerald-800 to-teal-900',
        image: 'https://picsum.photos/seed/hero1/1920/1080'
      },
      {
        title: 'Transparansi & Akuntabilitas',
        subtitle: 'Pemerintahan yang Terbuka untuk Masyarakat',
        gradient: 'from-blue-700 via-blue-800 to-cyan-900',
        image: 'https://picsum.photos/seed/hero2/1920/1080'
      },
      {
        title: 'Pembangunan Berkelanjutan',
        subtitle: 'Bersama Membangun Desa yang Lebih Baik',
        gradient: 'from-teal-700 via-emerald-800 to-green-900',
        image: 'https://picsum.photos/seed/hero3/1920/1080'
      }
    ]

    const slideGradients = [
      'from-emerald-700 via-emerald-800 to-teal-900',
      'from-blue-700 via-blue-800 to-cyan-900',
      'from-teal-700 via-emerald-800 to-green-900',
      'from-purple-700 via-indigo-800 to-blue-900',
      'from-amber-700 via-orange-800 to-red-900'
    ]

    const slides = ref(defaultSlides)

    watch(sharedDesaInfo, (val) => {
      if (val?.name && slides.value[0]) {
        slides.value = slides.value.map((s, i) =>
          i === 0 ? { ...s, title: `Selamat Datang di ${val.name}` } : s
        )
      }
    })

    const exploreItems = [
      { icon: 'building', title: 'Profil', desc: 'Sejarah & visi misi', link: '/profil', gradient: 'from-emerald-500 to-teal-600' },
      { icon: 'chart', title: 'Infografik', desc: 'Data statistik', link: '/infografik', gradient: 'from-blue-500 to-cyan-600' },
      { icon: 'pin', title: 'Peta', desc: 'Lokasi fasilitas', link: '/peta', gradient: 'from-emerald-600 to-green-700' },
      { icon: 'speaker', title: 'Berita', desc: 'Kabar terbaru', link: '/berita', gradient: 'from-amber-500 to-orange-600' },
      { icon: 'store', title: 'UMKM', desc: 'Usaha lokal', link: '/umkm', gradient: 'from-purple-500 to-pink-600' },
      { icon: 'scroll', title: 'PPID', desc: 'Info publik', link: '/ppid', gradient: 'from-rose-500 to-red-600' }
    ]

    const newsMediaUrl = (item) => mediaUrl(item, 'single')

    const formatDate = (dateString) => {
      if (!dateString) return '-'
      try {
        return new Date(dateString).toLocaleDateString('id-ID', {
          day: 'numeric', month: 'short', year: 'numeric'
        })
      } catch { return '-' }
    }

    const handleImageError = (event) => { event.target.style.display = 'none' }

    const nextSlide = () => { currentSlide.value = (currentSlide.value + 1) % slides.value.length }
    const previousSlide = () => { currentSlide.value = (currentSlide.value - 1 + slides.value.length) % slides.value.length }
    const startSlideTimer = () => { slideTimer = setInterval(nextSlide, 5000) }

    onMounted(async () => {
      const bannersPromise = getActiveBanners().catch(() => [])
      const [bannersResult, beritaResult] = await Promise.all([
        bannersPromise,
        getPublicBeritaList().catch(() => [])
      ])

      if (Array.isArray(bannersResult) && bannersResult.length > 0) {
        slides.value = bannersResult.slice(0, 5).map((b, i) => ({
          title: b.title || defaultSlides[0].title,
          subtitle: b.description || '',
          gradient: slideGradients[i % slideGradients.length],
          link: b.link || null,
          image: mediaUrl(b, 'single') || `https://picsum.photos/seed/hero${i + 1}/1920/1080`
        }))
      }

      latestNews.value = Array.isArray(beritaResult) ? beritaResult.slice(0, 4) : []
      newsLoading.value = false
      startSlideTimer()
    })

    onUnmounted(() => { if (slideTimer) clearInterval(slideTimer) })

    return {
      latestNews,
      desaInfo: sharedDesaInfo,
      desaLoading,
      newsLoading,
      currentSlide,
      slides,
      exploreItems,
      kepalaDesa,
      motto,
      wilayah,
      statistik,
      pengumuman,
      currentYear,
      formatDate,
      handleImageError,
      nextSlide,
      previousSlide,
      newsMediaUrl
    }
  }
}
</script>

<style scoped>
.ticker-track {
  animation: ticker-scroll 40s linear infinite;
}

@keyframes ticker-scroll {
  0% { transform: translateX(0); }
  100% { transform: translateX(-50%); }
}

@media (prefers-reduced-motion: reduce) {
  .ticker-track {
    animation: none;
  }
}
</style>