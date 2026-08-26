import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { nextTick } from 'vue'
import { createPinia, setActivePinia } from 'pinia'
import { useSettingsStore } from '../stores/settings'
import { useMCPTokensStore } from '../stores/mcptokens'
import MCPTokensPanel from './MCPTokensPanel.vue'

const mockApi = vi.hoisted(() => ({
  get: vi.fn(),
  post: vi.fn(),
  put: vi.fn(),
  delete: vi.fn(),
}))

vi.mock('../lib/api', () => ({ default: mockApi }))

const mockTokens = [
  { id: 1, name: 'Martijn laptop', display: 'gm_a1b2c3d4', read_only: false, last_used_at: '2026-08-25T10:00:00Z', created_at: '2026-08-20T10:00:00Z' },
  { id: 2, name: 'CI', display: 'gm_9f8e7d6c', read_only: true, last_used_at: null, created_at: '2026-08-21T10:00:00Z' },
]

function mountPanel() {
  return mount(MCPTokensPanel)
}

describe('MCPTokensPanel', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    mockApi.get.mockImplementation((url: string) => {
      if (url === '/mcp-tokens') return Promise.resolve({ data: { tokens: [...mockTokens] } })
      return Promise.resolve({ data: {} })
    })
  })

  it('lists tokens with their display prefix and access level', async () => {
    const wrapper = mountPanel()
    await flushPromises()

    const rows = wrapper.findAll('tbody tr')
    expect(rows).toHaveLength(2)
    expect(rows[0].text()).toContain('Martijn laptop')
    expect(rows[0].text()).toContain('gm_a1b2c3d4')
    expect(rows[0].text()).toContain('Full access')
    expect(rows[1].text()).toContain('Read only')
  })

  it('shows Never for a token that has not been used', async () => {
    const wrapper = mountPanel()
    await flushPromises()
    expect(wrapper.findAll('tbody tr')[1].text()).toContain('Never')
  })

  it('nudges toward one token per person when the list is empty', async () => {
    mockApi.get.mockResolvedValue({ data: { tokens: [] } })
    const wrapper = mountPanel()
    await flushPromises()

    expect(wrapper.find('tbody').exists()).toBe(false)
    expect(wrapper.text()).toContain('one per person')
  })

  it('creates a token and reveals the secret once', async () => {
    const secret = 'gm_a1b2c3d4e5f60718293a4b5c6d7e8f90'
    mockApi.post.mockResolvedValue({ data: { secret, token: mockTokens[0] } })

    const wrapper = mountPanel()
    await flushPromises()

    await wrapper.find('input#mcpTokenName').setValue('New laptop')
    await wrapper.find('form').trigger('submit')
    await flushPromises()

    expect(mockApi.post).toHaveBeenCalledWith('/mcp-tokens', { name: 'New laptop', read_only: false })
    expect(wrapper.text()).toContain(secret)
    expect(wrapper.text()).toContain('will not be shown again')

    // Dismissing clears it from the DOM: it cannot be fetched back.
    const done = wrapper.findAll('button').find(b => b.text() === 'Done')
    await done!.trigger('click')
    await nextTick()
    expect(wrapper.text()).not.toContain(secret)
  })

  it('passes the read-only flag through', async () => {
    mockApi.post.mockResolvedValue({ data: { secret: 'gm_x', token: mockTokens[1] } })

    const wrapper = mountPanel()
    await flushPromises()

    await wrapper.find('input#mcpTokenName').setValue('CI')
    await wrapper.find('input[type="checkbox"]:not(.sr-only)').setValue(true)
    await wrapper.find('form').trigger('submit')
    await flushPromises()

    expect(mockApi.post).toHaveBeenCalledWith('/mcp-tokens', { name: 'CI', read_only: true })
  })

  it('trims the name and refuses a blank one', async () => {
    mockApi.post.mockResolvedValue({ data: { secret: 'gm_x', token: mockTokens[0] } })

    const wrapper = mountPanel()
    await flushPromises()

    await wrapper.find('input#mcpTokenName').setValue('   ')
    await wrapper.find('form').trigger('submit')
    await flushPromises()
    expect(mockApi.post).not.toHaveBeenCalled()

    await wrapper.find('input#mcpTokenName').setValue('  Padded  ')
    await wrapper.find('form').trigger('submit')
    await flushPromises()
    expect(mockApi.post).toHaveBeenCalledWith('/mcp-tokens', { name: 'Padded', read_only: false })
  })

  it('confirms before revoking and skips the call when declined', async () => {
    globalThis.confirm = vi.fn(() => false)

    const wrapper = mountPanel()
    await flushPromises()

    await wrapper.find('button[aria-label="Revoke Martijn laptop"]').trigger('click')
    await flushPromises()

    expect(globalThis.confirm).toHaveBeenCalled()
    expect(mockApi.delete).not.toHaveBeenCalled()
  })

  it('revokes a single token when confirmed', async () => {
    globalThis.confirm = vi.fn(() => true)
    mockApi.delete.mockResolvedValue({ data: {} })

    const wrapper = mountPanel()
    await flushPromises()

    await wrapper.find('button[aria-label="Revoke CI"]').trigger('click')
    await flushPromises()

    expect(mockApi.delete).toHaveBeenCalledWith('/mcp-tokens/2')
  })

  it('toggles the endpoint through the settings store', async () => {
    mockApi.put.mockResolvedValue({ data: { login_captcha_enabled: false, mcp_enabled: true } })

    const wrapper = mountPanel()
    await flushPromises()

    await wrapper.find('input#mcpToggle').setValue(true)
    await flushPromises()

    expect(mockApi.put).toHaveBeenCalledWith('/settings', { mcp_enabled: true })
    expect(useSettingsStore().settings.mcp_enabled).toBe(true)
  })

  it('reverts the toggle and reports when the update fails', async () => {
    mockApi.put.mockRejectedValue(new Error('nope'))

    const wrapper = mountPanel()
    await flushPromises()

    await wrapper.find('input#mcpToggle').setValue(true)
    await flushPromises()

    expect(wrapper.text()).toContain('Failed to update setting.')
    // The panel refetches so the switch reflects the server, not the click.
    expect(mockApi.get).toHaveBeenCalledWith('/settings')
  })

  it('surfaces a failed creation without revealing a secret', async () => {
    mockApi.post.mockRejectedValue(new Error('nope'))

    const wrapper = mountPanel()
    await flushPromises()

    await wrapper.find('input#mcpTokenName').setValue('Nope')
    await wrapper.find('form').trigger('submit')
    await flushPromises()

    expect(wrapper.text()).toContain('Failed to create token.')
    expect(wrapper.text()).not.toContain('will not be shown again')
  })

  it('never renders a token digest', async () => {
    const wrapper = mountPanel()
    await flushPromises()

    expect(wrapper.html()).not.toContain('token_hash')
    expect(useMCPTokensStore().tokens[0]).not.toHaveProperty('token_hash')
  })
})
