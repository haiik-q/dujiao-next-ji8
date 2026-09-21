/**
 * ji8 · 本地 TOTP（RFC 6238）工具：Base32 解码、验证码计算与「任意粘贴内容 → 密钥候选」解析。
 * 纯函数、零依赖；仅用 Web Crypto（crypto.subtle HMAC），浏览器与 Node ≥ 20 均可运行。
 */

export type TotpAlgorithm = 'SHA-1' | 'SHA-256' | 'SHA-512'
export type TotpDigits = 6 | 7 | 8
export type TotpSource = 'otpauth' | 'url' | 'base32' | 'delivery'

export interface TotpCandidate {
  /** 规范化后的 Base32 密钥（大写、无分隔符、无 = 填充） */
  secret: string
  algorithm: TotpAlgorithm
  digits: TotpDigits
  period: number
  /** 账号 / 发行方（otpauth）或卡密里的邮箱、手机号 */
  label?: string
  source: TotpSource
}

export interface TotpOptions {
  secret: string
  algorithm?: TotpAlgorithm
  digits?: TotpDigits
  period?: number
  /** 毫秒时间戳，缺省 Date.now() */
  timestamp?: number
}

export const TOTP_DEFAULT_ALGORITHM: TotpAlgorithm = 'SHA-1'
export const TOTP_DEFAULT_DIGITS: TotpDigits = 6
export const TOTP_DEFAULT_PERIOD = 30

const MIN_SECRET_CHARS = 16
const MAX_SECRET_CHARS = 128
const BASE32_ALPHABET = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ234567'

/* ------------------------------------------------------------------ */
/* Base32                                                              */
/* ------------------------------------------------------------------ */

const ZERO_WIDTH_RE = /[\u200B-\u200D\u2060\uFEFF]/g
/** Base32 分组常见分隔符：空白、点、下划线、各种连字符 */
const SEPARATOR_RE = /[\s._\-\u2010-\u2015\u2212]+/g

/** 大写、去分隔符与零宽字符、去尾部 = 填充；不做字符集校验 */
export const normalizeBase32 = (input: string): string =>
  String(input ?? '')
    .replace(ZERO_WIDTH_RE, '')
    .replace(SEPARATOR_RE, '')
    .replace(/=+$/, '')
    .toUpperCase()

/**
 * 宽容的 Base32 解码：接受小写、空格 / 点 / 连字符分组、尾部 = 填充。
 * 非 Base32 字符或空输入抛出 Error。
 */
export function base32Decode(input: string): Uint8Array<ArrayBuffer> {
  const normalized = normalizeBase32(input)
  if (!normalized) throw new Error('empty base32 input')
  const bytes = new Uint8Array(Math.floor((normalized.length * 5) / 8))
  let buffer = 0
  let bits = 0
  let cursor = 0
  for (const char of normalized) {
    const value = BASE32_ALPHABET.indexOf(char)
    if (value < 0) throw new Error(`invalid base32 character: ${char}`)
    buffer = (buffer << 5) | value
    bits += 5
    if (bits >= 8) {
      bits -= 8
      bytes[cursor++] = (buffer >>> bits) & 0xff
      buffer &= (1 << bits) - 1
    }
  }
  return bytes
}

/** 是否为可用作 TOTP 密钥的 Base32 串（规范化后 16–128 字符且可解码） */
export const isBase32Secret = (input: string, minChars: number = MIN_SECRET_CHARS): boolean => {
  const normalized = normalizeBase32(input)
  if (normalized.length < minChars || normalized.length > MAX_SECRET_CHARS) return false
  try {
    base32Decode(normalized)
    return true
  } catch {
    return false
  }
}

/* ------------------------------------------------------------------ */
/* 参数规范化                                                          */
/* ------------------------------------------------------------------ */

export const normalizeAlgorithm = (value: unknown): TotpAlgorithm => {
  const compact = String(value ?? '').toUpperCase().replace(/[^A-Z0-9]/g, '')
  if (compact === 'SHA256') return 'SHA-256'
  if (compact === 'SHA512') return 'SHA-512'
  return TOTP_DEFAULT_ALGORITHM
}

