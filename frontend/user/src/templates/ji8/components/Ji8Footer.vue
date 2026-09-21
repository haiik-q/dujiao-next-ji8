<template>
  <footer class="j8-footer">
    <div>
      <Ji8BrandMark small />
      <p>
        <b>{{ siteName }}</b>
        <small>{{ description || `© ${year} ${siteName} · ${t('footer.rights')}` }}</small>
      </p>
    </div>
    <nav>
      <RouterLink :to="ordersPath">{{ t('ji8.nav.orders') }}</RouterLink>
      <RouterLink v-if="blogEnabled" to="/blog">{{ t('ji8.nav.tutorials') }}</RouterLink>
      <RouterLink v-if="noticeEnabled" to="/notice">{{ t('nav.notice') }}</RouterLink>
      <RouterLink v-if="aboutEnabled" to="/about">{{ t('nav.about') }}</RouterLink>
      <a
        v-for="(l, i) in footerLinks"
        :key="`fl-${i}`"
        :href="l.url || 'javascript:void(0)'"
        :target="l.url ? '_blank' : undefined"
        rel="noopener noreferrer"
      >{{ l.name }}</a>
      <RouterLink to="/terms">{{ t('footer.terms') }}</RouterLink>
      <RouterLink to="/privacy">{{ t('footer.privacy') }}</RouterLink>
      <a v-if="supportLink" :href="supportLink" target="_blank" rel="noopener noreferrer">{{ t('ji8.nav.support') }}</a>
    </nav>
  </footer>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '../../../stores/app'
import { getLocalizedText } from '../../../utils/resellerSiteConfig'
import { useJi8Nav } from '../composables/useJi8Nav'
import Ji8BrandMark from './Ji8BrandMark.vue'

const { t } = useI18n()
const appStore = useAppStore()
const { ordersPath, blogEnabled, noticeEnabled, aboutEnabled, supportLink } = useJi8Nav()

const year = new Date().getFullYear()

const siteName = computed(() => String(appStore.config?.brand?.site_name || '').trim())

/** 站点描述（多语言对象或纯字符串均兼容）；为空时页脚显示版权行 */
const description = computed(() => {
  const desc = appStore.config?.brand?.site_description
  if (typeof desc === 'string') return desc.trim()
  if (desc && typeof desc === 'object') return getLocalizedText(desc as Record<string, unknown>, appStore.locale)
  return ''
})

/** 后台配置的自定义页脚链接（与 VaultLayout 同一解析逻辑）；key 用下标，url 可空/重复（G15） */
const footerLinks = computed(() => {
  const links = appStore.config?.footer_links
  if (!Array.isArray(links)) return []
  return links
    .map((item: { name?: unknown; url?: unknown }) => ({
      name: typeof item?.name === 'string' ? item.name.trim() : getLocalizedText(item?.name as Record<string, string>, appStore.locale),
      url: String(item?.url || '').trim(),
    }))
    .filter((item) => item.name)
})
</script>
