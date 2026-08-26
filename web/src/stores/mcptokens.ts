import { defineStore } from 'pinia'
import { ref } from 'vue'
import api from '../lib/api'

export interface MCPToken {
  id: number
  name: string
  display: string
  read_only: boolean
  last_used_at: string | null
  created_at: string
}

export const useMCPTokensStore = defineStore('mcptokens', () => {
  const tokens = ref<MCPToken[]>([])
  const loading = ref(false)

  async function fetchTokens() {
    loading.value = true
    try {
      const { data } = await api.get('/mcp-tokens')
      // Never let a malformed body leave `tokens` undefined: the panel renders
      // straight off its length.
      tokens.value = data.tokens ?? []
    } finally {
      loading.value = false
    }
  }

  // The secret is returned only here; it cannot be read back afterwards.
  async function createToken(payload: { name: string; read_only: boolean }) {
    const { data } = await api.post('/mcp-tokens', payload)
    await fetchTokens()
    return data.secret as string
  }

  async function revokeToken(id: number) {
    await api.delete(`/mcp-tokens/${id}`)
    await fetchTokens()
  }

  return { tokens, loading, fetchTokens, createToken, revokeToken }
})