export const normalizeDigits = (value: unknown): TotpDigits => {
  const digits = Number(value)
  return digits === 7 || digits === 8 ? digits : TOTP_DEFAULT_DIGITS
}

export const normalizePeriod = (value: unknown): number => {
  const period = Number(value)
  return Number.isInteger(period) && period >= 10 && period <= 300 ? period : TOTP_DEFAULT_PERIOD
}

/* ------------------------------------------------------------------ */
/* TOTP                                                                */
/* ------------------------------------------------------------------ */

const counterToBytes = (counter: number): Uint8Array<ArrayBuffer> => {
  const bytes = new Uint8Array(8)
  const view = new DataView(bytes.buffer)
  const safe = Math.max(0, Math.floor(counter))
  view.setUint32(0, Math.floor(safe / 0x1_0000_0000))
  view.setUint32(4, safe >>> 0)
  return bytes
}

/** 当前周期序号 */
export const totpCounter = (period: number = TOTP_DEFAULT_PERIOD, timestamp: number = Date.now()): number =>
  Math.floor(timestamp / 1000 / period)

/** 当前周期剩余秒数（1 … period） */
export const totpRemainingSeconds = (period: number = TOTP_DEFAULT_PERIOD, timestamp: number = Date.now()): number => {
  const elapsed = (timestamp / 1000) % period
  return Math.min(period, Math.max(1, Math.ceil(period - elapsed)))
}

/** 当前周期剩余进度（0 … 1），用于倒计时环 */
export const totpProgress = (period: number = TOTP_DEFAULT_PERIOD, timestamp: number = Date.now()): number =>
  Math.min(1, Math.max(0, totpRemainingSeconds(period, timestamp) / period))

/** 是否可在当前环境做本地计算 */
export const isTotpSupported = (): boolean =>
  typeof globalThis.crypto !== 'undefined' && typeof globalThis.crypto.subtle?.sign === 'function'

/** RFC 6238：HMAC(secret, counter) → 动态截断 → 十进制验证码 */
export async function generateTotp(options: TotpOptions): Promise<string> {
  const subtle = globalThis.crypto?.subtle
  if (!subtle) throw new Error('Web Crypto is not available')
  const keyBytes = base32Decode(options.secret)
  if (keyBytes.length === 0) throw new Error('secret is too short')
  const algorithm = normalizeAlgorithm(options.algorithm)
  const digits = normalizeDigits(options.digits)
  const period = normalizePeriod(options.period)
  const timestamp = typeof options.timestamp === 'number' ? options.timestamp : Date.now()

  const key = await subtle.importKey('raw', keyBytes, { name: 'HMAC', hash: { name: algorithm } }, false, ['sign'])
  const mac = new Uint8Array(await subtle.sign('HMAC', key, counterToBytes(totpCounter(period, timestamp))))
  const offset = mac[mac.length - 1]! & 0x0f
  const binary =
    (((mac[offset]! & 0x7f) << 24) | (mac[offset + 1]! << 16) | (mac[offset + 2]! << 8) | mac[offset + 3]!) >>> 0
  return String(binary % 10 ** digits).padStart(digits, '0')
}

/** "123456" → "123 456"；8 位按 4+4 */
export const formatTotpCode = (code: string): string => {
  const value = String(code ?? '')
  if (!value) return ''
  const split = value.length === 8 ? 4 : Math.ceil(value.length / 2)
  return `${value.slice(0, split)} ${value.slice(split)}`
}

/** 密钥脱敏预览：JBSW••••••3PXP */
export const maskSecret = (secret: string): string => {
  const normalized = normalizeBase32(secret)
  if (!normalized) return ''
  if (normalized.length <= 10) return `${normalized.slice(0, 2)}••••${normalized.slice(-2)}`
  return `${normalized.slice(0, 4)}••••••${normalized.slice(-4)}`
}

/* ------------------------------------------------------------------ */
/* 输入解析                                                            */
/* ------------------------------------------------------------------ */

