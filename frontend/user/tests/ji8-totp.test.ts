import test from 'node:test'
import assert from 'node:assert/strict'
import {
  base32Decode,
  formatTotpCode,
  generateTotp,
  maskSecret,
  normalizeBase32,
  parseTotpInput,
  totpProgress,
  totpRemainingSeconds,
} from '../src/templates/ji8/utils/totp.ts'

// RFC 6238 附录 B：ASCII "12345678901234567890" 的 Base32
const RFC_SECRET = 'GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ'
const RFC_SECRET_256 = 'GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQGEZA'

test('base32Decode tolerates lowercase, spaces, dashes and padding', () => {
  const expected = Array.from('12345678901234567890', (c) => c.charCodeAt(0))
  assert.deepEqual(Array.from(base32Decode(RFC_SECRET)), expected)
  assert.deepEqual(Array.from(base32Decode('gezd gnbv-gy3t.qojq gezd_gnbv gy3t qojq====')), expected)
  assert.deepEqual(Array.from(base32Decode('JBSWY3DPEHPK3PXP')), Array.from('Hello!\u00de\u00ad\u00be\u00ef', (c) => c.charCodeAt(0)))
  assert.throws(() => base32Decode('JBSW1Y3D'))
  assert.throws(() => base32Decode(''))
  assert.equal(normalizeBase32(' jbsw y3dp-ehpk 3pxp== '), 'JBSWY3DPEHPK3PXP')
})

test('generateTotp matches RFC 6238 vectors', async () => {
  assert.equal(await generateTotp({ secret: RFC_SECRET, algorithm: 'SHA-1', digits: 6, period: 30, timestamp: 59_000 }), '287082')
  assert.equal(await generateTotp({ secret: RFC_SECRET, algorithm: 'SHA-1', digits: 8, period: 30, timestamp: 59_000 }), '94287082')
  assert.equal(await generateTotp({ secret: RFC_SECRET, algorithm: 'SHA-1', digits: 8, period: 30, timestamp: 1_111_111_109_000 }), '07081804')
  assert.equal(await generateTotp({ secret: RFC_SECRET_256, algorithm: 'SHA-256', digits: 8, period: 30, timestamp: 59_000 }), '46119246')
  // 缺省参数：SHA-1 / 6 位 / 30 秒
  assert.equal(await generateTotp({ secret: 'gezd gnbv gy3t qojq gezd gnbv gy3t qojq', timestamp: 59_000 }), '287082')
})

test('period helpers', () => {
  assert.equal(totpRemainingSeconds(30, 0), 30)
  assert.equal(totpRemainingSeconds(30, 29_999), 1)
  assert.equal(totpRemainingSeconds(30, 59_000), 1)
  assert.equal(totpRemainingSeconds(30, 60_000), 30)
  assert.equal(totpProgress(30, 0), 1)
  assert.equal(totpProgress(30, 15_000), 0.5)
  assert.equal(formatTotpCode('287082'), '287 082')
  assert.equal(formatTotpCode('94287082'), '9428 7082')
  assert.equal(maskSecret('JBSWY3DPEHPK3PXP'), 'JBSW••••••3PXP')
})

test('parses otpauth URL with issuer, account and parameters', () => {
  const list = parseTotpInput(
    '请扫码或复制 otpauth://totp/GitHub:alice%40example.com?secret=jbswy3dpehpk3pxp&issuer=GitHub&algorithm=SHA256&digits=8&period=60 到验证器。',
  )
  assert.equal(list.length, 1)
  assert.deepEqual(list[0], {
    secret: 'JBSWY3DPEHPK3PXP',
    algorithm: 'SHA-256',
    digits: 8,
    period: 60,
    label: 'GitHub · alice@example.com',
    source: 'otpauth',
  })
})

