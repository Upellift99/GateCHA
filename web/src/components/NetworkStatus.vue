<script setup lang="ts">
import { computed } from 'vue'
import type { APIKey } from '../stores/apikeys'
import type { DailyStat } from '../stores/stats'
import LineBullet from './LineBullet.vue'
import StatusChip from './StatusChip.vue'
import { lastDays, lineStatus, lineVars, STATUS_RANK, todayUTC } from '../lib/lines'
import { domainList } from '../lib/format'

// The traffic board: one row per line (API key), today's counts and status,
// a 7-day spark. Disturbed lines rise to the top; hovering or focusing a row
// selects its line, which the chart below uses to dim every other line.
const props = defineProps<{ keys: APIKey[]; daily: DailyStat[]; selected: number | null }>()
const emit = defineEmits<{ select: [id: number | null] }>()

const SPARK_DAYS = 7
const W = 88
const H = 20

const days = computed(() => lastDays(SPARK_DAYS))
const today = computed(() => todayUTC())

const byKey = computed(() => {
  const m = new Map<number, Map<string, DailyStat>>()
  for (const d of props.daily) {
    if (!m.has(d.api_key_id)) m.set(d.api_key_id, new Map())
    m.get(d.api_key_id)!.set(d.date, d)
  }
  return m
})

const rows = computed(() =>
  props.keys
    .map((key) => {
      const series = byKey.value.get(key.id)
      const todayRow = series?.get(today.value)
      const status = lineStatus(key.enabled, todayRow)
      const counts = days.value.map((d) => series?.get(d)?.challenges_issued ?? 0)
      const max = Math.max(1, ...counts)
      const points = counts
        .map((c, i) => `${((i / (SPARK_DAYS - 1)) * W).toFixed(1)},${(H - 2 - (c / max) * (H - 4)).toFixed(1)}`)
        .join(' ')
      return {
        key,
        status,
        domains: domainList(key.domain),
        challengesToday: todayRow?.challenges_issued ?? 0,
        failedToday: todayRow?.verifications_fail ?? 0,
        points,
        stroke: lineVars(key.id).stroke,
      }
    })
    .sort((a, b) => STATUS_RANK[a.status] - STATUS_RANK[b.status] || (a.key.name || a.key.key_id).localeCompare(b.key.name || b.key.key_id)),
)
</script>

<template>
  <section class="rounded-md border border-rule bg-panel" aria-labelledby="network-status-title">
    <div class="flex flex-wrap items-baseline justify-between gap-x-4 gap-y-1 border-b border-rule px-5 py-3">
      <h2 id="network-status-title" class="text-[17px] font-extrabold text-ink">
        Network status
        <span class="ml-2 text-[13px] font-medium text-muted">Today (UTC)<span class="hidden md:inline">&ensp;·&ensp;7-day trend</span></span>
      </h2>
      <router-link to="/keys" class="inline-flex min-h-6 items-center text-sm font-semibold text-link hover:text-link-hover">All keys →</router-link>
    </div>

    <div v-if="rows.length" class="overflow-x-auto">
      <table class="w-full border-collapse">
        <thead>
          <tr class="text-left text-[11px] font-bold uppercase tracking-[0.09em] text-muted">
            <th scope="col" class="w-14 py-2 pl-5 pr-3 font-bold">Line</th>
            <th scope="col" class="py-2 px-3 font-bold">Key</th>
            <th scope="col" class="hidden md:table-cell w-28 py-2 px-3 font-bold">7 days</th>
            <th scope="col" class="hidden sm:table-cell w-28 py-2 px-3 text-right font-bold">Challenges</th>
            <th scope="col" class="hidden sm:table-cell w-20 py-2 px-3 text-right font-bold">Failed</th>
            <th scope="col" class="hidden sm:table-cell py-2 pl-3 pr-5 font-bold">Status</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="row in rows"
            :key="row.key.id"
            class="border-t border-rule-soft transition-opacity"
            :class="{
              'bg-warn-soft !border-warn-edge [&+tr]:!border-warn-edge': row.status === 'disturbed',
              'opacity-45': selected !== null && selected !== row.key.id,
            }"
            @mouseenter="emit('select', row.key.id)"
            @mouseleave="emit('select', null)"
            @focusin="emit('select', row.key.id)"
            @focusout="emit('select', null)"
          >
            <td class="py-2 pl-5 pr-3">
              <LineBullet :id="row.key.id" />
            </td>
            <td class="py-2 px-3 max-w-0 w-full">
              <router-link
                :to="`/keys/${row.key.id}`"
                class="block truncate font-bold leading-[18px] text-ink hover:text-link"
                :class="{ 'text-muted': row.status === 'closed' }"
                :title="row.key.name || row.key.key_id"
              >{{ row.key.name || row.key.key_id }}</router-link>
              <div class="truncate text-xs leading-4 text-muted" :title="row.domains.join('\n') || undefined">
                <template v-if="row.domains.length">{{ row.domains[0] }}<span v-if="row.domains.length > 1"> +{{ row.domains.length - 1 }} more</span></template>
                <template v-else>Any domain</template>
              </div>
              <!-- Phones: the status moves under the name instead of its own column -->
              <StatusChip :status="row.status" :failed-today="row.failedToday" class="mt-1 sm:hidden" />
            </td>
            <td class="hidden md:table-cell py-2 px-3">
              <svg :width="88" :height="20" viewBox="0 0 88 20" role="img" :aria-label="`7-day challenges for line ${row.key.id}`">
                <polyline
                  fill="none"
                  :stroke="row.status === 'closed' ? 'var(--color-rule-strong)' : row.stroke"
                  stroke-width="1.5"
                  stroke-linejoin="round"
                  :stroke-dasharray="row.status === 'closed' ? '3 3' : undefined"
                  :points="row.points"
                />
              </svg>
            </td>
            <td class="hidden sm:table-cell py-2 px-3 text-right font-mono text-sm tabular-nums" :class="row.challengesToday ? 'text-ink' : 'text-muted'">
              {{ row.challengesToday.toLocaleString('en-US') }}
            </td>
            <td
              class="hidden sm:table-cell py-2 px-3 text-right font-mono text-sm tabular-nums"
              :class="row.failedToday ? 'font-bold text-fail' : 'text-muted'"
            >
              {{ row.failedToday.toLocaleString('en-US') }}
            </td>
            <td class="hidden sm:table-cell py-2 pl-3 pr-5">
              <StatusChip :status="row.status" :failed-today="row.failedToday" />
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div v-else class="px-5 py-10 text-center">
      <p class="text-sm text-muted">No lines yet. Each API key you create appears here as a line of the network.</p>
      <router-link to="/keys/new" class="mt-3 inline-flex min-h-8 items-center text-sm font-semibold text-link hover:text-link-hover">Create your first key →</router-link>
    </div>
  </section>
</template>