/** 链接（otpauth / http(s)）；不吞 CJK 与全角标点，避免把紧贴的中文说明当作路径 */
const URL_RE = /(?:otpauth:\/\/|https?:\/\/)[^\s<>"'`\u3000-\u303F\u4E00-\u9FFF\uFF00-\uFFEF]+/gi
/** 连续 16–128 位 Base32 字符（大小写均可）；前后不能再接 Base32 字符或 @（排除邮箱） */
const PLAIN_TOKEN_RE = /(^|[^A-Za-z2-7@])([A-Za-z2-7]{16,128})=*(?=$|[^A-Za-z2-7=@])/g
/** 4–8 位\u4E00组、以空白 / 点 / 下划线 / 连字符分隔的分组 Base32（末组可短） */
const GROUPED_TOKEN_RE =
  /(^|[^A-Za-z2-7@])((?:[A-Za-z2-7]{4,8}[ \t._\-\u2010-\u2015\u2212]+){1,31}[A-Za-z2-7]{2,8})=*(?=$|[^A-Za-z2-7=@])/g
/** 整行只由 Base32 字符与分隔符组成 */
const WHOLE_LINE_RE = /^[A-Za-z2-7=]+(?:[\s._\-\u2010-\u2015\u2212]+[A-Za-z2-7=]+)*$/
/** XXXX-XXXX-XXXX-XXXX 备用码 / 卡密序列号，不是 TOTP 密钥 */
const BACKUP_CODE_RE = /^(?:[A-Za-z0-9]{4}[\-\u2010-\u2015\u2212]){3}[A-Za-z0-9]{4}$/
/** 卡密字段分隔：两个以上连字符、竖线、冒号、分号、逗号（含全角） */
const FIELD_SEPARATOR_RE = /\s*(?:[\-\u2010-\u2015\u2212]{2,}|[|｜丨│¦:：;；,，])\s*/
/** 邮箱或手机号样式的账号字段 */
const ACCOUNT_RE = /^(?:[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}|\+?\d[\d\s-]{5,19}\d)$/
/** 密钥标签（2FA 密钥 / TOTP secret / MFA key / 谷歌验证器 …） */
const LABEL_RE =
  /(?:2[\s_-]*fa|two[\s_-]*factor|totp|mfa|otp|one[\s_-]*time[\s_-]*password|google[\s_-]*auth\w*|authenticator|谷歌(?:身份)?验证(?:器)?|动态口令|验证密钥|二次验证|双重验证|两步验证|二步验证)(?:[\s_-]*(?:secret|key|code|seed|密钥|密鑰|秘钥|秘鑰))?/gi
/** 2FA 链接路径里的标记段：其后\u4E00段即密钥 */
const PATH_MARKERS = new Set(['2fa', 'totp', 'otp', 'secret', 'key', 'seed', 'mfa'])
const QUERY_KEYS = new Set(['secret', 'key', 'totp', 'otp', 'seed', 'mfa', 'twofa', 'base32'])
const TWO_FA_HOSTS = new Set(['2fa.cn', '2fa.fun', '2fa.run', '2fa.pub', '2fa.live'])
const TRAILING_PUNCT_RE = /[，。！？、；：）】》〉)\]}>.,!?;:]+$/

const safeDecode = (value: string): string => {
  try {
    return decodeURIComponent(value)
  } catch {
    return value
  }
}

/** 反复 URL 解码（最多 3 层），返回所有中间形态 */
const decodeVariants = (value: string): string[] => {
  const variants = [value]
  let current = value
  for (let round = 0; round < 3; round += 1) {
    const next = safeDecode(current)
    if (next === current) break
    variants.push(next)
    current = next
  }
  return variants
}

const isBackupCode = (raw: string): boolean => BACKUP_CODE_RE.test(raw.trim())
const isAccountLike = (field: string): boolean => ACCOUNT_RE.test(field.trim())
const hasDigit = (raw: string): boolean => /[2-7]/.test(raw)
const isAllUpper = (raw: string): boolean => raw === raw.toUpperCase()
const stripPadding = (raw: string): string => raw.trim().replace(/=+$/, '')
const hasSeparators = (raw: string): boolean => /[\s._\-\u2010-\u2015\u2212]/.test(raw.trim())

interface RawToken {
  raw: string
  /** 该 token 是否就是参照文本（通常为整行）的全部内容 */
  whole: boolean
}

