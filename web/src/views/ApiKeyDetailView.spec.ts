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
  })

  it('computes challengeUrl correctly', async () => {
    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.text()).toContain('/api/v1/challenge?apiKey=gk_abc123def456')
  })

  it('computes widgetSnippet correctly', async () => {
    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.text()).toContain('<altcha-widget')
    expect(wrapper.text()).toContain('challengeurl=')
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

    // Confirm delete
    const confirmBtn = wrapper.findAll('button').find(b => b.text() === 'Delete' && b.classes().length > 0)
    // Find the confirm button inside the modal
    const modalButtons = wrapper.findAll('.fixed button')
    const confirmDelete = modalButtons.find(b => b.text() === 'Delete')
    await confirmDelete!.trigger('click')
    await flushPromises()

    expect(mockApi.delete).toHaveBeenCalledWith('/keys/1')
    expect(mockPush).toHaveBeenCalledWith('/keys')
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
