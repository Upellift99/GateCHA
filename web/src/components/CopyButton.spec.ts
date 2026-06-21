import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import CopyButton from './CopyButton.vue'

describe('CopyButton', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    Object.defineProperty(navigator, 'clipboard', {
      value: { writeText: vi.fn() },
      configurable: true,
    })
  })

  it('uses the label as accessible name and title when idle', () => {
    const wrapper = mount(CopyButton, { props: { value: 'gk_x', label: 'Copy key ID' } })
    const btn = wrapper.get('button')
    expect(btn.attributes('aria-label')).toBe('Copy key ID')
    expect(btn.attributes('title')).toBe('Copy key ID')
    expect(wrapper.text()).not.toContain('Copied!')
  })

  it('copies the value and shows a "Copied!" tooltip on click', async () => {
    const wrapper = mount(CopyButton, { props: { value: 'gk_secret' } })
    await wrapper.get('button').trigger('click')

    expect(navigator.clipboard.writeText).toHaveBeenCalledWith('gk_secret')
    expect(wrapper.text()).toContain('Copied!')
    expect(wrapper.get('button').attributes('aria-label')).toBe('Copied!')
  })

  it('clears the tooltip after the timeout', async () => {
    const wrapper = mount(CopyButton, { props: { value: 'v' } })
    await wrapper.get('button').trigger('click')
    expect(wrapper.text()).toContain('Copied!')

    vi.advanceTimersByTime(2000)
    await wrapper.vm.$nextTick()
    expect(wrapper.text()).not.toContain('Copied!')
  })

  it('applies the overlay variant styling', () => {
    const wrapper = mount(CopyButton, { props: { value: 'v', variant: 'overlay' } })
    expect(wrapper.get('button').classes()).toContain('bg-slate-700/80')
  })
})
