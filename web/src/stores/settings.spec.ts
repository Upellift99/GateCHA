import { describe, it, expect, vi, beforeEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useSettingsStore } from './settings'

const mockApi = vi.hoisted(() => ({
  get: vi.fn(),
  post: vi.fn(),
  put: vi.fn(),
  delete: vi.fn(),
}))

vi.mock('../lib/api', () => ({ default: mockApi }))

describe('settings store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('initializes with default settings', () => {
    const store = useSettingsStore()
    expect(store.settings).toEqual({ login_captcha_enabled: false })
    expect(store.loading).toBe(false)
  })

  it('fetchSettings loads settings and manages loading state', async () => {
    mockApi.get.mockResolvedValue({ data: { login_captcha_enabled: true } })
    const store = useSettingsStore()

    await store.fetchSettings()

    expect(mockApi.get).toHaveBeenCalledWith('/settings')
    expect(store.settings.login_captcha_enabled).toBe(true)
    expect(store.loading).toBe(false)
  })

  it('fetchSettings sets loading to false even on error', async () => {
    mockApi.get.mockRejectedValue(new Error('network error'))
    const store = useSettingsStore()

    await expect(store.fetchSettings()).rejects.toThrow('network error')
    expect(store.loading).toBe(false)
  })

  it('updateSettings sends PUT and updates state', async () => {
    mockApi.put.mockResolvedValue({ data: { login_captcha_enabled: true } })
    const store = useSettingsStore()

    await store.updateSettings({ login_captcha_enabled: true })

    expect(mockApi.put).toHaveBeenCalledWith('/settings', { login_captcha_enabled: true })
    expect(store.settings.login_captcha_enabled).toBe(true)
  })

  it('updateSettings with false disables captcha', async () => {
    mockApi.put.mockResolvedValue({ data: { login_captcha_enabled: false } })
    const store = useSettingsStore()

    await store.updateSettings({ login_captcha_enabled: false })

    expect(store.settings.login_captcha_enabled).toBe(false)
  })
})
