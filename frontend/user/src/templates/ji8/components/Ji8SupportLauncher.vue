<template>
  <div v-if="wechatId" ref="rootRef" class="j8-wechat-launcher">
    <div
      v-if="open"
      :id="panelId"
      class="j8-wechat-panel"
      role="dialog"
      :aria-label="t('ji8.wechat.title')"
    >
      <header>
        <span class="j8-wechat-badge"><svg viewBox="0 0 24 24" aria-hidden="true"><path :d="WECHAT_ICON_PATH" /></svg></span>
        <strong>{{ t('ji8.wechat.title') }}</strong>
        <button type="button" class="j8-wechat-close" :aria-label="t('ji8.wechat.close')" @click="open = false"><X /></button>
      </header>
      <div class="j8-wechat-id">
        <span>{{ t('ji8.wechat.idLabel') }}</span>
        <code>{{ wechatId }}</code>
        <button type="button" @click="copyId">{{ copied ? t('ji8.wechat.copied') : t('ji8.wechat.copy') }}</button>
      </div>
      <p>{{ t('ji8.wechat.tip') }}</p>
      <a v-if="supportLink" class="j8-wechat-other" :href="supportLink" target="_blank" rel="noopener noreferrer">
        <Headset /> {{ t('ji8.wechat.otherSupport') }}
      </a>
    </div>
    <button
      type="button"
      class="j8-support j8-support--wechat"
      :aria-expanded="open"
      :aria-controls="panelId"
      :aria-label="t('ji8.wechat.button')"
      @click="open = !open"
    >
      <svg viewBox="0 0 24 24" aria-hidden="true"><path :d="WECHAT_ICON_PATH" /></svg>
      <span>{{ t('ji8.wechat.button') }}</span>
    </button>
  </div>
  <a
    v-else-if="supportLink"
    class="j8-support"
    :href="supportLink"
    target="_blank"
    rel="noopener noreferrer"
    :aria-label="t('ji8.nav.support')"
  >
    <Headset />
    <span>{{ t('ji8.nav.support') }}</span>
  </a>
</template>

<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'
import { Headset, X } from 'lucide-vue-next'
import { useJi8Nav } from '../composables/useJi8Nav'
import { JI8_WECHAT_ID, WECHAT_ICON_PATH } from '../utils/contact'
import { copyText } from '../../../utils/clipboard'
import { toast } from '../../../composables/useToast'

/**
 * 右下角客服浮标：配置了微信号 → 微信按钮，点开卡片（微信号 + 复制，另有 telegram/whatsapp 时附链接）；
 * 未配置微信号 → 回落为 contact.telegram || contact.whatsapp 外链浮标，二者皆空则不渲染（D6）。
 */
const { t } = useI18n()
const route = useRoute()
const { supportLink } = useJi8Nav()

const wechatId = JI8_WECHAT_ID.trim()
const panelId = 'j8-wechat-panel'
const open = ref(false)
const copied = ref(false)
const rootRef = ref<HTMLElement | null>(null)
let copiedTimer: number | undefined

const copyId = async () => {
  try {
    await copyText(wechatId)
    copied.value = true
    toast.success(t('ji8.wechat.copied'))
    window.clearTimeout(copiedTimer)
    copiedTimer = window.setTimeout(() => (copied.value = false), 2000)
  } catch {
    toast.error(t('ji8.wechat.copyFailed'))
  }
}

// 点外部 / Esc / 换页 收起卡片
const onPointerDown = (event: PointerEvent) => {
  if (open.value && rootRef.value && !rootRef.value.contains(event.target as Node)) open.value = false
}
const onKeydown = (event: KeyboardEvent) => {
  if (event.key === 'Escape') open.value = false
}
watch(() => route.fullPath, () => (open.value = false))
onMounted(() => {
  document.addEventListener('pointerdown', onPointerDown)
  document.addEventListener('keydown', onKeydown)
})
onBeforeUnmount(() => {
  document.removeEventListener('pointerdown', onPointerDown)
  document.removeEventListener('keydown', onKeydown)
  window.clearTimeout(copiedTimer)
})
</script>
