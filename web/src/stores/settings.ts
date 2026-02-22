import { defineStore } from 'pinia'
import { ref } from 'vue'
import api from '../lib/api'

export interface Settings {
  login_captcha_enabled: boolean
}

export const useSettingsStore = defineStore('settings', () => {
  const settings = ref<Settings>({ login_captcha_enabled: false })
  const loading = ref(false)

  async function fetchSettings() {
    loading.value = true
    try {
      const { data } = await api.get('/settings')
      settings.value = data
    } finally {
      loading.value = false
    }
  }

  async function updateSettings(payload: Partial<Settings>) {
    const { data } = await api.put('/settings', payload)
    settings.value = data
  }

  return { settings, loading, fetchSettings, updateSettings }
})
