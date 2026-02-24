import { describe, it, expect, vi, beforeEach } from 'vitest'
import axios from 'axios'

vi.mock('axios', () => {
  const interceptors = {
    request: { use: vi.fn(), eject: vi.fn() },
    response: { use: vi.fn(), eject: vi.fn() },
  }
  const instance = {
    interceptors,
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn(),
  }
  return {
    default: {
      create: vi.fn(() => instance),
    },
  }
})

describe('api', () => {
  beforeEach(() => {
    vi.resetModules()
    localStorage.clear()
  })

  it('creates axios instance with /api/admin baseURL', async () => {
    await import('./api')
    expect(axios.create).toHaveBeenCalledWith({ baseURL: '/api/admin' })
  })

  it('request interceptor adds Authorization header when token exists', async () => {
    localStorage.setItem('gatecha_token', 'test-token')
    await import('./api')

    const requestUse = (axios.create as ReturnType<typeof vi.fn>).mock.results[0].value.interceptors.request.use
    const onFulfilled = requestUse.mock.calls[0][0]

    const config = { headers: {} as Record<string, string> }
    const result = onFulfilled(config)

    expect(result.headers.Authorization).toBe('Bearer test-token')
  })

  it('request interceptor does not add header when no token', async () => {
    await import('./api')

    const requestUse = (axios.create as ReturnType<typeof vi.fn>).mock.results[0].value.interceptors.request.use
    const onFulfilled = requestUse.mock.calls[0][0]

    const config = { headers: {} as Record<string, string> }
    const result = onFulfilled(config)

    expect(result.headers.Authorization).toBeUndefined()
  })

  it('response interceptor removes token and redirects on 401', async () => {
    localStorage.setItem('gatecha_token', 'test-token')

    const originalHref = globalThis.location.href
    Object.defineProperty(globalThis, 'location', {
      value: { href: originalHref },
      writable: true,
      configurable: true,
    })

    await import('./api')

    const responseUse = (axios.create as ReturnType<typeof vi.fn>).mock.results[0].value.interceptors.response.use
    const onRejected = responseUse.mock.calls[0][1]

    const error = { response: { status: 401 } }
    await expect(onRejected(error)).rejects.toEqual(error)

    expect(localStorage.getItem('gatecha_token')).toBeNull()
    expect(globalThis.location.href).toBe('/login')
  })

  it('response interceptor passes through non-401 errors', async () => {
    await import('./api')

    const responseUse = (axios.create as ReturnType<typeof vi.fn>).mock.results[0].value.interceptors.response.use
    const onRejected = responseUse.mock.calls[0][1]

    const error = { response: { status: 500 } }
    await expect(onRejected(error)).rejects.toEqual(error)
  })

  it('response interceptor passes through successful responses', async () => {
    await import('./api')

    const responseUse = (axios.create as ReturnType<typeof vi.fn>).mock.results[0].value.interceptors.response.use
    const onFulfilled = responseUse.mock.calls[0][0]

    const response = { data: 'ok', status: 200 }
    expect(onFulfilled(response)).toEqual(response)
  })
})
