import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import ToggleSwitch from './ToggleSwitch.vue'

function mountSwitch(props: Record<string, unknown> = {}) {
  return mount(ToggleSwitch, { props: { id: 'demoToggle', modelValue: false, ...props } })
}

describe('ToggleSwitch', () => {
  it('reflects the bound value on the checkbox', () => {
    const on = mountSwitch({ modelValue: true })
    expect((on.get('input').element as HTMLInputElement).checked).toBe(true)
    const off = mountSwitch({ modelValue: false })
    expect((off.get('input').element as HTMLInputElement).checked).toBe(false)
  })

  it('emits the new state on change', async () => {
    const wrapper = mountSwitch()
    await wrapper.get('input').setValue(true)
    expect(wrapper.emitted('update:modelValue')).toEqual([[true]])
  })

  it('carries the id so a label can be bound to it', () => {
    expect(mountSwitch().get('input').attributes('id')).toBe('demoToggle')
  })

  it('disables the checkbox when asked', () => {
    expect(mountSwitch({ disabled: true }).get('input').attributes('disabled')).toBeDefined()
  })

  // The checkbox is sr-only, so the pill is the only visible target and it has to
  // stay a sibling of the input: the peer-checked styling depends on it. See #146.
  it('draws the pill as a sibling of the checkbox', () => {
    const wrapper = mountSwitch()
    const pill = wrapper.get('#demoToggle ~ span')
    expect(pill.classes()).toContain('peer-checked:bg-teal-600')
  })

  // The knob is absolutely positioned, so it needs the pill itself as its
  // containing block; against any other box it drifts off the pill's midline.
  it('positions the knob against the pill, centred on its midline', () => {
    const pill = mountSwitch().get('#demoToggle ~ span')
    expect(pill.classes()).toContain('relative')
    expect(pill.classes()).toContain('after:top-1/2')
    expect(pill.classes()).toContain('after:-translate-y-1/2')
  })

  // The three sizes are one sum: a 44px pill holding an 18px knob inset by 3px
  // leaves exactly 20px of travel. Changing one without the others either strands
  // the knob short of the edge or pushes it out of the pill.
  it('keeps the knob inset on both ends of its travel', () => {
    const pill = mountSwitch().get('#demoToggle ~ span')
    expect(pill.classes()).toEqual(expect.arrayContaining(['w-11', 'after:w-4.5', 'after:left-0.75']))
    expect(pill.classes()).toContain('peer-checked:after:translate-x-5')
  })

  // A ring left behind by a mouse click reads as a permanent outline, so the
  // focus ring is keyboard only.
  it('rings only on keyboard focus', () => {
    const classes = mountSwitch().get('#demoToggle ~ span').classes()
    expect(classes).toContain('peer-focus-visible:ring-2')
    expect(classes.some((c) => c.startsWith('peer-focus:'))).toBe(false)
  })
})
