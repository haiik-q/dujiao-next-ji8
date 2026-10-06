<template>
  <!-- ji8：链上付款按金额匹配订单，交易所提币扣手续费会导致少到账、订单识别不到 -->
  <div class="w-full rounded-xl border border-amber-300 bg-amber-50 p-3 text-left text-amber-900 dark:border-amber-500/40 dark:bg-amber-500/10 dark:text-amber-100">
    <div class="flex items-start gap-2">
      <AlertTriangle class="mt-0.5 h-4 w-4 shrink-0 text-amber-600 dark:text-amber-400" />
      <div class="min-w-0 space-y-1.5 text-xs leading-relaxed sm:text-sm">
        <i18n-t :keypath="amount ? 'payment.exactAmount.title' : 'payment.exactAmount.titleGeneric'" tag="p" class="font-bold">
          <template #received><span :class="HIGHLIGHT">{{ t('payment.exactAmount.receivedWord') }}</span></template>
          <template #amount><span :class="HIGHLIGHT">{{ amountLabel }}</span></template>
        </i18n-t>
        <!-- 钱包转账和交易所提币分开写：钱包不扣 USDT 手续费，照抄交易所的「+ 手续费」会多付（10-06 有客人钱包多转了 0.01） -->
        <i18n-t :keypath="amount ? 'payment.exactAmount.wallet' : 'payment.exactAmount.walletGeneric'" tag="p">
          <template #label><span class="font-bold">• {{ t('payment.exactAmount.walletLabel') }}</span></template>
          <template #amount><span :class="HIGHLIGHT">{{ amountLabel }}</span></template>
        </i18n-t>
        <i18n-t :keypath="exampleFill ? 'payment.exactAmount.exchange' : 'payment.exactAmount.exchangeGeneric'" tag="p">
          <template #label><span class="font-bold">• {{ t('payment.exactAmount.exchangeLabel') }}</span></template>
          <template #fill><span class="font-bold">{{ exampleFill }}</span></template>
          <template #receivedTag><span :class="HIGHLIGHT">{{ t('payment.exactAmount.receivedTag') }}</span></template>
        </i18n-t>
        <p class="font-semibold text-red-600 dark:text-red-400">{{ t('payment.exactAmount.mismatch') }}</p>
        <button
          type="button"
          class="inline-flex items-center gap-1 font-semibold text-amber-700 underline-offset-2 hover:underline dark:text-amber-300"
          @click="showGuide = !showGuide"
        >
          <ImageIcon class="h-3.5 w-3.5" />
          {{ showGuide ? t('payment.exactAmount.hideExample') : t('payment.exactAmount.showExample') }}
        </button>
        <a v-if="showGuide" :href="GUIDE_IMAGE" target="_blank" rel="noopener noreferrer" class="block">
          <img :src="GUIDE_IMAGE" :alt="t('payment.exactAmount.showExample')" loading="lazy" class="mt-1 w-full max-w-sm rounded-lg border border-amber-200 dark:border-amber-500/30" />
        </a>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { AlertTriangle, Image as ImageIcon } from 'lucide-vue-next'

const GUIDE_IMAGE = '/images/usdt-fee-guide.jpg?v=2'
// 「实际到账数量」、应付金额标红加粗，提醒客人核对到账数量
const HIGHLIGHT = 'font-extrabold text-red-600 dark:text-red-400'

const props = defineProps<{
  details: Array<{ key: string; value: string }>
}>()

const { t } = useI18n()
// 示例图默认展开，客人可以收起
const showGuide = ref(true)

const amount = computed(() => props.details.find((d) => d.key === 'amount')?.value?.trim() || '')
const token = computed(() => props.details.find((d) => d.key === 'token')?.value?.trim() || '')
const amountLabel = computed(() => (token.value && !amount.value.includes(token.value) ? `${amount.value} ${token.value}` : amount.value))

// 示例：手续费按 0.01 算，提币数量 = 应付 + 0.01（保留应付金额的小数位数，至少 2 位）
const exampleFill = computed(() => {
  const m = amount.value.match(/\d+(?:\.(\d+))?/)
  if (!m) return ''
  const decimals = Math.max(2, (m[1] || '').length)
  const value = Number(m[0]) + 0.01
  return Number.isFinite(value) ? value.toFixed(decimals) : ''
})
</script>
