<template>
  <div class="space-y-1.5">
    <button
      type="button"
      class="inline-flex items-center gap-1.5 rounded-md border border-primary/40 bg-primary/10 px-3 py-1.5 text-xs font-semibold text-primary transition-colors hover:bg-primary/15 disabled:cursor-not-allowed disabled:opacity-60"
      :disabled="checking || !handle.trim()"
      @click="run"
    >
      <Loader2 v-if="checking" class="h-3.5 w-3.5 animate-spin" />
      <SearchCheck v-else class="h-3.5 w-3.5" />
      {{ checking ? t('ji8.xcheck.checking') : t('checkout.xCheckButton') }}
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
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { CircleAlert, CircleCheck, CircleX, Loader2, SearchCheck } from 'lucide-vue-next'
import { xCheckAPI } from '../../api'

/** 结算页「X 用户名」字段旁的赠礼资格自检（结果仅供参考，不阻止下单）。 */
const props = defineProps<{ handle: string }>()

const { t } = useI18n()
const checking = ref(false)
const result = ref<{ tone: 'good' | 'bad' | 'warn'; text: string } | null>(null)

// 改了用户名后旧结果作废
watch(() => props.handle, () => { result.value = null })

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

async function run() {
  if (checking.value) return
  checking.value = true
  result.value = null
  try {
    const res = await xCheckAPI.check(props.handle.trim())
    const data = res.data?.data || {}
    const who = data.profile?.name ? `${data.profile.name}（@${data.handle}）` : `@${data.handle}`
    result.value = data.eligible
      ? { tone: 'good', text: `${who}：${t('ji8.xcheck.eligible')}` }
      : { tone: 'bad', text: `${who}：${reasonText(data.reason)}` }
  } catch (err: any) {
    result.value = { tone: 'warn', text: err?.message || t('ji8.xcheck.failed') }
  } finally {
    checking.value = false
  }
}
</script>