/**
 * 分组 token 里剔除前导的「异形」组：
 * `password jbsw y3dp ehpk 3pxp` → 取组长\u4E00致的尾段 `jbsw y3dp ehpk 3pxp`。
 * 合法分组密钥各组等长（末组可短）；找不到等长尾段则原样返回。
 */
const trimGroupedRun = (raw: string): string => {
  const groups = raw.split(/[ \t._\-\u2010-\u2015\u2212]+/).filter(Boolean)
  if (groups.length < 2) return raw
  for (let start = 0; start < groups.length - 1; start += 1) {
    const size = groups[start]!.length
    const body = groups.slice(start, -1)
    const tail = groups[groups.length - 1]!
    if (body.every((g) => g.length === size) && tail.length <= size) {
      return groups.slice(start).join(' ')
    }
  }
  return raw
}

/** 从\u4E00段文本里找出所有可能的 Base32 token（连续或分组）；reference 用来判断 token 是否占满整行 */
const collectTokens = (zone: string, reference: string = zone): RawToken[] => {
  const tokens: RawToken[] = []
  const wholeText = stripPadding(reference)
  PLAIN_TOKEN_RE.lastIndex = 0
  for (let match = PLAIN_TOKEN_RE.exec(zone); match; match = PLAIN_TOKEN_RE.exec(zone)) {
    const raw = match[2]!
    tokens.push({ raw, whole: wholeText === raw })
  }
  GROUPED_TOKEN_RE.lastIndex = 0
  for (let match = GROUPED_TOKEN_RE.exec(zone); match; match = GROUPED_TOKEN_RE.exec(zone)) {
    const found = match[2]!
    if (isBackupCode(found)) continue
    tokens.push({ raw: trimGroupedRun(found), whole: wholeText === found })
  }
  return tokens
}

interface CandidateExtras {
  algorithm?: unknown
  digits?: unknown
  period?: unknown
  label?: string
}

const makeCandidate = (secret: string, source: TotpSource, extra: CandidateExtras = {}): TotpCandidate => {
  const candidate: TotpCandidate = {
    secret,
    algorithm: normalizeAlgorithm(extra.algorithm),
    digits: normalizeDigits(extra.digits),
    period: normalizePeriod(extra.period),
    source,
  }
  const label = String(extra.label ?? '').trim()
  if (label) candidate.label = label
  return candidate
}

/** otpauth://totp/Issuer:account?secret=…&issuer=…&algorithm=…&digits=…&period=… */
const parseOtpauth = (rawUrl: string): TotpCandidate | null => {
  let url: URL
  try {
    url = new URL(rawUrl.replace(TRAILING_PUNCT_RE, ''))
  } catch {
    return null
  }
  if (url.protocol !== 'otpauth:' || url.hostname.toLowerCase() !== 'totp') return null
  const secret = normalizeBase32(url.searchParams.get('secret') ?? '')
  if (!isBase32Secret(secret, 8)) return null

  let issuer = (url.searchParams.get('issuer') ?? '').trim()
  let account = safeDecode(url.pathname.replace(/^\/+/, '')).trim()
  const colon = account.indexOf(':')
  if (colon > 0) {
    const prefix = account.slice(0, colon).trim()
    if (!issuer) issuer = prefix
    if (prefix.toLowerCase() === issuer.toLowerCase()) account = account.slice(colon + 1).trim()
  }
  const label = Array.from(new Set([issuer, account].filter(Boolean))).join(' · ')

  return makeCandidate(secret, 'otpauth', {
    algorithm: url.searchParams.get('algorithm'),
    digits: url.searchParams.get('digits'),
    period: url.searchParams.get('period'),
    label,
  })
}

/** 从\u4E00个 URL 片段值（query 值 / 路径段）里提取密钥：直接、多层解码、按卡密分隔再试 */
const secretFromUrlValue = (value: string, strictUpper: boolean): string => {
  for (const variant of decodeVariants(value)) {
    const direct = normalizeBase32(variant)
    if (isBase32Secret(direct) && !isBackupCode(variant) && (!strictUpper || isAllUpper(variant) || hasDigit(variant))) {
      return direct
    }
    if (strictUpper) continue
    for (const fragment of variant.split(FIELD_SEPARATOR_RE)) {
      for (const token of collectTokens(fragment)) {
        const normalized = normalizeBase32(token.raw)
        if (isBase32Secret(normalized)) return normalized
      }
    }
  }
  return ''
}

