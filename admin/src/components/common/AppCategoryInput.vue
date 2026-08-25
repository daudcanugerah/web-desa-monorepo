<template>
  <div>
    <div class="flex items-end gap-2">
      <div class="flex-1">
        <AppInput
          :id="id"
          v-model="categoryName"
          :label="label"
          :required="required"
          :error="error"
          :placeholder="placeholder"
          :list="listId"
          @change="onNameInput"
        />
        <datalist :id="listId">
          <option v-for="cat in suggestions" :key="cat.id || cat.name" :value="cat.name" />
        </datalist>
      </div>
      <AppButton variant="secondary" type="button" @click="$emit('manage')">
        {{ manageLabel }}
      </AppButton>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  modelValue: {
    type: Object,
    default: () => ({ id: '', name: '' }),
  },
  suggestions: {
    type: Array,
    default: () => [],
  },
  service: { type: Object, required: true },
  label: { type: String, default: 'Kategori' },
  placeholder: { type: String, default: 'Contoh: Pengumuman, Berita' },
  required: { type: Boolean, default: false },
  error: { type: String, default: '' },
  id: { type: String, default: 'app-category-input' },
  manageLabel: { type: String, default: 'Kelola Kategori' },
})

const emit = defineEmits(['update:modelValue', 'manage'])

const listId = computed(() => `${props.id}-suggestions`)

const categoryName = computed({
  get: () => props.modelValue?.name || '',
  set: (val) => {
    emit('update:modelValue', { id: props.modelValue?.id || '', name: val })
  },
})

function onNameInput() {
  const name = (props.modelValue?.name || '').trim()
  if (!name) {
    emit('update:modelValue', { id: '', name: '' })
    return
  }
  const match = props.suggestions.find((c) => c.name?.toLowerCase() === name.toLowerCase())
  emit('update:modelValue', { id: match?.id || '', name })
}
</script>