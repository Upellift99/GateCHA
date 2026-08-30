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

  // The checkbox is sr-only, so the switch is the only visible target and it has
  // to stay a sibling of the input: the focus ring is a peer variant. See #146.
  it('draws the switch as a sibling of the checkbox', () => {
    const svg = mountSwitch().get('#demoToggle ~ svg')
    expect(svg.attributes('viewBox')).toBe('0 0 44 24')
    expect(svg.classes()).toEqual(expect.arrayContaining(['h-6', 'w-11']))
  })

  // The switch is one SVG rather than a box with an ::after knob because a box
  // and its pseudo-element snap to whole device pixels independently, which
  // splits the track unevenly above and below the knob at fractional page
  // scales. Both shapes have to stay inside the same coordinate system for the
  // gap to be antialiased symmetrically.
  it('draws the track and the knob in one coordinate system', () => {
    const svg = mountSwitch().get('#demoToggle ~ svg')
    expect(svg.find('rect').exists()).toBe(true)
    expect(svg.find('circle').exists()).toBe(true)
  })

  // The numbers are one sum: a knob of r=8 centred at 12 leaves 4 units of track
  // on every side of a 44x24 box, and 20 units of travel to the other end.
  // Changing one without the others either strands the knob short of the edge or
  // pushes it through the cap.
  it('leaves the same track on every side of the knob', () => {
    const circle = mountSwitch().get('#demoToggle ~ svg circle')
    expect(circle.attributes('cx')).toBe('12')
    expect(circle.attributes('cy')).toBe('12')
    expect(circle.attributes('r')).toBe('8')
    expect(mountSwitch({ modelValue: true }).get('#demoToggle ~ svg circle').classes()).toContain(
      'translate-x-5',
    )
  })

  // A stroke is centred on the path, so it pushes the knob's antialiased edge
  // half a unit past r=8 and brings back the uneven rounding the SVG removes.
  it('draws the knob without a stroke', () => {
    const circle = mountSwitch().get('#demoToggle ~ svg circle')
    expect(circle.attributes('stroke')).toBeUndefined()
    expect(circle.classes().some((c) => c.startsWith('stroke-'))).toBe(false)
  })

  it('colours the track from the bound value', () => {
    expect(mountSwitch().get('#demoToggle ~ svg rect').classes()).toContain('fill-slate-200')
    expect(mountSwitch({ modelValue: true }).get('#demoToggle ~ svg rect').classes()).toContain(
      'fill-teal-600',
    )
  })

  // A ring left behind by a mouse click reads as a permanent outline, so the
  // focus ring is keyboard only.
  it('rings only on keyboard focus', () => {
    const classes = mountSwitch().get('#demoToggle ~ svg').classes()
    expect(classes).toContain('peer-focus-visible:ring-2')
    expect(classes.some((c) => c.startsWith('peer-focus:'))).toBe(false)
  })
})
