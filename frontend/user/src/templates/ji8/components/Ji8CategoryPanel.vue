<template>
  <aside class="j8-panel">
    <header>
      <span class="j8-eyebrow">{{ t('ji8.categories.eyebrow') }}</span>
      <h2>{{ t('ji8.categories.title') }}</h2>
    </header>
    <div v-if="loading && !groups.length" class="j8-skeleton-row"></div>
    <Ji8CategoryList
      v-else
      :groups="groups"
      :selected="selected"
      :total="total"
      :counts="counts"
      @select="emit('select', $event)"
    />
  </aside>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { CategoryGroup } from '../../../utils/category'
import Ji8CategoryList from './Ji8CategoryList.vue'

defineProps<{
  groups: CategoryGroup[]
  selected: number | null
  total: number
  counts: Record<number, number>
  loading?: boolean
}>()

const emit = defineEmits<{
  select: [id: number | null]
}>()

const { t } = useI18n()
</script>
