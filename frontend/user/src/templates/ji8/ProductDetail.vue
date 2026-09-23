<template>
  <div class="j8-container j8-pd">
    <!-- 加载骨架 -->
    <div v-if="loading" class="j8-pd-skeleton">
      <div class="j8-pd-sk-back"></div>
      <div class="j8-pd-card">
        <div class="j8-pd-media"><div class="j8-pd-sk-block"></div></div>
        <div class="j8-pd-info">
          <div class="j8-pd-sk-line is-lg"></div>
          <div class="j8-pd-sk-line"></div>
          <div class="j8-pd-sk-line is-sm"></div>
          <div class="j8-pd-sk-line is-btn"></div>
        </div>
      </div>
    </div>

    <template v-else-if="product">
      <router-link to="/products" class="j8-pd-back">
        <ChevronLeft />
        <span>{{ t('ji8.product.back') }}</span>
      </router-link>

      <section class="j8-pd-card">
        <!-- 左：图片 / 品牌 logo -->
        <div class="j8-pd-media">
          <div
            class="j8-pd-stage"
            @touchstart.passive="onTouchStart"
            @touchend.passive="onTouchEnd"
          >
            <img v-if="currentImage" :src="currentImage" :alt="title" />
            <Ji8BrandIcon v-else class="j8-pd-logo" :slug="product.category?.slug" :name="title" />
          </div>
          <div v-if="images.length > 1" class="j8-pd-thumbs">
            <button
              v-for="(image, index) in images"
              :key="index"
              type="button"
              :class="{ 'is-active': currentImage === image }"
              @click="currentImage = image"
            >
              <img :src="image" :alt="`${title} ${index + 1}`" loading="lazy" />
            </button>
          </div>
        </div>

        <!-- 右：购买区 -->
        <div class="j8-pd-info">
          <h1>{{ title }}</h1>

          <div class="j8-pd-chips">
            <span class="is-accent">
              <Zap v-if="product.fulfillment_type === 'auto'" />
              <Pencil v-else />
              {{ getFulfillmentTypeLabel(product.fulfillment_type) }}
            </span>
            <span :class="product.purchase_type === 'guest' ? 'is-accent' : 'is-warm'">
              {{ getPurchaseTypeLabel(product.purchase_type) }}
            </span>
            <span :class="product.is_sold_out ? 'is-danger' : 'is-accent'">{{ getStockStatusLabel(product) }}</span>
            <span v-for="(tag, index) in product.tags || []" :key="`tag-${index}`" class="is-accent">{{ tag }}</span>
            <button type="button" class="j8-pd-share" @click="share">
              <Share2 />
              {{ t('ji8.product.share') }}
            </button>
          </div>

          <div class="j8-pd-price-row">
            <div class="j8-pd-price">
              <span class="j8-pd-price-sym">{{ currencySymbol(siteCurrency) }}</span>
              <strong>{{ plainAmount(priceInfo.final) }}</strong>
              <del v-if="priceInfo.original !== null">{{ money(priceInfo.original) }}</del>
              <span v-if="priceInfo.tag" class="j8-pd-price-tag" :class="`is-${priceInfo.tag}`">
                {{ priceTagText }}
              </span>
            </div>

            <router-link v-if="promoCard" :to="promoCard.to" class="j8-pd-promo">
              <Crown />
              <span>
                <b>{{ promoCard.title }}</b>
                <small>{{ t('ji8.product.promoDesc') }}</small>
              </span>
              <em>{{ t('ji8.product.promoLink') }} →</em>
            </router-link>
          </div>

          <!-- 批发价 / 活动规则 -->
          <div v-if="selectedSkuWholesaleRules.length" class="j8-pd-note is-accent">
            <b>{{ t('products.wholesaleRulesTitle') }}</b>
            <span v-for="tier in selectedSkuWholesaleRules" :key="`${tier.sku_id || tier.sku_code || 'all'}-${tier.min_quantity}`">
              {{ formatWholesaleTier(tier) }}
            </span>
          </div>
          <div v-if="hasPromotionRules(product)" class="j8-pd-note is-warm">
            <b><Tag /> {{ t('products.promotionRulesTitle') }}</b>
            <span v-for="rule in getPromotionRules(product)" :key="rule.id">{{ formatPromotionRule(rule) }}</span>
          </div>

          <div v-if="activeSkus.length" class="j8-pd-field">
            <label>{{ t('ji8.product.skuTitle') }}</label>
            <div class="j8-pd-skus">
              <button
                v-for="sku in activeSkus"
                :key="sku.id"
                type="button"
                :class="{ 'is-active': normalizeSkuId(sku.id) === selectedSkuId, 'is-disabled': !isSkuPurchasable(sku) }"
                :disabled="!isSkuPurchasable(sku)"
                @click="selectedSkuId = normalizeSkuId(sku.id)"
              >
                <b>{{ skuDisplayText(sku) }}</b>
                <small>{{ money(sku.price_amount) }} · {{ skuStockText(sku) }}</small>
              </button>
            </div>
            <p v-if="requiresSKUSelection" class="j8-pd-hint">{{ t('productDetail.skuRequired') }}</p>
          </div>

          <div class="j8-pd-field">
            <label>{{ t('productDetail.quantity') }}</label>
            <div class="j8-pd-qty">
              <button
                type="button"
                :disabled="quantity <= quantityEffectiveMin"
                :aria-label="t('ji8.product.decrease')"
                @click="quantity = Math.max(quantityEffectiveMin, quantity - 1)"
              >
                <Minus />
              </button>
              <input
                type="text"
                inputmode="numeric"
                :value="quantity"
                @change="handleQuantityInput($event)"
                @keydown.enter.prevent="($event.target as HTMLInputElement)?.blur()"
              />
              <button
                type="button"
                :disabled="quantityEffectiveLimit !== null && quantity >= quantityEffectiveLimit"
                :aria-label="t('ji8.product.increase')"
                @click="quantity = quantity + 1"
              >
                <Plus />
              </button>
            </div>
          </div>

          <p v-if="description" class="j8-pd-desc">{{ description }}</p>

          <div ref="purchaseActionsRef" class="j8-pd-actions">
            <p v-if="cannotPurchaseReason" class="j8-pd-alert is-danger">{{ cannotPurchaseReason }}</p>
            <p v-if="purchaseWarning" class="j8-pd-alert is-warm">{{ purchaseWarning }}</p>

            <div class="j8-pd-total">
              <span>{{ t('ji8.product.payable') }}</span>
              <strong>{{ money(totalAmount) }}</strong>
            </div>
            <p class="j8-pd-total-note">{{ t('ji8.product.payableNote') }}</p>

            <button v-if="requiresLogin" type="button" class="j8-pd-buy" @click="goLogin">
              {{ t('productDetail.loginToBuy') }}
            </button>
            <div v-else class="j8-pd-buttons">
              <button type="button" class="j8-pd-cart" :disabled="!canPurchase" @click="addToCart">
                <ShoppingCart />
                {{ t('productDetail.addToCart') }}
              </button>
              <button type="button" class="j8-pd-buy" :disabled="!canPurchase" @click="buyNow">
                {{ t('productDetail.buyNow') }}
              </button>
            </div>
          </div>
        </div>
      </section>

      <!-- 商品详情 -->
      <section class="j8-pd-detail">
        <template v-if="product.content">
          <h2>{{ t('ji8.product.detailTitle') }}</h2>
          <div
            class="theme-prose j8-pd-prose"
            v-html="processHtmlForDisplay(getLocalizedText(product.content))"
          ></div>
        </template>
        <div class="j8-pd-notice" :class="{ 'is-first': !product.content }">
          <h2>{{ t('ji8.product.noticeTitle') }}</h2>
          <p>{{ t('ji8.product.noticeBody') }}</p>
        </div>
      </section>

      <!-- 相关文章 -->
      <section v-if="relatedPosts.length" class="j8-pd-detail">
        <h2>{{ t('productDetail.relatedPosts') }}</h2>
        <div class="j8-pd-posts">
          <router-link v-for="rp in relatedPosts" :key="rp.id" :to="`/blog/${rp.slug}`">
            <img v-if="rp.thumbnail" :src="getImageUrl(rp.thumbnail)" :alt="getLocalizedText(rp.title)" loading="lazy" />
            <b>{{ getLocalizedText(rp.title) }}</b>
            <small>{{ formatRelatedPostDate(rp.published_at) }}</small>
          </router-link>
        </div>
      </section>

      <ProductMobileBar
        :visible="showMobileBar && !!product && !loading"
        :requires-login="requiresLogin"
        :can-purchase="canPurchase"
        :show-member-price="mobileBarShowMemberPrice"
        :member-price-display="mobileBarMemberPriceDisplay"
        :show-sku-promotion-price="mobileBarShowSkuPromotionPrice"
        :sku-promotion-price-display="mobileBarSkuPromotionPriceDisplay"
        :show-sku-price="mobileBarShowSkuPrice"
        :sku-price-display="mobileBarSkuPriceDisplay"
        :show-product-promotion-price="mobileBarShowProductPromotionPrice"
        :product-promotion-price-display="mobileBarProductPromotionPriceDisplay"
        :product-price-display="mobileBarProductPriceDisplay"
        @add-to-cart="addToCart"
        @buy-now="buyNow"
        @go-login="goLogin"
      />
    </template>

    <EmptyState v-else size="lg" icon="alert" :title="t('productDetail.notFound')">
      <template #action>
        <Button class="rounded-full h-10" @click="loadProduct">
          <RotateCw />
          {{ t('errorBoundary.retry') }}
        </Button>
        <Button variant="secondary" as-child class="rounded-full h-10">
          <router-link to="/products">{{ t('productDetail.backToProducts') }}</router-link>
        </Button>
      </template>
    </EmptyState>
  </div>