/**
 * http(s) 链接：
 * - 2fa.cn / 2fa.fun / 2fa.run 之类的专用站点：query、hash、任意路径段都可能带密钥；
 * - 其它站点：只信任 secret/key/totp/otp/seed/mfa/twofa/base32 这些 query 键、
 *   标记段（/2fa/、/totp/、/secret/ …）之后的路径段，以及形如大写 Base32 的路径段。
 */
const parseHttpUrl = (rawUrl: string): TotpCandidate | null => {
  let url: URL
  try {
    url = new URL(rawUrl.replace(TRAILING_PUNCT_RE, ''))
  } catch {
    return null
  }
  const host = url.hostname.toLowerCase().replace(/^www\./, '')
  const dedicated = TWO_FA_HOSTS.has(host) || /(^|\.)2fa\.[a-z]+$/.test(host)
  const values: Array<{ value: string; strict: boolean }> = []
  const pushValue = (value: string | null | undefined, strict: boolean) => {
    const trimmed = String(value ?? '').trim()
    if (trimmed) values.push({ value: trimmed, strict })
  }
  const collectParams = (params: URLSearchParams) => {
    params.forEach((value, key) => {
      if (QUERY_KEYS.has(key.trim().toLowerCase())) pushValue(value, false)
    })
  }

  collectParams(url.searchParams)
  const hash = url.hash.replace(/^#\/?/, '')
  if (hash) {
    const question = hash.indexOf('?')
    collectParams(new URLSearchParams(question >= 0 ? hash.slice(question + 1) : hash))
    if (dedicated) hash.split('/').forEach((segment) => pushValue(segment, false))
  }
  const segments = url.pathname.split('/').filter(Boolean)
  segments.forEach((segment, index) => {
    if (PATH_MARKERS.has(safeDecode(segment).trim().toLowerCase())) pushValue(segments[index + 1], false)
  })
  const reversed = [...segments].reverse()
  reversed.forEach((segment) => pushValue(segment, !dedicated))

  for (const { value, strict } of values) {
    const secret = secretFromUrlValue(value, strict)
    if (secret) return makeCandidate(secret, 'url')
  }
  return null
}

/** 拆卡密字段 */
const splitFields = (line: string): string[] =>
  line
    .split(FIELD_SEPARATOR_RE)
    .map((field) => field.trim())
    .filter(Boolean)

/**
 * 区域内 token → 规范化密钥。
 * strict 为真（无标签 / 账号上下文）时，拒绝「小写且不含数字、又不占满整行」的 token，
 * 避免把普通英文单词当密钥；reference 用于判断是否占满整行。
 */
const tokensToSecrets = (zone: string, strict: boolean, reference: string = zone): string[] => {
  const secrets: string[] = []
  for (const token of collectTokens(zone, reference)) {
    if (isBackupCode(token.raw)) continue
    if (strict && !token.whole && !hasDigit(token.raw) && !isAllUpper(token.raw)) continue
    const normalized = normalizeBase32(token.raw)
    if (isBase32Secret(normalized)) secrets.push(normalized)
  }
  return secrets
}

/**
 * 解析任意粘贴内容，返回按出现顺序、按密钥去重的 TOTP 候选：
 * otpauth 链接、2FA / 通用链接、带标签的行（2FA密钥: …）、
 * 卡密元组（邮箱--密码--密钥）、独立 Base32（含分组写法）。
 */
export function parseTotpInput(text: string): TotpCandidate[] {
  const source = String(text ?? '').replace(ZERO_WIDTH_RE, '').replace(/\r\n?/g, '\n')
  const candidates: TotpCandidate[] = []
  const bySecret = new Map<string, TotpCandidate>()
  const add = (candidate: TotpCandidate | null) => {
    if (!candidate) return
    const existing = bySecret.get(candidate.secret)
    if (existing) {
      if (!existing.label && candidate.label) existing.label = candidate.label
      return
    }
    bySecret.set(candidate.secret, candidate)
    candidates.push(candidate)
  }

  const lines = source.split('\n')
  const singleLine = lines.filter((l) => l.trim()).length === 1
  const consumed = new Set<number>()

  lines.forEach((originalLine, index) => {
    // 1. 链接：解析后从行内抹去，后续步骤不再重复扫描
    let line = originalLine
    URL_RE.lastIndex = 0
    for (let match = URL_RE.exec(originalLine); match; match = URL_RE.exec(originalLine)) {
      const rawUrl = match[0]
      add(rawUrl.toLowerCase().startsWith('otpauth:') ? parseOtpauth(rawUrl) : parseHttpUrl(rawUrl))
      line = line.replace(rawUrl, ' '.repeat(rawUrl.length))
    }
    const trimmed = line.trim()
    if (consumed.has(index) || !trimmed || isAccountLike(trimmed)) return

    const fields = splitFields(line)
    const accountIndex = fields.findIndex(isAccountLike)
    const contextLabel =
      (accountIndex >= 0 ? fields[accountIndex] : undefined) ||
      lines
        .slice(Math.max(0, index - 2), index)
        .reverse()
        .map((l) => splitFields(l).find(isAccountLike))
        .find(Boolean)

    // 2. 带标签的行：标签之后的内容优先，同行没有则看下\u4E00行
    LABEL_RE.lastIndex = 0
    let labelEnd = -1
    for (let match = LABEL_RE.exec(line); match; match = LABEL_RE.exec(line)) labelEnd = match.index + match[0].length
    if (labelEnd >= 0) {
      let found = tokensToSecrets(line.slice(labelEnd), false, line)
      const nextLine = lines[index + 1]
      if (!found.length && nextLine !== undefined) {
        found = tokensToSecrets(nextLine.replace(URL_RE, ' '), false)
        if (found.length) consumed.add(index + 1)
      }
      found.forEach((secret) => add(makeCandidate(secret, 'delivery', { label: contextLabel })))
      if (found.length) return
    }

    // 3. 卡密元组：邮箱/手机号 -- 密码 -- 密钥 …（跳过账号字段与紧随其后的密码字段）
    if (fields.length >= 2) {
      const skip = new Set<number>()
      if (accountIndex >= 0) {
        skip.add(accountIndex)
        if (fields.length >= 3 && accountIndex === 0) skip.add(1)
      }
      const found: string[] = []
      fields.forEach((field, fieldIndex) => {
        if (skip.has(fieldIndex)) return
        const words = field.split(/\s+/)
        if (words.length >= 3 && isAccountLike(words[0]!)) {
          found.push(...tokensToSecrets(words.slice(2).join(' '), false))
          return
        }
        found.push(...tokensToSecrets(field, accountIndex < 0, line))
      })
      found.forEach((secret) => add(makeCandidate(secret, accountIndex >= 0 ? 'delivery' : 'base32', { label: contextLabel })))
      if (found.length) return
    }

    // 4. 空白分隔的元组：邮箱 密码 密钥
    const words = trimmed.split(/\s+/)
    if (words.length >= 3 && isAccountLike(words[0]!)) {
      tokensToSecrets(words.slice(2).join(' '), false).forEach((secret) =>
        add(makeCandidate(secret, 'delivery', { label: words[0] })),
      )
      return
    }

    // 5. 整行就是\u4E00个（可能分组的）Base32 密钥；单行输入时连 XXXX-XXXX-XXXX-XXXX 也放行。
    //    分组写法要求各组等长且（含数字或全大写），避免把\u4E00句小写英文当密钥
    if (
      WHOLE_LINE_RE.test(trimmed) &&
      isBase32Secret(trimmed) &&
      (singleLine || !isBackupCode(trimmed)) &&
      (!hasSeparators(trimmed) || (trimGroupedRun(trimmed) === trimmed.replace(SEPARATOR_RE, ' ') && (hasDigit(trimmed) || isAllUpper(trimmed))))
    ) {
      add(makeCandidate(normalizeBase32(trimmed), contextLabel ? 'delivery' : 'base32', { label: contextLabel }))
      return
    }

    // 6. 行内独立 Base32
    tokensToSecrets(line, true).forEach((secret) =>
      add(makeCandidate(secret, contextLabel ? 'delivery' : 'base32', { label: contextLabel })),
    )
  })

  return candidates
}
