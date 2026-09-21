import { ref } from 'vue'
import { productAPI } from '../../../api'
import type { CategoryGroup } from '../../../utils/category'

/**
 * 分类商品计数（D15 / G6）。公开 API 没有 product_count 字段，
 * 只能按分类各发一次 page_size:1 的列表请求读 pagination.total。
 *
 * key 0 = 全部商品（无筛选）；模块级缓存，跨 Home / Products 页面复用，
 * 已请求过（含进行中）的分类不再重复请求。
 */
const cache = ref<Record<number, number>>({})
const requested = new Set<number>()

/** 同时进行的请求上限，避免分类很多时一次打出几十个请求 */
const CONCURRENCY = 4

const fetchCount = async (id: number): Promise<void> => {
  const params: Record<string, number> = { page: 1, page_size: 1 }
  if (id) params.category_id = id
  try {
    const res = await productAPI.list(params)
    cache.value = { ...cache.value, [id]: Number(res.data?.pagination?.total) || 0 }
  } catch {
    // 计数只是装饰信息，失败时允许下次 load 重试
    requested.delete(id)
  }
}

export function useCategoryCounts() {
  const counts = cache

  const load = async (groups: CategoryGroup[]): Promise<void> => {
    const ids = [0, ...groups.flatMap((g) => [g.id, ...g.children.map((c) => c.id)])]
    const pending = ids.filter((id) => Number.isFinite(id) && id >= 0 && !requested.has(id))
    if (!pending.length) return
    pending.forEach((id) => requested.add(id))

    const queue = pending.slice()
    const worker = async () => {
      while (queue.length) {
        const id = queue.shift()
        if (id === undefined) return
        await fetchCount(id)
      }
    }
    await Promise.allSettled(Array.from({ length: Math.min(CONCURRENCY, queue.length) }, worker))
  }

  return { counts, load }
}
