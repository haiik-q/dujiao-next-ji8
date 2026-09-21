<template>
  <header class="j8-header">
    <RouterLink class="j8-brand" to="/" :title="siteName">
      <Ji8BrandMark />
      <span class="j8-brand-text"><strong>{{ siteName }}</strong><small>{{ tagline }}</small></span>
    </RouterLink>

    <nav class="j8-nav">
      <RouterLink
        v-for="item in promoItems"
        :key="item.key"
        :to="item.path"
        :class="toneClass[item.tone ?? 'promo']"
      ><component :is="item.icon" />{{ item.label }}</RouterLink>
      <template v-for="item in mainItems" :key="item.key">
        <RouterLink
          v-if="item.type === 'route'"
          :to="item.path"
          :class="{ 'is-active': isActive(item, route.path) }"
        ><component :is="item.icon" />{{ item.label }}</RouterLink>
        <a
          v-else
          :href="item.path"
          :target="item.target"
          rel="noopener noreferrer"
        ><component :is="item.icon" />{{ item.label }}</a>
      </template>
    </nav>

    <div class="j8-actions">
      <!-- 移动端：只放 1 枚促销 pill（D16） -->
      <RouterLink
        v-if="mobilePromo"
        :to="mobilePromo.path"
        :class="['j8-growth', 'j8-only-mobile', { 'is-promo': mobilePromo.tone === 'promo' }]"
      >
        <component :is="mobilePromo.icon" /><span>{{ mobilePromo.label }}</span>
      </RouterLink>
      <RouterLink v-if="noticeEnabled" to="/notice" class="j8-icon-btn j8-only-mobile" :aria-label="t('nav.notice')">
        <Megaphone />
      </RouterLink>
      <!-- 移动端：2FA 动态验证码入口（桌面走主导航） -->
      <RouterLink to="/2fa" class="j8-icon-btn j8-only-mobile" :aria-label="t('ji8.nav.totp')" :title="t('ji8.nav.totp')">
        <KeyRound />
      </RouterLink>
      <RouterLink to="/cart" class="j8-icon-btn" :aria-label="t('navbar.cart')">
        <ShoppingCart />
        <span v-if="cartCount > 0" class="j8-cart-count">{{ cartCount > 99 ? '99+' : cartCount }}</span>
      </RouterLink>

      <!-- 语言切换（桌面） -->
      <div ref="langEl" class="relative j8-only-desktop">
        <button type="button" class="j8-icon-btn" :aria-label="t('navbar.selectLanguage')" @click="langOpen = !langOpen">
          <Languages />
        </button>
        <div v-if="langOpen" class="j8-popover">
          <button
            v-for="lang in languages"
            :key="lang.code"
            type="button"
            :class="{ 'is-active': appStore.locale === lang.code }"
            @click="changeLanguage(lang.code)"
          >{{ lang.name }}</button>
        </div>
      </div>

      <!-- 桌面：单按钮，未登录=进入登录，已登录=退出（参考站 .desktop-auth-entry.is-authenticated） -->
      <button
        v-if="userAuthStore.isAuthenticated"
        type="button"
        class="j8-icon-btn is-danger j8-only-desktop"
        :aria-label="t('navbar.logout')"
        :title="t('navbar.logout')"
        @click="userAuthStore.logout()"
      ><LogOut /></button>
      <RouterLink
        v-else
        to="/auth/login"
        class="j8-icon-btn j8-only-desktop"
        :aria-label="t('navbar.login')"
        :title="t('navbar.login')"
      ><User /></RouterLink>

      <!-- 移动：账号（登录后进个人中心，否则去登录） -->
      <RouterLink :to="accountPath" class="j8-icon-btn j8-only-mobile" :aria-label="t('navbar.personalCenter')">
        <User />
      </RouterLink>
    </div>
  </header>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ShoppingCart, User, LogOut, Languages, Megaphone, KeyRound } from 'lucide-vue-next'
import { useAppStore } from '../../../stores/app'
import { useCartStore } from '../../../stores/cart'
import { useUserAuthStore } from '../../../stores/userAuth'
import { useJi8Nav, type Ji8NavItem } from '../composables/useJi8Nav'
import Ji8BrandMark from './Ji8BrandMark.vue'

const { t } = useI18n()
const route = useRoute()
const appStore = useAppStore()
const cartStore = useCartStore()
const userAuthStore = useUserAuthStore()
const { mainItems, promoItems, accountPath, isActive, noticeEnabled } = useJi8Nav()

const siteName = computed(() => String(appStore.config?.brand?.site_name || '').trim())

/** 顶栏副标题：站点 URL 的 host 大写（如 JI8.AI），缺省用当前域名（D9 / G11） */
const tagline = computed(() => {
  const raw = String(appStore.config?.brand?.site_url || '').trim()
  let host = ''
  try {
    host = raw ? new URL(raw).host : ''
  } catch {
    host = ''
  }
  return (host || window.location.host).toUpperCase()
})

/** 促销项语义 → 样式类（G4：growth 走带边框 pill，gold 只改字色） */
const toneClass: Record<NonNullable<Ji8NavItem['tone']>, string> = {
  promo: 'j8-nav-promo',
  growth: 'j8-growth',
  gold: 'j8-nav-gold',
}

const mobilePromo = computed<Ji8NavItem | null>(() => promoItems.value[0] ?? null)

const cartCount = computed(() => cartStore.totalItems)

const languages = [
  { code: 'zh-CN', name: '简体中文' },
  { code: 'zh-TW', name: '繁體中文' },
  { code: 'en-US', name: 'English' },
]

const langOpen = ref(false)
const langEl = ref<HTMLElement | null>(null)

const changeLanguage = (code: string) => {
  appStore.setLocale(code)
  langOpen.value = false
}

const onDocClick = (e: MouseEvent) => {
  const target = e.target as Node
  if (langOpen.value && langEl.value && !langEl.value.contains(target)) langOpen.value = false
}

onMounted(() => document.addEventListener('click', onDocClick))
onUnmounted(() => document.removeEventListener('click', onDocClick))
</script>
