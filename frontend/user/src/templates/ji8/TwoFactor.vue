<template>
  <div class="j8-container j8-twofa-main">
    <section class="j8-twofa-hero">
      <div>
        <span class="j8-eyebrow">{{ t('ji8.totp.eyebrow') }}</span>
        <h1>{{ t('ji8.totp.title') }}</h1>
      </div>
      <div class="j8-twofa-privacy-badges" :aria-label="t('ji8.totp.badgesLabel')">
        <span>{{ t('ji8.totp.badgeLocal') }}</span>
        <span>{{ t('ji8.totp.badgeNoUpload') }}</span>
        <span>{{ t('ji8.totp.badgeEphemeral') }}</span>
      </div>
    </section>

    <section class="j8-twofa-panel">
      <div class="j8-totp-tool" :class="{ 'is-empty': !hasResult }">
        <form class="j8-totp-input-card" autocomplete="off" @submit.prevent="generate">
          <header>
            <span>{{ t('ji8.totp.cardEyebrow') }}</span>
            <h2>{{ t('ji8.totp.cardTitle') }}</h2>
            <p>{{ t('ji8.totp.cardHelp') }}</p>
          </header>
          <label>
            <span>{{ t('ji8.totp.inputLabel') }}</span>
            <textarea
              ref="textareaEl"
              v-model="input"
              rows="7"
              maxlength="8192"
              autocomplete="off"
              autocapitalize="characters"
              spellcheck="false"
              data-1p-ignore="true"
              :placeholder="t('ji8.totp.placeholder')"
              @input="onInput"
              @keydown.ctrl.enter.prevent="generate"
              @keydown.meta.enter.prevent="generate"
            ></textarea>
          </label>
          <div class="j8-totp-input-actions">
            <button type="submit" class="j8-totp-btn-primary">{{ t('ji8.totp.generate') }}</button>
            <button type="button" class="j8-totp-btn-secondary" @click="clear">{{ t('ji8.totp.clear') }}</button>
          </div>
          <p class="j8-totp-shortcut">{{ t('ji8.totp.shortcut') }}</p>
          <p class="j8-totp-local-note">
            <b>{{ t('ji8.totp.noteTitle') }}</b>
            <span>{{ t('ji8.totp.noteBody') }}</span>
          </p>
        </form>

        <section v-if="hasResult" class="j8-totp-result-card" aria-live="polite">
          <!-- 单密钥 -->
          <div v-if="single" class="j8-totp-single-result">
            <div class="j8-totp-code-panel">
              <div>
                <span>{{ t('ji8.totp.currentCode') }}</span>
                <output :class="{ 'is-error': single.slot.error }">{{ codeText(single.slot) }}</output>
                <small>{{ candidateLabel(single.candidate) }} · {{ maskSecret(single.candidate.secret) }}</small>
              </div>
              <div class="j8-totp-countdown" :style="{ '--totp-progress': String(single.slot.progress) }" aria-hidden="true">
                <b>{{ single.slot.remaining }}</b><small>{{ t('ji8.totp.seconds') }}</small>
              </div>
            </div>
            <button type="button" class="j8-totp-copy-button" :disabled="!single.slot.code" @click="copyCode(0)">
              <component :is="copiedIndex === 0 ? Check : Copy" />
              {{ copiedIndex === 0 ? t('ji8.totp.copied') : t('ji8.totp.copy') }}
            </button>
            <dl class="j8-totp-meta">
              <div><dt>{{ t('ji8.totp.metaAlgorithm') }}</dt><dd>{{ single.candidate.algorithm }}</dd></div>
              <div><dt>{{ t('ji8.totp.metaDigits') }}</dt><dd>{{ t('ji8.totp.digitsValue', { n: single.candidate.digits }) }}</dd></div>
              <div><dt>{{ t('ji8.totp.metaPeriod') }}</dt><dd>{{ t('ji8.totp.periodValue', { n: single.candidate.period }) }}</dd></div>
              <div><dt>{{ t('ji8.totp.metaAccount') }}</dt><dd>{{ single.candidate.label || t('ji8.totp.unknownAccount') }}</dd></div>
            </dl>
            <p class="j8-totp-time-note">{{ t('ji8.totp.timeNote') }}</p>
          </div>

          <!-- 多密钥 -->
          <section v-else class="j8-totp-multiple-result">
            <header class="j8-totp-multiple-header">
              <div>
                <span>{{ t('ji8.totp.multiEyebrow') }}</span>
                <i18n-t keypath="ji8.totp.multiTitle" tag="h3">
                  <template #n><b>{{ candidates.length }}</b></template>
                </i18n-t>
              </div>
              <p>{{ t('ji8.totp.multiHint') }}</p>
            </header>
            <div class="j8-totp-multiple-list">
              <article
                v-for="(candidate, index) in candidates"
                :key="`${candidate.secret}-${index}`"
                class="j8-totp-multiple-card"
                :aria-label="t('ji8.totp.cardAria', { n: index + 1 })"
              >
                <header class="j8-totp-multiple-card-header">
                  <div>
                    <b>{{ t('ji8.totp.cardIndex', { n: index + 1 }) }}</b>
                    <span>{{ candidateLabel(candidate) }} · {{ maskSecret(candidate.secret) }}</span>
                  </div>
                  <span class="j8-totp-multiple-settings">
                    {{ candidate.algorithm }} · {{ t('ji8.totp.digitsValue', { n: candidate.digits }) }} · {{ t('ji8.totp.periodShort', { n: candidate.period }) }}
                  </span>
                </header>
                <div class="j8-totp-multiple-code-row">
                  <div class="j8-totp-multiple-code">
                    <span>{{ t('ji8.totp.currentCode') }}</span>
                    <output :class="{ 'is-error': slots[index]?.error }">{{ codeText(slots[index]) }}</output>
                  </div>
                  <div
                    class="j8-totp-countdown is-compact"
                    :style="{ '--totp-progress': String(slots[index]?.progress ?? 1) }"
                    aria-hidden="true"
                  >
                    <b>{{ slots[index]?.remaining ?? candidate.period }}</b><small>{{ t('ji8.totp.seconds') }}</small>
                  </div>
                  <button type="button" class="j8-totp-multiple-copy" :disabled="!slots[index]?.code" @click="copyCode(index)">
                    <component :is="copiedIndex === index ? Check : Copy" />
                    {{ copiedIndex === index ? t('ji8.totp.copied') : t('ji8.totp.copyCard') }}
                  </button>
                </div>
                <p class="j8-totp-multiple-card-status" :class="{ 'is-error': slots[index]?.error }">{{ cardStatus(candidate, slots[index]) }}</p>
              </article>
            </div>
            <p class="j8-totp-time-note">{{ t('ji8.totp.multiTimeNote') }}</p>
          </section>
        </section>

        <p class="j8-totp-status" :class="statusClass" role="status">{{ status.text }}</p>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { Check, Copy } from 'lucide-vue-next'
