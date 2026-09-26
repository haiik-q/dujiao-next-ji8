<template>
  <RouterLink :to="`/products/${product.slug}`" class="j8-card" :class="{ 'is-sold-out': soldOut }">
    <div class="j8-card-top">
      <span class="j8-card-icon" :class="{ 'is-logo': !productImage }">
        <Ji8BrandIcon :slug="product.category?.slug" :name="title" :image="iconImage" />
      </span>
      <span class="j8-card-badge" :class="{ sold: soldOut }">{{ badgeLabel }}</span>
    </div>
    <h3>{{ title }}</h3>
    <div class="j8-card-chips">
      <template v-if="tags.length">
        <span v-for="(tag, i) in tags" :key="i" :class="`tone-${(i % 3) + 1}`">{{ tag }}</span>
      </template>
      <template v-else>
        <span class="tone-1">{{ getFulfillmentTypeLabel(product.fulfillment_type) }}</span>
        <span class="tone-2">{{ getPurchaseTypeLabel(product.purchase_type) }}</span>
      </template>
    </div>
    <div class="j8-card-bottom">
      <div class="j8-card-price">
        <small>{{ promo ? t('products.promotionTag') : t('ji8.card.priceLabel') }}</small>
        <strong>{{ priceText }}</strong>
        <del v-if="promo">{{ originalPriceText }}</del>
      </div>
      <span class="j8-card-stock">{{ stockText }}</span>
    </div>
  </RouterLink>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { getFirstImageUrl, getImageUrl } from '../../../utils/image'
import { useLocalized, useProductLabels } from '../../../composables/useProduct'
import { formatMoney } from '../utils/price'
import Ji8BrandIcon from './Ji8BrandIcon.vue'

const props = withDefaults(defineProps<{
  product: any
  maxTags?: number
}>(), {
  maxTags: 3,
})

const { t } = useI18n()
const { getLocalizedText, siteCurrency } = useLocalized()
const {
  isSoldOut,
  hasPromotionPrice,
  getPromotionPriceAmount,
  hasPromotionRules,
  hasWholesalePrices,
  getFulfillmentTypeLabel,
  getPurchaseTypeLabel,
  getStockStatusLabel,
} = useProductLabels()

const title = computed(() => getLocalizedText(props.product?.title))
const soldOut = computed(() => isSoldOut(props.product))
const promo = computed(() => hasPromotionPrice(props.product))

// 图标：商品首图 → 分类图标 → 品牌 logo / 通用图标 → 标题首字（后两级及图片加载失败回落由 Ji8BrandIcon 处理）
const productImage = computed(() => getFirstImageUrl(props.product?.images))
const iconImage = computed(() => productImage.value || getImageUrl(props.product?.category?.icon))

// 标签优先（最多 maxTags 枚，色调 tone-1/2/3 循环）；没有标签时才回落到交付 / 购买类型 chips
const tags = computed<string[]>(() => {
  const raw = props.product?.tags
  if (!Array.isArray(raw)) return []
  return raw
    .map((tag: unknown) => String(tag).trim())
    .filter(Boolean)
    .slice(0, props.maxTags)
})

const badgeLabel = computed(() => {
  if (soldOut.value) return t('products.stockStatus.outOfStock')
  if (promo.value) return t('products.promotionTag')
  if (hasPromotionRules(props.product)) return t('products.promotionBadge')
  if (hasWholesalePrices(props.product)) return t('products.wholesaleTag')
  return t('ji8.card.recommended')
})

const priceText = computed(() =>
  formatMoney(promo.value ? getPromotionPriceAmount(props.product) : props.product?.price_amount, siteCurrency.value),
)
const originalPriceText = computed(() => formatMoney(props.product?.price_amount, siteCurrency.value))

// 精确库存数：与 useProduct.getStockStatusLabel 同源（manual → manual_stock_available，其余 → auto_stock_available）
const availableStock = computed<number | null>(() => {
  const p = props.product
  const count = Number(p?.fulfillment_type === 'manual' ? p?.manual_stock_available : p?.auto_stock_available)
  return Number.isFinite(count) && count > 0 ? Math.floor(count) : null
})

// 库存文案：售罄 → exact 模式（无限=库存充足 / 有限=库存 N）→ status/range/hidden 模式沿用通用状态文案
const stockText = computed(() => {
  const p = props.product
  if (soldOut.value) return t('products.stockStatus.outOfStock')
  const exact = String(p?.stock_display_mode || '').trim() === 'exact' && p?.stock_quantity_hidden !== true
  if (exact) {
    if (p?.stock_status === 'unlimited') return t('ji8.card.stockUnlimited')
    if (availableStock.value !== null) return t('ji8.card.stock', { count: availableStock.value })
  }
  return getStockStatusLabel(p)
})
</script>
