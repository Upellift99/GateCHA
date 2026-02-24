import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import ApiKeyFormView from './ApiKeyFormView.vue'

const mockPush = vi.fn()
const mockRoute = vi.hoisted(() => ({
  name: 'keys-new' as string,
  params: {} as Record<string, string>,
}))

vi.mock('vue-router', () => ({
  useRouter: () => ({ push: mockPush }),
  useRoute: () => mockRoute,
}))

const mockApi = vi.hoisted(() => ({
  get: vi.fn(),
  post: vi.fn(),
  put: vi.fn(),
  delete: vi.fn(),
}))

vi.mock('../lib/api', () => ({ default: mockApi }))

function mountView() {
  return mount(ApiKeyFormView, {
    global: {
      stubs: { 'router-link': { template: '<a><slot /></a>' } },
    },
  })
}

describe('ApiKeyFormView', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    mockRoute.name = 'keys-new'
    mockRoute.params = {}
  })

  it('renders create form by default', () => {
    const wrapper = mountView()
    expect(wrapper.text()).toContain('Create API Key')
  })

  it('renders edit form when route is keys-edit', async () => {
    mockRoute.name = 'keys-edit'
    mockRoute.params = { id: '1' }
    mockApi.get.mockResolvedValue({
      data: { id: 1, name: 'Existing', domain: 'test.com', max_number: 50000, expire_seconds: 120, algorithm: 'SHA-256' },
    })

    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.text()).toContain('Edit API Key')
    expect(mockApi.get).toHaveBeenCalledWith('/keys/1')
  })

  it('creates key on submit', async () => {
    mockApi.post.mockResolvedValue({
      data: { key_id: 'gk_new123', hmac_secret: 'secret123' },
    })
    mockApi.get.mockResolvedValue({ data: { keys: [] } })

    const wrapper = mountView()
    await wrapper.find('#key-name').setValue('New Key')
    await wrapper.find('form').trigger('submit')
    await flushPromises()

    expect(mockApi.post).toHaveBeenCalledWith('/keys', expect.objectContaining({
      name: 'New Key',
      algorithm: 'SHA-256',
    }))
    // Shows created key info
    expect(wrapper.text()).toContain('Key Created Successfully')
    expect(wrapper.text()).toContain('gk_new123')
  })

  it('updates key on submit in edit mode', async () => {
    mockRoute.name = 'keys-edit'
    mockRoute.params = { id: '5' }
    mockApi.get.mockResolvedValue({
      data: { id: 5, name: 'Old', domain: '', max_number: 100000, expire_seconds: 300, algorithm: 'SHA-256' },
    })
    mockApi.put.mockResolvedValue({ data: { id: 5, name: 'Updated' } })

    const wrapper = mountView()
    await flushPromises()

    await wrapper.find('#key-name').setValue('Updated')
    await wrapper.find('form').trigger('submit')
    await flushPromises()

    expect(mockApi.put).toHaveBeenCalledWith('/keys/5', expect.objectContaining({ name: 'Updated' }))
    expect(mockPush).toHaveBeenCalledWith('/keys/5')
  })

  it('shows error on submit failure', async () => {
    mockApi.post.mockRejectedValue(new Error('fail'))
    const wrapper = mountView()

    await wrapper.find('form').trigger('submit')
    await flushPromises()

    expect(wrapper.text()).toContain('Failed to save key')
  })

  it('has default form values', () => {
    const wrapper = mountView()
    const maxNumber = wrapper.find('#key-max-number')
    expect((maxNumber.element as HTMLInputElement).value).toBe('100000')
  })
})
