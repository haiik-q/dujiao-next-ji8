<template>
  <div class="j8-container j8-xcheck-main">
    <section class="j8-xcheck-hero">
      <span class="j8-eyebrow">{{ t('ji8.xcheck.eyebrow') }}</span>
      <h1>{{ t('ji8.xcheck.title') }}</h1>
      <p>{{ t('ji8.xcheck.subtitle') }}</p>
    </section>
    <section class="j8-xcheck-panel">
      <Ji8XCheck />
      <aside class="j8-xcheck-rules">
        <h2>{{ t('ji8.xcheck.rulesTitle') }}</h2>
        <ul>
          <li v-for="key in ruleKeys" :key="key">{{ t(`ji8.xcheck.reasons.${key}`) }}</li>
        </ul>
      </aside>
    </section>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Ji8XCheck from './components/Ji8XCheck.vue'
import './styles/xcheck.css'

const { t } = useI18n()
const ruleKeys = ['already_premium', 'verified_organization', 'protected', 'hidden_subscriptions', 'suspended']
</script>
