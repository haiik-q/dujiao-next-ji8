import { amountToCents, centsToAmount } from '../../../utils/money'

/** 常见币种符号；未收录的币种回退为 "CODE " 前缀（如 "THB 12.00"） */
const SYMBOLS: Record<string, string> = {
  CNY: '¥',
  USD: '$',
  EUR: '€',
  GBP: '£',
  JPY: '¥',
  HKD: 'HK$',
  TWD: 'NT$',
  SGD: 'S$',
  MYR: 'RM',
  KRW: '₩',
  AUD: 'A$',
  CAD: 'C$',
}

export const currencySymbol = (code: string): string => {
  const normalized = String(code || '').trim().toUpperCase()
  if (!normalized) return ''
  return SYMBOLS[normalized] ?? `${normalized} `
}

/**
 * 金额格式化：走 amountToCents/centsToAmount 保证两位小数且无浮点误差；
 * 无法解析的金额原样输出，空值输出 "-"。
 */
export const formatMoney = (amount: unknown, currency: string): string => {
  if (amount === null || amount === undefined || amount === '') return '-'
  const cents = amountToCents(amount)
  return `${currencySymbol(currency)}${cents === null ? String(amount) : centsToAmount(cents)}`
}
