<script setup lang="ts">
import { computed } from 'vue'
import type { APIKey } from '../stores/apikeys'
import type { DailyStat } from '../stores/stats'
import LineBullet from './LineBullet.vue'
import { lastDays, lineVars } from '../lib/lines'

// Challenges per day, one line per key, drawn in its line colour. The selected
// line (hovered on the board, or pinned from the legend) keeps full strength;
// every other line steps back. Quiet days are real zeros, not gaps.
const props = defineProps<{ keys: APIKey[]; daily: DailyStat[]; selected: number | null; pinned: number | null; days?: number }>()
const emit = defineEmits<{ pin: [id: number | null] }>()

const DAYS = computed(() => props.days ?? 30)
const W = 900
const H = 180
const PAD = 8

const dates = computed(() => lastDays(DAYS.value))

const series = computed(() => {
  const byKey = new Map<number, Map<string, number>>()
  for (const d of props.daily) {
    if (!byKey.has(d.api_key_id)) byKey.set(d.api_key_id, new Map())
    byKey.get(d.api_key_id)!.set(d.date, d.challenges_issued)
  }
  return props.keys.map((key) => {
    const m = byKey.get(key.id)
    const values = dates.value.map((d) => m?.get(d) ?? 0)
    return { key, values, total: values.reduce((a, b) => a + b, 0) }
  })
})

const max = computed(() => Math.max(1, ...series.value.flatMap((s) => s.values)))

function points(values: number[]): string {
  const n = values.length - 1 || 1
  return values
    .map((v, i) => `${((i / n) * W).toFixed(1)},${(H - PAD - (v / max.value) * (H - 2 * PAD)).toFixed(1)}`)
    .join(' ')
}

const active = computed(() => props.selected ?? props.pinned)

// Draw the active line last so it sits on top of the others.
const ordered = computed(() =>
  [...series.value].sort((a, b) => Number(a.key.id === active.value) - Number(b.key.id === active.value)),
)

function shortDate(iso: string): string {
  return new Date(iso + 'T00:00:00Z').toLocaleDateString('en-US', { month: 'short', day: 'numeric', timeZone: 'UTC' })
}

const summary = computed(() =>
  series.value.map((s) => `line ${s.key.id} ${s.key.name || s.key.key_id}: ${s.total} challenges`).join('; '),
)
</script>

<template>
  <section class="rounded-md border border-rule bg-panel" aria-labelledby="lines-chart-title">
    <div class="flex flex-wrap items-center justify-between gap-3 border-b border-rule px-5 py-3">
      <h2 id="lines-chart-title" class="text-[17px] font-extrabold text-ink">
        Challenges, {{ DAYS }} days
        <span v-if="active !== null" class="ml-2 text-[13px] font-medium text-muted">Line {{ active }} selected</span>
      </h2>
      <div class="flex flex-wrap items-center gap-1" role="group" aria-label="Highlight a line">
        <button
          v-for="s in series"
          :key="s.key.id"
          type="button"
          class="inline-flex min-h-7 items-center gap-1.5 rounded px-1.5 text-xs font-semibold transition-opacity focus:outline-none focus-visible:ring-2 focus-visible:ring-focus"
          :class="active === null || active === s.key.id ? 'opacity-100 text-ink' : 'opacity-70 text-muted'"
          :aria-pressed="pinned === s.key.id"
          :aria-label="`Line ${s.key.id}, ${s.key.name || s.key.key_id}`"
          :title="s.key.name || s.key.key_id"
          @click="emit('pin', pinned === s.key.id ? null : s.key.id)"
        >
          <LineBullet :id="s.key.id" size="sm" />
          <span v-if="active === s.key.id" class="hidden max-w-40 truncate sm:inline">{{ s.key.name || s.key.key_id }}</span>
        </button>
      </div>
    </div>

    <div class="px-5 pt-4 pb-3">
      <div class="flex gap-3">
        <div class="flex w-10 shrink-0 flex-col justify-between py-1 text-right font-mono text-[11px] tabular-nums text-muted" aria-hidden="true">
          <span>{{ max.toLocaleString('en-US') }}</span>
          <span>0</span>
        </div>
        <svg :viewBox="`0 0 ${W} ${H}`" preserveAspectRatio="none" class="block h-44 w-full" role="img" :aria-label="`Challenges per day over ${DAYS} days: ${summary}`">
          <g stroke="var(--color-rule-soft)" stroke-width="1" vector-effect="non-scaling-stroke">
            <line :x1="0" :x2="W" :y1="PAD" :y2="PAD" vector-effect="non-scaling-stroke" />
            <line :x1="0" :x2="W" :y1="H / 2" :y2="H / 2" vector-effect="non-scaling-stroke" />
            <line :x1="0" :x2="W" :y1="H - PAD" :y2="H - PAD" stroke="var(--color-rule)" vector-effect="non-scaling-stroke" />
          </g>
          <polyline
            v-for="s in ordered"
            :key="s.key.id"
            fill="none"
            :stroke="s.key.enabled ? lineVars(s.key.id).stroke : 'var(--color-rule-strong)'"
            :stroke-width="active === s.key.id ? 2.75 : 1.5"
            :stroke-opacity="active === null || active === s.key.id ? 1 : 0.22"
            :stroke-dasharray="s.key.enabled ? undefined : '4 4'"
            stroke-linejoin="round"
            vector-effect="non-scaling-stroke"
            :points="points(s.values)"
          />
        </svg>
      </div>
      <div class="ml-[52px] mt-1 flex justify-between font-mono text-[11px] text-muted" aria-hidden="true">
        <span>{{ shortDate(dates[0]!) }}</span>
        <span>{{ shortDate(dates[Math.floor(dates.length / 2)]!) }}</span>
        <span>Today</span>
      </div>
    </div>
  </section>
</template>
