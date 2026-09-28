import { ref } from 'vue'

const parseFlag = (value) => value === true || value === 'true'

export const features = ref({
  galleryPublic: parseFlag(import.meta.env.VITE_FEATURE_GALLERY_PUBLIC),
  beritaSearch: parseFlag(import.meta.env.VITE_FEATURE_BERITA_SEARCH),
  infographicSort: parseFlag(import.meta.env.VITE_FEATURE_INFOGRAPHIC_SORT)
})

export const useFeatureFlags = () => features
