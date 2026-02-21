<script setup lang="ts">
import { onMounted } from 'vue'
import { useStatsStore } from '../stores/stats'
import StatsChart from '../components/StatsChart.vue'
import StatsSummaryCard from '../components/StatsSummaryCard.vue'

const statsStore = useStatsStore()

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
