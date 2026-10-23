<script setup lang="ts">
import { computed } from 'vue'
import { Line } from 'vue-chartjs'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  Filler,
} from 'chart.js'
import type { DailyStat } from '../stores/stats'

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, Title, Tooltip, Legend, Filler)

const props = defineProps<{
  data: DailyStat[]
}>()

// The API only returns days that had traffic. Plotting those rows as-is draws
// a line straight across quiet days, and a single day collapses the chart, so
// the window is filled with zero days. Dates are UTC, like the stored stats.
const DAYS = 30

function lastDays(n: number): string[] {
  const today = new Date()
  return Array.from({ length: n }, (_, i) => {
    const d = new Date(Date.UTC(today.getUTCFullYear(), today.getUTCMonth(), today.getUTCDate() - (n - 1 - i)))
    return d.toISOString().slice(0, 10)
  })
}

const chartData = computed(() => {
  const byDate = new Map(props.data.map((d) => [d.date, d]))
  const sorted = lastDays(DAYS).map(
    (date): DailyStat =>
      byDate.get(date) ?? {
        date,
        challenges_issued: 0,
        verifications_ok: 0,
        verifications_fail: 0,
        his_observations: 0,
        his_bot_suspected: 0,
      },
  )
  return {
    labels: sorted.map((d) => d.date.slice(5)),
    datasets: [
      {
        label: 'Challenges',
        data: sorted.map((d) => d.challenges_issued),
        borderColor: '#6366f1',
        backgroundColor: 'rgba(99, 102, 241, 0.1)',
        fill: true,
        tension: 0.3,
      },
      {
        label: 'Verified OK',
        data: sorted.map((d) => d.verifications_ok),
        borderColor: '#15803d', // green-700: a line needs 3:1 on white
        backgroundColor: 'rgba(21, 128, 61, 0.08)',
        fill: true,
        tension: 0.3,
      },
      {
        label: 'Failed',
        data: sorted.map((d) => d.verifications_fail),
        borderColor: '#ef4444',
        backgroundColor: 'rgba(239, 68, 68, 0.1)',
        fill: true,
        tension: 0.3,
      },
    ],
  }
})

const chartOptions = {
  responsive: true,
  interaction: {
    intersect: false,
    mode: 'index' as const,
  },
  plugins: {
    legend: {
      position: 'bottom' as const,
    },
  },
  scales: {
    y: {
      beginAtZero: true,
    },
  },
}
</script>

<template>
  <Line :data="chartData" :options="chartOptions" />
</template>
