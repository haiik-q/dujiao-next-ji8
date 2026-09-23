<template>
  <div class="j8-container j8-acc">
    <!-- 侧栏：用户卡 + 分组导航 + 返回购物主页（桌面）；移动端收起为横向 chips -->
    <aside class="j8-acc-side">
      <div class="j8-acc-user">
        <span class="j8-acc-avatar">{{ displayInitial }}</span>
        <div>
          <b>{{ userProfileStore.displayName }}</b>
          <small>{{ levelName(userProfileStore.currentLevel) }}</small>
        </div>
      </div>

      <nav class="j8-acc-nav">
        <template v-for="group in navGroups" :key="group.key">
          <span class="j8-acc-nav-label">{{ group.label }}</span>
          <button
            v-for="item in group.items"
            :key="item.key"
            type="button"
            :class="{ 'is-active': currentSection === item.key }"
            @click="switchSection(item.key)"
          >
            <i><component :is="item.icon" /></i>
            <span>{{ item.label }}</span>
          </button>
        </template>
      </nav>

      <router-link to="/" class="j8-acc-back">
        <i><ChevronLeft /></i>
        <span>{{ t('ji8.account.backToShop') }}</span>
      </router-link>
    </aside>

    <div class="j8-acc-chips">
      <button
        v-for="item in flatNavItems"
        :key="item.key"
        type="button"
        :class="{ 'is-active': currentSection === item.key }"
        @click="switchSection(item.key)"
      >
        <component :is="item.icon" />
        <span>{{ item.label }}</span>
      </button>
    </div>

    <section class="j8-acc-main">
      <header class="j8-acc-head">
        <div>
          <span class="j8-eyebrow">MEMBER CENTER</span>
          <h1>{{ currentTitle }}</h1>
        </div>
        <div class="j8-acc-head-actions">
          <router-link to="/" class="j8-acc-btn">{{ t('ji8.account.shop') }} <ArrowUpRight /></router-link>
          <button type="button" class="j8-acc-btn is-danger" @click="userAuthStore.logout()">{{ t('ji8.account.logout') }}</button>
        </div>
      </header>

      <Alert
        v-if="globalAlert"
        :variant="pageAlertVariant(globalAlert.level)"
        :class="pageAlertToneClass(globalAlert.level)"
      >
        <AlertDescription>{{ globalAlert.message }}</AlertDescription>
      </Alert>

      <template v-if="currentSection === 'overview'">
        <!-- 欢迎卡 -->
        <section class="j8-acc-hero">
          <div class="j8-acc-hero-main">
            <span class="j8-acc-avatar is-lg">{{ displayInitial }}</span>
            <div>
              <span class="j8-acc-hero-eyebrow">{{ t('ji8.account.welcome') }}</span>
              <h2>{{ userProfileStore.profile?.email || userProfileStore.displayName }}</h2>
              <p>{{ t('ji8.account.welcomeDesc') }}</p>
            </div>
          </div>
          <div class="j8-acc-hero-side">
            <div v-if="affiliate?.affiliate_code" class="j8-acc-invite">
              <label>{{ t('ji8.account.inviteCode') }}</label>
              <div>
                <b>{{ affiliate.affiliate_code }}</b>
                <button type="button" @click="copyInvite">{{ t('ji8.account.copy') }}</button>
              </div>
            </div>
            <router-link to="/" class="j8-acc-cta">
              <span>{{ t('ji8.account.continueShopping') }}</span>
              <ArrowRight />
            </router-link>
          </div>
        </section>

        <!-- 快捷入口 -->
        <div class="j8-acc-quick">
          <button v-for="q in quickItems" :key="q.key" type="button" @click="switchSection(q.key)">
            <i><component :is="q.icon" /></i>
            <span>
              <b>{{ q.title }}</b>
              <small>{{ q.desc }}</small>
            </span>
            <ChevronRight class="j8-acc-quick-arrow" />
          </button>
        </div>

        <!-- 资金统计 -->
        <div class="j8-acc-stats">
          <div v-for="s in statItems" :key="s.key" class="j8-acc-stat">
            <i><component :is="s.icon" /></i>
            <div>
              <small>{{ s.label }}</small>
              <strong>{{ s.value }}</strong>
              <span>{{ s.hint }}</span>
            </div>
          </div>
        </div>

        <!-- 个人资料 -->
        <section class="j8-acc-card">
          <header class="j8-acc-card-head">
            <i><IdCard /></i>
            <div>
              <span>PROFILE</span>
              <h3>{{ t('ji8.account.profileTitle') }}</h3>
            </div>
            <button type="button" class="j8-acc-link" @click="switchSection('profile')">{{ t('ji8.account.editProfile') }} →</button>
          </header>
          <dl class="j8-acc-info">
            <div v-for="row in profileRows" :key="row.label">
              <dt>{{ row.label }}</dt>
              <dd>{{ row.value }}</dd>
            </div>
          </dl>
        </section>

        <!-- 最近购买 -->
        <section class="j8-acc-card">
          <header class="j8-acc-card-head">
            <i><ReceiptText /></i>
            <div>
              <span>RECENT PURCHASES</span>
              <h3>{{ t('ji8.account.recentTitle') }}</h3>
            </div>
            <button type="button" class="j8-acc-link" @click="switchSection('orders')">{{ t('ji8.account.allRecords') }} →</button>
          </header>

          <div v-if="userProfileStore.loadingOrders" class="j8-acc-empty">…</div>
          <div v-else-if="userProfileStore.recentOrders.length === 0" class="j8-acc-empty">
            {{ t('personalCenter.overview.emptyOrders') }}
            <router-link to="/">{{ t('ji8.account.goShopping') }} →</router-link>
          </div>
          <div v-else class="j8-acc-table-wrap">
            <table class="j8-acc-table">
              <thead>
                <tr>
                  <th>{{ t('ji8.account.col.orderNo') }}</th>
                  <th>{{ t('ji8.account.col.time') }}</th>
                  <th>{{ t('ji8.account.col.product') }}</th>
                  <th>{{ t('ji8.account.col.qty') }}</th>
                  <th>{{ t('ji8.account.col.amount') }}</th>
                  <th>{{ t('ji8.account.col.status') }}</th>
                  <th>{{ t('ji8.account.col.action') }}</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="order in recentRows" :key="order.order_no">
                  <td :data-label="t('ji8.account.col.orderNo')"><b class="j8-acc-mono">{{ order.order_no }}</b></td>
                  <td :data-label="t('ji8.account.col.time')">{{ formatDate(order.created_at) }}</td>
                  <td :data-label="t('ji8.account.col.product')">
                    <span class="j8-acc-product">
                      <Ji8BrandIcon :name="order.title" />
                      <span>{{ order.title || '-' }}</span>
                    </span>
                  </td>
                  <td :data-label="t('ji8.account.col.qty')">{{ order.quantity }}</td>
                  <td :data-label="t('ji8.account.col.amount')"><b class="j8-acc-amount">{{ money(order.total_amount, order.currency) }}</b></td>
                  <td :data-label="t('ji8.account.col.status')">
                    <span class="j8-acc-pill" :class="`is-${statusVariant(order.status)}`">{{ statusLabel(order.status) }}</span>
                  </td>
                  <td :data-label="t('ji8.account.col.action')">
                    <div class="j8-acc-row-actions">
                      <router-link v-if="order.status === 'pending_payment'" :to="`/pay?order_no=${order.order_no}`" class="j8-acc-btn-primary is-warm">{{ t('orders.payNow') }}</router-link>
                      <router-link :to="`/orders/${order.order_no}`" class="j8-acc-btn-primary">{{ t('ji8.account.viewOrder') }}</router-link>
                    </div>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </section>
      </template>

      <ProfilePanel v-else-if="currentSection === 'profile'" />
      <SecurityPanel v-else-if="currentSection === 'security'" />
      <OrdersPanel v-else-if="currentSection === 'orders'" />
      <WalletPanel v-else-if="currentSection === 'wallet'" />
      <AffiliatePanel v-else-if="currentSection === 'affiliate'" />
      <GiftCardPanel v-else-if="currentSection === 'giftCard'" />
      <ApiPanel v-else-if="currentSection === 'api'" />
      <OrdersPanel v-else />
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, type Component } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  ArrowRight, ArrowUpRight, Banknote, ChevronLeft, ChevronRight, CircleDollarSign, Coins, Crown, Gift,
  House, IdCard, Key, Megaphone, ReceiptText, ShieldCheck, ShoppingBag, UserCircle, Wallet,
} from 'lucide-vue-next'
import { affiliateAPI, walletAPI, type AffiliateDashboardData } from '../../api'
import { useUserAuthStore } from '../../stores/userAuth'
import { useAppStore } from '../../stores/app'
import { useLocalized } from '../../composables/useProduct'
import { usePersonalCenter, type PersonalSection } from '../../composables/usePersonalCenter'
import { pageAlertVariant, pageAlertToneClass } from '../../utils/alerts'
import { copyText } from '../../utils/clipboard'
import { toast } from '../../composables/useToast'
import { Alert, AlertDescription } from '@/components/ui/alert'
import ProfilePanel from '../../views/personal/ProfilePanel.vue'
import SecurityPanel from '../../views/personal/SecurityPanel.vue'
import OrdersPanel from '../../views/personal/OrdersPanel.vue'
import WalletPanel from '../../views/personal/WalletPanel.vue'
import GiftCardPanel from '../../views/personal/GiftCardPanel.vue'
import AffiliatePanel from '../../views/personal/AffiliatePanel.vue'
import ApiPanel from '../../views/personal/ApiPanel.vue'
import Ji8BrandIcon from './components/Ji8BrandIcon.vue'
import { formatMoney } from './utils/price'
import './styles/account.css'

