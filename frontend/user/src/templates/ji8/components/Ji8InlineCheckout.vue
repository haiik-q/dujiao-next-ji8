<template>
  <form class="j8-ic" novalidate @submit.prevent="submit">
    <!-- 交付信息（人工交付商品的下单表单） -->
    <template v-for="formProduct in manualFormProducts" :key="formProduct.itemKey">
      <div v-for="field in formProduct.fields" :key="`${formProduct.itemKey}-${field.key}`" class="j8-ic-field">
        <label :for="`j8-ic-${field.key}`">
          {{ getManualFieldLabel(field) }}<b v-if="field.required">*</b>
        </label>

        <textarea
          v-if="field.type === 'textarea'"
          :id="`j8-ic-${field.key}`"
          :value="fieldValue(formProduct.itemKey, field.key)"
          rows="3"
          :placeholder="getManualFieldPlaceholder(field)"
          @input="setField(formProduct.itemKey, field.key, ($event.target as HTMLTextAreaElement).value)"
        ></textarea>

        <select
          v-else-if="field.type === 'select'"
          :id="`j8-ic-${field.key}`"
          :value="fieldValue(formProduct.itemKey, field.key)"
          @change="setField(formProduct.itemKey, field.key, ($event.target as HTMLSelectElement).value)"
        >
          <option value="">{{ t('checkout.manualFormSelectPlaceholder') }}</option>
          <option v-for="option in field.options" :key="option" :value="option">{{ option }}</option>
        </select>

        <div v-else-if="field.type === 'radio' || field.type === 'checkbox'" class="j8-ic-options">
          <label v-for="option in field.options" :key="option">
            <input
              :type="field.type"
              :name="`j8-ic-${formProduct.itemKey}-${field.key}`"
              :checked="isOptionChecked(formProduct.itemKey, field.key, option, field.type)"
              @change="toggleOption(formProduct.itemKey, field.key, option, field.type, ($event.target as HTMLInputElement).checked)"
            />
            <span>{{ option }}</span>
          </label>
        </div>

        <div v-else :class="{ 'j8-ic-with-action': field.key === 'x_handle' }">
          <input
            :id="`j8-ic-${field.key}`"
            :value="fieldValue(formProduct.itemKey, field.key)"
            :type="field.type === 'number' ? 'number' : field.type === 'email' ? 'email' : field.type === 'phone' ? 'tel' : 'text'"
            :placeholder="getManualFieldPlaceholder(field)"
            autocomplete="off"
            spellcheck="false"
            @input="setField(formProduct.itemKey, field.key, ($event.target as HTMLInputElement).value)"
          />
          <XHandleCheck v-if="field.key === 'x_handle'" :handle="String(fieldValue(formProduct.itemKey, field.key) || '')" />
        </div>

        <p v-if="submitAttempted && manualFieldError(formProduct.itemKey, field.key)" class="j8-ic-error">
          {{ manualFieldError(formProduct.itemKey, field.key) }}
        </p>
      </div>
    </template>

    <!-- 游客信息：邮箱 + 自设订单查询密码（登录会员不显示） -->
    <template v-if="!isAuthenticated">
      <div class="j8-ic-field">
        <label for="j8-ic-email">{{ t('ji8.checkout.email') }}<b>*</b></label>
        <input id="j8-ic-email" v-model="guestEmail" type="email" autocomplete="email" :placeholder="t('ji8.checkout.emailPlaceholder')" />
        <p v-if="guestEmail && !guestEmailValid" class="j8-ic-error">{{ t('error.email_invalid') }}</p>
      </div>
      <div class="j8-ic-field">
        <label for="j8-ic-password">{{ t('ji8.checkout.orderPassword') }}<b>*</b></label>
        <input
          id="j8-ic-password"
          v-model="guestPassword"
          type="password"
          autocomplete="new-password"
          :placeholder="t('ji8.checkout.orderPasswordPlaceholder')"
        />
      </div>
      <div v-if="guestCaptchaEnabled" class="j8-ic-field">
        <label>{{ t('auth.common.captchaLabel') }}<b>*</b></label>
        <ImageCaptcha
          v-if="captchaProvider === 'image'"
          :ref="(el: any) => (guestImageCaptchaRef = el)"
          v-model="guestCaptchaPayload"
          :disabled="submitting"
          @config-stale="handleGuestCaptchaConfigStale"
        />
        <TurnstileCaptcha
          v-else-if="captchaProvider === 'turnstile'"
          :ref="(el: any) => (guestTurnstileRef = el)"
          v-model="guestTurnstileToken"
          :site-key="guestTurnstileSiteKey"
        />
      </div>
      <p class="j8-ic-hint">
        {{ t('ji8.checkout.guestHint') }}
        <router-link :to="loginLink">{{ t('ji8.checkout.loginInstead') }}</router-link>
      </p>
    </template>

    <!-- 付款方式 -->
    <section class="j8-ic-pay">
      <h3><CreditCard />{{ t('checkout.paymentMethod') }}</h3>

      <label v-if="showBalanceOption" class="j8-ic-balance" :class="{ 'is-active': useBalance }">
        <input v-model="useBalance" type="checkbox" :disabled="walletOnlyPayment" />
        <Wallet />
        <span>
          <b>{{ t('ji8.checkout.useBalance') }}</b>
          <small>{{ t('ji8.checkout.balance') }} {{ walletLoading ? t('common.loading') : formatMoney(walletBalance, previewCurrency) }}</small>
        </span>
        <em v-if="useBalance && requiresOnlineChannel">{{ t('ji8.checkout.balanceNotEnough', { amount: formatMoney(centsToAmount(expectedOnlinePayCents), previewCurrency) }) }}</em>
      </label>

      <template v-if="!walletOnlyPayment && requiresOnlineChannel">
        <div v-if="paymentChannels.length" class="j8-ic-channels">
          <button
            v-for="channel in paymentChannels"
            :key="channel.id"
            type="button"
            class="j8-ic-channel"
            :class="{ 'is-active': selectedChannelId === channel.id && !isChannelDisabledForAmount(channel) }"
            :disabled="isChannelDisabledForAmount(channel)"
            :title="isChannelDisabledForAmount(channel) ? channelAmountLimitHint(channel) : ''"
            @click="handleSelectChannel(channel)"
          >
            <i aria-hidden="true"></i>
            <img v-if="channel.icon" :src="getImageUrl(channel.icon)" alt="" loading="lazy" />
            <span>
              <b>{{ channel.name }}</b>
              <small v-if="channelFeeText(channel)">{{ channelFeeText(channel) }}</small>
              <small v-else-if="isChannelDisabledForAmount(channel)">{{ channelAmountLimitHint(channel) }}</small>
            </span>
          </button>
        </div>
        <p v-else class="j8-ic-hint">{{ t('checkout.noPaymentChannels') }}</p>
      </template>
      <p v-else-if="!requiresOnlineChannel" class="j8-ic-hint is-ok">{{ t('checkout.walletCoversAll') }}</p>
    </section>

    <!-- 应付金额 -->
    <div class="j8-ic-total">
      <span>{{ t('ji8.product.payable') }}</span>
      <strong>{{ formatMoney(displayTotal, previewCurrency) }}</strong>
    </div>
    <p v-if="selectedFeeCents > 0" class="j8-ic-hint is-right">
      {{ t('payment.feeLabel') }} {{ formatMoney(centsToAmount(selectedFeeCents), previewCurrency) }}
    </p>
    <p v-if="Number(previewMemberDiscount) > 0" class="j8-ic-hint is-right">
      {{ t('checkout.previewMemberDiscount') }} -{{ formatMoney(previewMemberDiscount, previewCurrency) }}
    </p>

    <p v-if="alertText" class="j8-ic-alert" :class="{ 'is-error': !!(error || previewError) || xHandleRejected }">{{ alertText }}</p>

    <div class="j8-pd-buttons">
      <button type="button" class="j8-pd-cart" :disabled="disabled" @click="emit('add-to-cart')">
        <ShoppingCart />
        {{ t('productDetail.addToCart') }}
      </button>
      <button type="submit" class="j8-pd-buy" :disabled="disabled || submitting || !!xHandleBlockedReason" :title="xHandleBlockedReason">
        <Loader2 v-if="submitting" class="j8-ic-spin" />
        {{ submitting ? t('checkout.submitting') : t('checkout.submitButton') }}
      </button>
    </div>
  </form>
