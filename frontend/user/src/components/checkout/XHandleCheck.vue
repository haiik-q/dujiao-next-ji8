<template>
  <div class="space-y-1.5">
    <button
      type="button"
      class="inline-flex items-center gap-1.5 rounded-md border border-primary/40 bg-primary/10 px-3 py-1.5 text-xs font-semibold text-primary transition-colors hover:bg-primary/15 disabled:cursor-not-allowed disabled:opacity-60"
      :disabled="checking || !valid"
      @click="run(true)"
    >
      <Loader2 v-if="checking" class="h-3.5 w-3.5 animate-spin" />
      <SearchCheck v-else class="h-3.5 w-3.5" />
      {{ checking ? t('ji8.xcheck.checking') : state ? t('checkout.xCheckRetry') : t('checkout.xCheckButton') }}
    </button>
    <p v-if="result" class="flex items-start gap-1.5 text-xs leading-relaxed" :class="toneClass">
      <CircleCheck v-if="result.tone === 'good'" class="mt-px h-3.5 w-3.5 shrink-0" />
      <CircleX v-else-if="result.tone === 'bad'" class="mt-px h-3.5 w-3.5 shrink-0" />
      <CircleAlert v-else class="mt-px h-3.5 w-3.5 shrink-0" />
      <span>{{ result.text }}</span>
    </p>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { CircleAlert, CircleCheck, CircleX, Loader2, SearchCheck } from 'lucide-vue-next'
import { checkXHandle, getXHandleState, normalizeXHandle } from '../../composables/useXHandleGate'

/**
 * 结算页「X 用户名」字段旁的赠礼资格检测：填好用户名自动检测，结果共享给 useCheckout，
 * 检测通过前不能提交订单（后端创建订单时还会再核实）。
 */
const props = defineProps<{ handle: string }>()

const { t } = useI18n()

// 停止输入 1 秒后自动检测，避免逐字查询吃掉按 IP 的查询额度
const AUTO_CHECK_DELAY_MS = 1000
let timer: ReturnType<typeof setTimeout> | undefined

const valid = computed(() => normalizeXHandle(props.handle) !== '')
const state = computed(() => getXHandleState(props.handle))
const checking = computed(() => state.value?.status === 'checking')

const run = (force = false) => {
  if (timer) clearTimeout(timer)
  if (valid.value) void checkXHandle(props.handle, force)
}

watch(() => props.handle, () => {
  if (timer) clearTimeout(timer)
  if (valid.value && !state.value) timer = setTimeout(() => run(), AUTO_CHECK_DELAY_MS)
}, { immediate: true })

onBeforeUnmount(() => {
  if (timer) clearTimeout(timer)
})

const result = computed<{ tone: 'good' | 'bad' | 'warn'; text: string } | null>(() => {
  const s = state.value
  if (!s || s.status === 'checking') return null
  if (s.status === 'error') return { tone: 'warn', text: s.message || t('ji8.xcheck.failed') }
  const who = s.name ? `${s.name}（@${s.handle}）` : `@${s.handle}`
  return s.status === 'eligible'
    ? { tone: 'good', text: `${who}：${t('ji8.xcheck.eligible')}` }
    : { tone: 'bad', text: `${who}：${reasonText(s.reason)}${t('checkout.xHandleBlockedSuffix')}` }
})

const toneClass = computed(() => {
  if (result.value?.tone === 'good') return 'text-emerald-600 dark:text-emerald-400'
  if (result.value?.tone === 'bad') return 'text-destructive'
  return 'text-amber-600 dark:text-amber-400'
})

function reasonText(reason?: string): string {
  const key = `ji8.xcheck.reasons.${reason || 'not_eligible'}`
  const text = t(key)
  return text === key ? t('ji8.xcheck.reasons.not_eligible') : text
}
</script>