/**
 * ji8 个人中心（仿 ai.bahk.cn/account.html）：
 * 分组侧栏 + MEMBER CENTER 标题条 + 概览（欢迎卡 / 快捷入口 / 资金统计 / 个人资料 / 最近购买）。
 * 非概览分区直接复用 classic 的各 Panel（已被 ji8 令牌重新着色）。
 */
const { t } = useI18n()
const userAuthStore = useUserAuthStore()
const appStore = useAppStore()
const { getLocalizedText, siteCurrency } = useLocalized()

const props = withDefaults(defineProps<{ section?: PersonalSection }>(), {
  section: 'overview',
})

const {
  userProfileStore, canAccessResellerConsole, currentSection, globalAlert,
  displayInitial, switchSection, statusLabel, statusVariant, levelName, discountText,
} = usePersonalCenter(() => props.section)

interface NavItem { key: PersonalSection; label: string; icon: Component }

const affiliateEnabled = computed(() => appStore.config?.affiliate?.enabled === true)

const navGroups = computed(() => {
  const promo: NavItem[] = []
  if (affiliateEnabled.value) promo.push({ key: 'affiliate', label: t('personalCenter.tabs.affiliate'), icon: Megaphone })
  if (canAccessResellerConsole.value) promo.push({ key: 'reseller', label: t('personalCenter.tabs.reseller'), icon: Banknote })
  promo.push({ key: 'api', label: t('personalCenter.tabs.api'), icon: Key })
  return [
    {
      key: 'account',
      label: t('ji8.account.group.account'),
      items: [
        { key: 'overview', label: t('ji8.account.home'), icon: House },
        { key: 'profile', label: t('personalCenter.tabs.profile'), icon: UserCircle },
        { key: 'security', label: t('personalCenter.tabs.security'), icon: ShieldCheck },
      ] as NavItem[],
    },
    {
      key: 'finance',
      label: t('ji8.account.group.finance'),
      items: [
        { key: 'wallet', label: t('personalCenter.tabs.wallet'), icon: Wallet },
        { key: 'orders', label: t('personalCenter.tabs.orders'), icon: ShoppingBag },
        { key: 'giftCard', label: t('personalCenter.tabs.giftCard'), icon: Gift },
      ] as NavItem[],
    },
    { key: 'promo', label: t('ji8.account.group.promo'), items: promo },
  ]
})

