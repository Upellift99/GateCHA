import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import LoginView from './LoginView.vue'

const mockPush = vi.fn()

vi.mock('vue-router', () => ({
  useRouter: () => ({ push: mockPush }),
}))

const mockApi = vi.hoisted(() => ({
  get: vi.fn(),
  post: vi.fn(),
  put: vi.fn(),
  delete: vi.fn(),
}))

vi.mock('../lib/api', () => ({ default: mockApi }))

const mockAxios = vi.hoisted(() => ({
  default: { get: vi.fn() },
}))

vi.mock('axios', () => mockAxios)

vi.mock('altcha', () => ({}))

function mountView() {
  return mount(LoginView, {
    global: {
      stubs: {
        'altcha-widget': { template: '<div />' },
      },
    },
  })
}

describe('LoginView', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    mockAxios.default.get.mockResolvedValue({ data: { captcha_required: false } })
    mockApi.post.mockResolvedValue({ data: { token: 'test-token' } })
  })

  it('renders login form', () => {
    const wrapper = mountView()
    expect(wrapper.text()).toContain('GateCHA')
    expect(wrapper.text()).toContain('Sign in')
  })

  it('fetches login config on mount', async () => {
    mountView()
    await flushPromises()
    expect(mockAxios.default.get).toHaveBeenCalledWith('/api/public/login-config')
  })

  it('handles login config fetch error gracefully', async () => {
    mockAxios.default.get.mockRejectedValue(new Error('fail'))
    const wrapper = mountView()
    await flushPromises()
    // Should not crash, captchaRequired stays false
    expect(wrapper.text()).toContain('Sign in')
  })

  it('logs in successfully and redirects', async () => {
    const wrapper = mountView()
    await flushPromises()

    await wrapper.find('#username').setValue('admin')
    await wrapper.find('#password').setValue('password123')
    await wrapper.find('form').trigger('submit')
    await flushPromises()

    expect(mockApi.post).toHaveBeenCalledWith('/login', {
      username: 'admin',
      password: 'password123',
    })
    expect(mockPush).toHaveBeenCalledWith('/')
  })

  it('shows error on login failure', async () => {
    mockApi.post.mockRejectedValue(new Error('unauthorized'))
    const wrapper = mountView()
    await flushPromises()

    await wrapper.find('#username').setValue('admin')
    await wrapper.find('#password').setValue('wrong')
    await wrapper.find('form').trigger('submit')
    await flushPromises()

    expect(wrapper.text()).toContain('Invalid credentials')
  })

  it('shows captcha widget when required', async () => {
    mockAxios.default.get.mockResolvedValue({
      data: { captcha_required: true, challenge_url: '/api/v1/challenge?apiKey=gk_test' },
    })
    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.find('altcha-widget-stub').exists() || wrapper.find('[challengeurl]').exists()).toBe(true)
  })

  it('canSubmit is false when captcha required but not verified', async () => {
    mockAxios.default.get.mockResolvedValue({
      data: { captcha_required: true, challenge_url: '/api/v1/challenge?apiKey=gk_test' },
    })
    const wrapper = mountView()
    await flushPromises()

    const button = wrapper.find('button[type="submit"]')
    expect(button.attributes('disabled')).toBeDefined()
  })
})