import { usePageSeo } from '../../composables/usePageSeo'
import { copyText } from '../../utils/clipboard'
import {
  formatTotpCode,
  generateTotp,
  isTotpSupported,
  maskSecret,
  parseTotpInput,
  totpCounter,
  totpProgress,
  totpRemainingSeconds,
  type TotpCandidate,
} from './utils/totp'
import './styles/totp.css'

/**
 * 2FA 动态验证码（仿 ai.bahk.cn/2fa.html）。
 * 全部在浏览器内完成：解析粘贴内容 → Web Crypto HMAC → 展示验证码与倒计时；
 * 密钥只存在于组件状态里，不写 localStorage、不写 URL，离开页面即清除。
 * 入口：/2fa（/2fa.html 重定向）；支持 ?secret= 预填（只读，不回写）。
 */

type StatusKind = 'info' | 'error' | 'success'

interface CodeSlot {
  code: string
  remaining: number
  progress: number
  error: boolean
}

const { t } = useI18n()
const route = useRoute()

usePageSeo({
  title: () => t('ji8.totp.title'),
  description: () => t('ji8.totp.seoDescription'),
  canonicalPath: () => '/2fa',
})

const TICK_MS = 250
const COPIED_MS = 1600

const input = ref('')
const textareaEl = ref<HTMLTextAreaElement | null>(null)
const candidates = ref<TotpCandidate[]>([])
const slots = ref<CodeSlot[]>([])
const status = ref<{ text: string; kind: StatusKind }>({ text: '', kind: 'info' })
const copiedIndex = ref<number | null>(null)

let timer: number | null = null
let copiedTimer: number | null = null
/** 每次开始/停止都 +1，用来丢弃过期的异步计算结果 */
let generation = 0
let lastCounters: Array<number | null> = []
let pendingCounters: Array<number | null> = []

const hasResult = computed(() => candidates.value.length > 0)
const single = computed(() => {
  if (candidates.value.length !== 1) return null
  const candidate = candidates.value[0]
  const slot = slots.value[0]
  return candidate && slot ? { candidate, slot } : null
})
const statusClass = computed(() => (status.value.text ? `is-${status.value.kind}` : ''))

const setStatus = (text: string, kind: StatusKind = 'info') => {
  status.value = { text, kind }
}

const candidateLabel = (candidate: TotpCandidate): string => candidate.label || t(`ji8.totp.source.${candidate.source}`)

const codeText = (slot: CodeSlot | undefined): string => {
  if (!slot) return '--- ---'
  if (slot.error) return t('ji8.totp.generateFailed')
  return slot.code ? formatTotpCode(slot.code) : '--- ---'
}

