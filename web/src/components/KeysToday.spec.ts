import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import KeysToday from './KeysToday.vue'
import AttentionBanner from './AttentionBanner.vue'
import { todayUTC } from '../lib/status'

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

describe('KeysToday', () => {
  const keys = [key(1, 'Alpha'), key(2, 'Bravo'), key(3, 'Charlie', false), key(4, 'Delta')]
  const daily = [row(1, 50, 49, 1), row(2, 74, 40, 31)]

  it('groups keys by what needs doing, failing first', () => {
    const wrapper = mount(KeysToday, { props: { keys, daily }, global: { stubs } })
    const groups = wrapper.findAll('th[scope="rowgroup"]').map((th) => th.text())
    expect(groups).toEqual(['Needs attention · 1', 'Working · 1', 'No traffic today · 1', 'Disabled · 1'])
    const names = wrapper.findAll('tbody tr td:first-child a').map((a) => a.text())
    expect(names).toEqual(['Bravo', 'Alpha', 'Delta', 'Charlie'])
  })

  it("shows today's challenges and failures", () => {
    const wrapper = mount(KeysToday, { props: { keys, daily }, global: { stubs } })
    const failing = wrapper.findAll('tbody tr').find((r) => r.text().includes('Bravo'))!
    expect(failing.text()).toContain('74')
    expect(failing.text()).toContain('31')
  })

  it('invites to create a key when there are none', () => {
    const wrapper = mount(KeysToday, { props: { keys: [], daily: [] }, global: { stubs } })
    expect(wrapper.text()).toContain('Create your first key')
  })
})

describe('AttentionBanner', () => {
  it('names the failing key with its counts', () => {
    const wrapper = mount(AttentionBanner, {
      props: { failing: [{ key: key(2, 'Bravo'), failed: 31, verified: 71 }], activeKeys: 4, checkedToday: 120 },
      global: { stubs },
    })
    expect(wrapper.text()).toContain('1 key needs attention')
    expect(wrapper.text()).toContain('Bravo')
    expect(wrapper.text()).toContain('31')
    expect(wrapper.text()).toContain('(44%)')
  })

  it('lists several failing keys', () => {
    const wrapper = mount(AttentionBanner, {
      props: {
        failing: [
          { key: key(2, 'Bravo'), failed: 31, verified: 71 },
          { key: key(4, 'Delta'), failed: 9, verified: 20 },
        ],
        activeKeys: 4,
        checkedToday: 120,
      },
      global: { stubs },
    })
    expect(wrapper.text()).toContain('2 keys need attention')
    expect(wrapper.findAll('li')).toHaveLength(2)
  })

  it('says all is well when nothing fails', () => {
    const wrapper = mount(AttentionBanner, { props: { failing: [], activeKeys: 4, checkedToday: 120 }, global: { stubs } })
    expect(wrapper.text()).toContain('All keys are working today.')
  })

  it('does not claim keys work before any verification today', () => {
    const wrapper = mount(AttentionBanner, { props: { failing: [], activeKeys: 4, checkedToday: 0 }, global: { stubs } })
    expect(wrapper.text()).toContain('No verification yet today.')
    expect(wrapper.text()).not.toContain('working')
  })
})
