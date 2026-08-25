<template>
  <nav aria-label="Breadcrumb">
    <ol class="text-sm text-secondary-500 dark:text-secondary-400 flex items-center gap-1">
      <li>
        <router-link to="/" class="hover:text-secondary-700 dark:text-secondary-300">Home</router-link>
      </li>
      <template v-for="(crumb, index) in crumbs" :key="crumb.path">
        <li aria-hidden="true">/</li>
        <li>
          <router-link
            v-if="index < crumbs.length - 1"
            :to="crumb.path"
            class="hover:text-secondary-700 dark:text-secondary-300"
          >
            {{ crumb.label }}
          </router-link>
          <span v-else class="text-secondary-700 dark:text-secondary-300 font-medium" aria-current="page">
            {{ crumb.label }}
          </span>
        </li>
      </template>
    </ol>
  </nav>
</template>

<script setup>
import { computed } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()

const crumbs = computed(() =>
  route.matched
    .filter((r) => r.meta?.breadcrumb)
    .map((r) => ({ label: r.meta.breadcrumb, path: r.path || '/' }))
)
</script>
