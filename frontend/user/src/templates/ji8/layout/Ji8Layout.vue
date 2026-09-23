<template>
  <div class="ji8-scope" :class="{ 'is-product': route.name === 'product-detail' }">
    <Ji8Header />
    <main class="j8-main"><slot /></main>
    <Ji8Footer />
    <Ji8SupportLauncher />
    <BackToTop />
    <Ji8BottomNav />
  </div>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useHead } from '@unhead/vue'
import '../styles/ji8.css'
import Ji8Header from '../components/Ji8Header.vue'
import Ji8Footer from '../components/Ji8Footer.vue'
import Ji8BottomNav from '../components/Ji8BottomNav.vue'
import Ji8SupportLauncher from '../components/Ji8SupportLauncher.vue'
import BackToTop from '../../../components/BackToTop.vue'

/**
 * ji8 模板外壳（仿 ai.bahk.cn）。
 * - 根节点 `.ji8-scope` 承载全部令牌与 `.j8-*` 样式；商品详情路由（classic 回退页自带
 *   ProductMobileBar）加 `is-product`，由 ji8.css 隐藏浮动底栏并上移客服浮标 / 回顶按钮（D14）。
 * - Teleport 到 body 的浮层（Toast / ConfirmDialog / AnnouncementModal）在 `.ji8-scope` 之外，
 *   靠 body 上的 `ji8-tokens` class 拿到同一套令牌，见 ji8.css 顶部选择器。
 */
const route = useRoute()

useHead({
  meta: [
    { name: 'theme-color', content: '#f7f9fc', media: '(prefers-color-scheme: light)' },
    { name: 'theme-color', content: '#0f1620', media: '(prefers-color-scheme: dark)' },
  ],
})

// 单页应用切换路由时 DOM 被替换但选区不会重置：开启光标浏览(F7)或刚点过文字时，
// 残留光标会落到新页面第一段文字（如商品页的「返回购物」）上闪烁。换页后清掉即可。
watch(
  () => route.path,
  () => {
    const selection = window.getSelection()
    if (selection && selection.type !== 'Range') selection.removeAllRanges()
  },
  { flush: 'post' },
)

onMounted(() => document.body.classList.add('ji8-tokens'))
onUnmounted(() => document.body.classList.remove('ji8-tokens'))
</script>
