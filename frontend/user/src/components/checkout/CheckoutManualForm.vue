<template>
  <div
    v-if="manualFormProducts.length"
    class="rounded-2xl border bg-card text-card-foreground p-6"
  >
    <h2 class="mb-2 text-lg font-bold text-foreground">{{ t('checkout.manualFormTitle') }}</h2>
    <p class="mb-4 text-xs text-muted-foreground">{{ t('checkout.manualFormTip') }}</p>
    <div class="space-y-5">
      <div
        v-for="manualItem in manualFormProducts"
        :key="manualItem.itemKey"
        class="rounded-xl border bg-secondary p-4"
      >
        <h3 class="mb-3 text-sm font-semibold text-foreground">{{ manualItemTitle(manualItem) }}</h3>
        <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
          <div v-for="field in manualItem.fields" :key="`${manualItem.itemKey}-${field.key}`" class="space-y-1.5">
            <label class="text-xs font-semibold text-muted-foreground">
              {{ getManualFieldLabel(field) }}
              <span v-if="field.required" class="ml-1 text-destructive">*</span>
            </label>

            <Textarea
              v-if="field.type === 'textarea'"
              :model-value="getFieldValue(manualItem.itemKey, field.key)"
              @update:model-value="updateFieldValue(manualItem.itemKey, field.key, $event)"
              rows="3"
              :placeholder="getManualFieldPlaceholder(field)"
            />

            <select
              v-else-if="field.type === 'select'"
              :value="getFieldValue(manualItem.itemKey, field.key)"
              @change="updateFieldValue(manualItem.itemKey, field.key, ($event.target as HTMLSelectElement).value)"
              :class="selectClass"
            >
              <option value="">{{ t('checkout.manualFormSelectPlaceholder') }}</option>
              <option v-for="option in field.options" :key="option" :value="option">{{ option }}</option>
            </select>

            <div v-else-if="field.type === 'radio'" class="space-y-2 rounded-xl border bg-secondary p-3">
              <label v-for="option in field.options" :key="option" class="flex items-center gap-2 text-sm text-muted-foreground">
                <input
                  :checked="getFieldValue(manualItem.itemKey, field.key) === option"
                  @change="updateFieldValue(manualItem.itemKey, field.key, option)"
                  type="radio"
                  :name="`manual-radio-${manualItem.itemKey}-${field.key}`"
                  :value="option"
                  class="h-4 w-4 accent-primary"
                />
                <span>{{ option }}</span>
              </label>
            </div>

            <div v-else-if="field.type === 'checkbox'" class="space-y-2 rounded-xl border bg-secondary p-3">
              <label v-for="option in field.options" :key="option" class="flex items-center gap-2 text-sm text-muted-foreground">
                <input
                  :checked="isCheckboxChecked(manualItem.itemKey, field.key, option)"
                  @change="toggleCheckboxValue(manualItem.itemKey, field.key, option, ($event.target as HTMLInputElement).checked)"
                  type="checkbox"
                  :value="option"
                  class="h-4 w-4 accent-primary"
                />
                <span>{{ option }}</span>
              </label>
            </div>

            <Input
              v-else
              :model-value="getFieldValue(manualItem.itemKey, field.key)"
              @update:model-value="updateFieldValue(manualItem.itemKey, field.key, $event)"
              :type="field.type === 'number' ? 'number' : field.type === 'email' ? 'email' : field.type === 'phone' ? 'tel' : 'text'"
              :placeholder="getManualFieldPlaceholder(field)"
            />

            <XHandleCheck
              v-if="field.key === 'x_handle'"
              :handle="String(getFieldValue(manualItem.itemKey, field.key) || '')"
            />

            <p
              v-if="submitAttempted && manualFieldError(manualItem.itemKey, field.key)"
              class="text-xs text-destructive"
            >
              {{ manualFieldError(manualItem.itemKey, field.key) }}
            </p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { useLocalized } from '../../composables/useProduct'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import XHandleCheck from './XHandleCheck.vue'

const selectClass =
  'h-9 w-full rounded-md border border-input bg-transparent px-3 text-sm shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring'

interface ManualFormField {
  key: string
  type: string
  required: boolean
  label?: Record<string, string>
  placeholder?: Record<string, string>
  regex?: string
  min?: number
  max?: number
  max_len?: number
  options: string[]
}

interface ManualFormProduct {
  itemKey: string
  productId: number
  title: any
  fields: ManualFormField[]
  skuCount: number
}

const props = defineProps<{
  manualFormProducts: ManualFormProduct[]
  modelValue: Record<string, Record<string, any>>
  submitAttempted: boolean
  getManualFieldLabel: (field: ManualFormField) => string
  getManualFieldPlaceholder: (field: ManualFormField) => string
  manualFieldError: (itemKey: string, fieldKey: string) => string
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: Record<string, Record<string, any>>): void
}>()

const { t } = useI18n()
const { getLocalizedText } = useLocalized()

const manualItemTitle = (manualItem: ManualFormProduct) => {
  const productTitle = getLocalizedText(manualItem.title)
  if (manualItem.skuCount <= 1) return productTitle
  return `${productTitle} (${t('checkout.manualFormAppliesToSkuCount', { count: manualItem.skuCount })})`
}

const getFieldValue = (itemKey: string, fieldKey: string) => {
  return props.modelValue[itemKey]?.[fieldKey] ?? ''
}

const updateFieldValue = (itemKey: string, fieldKey: string, value: any) => {
  const updated = { ...props.modelValue }
  if (!updated[itemKey]) {
    updated[itemKey] = {}
  }
  updated[itemKey] = { ...updated[itemKey], [fieldKey]: value }
  emit('update:modelValue', updated)
}

const isCheckboxChecked = (itemKey: string, fieldKey: string, option: string) => {
  const value = props.modelValue[itemKey]?.[fieldKey]
  return Array.isArray(value) && value.includes(option)
}

const toggleCheckboxValue = (itemKey: string, fieldKey: string, option: string, checked: boolean) => {
  const current = props.modelValue[itemKey]?.[fieldKey]
  const list = Array.isArray(current) ? [...current] : []
  if (checked) {
    if (!list.includes(option)) list.push(option)
  } else {
    const idx = list.indexOf(option)
    if (idx !== -1) list.splice(idx, 1)
  }
  updateFieldValue(itemKey, fieldKey, list)
}
</script>
