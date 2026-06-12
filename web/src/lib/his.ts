// Client-side collector for GateCHA's Human Interaction Signature (HIS).
//
// It observes interaction events and emits only privacy-preserving *aggregates*
// (counts, total pointer distance, durations, timing variance) — never raw
// coordinates, timestamps, or key contents. The server scores these in Monitor
// mode to estimate automation probability without ever blocking.

export interface HISSignals {
  duration_ms: number
  time_to_first_ms: number
  pointer_events: number
  pointer_distance: number
  scrolls: number
  touches: number
  keydowns: number
  key_interval_stdev_ms: number
}

export interface HISCollector {
  /** Snapshot the aggregates collected so far. */
  signals(): HISSignals
  /** Detach all listeners. Safe to call multiple times. */
  stop(): void
}

function stdevOfIntervals(times: number[]): number {
  if (times.length < 2) return 0
  const intervals: number[] = []
  for (let i = 1; i < times.length; i++) intervals.push(times[i] - times[i - 1])
  const mean = intervals.reduce((a, b) => a + b, 0) / intervals.length
  const variance = intervals.reduce((a, b) => a + (b - mean) ** 2, 0) / intervals.length
  return Math.sqrt(variance)
}

/**
 * Start collecting interaction signals on `target` (defaults to the document).
 * Call `signals()` to read aggregates and `stop()` to detach listeners.
 */
export function startHISCollector(target: EventTarget = document): HISCollector {
  const start = performance.now()
  let firstAt = -1
  let pointerEvents = 0
  let pointerDistance = 0
  let lastX = NaN
  let lastY = NaN
  let scrolls = 0
  let touches = 0
  let keydowns = 0
  const keyTimes: number[] = []

  const markFirst = () => {
    if (firstAt < 0) firstAt = performance.now() - start
  }

  const onPointer = (e: Event) => {
    markFirst()
    pointerEvents++
    const me = e as MouseEvent
    if (!Number.isNaN(lastX)) {
      pointerDistance += Math.hypot(me.clientX - lastX, me.clientY - lastY)
    }
    lastX = me.clientX
    lastY = me.clientY
  }
  const onScroll = () => {
    markFirst()
    scrolls++
  }
  const onTouch = () => {
    markFirst()
    touches++
  }
  const onKey = () => {
    markFirst()
    keydowns++
    keyTimes.push(performance.now())
  }

  const opts: AddEventListenerOptions = { passive: true }
  target.addEventListener('pointermove', onPointer, opts)
  target.addEventListener('scroll', onScroll, { passive: true, capture: true })
  target.addEventListener('touchstart', onTouch, opts)
  target.addEventListener('keydown', onKey, opts)

  let stopped = false
  const stop = () => {
    if (stopped) return
    stopped = true
    target.removeEventListener('pointermove', onPointer)
    target.removeEventListener('scroll', onScroll, true)
    target.removeEventListener('touchstart', onTouch)
    target.removeEventListener('keydown', onKey)
  }

  const signals = (): HISSignals => ({
    duration_ms: Math.round(performance.now() - start),
    time_to_first_ms: firstAt < 0 ? -1 : Math.round(firstAt),
    pointer_events: pointerEvents,
    pointer_distance: Math.round(pointerDistance),
    scrolls,
    touches,
    keydowns,
    key_interval_stdev_ms: Math.round(stdevOfIntervals(keyTimes)),
  })

  return { signals, stop }
}
