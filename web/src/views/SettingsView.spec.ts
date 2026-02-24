import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { useSettingsStore } from '../stores/settings'
import SettingsView from './SettingsView.vue'

const mockApi = vi.hoisted(() => ({
  get: vi.fn(),
  post: vi.fn(),
  put: vi.fn(),
  delete: vi.fn(),
}))

vi.mock('../lib/api', () => ({ default: mockApi }))

function mountView() {
  return mount(SettingsView, {
    global: {
      stubs: { 'router-link': { template: '<a><slot /></a>' } },
    },
  })
}

describe('SettingsView', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    mockApi.get.mockResolvedValue({ data: { login_captcha_enabled: false } })
  })

  it('renders the settings form', () => {
    const wrapper = mountView()
    expect(wrapper.text()).toContain('Change Password')
    expect(wrapper.text()).toContain('Security')
  })

  it('validates password mismatch', async () => {
    const wrapper = mountView()

    await wrapper.find('#currentPassword').setValue('oldpass')
    await wrapper.find('#newPassword').setValue('newpass12')
    await wrapper.find('#confirmPassword').setValue('different')
    await wrapper.find('form').trigger('submit')

    expect(wrapper.text()).toContain('Passwords do not match')
  })

  it('validates password length', async () => {
    const wrapper = mountView()

    await wrapper.find('#currentPassword').setValue('oldpass')
    await wrapper.find('#newPassword').setValue('short')
    await wrapper.find('#confirmPassword').setValue('short')
    await wrapper.find('form').trigger('submit')

    expect(wrapper.text()).toContain('Password must be at least 8 characters')
  })

  it('submits password change successfully', async () => {
    mockApi.post.mockResolvedValue({ data: {} })
    const wrapper = mountView()

    await wrapper.find('#currentPassword').setValue('oldpassword')
    await wrapper.find('#newPassword').setValue('newpassword123')
    await wrapper.find('#confirmPassword').setValue('newpassword123')
    await wrapper.find('form').trigger('submit')
    await flushPromises()

    expect(mockApi.post).toHaveBeenCalledWith('/change-password', {
      current_password: 'oldpassword',
      new_password: 'newpassword123',
    })
    expect(wrapper.text()).toContain('Password changed successfully')
  })

  it('shows error on password change failure', async () => {
    mockApi.post.mockRejectedValue(new Error('bad'))
    const wrapper = mountView()

    await wrapper.find('#currentPassword').setValue('oldpassword')
    await wrapper.find('#newPassword').setValue('newpassword123')
    await wrapper.find('#confirmPassword').setValue('newpassword123')
    await wrapper.find('form').trigger('submit')
    await flushPromises()

    expect(wrapper.text()).toContain('Failed to change password')
  })

  it('toggles captcha setting', async () => {
    mockApi.put.mockResolvedValue({ data: { login_captcha_enabled: true } })
    const wrapper = mountView()

    const checkbox = wrapper.find('#loginCaptchaToggle')
    await checkbox.setValue(true)
    await flushPromises()

    const store = useSettingsStore()
    expect(store.settings.login_captcha_enabled).toBe(true)
  })

  it('reverts captcha toggle on error', async () => {
    mockApi.put.mockRejectedValue(new Error('fail'))
    mockApi.get.mockResolvedValue({ data: { login_captcha_enabled: false } })
    const wrapper = mountView()

    const checkbox = wrapper.find('#loginCaptchaToggle')
    await checkbox.setValue(true)
    await flushPromises()

    expect(wrapper.text()).toContain('Failed to update setting')
  })
})
