import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { nextTick } from 'vue'
import { createPinia, setActivePinia } from 'pinia'
import { useApiKeysStore } from '../stores/apikeys'
import { useStatsStore } from '../stores/stats'
import ApiKeysView from './ApiKeysView.vue'

const mockApi = vi.hoisted(() => ({
  get: vi.fn(),
  post: vi.fn(),
  put: vi.fn(),
  delete: vi.fn(),
}))

vi.mock('../lib/api', () => ({ default: mockApi }))

const mockKeys = [
  { id: 1, key_id: 'gk_aaa', name: 'Bravo', domain: 'b.com', max_number: 100, expire_seconds: 60, algorithm: 'SHA-256', enabled: true, created_at: '', updated_at: '' },
  { id: 2, key_id: 'gk_bbb', name: 'Alpha', domain: 'a.com', max_number: 200, expire_seconds: 120, algorithm: 'SHA-256', enabled: false, created_at: '', updated_at: '' },
  { id: 3, key_id: 'gk_ccc', name: 'Charlie', domain: '', max_number: 300, expire_seconds: 300, algorithm: 'SHA-512', enabled: true, created_at: '', updated_at: '' },
]

const mockSummary = {
  '1': { api_key_id: 1, challenges_issued: 50, verifications_ok: 30, verifications_fail: 5 },
  '2': { api_key_id: 2, challenges_issued: 100, verifications_ok: 10, verifications_fail: 20 },
  '3': { api_key_id: 3, challenges_issued: 10, verifications_ok: 80, verifications_fail: 1 },
}

function mountView() {
  return mount(ApiKeysView, {
    global: {
      stubs: { 'router-link': { template: '<a><slot /></a>' } },
    },
  })
}

describe('ApiKeysView', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    // Default: return mockKeys so onMounted fetchKeys doesn't wipe store.keys
    mockApi.get.mockImplementation((url: string) => {
      if (url === '/keys') return Promise.resolve({ data: { keys: [...mockKeys] } })
      if (url === '/stats/keys-summary') return Promise.resolve({ data: { keys: {} } })
      return Promise.resolve({ data: {} })
    })
  })

  it('renders loading state', () => {
    const store = useApiKeysStore()
    store.loading = true
    const wrapper = mountView()
    expect(wrapper.text()).toContain('Loading...')
  })

  it('renders empty state when no keys', async () => {
    mockApi.get.mockImplementation((url: string) => {
      if (url === '/keys') return Promise.resolve({ data: { keys: [] } })
      if (url === '/stats/keys-summary') return Promise.resolve({ data: { keys: {} } })
      return Promise.resolve({ data: {} })
    })
    const wrapper = mountView()
    await flushPromises()
    expect(wrapper.text()).toContain('No API keys yet')
  })

  it('renders table when keys exist', async () => {
    const wrapper = mountView()
    await flushPromises()
    expect(wrapper.find('table').exists()).toBe(true)
    expect(wrapper.findAll('tbody tr')).toHaveLength(3)
  })

  it('getKeyStat returns stat value when summary exists', async () => {
    const wrapper = mountView()
    await flushPromises()
    const statsStore = useStatsStore()
    statsStore.keysSummary = { ...mockSummary }
    await nextTick()
    expect(wrapper.text()).toContain('50')
    expect(wrapper.text()).toContain('30')
  })

  it('getKeyStat returns 0 when no summary', async () => {
    const wrapper = mountView()
    await flushPromises()
    const cells = wrapper.findAll('td')
    const statCells = cells.filter(c => c.text() === '0')
    expect(statCells.length).toBeGreaterThan(0)
  })

  it('sorts by name ascending by default', async () => {
    const wrapper = mountView()
    await flushPromises()
    const rows = wrapper.findAll('tbody tr')
    expect(rows[0].text()).toContain('Alpha')
    expect(rows[1].text()).toContain('Bravo')
    expect(rows[2].text()).toContain('Charlie')
  })

  it('toggleSort reverses direction when same column clicked', async () => {
    const wrapper = mountView()
    await flushPromises()

    const nameHeader = wrapper.findAll('th')[0]
    await nameHeader.trigger('click')
    await nextTick()

    const rows = wrapper.findAll('tbody tr')
    expect(rows).toHaveLength(3)
    expect(rows[0].text()).toContain('Charlie')
    expect(rows[2].text()).toContain('Alpha')
  })

  it('toggleSort changes column and resets to asc', async () => {
    const wrapper = mountView()
    await flushPromises()

    const domainHeader = wrapper.findAll('th')[1]
    await domainHeader.trigger('click')
    await nextTick()

    const rows = wrapper.findAll('tbody tr')
    expect(rows).toHaveLength(3)
    expect(rows[0].text()).toContain('Charlie') // domain: ''
    expect(rows[1].text()).toContain('Alpha')   // domain: 'a.com'
    expect(rows[2].text()).toContain('Bravo')   // domain: 'b.com'
  })

  it('sorts by enabled status', async () => {
    const wrapper = mountView()
    await flushPromises()

    const statusHeader = wrapper.findAll('th')[2]
    await statusHeader.trigger('click')
    await nextTick()

    const rows = wrapper.findAll('tbody tr')
    expect(rows).toHaveLength(3)
    expect(rows[0].text()).toContain('Disabled')
  })

  it('sorts by challenges stat', async () => {
    const wrapper = mountView()
    await flushPromises()
    const statsStore = useStatsStore()
    statsStore.keysSummary = { ...mockSummary }
    await nextTick()

    const challengesHeader = wrapper.findAll('th')[3]
    await challengesHeader.trigger('click')
    await nextTick()

    const rows = wrapper.findAll('tbody tr')
    expect(rows).toHaveLength(3)
    expect(rows[0].text()).toContain('Charlie')
    expect(rows[2].text()).toContain('Alpha')
  })

  it('sorts by verified stat descending', async () => {
    const wrapper = mountView()
    await flushPromises()
    const statsStore = useStatsStore()
    statsStore.keysSummary = { ...mockSummary }
    await nextTick()

    // Click verified header twice for descending
    const verifiedHeader = wrapper.findAll('th')[4]
    await verifiedHeader.trigger('click')
    await nextTick()
    await verifiedHeader.trigger('click')
    await nextTick()

    const rows = wrapper.findAll('tbody tr')
    expect(rows).toHaveLength(3)
    // desc: Charlie(80) > Bravo(30) > Alpha(10)
    expect(rows[0].text()).toContain('Charlie')
    expect(rows[2].text()).toContain('Alpha')
  })

  it('sorts by failed stat', async () => {
    const wrapper = mountView()
    await flushPromises()
    const statsStore = useStatsStore()
    statsStore.keysSummary = { ...mockSummary }
    await nextTick()

    const failedHeader = wrapper.findAll('th')[5]
    await failedHeader.trigger('click')
    await nextTick()

    const rows = wrapper.findAll('tbody tr')
    expect(rows).toHaveLength(3)
    expect(rows[0].text()).toContain('Charlie')
    expect(rows[2].text()).toContain('Alpha')
  })

  it('shows sort indicator on active column', async () => {
    const wrapper = mountView()
    await flushPromises()
    const nameHeader = wrapper.findAll('th')[0]
    expect(nameHeader.text()).toContain('\u25B2')
  })

  it('fetches keys and summary on mount', async () => {
    mountView()
    await flushPromises()
    expect(mockApi.get).toHaveBeenCalledWith('/keys')
    expect(mockApi.get).toHaveBeenCalledWith('/stats/keys-summary')
  })
})
