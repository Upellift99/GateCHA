import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import CountryTraffic from './CountryTraffic.vue'
import type { CountryStat } from '../stores/stats'

const countries: CountryStat[] = [
  { country: 'FR', verifications_ok: 80, verifications_fail: 20, total: 100 },
  { country: 'US', verifications_ok: 40, verifications_fail: 10, total: 50 },
  { country: '', verifications_ok: 1, verifications_fail: 9, total: 10 },
]

describe('CountryTraffic', () => {
  it('renders a flag and resolved name per country', () => {
    const wrapper = mount(CountryTraffic, { props: { countries } })
    const text = wrapper.text()
    expect(text).toContain('🇫🇷')
    expect(text).toContain('France')
    expect(text).toContain('🇺🇸')
    expect(text).toContain('United States')
  })

  it('shows a globe and "Unknown" for an empty (unlocatable) country code', () => {
    const wrapper = mount(CountryTraffic, { props: { countries: [countries[2]] } })
    expect(wrapper.text()).toContain('🌐')
    expect(wrapper.text()).toContain('Unknown')
  })

  it('computes failure rate and flags high-failure sources', () => {
    const wrapper = mount(CountryTraffic, { props: { countries } })
    // FR: 20/100 = 20%, the unknown bucket: 9/10 = 90%.
    expect(wrapper.text()).toContain('20%')
    expect(wrapper.text()).toContain('90%')
  })

  it('honors the limit prop', () => {
    const wrapper = mount(CountryTraffic, { props: { countries, limit: 2 } })
    expect(wrapper.findAll('li')).toHaveLength(2)
  })

  it('renders an empty state when there is no data', () => {
    const wrapper = mount(CountryTraffic, { props: { countries: [] } })
    expect(wrapper.text()).toContain('No verifications recorded yet')
    expect(wrapper.findAll('li')).toHaveLength(0)
  })
})