const cardStatus = (candidate: TotpCandidate, slot: CodeSlot | undefined): string => {
  if (!slot || !slot.code) return slot?.error ? t('ji8.totp.cardStatusFailed') : t('ji8.totp.cardStatusPending')
  return t('ji8.totp.cardStatusReady', { secret: maskSecret(candidate.secret) })
}

const stopTimer = () => {
  if (timer !== null) {
    window.clearInterval(timer)
    timer = null
  }
}

/** 每 250ms 刷新倒计时；只有跨过周期边界时才重新做 HMAC */
const tick = async (force = false) => {
  const list = candidates.value
  if (!list.length) return
  const lifecycle = generation
  const now = Date.now()
  const jobs: Array<Promise<{ index: number; counter: number; code?: string; failed?: boolean }>> = []

  list.forEach((candidate, index) => {
    const slot = slots.value[index]
    if (!slot) return
    slot.remaining = totpRemainingSeconds(candidate.period, now)
    slot.progress = totpProgress(candidate.period, now)
    const counter = totpCounter(candidate.period, now)
    if (!force && (lastCounters[index] === counter || pendingCounters[index] === counter)) return
    pendingCounters[index] = counter
    jobs.push(
      generateTotp({
        secret: candidate.secret,
        algorithm: candidate.algorithm,
        digits: candidate.digits,
        period: candidate.period,
        timestamp: now,
      })
        .then((code) => ({ index, counter, code }))
        .catch(() => ({ index, counter, failed: true })),
    )
  })
  if (!jobs.length) return

  const results = await Promise.all(jobs)
  if (lifecycle !== generation) return

  let failures = 0
  results.forEach(({ index, counter, code, failed }) => {
    if (pendingCounters[index] === counter) pendingCounters[index] = null
    const slot = slots.value[index]
    if (!slot) return
    if (failed || !code) {
      failures += 1
      lastCounters[index] = null
      slot.code = ''
      slot.error = true
      return
    }
    lastCounters[index] = counter
    slot.code = code
    slot.error = false
  })

  if (list.length > 1) {
    setStatus(
      failures ? t('ji8.totp.statusFailedCount', { n: failures }) : t('ji8.totp.statusMulti', { n: list.length }),
      failures ? 'error' : 'success',
    )
  } else {
    setStatus(failures ? t('ji8.totp.statusFailed') : t('ji8.totp.statusSingle'), failures ? 'error' : 'success')
  }
}

const start = () => {
  stopTimer()
  generation += 1
  lastCounters = candidates.value.map(() => null)
  pendingCounters = candidates.value.map(() => null)
  void tick(true)
  timer = window.setInterval(() => void tick(false), TICK_MS)
}

/** 停止计算并清掉结果（保留输入框内容） */
const reset = () => {
  generation += 1
  stopTimer()
  candidates.value = []
  slots.value = []
  lastCounters = []
  pendingCounters = []
  copiedIndex.value = null
  setStatus('')
}

const generate = () => {
  const text = input.value.trim()
  reset()
  if (!text) {
    setStatus(t('ji8.totp.statusEmpty'), 'info')
    return
  }
  if (!isTotpSupported()) {
    setStatus(t('ji8.totp.unsupported'), 'error')
    return
  }
  const found = parseTotpInput(text)
  if (!found.length) {
    setStatus(t('ji8.totp.statusNone'), 'error')
    return
  }
  candidates.value = found
  slots.value = found.map((candidate) => ({ code: '', remaining: candidate.period, progress: 1, error: false }))
  start()
}

const clear = () => {
  reset()
  input.value = ''
  textareaEl.value?.focus({ preventScroll: true })
}

/** 内容一变就撤下旧结果，避免验证码与输入不一致 */
const onInput = () => {
  if (hasResult.value) reset()
}

const copyCode = async (index: number) => {
  const code = slots.value[index]?.code
  if (!code) return
  try {
    await copyText(code)
    copiedIndex.value = index
    if (copiedTimer !== null) window.clearTimeout(copiedTimer)
    copiedTimer = window.setTimeout(() => {
      copiedIndex.value = null
      copiedTimer = null
    }, COPIED_MS)
    setStatus(
      candidates.value.length > 1 ? t('ji8.totp.copiedCard', { n: index + 1 }) : t('ji8.totp.copiedDetail'),
      'success',
    )
  } catch {
    setStatus(t('ji8.totp.copyFailed'), 'error')
  }
}

onMounted(() => {
  // ?secret= 只读取、不回写；密钥永远不由本页写进地址栏
  const raw = route.query.secret
  const value = Array.isArray(raw) ? raw[0] : raw
  if (typeof value === 'string' && value.trim()) {
    input.value = value.trim()
    generate()
  }
})

onBeforeUnmount(() => {
  reset()
  if (copiedTimer !== null) window.clearTimeout(copiedTimer)
  input.value = ''
})
</script>