const flatNavItems = computed(() => navGroups.value.flatMap((g) => g.items))

const currentTitle = computed(() => {
  const item = flatNavItems.value.find((i) => i.key === currentSection.value)
  return item?.label || t('ji8.account.home')
})

// ---- 概览数据：钱包余额、推广数据（与 Panel 自身加载互不影响） ----
const walletBalance = ref<string | null>(null)
const affiliate = ref<AffiliateDashboardData | null>(null)

const loadOverviewExtras = async () => {
  const tasks: Promise<unknown>[] = [
    walletAPI.account().then((res) => { walletBalance.value = res.data.data?.balance ?? null }).catch(() => undefined),
    userProfileStore.loadRecentLoginLogs(10),
  ]
  if (appStore.config?.affiliate?.enabled !== false) {
    tasks.push(affiliateAPI.dashboard().then((res) => { affiliate.value = res.data.data || null }).catch(() => undefined))
  }
  await Promise.all(tasks)
}

onMounted(() => { loadOverviewExtras() })

// 与参考站一致：YYYY-MM-DD HH:mm:ss（本地时区）
const pad = (n: number) => String(n).padStart(2, '0')
const formatDate = (raw?: string) => {
  if (!raw) return '-'
  const d = new Date(raw)
  if (Number.isNaN(d.getTime())) return raw
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
}