</template>

<script setup lang="ts">
import { computed, toRef } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { CreditCard, Loader2, ShoppingCart, Wallet } from 'lucide-vue-next'
import type { CartItem } from '../../../stores/cart'
import { useUserAuthStore } from '../../../stores/userAuth'
import { useCheckout } from '../../../composables/useCheckout'
import { getImageUrl } from '../../../utils/image'
import ImageCaptcha from '../../../components/captcha/ImageCaptcha.vue'
import TurnstileCaptcha from '../../../components/captcha/TurnstileCaptcha.vue'
import XHandleCheck from '../../../components/checkout/XHandleCheck.vue'
import { getXHandleState } from '../../../composables/useXHandleGate'
import { formatMoney } from '../utils/price'
import { amountToCents, basisPointsToPercent, centsToAmount, rateToBasisPoints } from '../../../utils/money'

/**
 * 商品页一步下单：交付信息 + 游客邮箱/查询密码 + 付款方式 + 提交，直接调用 create-and-pay 后进入付款页。
 * 逻辑完全复用 useCheckout（价格试算、渠道、验证码、校验与提交），只是订单项取当前选中的规格 × 数量。
 */
const props = defineProps<{ items: CartItem[]; disabled?: boolean }>()
const emit = defineEmits<{ (e: 'add-to-cart'): void }>()

