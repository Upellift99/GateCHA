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

  it('links the version to the GitHub releases page', () => {
    const auth = useAuthStore()
    auth.token = 'tok'
    auth.version = '0.2.1'
    const wrapper = mount(Footer)
    const link = wrapper.findAll('a').find((a) => a.text() === '0.2.1')!
    expect(link.attributes('href')).toBe('https://github.com/Upellift99/GateCHA/releases')
    expect(link.attributes('target')).toBe('_blank')
    expect(link.attributes('rel')).toContain('noopener')
  })

  it('links the GateCHA wordmark to gatecha.org', () => {
    const wrapper = mount(Footer)
    const brand = wrapper.findAll('a').find((a) => a.text() === 'GateCHA')!
    expect(brand).toBeTruthy()
    expect(brand.attributes('href')).toBe('https://gatecha.org')
  })

  it('does not show a version when logged out', () => {
    const wrapper = mount(Footer)
    expect(wrapper.text()).not.toContain('v1')
    expect(mockApi.get).not.toHaveBeenCalled()
  })

  // Regression: Footer is mounted in App.vue outside <router-view>, so it mounts
  // once on the login page (unauthenticated, showVersion=false) and never remounts.
  // Fetching only in onMounted meant the version was never loaded after logging in.
  it('fetches the version after logging in, without remounting', async () => {
    mockApi.get.mockResolvedValue({ data: { version: '0.3.2' } })
    const auth = useAuthStore()
    const wrapper = mount(Footer, { props: { showVersion: false } })
    expect(mockApi.get).not.toHaveBeenCalled()

    // login: token appears and the route leaves the login page
    auth.token = 'tok'
    await wrapper.setProps({ showVersion: true })
    await vi.waitFor(() => expect(mockApi.get).toHaveBeenCalledWith('/version'))
    // waitFor the rendered output, not just the call: the version lands only after
    // the request resolves and the reactive update flushes.
    await vi.waitFor(() => expect(wrapper.text()).toContain('0.3.2'))
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
