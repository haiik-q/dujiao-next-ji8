<template>
  <div class="j8-cat-list">
    <button type="button" class="j8-cat-item" :class="{ 'is-active': selected === null }" @click="emit('select', null)">
      <span class="j8-cat-logo hot"><Star /></span>
      <span>{{ t('products.allCategories') }}</span>
      <b v-if="countOf(0) !== undefined">{{ countOf(0) }}</b>
    </button>
    <template v-for="g in groups" :key="g.id">
      <button type="button" class="j8-cat-item" :class="{ 'is-active': selected === g.id }" @click="emit('select', g.id)">
        <span class="j8-cat-logo"><Ji8BrandIcon :slug="g.slug" :name="name(g)" :image="icon(g)" /></span>
        <span>{{ name(g) }}</span>
        <b v-if="countOf(g.id) !== undefined">{{ countOf(g.id) }}</b>
      </button>
      <button
        v-for="c in g.children"
        :key="c.id"
        type="button"
        class="j8-cat-item is-child"
        :class="{ 'is-active': selected === c.id }"
        @click="emit('select', c.id)"
      >
        <span class="j8-cat-logo"><Ji8BrandIcon :slug="c.slug" :name="name(c)" :image="icon(c)" /></span>
        <span>{{ name(c) }}</span>
        <b v-if="countOf(c.id) !== undefined">{{ countOf(c.id) }}</b>
      </button>
    </template>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { Star } from 'lucide-vue-next'
import type { CategoryGroup, PublicCategory } from '../../../utils/category'
import { getImageUrl } from '../../../utils/image'
import { useLocalized } from '../../../composables/useProduct'
import Ji8BrandIcon from './Ji8BrandIcon.vue'

const props = defineProps<{
  groups: CategoryGroup[]
  selected: number | null
  total: number
  counts: Record<number, number>
}>()

const emit = defineEmits<{
  select: [id: number | null]
}>()

const { t } = useI18n()
const { getLocalizedText } = useLocalized()

const name = (category: PublicCategory): string => getLocalizedText(category.name)
// 后台图标为空或加载失败时，由 Ji8BrandIcon 回落到品牌 logo / 通用图标 / 首字
const icon = (category: PublicCategory): string => (category.icon ? getImageUrl(category.icon) : '')

// id 0 = 全部商品；选中项显示当前列表总数，其余读取预取的分类计数
const countOf = (id: number): number | undefined => (props.selected === id ? props.total : props.counts[id])
</script>
