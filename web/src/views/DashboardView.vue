<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useStatsStore } from '../stores/stats'
import StatsChart from '../components/StatsChart.vue'
import StatsSummaryCard from '../components/StatsSummaryCard.vue'

const statsStore = useStatsStore()

const hisSuspectedPct = computed(() => {
  const o = statsStore.overview
  if (!o || o.total_his_observations === 0) return 0
  return Math.round((o.total_his_bot_suspected / o.total_his_observations) * 100)
})

onMounted(() => {
  statsStore.fetchOverview(30)
})
</script>

<template>
  <div>
    <h1 class="text-2xl font-bold text-gray-900 mb-6">Dashboard</h1>

    <div v-if="statsStore.overview" class="space-y-6">
      <!-- Summary Cards -->
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <StatsSummaryCard
          label="Total Challenges"
          :value="statsStore.overview.total_challenges"
          color="blue"
        />
        <StatsSummaryCard
          label="Verifications OK"
          :value="statsStore.overview.total_verifications_ok"
          color="green"
        />
        <StatsSummaryCard
          label="Verifications Failed"
          :value="statsStore.overview.total_verifications_fail"
          color="red"
        />
        <StatsSummaryCard
          label="Active Keys"
          :value="statsStore.overview.active_keys"
          color="purple"
        />
      </div>

      <!-- HIS (Human Interaction Signature) — Monitor mode -->
      <div class="bg-white shadow rounded-lg p-6">
        <div class="flex items-center justify-between mb-4">
          <h2 class="text-lg font-medium text-gray-900">Human Interaction Signature</h2>
          <span class="text-xs font-medium px-2 py-1 rounded bg-amber-50 text-amber-700">Monitor mode</span>
        </div>
        <p class="text-sm text-gray-500 mb-4">
          Observes interaction behavior to estimate automation probability. In Monitor mode it only records — it never blocks a verification.
        </p>
        <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
          <StatsSummaryCard
            label="HIS Observations"
            :value="statsStore.overview.total_his_observations"
            color="blue"
          />
          <StatsSummaryCard
            label="Bot Suspected"
            :value="statsStore.overview.total_his_bot_suspected"
            color="amber"
          />
          <StatsSummaryCard
            label="Bot Suspected %"
            :value="hisSuspectedPct"
            color="red"
          />
        </div>
      </div>

      <!-- Chart -->
      <div class="bg-white shadow rounded-lg p-6">
        <h2 class="text-lg font-medium text-gray-900 mb-4">Last 30 Days</h2>
        <StatsChart
          v-if="statsStore.overview.daily?.length"
          :data="statsStore.overview.daily"
        />
        <p v-else class="text-gray-500 text-center py-12">No data yet</p>
      </div>
    </div>

    <div v-else class="text-center py-12 text-gray-500">Loading...</div>
  </div>
</template>
