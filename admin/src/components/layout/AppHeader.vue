<template>
  <header class="bg-white border-b border-secondary-200 dark:bg-secondary-800 dark:border-secondary-700 px-4 py-3 flex items-center justify-between">
    <!-- Left: hamburger + title -->
    <div class="flex items-center gap-3">
      <button
        class="rounded p-1.5 text-secondary-600 hover:bg-secondary-100 dark:text-secondary-300 dark:hover:bg-secondary-700 focus:outline-none focus:ring-2 focus:ring-primary-500 md:hidden"
        aria-label="Toggle sidebar"
        @click="uiStore.toggleSidebar"
      >
        <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
        </svg>
      </button>
      <span class="text-sm font-semibold text-secondary-700 dark:text-secondary-100 hidden md:block">Admin Panel</span>
    </div>

    <!-- Center: theme toggle + language switcher -->
    <div class="flex items-center gap-2">
      <ThemeToggle />
      <LanguageSwitcher />
    </div>

    <!-- Right: avatar dropdown -->
    <div class="relative" ref="dropdownRef">
      <button
        class="flex items-center gap-2 rounded-full focus:outline-none focus:ring-2 focus:ring-primary-500"
        aria-label="User menu"
        @click="toggleDropdown"
      >
        <!-- Avatar: image or initial -->
        <SafeImg
          v-if="authStore.currentUser?.profile_image_url"
          :src="authStore.currentUser.profile_image_url"
          :alt="authStore.currentUser.name"
          circle
          class="w-8 h-8"
        />
        <span
          v-else
          class="w-8 h-8 rounded-full bg-primary-600 text-white text-sm font-semibold flex items-center justify-center select-none"
        >
          {{ userInitial }}
        </span>
        <span class="text-sm text-secondary-700 dark:text-secondary-100 hidden sm:block">{{ authStore.currentUser?.name ?? 'Admin' }}</span>
        <svg class="w-4 h-4 text-secondary-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
        </svg>
      </button>

      <!-- Dropdown menu -->
      <transition
        enter-active-class="transition ease-out duration-100"
        enter-from-class="opacity-0 scale-95"
        enter-to-class="opacity-100 scale-100"
        leave-active-class="transition ease-in duration-75"
        leave-from-class="opacity-100 scale-100"
        leave-to-class="opacity-0 scale-95"
      >
        <div
          v-if="open"
          class="absolute right-0 mt-2 w-44 bg-white dark:bg-secondary-800 rounded-lg shadow-lg border border-secondary-200 dark:border-secondary-700 py-1 z-50"
          role="menu"
        >
          <button
            class="w-full text-left px-4 py-2 text-sm text-secondary-700 dark:text-secondary-100 hover:bg-secondary-50 dark:hover:bg-secondary-700 transition-colors"
            role="menuitem"
            @click="goToProfile"
          >
            {{ t('navigation.myProfile') }}
          </button>
          <hr class="my-1 border-secondary-200 dark:border-secondary-700" />
          <button
            class="w-full text-left px-4 py-2 text-sm text-danger-600 dark:text-danger-400 hover:bg-danger-50 dark:hover:bg-danger-900/20 transition-colors"
            role="menuitem"
            @click="logout"
          >
            {{ t('auth.logout') }}
          </button>
        </div>
      </transition>
    </div>
  </header>
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '../../stores/auth'
import { useUiStore } from '../../stores/ui'
import LanguageSwitcher from '../common/LanguageSwitcher.vue'
import ThemeToggle from '../common/ThemeToggle.vue'
import SafeImg from '../common/SafeImg.vue'

const router = useRouter()
const { t } = useI18n()
const authStore = useAuthStore()
const uiStore = useUiStore()

const open = ref(false)
const dropdownRef = ref(null)

const userInitial = computed(() => {
  const name = authStore.currentUser?.name ?? 'A'
  return name.charAt(0).toUpperCase()
})

function toggleDropdown() {
  open.value = !open.value
}

function goToProfile() {
  open.value = false
  router.push('/me')
}

function logout() {
  open.value = false
  authStore.clearTokens()
  router.push('/login')
}

function onClickOutside(e) {
  if (dropdownRef.value && !dropdownRef.value.contains(e.target)) {
    open.value = false
  }
}

onMounted(() => document.addEventListener('click', onClickOutside, true))
onBeforeUnmount(() => document.removeEventListener('click', onClickOutside, true))
</script>