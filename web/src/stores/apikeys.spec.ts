import { describe, it, expect, vi, beforeEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useApiKeysStore } from './apikeys'

const mockApi = vi.hoisted(() => ({
  get: vi.fn(),
  post: vi.fn(),
  put: vi.fn(),
  delete: vi.fn(),
}))

vi.mock('../lib/api', () => ({ default: mockApi }))

const mockKey = {
  id: 1,
  key_id: 'gk_abc123',
  name: 'Test Key',
  domain: 'test.com',
  max_number: 50000,
  expire_seconds: 120,
  algorithm: 'SHA-256',
  enabled: true,
  created_at: '2026-01-01T00:00:00Z',
  updated_at: '2026-01-01T00:00:00Z',
}

describe('apikeys store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('initializes with empty keys', () => {
    const store = useApiKeysStore()
    expect(store.keys).toEqual([])
    expect(store.loading).toBe(false)
  })

  it('fetchKeys loads keys and manages loading state', async () => {
    mockApi.get.mockResolvedValue({ data: { keys: [mockKey] } })
    const store = useApiKeysStore()

    await store.fetchKeys()

    expect(mockApi.get).toHaveBeenCalledWith('/keys')
    expect(store.keys).toEqual([mockKey])
    expect(store.loading).toBe(false)
  })

  it('fetchKeys sets loading to false even on error', async () => {
    mockApi.get.mockRejectedValue(new Error('network error'))
    const store = useApiKeysStore()

    await expect(store.fetchKeys()).rejects.toThrow('network error')
    expect(store.loading).toBe(false)
  })

  it('createKey posts and refreshes keys', async () => {
    mockApi.post.mockResolvedValue({ data: mockKey })
    mockApi.get.mockResolvedValue({ data: { keys: [mockKey] } })
    const store = useApiKeysStore()

    const payload = { name: 'Test Key', domain: 'test.com', max_number: 50000, expire_seconds: 120 }
    const result = await store.createKey(payload)

    expect(mockApi.post).toHaveBeenCalledWith('/keys', payload)
    expect(result).toEqual(mockKey)
    expect(mockApi.get).toHaveBeenCalledWith('/keys')
  })

  it('getKey fetches single key', async () => {
    mockApi.get.mockResolvedValue({ data: mockKey })
    const store = useApiKeysStore()

    const result = await store.getKey(1)

    expect(mockApi.get).toHaveBeenCalledWith('/keys/1')
    expect(result).toEqual(mockKey)
  })

  it('updateKey puts and refreshes keys', async () => {
    const updated = { ...mockKey, name: 'Updated' }
    mockApi.put.mockResolvedValue({ data: updated })
    mockApi.get.mockResolvedValue({ data: { keys: [updated] } })
    const store = useApiKeysStore()

    const result = await store.updateKey(1, { name: 'Updated' })

    expect(mockApi.put).toHaveBeenCalledWith('/keys/1', { name: 'Updated' })
    expect(result.name).toBe('Updated')
  })

  it('deleteKey deletes and refreshes keys', async () => {
    mockApi.delete.mockResolvedValue({})
    mockApi.get.mockResolvedValue({ data: { keys: [] } })
    const store = useApiKeysStore()

    await store.deleteKey(1)

    expect(mockApi.delete).toHaveBeenCalledWith('/keys/1')
    expect(mockApi.get).toHaveBeenCalledWith('/keys')
    expect(store.keys).toEqual([])
  })

  it('rotateSecret calls API and returns new secret', async () => {
    mockApi.post.mockResolvedValue({ data: { hmac_secret: 'new-secret-123' } })
    const store = useApiKeysStore()

    const result = await store.rotateSecret(1)

    expect(mockApi.post).toHaveBeenCalledWith('/keys/1/rotate-secret')
    expect(result).toBe('new-secret-123')
  })
})
