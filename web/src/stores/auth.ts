import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import api from '../lib/api'

export const useAuthStore = defineStore('auth', () => {
  const token = ref(localStorage.getItem('gatecha_token') || '')
  const isAuthenticated = computed(() => !!token.value)

  async function login(username: string, password: string, altchaPayload?: string) {
    const body: Record<string, string> = { username, password }
    if (altchaPayload) {
      body.altcha_payload = altchaPayload
    }
    const { data } = await api.post('/login', body)
    token.value = data.token
    localStorage.setItem('gatecha_token', data.token)
  }

  function logout() {
    token.value = ''
    localStorage.removeItem('gatecha_token')
  }

  async function checkAuth() {
    try {
      await api.get('/me')
      return true
    } catch {
      logout()
      return false
    }
  }

  return { token, isAuthenticated, login, logout, checkAuth }
})
