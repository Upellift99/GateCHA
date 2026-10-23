import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import NetworkStatus from './NetworkStatus.vue'
import LinesChart from './LinesChart.vue'
import { todayUTC } from '../lib/lines'

const key = (id: number, name: string, enabled = true) => ({
  id,
  key_id: `gk_${id}`,
  name,
  domain: `${name.toLowerCase()}.example`,
  max_number: 100000,
  expire_seconds: 300,
  algorithm: 'SHA-256',
  enabled,
  created_at: '',
  updated_at: '',
})

const today = todayUTC()
const row = (api_key_id: number, challenges_issued: number, verifications_ok: number, verifications_fail: number) => ({
  api_key_id,
  date: today,
  challenges_issued,
  verifications_ok,
  verifications_fail,
  his_observations: 0,
  his_bot_suspected: 0,
})

const stubs = { 'router-link': { template: '<a><slot /></a>' } }

describe('NetworkStatus', () => {
  const keys = [key(1, 'Alpha'), key(2, 'Bravo'), key(3, 'Charlie', false), key(4, 'Delta')]
  const daily = [row(1, 50, 49, 1), row(2, 74, 40, 31)]

  function mountBoard(selected: number | null = null) {
    return mount(NetworkStatus, { props: { keys, daily, selected }, global: { stubs } })
  }

  it('puts disturbed lines first and closed lines last', () => {
    const names = mountBoard().findAll('tbody tr').map((r) => r.find('a').text())
    expect(names).toEqual(['Bravo', 'Alpha', 'Delta', 'Charlie'])
  })

  it("shows today's counts and the transit status of each line", () => {
    const rows = mountBoard().findAll('tbody tr')
    expect(rows[0]!.text()).toContain('Disturbed')
    expect(rows[0]!.text()).toContain('31 failed today')
    expect(rows[1]!.text()).toContain('Normal service')
    expect(rows[2]!.text()).toContain('No traffic today')
    expect(rows[3]!.text()).toContain('Closed')
  })

  it('selects a line on hover and dims the others', async () => {
    const wrapper = mountBoard()
    await wrapper.findAll('tbody tr')[1]!.trigger('mouseenter')
    expect(wrapper.emitted('select')![0]).toEqual([1])

    const dimmed = mountBoard(1).findAll('tbody tr').map((r) => r.classes().includes('opacity-45'))
    expect(dimmed).toEqual([true, false, true, true])
  })

  it('invites to create a key when there are no lines', () => {
    const wrapper = mount(NetworkStatus, { props: { keys: [], daily: [], selected: null }, global: { stubs } })
    expect(wrapper.text()).toContain('Create your first key')
  })
})

describe('LinesChart', () => {
  it('pins and unpins a line from its legend', async () => {
    const keys = [key(1, 'Alpha'), key(2, 'Bravo')]
    const wrapper = mount(LinesChart, { props: { keys, daily: [row(1, 5, 5, 0)], selected: null, pinned: null } })
    const legend = wrapper.findAll('button')
    expect(legend).toHaveLength(2)
    await legend[1]!.trigger('click')
    expect(wrapper.emitted('pin')![0]).toEqual([2])

    await wrapper.setProps({ pinned: 2 })
    expect(wrapper.findAll('button')[1]!.attributes('aria-pressed')).toBe('true')
    await wrapper.findAll('button')[1]!.trigger('click')
    expect(wrapper.emitted('pin')![1]).toEqual([null])
  })
})
