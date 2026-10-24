import { describe, it, expect } from 'vitest'
import { keyStatus, lastDays, todayUTC } from './status'

const day = (challenges_issued: number, verifications_ok: number, verifications_fail: number) => ({
  challenges_issued,
  verifications_ok,
  verifications_fail,
})

describe('keyStatus', () => {
  it('reports disabled keys whatever their traffic', () => {
    expect(keyStatus(false, day(100, 0, 100))).toBe('disabled')
  })

  it('calls a key idle when nothing happened today', () => {
    expect(keyStatus(true, undefined)).toBe('idle')
    expect(keyStatus(true, day(0, 0, 0))).toBe('idle')
  })

  it('needs both enough failures and a large share to call a key failing', () => {
    expect(keyStatus(true, day(40, 34, 6))).toBe('ok') // 6 failures but 15%
    expect(keyStatus(true, day(6, 0, 4))).toBe('ok') // 100% but only 4
    expect(keyStatus(true, day(74, 40, 31))).toBe('failing')
    expect(keyStatus(true, day(10, 20, 5))).toBe('failing') // exactly 5 and 20%
  })
})

describe('days', () => {
  it('lists UTC days oldest first, ending today', () => {
    const now = new Date('2026-10-24T01:30:00Z')
    expect(todayUTC(now)).toBe('2026-10-24')
    expect(lastDays(3, now)).toEqual(['2026-10-22', '2026-10-23', '2026-10-24'])
  })
})
