import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import Navbar from './Navbar.vue'

vi.mock('vue-router', () => ({ useRouter: () => ({ push: vi.fn() }) }))

const mockApi = vi.hoisted(() => ({ get: vi.fn() }))
vi.mock('../lib/api', () => ({ default: mockApi }))

const stubs = { 'router-link': { template: '<a><slot /></a>' } }

describe('Navbar', () => {
  beforeEach(() => {
    localStorage.clear()
    mockApi.get.mockReset()
    mockApi.get.mockResolvedValue({ data: { version: 'v1.2.3' } })
  })

  // A signed-out /version call answers 401, which reloads /login: an endless loop.
  it('does not ask for the version when signed out', async () => {
    setActivePinia(createPinia())
    mount(Navbar, { global: { stubs } })
    await flushPromises()
    expect(mockApi.get).not.toHaveBeenCalled()
  })

  it('shows the version once signed in', async () => {
    localStorage.setItem('gatecha_token', 'tok')
    setActivePinia(createPinia())
    const wrapper = mount(Navbar, { global: { stubs } })
    await flushPromises()
    expect(mockApi.get).toHaveBeenCalledWith('/version')
    expect(wrapper.text()).toContain('v1.2.3')
  })
})
