<template>
  <section v-if="groups.length >= 2" ref="root" class="j8-featured" :data-count="groups.length">
    <div class="j8-section-heading">
      <div>
        <span class="j8-eyebrow">{{ t('ji8.featured.eyebrow') }}</span>
        <h2>{{ t('ji8.featured.title') }}</h2>
      </div>
    </div>
    <div class="j8-featured-grid">
      <RouterLink
        v-for="(g, i) in groups"
        :key="g.art"
        :to="linkOf(g)"
        class="j8-fcard j8-reveal"
        :class="[`is-${g.art}`, { 'is-wide-lg': i === 0 && groups.length % 3 === 2, 'is-wide-md': i === 0 && groups.length % 2 === 1, 'is-sold-out': g.allSoldOut }]"
      >
        <div class="j8-fcard-head">
          <div>
            <h3>{{ t(`ji8.featured.name.${g.art}`) }}<em v-if="highlightOf(g)">{{ highlightOf(g) }}</em></h3>
            <p>{{ t(`ji8.featured.sub.${g.art}`) }}</p>
          </div>
          <span v-if="!highlightOf(g)" class="j8-fcard-tag">{{ t('ji8.featured.count', { count: g.products.length }) }}</span>
        </div>

        <div class="j8-fart" :class="`is-${g.art}`" aria-hidden="true">
          <!-- ChatGPT：套餐卡叠放，悬停时散开 -->
          <div v-if="g.art === 'chatgpt'" class="fa-gpt">
            <div v-for="(plan, k) in [...g.plans].reverse()" :key="plan.name + k" class="fa-gpt-card" :class="`c${g.plans.length - k}`">
              <small>ChatGPT</small>
              <b>{{ plan.name }}</b>
              <em>{{ money(plan.price) }}</em>
            </div>
          </div>

          <!-- Gemini：渐变四角星 + 轨道 + 闪烁 -->
          <div v-else-if="g.art === 'gemini'" class="fa-gem">
            <span class="fa-glow"></span>
            <span class="fa-gem-orbit"><i></i></span>
            <svg class="fa-gem-star" viewBox="0 0 100 100">
              <defs>
                <linearGradient id="j8-gem-grad" x1="0" y1="0" x2="1" y2="1">
                  <stop offset="0" stop-color="#4f8bff" />
                  <stop offset=".55" stop-color="#8d6bff" />
                  <stop offset="1" stop-color="#e46fbf" />
                </linearGradient>
              </defs>
              <path :d="STAR" fill="url(#j8-gem-grad)" />
            </svg>
            <svg v-for="n in 3" :key="n" class="fa-twinkle" :class="`t${n}`" viewBox="0 0 100 100"><path :d="STAR" /></svg>
            <div v-if="bigOf(g)" class="fa-num"><b><span class="fa-count" :style="{ '--to': bigOf(g)!.n }"></span><small>{{ bigOf(g)!.unit }}</small></b><span>{{ t('ji8.featured.caption.gemini') }}</span></div>
          </div>

          <!-- X：黑色会员卡 + 扫光 -->
          <div v-else-if="g.art === 'x'" class="fa-x">
            <div class="fa-x-card">
              <div class="fa-x-logo">
                <svg v-if="xIcon" :viewBox="xIcon.viewBox" fill="currentColor"><path :d="xIcon.path" /></svg>
                <b>Premium</b>
              </div>
              <div class="fa-x-meta">
                <small>MEMBERSHIP</small>
                <em v-if="g.months.length">{{ g.months.join(' / ') }} MONTHS</em>
              </div>
              <i class="fa-x-sheen"></i>
            </div>
          </div>

          <!-- 多邻国：图标 + 漂浮的多语问候 -->
          <div v-else-if="g.art === 'duolingo'" class="fa-duo">
            <span class="fa-glow"></span>
            <span class="fa-icon"><Ji8BrandIcon :slug="g.category?.slug" :name="nameOf(g)" :image="iconOf(g)" /></span>
            <span v-for="(word, k) in WORDS" :key="word" class="fa-word" :class="`w${k + 1}`">{{ word }}</span>
            <div v-if="bigOf(g)" class="fa-num"><b><span class="fa-count" :style="{ '--to': bigOf(g)!.n }"></span><small>{{ bigOf(g)!.unit }}</small></b><span>{{ t('ji8.featured.caption.duolingo') }}</span></div>
          </div>

          <!-- Muse：吉祥物 + 绕圈的助力头像 + 冒出的词元气泡 + 词元计数 -->
          <div v-else class="fa-muse">
            <span class="fa-glow"></span>
            <span class="fa-muse-ring"><i v-for="n in 4" :key="n" :class="`a${n}`"><UserRound /></i></span>
            <svg v-for="n in 2" :key="`s${n}`" class="fa-twinkle" :class="`t${n}`" viewBox="0 0 100 100"><path :d="STAR" /></svg>
            <span class="fa-icon"><Ji8BrandIcon :slug="g.category?.slug" :name="nameOf(g)" :image="iconOf(g)" /></span>
            <span v-for="n in 3" :key="n" class="fa-chip" :class="`c${n}`">{{ t('ji8.featured.caption.museChip') }}</span>
            <div v-if="g.tokens" class="fa-num"><b><span class="fa-count" :style="{ '--to': tokenNum(g) }"></span><small>{{ t('ji8.featured.unit.token') }}</small></b><span>{{ t('ji8.featured.caption.muse') }}</span></div>
          </div>
        </div>

        <div class="j8-fcard-foot">
          <div class="j8-fcard-price">
            <strong>{{ money(g.minPrice) }}</strong>
            <small v-if="g.priceVaries">{{ t('ji8.featured.from') }}</small>
          </div>
          <span class="j8-fcard-cta">{{ ctaOf(g) }}<ChevronRight /></span>
        </div>
      </RouterLink>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { RouterLink } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ChevronRight, UserRound } from 'lucide-vue-next'
