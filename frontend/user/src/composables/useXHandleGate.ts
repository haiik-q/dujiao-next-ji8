import { reactive } from 'vue'
import { xCheckAPI } from '../api'

/**
 * ji8：X 会员下单前的赠礼资格检测结果，按用户名在各结算组件间共享。
 * 检测通过前结算按钮不可点；后端创建订单时还会再向 X 核实一次，前端只是第一道关。
 */
export type XHandleState =
  | { status: 'checking' }
  | { status: 'eligible'; handle: string; name: string; at: number }
  | { status: 'ineligible'; handle: string; name: string; reason: string; at: number }
  | { status: 'error'; message: string; at: number }

// 同一用户名 5 分钟内复用结果（后端自检接口另有 10 分钟缓存和按 IP 限流）
const REUSE_MS = 5 * 60 * 1000
const handlePattern = /^[A-Za-z0-9_]{1,15}$/

const states = reactive<Record<string, XHandleState>>({})
const inflight = new Map<string, Promise<XHandleState>>()

/** 接受 @name、name 或 x.com / twitter.com 主页链接，返回小写用户名；不合法返回空串。 */
export function normalizeXHandle(raw: unknown): string {
  let s = String(raw ?? '').trim()
  s = s.replace(/^https?:\/\//i, '').replace(/^(www\.|mobile\.)?(x|twitter)\.com\//i, '')
  s = (s.split(/[/?#]/)[0] ?? '').replace(/^@/, '')
  return handlePattern.test(s) ? s.toLowerCase() : ''
}

export function getXHandleState(raw: unknown): XHandleState | undefined {
  const key = normalizeXHandle(raw)
  return key ? states[key] : undefined
}

export function checkXHandle(raw: unknown, force = false): Promise<XHandleState | undefined> {
  const key = normalizeXHandle(raw)
  if (!key) return Promise.resolve(undefined)
  const current = states[key]
  if (!force && current && current.status !== 'checking' && current.status !== 'error' && Date.now() - current.at < REUSE_MS) {
    return Promise.resolve(current)
  }
  const pending = inflight.get(key)
  if (pending) return pending

  states[key] = { status: 'checking' }
  const task = xCheckAPI.check(key)
    .then((res: any): XHandleState => {
      const data = res?.data?.data || {}
      const handle = String(data.handle || key)
      const name = String(data.profile?.name || '')
      return data.eligible
        ? { status: 'eligible', handle, name, at: Date.now() }
        : { status: 'ineligible', handle, name, reason: String(data.reason || 'not_eligible'), at: Date.now() }
    })
    .catch((err: any): XHandleState => ({ status: 'error', message: String(err?.message || ''), at: Date.now() }))
    .then((state) => {
      states[key] = state
      inflight.delete(key)
      return state
    })
  inflight.set(key, task)
  return task
}
