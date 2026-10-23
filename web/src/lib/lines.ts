// "Info trafic": every API key is a numbered line of the network. The number is
// the key's database id, so it never changes and reads the same in the list,
// the detail page and the charts. Colours cycle over eight metro inks; the
// number, not the colour, is what identifies a line.

import type { DailyStat } from '../stores/stats'

export const LINE_COLOURS = 8

export function lineSlot(id: number): number {
  return ((Math.max(1, id) - 1) % LINE_COLOURS) + 1
}

/** CSS custom properties for a line's bullet fill, its number and its chart stroke. */
export function lineVars(id: number) {
  const n = lineSlot(id)
  return {
    fill: `var(--color-line-${n})`,
    ink: `var(--color-line-${n}-ink)`,
    stroke: `var(--color-line-${n}-stroke)`,
  }
}

export type LineStatus = 'normal' | 'quiet' | 'disturbed' | 'closed'

// A line is disturbed when today's failures are both numerous and a large
// share of its verifications: a couple of failed solves on a busy site is
// normal service, a run of them is not. Stats are daily, so "today" (UTC) is
// the finest grain the server records.
export const DISTURBED_MIN_FAILURES = 5
export const DISTURBED_MIN_SHARE = 0.2

export function lineStatus(enabled: boolean, today: Pick<DailyStat, 'challenges_issued' | 'verifications_ok' | 'verifications_fail'> | undefined): LineStatus {
  if (!enabled) return 'closed'
  if (!today || (today.challenges_issued === 0 && today.verifications_ok + today.verifications_fail === 0)) return 'quiet'
  const verified = today.verifications_ok + today.verifications_fail
  if (today.verifications_fail >= DISTURBED_MIN_FAILURES && verified > 0 && today.verifications_fail / verified >= DISTURBED_MIN_SHARE) {
    return 'disturbed'
  }
  return 'normal'
}

export const STATUS_LABEL: Record<LineStatus, string> = {
  normal: 'Normal service',
  quiet: 'No traffic today',
  disturbed: 'Disturbed',
  closed: 'Closed',
}

/** Rank used to put disturbed lines first, closed lines last. */
export const STATUS_RANK: Record<LineStatus, number> = { disturbed: 0, normal: 1, quiet: 2, closed: 3 }

export function todayUTC(now = new Date()): string {
  return new Date(Date.UTC(now.getUTCFullYear(), now.getUTCMonth(), now.getUTCDate())).toISOString().slice(0, 10)
}

/** The last n UTC calendar days, oldest first, ending today. */
export function lastDays(n: number, now = new Date()): string[] {
  return Array.from({ length: n }, (_, i) => {
    const d = new Date(Date.UTC(now.getUTCFullYear(), now.getUTCMonth(), now.getUTCDate() - (n - 1 - i)))
    return d.toISOString().slice(0, 10)
  })
}