import { useLocalized, useProductLabels } from '../../../composables/useProduct'
import { getImageUrl } from '../../../utils/image'
import { formatMoney } from '../utils/price'
import { matchBrandIcon } from '../utils/brandIcons'
import { buildFeaturedGroups, type FeaturedGroup } from '../utils/featured'
import { useReveal } from '../composables/useReveal'
import Ji8BrandIcon from './Ji8BrandIcon.vue'
import '../styles/featured.css'

const props = defineProps<{ products: any[] }>()

const { t, locale } = useI18n()
const { getLocalizedText, siteCurrency } = useLocalized()
const { isSoldOut } = useProductLabels()

// 四角星（Gemini 主图与闪烁小星共用）
const STAR = 'M50 2C54 30 70 46 98 50C70 54 54 70 50 98C46 70 30 54 2 50C30 46 46 30 50 2Z'
const WORDS = ['Hello!', 'Bonjour', 'Hola', 'こんにちは']
const xIcon = matchBrandIcon('twitter-x')

const groups = computed(() =>
  buildFeaturedGroups(props.products || [], (p) => getLocalizedText(p?.title), (p) => isSoldOut(p)),
)

const money = (amount: number) => formatMoney(amount, siteCurrency.value).replace(/\.00$/, '')
const nameOf = (g: FeaturedGroup) => getLocalizedText(g.category?.name) || g.art
const iconOf = (g: FeaturedGroup) => getImageUrl(g.category?.icon)

// 词元数：中文按「亿」显示，英文换算成 B（10 亿 = 1B）
const tokenNum = (g: FeaturedGroup) => (String(locale.value).startsWith('en') ? Math.round(g.tokens / 10) : g.tokens)

// 标题里高亮的部分：时长（12 的倍数写成「N 年」）、ChatGPT 套餐档位或 Muse 最高词元；没有就不高亮，改在右上角显示款数
const highlightOf = (g: FeaturedGroup) => {
  if (g.art === 'chatgpt' && g.plans.length) return [...new Set(g.plans.map((p) => p.name.split(' ')[0]))].join(' / ')
  if (g.art === 'muse' && g.tokens) return t('ji8.featured.tokens', { n: tokenNum(g) })
  if (!g.months.length) return ''
  if (g.months.every((m) => m % 12 === 0)) {
    const years = g.months.map((m) => m / 12)
    return t('ji8.featured.years', { n: years.join(' / ') }, years[years.length - 1] ?? 1)
  }
  return t('ji8.featured.months', { n: g.months.join(' / ') })
}

// 插画里的大数字：12 个月写成 365 天，其他 12 的倍数写成年
const bigOf = (g: FeaturedGroup) => {
  const m = g.months[g.months.length - 1]
  if (!m) return null
  if (m === 12) return { n: 365, unit: t('ji8.featured.unit.day') }
  if (m % 12 === 0) return { n: m / 12, unit: t('ji8.featured.unit.year') }
  return { n: m, unit: t('ji8.featured.unit.month') }
}

const linkOf = (g: FeaturedGroup) =>
  g.products.length === 1 ? `/products/${g.products[0].slug}` : `/categories/${g.category?.slug}`

const ctaOf = (g: FeaturedGroup) => {
  if (g.allSoldOut) return t('ji8.featured.soldOut')
  return g.products.length === 1 ? t('ji8.featured.viewProduct') : t('ji8.featured.choosePlan')
}

// 进入视口时依次浮现
const root = ref<HTMLElement | null>(null)
const { scan } = useReveal(root, '.j8-reveal')
watch(groups, () => void scan(), { immediate: true, flush: 'post' })
</script>
