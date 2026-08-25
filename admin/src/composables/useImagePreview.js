import { ref } from 'vue'
import { isValidImageType, isWithinSize, formatFileSize } from '../utils/validators'

export function useImagePreview({ maxMB = 10, warnMB = 5 } = {}) {
  const previewUrl = ref(null)
  const fileSize = ref('')
  const error = ref('')
  const warning = ref('')
  const selectedFile = ref(null)

  function reset() {
    previewUrl.value = null
    fileSize.value = ''
    error.value = ''
    warning.value = ''
    selectedFile.value = null
  }

  function processFile(file) {
    error.value = ''
    warning.value = ''

    if (!isValidImageType(file)) {
      error.value = 'Only JPEG, PNG, or WebP images are allowed.'
      return false
    }
    if (!isWithinSize(file, maxMB)) {
      error.value = `File exceeds the ${maxMB}MB limit.`
      return false
    }
    if (!isWithinSize(file, warnMB)) {
      warning.value = `File is larger than ${warnMB}MB — consider compressing it.`
    }

    selectedFile.value = file
    fileSize.value = formatFileSize(file.size)
    previewUrl.value = URL.createObjectURL(file)
    return true
  }

  return { previewUrl, fileSize, error, warning, selectedFile, processFile, reset }
}
