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

const chartData = computed(() => {
  const sorted = [...props.data].sort((a, b) => a.date.localeCompare(b.date))
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
        borderColor: '#22c55e',
        backgroundColor: 'rgba(34, 197, 94, 0.1)',
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
