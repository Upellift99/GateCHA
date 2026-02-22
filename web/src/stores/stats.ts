import { defineStore } from 'pinia'
import { ref } from 'vue'
import api from '../lib/api'

export interface DailyStat {
  date: string
  challenges_issued: number
  verifications_ok: number
  verifications_fail: number
}

export interface StatsOverview {
  total_challenges: number
  total_verifications_ok: number
  total_verifications_fail: number
  active_keys: number
  daily: DailyStat[]
}

export interface KeyStatsSummary {
  api_key_id: number
  challenges_issued: number
  verifications_ok: number
  verifications_fail: number
}

export const useStatsStore = defineStore('stats', () => {
  const overview = ref<StatsOverview | null>(null)
  const keyStats = ref<DailyStat[]>([])
  const keysSummary = ref<Record<string, KeyStatsSummary>>({})

  async function fetchOverview(days = 30) {
    const { data } = await api.get(`/stats/overview?days=${days}`)
    overview.value = data
  }

  async function fetchKeyStats(keyId: number, days = 30) {
    const { data } = await api.get(`/stats/keys/${keyId}?days=${days}`)
    keyStats.value = data.days
  }

  async function fetchKeysSummary() {
    const { data } = await api.get('/stats/keys-summary')
    keysSummary.value = data.keys
  }

  return { overview, keyStats, keysSummary, fetchOverview, fetchKeyStats, fetchKeysSummary }
})
