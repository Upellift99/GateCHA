import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import HISCalibrationPanel from './HISCalibrationPanel.vue'

const mockApi = vi.hoisted(() => ({ get: vi.fn() }))
vi.mock('../lib/api', () => ({ default: mockApi }))

function calibration(samples: number) {
  return {
    samples,
    suspected: 8,
    threshold: 0.8,
    avg_duration_ms: 120,
    avg_pointer_events: 2,
    no_motion_pct: 75,
    score_histogram: Array.from({ length: 10 }, (_, i) => ({
      lo: i / 10,
      hi: (i + 1) / 10,
      count: i === 9 ? 8 : i,
    })),
  }
}

describe('HISCalibrationPanel', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    mockApi.get.mockReset()
  })

  it('renders the histogram and suspected rate when samples exist', async () => {
    mockApi.get.mockResolvedValue({ data: calibration(40) })
    const wrapper = mount(HISCalibrationPanel, { props: { keyId: 1 } })
    await flushPromises()

    expect(mockApi.get).toHaveBeenCalledWith(expect.stringContaining('/his/calibration'))
    expect(mockApi.get).toHaveBeenCalledWith(expect.stringContaining('key_id=1'))
    // 8/40 suspected = 20%
    expect(wrapper.text()).toContain('20')
    expect(wrapper.text()).toContain('threshold 0.80')
    // 10 histogram bars.
    expect(wrapper.findAll('.h-32 > div')).toHaveLength(10)
  })

  it('points at the collector when no sample and no observation arrived', async () => {
    mockApi.get.mockResolvedValue({ data: calibration(0) })
    const wrapper = mount(HISCalibrationPanel, { props: { keyId: 2, observations: 0 } })
    await flushPromises()

    // The panel only renders when sampling is on, so it must never suggest
    // enabling it: that was the whole defect.
    expect(wrapper.text()).not.toContain('Enable')
    expect(wrapper.text()).toContain('HIS sampling is on for this key')
    expect(wrapper.text()).toContain('globalThis.gatechaHIS.signals()')
  })

  it('says the histogram is merely young when signals are arriving', async () => {
    mockApi.get.mockResolvedValue({ data: calibration(0) })
    const wrapper = mount(HISCalibrationPanel, { props: { keyId: 2, observations: 120 } })
    await flushPromises()

    expect(wrapper.text()).toContain('120')
    expect(wrapper.text()).toContain('fills from now on')
    // Signals are getting through, so the integration hints would be misleading.
    expect(wrapper.text()).not.toContain('not getting through')
  })

  it('reports a failed request as a failed request, not as an empty histogram', async () => {
    mockApi.get.mockRejectedValue(new Error('401'))
    const wrapper = mount(HISCalibrationPanel, { props: { keyId: 2 } })
    await flushPromises()

    expect(wrapper.text()).toContain('Could not load')
    expect(wrapper.text()).not.toContain('HIS sampling is on for this key')
  })
})
