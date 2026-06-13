import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import api from '../lib/api'
import type { HISSignals } from '../lib/his'

export const useAuthStore = defineStore('auth', () => {
  const token = ref(localStorage.getItem('gatecha_token') || '')
  const isAuthenticated = computed(() => !!token.value)
  const version = ref('')

  async function login(username: string, password: string, altchaPayload?: string, hisSignals?: HISSignals) {
    const body: Record<string, unknown> = { username, password }
    if (altchaPayload) {
      body.altcha_payload = altchaPayload
    }
    if (hisSignals) {
      body.his_signals = hisSignals
    }
    const { data } = await api.post('/login', body)
    token.value = data.token
    localStorage.setItem('gatecha_token', data.token)
  }

  function logout() {
    token.value = ''
    version.value = ''
    localStorage.removeItem('gatecha_token')
  }

  // fetchVersion loads the build version once (authenticated endpoint). Silently
  // no-ops if it fails so the footer simply omits the version.
  async function fetchVersion() {
    if (version.value) return
    try {
      const { data } = await api.get('/version')
      version.value = data.version ?? ''
    } catch {
      version.value = ''
    }
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

  return { token, isAuthenticated, version, login, logout, checkAuth, fetchVersion }
})
