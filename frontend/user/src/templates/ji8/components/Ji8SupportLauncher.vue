<template>
  <div v-if="hasChannels" ref="rootRef" class="j8-contact-launcher">
    <div
      v-if="open"
      :id="panelId"
      class="j8-contact-panel"
      role="dialog"
      :aria-label="t('ji8.contact.title')"
    >
      <header>
        <span class="j8-contact-badge"><Headset /></span>
        <strong>{{ t('ji8.contact.title') }}</strong>
        <button type="button" class="j8-contact-close" :aria-label="t('ji8.contact.close')" @click="open = false"><X /></button>
      </header>
      <ul class="j8-contact-list">
        <li v-if="channels.telegramUser">
          <span class="j8-contact-icon is-telegram"><svg viewBox="0 0 24 24" aria-hidden="true"><path :d="TELEGRAM_ICON_PATH" /></svg></span>
          <div class="j8-contact-text">
            <span>{{ t('ji8.contact.telegramUser') }}</span>
            <code>@{{ channels.telegramUser }}</code>
          </div>
          <a :href="`https://t.me/${channels.telegramUser}`" target="_blank" rel="noopener noreferrer">{{ t('ji8.contact.chat') }}</a>
        </li>
        <li v-if="channels.telegramGroup">
          <span class="j8-contact-icon is-telegram"><svg viewBox="0 0 24 24" aria-hidden="true"><path :d="TELEGRAM_ICON_PATH" /></svg></span>
          <div class="j8-contact-text">
            <span>{{ t('ji8.contact.telegramGroup') }}</span>
            <code class="is-link">{{ channels.telegramGroup.replace(/^https?:\/\//, '') }}</code>
          </div>
          <a :href="channels.telegramGroup" target="_blank" rel="noopener noreferrer">{{ t('ji8.contact.join') }}</a>
        </li>
        <li v-if="channels.qqGroup">
          <span class="j8-contact-icon is-qq"><svg viewBox="0 0 24 24" aria-hidden="true"><path :d="QQ_ICON_PATH" /></svg></span>
          <div class="j8-contact-text">
            <span>{{ t('ji8.contact.qqGroup') }}</span>
            <code>{{ channels.qqGroup }}</code>
          </div>
          <button type="button" @click="copyValue('qqGroup')">{{ copiedKey === 'qqGroup' ? t('ji8.contact.copied') : t('ji8.contact.copy') }}</button>
        </li>
        <li v-if="channels.wechat">
          <span class="j8-contact-icon is-wechat"><MessageCircle /></span>
          <div class="j8-contact-text">
            <span>{{ t('ji8.contact.wechat') }}</span>
            <code>{{ channels.wechat }}</code>
          </div>
          <button type="button" @click="copyValue('wechat')">{{ copiedKey === 'wechat' ? t('ji8.contact.copied') : t('ji8.contact.copy') }}</button>
        </li>
        <li v-if="channels.qq">
          <span class="j8-contact-icon is-qq"><svg viewBox="0 0 24 24" aria-hidden="true"><path :d="QQ_ICON_PATH" /></svg></span>
          <div class="j8-contact-text">
            <span>{{ t('ji8.contact.qq') }}</span>
            <code>{{ channels.qq }}</code>
          </div>
          <button type="button" @click="copyValue('qq')">{{ copiedKey === 'qq' ? t('ji8.contact.copied') : t('ji8.contact.copy') }}</button>
        </li>
      </ul>
      <p>{{ t('ji8.contact.tip') }}</p>
      <a v-if="supportLink" class="j8-contact-other" :href="supportLink" target="_blank" rel="noopener noreferrer">
        <Headset /> {{ t('ji8.contact.otherSupport') }}
      </a>
    </div>
    <button
      type="button"
      class="j8-support j8-support--contact"
      :aria-expanded="open"
      :aria-controls="panelId"
      :aria-label="t('ji8.contact.button')"
      @click="open = !open"
    >
      <Headset />
      <span>{{ t('ji8.contact.button') }}</span>
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
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'
import { Headset, MessageCircle, X } from 'lucide-vue-next'
import { useAppStore } from '../../../stores/app'
import { useJi8Nav } from '../composables/useJi8Nav'
import { JI8_SUPPORT, QQ_ICON_PATH, TELEGRAM_ICON_PATH } from '../utils/contact'
import { copyText } from '../../../utils/clipboard'
import { toast } from '../../../composables/useToast'

/**
 * 右下角客服浮标：主站 → 「联系客服」按钮，点开卡片（Telegram 客服 / Telegram 群 / QQ 群，另有 telegram/whatsapp 时附链接）；
 * 分销站填了微信号 / QQ → 同样的卡片（只复制），其余渠道走卡片底部链接；
 * 渠道全空 → 回落为 contact 外链浮标（分销站为分销商自己的客服），皆空则不渲染（D6）。
 */
const { t } = useI18n()
const route = useRoute()
const appStore = useAppStore()
const { supportLink } = useJi8Nav()

// Telegram 客服 / 群 / QQ 群是主站的：分销站不显示，避免分销商的客人找到主站客服；
// 分销站只显示分销商在站点设置里填的微信号 / QQ
const channels = computed(() => {
  if (appStore.isResellerTenant) {
    const contact = (appStore.config?.contact || {}) as Record<string, unknown>
    return {
      telegramUser: '',
      telegramGroup: '',
      qqGroup: '',
      wechat: String(contact.wechat || '').trim(),
      qq: String(contact.qq || '').trim(),
    }
  }
  return {
    telegramUser: JI8_SUPPORT.telegramUser.trim().replace(/^@/, ''),
    telegramGroup: JI8_SUPPORT.telegramGroup.trim(),
    qqGroup: JI8_SUPPORT.qqGroup.trim(),
    wechat: '',
    qq: '',
  }
})
const hasChannels = computed(() => Object.values(channels.value).some(Boolean))
const panelId = 'j8-contact-panel'
const open = ref(false)
const copiedKey = ref('')
const rootRef = ref<HTMLElement | null>(null)
let copiedTimer: number | undefined

const copyValue = async (key: 'qqGroup' | 'wechat' | 'qq') => {
  try {
    await copyText(channels.value[key])
    copiedKey.value = key
    toast.success(t('ji8.contact.copied'))
    window.clearTimeout(copiedTimer)
    copiedTimer = window.setTimeout(() => (copiedKey.value = ''), 2000)
  } catch {
    toast.error(t('ji8.contact.copyFailed'))
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
