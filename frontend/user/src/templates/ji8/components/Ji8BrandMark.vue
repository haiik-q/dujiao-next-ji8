<template>
  <span class="j8-brand-mark" :class="{ small }">
    <img
      v-if="logo && !errored"
      :src="logo"
      :alt="siteName"
      width="200"
      height="200"
      @error="errored = true"
    />
    <b v-else>{{ initial }}</b>
  </span>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useAppStore } from '../../../stores/app'
import { getImageUrl } from '../../../utils/image'

/**
 * 品牌标记（顶栏 / 页脚共用）：优先显示后台 site_logo，缺省或加载失败时显示站名首字。
 * `small` 用于页脚的 32px 版本（样式见 ji8.css `.j8-brand-mark.small`）。
 */
defineProps<{ small?: boolean }>()

const appStore = useAppStore()
const errored = ref(false)

const siteName = computed(() => String(appStore.config?.brand?.site_name || '').trim())
const logo = computed(() => {
  const raw = String(appStore.config?.brand?.site_logo || '').trim()
  return raw ? getImageUrl(raw) : ''
})
const initial = computed(() => siteName.value.charAt(0).toUpperCase() || 'A')

// logo 地址变化（配置重新加载）时重置失败标记，重新尝试加载
watch(logo, () => {
  errored.value = false
})
</script>
