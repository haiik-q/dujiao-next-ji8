import { computed, type Component } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  ShoppingBag,
  ClipboardList,
  BookOpen,
  Megaphone,
  Info,
  Wallet,
  Gift,
  Crown,
  Ticket,
  ShoppingCart,
  User,
} from 'lucide-vue-next'
import { useAppStore } from '../../../stores/app'
import { useUserAuthStore } from '../../../stores/userAuth'
import { useNavConfig } from '../../../composables/useNavConfig'

/**
 * ji8 模板导航项。与 useNavConfig 的 NavItem 兼容（target 可选），
 * 额外带 exact（仅首页精确匹配）与 tone（顶栏促销 pill 的视觉变体）。
 */
export interface Ji8NavItem {
  key: string
  path: string
  label: string
  icon: Component
  /** route 走 <RouterLink>，link 走原生 <a> */
  type: 'route' | 'link'
  target?: string
  exact?: boolean
  /** promo=无边框渐变（充值优惠）、growth=带边框渐变（推广领佣金）、gold=金字（成为分销） */
  tone?: 'promo' | 'growth' | 'gold'
}

/**
 * ji8 模板的导航数据源（D6 语义映射）：
 * - 充值优惠 → /me/wallet（wallet_recharge_channel_ids 非空）
 * - 推广领佣金 → /me/affiliate（affiliate.enabled）
 * - 成为分销 → /reseller（canAccessResellerConsole）
 * - 提交工单/客服 → contact.telegram || contact.whatsapp（外链）
 * - 查订单 → 登录 /me/orders，否则 /guest/orders
 * - 个人中心 → 登录 /me，否则 /auth/login
 * 显隐全部由 appStore.config 决定，不硬编码。
 */
export function useJi8Nav() {
  const appStore = useAppStore()
  const userAuthStore = useUserAuthStore()
  const { t } = useI18n()
  const { blogEnabled, noticeEnabled, aboutEnabled, customNavItems } = useNavConfig()

  const ordersPath = computed(() => (userAuthStore.isAuthenticated ? '/me/orders' : '/guest/orders'))
  const accountPath = computed(() => (userAuthStore.isAuthenticated ? '/me' : '/auth/login'))
  const supportLink = computed(() =>
    String(appStore.config?.contact?.telegram || appStore.config?.contact?.whatsapp || '').trim(),
  )

  /** 顶栏促销 pill（桌面全部显示；移动端只显示第一枚，D16） */
  const promoItems = computed<Ji8NavItem[]>(() => {
    const items: Ji8NavItem[] = []
    const ids = appStore.config?.wallet_recharge_channel_ids
    if (Array.isArray(ids) && ids.length) {
      items.push({ key: 'recharge', path: '/me/wallet', label: t('ji8.nav.recharge'), icon: Wallet, type: 'route', tone: 'promo' })
    }
    if (appStore.config?.affiliate?.enabled === true) {
      items.push({ key: 'affiliate', path: '/me/affiliate', label: t('ji8.nav.affiliate'), icon: Gift, type: 'route', tone: 'growth' })
    }
    if (appStore.canAccessResellerConsole) {
      items.push({ key: 'reseller', path: '/reseller', label: t('ji8.nav.reseller'), icon: Crown, type: 'route', tone: 'gold' })
    }
    return items
  })

  /** 桌面顶栏主导航 */
  const mainItems = computed<Ji8NavItem[]>(() => {
    const items: Ji8NavItem[] = [
      { key: 'shop', path: '/', label: t('ji8.nav.shop'), icon: ShoppingBag, type: 'route', exact: true },
    ]
    items.push({ key: 'orders', path: ordersPath.value, label: t('ji8.nav.orders'), icon: ClipboardList, type: 'route' })
    if (blogEnabled.value) {
      items.push({ key: 'tutorials', path: '/blog', label: t('ji8.nav.tutorials'), icon: BookOpen, type: 'route' })
    }
    if (supportLink.value) {
      items.push({ key: 'support', path: supportLink.value, label: t('ji8.nav.support'), icon: Ticket, type: 'link', target: '_blank' })
    }
    if (noticeEnabled.value) {
      items.push({ key: 'notice', path: '/notice', label: t('nav.notice'), icon: Megaphone, type: 'route' })
    }
    if (aboutEnabled.value) {
      items.push({ key: 'about', path: '/about', label: t('nav.about'), icon: Info, type: 'route' })
    }
    customNavItems.value.forEach((c) => {
      items.push({ key: c.key, path: c.path, label: c.label, icon: c.icon, type: c.type, target: c.target })
    })
    items.push({ key: 'me', path: accountPath.value, label: t('navbar.personalCenter'), icon: User, type: 'route' })
    return items
  })

  /** 移动端浮动底栏（4–5 项） */
  const bottomItems = computed<Ji8NavItem[]>(() => [
    { key: 'shop', path: '/', label: t('ji8.nav.shop'), icon: ShoppingBag, type: 'route', exact: true },
    { key: 'orders', path: ordersPath.value, label: t('ji8.nav.orders'), icon: ClipboardList, type: 'route' },
    ...(blogEnabled.value
      ? [{ key: 'tutorials', path: '/blog', label: t('ji8.nav.tutorials'), icon: BookOpen, type: 'route' as const }]
      : []),
    { key: 'cart', path: '/cart', label: t('bottomNav.cart'), icon: ShoppingCart, type: 'route' },
    { key: 'me', path: accountPath.value, label: t('bottomNav.me'), icon: User, type: 'route' },
  ])

  /** 按 key 判活（G8）：避免已登录时 /me 与 /me/orders 在 /me/orders 同时高亮 */
  const isActive = (item: Ji8NavItem, path: string): boolean => {
    if (item.type === 'link') return false
    if (item.exact || item.path === '/') return path === '/'
    if (item.key === 'me') {
      return path === '/me' || (path.startsWith('/me/') && !path.startsWith('/me/orders')) || path.startsWith('/auth')
    }
    return path.startsWith(item.path)
  }

  return {
    ordersPath,
    accountPath,
    supportLink,
    promoItems,
    mainItems,
    bottomItems,
    isActive,
    blogEnabled,
    noticeEnabled,
    aboutEnabled,
  }
}
