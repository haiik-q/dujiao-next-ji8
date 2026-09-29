<template>
  <div class="j8-xcheck" :class="{ 'is-compact': compact }">
    <form class="j8-xcheck-form" autocomplete="off" @submit.prevent="run">
      <label>
        <span>{{ t('ji8.xcheck.inputLabel') }}</span>
        <textarea
          v-if="!compact"
          v-model="input"
          rows="4"
          maxlength="600"
          spellcheck="false"
          :placeholder="t('ji8.xcheck.placeholderMulti')"
          @keydown.ctrl.enter.prevent="run"
          @keydown.meta.enter.prevent="run"
        ></textarea>
        <div v-else class="j8-xcheck-inline">
          <input v-model="input" type="text" maxlength="120" spellcheck="false" :placeholder="t('ji8.xcheck.placeholder')" />
          <button type="submit" class="j8-xcheck-btn" :disabled="checking">
            <Loader2 v-if="checking" class="j8-xcheck-spin" />
            <SearchCheck v-else />
            {{ checking ? t('ji8.xcheck.checking') : t('ji8.xcheck.check') }}
          </button>
        </div>
      </label>
      <button v-if="!compact" type="submit" class="j8-xcheck-btn is-block" :disabled="checking">
        <Loader2 v-if="checking" class="j8-xcheck-spin" />
        <SearchCheck v-else />
        {{ checking ? t('ji8.xcheck.checking') : t('ji8.xcheck.check') }}
      </button>
      <p class="j8-xcheck-note">{{ t('ji8.xcheck.note') }}</p>
    </form>

    <p v-if="formError" class="j8-xcheck-error">{{ formError }}</p>

    <ul v-if="rows.length" class="j8-xcheck-results" aria-live="polite">
      <li v-for="row in rows" :key="row.input" :class="`is-${row.tone}`">
        <img v-if="row.avatar" :src="row.avatar" alt="" referrerpolicy="no-referrer" loading="lazy" />
        <span v-else class="j8-xcheck-avatar"><AtSign /></span>
        <div>
          <b>
            <template v-if="row.name">{{ row.name }} <small>@{{ row.handle }}</small></template>
            <template v-else>@{{ row.handle }}</template>
          </b>
          <p>{{ row.message }}</p>
        </div>
        <CircleCheck v-if="row.tone === 'good'" class="j8-xcheck-icon" />
        <CircleX v-else-if="row.tone === 'bad'" class="j8-xcheck-icon" />
        <CircleAlert v-else class="j8-xcheck-icon" />
      </li>
    </ul>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { AtSign, CircleAlert, CircleCheck, CircleX, Loader2, SearchCheck } from 'lucide-vue-next'
import { xCheckAPI } from '../../../api'

/** X Premium 赠礼资格自检：compact=商品详情里的单行版，否则为工具页的多行批量版（最多 5 个）。 */
const props = withDefaults(defineProps<{ compact?: boolean }>(), { compact: false })

const { t } = useI18n()
const MAX_BATCH = 5
const HANDLE_RE = /^[A-Za-z0-9_]{1,15}$/

interface Row {
  input: string
  handle: string
  name?: string
  avatar?: string
  tone: 'good' | 'bad' | 'warn'
  message: string
}

const input = ref('')
const checking = ref(false)
const formError = ref('')
const rows = ref<Row[]>([])

/** 与后端 NormalizeHandle 一致：接受 @name / name / x.com 或 twitter.com 主页链接 */
function normalize(raw: string): string {
  let s = raw.trim().replace(/^https?:\/\//i, '')
  s = s.replace(/^(www\.|mobile\.)?(x|twitter)\.com\//i, '')
  return (s.split(/[/?#]/)[0] ?? '').replace(/^@/, '')
}

function reasonText(reason?: string): string {
  const key = `ji8.xcheck.reasons.${reason || 'not_eligible'}`
  const text = t(key)
  return text === key ? t('ji8.xcheck.reasons.not_eligible') : text
}

async function checkOne(raw: string): Promise<Row> {
  const handle = normalize(raw)
  if (!HANDLE_RE.test(handle)) {
    return { input: raw, handle: raw.trim(), tone: 'warn', message: t('ji8.xcheck.invalid') }
  }
  try {
    const res = await xCheckAPI.check(handle)
    const data = res.data?.data || {}
    const row: Row = {
      input: raw,
      handle: data.handle || handle,
      name: data.profile?.name,
      avatar: data.profile?.avatar_url,
      tone: data.eligible ? 'good' : 'bad',
      message: data.eligible ? t('ji8.xcheck.eligible') : reasonText(data.reason),
    }
    return row
  } catch (err: any) {
    return { input: raw, handle, tone: 'warn', message: err?.message || t('ji8.xcheck.failed') }
  }
}

async function run() {
  if (checking.value) return
  formError.value = ''
  const items = Array.from(new Set(input.value.split(/[\s,，;；]+/).map((s) => s.trim()).filter(Boolean)))
  if (!items.length) {
    formError.value = t('ji8.xcheck.empty')
    return
  }
  const batch = props.compact ? items.slice(0, 1) : items.slice(0, MAX_BATCH)
  if (!props.compact && items.length > MAX_BATCH) formError.value = t('ji8.xcheck.tooMany', { n: MAX_BATCH })
  checking.value = true
  rows.value = []
  try {
    for (const item of batch) {
      rows.value = [...rows.value, await checkOne(item)]
    }
  } finally {
    checking.value = false
  }
}
</script>
