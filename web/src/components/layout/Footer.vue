<template>
  <footer class="bg-gray-900 text-gray-300">
    <div class="border-b border-gray-700">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
        <div class="flex flex-col md:flex-row md:items-center md:justify-between gap-4">
          <div>
            <p class="text-sm text-gray-400 mb-1">Tetap terhubung dengan Pemerintah {{ brandName }}</p>
            <div class="flex items-center gap-3">
              <a v-for="sm in socialMedia" :key="sm.label" :href="sm.href" :aria-label="sm.label" class="w-9 h-9 rounded-full bg-gray-800 hover:bg-emerald-600 flex items-center justify-center transition-colors">
                <Icon :name="sm.icon" class="w-4 h-4" />
              </a>
            </div>
          </div>
          <div class="text-xs text-gray-500 md:text-right">
            <p class="mb-1">Bagian dari Sistem Pemerintahan Republik Indonesia</p>
            <p>Berkoordinasi dengan Kemendagri · Kemendes PDTT</p>
          </div>
        </div>
      </div>
    </div>

    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
      <div class="grid grid-cols-1 md:grid-cols-4 gap-8">
        <div class="md:col-span-1">
          <div class="flex items-center gap-3 mb-4">
            <div class="w-10 h-10 bg-emerald-600 rounded flex items-center justify-center">
              <span class="text-white font-bold text-sm">{{ brandInitials }}</span>
            </div>
            <h3 class="text-white font-bold text-lg">{{ brandName }}</h3>
          </div>
          <p class="text-sm leading-relaxed text-gray-400">
            Website resmi Pemerintah {{ brandName }}. Portal informasi dan layanan publik desa berbasis digital.
          </p>
        </div>

        <div>
          <h3 class="text-white font-bold text-sm uppercase tracking-wider mb-4">Kontak</h3>
          <ul class="space-y-2 text-sm">
            <li v-if="desaInfo.phone" class="flex items-start gap-2">
              <Icon name="phone" class="w-4 h-4 text-emerald-500 mt-0.5 flex-shrink-0" />
              <a :href="`tel:${desaInfo.phone}`" class="hover:text-white transition-colors">{{ desaInfo.phone }}</a>
            </li>
            <li v-if="desaInfo.email" class="flex items-start gap-2">
              <Icon name="mail" class="w-4 h-4 text-emerald-500 mt-0.5 flex-shrink-0" />
              <a :href="`mailto:${desaInfo.email}`" class="hover:text-white transition-colors break-all">{{ desaInfo.email }}</a>
            </li>
            <li v-if="desaInfo.address" class="flex items-start gap-2">
              <Icon name="pin" class="w-4 h-4 text-emerald-500 mt-0.5 flex-shrink-0" />
              <span>{{ desaInfo.address }}</span>
            </li>
            <li v-if="!desaInfo.phone && !desaInfo.email && !desaInfo.address" class="text-gray-500 italic text-sm">Belum ada informasi kontak.</li>
          </ul>
        </div>

        <div>
          <h3 class="text-white font-bold text-sm uppercase tracking-wider mb-4">Tautan Cepat</h3>
          <ul class="space-y-2 text-sm">
            <li><RouterLink to="/profil" class="hover:text-white transition-colors inline-flex items-center gap-1.5"><span class="w-1 h-1 bg-emerald-500 rounded-full"></span>Profil Desa</RouterLink></li>
            <li><RouterLink to="/berita" class="hover:text-white transition-colors inline-flex items-center gap-1.5"><span class="w-1 h-1 bg-emerald-500 rounded-full"></span>Berita</RouterLink></li>
            <li><RouterLink to="/infografik" class="hover:text-white transition-colors inline-flex items-center gap-1.5"><span class="w-1 h-1 bg-emerald-500 rounded-full"></span>Infografik</RouterLink></li>
            <li><RouterLink to="/peta" class="hover:text-white transition-colors inline-flex items-center gap-1.5"><span class="w-1 h-1 bg-emerald-500 rounded-full"></span>Peta Wilayah</RouterLink></li>
            <li><RouterLink to="/umkm" class="hover:text-white transition-colors inline-flex items-center gap-1.5"><span class="w-1 h-1 bg-emerald-500 rounded-full"></span>UMKM</RouterLink></li>
            <li><RouterLink to="/ppid" class="hover:text-white transition-colors inline-flex items-center gap-1.5"><span class="w-1 h-1 bg-emerald-500 rounded-full"></span>PPID</RouterLink></li>
          </ul>
        </div>

        <div>
          <h3 class="text-white font-bold text-sm uppercase tracking-wider mb-4">Layanan</h3>
          <ul class="space-y-2 text-sm">
            <li><a href="/ppid" class="hover:text-white transition-colors inline-flex items-center gap-1.5"><span class="w-1 h-1 bg-emerald-500 rounded-full"></span>Permohonan Dokumen</a></li>
            <li><a href="#" class="hover:text-white transition-colors inline-flex items-center gap-1.5"><span class="w-1 h-1 bg-emerald-500 rounded-full"></span>Pengaduan Masyarakat</a></li>
            <li><a href="#" class="hover:text-white transition-colors inline-flex items-center gap-1.5"><span class="w-1 h-1 bg-emerald-500 rounded-full"></span>Sistem Informasi Desa</a></li>
            <li><a href="#" class="hover:text-white transition-colors inline-flex items-center gap-1.5"><span class="w-1 h-1 bg-emerald-500 rounded-full"></span>Transparansi APBDesa</a></li>
          </ul>
        </div>
      </div>
    </div>

    <div class="border-t border-gray-700 bg-gray-950">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-5">
        <div class="flex flex-col md:flex-row md:items-center md:justify-between gap-3 text-xs text-gray-500">
          <p>&copy; {{ currentYear }} Pemerintah {{ brandName }}. Semua hak dilindungi.</p>
          <p class="flex items-center gap-1.5">
            <Icon name="shield" class="w-3 h-3 text-emerald-500" />
            Konten dikelola resmi oleh Pemerintah {{ brandName }}
          </p>
        </div>
      </div>
    </div>
  </footer>
</template>

<script>
import { computed } from 'vue'
import { useDesaInfo } from '../../composables/useDesaInfo'
import Icon from '../common/Icon.vue'

export default {
  name: 'Footer',
  components: { Icon },
  setup() {
    const { desaInfo, initials } = useDesaInfo()
    const brandName = computed(() => desaInfo.value?.name || 'Desa')
    const brandInitials = computed(() => initials(desaInfo.value?.name))

    const socialMedia = [
      { label: 'Facebook', icon: 'facebook', href: '#' },
      { label: 'Instagram', icon: 'instagram', href: '#' },
      { label: 'YouTube', icon: 'youtube', href: '#' }
    ]

    const currentYear = new Date().getFullYear()
    return { desaInfo, brandName, brandInitials, socialMedia, currentYear }
  }
}
</script>