<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useStatsStore } from '../stores/stats'
import StatsChart from '../components/StatsChart.vue'
import StatsSummaryCard from '../components/StatsSummaryCard.vue'
import CountryTraffic from '../components/CountryTraffic.vue'

const statsStore = useStatsStore()

const hisSuspectedPct = computed(() => {
  const o = statsStore.overview
  if (!o || o.total_his_observations === 0) return 0
  return Math.round((o.total_his_bot_suspected / o.total_his_observations) * 100)
})

const successRate = computed(() => {
  const o = statsStore.overview
  if (!o) return 0
  const total = o.total_verifications_ok + o.total_verifications_fail
  return total === 0 ? 0 : Math.round((o.total_verifications_ok / total) * 100)
})

onMounted(() => {
  statsStore.fetchOverview(30)
})
</script>

<template>
  <div>
    <div class="mb-6">
      <h1 class="text-2xl font-bold tracking-tight text-slate-900">Dashboard</h1>
      <p class="mt-1 text-sm text-slate-500">Activity across all your API keys over the last 30 days.</p>
    </div>

    <div v-if="statsStore.overview" class="space-y-6">
      <!-- Summary Cards -->
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <StatsSummaryCard label="Total Challenges" :value="statsStore.overview.total_challenges" color="brand">
          <template #icon>
            <svg class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
              <path fill-rule="evenodd" d="M9.661 2.237a.531.531 0 0 1 .678 0 11.947 11.947 0 0 0 7.078 2.749.5.5 0 0 1 .479.425c.069.52.104 1.05.104 1.59 0 5.162-3.26 9.563-7.834 11.256a.48.48 0 0 1-.332 0C5.16 16.564 1.9 12.163 1.9 7c0-.538.035-1.069.104-1.589a.5.5 0 0 1 .48-.425 11.947 11.947 0 0 0 7.077-2.75Zm4.196 5.954a.75.75 0 0 0-1.214-.882l-3.236 4.53-1.55-1.55a.75.75 0 0 0-1.06 1.061l2.18 2.18a.75.75 0 0 0 1.137-.089l3.743-5.25Z" clip-rule="evenodd" />
            </svg>
          </template>
        </StatsSummaryCard>
        <StatsSummaryCard label="Verifications OK" :value="statsStore.overview.total_verifications_ok" color="green" :hint="`${successRate}% success rate`">
          <template #icon>
            <svg class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
              <path fill-rule="evenodd" d="M16.704 4.153a.75.75 0 0 1 .143 1.052l-8 10.5a.75.75 0 0 1-1.127.075l-4.5-4.5a.75.75 0 0 1 1.06-1.06l3.894 3.893 7.48-9.817a.75.75 0 0 1 1.05-.143Z" clip-rule="evenodd" />
            </svg>
          </template>
        </StatsSummaryCard>
        <StatsSummaryCard label="Verifications Failed" :value="statsStore.overview.total_verifications_fail" color="red">
          <template #icon>
            <svg class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
              <path d="M6.28 5.22a.75.75 0 0 0-1.06 1.06L8.94 10l-3.72 3.72a.75.75 0 1 0 1.06 1.06L10 11.06l3.72 3.72a.75.75 0 1 0 1.06-1.06L11.06 10l3.72-3.72a.75.75 0 0 0-1.06-1.06L10 8.94 6.28 5.22Z" />
            </svg>
          </template>
        </StatsSummaryCard>
        <StatsSummaryCard label="Active Keys" :value="statsStore.overview.active_keys" color="purple">
          <template #icon>
            <svg class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
              <path fill-rule="evenodd" d="M8 7a5 5 0 1 1 3.61 4.804l-1.903 1.903A1 1 0 0 1 9 14H8v1a1 1 0 0 1-1 1H6v1a1 1 0 0 1-1 1H3a1 1 0 0 1-1-1v-2a1 1 0 0 1 .293-.707L8.196 8.39A5.002 5.002 0 0 1 8 7Zm5-3a.75.75 0 0 0 0 1.5A1.5 1.5 0 0 1 14.5 7 .75.75 0 0 0 16 7a3 3 0 0 0-3-3Z" clip-rule="evenodd" />
            </svg>
          </template>
        </StatsSummaryCard>
      </div>

      <!-- HIS (Human Interaction Signature) — Monitor mode -->
      <div class="rounded-xl border border-slate-200 bg-white p-6 shadow-sm">
        <div class="flex items-center justify-between mb-4">
          <h2 class="text-lg font-semibold text-slate-900">Human Interaction Signature</h2>
          <span class="text-xs font-medium px-2.5 py-1 rounded-full bg-amber-50 text-amber-700 ring-1 ring-amber-200">Monitor mode</span>
        </div>
        <p class="text-sm text-slate-500 mb-4">
          Observes interaction behavior to estimate automation probability. In Monitor mode it only records — it never blocks a verification.
        </p>
        <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
          <StatsSummaryCard label="HIS Observations" :value="statsStore.overview.total_his_observations" color="blue" />
          <StatsSummaryCard label="Bot Suspected" :value="statsStore.overview.total_his_bot_suspected" color="amber" />
          <StatsSummaryCard label="Bot Suspected" :value="hisSuspectedPct" suffix="%" color="red" hint="share of observed samples" />
        </div>
      </div>

      <!-- Chart + Traffic by country -->
      <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div class="rounded-xl border border-slate-200 bg-white p-6 shadow-sm lg:col-span-2">
          <h2 class="text-lg font-semibold text-slate-900 mb-4">Last 30 Days</h2>
          <StatsChart v-if="statsStore.overview.daily?.length" :data="statsStore.overview.daily" />
          <p v-else class="text-slate-400 text-center py-12">No data yet</p>
        </div>
        <CountryTraffic :countries="statsStore.overview.countries ?? []" />
      </div>
    </div>

    <div v-else class="text-center py-12 text-slate-400">Loading…</div>
  </div>
</template>
