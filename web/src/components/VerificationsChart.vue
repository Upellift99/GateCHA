<script setup lang="ts">
import { computed } from 'vue'
import type { DailyStat } from '../stores/stats'
import { lastDays } from '../lib/status'

// Daily verifications as stacked bars (verified, then failed on top) with the
// challenges issued drawn as a line. Days without traffic are zero, not gaps.
const props = withDefaults(defineProps<{ data: DailyStat[]; days?: number }>(), { days: 30 })

const W = 380
const H = 120

const series = computed(() => {
  const byDate = new Map(props.data.map((d) => [d.date, d]))
  return lastDays(props.days).map((date) => {
    const d = byDate.get(date)
    return { date, issued: d?.challenges_issued ?? 0, ok: d?.verifications_ok ?? 0, fail: d?.verifications_fail ?? 0 }
  })
})

const max = computed(() => Math.max(1, ...series.value.map((d) => Math.max(d.issued, d.ok + d.fail))))
const y = (v: number) => H - (v / max.value) * (H - 6)
const bw = computed(() => W / series.value.length)

const label = (date: string) =>
  new Date(date + 'T00:00:00Z').toLocaleDateString('en-US', { month: 'short', day: 'numeric', timeZone: 'UTC' })

const line = computed(() =>
  series.value.map((d, i) => `${(i * bw.value + bw.value / 2).toFixed(1)},${y(d.issued).toFixed(1)}`).join(' '),
)

const summary = computed(() => {
  const t = series.value.reduce((a, d) => ({ issued: a.issued + d.issued, ok: a.ok + d.ok, fail: a.fail + d.fail }), { issued: 0, ok: 0, fail: 0 })
  return `${t.issued.toLocaleString('en-US')} challenges, ${t.ok.toLocaleString('en-US')} verified, ${t.fail.toLocaleString('en-US')} failed`
})
</script>

<template>
  <figure class="m-0">
    <svg :viewBox="`0 0 ${W} ${H}`" preserveAspectRatio="none" class="block h-32 w-full" role="img" :aria-label="`Last ${days} days: ${summary}`">
      <line v-for="f in [0.5, 1]" :key="f" x1="0" :x2="W" :y1="y(max * f)" :y2="y(max * f)" stroke="var(--color-rule-soft)" vector-effect="non-scaling-stroke" />
      <g v-for="(d, i) in series" :key="d.date">
        <title>{{ label(d.date) }}: {{ d.issued.toLocaleString('en-US') }} challenges, {{ d.ok.toLocaleString('en-US') }} verified, {{ d.fail.toLocaleString('en-US') }} failed</title>
        <rect :x="i * bw + 1" :y="y(d.ok)" :width="bw - 2" :height="H - y(d.ok)" fill="var(--color-ok-edge)" />
        <rect :x="i * bw + 1" :y="y(d.ok + d.fail)" :width="bw - 2" :height="H - y(d.fail)" fill="var(--color-fail)" />
      </g>
      <polyline :points="line" fill="none" stroke="var(--color-ink)" stroke-width="1.75" stroke-linejoin="round" vector-effect="non-scaling-stroke" />
    </svg>
    <div class="mt-1 flex justify-between font-mono text-[11px] text-muted" aria-hidden="true">
      <span>{{ label(series[0]!.date) }}</span><span>Today</span>
    </div>
    <figcaption class="mt-2 flex flex-wrap gap-x-4 gap-y-1 text-xs font-semibold text-muted">
      <span class="inline-flex items-center gap-1.5"><span class="h-2.5 w-2.5 rounded-sm bg-ok-edge" aria-hidden="true"></span>Verified</span>
      <span class="inline-flex items-center gap-1.5"><span class="h-2.5 w-2.5 rounded-sm bg-fail" aria-hidden="true"></span>Failed</span>
      <span class="inline-flex items-center gap-1.5"><span class="h-0.5 w-3 bg-ink" aria-hidden="true"></span>Challenges</span>
    </figcaption>
  </figure>
</template>