test('parses 2FA links and generic URLs carrying a secret', () => {
  const twoFa = parseTotpInput('https://2fa.cn/JBSWY3DPEHPK3PXP')
  assert.equal(twoFa.length, 1)
  assert.equal(twoFa[0]?.secret, 'JBSWY3DPEHPK3PXP')
  assert.equal(twoFa[0]?.source, 'url')

  const query = parseTotpInput('看这里 https://2fa.fun/?secret=jbsw%20y3dp%20ehpk%203pxp&x=1 ，')
  assert.equal(query[0]?.secret, 'JBSWY3DPEHPK3PXP')

  const generic = parseTotpInput('https://example.com/account/2fa/GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ?tab=security')
  assert.equal(generic[0]?.secret, RFC_SECRET)

  const nested = parseTotpInput('https://2fa.run/#/JBSWY3DPEHPK3PXP')
  assert.equal(nested[0]?.secret, 'JBSWY3DPEHPK3PXP')

  // 普通链接里的小写 slug 不当密钥
  assert.equal(parseTotpInput('https://example.com/internationalization').length, 0)
})

test('parses plain and grouped base32 strings', () => {
  assert.equal(parseTotpInput('JBSWY3DPEHPK3PXP')[0]?.secret, 'JBSWY3DPEHPK3PXP')
  assert.equal(parseTotpInput('jbsw y3dp ehpk 3pxp')[0]?.secret, 'JBSWY3DPEHPK3PXP')
  assert.equal(parseTotpInput('  gezd-gnbv-gy3t-qojq-gezd-gnbv-gy3t-qojq  ')[0]?.secret, RFC_SECRET)
  assert.equal(parseTotpInput('jbsw y3dp ehpk 3pxp')[0]?.source, 'base32')
  // 尾部 = 填充
  assert.equal(parseTotpInput('JBSWY3DPEHPK3PXP====')[0]?.secret, 'JBSWY3DPEHPK3PXP')
})

test('parses labeled lines, including secret on the following line', () => {
  const sameLine = parseTotpInput('账号：alice@example.com\n密码：Passw0rd!\n2FA密钥: JBSWY3DPEHPK3PXP\n备注：勿泄露')
  assert.equal(sameLine.length, 1)
  assert.equal(sameLine[0]?.secret, 'JBSWY3DPEHPK3PXP')
  assert.equal(sameLine[0]?.label, 'alice@example.com')
  assert.equal(sameLine[0]?.source, 'delivery')

  const nextLine = parseTotpInput('TOTP secret =\njbsw y3dp ehpk 3pxp\nbackup: ABCD-EFGH-IJKL-MNOP')
  assert.equal(nextLine.length, 1)
  assert.equal(nextLine[0]?.secret, 'JBSWY3DPEHPK3PXP')
})

test('parses delivery tuples and uses the account as label', () => {
  const dashed = parseTotpInput('bob@example.com--Secret123--JBSWY3DPEHPK3PXP--recovery:ABCD-EFGH-IJKL-MNOP')
  assert.equal(dashed.length, 1)
  assert.deepEqual(dashed[0], {
    secret: 'JBSWY3DPEHPK3PXP',
    algorithm: 'SHA-1',
    digits: 6,
    period: 30,
    label: 'bob@example.com',
    source: 'delivery',
  })

  const piped = parseTotpInput('+8613800138000｜abcdefghijklmnop｜GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ')
  assert.equal(piped.length, 1)
  assert.equal(piped[0]?.secret, RFC_SECRET)
  assert.equal(piped[0]?.label, '+8613800138000')

  const spaced = parseTotpInput('carol@example.com hunter22 jbsw y3dp ehpk 3pxp')
  assert.equal(spaced.length, 1)
  assert.equal(spaced[0]?.label, 'carol@example.com')

  const multi = parseTotpInput(
    'a@x.com--pw--JBSWY3DPEHPK3PXP\nb@x.com--pw--GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ\nJBSWY3DPEHPK3PXP',
  )
  assert.deepEqual(
    multi.map((c) => [c.secret, c.label]),
    [
      ['JBSWY3DPEHPK3PXP', 'a@x.com'],
      [RFC_SECRET, 'b@x.com'],
    ],
  )
})

test('ignores backup codes, emails and prose', () => {
  assert.equal(parseTotpInput('ABCD-EFGH-IJKL-MNOP\nQRST-UVWX-YZAB-CDEF').length, 0)
  assert.equal(parseTotpInput('jbswy3dpehpk3pxp@gmail.com').length, 0)
  assert.equal(parseTotpInput('please send me the code before tomorrow').length, 0)
  assert.equal(parseTotpInput('').length, 0)
})
