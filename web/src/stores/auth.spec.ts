import { describe, it, expect, vi, beforeEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useAuthStore } from './auth'

const mockApi = vi.hoisted(() => ({
  get: vi.fn(),
  post: vi.fn(),
  put: vi.fn(),
  delete: vi.fn(),
}))

vi.mock('../lib/api', () => ({ default: mockApi }))

describe('auth store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    localStorage.clear()
    vi.clearAllMocks()
  })

  it('initializes with empty token when no localStorage', () => {
    const store = useAuthStore()
    expect(store.token).toBe('')
    expect(store.isAuthenticated).toBe(false)
  })

  it('initializes with token from localStorage', () => {
    localStorage.setItem('gatecha_token', 'saved-token')
    const store = useAuthStore()
    expect(store.token).toBe('saved-token')
    expect(store.isAuthenticated).toBe(true)
  })

  it('login stores token', async () => {
    mockApi.post.mockResolvedValue({ data: { token: 'new-token' } })
    const store = useAuthStore()

    await store.login('admin', 'password')

    expect(mockApi.post).toHaveBeenCalledWith('/login', {
      username: 'admin',
      password: 'password',
    })
    expect(store.token).toBe('new-token')
    expect(localStorage.getItem('gatecha_token')).toBe('new-token')
    expect(store.isAuthenticated).toBe(true)
  })

  it('login with altcha payload includes it in body', async () => {
    mockApi.post.mockResolvedValue({ data: { token: 'token' } })
    const store = useAuthStore()

    await store.login('admin', 'password', 'captcha-payload')

    expect(mockApi.post).toHaveBeenCalledWith('/login', {
      username: 'admin',
      password: 'password',
      altcha_payload: 'captcha-payload',
    })
  })

  it('logout clears token', () => {
    localStorage.setItem('gatecha_token', 'token')
    const store = useAuthStore()
    expect(store.isAuthenticated).toBe(true)

    store.logout()

    expect(store.token).toBe('')
    expect(store.isAuthenticated).toBe(false)
    expect(localStorage.getItem('gatecha_token')).toBeNull()
  })

  it('checkAuth returns true when authenticated', async () => {
    mockApi.get.mockResolvedValue({ data: { username: 'admin' } })
    const store = useAuthStore()

    const result = await store.checkAuth()

    expect(result).toBe(true)
    expect(mockApi.get).toHaveBeenCalledWith('/me')
  })

  it('checkAuth returns false and logs out on error', async () => {
    mockApi.get.mockRejectedValue(new Error('unauthorized'))
    localStorage.setItem('gatecha_token', 'old-token')
    const store = useAuthStore()

    const result = await store.checkAuth()

    expect(result).toBe(false)
    expect(store.token).toBe('')
    expect(localStorage.getItem('gatecha_token')).toBeNull()
  })
})
