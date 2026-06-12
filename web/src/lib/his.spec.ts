import { describe, it, expect } from 'vitest'
import { startHISCollector } from './his'

describe('startHISCollector', () => {
  it('reports no interaction before any events', () => {
    const target = new EventTarget()
    const c = startHISCollector(target)
    const s = c.signals()
    expect(s.pointer_events).toBe(0)
    expect(s.time_to_first_ms).toBe(-1)
    expect(s.pointer_distance).toBe(0)
    expect(s.duration_ms).toBeGreaterThanOrEqual(0)
    c.stop()
  })

  it('aggregates pointer movement into distance and counts', () => {
    const target = new EventTarget()
    const c = startHISCollector(target)

    target.dispatchEvent(Object.assign(new Event('pointermove'), { clientX: 0, clientY: 0 }))
    target.dispatchEvent(Object.assign(new Event('pointermove'), { clientX: 3, clientY: 4 })) // +5px
    target.dispatchEvent(Object.assign(new Event('pointermove'), { clientX: 3, clientY: 4 })) // +0px

    const s = c.signals()
    expect(s.pointer_events).toBe(3)
    expect(s.pointer_distance).toBe(5)
    expect(s.time_to_first_ms).toBeGreaterThanOrEqual(0)
    c.stop()
  })

  it('counts scrolls, touches and keydowns', () => {
    const target = new EventTarget()
    const c = startHISCollector(target)

    target.dispatchEvent(new Event('scroll'))
    target.dispatchEvent(new Event('touchstart'))
    target.dispatchEvent(new Event('keydown'))
    target.dispatchEvent(new Event('keydown'))

    const s = c.signals()
    expect(s.scrolls).toBe(1)
    expect(s.touches).toBe(1)
    expect(s.keydowns).toBe(2)
    c.stop()
  })

  it('stops collecting after stop()', () => {
    const target = new EventTarget()
    const c = startHISCollector(target)
    c.stop()
    target.dispatchEvent(Object.assign(new Event('pointermove'), { clientX: 1, clientY: 1 }))
    expect(c.signals().pointer_events).toBe(0)
  })
})
