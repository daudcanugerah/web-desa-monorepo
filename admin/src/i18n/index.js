import { createI18n } from 'vue-i18n'
import id from './locales/id.json'
import en from './locales/en.json'

const messages = {
  id,
  en,
}

const i18n = createI18n({
  legacy: false,
  locale: localStorage.getItem('locale') || 'id',
  fallbackLocale: 'id',
  messages,
})

export default i18n
