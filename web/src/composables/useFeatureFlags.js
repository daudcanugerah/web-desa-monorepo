import { ref } from 'vue'
import { env } from '../utils/runtimeEnv.js'

const parseFlag = (value) => value === true || value === 'true'

export const features = ref({
  galleryPublic: parseFlag(env.VITE_FEATURE_GALLERY_PUBLIC),
  beritaSearch: parseFlag(env.VITE_FEATURE_BERITA_SEARCH),
  infographicSort: parseFlag(env.VITE_FEATURE_INFOGRAPHIC_SORT)
})

export const useFeatureFlags = () => features
