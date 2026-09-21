<template>
  <span class="j8-brand-icon">
    <img v-if="showImage" :src="image" :alt="name || ''" loading="lazy" decoding="async" @error="errored = true" />
    <svg v-else-if="brand" class="j8-brand-svg" :viewBox="brand.viewBox" fill="currentColor" aria-hidden="true">
      <path :d="brand.path" />
    </svg>
    <component :is="generic" v-else-if="generic" aria-hidden="true" />
    <b v-else>{{ initial }}</b>
  </span>
</template>

<script setup lang="ts">
import { computed, ref, watch, type Component } from 'vue'
import { BookOpen, CreditCard, Palette, UserRound, Wallet } from 'lucide-vue-next'
import { matchBrandIcon } from '../utils/brandIcons'

/**
 * 分类 / 商品图标，按优先级回落：
 * 后台图片 → simple-icons 品牌 logo（按 slug/名称关键字）→ lucide 通用图标（按品类关键字）→ 名称首字。
 */
const props = defineProps<{
  slug?: string
  name?: string
  image?: string
}>()

// 图片加载失败则继续向下回落；换图后重置，避免旧的失败状态挡住新图
const errored = ref(false)
watch(
  () => props.image,
  () => {
    errored.value = false
  },
)

const showImage = computed(() => Boolean(props.image) && !errored.value)
const brand = computed(() => matchBrandIcon(props.slug, props.name))

const GENERIC_RULES: ReadonlyArray<readonly [RegExp, Component]> = [
  [/教程|tutorial|guide|course|学/, BookOpen],
  [/生图|prompt|绘|art|image|midjourney|stable|画/, Palette],
  [/卡|card|visa|master/, CreditCard],
  [/账号|account/, UserRound],
  [/充值|recharge|top-up|topup/, Wallet],
]
const generic = computed<Component | null>(() => {
  const haystack = `${props.slug ?? ''} ${props.name ?? ''}`.toLowerCase()
  return GENERIC_RULES.find(([pattern]) => pattern.test(haystack))?.[1] ?? null
})

const initial = computed(() => (props.name ?? '').trim().charAt(0).toUpperCase() || '?')
</script>