const money = (amount: unknown, currency?: string) => formatMoney(amount, currency || siteCurrency.value)

const copyInvite = async () => {
  const code = affiliate.value?.affiliate_code
  if (!code) return
  const path = affiliate.value?.promotion_path || `/?aff=${code}`
  try {
    await copyText(`${window.location.origin}${path}`)
    toast.success(t('ji8.account.copied'))
  } catch {
    toast.error(t('ji8.account.copyFailed'))
  }
}

const quickItems = computed(() => {
  const items: Array<{ key: PersonalSection; title: string; desc: string; icon: Component }> = [
    { key: 'wallet', title: t('ji8.account.quick.wallet'), desc: t('ji8.account.quick.walletDesc'), icon: Wallet },
    { key: 'orders', title: t('ji8.account.quick.orders'), desc: t('ji8.account.quick.ordersDesc'), icon: ReceiptText },
    { key: 'giftCard', title: t('ji8.account.quick.giftCard'), desc: t('ji8.account.quick.giftCardDesc'), icon: Gift },
    { key: 'security', title: t('ji8.account.quick.security'), desc: t('ji8.account.quick.securityDesc'), icon: ShieldCheck },
  ]
  if (affiliateEnabled.value) {
    items.push({
      key: 'affiliate',
      title: t('ji8.account.quick.affiliate'),
      desc: t('ji8.account.quick.affiliateDesc', { n: affiliate.value?.valid_order_count ?? 0 }),
      icon: Megaphone,
    })
  }
  return items
})

const statItems = computed(() => {
  const items = [
    { key: 'balance', label: t('ji8.account.stat.balance'), value: walletBalance.value === null ? '—' : money(walletBalance.value), hint: t('ji8.account.stat.balanceHint'), icon: CircleDollarSign },
    { key: 'recharged', label: t('ji8.account.stat.recharged'), value: money(userProfileStore.profile?.total_recharged ?? 0), hint: t('ji8.account.stat.rechargedHint'), icon: Coins },
    { key: 'spent', label: t('ji8.account.stat.spent'), value: money(userProfileStore.profile?.total_spent ?? 0), hint: t('ji8.account.stat.spentHint', { n: userProfileStore.ordersTotal }), icon: ShoppingBag },
    { key: 'level', label: t('ji8.account.stat.level'), value: levelName(userProfileStore.currentLevel), hint: discountText.value, icon: Crown },
  ]
  if (affiliateEnabled.value) {
    items.push({ key: 'commission', label: t('ji8.account.stat.commission'), value: money(affiliate.value?.available_commission ?? 0), hint: t('ji8.account.stat.commissionHint', { amount: money(affiliate.value?.pending_commission ?? 0) }), icon: Banknote })
  }
  return items
})

const successLogins = computed(() => userProfileStore.recentLoginLogs.filter((log) => log.status === 'success'))

const profileRows = computed(() => {
  const p = userProfileStore.profile
  const current = successLogins.value[0]
  const previous = successLogins.value[1]
  return [
    { label: t('ji8.account.info.account'), value: p?.email || '-' },
    { label: t('ji8.account.info.nickname'), value: p?.nickname || '-' },
    { label: t('ji8.account.info.email'), value: p?.email_verified_at ? t('personalCenter.overview.emailVerified') : t('personalCenter.overview.emailUnverified') },
    { label: t('ji8.account.info.level'), value: `${levelName(userProfileStore.currentLevel)} · ${discountText.value}` },
    { label: t('ji8.account.info.currentIp'), value: current?.client_ip || '-' },
    { label: t('ji8.account.info.currentTime'), value: formatDate(current?.created_at) },
    { label: t('ji8.account.info.lastIp'), value: previous?.client_ip || '-' },
    { label: t('ji8.account.info.lastTime'), value: formatDate(previous?.created_at) },
  ]
})

interface OrderItemLite { title?: Record<string, string>; quantity?: number }

const recentRows = computed(() =>
  userProfileStore.recentOrders.map((order) => {
    const items = ((order as { items?: OrderItemLite[] }).items || []) as OrderItemLite[]
    const first = items[0]
    const title = first ? String(getLocalizedText(first.title) || '') : ''
    const extra = items.length > 1 ? t('ji8.account.moreItems', { n: items.length - 1 }) : ''
    const quantity = items.reduce((sum, item) => sum + (Number(item.quantity) || 0), 0)
    return { ...order, title: title + extra, quantity: quantity || '-' }
  }),
)
</script>
