import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import ApiKeyDetailView from './ApiKeyDetailView.vue'

const mockPush = vi.fn()

vi.mock('vue-router', () => ({
  useRouter: () => ({ push: mockPush }),
  useRoute: () => ({ params: { id: '1' } }),
}))

const mockApi = vi.hoisted(() => ({
  get: vi.fn(),
  post: vi.fn(),
  put: vi.fn(),
  delete: vi.fn(),
}))

vi.mock('../lib/api', () => ({ default: mockApi }))

const mockKey = {
  id: 1,
  key_id: 'gk_abc123def456',
  hmac_secret: 'secret-value-here',
  name: 'Test Key',
  domain: 'test.com',
  max_number: 100000,
  expire_seconds: 300,
  algorithm: 'SHA-256',
  rate_limit_per_min: 60,
  adaptive_difficulty: true,
  his_sampling: false,
  his_enforce: false,
  his_threshold: 0.8,
  enabled: true,
  created_at: '2026-01-01T00:00:00Z',
  updated_at: '2026-01-01T00:00:00Z',
}

function mountView() {
  return mount(ApiKeyDetailView, {
    global: {
      stubs: {
        'router-link': { template: '<a><slot /></a>' },
        StatsChart: { template: '<div />' },
      },
    },
  })
}

describe('ApiKeyDetailView', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    mockApi.get
      .mockResolvedValueOnce({ data: mockKey }) // getKey
      .mockResolvedValueOnce({ data: { days: [] } }) // fetchKeyStats
  })

  it('loads and displays key details', async () => {
    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.text()).toContain('Test Key')
    expect(wrapper.text()).toContain('gk_abc123def456')
    expect(wrapper.text()).toContain('test.com')
    expect(wrapper.text()).toContain('SHA-256')
    expect(wrapper.text()).toContain('60 req/min')
    expect(wrapper.text()).toContain('Adaptive difficulty')
    expect(wrapper.text()).toContain('On')
  })

  it('shows "Unlimited" when no per-key rate limit is set', async () => {
    mockApi.get
      .mockReset()
      .mockResolvedValueOnce({ data: { ...mockKey, rate_limit_per_min: 0 } })
      .mockResolvedValueOnce({ data: { days: [] } })

    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.text()).toContain('Unlimited')
  })

  it('computes challengeUrl correctly', async () => {
    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.text()).toContain('/api/v1/challenge?apiKey=gk_abc123def456')
  })

  it('copies the instance URL to clipboard', async () => {
    const writeText = vi.fn()
    Object.defineProperty(navigator, 'clipboard', { value: { writeText }, configurable: true })

    const wrapper = mountView()
    await flushPromises()

    const copyUrlBtn = wrapper.find('button[aria-label="Copy instance URL"]')
    expect(copyUrlBtn.exists()).toBe(true)
    await copyUrlBtn.trigger('click')

    expect(writeText).toHaveBeenCalledWith(globalThis.location.origin)
    expect(copyUrlBtn.text()).toBe('Copied!') // tooltip, not the button label
  })

  it('computes widgetSnippet correctly', async () => {
    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.text()).toContain('<altcha-widget')
    expect(wrapper.text()).toContain('challenge=')
  })

  // The closing tag cannot be written literally in the component, so assert the
  // rendered snippet is a complete script tag a user can paste as-is.
  it('renders a complete HIS collector snippet', async () => {
    const wrapper = mountView()
    await flushPromises()

    const text = wrapper.text()
    expect(text).toContain('/api/public/his.js')
    expect(text).toContain('</' + 'script>')
  })

  // Integrators kept asking how to switch Monitor into blocking (see #149).
  // Now that the switch exists, the page has to name it rather than say there
  // is nothing to switch, which is how this test read before enforcement shipped.
  it('names the switch when the key is in Monitor mode', async () => {
    const wrapper = mountView()
    await flushPromises()

    const text = wrapper.text()
    expect(text).toContain('never blocked')
    expect(text).toContain('Block suspected bots')
    expect(text).toContain('Monitor only')
  })

  it('says what blocking does when the key enforces', async () => {
    // beforeEach queues the key response with mockResolvedValueOnce, so the
    // queue has to be rebuilt rather than shadowed by a default.
    mockApi.get.mockReset()
    mockApi.get
      .mockResolvedValueOnce({ data: { ...mockKey, his_enforce: true, his_threshold: 0.5 } })
      .mockResolvedValue({ data: { days: [] } })
    const wrapper = mountView()
    await flushPromises()

    const text = wrapper.text()
    expect(text).toContain('Blocking')
    // The threshold has to be the key's own, not the 0.8 default: a key lowered
    // to 0.5 that still advertised 0.8 would describe a policy it does not apply.
    expect(text).toContain('0.50')
    expect(text).toContain('bot_suspected')
  })

  it('toggles key enabled state', async () => {
    mockApi.put.mockResolvedValue({ data: { ...mockKey, enabled: false } })
    mockApi.get.mockResolvedValue({ data: { ...mockKey, enabled: false } })

    const wrapper = mountView()
    await flushPromises()

    const disableBtn = wrapper.findAll('button').find(b => b.text() === 'Disable')
    expect(disableBtn).toBeDefined()
    await disableBtn!.trigger('click')
    await flushPromises()

    expect(mockApi.put).toHaveBeenCalledWith('/keys/1', { enabled: false })
  })

  it('handles delete flow', async () => {
    mockApi.delete.mockResolvedValue({})
    mockApi.get
      .mockReset()
      .mockResolvedValueOnce({ data: mockKey })
      .mockResolvedValueOnce({ data: { days: [] } })
      .mockResolvedValue({ data: { keys: [] } })

    const wrapper = mountView()
    await flushPromises()

    // Open confirmation
    const deleteBtn = wrapper.findAll('button').find(b => b.text() === 'Delete')
    await deleteBtn!.trigger('click')

    expect(wrapper.text()).toContain('Delete API Key?')

    // Confirm delete — find the confirm button inside the modal
    const modalButtons = wrapper.findAll('dialog button')
    const confirmDelete = modalButtons.find(b => b.text() === 'Delete')
    await confirmDelete!.trigger('click')
    await flushPromises()

    expect(mockApi.delete).toHaveBeenCalledWith('/keys/1')
    expect(mockPush).toHaveBeenCalledWith('/keys')
  })

  // The dialog stays up until the navigation lands, so an impatient second click
  // used to send a second DELETE.
  it('sends one delete however many times the button is clicked', async () => {
    let resolveDelete: (value: unknown) => void = () => {}
    mockApi.delete.mockReturnValue(new Promise((resolve) => { resolveDelete = resolve }))
    mockApi.get
      .mockReset()
      .mockResolvedValueOnce({ data: mockKey })
      .mockResolvedValueOnce({ data: { days: [] } })
      .mockResolvedValue({ data: { keys: [] } })

    const wrapper = mountView()
    await flushPromises()

    await wrapper.findAll('button').find(b => b.text() === 'Delete')!.trigger('click')
    const confirmDelete = wrapper.findAll('dialog button').find(b => b.text() === 'Delete')
    await confirmDelete!.trigger('click')
    await confirmDelete!.trigger('click')

    resolveDelete({})
    await flushPromises()

    expect(mockApi.delete).toHaveBeenCalledTimes(1)
  })

  it('rotates secret with confirmation', async () => {
    globalThis.confirm = vi.fn(() => true)
    mockApi.post.mockResolvedValue({ data: { hmac_secret: 'new-secret' } })

    const wrapper = mountView()
    await flushPromises()

    const rotateBtn = wrapper.findAll('button').find(b => b.text() === 'Rotate')
    await rotateBtn!.trigger('click')
    await flushPromises()

    expect(mockApi.post).toHaveBeenCalledWith('/keys/1/rotate-secret')
  })

  it('cancels rotate secret when not confirmed', async () => {
    globalThis.confirm = vi.fn(() => false)

    const wrapper = mountView()
    await flushPromises()

    const rotateBtn = wrapper.findAll('button').find(b => b.text() === 'Rotate')
    await rotateBtn!.trigger('click')
    await flushPromises()

    expect(mockApi.post).not.toHaveBeenCalled()
  })

  it('shows loading state before key loads', () => {
    mockApi.get.mockReset().mockReturnValue(new Promise(() => {})) // never resolves
    const wrapper = mountView()
    expect(wrapper.text()).toContain('Loading...')
  })
})
