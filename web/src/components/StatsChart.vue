<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
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
import { lastDays, lineVars } from '../lib/lines'

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, Title, Tooltip, Legend, Filler)

const props = defineProps<{
  data: DailyStat[]
  /** The key's line: its challenges series takes that line's colour. */
  lineId?: number
}>()

// Chart.js paints on canvas and cannot read CSS variables, so resolve the
// theme tokens here and read them again when the system theme flips.
let media: MediaQueryList | undefined
const onThemeChange = () => {
  colours.value = readColours()
}
onMounted(() => {
  media = globalThis.matchMedia?.('(prefers-color-scheme: dark)')
  media?.addEventListener('change', onThemeChange)
})
onBeforeUnmount(() => media?.removeEventListener('change', onThemeChange))

function token(name: string, fallback: string): string {
  const v = getComputedStyle(document.documentElement).getPropertyValue(name).trim()
  return v || fallback
}

function resolve(cssValue: string, fallback: string): string {
  const m = /^var\((--[\w-]+)\)$/.exec(cssValue)
  return m ? token(m[1]!, fallback) : cssValue
}

function readColours() {
  return {
    challenges: props.lineId ? resolve(lineVars(props.lineId).stroke, '#1F4FC1') : token('--color-link', '#1F4FC1'),
    ok: token('--color-ok', '#00754E'),
    fail: token('--color-fail', '#C8001A'),
    muted: token('--color-muted', '#5B6478'),
    rule: token('--color-rule-soft', '#EBEEF2'),
  }
}

const colours = ref(readColours())
watch(() => props.lineId, onThemeChange)

// The API only returns days that had traffic. Plotting those rows as-is draws
// a line straight across quiet days, and a single day collapses the chart, so
// the window is filled with zero days. Dates are UTC, like the stored stats.
const DAYS = 30

const chartData = computed(() => {
  const byDate = new Map(props.data.map((d) => [d.date, d]))
  const sorted = lastDays(DAYS).map(
    (date): DailyStat =>
      byDate.get(date) ?? {
        api_key_id: 0,
        date,
        challenges_issued: 0,
        verifications_ok: 0,
        verifications_fail: 0,
        his_observations: 0,
        his_bot_suspected: 0,
      },
  )
  return {
    labels: sorted.map((d) =>
      new Date(d.date + 'T00:00:00Z').toLocaleDateString('en-US', { month: 'short', day: 'numeric', timeZone: 'UTC' }),
    ),
    datasets: [
      {
        label: 'Challenges',
        data: sorted.map((d) => d.challenges_issued),
        borderColor: colours.value.challenges,
        backgroundColor: colours.value.challenges,
        fill: false,
        tension: 0,
        pointRadius: 0,
        pointHoverRadius: 3,
        borderWidth: 2.5,
      },
      {
        label: 'Verified OK',
        data: sorted.map((d) => d.verifications_ok),
        borderColor: colours.value.ok,
        backgroundColor: colours.value.ok,
        fill: false,
        tension: 0,
        pointRadius: 0,
        pointHoverRadius: 3,
        borderWidth: 1.5,
      },
      {
        label: 'Failed',
        data: sorted.map((d) => d.verifications_fail),
        borderColor: colours.value.fail,
        backgroundColor: colours.value.fail,
        fill: false,
        tension: 0,
        pointRadius: 0,
        pointHoverRadius: 3,
        borderWidth: 1.5,
      },
    ],
  }
})

const chartOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  interaction: {
    intersect: false,
    mode: 'index' as const,
  },
  plugins: {
    legend: {
      position: 'bottom' as const,
      labels: { color: colours.value.muted, font: { family: 'Overpass', weight: 600 }, usePointStyle: true, pointStyle: 'circle' as const, boxWidth: 8, boxHeight: 8 },
    },
  },
  scales: {
    x: {
      grid: { display: false },
      border: { color: colours.value.rule },
      ticks: { color: colours.value.muted, font: { family: 'Overpass Mono', size: 11 }, maxRotation: 0, autoSkip: true, maxTicksLimit: 5 },
    },
    y: {
      beginAtZero: true,
      border: { display: false },
      grid: { color: colours.value.rule },
      ticks: { color: colours.value.muted, font: { family: 'Overpass Mono', size: 11 }, maxTicksLimit: 4, precision: 0 },
    },
  },
}))
</script>

<template>
  <div class="h-60">
    <Line :data="chartData" :options="chartOptions" />
  </div>
</template>
