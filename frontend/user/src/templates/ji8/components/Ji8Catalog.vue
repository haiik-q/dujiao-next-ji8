<template>
  <div class="j8-container">
    <section class="j8-hero" :class="{ 'is-home': variant === 'home' }">
      <div class="j8-hero-copy">
        <span class="j8-eyebrow">{{ heroEyebrow }}</span>
        <h1>{{ heroTitle }}</h1>
        <p>{{ heroLede }}</p>
      </div>
      <i18n-t v-if="variant === 'home'" keypath="ji8.hero.mobileLead" tag="p" class="j8-shortlink">
        <template #host><em>{{ siteHost }}</em></template>
      </i18n-t>
      <div class="j8-hero-actions">
        <div class="j8-search" role="search">
          <Search />
          <input
            v-model="searchQuery"
            type="search"
            :placeholder="t('products.searchBoxPlaceholder')"
            :aria-label="t('products.searchLabel')"
            @keydown.enter.prevent="onSearch"
          />
          <button v-if="searchQuery" type="button" class="j8-search-clear" :aria-label="t('blog.searchClear')" @click="clearSearch">
            <X />
          </button>
        </div>
        <button type="button" class="j8-share" @click="shareSite"><Share2 />{{ t('ji8.hero.share') }}</button>
      </div>
    </section>

    <button type="button" class="j8-cat-btn" @click="openDrawer">
      <ShoppingBag />
      <span>{{ currentCategoryName }}</span>
      <ChevronDown />
    </button>

    <section class="j8-catalog">
      <Ji8CategoryPanel
        :groups="categoryGroups"
        :selected="selectedCategory"
        :total="total"
        :counts="counts"
        :loading="loading"
        @select="selectCategory($event, true)"
      />
      <div>
        <div class="j8-section-heading">
          <div>
            <span class="j8-eyebrow">{{ sectionEyebrow }}</span>
            <h2>{{ currentCategoryName }}</h2>
          </div>
          <span v-if="!loading">{{ t('ji8.section.count', { count: total }) }}</span>
        </div>
        <div v-if="loading" class="j8-grid">
          <div v-for="i in 6" :key="i" class="j8-skeleton-card"></div>
        </div>
        <div v-else-if="products.length" class="j8-grid">
          <Ji8ProductCard v-for="p in products" :key="p.id" :product="p" />
        </div>
        <div v-else class="j8-empty">
          <component :is="hasFilter ? SearchX : PackageOpen" />
          <b>{{ hasFilter ? t('products.emptyFiltered') : t('ji8.empty.title') }}</b>
          <span>{{ t('ji8.empty.hint') }}</span>
          <button v-if="hasFilter" type="button" class="j8-btn-primary" @click="resetFilters">{{ t('products.clearFilters') }}</button>
        </div>
        <PaginationNav
          :current-page="currentPage"
          :total-pages="totalPages"
          :loading="loading"
          :scroll-top="false"
          compact
          @change-page="changePage"
        />
      </div>
    </section>

    <Ji8CategoryDrawer
      v-model:open="showFilterDrawer"
      :groups="categoryGroups"
      :selected="selectedCategory"
      :total="total"
      :counts="counts"
      @select="selectCategory($event, true)"
    />

    <AnnouncementModal
      v-if="variant === 'home' && activeAnnouncement"
      :announcement="activeAnnouncement"
      :visible="announcementVisible"
      @update:visible="announcementVisible = $event"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ChevronDown, PackageOpen, Search, SearchX, Share2, ShoppingBag, X } from 'lucide-vue-next'
import { useAppStore } from '../../../stores/app'
import { useProductList } from '../../../composables/useProductList'
import { useLocalized } from '../../../composables/useProduct'
import { useAnnouncement, type HomeAnnouncement } from '../../../composables/useAnnouncement'
import { usePageSeo } from '../../../composables/usePageSeo'
import { useToast } from '../../../composables/useToast'
import { copyText } from '../../../utils/clipboard'
import PaginationNav from '../../../components/PaginationNav.vue'
import AnnouncementModal from '../../../components/AnnouncementModal.vue'
import { useCategoryCounts } from '../composables/useCategoryCounts'
import Ji8CategoryPanel from './Ji8CategoryPanel.vue'
import Ji8CategoryDrawer from './Ji8CategoryDrawer.vue'
import Ji8ProductCard from './Ji8ProductCard.vue'

