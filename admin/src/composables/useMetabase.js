import { ref, onMounted } from 'vue'

const isMetabaseInitialized = ref(false)

export function useMetabase() {
  const metabaseUrl = import.meta.env.VITE_METABASE_URL || 'http://localhost:3000'

  function initializeMetabase() {
    if (isMetabaseInitialized.value) return Promise.resolve()

    return new Promise((resolve, reject) => {
      // Load Metabase embed script
      const embedScript = document.createElement('script')
      embedScript.defer = true
      embedScript.src = `${metabaseUrl}/app/embed.js`
      
      embedScript.onload = () => {
        // Define Metabase config
        const configScript = document.createElement('script')
        configScript.textContent = `
          function defineMetabaseConfig(config) {
            window.metabaseConfig = config;
          }
          defineMetabaseConfig({
            "theme": {"preset": "light"},
            "isGuest": true,
            "instanceUrl": "${metabaseUrl}"
          });
        `
        
        document.head.appendChild(configScript)
        isMetabaseInitialized.value = true
        resolve()
      }
      
      embedScript.onerror = () => {
        reject(new Error('Failed to load Metabase embed script'))
      }
      
      document.head.appendChild(embedScript)
    })
  }

  return {
    initializeMetabase,
    isMetabaseInitialized,
  }
}