import { reactive, computed } from 'vue'

const EMAIL_RE = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
const PHONE_RE = /^[0-9+\-\s()]{7,20}$/
const URL_RE = /^https?:\/\/.+/i
const SLUG_RE = /^[a-z0-9]+(?:-[a-z0-9]+)*$/

export const rules = {
  required: (label = 'Field') => (v) =>
    (typeof v === 'string' ? v.trim().length > 0 : !!v) || `${label} wajib diisi`,
  email: (v) => !v || EMAIL_RE.test(v) || 'Format email tidak valid',
  minLength: (n, label = 'Field') => (v) =>
    !v || v.length >= n || `${label} minimal ${n} karakter`,
  maxLength: (n, label = 'Field') => (v) =>
    !v || v.length <= n || `${label} maksimal ${n} karakter`,
  url: (v) => !v || URL_RE.test(v) || 'URL harus dimulai dengan http:// atau https://',
  phone: (v) => !v || PHONE_RE.test(v) || 'Format telepon tidak valid',
  integer: (label = 'Field') => (v) =>
    v === null || v === undefined || v === '' || Number.isInteger(Number(v)) || `${label} harus berupa angka bulat`,
  positive: (label = 'Field') => (v) =>
    !v || Number(v) > 0 || `${label} harus lebih besar dari 0`,
  range: (min, max, label = 'Field') => (v) => {
    if (v === '' || v === null || v === undefined) return true
    const n = Number(v)
    if (Number.isNaN(n)) return `${label} harus berupa angka`
    return (n >= min && n <= max) || `${label} harus antara ${min} dan ${max}`
  },
  matches: (otherRef, label = 'Field') => (v) =>
    v === otherRef.value || `${label} tidak cocok`,
  slug: (v) => !v || SLUG_RE.test(v) || 'Hanya huruf kecil, angka, dan tanda hubung',
}

export function useForm(initialValues, schema) {
  const values = reactive({ ...initialValues })
  const errors = reactive({})
  const touched = reactive({})

  function validateField(field) {
    const fieldRules = schema[field]
    if (!fieldRules) return true
    const value = values[field]
    for (const rule of fieldRules) {
      const result = rule(value)
      if (result !== true) {
        errors[field] = result
        return false
      }
    }
    errors[field] = ''
    return true
  }

  function validateAll() {
    let valid = true
    let firstInvalid = null
    for (const field of Object.keys(schema)) {
      const ok = validateField(field)
      if (!ok) {
        valid = false
        if (!firstInvalid) firstInvalid = field
      }
    }
    if (firstInvalid) {
      const el = document.getElementById(firstInvalid)
      if (el && typeof el.focus === 'function') {
        el.focus()
        el.scrollIntoView({ behavior: 'smooth', block: 'center' })
      }
    }
    return valid
  }

  function touch(field) {
    touched[field] = true
    validateField(field)
  }

  function reset(toValues = initialValues) {
    Object.assign(values, toValues)
    Object.keys(errors).forEach((k) => (errors[k] = ''))
    Object.keys(touched).forEach((k) => (touched[k] = false))
  }

  const isValid = computed(() =>
    Object.values(errors).every((e) => !e)
  )

  return {
    values,
    errors,
    touched,
    validateField,
    validateAll,
    touch,
    reset,
    isValid,
  }
}