const props = defineProps<{
  variant: 'home' | 'store'
  homeRouteName: 'home' | 'products'
}>()

const route = useRoute()
const { t } = useI18n()
const appStore = useAppStore()
const { getLocalizedText } = useLocalized()
const { shouldShow } = useAnnouncement()
const { toast } = useToast()
const { counts, load } = useCategoryCounts()

const {
  loading,
  products,
  categoryGroups,
  categoryMap,
  selectedCategory,
  searchQuery,
  currentPage,
  totalPages,
  total,
  showFilterDrawer,
  selectCategory,
  changePage,
  clearSearch,
  onSearch,
  initialize,
  cleanup,
} = useProductList({ pageSize: 24, homeRouteName: props.homeRouteName })

// ==================== 品牌 / 文案 ====================
const siteName = computed(() => String(appStore.config?.brand?.site_name || '').trim())
const siteHost = computed(() => {
  try {
    return new URL(String(appStore.config?.brand?.site_url || '').trim()).host || window.location.host
  } catch {
    return window.location.host
  }
})

const heroEyebrow = computed(() => (props.variant === 'home' ? t('ji8.hero.eyebrow') : t('ji8.store.eyebrow')))
const heroTitle = computed(() =>
  props.variant === 'home' ? t('ji8.hero.title', { site: siteName.value || siteHost.value }) : t('ji8.store.title'),
)
const heroLede = computed(() => (props.variant === 'home' ? t('ji8.hero.subtitle') : t('products.subtitle')))
const sectionEyebrow = computed(() => (props.variant === 'home' ? t('ji8.section.eyebrow') : t('ji8.store.sectionEyebrow')))

const selectedCategoryName = computed(() => {
  if (!selectedCategory.value) return ''
  const category = categoryMap.value.get(selectedCategory.value)
  return category ? getLocalizedText(category.name) : ''
})
const currentCategoryName = computed(() => selectedCategoryName.value || t('products.allCategories'))
const hasFilter = computed(() => selectedCategory.value !== null || searchQuery.value.trim() !== '')

// ==================== SEO ====================
usePageSeo({
  canonicalPath: () => route.path,
  title: () => {
    if (route.name === 'category-products') return selectedCategoryName.value || t('nav.products')
    if (route.name === 'products') return t('nav.products')
    return undefined
  },
})

// ==================== 公告 ====================
const activeAnnouncement = ref<HomeAnnouncement | null>(null)
const announcementVisible = ref(false)

const showAnnouncementIfNeeded = () => {
  const announcement = appStore.config?.announcement as HomeAnnouncement | undefined
  if (announcement && shouldShow(announcement)) {
    activeAnnouncement.value = announcement
    announcementVisible.value = true
  }
}

// ==================== 交互 ====================
const openDrawer = () => {
  showFilterDrawer.value = true
}

const resetFilters = () => {
  clearSearch()
  selectCategory(null)
}

const shareSite = async () => {
  const url = String(appStore.config?.brand?.site_url || '').trim() || window.location.origin
  if (typeof navigator.share === 'function') {
    try {
      await navigator.share({ title: siteName.value || siteHost.value, url })
    } catch {
      // 用户取消或系统拒绝分享时不再回落
    }
    return
  }
  try {
    await copyText(url)
    toast.success(t('ji8.hero.shareCopied'))
  } catch (error) {
    console.error('Failed to copy site link:', error)
  }
}

// ==================== 生命周期 ====================
// config 已由 router.beforeEach 保证加载；这里不再调用 loadConfig()，否则每次切换路由都会重复执行站长自定义脚本
onMounted(async () => {
  await initialize()
  void load(categoryGroups.value)
  if (props.variant === 'home') showAnnouncementIfNeeded()
})

onUnmounted(cleanup)
</script>
