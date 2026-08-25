<template>
  <div class="relative">
    <div ref="quillRef"></div>
    
    <!-- Table Grid Picker -->
    <div
      v-if="showTablePicker"
      class="absolute top-12 left-0 bg-white dark:bg-secondary-800 border border-secondary-300 dark:border-secondary-600 rounded shadow-lg p-2 z-50"
      @mouseleave="showTablePicker = false"
    >
      <div class="mb-2 text-xs text-secondary-600 dark:text-secondary-400 font-semibold">{{ selectedRows }} × {{ selectedCols }}</div>
      <div class="grid gap-1" style="grid-template-columns: repeat(10, 1fr)">
        <button
          v-for="(_, idx) in 100"
          :key="idx"
          type="button"
          class="w-5 h-5 border border-secondary-300 dark:border-secondary-600 rounded hover:bg-blue-500 dark:hover:bg-blue-600"
          :class="Math.ceil((idx + 1) / 10) <= selectedRows && ((idx + 1) % 10 || 10) <= selectedCols ? 'bg-blue-500' : 'bg-white dark:bg-secondary-800'"
          @mouseenter="selectedRows = Math.ceil((idx + 1) / 10); selectedCols = (idx + 1) % 10 || 10"
          @click="insertTable(Math.ceil((idx + 1) / 10), (idx + 1) % 10 || 10)"
        />
      </div>
    </div>
  </div>
</template>

<style scoped>
:deep(.ql-editor) {
  min-height: 200px;
}

:deep(.ql-editor table) {
  border-collapse: collapse;
  width: 100%;
}

:deep(.ql-editor table td,
.ql-editor table th) {
  border: 1px solid #ddd;
  padding: 8px;
}

:deep(.ql-editor table th) {
  background-color: #f3f4f6;
}

:deep(.ql-toolbar.ql-snow) {
  border: 1px solid #ccc;
  background-color: #fff;
}

:deep(.ql-container.ql-snow) {
  border: 1px solid #ccc;
  background-color: #fff;
}

:deep(.ql-editor.ql-blank::before) {
  color: #9ca3af;
  font-style: normal;
}
</style>

<!--
  Quill renders its own DOM that is NOT scoped to this component, and the
  `.dark` class lives on <html> (an ancestor of the editor). Scoped `:deep(.dark ...)`
  compiles to `[data-v] .dark ...` which can never match an ancestor, so dark
  overrides MUST live in a non-scoped block.
-->
<style>
.dark .ql-toolbar.ql-snow,
.dark .ql-container.ql-snow {
  border-color: #475569;
  background-color: #1f2937;
  color: #e5e7eb;
}

.dark .ql-editor.ql-blank::before {
  color: #9ca3af;
}

.dark .ql-editor table th {
  background-color: #374151;
}

.dark .ql-editor table td,
.dark .ql-editor table th {
  border-color: #4b5563;
}

.dark .ql-snow .ql-stroke {
  stroke: #cbd5e1;
}

.dark .ql-snow .ql-fill {
  fill: #cbd5e1;
}

.dark .ql-snow .ql-picker {
  color: #cbd5e1;
}

.dark .ql-snow .ql-picker-options {
  background-color: #1f2937;
  border-color: #475569;
  color: #cbd5e1;
}

.dark .ql-snow .ql-picker-label::before,
.dark .ql-snow .ql-picker-item::before {
  color: #cbd5e1;
}

.dark .ql-snow .ql-active .ql-stroke,
.dark .ql-snow button:hover .ql-stroke,
.dark .ql-snow .ql-picker-label:hover .ql-stroke,
.dark .ql-snow .ql-picker-label.ql-active .ql-stroke {
  stroke: #60a5fa;
}

.dark .ql-snow .ql-active .ql-fill,
.dark .ql-snow button:hover .ql-fill,
.dark .ql-snow .ql-picker-label:hover .ql-fill,
.dark .ql-snow .ql-picker-label.ql-active .ql-fill {
  fill: #60a5fa;
}

/* Toolbar button hover states */
.dark .ql-snow .ql-toolbar button:hover {
  color: #60a5fa;
}

.dark .ql-snow .ql-toolbar .ql-active {
  color: #60a5fa;
}

/* Selected dropdown option */
.dark .ql-snow .ql-picker-item.ql-selected {
  color: #60a5fa;
}
</style>

<script setup>
import 'quill/dist/quill.snow.css'
import 'quill-better-table-plus/dist/quill-better-table-plus.css'
import Quill from 'quill'
import quillBetterTablePlus from 'quill-better-table-plus'
import { ref, onMounted, watch, onBeforeUnmount } from 'vue'
import { newsService } from '../../services/news.service'
import { useNotificationStore } from '../../stores/notification'
import { getImageUrl } from '../../utils/imageUrl'

