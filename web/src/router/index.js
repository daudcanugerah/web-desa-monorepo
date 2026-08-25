import { createRouter, createWebHistory } from 'vue-router'
import Layout from '../components/layout/Layout.vue'
import Home from '../pages/Home.vue'
import Profil from '../pages/Profil.vue'
import Infografik from '../pages/Infografik.vue'
import Peta from '../pages/Peta.vue'
import Berita from '../pages/Berita.vue'
import BeritaDetail from '../pages/BeritaDetail.vue'
import UMKM from '../pages/UMKM.vue'
import PPID from '../pages/PPID.vue'
import Galeri from '../pages/Galeri.vue'
import GaleriDetail from '../pages/GaleriDetail.vue'

const routes = [
  {
    path: '/',
    component: Layout,
    children: [
      { path: '', component: Home },
      { path: 'profil', component: Profil },
      { path: 'infografik', component: Infografik },
      { path: 'peta', component: Peta },
      { path: 'berita', component: Berita },
      { path: 'berita/:id', component: BeritaDetail },
      { path: 'umkm', component: UMKM },
      { path: 'ppid', component: PPID },
      { path: 'galeri', component: Galeri },
      { path: 'galeri/:id', component: GaleriDetail },
      { path: ':pathMatch(.*)*', component: () => import('../pages/NotFound.vue') }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes,
  scrollBehavior(to, from, savedPosition) {
    if (savedPosition) return savedPosition
    if (to.hash) return { el: to.hash, behavior: 'smooth' }
    return { top: 0, behavior: 'smooth' }
  }
})

export default router
