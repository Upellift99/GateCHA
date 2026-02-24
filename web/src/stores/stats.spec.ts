import { describe, it, expect, vi, beforeEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useStatsStore } from './stats'

const mockApi = vi.hoisted(() => ({
  get: vi.fn(),
  post: vi.fn(),
  put: vi.fn(),
  delete: vi.fn(),
}))

vi.mock('../lib/api', () => ({ default: mockApi }))

const mockOverview = {
  total_challenges: 100,
  total_verifications_ok: 80,
  total_verifications_fail: 20,
  active_keys: 3,
  daily: [
    { date: '2026-02-24', challenges_issued: 10, verifications_ok: 8, verifications_fail: 2 },
  ],
}

describe('stats store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('initializes with null overview and empty stats', () => {
    const store = useStatsStore()
    expect(store.overview).toBeNull()
    expect(store.keyStats).toEqual([])
    expect(store.keysSummary).toEqual({})
  })

  it('fetchOverview loads overview with default days', async () => {
    mockApi.get.mockResolvedValue({ data: mockOverview })
    const store = useStatsStore()

    await store.fetchOverview()

    expect(mockApi.get).toHaveBeenCalledWith('/stats/overview?days=30')
    expect(store.overview).toEqual(mockOverview)
  })

  it('fetchOverview loads overview with custom days', async () => {
    mockApi.get.mockResolvedValue({ data: mockOverview })
    const store = useStatsStore()

    await store.fetchOverview(7)

    expect(mockApi.get).toHaveBeenCalledWith('/stats/overview?days=7')
  })

  it('fetchKeyStats loads key stats', async () => {
    const days = [
      { date: '2026-02-24', challenges_issued: 5, verifications_ok: 4, verifications_fail: 1 },
    ]
    mockApi.get.mockResolvedValue({ data: { days } })
    const store = useStatsStore()

    await store.fetchKeyStats(1)

    expect(mockApi.get).toHaveBeenCalledWith('/stats/keys/1?days=30')
    expect(store.keyStats).toEqual(days)
  })

  it('fetchKeyStats with custom days', async () => {
    mockApi.get.mockResolvedValue({ data: { days: [] } })
    const store = useStatsStore()

    await store.fetchKeyStats(1, 7)

    expect(mockApi.get).toHaveBeenCalledWith('/stats/keys/1?days=7')
  })

  it('fetchKeysSummary loads keys summary', async () => {
    const keys = {
      '1': { api_key_id: 1, challenges_issued: 10, verifications_ok: 8, verifications_fail: 2 },
    }
    mockApi.get.mockResolvedValue({ data: { keys } })
    const store = useStatsStore()

    await store.fetchKeysSummary()

    expect(mockApi.get).toHaveBeenCalledWith('/stats/keys-summary')
    expect(store.keysSummary).toEqual(keys)
  })
})
