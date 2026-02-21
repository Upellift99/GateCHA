import { defineStore } from 'pinia'
import { ref } from 'vue'
import api from '../lib/api'

export interface APIKey {
  id: number
  key_id: string
  hmac_secret?: string
  name: string
  domain: string
  max_number: number
  expire_seconds: number
  algorithm: string
  enabled: boolean
  created_at: string
  updated_at: string
}

export const useApiKeysStore = defineStore('apikeys', () => {
  const keys = ref<APIKey[]>([])
  const loading = ref(false)

  async function fetchKeys() {
    loading.value = true
    try {
      const { data } = await api.get('/keys')
      keys.value = data.keys
    } finally {
      loading.value = false
    }
  }

  async function createKey(payload: { name: string; domain: string; max_number: number; expire_seconds: number; algorithm?: string }) {
    const { data } = await api.post('/keys', payload)
    await fetchKeys()
    return data as APIKey
  }

  async function getKey(id: number) {
    const { data } = await api.get(`/keys/${id}`)
    return data as APIKey
  }

  async function updateKey(id: number, payload: Partial<APIKey>) {
    const { data } = await api.put(`/keys/${id}`, payload)
    await fetchKeys()
    return data as APIKey
  }

  async function deleteKey(id: number) {
    await api.delete(`/keys/${id}`)
    await fetchKeys()
  }

  async function rotateSecret(id: number) {
    const { data } = await api.post(`/keys/${id}/rotate-secret`)
    return data.hmac_secret as string
  }

  return { keys, loading, fetchKeys, createKey, getKey, updateKey, deleteKey, rotateSecret }
})