const { t } = useI18n()
const route = useRoute()
const userAuthStore = useUserAuthStore()
const isAuthenticated = computed(() => userAuthStore.isAuthenticated)
const loginLink = computed(() => `/auth/login?redirect=${encodeURIComponent(route.fullPath)}`)

const {
  manualFormProducts, manualFormData, submitAttempted,
  getManualFieldLabel, getManualFieldPlaceholder, manualFieldError, xHandleBlockedReason,
  guestEmail, guestPassword, guestEmailValid,
  guestCaptchaEnabled, captchaProvider, guestCaptchaPayload, guestTurnstileToken, guestTurnstileSiteKey,
  guestImageCaptchaRef, guestTurnstileRef, handleGuestCaptchaConfigStale,
  previewCurrency, previewTotal, previewMemberDiscount,
  error, previewError, canSubmit, submitBlockedReason,
  showBalanceOption, walletLoading, walletBalance, useBalance, walletOnlyPayment, expectedOnlinePayCents,
  requiresOnlineChannel, paymentChannels, selectedChannelId,
  isChannelDisabledForAmount, channelAmountLimitHint, handleSelectChannel,
  submitting, handleSubmit,
} = useCheckout({ items: toRef(props, 'items') })

// 买家承担手续费（后台 payment_config.customer_fee_enabled）时，按选中渠道把手续费算进应付金额，与后端 calculatePaymentAmounts 一致
const channelFee = (channel: any) => ({
  bp: channel?.fee_policy === 'customer_surcharge' ? rateToBasisPoints(channel?.fee_rate) || 0 : 0,
  fixedCents: channel?.fee_policy === 'customer_surcharge' ? amountToCents(String(channel?.fixed_fee ?? '')) || 0 : 0,
})
const channelFeeText = (channel: any) => {
  const { bp, fixedCents } = channelFee(channel)
  const parts: string[] = []
  if (bp > 0) parts.push(`${Number(basisPointsToPercent(bp))}%`)
  if (fixedCents > 0) parts.push(formatMoney(centsToAmount(fixedCents), previewCurrency.value))
  return parts.length ? `${t('payment.feeLabel')} ${parts.join(' + ')}` : ''
}
const selectedFeeCents = computed(() => {
  if (walletOnlyPayment.value || !requiresOnlineChannel.value) return 0
  const channel = paymentChannels.value.find((c: any) => c.id === selectedChannelId.value)
  if (!channel || isChannelDisabledForAmount(channel)) return 0
  const { bp, fixedCents } = channelFee(channel)
  return Math.round((expectedOnlinePayCents.value * bp) / 10000) + fixedCents
})

// 服务端试算前（如游客未填邮箱）previewTotal 自动回落为单价 × 数量
const displayTotal = computed(() => centsToAmount((amountToCents(previewTotal.value) || 0) + selectedFeeCents.value))

// X 账号检测结果是「不能接收」或检测失败（检测中、待检测只是黄色提示）
const xHandleRejected = computed(() => manualFormProducts.value.some((product) => {
  const status = getXHandleState(manualFormData.value[product.itemKey]?.x_handle)?.status
  return status === 'ineligible' || status === 'error'
}))

// 未点提交前只显示接口错误，不提前把「请填写…」类提示甩给客人
const alertText = computed(() => {
  if (error.value) return error.value
  if (previewError.value) return previewError.value
  // X 账号没检测通过时按钮是灰的，直接说明原因
  if (xHandleBlockedReason.value) return xHandleBlockedReason.value
  if (submitAttempted.value && !canSubmit.value) return submitBlockedReason.value
  return ''
})

const fieldValue = (itemKey: string, key: string) => manualFormData.value[itemKey]?.[key] ?? ''
const setField = (itemKey: string, key: string, value: unknown) => {
  manualFormData.value = { ...manualFormData.value, [itemKey]: { ...(manualFormData.value[itemKey] || {}), [key]: value } }
}
const isOptionChecked = (itemKey: string, key: string, option: string, type: string) => {
  const value = fieldValue(itemKey, key)
  return type === 'checkbox' ? Array.isArray(value) && value.includes(option) : value === option
}
const toggleOption = (itemKey: string, key: string, option: string, type: string, checked: boolean) => {
  if (type === 'radio') return setField(itemKey, key, option)
  const current = fieldValue(itemKey, key)
  const list = Array.isArray(current) ? current.filter((v: string) => v !== option) : []
  setField(itemKey, key, checked ? [...list, option] : list)
}

const submit = () => {
  if (props.disabled || xHandleBlockedReason.value) return
  void handleSubmit()
}

defineExpose({ submit })
</script>
