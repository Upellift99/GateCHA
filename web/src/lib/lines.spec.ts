import { describe, it, expect } from 'vitest'
import { lastDays, lineSlot, lineStatus, lineVars, todayUTC } from './lines'

const day = (challenges_issued: number, verifications_ok: number, verifications_fail: number) => ({
  challenges_issued,
  verifications_ok,
  verifications_fail,
})

describe('lines', () => {
  it('cycles eight colours while the number stays the identity', () => {
    expect([1, 8, 9, 16, 17].map(lineSlot)).toEqual([1, 8, 1, 8, 1])
    expect(lineVars(9).fill).toBe('var(--color-line-1)')
    expect(lineVars(3).stroke).toBe('var(--color-line-3-stroke)')
  })

  it('closes disabled keys whatever their traffic', () => {
    expect(lineStatus(false, day(100, 0, 100))).toBe('closed')
  })

  it('calls a line quiet when nothing happened today', () => {
    expect(lineStatus(true, undefined)).toBe('quiet')
    expect(lineStatus(true, day(0, 0, 0))).toBe('quiet')
  })

  it('needs both enough failures and a large share to call a line disturbed', () => {
    expect(lineStatus(true, day(40, 34, 6))).toBe('normal') // 6 failures but 15%
    expect(lineStatus(true, day(6, 0, 4))).toBe('normal') // 100% but only 4
    expect(lineStatus(true, day(74, 40, 31))).toBe('disturbed')
    expect(lineStatus(true, day(10, 20, 5))).toBe('disturbed') // exactly 5 and 20%
  })

  it('lists UTC days oldest first, ending today', () => {
    const now = new Date('2026-10-24T01:30:00Z')
    expect(todayUTC(now)).toBe('2026-10-24')
    expect(lastDays(3, now)).toEqual(['2026-10-22', '2026-10-23', '2026-10-24'])
  })
})
