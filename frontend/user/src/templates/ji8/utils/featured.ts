/**
 * 首页「为你精选」：把商品按分类归成几个产品线，每条线一张大卡片 + 一幅 CSS 动态插画。
 *
 * 价格、套餐名、月数全部取自当前站点的商品数据（分销站加价后显示的也是分销价），
 * 插画类型只按分类 slug 匹配；不在下表里的分类不出精选卡，照常出现在下方全部商品里。
 */
export type FeaturedArt = 'chatgpt' | 'gemini' | 'x' | 'duolingo' | 'muse'

/** 顺序即卡片顺序（桌面端第一张占两列） */
const FEATURED_DEFS: { art: FeaturedArt; slugs: string[] }[] = [
  { art: 'chatgpt', slugs: ['chatgpt'] },
  { art: 'gemini', slugs: ['gemini'] },
  { art: 'x', slugs: ['twitter-x', 'twitter', 'x'] },
  { art: 'duolingo', slugs: ['duolingo'] },
  { art: 'muse', slugs: ['muse'] },
]

export interface FeaturedPlan {
  name: string
  price: number
}

export interface FeaturedGroup {
  art: FeaturedArt
  category: any
  products: any[]
  /** 在售商品里的最低价；全部售罄时取全部商品的最低价 */
  minPrice: number
  /** 价格不止一个时显示「起」 */
  priceVaries: boolean
  /** 标题里解析出的月数，去重升序（如 [3, 6]） */
  months: number[]
  /** ChatGPT 卡片叠放的套餐（按价格升序，最多 3 个） */
  plans: FeaturedPlan[]
  allSoldOut: boolean
}

const toNumber = (value: unknown) => {
  const n = Number(value)
  return Number.isFinite(n) ? n : 0
}

const PLAN_RE = /\b(Plus|Go|Team|Business|Pro\s*\d+\s*x?|Pro)\b/i
const MONTHS_RE = /(\d{1,2})\s*(?:个月|個月|months?)/i

export function buildFeaturedGroups(
  products: any[],
  getTitle: (product: any) => string,
  isSoldOut: (product: any) => boolean,
): FeaturedGroup[] {
  const groups: FeaturedGroup[] = []
  for (const def of FEATURED_DEFS) {
    const items = products.filter((p) => def.slugs.includes(String(p?.category?.slug || '').trim().toLowerCase()))
    if (!items.length) continue

    const onSale = items.filter((p) => !isSoldOut(p))
    const pricedPool = onSale.length ? onSale : items
    const prices = pricedPool.map((p) => toNumber(p?.price_amount)).filter((n) => n > 0)
    const months = [...new Set(items.map((p) => Number(MONTHS_RE.exec(getTitle(p))?.[1] || 0)).filter((n) => n > 0))]
      .sort((a, b) => a - b)

    const plans = [...pricedPool]
      .sort((a, b) => toNumber(a?.price_amount) - toNumber(b?.price_amount))
      .slice(0, 3)
      .map((p) => {
        const plan = PLAN_RE.exec(getTitle(p))?.[1]
        const name = plan ? plan.replace(/\s+/g, ' ').replace(/^pro/i, 'Pro').replace(/^plus/i, 'Plus') : getTitle(p).slice(0, 10)
        return { name, price: toNumber(p?.price_amount) }
      })

    groups.push({
      art: def.art,
      category: items[0].category,
      products: items,
      minPrice: prices.length ? Math.min(...prices) : 0,
      priceVaries: new Set(prices).size > 1,
      months,
      plans,
      allSoldOut: !onSale.length,
    })
  }
  return groups
}
