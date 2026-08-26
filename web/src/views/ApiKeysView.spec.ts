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

  it('copies a key id to clipboard from the list', async () => {
    const writeText = vi.fn()
    Object.defineProperty(navigator, 'clipboard', { value: { writeText }, configurable: true })

    const wrapper = mountView()
    await flushPromises()

    // Alpha (gk_bbb) sorts first
    const copyBtn = wrapper.find('tbody button[aria-label="Copy gk_bbb"]')
    expect(copyBtn.exists()).toBe(true)
    await copyBtn.trigger('click')

    expect(writeText).toHaveBeenCalledWith('gk_bbb')
    expect(copyBtn.text()).toBe('Copied!') // tooltip
  })

  it('fetches keys and summary on mount', async () => {
    mountView()
    await flushPromises()
    expect(mockApi.get).toHaveBeenCalledWith('/keys')
    expect(mockApi.get).toHaveBeenCalledWith('/stats/keys-summary')
  })

  // --- search + pagination (issue #136) ---

  // 60 keys named Key 001..Key 060, each on its own domain, so paging and
  // filtering can be exercised past the default page size of 25.
  const manyKeys = Array.from({ length: 60 }, (_, i) => {
    const n = String(i + 1).padStart(3, '0')
    return {
      id: 100 + i,
      key_id: `gk_${n}`,
      name: `Key ${n}`,
      domain: `site${n}.example.com`,
      max_number: 100,
      expire_seconds: 60,
      algorithm: 'SHA-256',
      enabled: true,
      created_at: '',
      updated_at: '',
    }
  })

  function mockManyKeys() {
    mockApi.get.mockImplementation((url: string) => {
      if (url === '/keys') return Promise.resolve({ data: { keys: [...manyKeys] } })
      if (url === '/stats/keys-summary') return Promise.resolve({ data: { keys: {} } })
      return Promise.resolve({ data: {} })
    })
  }

  async function setSearch(wrapper: ReturnType<typeof mountView>, value: string) {
    await wrapper.find('input#key-search').setValue(value)
    await nextTick()
  }

  it('filters by name, case-insensitively', async () => {
    const wrapper = mountView()
    await flushPromises()
    await setSearch(wrapper, 'alp')

    const rows = wrapper.findAll('tbody tr')
    expect(rows).toHaveLength(1)
    expect(rows[0].text()).toContain('Alpha')
  })

  it('filters by domain', async () => {
    const wrapper = mountView()
    await flushPromises()
    await setSearch(wrapper, 'b.com')

    const rows = wrapper.findAll('tbody tr')
    expect(rows).toHaveLength(1)
    expect(rows[0].text()).toContain('Bravo')
  })

  it('filters by key id', async () => {
    const wrapper = mountView()
    await flushPromises()
    await setSearch(wrapper, 'GK_CCC')

    const rows = wrapper.findAll('tbody tr')
    expect(rows).toHaveLength(1)
    expect(rows[0].text()).toContain('Charlie')
  })

  it('ignores surrounding whitespace in the search term', async () => {
    const wrapper = mountView()
    await flushPromises()
    await setSearch(wrapper, '  alpha  ')

    expect(wrapper.findAll('tbody tr')).toHaveLength(1)
  })

  it('shows a no-results state distinct from the empty state', async () => {
    const wrapper = mountView()
    await flushPromises()
    await setSearch(wrapper, 'nothingmatches')

    expect(wrapper.find('table').exists()).toBe(false)
    expect(wrapper.text()).toContain('No keys match "nothingmatches"')
    expect(wrapper.text()).not.toContain('No API keys yet')
  })

  it('clears the search from the no-results state', async () => {
    const wrapper = mountView()
    await flushPromises()
    await setSearch(wrapper, 'nothingmatches')

    const clear = wrapper.findAll('button').find(b => b.text() === 'Clear search')
    expect(clear).toBeDefined()
    await clear!.trigger('click')
    await nextTick()

    expect(wrapper.findAll('tbody tr')).toHaveLength(3)
  })

  it('clears the search from the input clear button', async () => {
    const wrapper = mountView()
    await flushPromises()
    await setSearch(wrapper, 'alpha')

    await wrapper.find('button[aria-label="Clear search"]').trigger('click')
    await nextTick()

    expect(wrapper.findAll('tbody tr')).toHaveLength(3)
  })

  it('does not paginate when keys fit on one page', async () => {
    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.findAll('tbody tr')).toHaveLength(3)
    expect(wrapper.find('button[aria-label="Next page"]').exists()).toBe(false)
    expect(wrapper.text()).toContain('Showing 1–3 of 3')
  })

  it('caps the first page at the default page size', async () => {
    mockManyKeys()
    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.findAll('tbody tr')).toHaveLength(25)
    expect(wrapper.text()).toContain('Showing 1–25 of 60')
    expect(wrapper.text()).toContain('Page 1 of 3')
  })

  it('navigates between pages', async () => {
    mockManyKeys()
    const wrapper = mountView()
    await flushPromises()

    await wrapper.find('button[aria-label="Next page"]').trigger('click')
    await nextTick()

    expect(wrapper.text()).toContain('Showing 26–50 of 60')
    const rows = wrapper.findAll('tbody tr')
    expect(rows[0].text()).toContain('Key 026')

    await wrapper.find('button[aria-label="Previous page"]').trigger('click')
    await nextTick()
    expect(wrapper.findAll('tbody tr')[0].text()).toContain('Key 001')
  })

  it('shows the remainder on the last page and disables Next', async () => {
    mockManyKeys()
    const wrapper = mountView()
    await flushPromises()

    const next = wrapper.find('button[aria-label="Next page"]')
    await next.trigger('click')
    await nextTick()
    await next.trigger('click')
    await nextTick()

    expect(wrapper.findAll('tbody tr')).toHaveLength(10)
    expect(wrapper.text()).toContain('Showing 51–60 of 60')
    expect(next.attributes('disabled')).toBeDefined()
  })

  it('disables Previous on the first page', async () => {
    mockManyKeys()
    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.find('button[aria-label="Previous page"]').attributes('disabled')).toBeDefined()
  })

  it('sorts across the whole set, not just the current page', async () => {
    mockManyKeys()
    const wrapper = mountView()
    await flushPromises()

    // Descending by name: page 1 must start at the global maximum (Key 060),
    // which lives on the last page when sorted ascending.
    await wrapper.findAll('th')[0].trigger('click')
    await nextTick()

    expect(wrapper.findAll('tbody tr')[0].text()).toContain('Key 060')

    await wrapper.find('button[aria-label="Next page"]').trigger('click')
    await nextTick()
    expect(wrapper.findAll('tbody tr')[0].text()).toContain('Key 035')
  })

  it('resets to the first page when the search changes', async () => {
    mockManyKeys()
    const wrapper = mountView()
    await flushPromises()

    await wrapper.find('button[aria-label="Next page"]').trigger('click')
    await nextTick()
    expect(wrapper.text()).toContain('Page 2 of 3')

    // 'Key 0' matches everything, so the result set still spans 3 pages:
    // landing back on page 1 has to come from the reset, not from clamping.
    await setSearch(wrapper, 'Key 0')
    expect(wrapper.text()).toContain('Page 1 of 3')
    expect(wrapper.findAll('tbody tr')[0].text()).toContain('Key 001')
  })

  it('resets to the first page when the page size changes', async () => {
    mockManyKeys()
    const wrapper = mountView()
    await flushPromises()

    await wrapper.find('button[aria-label="Next page"]').trigger('click')
    await nextTick()

    await wrapper.find('select#key-page-size').setValue('50')
    await nextTick()

    expect(wrapper.findAll('tbody tr')).toHaveLength(50)
    expect(wrapper.text()).toContain('Showing 1–50 of 60')
  })

  it('clamps the page when the underlying key list shrinks', async () => {
    mockManyKeys()
    const wrapper = mountView()
    await flushPromises()

    const next = wrapper.find('button[aria-label="Next page"]')
    await next.trigger('click')
    await nextTick()
    await next.trigger('click')
    await nextTick()
    expect(wrapper.text()).toContain('Page 3 of 3')

    // A refetch after keys were deleted elsewhere must not strand the view on a
    // page that no longer exists. The search is untouched, so nothing resets it.
    const store = useApiKeysStore()
    store.keys = manyKeys.slice(0, 30)
    await nextTick()

    expect(wrapper.text()).toContain('Page 2 of 2')
    expect(wrapper.text()).toContain('Showing 26–30 of 30')
    expect(wrapper.findAll('tbody tr')).toHaveLength(5)
  })

  it('reports the filtered total against the full count', async () => {
    mockManyKeys()
    const wrapper = mountView()
    await flushPromises()
    await setSearch(wrapper, 'Key 01')

    expect(wrapper.text()).toContain('Showing 1–10 of 10')
    expect(wrapper.text()).toContain('filtered from 60')
  })

  it('hides the page-size control for short lists', async () => {
    const wrapper = mountView()
    await flushPromises()
    expect(wrapper.find('select#key-page-size').exists()).toBe(false)
  })

  it('keeps the page-size control reachable when a larger size fits on one page', async () => {
    mockManyKeys()
    const wrapper = mountView()
    await flushPromises()

    await wrapper.find('select#key-page-size').setValue('100')
    await nextTick()

    expect(wrapper.findAll('tbody tr')).toHaveLength(60)
    expect(wrapper.find('button[aria-label="Next page"]').exists()).toBe(false)
    expect(wrapper.find('select#key-page-size').exists()).toBe(true)
  })

  it('hides the search bar when there are no keys at all', async () => {
    mockApi.get.mockImplementation((url: string) => {
      if (url === '/keys') return Promise.resolve({ data: { keys: [] } })
      if (url === '/stats/keys-summary') return Promise.resolve({ data: { keys: {} } })
      return Promise.resolve({ data: {} })
    })
    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.find('input#key-search').exists()).toBe(false)
    expect(wrapper.text()).toContain('No API keys yet')
  })
})
