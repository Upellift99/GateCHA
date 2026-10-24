// Today's state of an API key, read from its daily stats row.

import type { APIKey } from '../stores/apikeys'
import type { DailyStat } from '../stores/stats'

export type KeyStatus = 'ok' | 'idle' | 'failing' | 'disabled'

// A key is failing when today's failures are both numerous and a large share
// of its verifications: a couple of failed solves on a busy site is normal, a
// run of them is not. Stats are daily, so "today" (UTC) is the finest grain
// the server records.
export const FAILING_MIN_FAILURES = 5
export const FAILING_MIN_SHARE = 0.2

export function keyStatus(enabled: boolean, today: Pick<DailyStat, 'challenges_issued' | 'verifications_ok' | 'verifications_fail'> | undefined): KeyStatus {
  if (!enabled) return 'disabled'
  if (!today || (today.challenges_issued === 0 && today.verifications_ok + today.verifications_fail === 0)) return 'idle'
  const verified = today.verifications_ok + today.verifications_fail
  if (today.verifications_fail >= FAILING_MIN_FAILURES && verified > 0 && today.verifications_fail / verified >= FAILING_MIN_SHARE) {
    return 'failing'
  }
  return 'ok'
}

/** A failing key with today's counts, as the dashboard banner lists it. */
export interface FailingKey {
  key: APIKey
  failed: number
  verified: number
}

export const STATUS_LABEL: Record<KeyStatus, string> = {
  ok: 'OK',
  idle: 'No traffic',
  failing: 'Failing',
  disabled: 'Disabled',
}

/** Rank used to put failing keys first, disabled keys last. */
export const STATUS_RANK: Record<KeyStatus, number> = { failing: 0, ok: 1, idle: 2, disabled: 3 }

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
