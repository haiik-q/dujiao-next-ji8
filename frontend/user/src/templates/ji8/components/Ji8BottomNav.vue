<template>
  <nav class="j8-bottom-nav" :style="{ gridTemplateColumns: `repeat(${bottomItems.length}, 1fr)` }">
    <RouterLink
      v-for="item in bottomItems"
      :key="item.key"
      :to="item.path"
      :class="{ 'is-active': isActive(item, route.path) }"
    >
      <component :is="item.icon" />
      <span v-if="item.key === 'cart' && cartCount > 0" class="j8-cart-count">{{ cartCount > 99 ? '99+' : cartCount }}</span>
      <span>{{ item.label }}</span>
    </RouterLink>
  </nav>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useCartStore } from '../../../stores/cart'
import { useJi8Nav } from '../composables/useJi8Nav'

/**
 * 移动端浮动底栏（≤760px 由 ji8.css 显示，桌面 display:none）。
 * 商品详情路由由 `.ji8-scope.is-product .j8-bottom-nav{display:none}` 隐藏（D14），
 * 让位给 classic 回退页自带的 ProductMobileBar。
 */
const route = useRoute()
const cartStore = useCartStore()
const { bottomItems, isActive } = useJi8Nav()

const cartCount = computed(() => cartStore.totalItems)
</script>