</template>

<script setup lang="ts">
import { computed, onUnmounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { ChevronLeft, Crown, Minus, Pencil, Plus, RotateCw, Share2, ShoppingCart, Tag, Zap } from 'lucide-vue-next'
import { affiliateAPI } from '../../api'
import { useAppStore } from '../../stores/app'
import { useUserAuthStore } from '../../stores/userAuth'
import { useProductDetail } from '../../composables/useProductDetail'
import { toast } from '../../composables/useToast'
import { copyText } from '../../utils/clipboard'
import { getImageUrl } from '../../utils/image'
import { processHtmlForDisplay } from '../../utils/content'
import { amountToCents, centsToAmount } from '../../utils/money'
import ProductMobileBar from '../../components/product/ProductMobileBar.vue'
import EmptyState from '../../components/EmptyState.vue'
import { Button } from '@/components/ui/button'
import Ji8BrandIcon from './components/Ji8BrandIcon.vue'
import { currencySymbol, formatMoney } from './utils/price'
import './styles/product.css'

/**
 * ji8 商品详情（仿 ai.bahk.cn/product.html）：
 * 左大图（无图时显示品牌 logo）+ 右侧购买区（标签 / 价格 / 推广卡 / 规格卡 / 数量 / 应付金额），
 * 下方商品详情。业务逻辑全部来自 useProductDetail，与 classic / vault 一致。
 */
const { t } = useI18n()
const appStore = useAppStore()
const userAuthStore = useUserAuthStore()

// 移动端固定购买条：桌面购买区出屏后显示
const purchaseActionsRef = ref<HTMLElement | null>(null)
const showMobileBar = ref(false)
let observer: IntersectionObserver | null = null
const setupMobileBarObserver = () => {
  if (observer) observer.disconnect()
  if (!purchaseActionsRef.value) return
  observer = new IntersectionObserver(
    (entries) => {
      const entry = entries[0]
      if (entry) showMobileBar.value = !entry.isIntersecting
    },
    { threshold: 0.1 },
  )
  observer.observe(purchaseActionsRef.value)
}

const {
  getLocalizedText, siteCurrency,
  getPurchaseTypeLabel, getFulfillmentTypeLabel, getStockStatusLabel,
  hasPromotionPrice, getPromotionPriceAmount, hasSkuPromotionPrice,
  hasPromotionRules, getPromotionRules,
  formatPromotionRule, formatWholesaleTier, formatRelatedPostDate, normalizeSkuId,
  loading, product, relatedPosts, currentImage, selectedSkuId, quantity, purchaseWarning,
  activeSkus, selectedSku,
  selectedSkuMemberPrice, hasMemberPrice,
  hasSelectedSkuWholesalePrice, selectedSkuWholesaleFinalIsMember, selectedSkuWholesaleFinalPrice,
  selectedSkuWholesaleRules,
  selectedSkuPromotionPrice, selectedSkuPromotionFinalIsMember, selectedSkuPromotionFinalPrice,
  isSkuPurchasable, skuDisplayText, skuStockText,
  quantityEffectiveLimit, quantityEffectiveMin, handleQuantityInput,
  requiresLogin, requiresSKUSelection, canPurchase, cannotPurchaseReason,
  images,
  addToCart, buyNow, goLogin, loadProduct,
  mobileBarShowMemberPrice, mobileBarMemberPriceDisplay,
  mobileBarShowSkuPromotionPrice, mobileBarSkuPromotionPriceDisplay,
  mobileBarShowSkuPrice, mobileBarSkuPriceDisplay,
  mobileBarShowProductPromotionPrice, mobileBarProductPromotionPriceDisplay, mobileBarProductPriceDisplay,
} = useProductDetail({ onLoaded: () => setupMobileBarObserver() })

onUnmounted(() => {
  observer?.disconnect()
  observer = null
})

const title = computed(() => (product.value ? String(getLocalizedText(product.value.title) || '') : ''))
const description = computed(() => (product.value ? String(getLocalizedText(product.value.description) || '').trim() : ''))

const money = (amount: unknown) => formatMoney(amount, siteCurrency.value)
const plainAmount = (amount: unknown) => {
  const cents = amountToCents(amount)
  return cents === null ? String(amount ?? '-') : centsToAmount(cents)
}

type PriceTag = 'promo' | 'member' | 'wholesale' | null

/** 与 classic 价格区的分支一一对应：当前展示的成交单价、划线原价、标签 */
const priceInfo = computed<{ final: unknown; original: unknown | null; tag: PriceTag }>(() => {
  const p = product.value
  const sku = selectedSku.value
  if (!p) return { final: 0, original: null, tag: null }
  if (sku && hasSelectedSkuWholesalePrice.value) {
    return { final: selectedSkuWholesaleFinalPrice.value, original: sku.price_amount, tag: selectedSkuWholesaleFinalIsMember.value ? 'member' : 'wholesale' }
  }
  if (sku && hasSkuPromotionPrice(sku)) {
    return selectedSkuPromotionFinalIsMember.value
      ? { final: selectedSkuPromotionFinalPrice.value, original: sku.price_amount, tag: 'member' }
      : { final: selectedSkuPromotionPrice.value, original: sku.price_amount, tag: 'promo' }
  }
  if (sku && hasMemberPrice.value) {
    return { final: selectedSkuMemberPrice.value, original: sku.price_amount, tag: 'member' }
  }
  if (sku) return { final: sku.price_amount, original: null, tag: null }
  if (hasPromotionPrice(p)) return { final: getPromotionPriceAmount(p), original: p.price_amount, tag: 'promo' }
  return { final: p.price_amount, original: null, tag: null }
})

const priceTagText = computed(() => {
  const info = priceInfo.value
  const saved = (amountToCents(info.original) ?? 0) - (amountToCents(info.final) ?? 0)
  const label = info.tag === 'member'
    ? t('products.memberPriceTag')
    : info.tag === 'wholesale' ? t('products.wholesaleTag') : t('products.promotionTag')
  return saved > 0 ? `${label} · ${t('ji8.product.saved', { amount: money(centsToAmount(saved)) })}` : label
})

/** 预估应付 = 当前单价 × 数量（优惠券等在结算页计算） */
const totalAmount = computed(() => {
  const cents = amountToCents(priceInfo.value.final)
  if (cents === null) return priceInfo.value.final
  return centsToAmount(cents * Math.max(1, Number(quantity.value) || 1))
})

const promoCard = computed(() => {
  if (appStore.canAccessResellerConsole) return { to: '/reseller', title: t('ji8.product.promoReseller') }
  if (appStore.config?.affiliate?.enabled === true) return { to: '/me/affiliate', title: t('ji8.product.promoAffiliate') }
  return null
})

// 分享：登录且开通推广时自动带上自己的推广码
let cachedAffCode: string | null = null
const share = async () => {
  const url = new URL(window.location.href)
  url.search = ''
  if (userAuthStore.isAuthenticated && appStore.config?.affiliate?.enabled === true) {
    try {
      if (cachedAffCode === null) {
        const res = await affiliateAPI.dashboard()
        cachedAffCode = String(res.data.data?.affiliate_code || '')
      }
      if (cachedAffCode) url.searchParams.set('aff', cachedAffCode)
    } catch {
      /* 取推广码失败时分享普通链接 */
    }
  }
  try {
    await copyText(url.toString())
    toast.success(t(url.searchParams.has('aff') ? 'ji8.product.sharedWithAff' : 'ji8.product.shared'))
  } catch {
    toast.error(t('ji8.account.copyFailed'))
  }
}

// 移动端左右滑切图
let touchStartX = 0
const onTouchStart = (e: TouchEvent) => { touchStartX = e.touches[0]?.clientX ?? 0 }
const onTouchEnd = (e: TouchEvent) => {
  const dx = (e.changedTouches[0]?.clientX ?? 0) - touchStartX
  if (Math.abs(dx) < 40 || images.value.length < 2) return
  const idx = images.value.indexOf(currentImage.value)
  const next = (idx + (dx < 0 ? 1 : -1) + images.value.length) % images.value.length
  currentImage.value = images.value[next] ?? currentImage.value
}
</script>
