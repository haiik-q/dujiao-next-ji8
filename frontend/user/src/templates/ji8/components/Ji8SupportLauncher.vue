<template>
  <a
    v-if="supportLink"
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
import { useI18n } from 'vue-i18n'
import { Headset } from 'lucide-vue-next'
import { useJi8Nav } from '../composables/useJi8Nav'

/** 右下角客服浮标：contact.telegram || contact.whatsapp 有值才渲染（D6） */
const { t } = useI18n()
const { supportLink } = useJi8Nav()
</script>
