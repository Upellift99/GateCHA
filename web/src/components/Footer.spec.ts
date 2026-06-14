import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import Footer from './Footer.vue'
import { useAuthStore } from '../stores/auth'

const mockApi = vi.hoisted(() => ({ get: vi.fn() }))
vi.mock('../lib/api', () => ({ default: mockApi }))

describe('Footer', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    mockApi.get.mockReset()
    localStorage.clear()
  })

  it('always links to gatecha.org + docs and shows the Sceaux tagline', () => {
    const wrapper = mount(Footer)
    const hrefs = wrapper.findAll('a').map((a) => a.attributes('href'))
    expect(hrefs).toContain('https://gatecha.org')
    expect(hrefs).toContain('https://gatecha.org/docs')
    const site = wrapper.findAll('a').find((a) => a.attributes('href') === 'https://gatecha.org')!
    expect(site.attributes('target')).toBe('_blank')
    expect(site.attributes('rel')).toContain('noopener')
    expect(wrapper.text()).toContain('Made with')
    expect(wrapper.text()).toContain('Sceaux, France')
  })

  it('shows the version when authenticated', () => {
    const auth = useAuthStore()
    auth.token = 'tok'
    auth.version = 'v1.2.3'
    const wrapper = mount(Footer)
    expect(wrapper.text()).toContain('v1.2.3')
    // version is already set, so no fetch is triggered
    expect(mockApi.get).not.toHaveBeenCalled()
  })

  it('does not show a version when logged out', () => {
    const wrapper = mount(Footer)
    expect(wrapper.text()).not.toContain('v1')
    expect(mockApi.get).not.toHaveBeenCalled()
  })

  it('never shows or fetches the version when showVersion is false (login page)', () => {
    const auth = useAuthStore()
    auth.token = 'tok'
    auth.version = 'v1.2.3'
    const wrapper = mount(Footer, { props: { showVersion: false } })
    expect(wrapper.text()).not.toContain('v1.2.3')
    expect(mockApi.get).not.toHaveBeenCalled()
  })
})
