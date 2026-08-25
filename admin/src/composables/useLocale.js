import { useI18n } from 'vue-i18n'
import { computed } from 'vue'

export function useLocale() {
  const i18n = useI18n()

  const currentLocale = computed({
    get: () => i18n.locale.value,
    set: (value) => {
      i18n.locale.value = value
      localStorage.setItem('locale', value)
      document.documentElement.lang = value
    },
  })

  const availableLocales = computed(() => [
    { code: 'id', name: 'Bahasa Indonesia' },
    { code: 'en', name: 'English' },
  ])

  return {
    currentLocale,
    availableLocales,
    t: i18n.t,
  }
}