Quill.register({ 'modules/better-table-plus': quillBetterTablePlus }, true)

const props = defineProps({
  modelValue: { type: String, default: '' },
})
const emit = defineEmits(['update:modelValue'])

const notificationStore = useNotificationStore()
const quillRef = ref(null)
let quill = null
let tableModule = null

const showTablePicker = ref(false)
const selectedRows = ref(1)
const selectedCols = ref(1)

const options = {
  theme: 'snow',
  modules: {
    toolbar: {
      container: [
        ['bold', 'italic', 'underline', 'strike'],
        ['blockquote', 'code-block'],
        ['link', 'image', 'video', 'formula'],
        [{ header: 1 }, { header: 2 }],
        [{ list: 'ordered' }, { list: 'bullet' }, { list: 'check' }],
        [{ script: 'sub' }, { script: 'super' }],
        [{ indent: '-1' }, { indent: '+1' }],
        [{ direction: 'rtl' }],
        [{ size: ['small', false, 'large', 'huge'] }],
        [{ header: [1, 2, 3, 4, 5, 6, false] }],
        [{ color: [] }, { background: [] }],
        [{ font: [] }],
        [{ align: [] }],
        ['table'],
        ['clean'],
      ],
      handlers: {
        table: handleTableClick,
      },
    },
    table: false,
    'better-table-plus': {
      operationMenu: {
        items: {
          unmergeCells: {
            text: 'Unmerge cells',
          },
        },
      },
    },
    keyboard: {
      bindings: quillBetterTablePlus.keyboardBindings,
    },
  },
}

async function handleImageUpload() {
  const input = document.createElement('input')
  input.setAttribute('type', 'file')
  input.setAttribute('accept', 'image/*')
  input.click()

  input.onchange = async () => {
    const file = input.files[0]
    if (!file) return

    try {
      const res = await newsService.uploadMedia(file, 'image')
      const imageFileName = res.data.url || res.data.filename
      const imageUrl = getImageUrl(imageFileName)
      const range = quill.getSelection()
      quill.insertEmbed(range.index, 'image', imageUrl)
      notificationStore.success('Gambar berhasil diunggah')
    } catch (err) {
      notificationStore.error(err.response?.data?.error || 'Gagal mengunggah gambar')
    }
  }
}

async function handleVideoUpload() {
  const input = document.createElement('input')
  input.setAttribute('type', 'file')
  input.setAttribute('accept', 'video/*')
  input.click()

  input.onchange = async () => {
    const file = input.files[0]
    if (!file) return

    try {
      const res = await newsService.uploadMedia(file, 'video')
      const videoFileName = res.data.url || res.data.filename
      const videoUrl = getImageUrl(videoFileName)
      const range = quill.getSelection()
      quill.insertEmbed(range.index, 'video', videoUrl)
      notificationStore.success('Video berhasil diunggah')
    } catch (err) {
      notificationStore.error(err.response?.data?.error || 'Gagal mengunggah video')
    }
  }
}

function handleTableClick() {
  showTablePicker.value = !showTablePicker.value
  selectedRows.value = 1
  selectedCols.value = 1
}

function insertTable(rows, cols) {
  tableModule.insertTable(rows, cols)
  showTablePicker.value = false
}

function initValue(html) {
  if (!html) return
  const delta = quill.clipboard.convert({ html })
  const [range] = quill.selection.getRange()
  quill.updateContents(delta, Quill.sources.USER)
  quill.setSelection(delta.length() - (range?.length || 0), Quill.sources.SILENT)
  quill.scrollSelectionIntoView()
}

onMounted(() => {
  quill = new Quill(quillRef.value, options)
  tableModule = quill.getModule('better-table-plus')

  // Attach image and video handlers
  const toolbar = quill.getModule('toolbar')
  toolbar.addHandler('image', handleImageUpload)
  toolbar.addHandler('video', handleVideoUpload)

  // Set initial content
  if (props.modelValue) {
    initValue(props.modelValue)
  }

  // Handle content changes
  quill.on('text-change', () => {
    emit('update:modelValue', quill.root.innerHTML)
  })
})

watch(
  () => props.modelValue,
  (newVal) => {
    if (quill && quill.root.innerHTML !== newVal && newVal) {
      initValue(newVal)
    }
  }
)

onBeforeUnmount(() => {
  if (quill) {
    quill = null
  }
})
</script>
