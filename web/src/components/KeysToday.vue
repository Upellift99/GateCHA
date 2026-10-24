<script setup lang="ts">
import { computed } from 'vue'
import type { APIKey } from '../stores/apikeys'
import type { DailyStat } from '../stores/stats'
import { keyStatus, lastDays, todayUTC, type KeyStatus } from '../lib/status'
import { domainList } from '../lib/format'

// Every key with today's counts, grouped by what the admin has to do about it:
// failing keys first, then the ones working, the quiet ones, the disabled ones.
const props = defineProps<{ keys: APIKey[]; daily: DailyStat[] }>()

const SPARK_DAYS = 7
const W = 88
const H = 22

const GROUPS: { status: KeyStatus; title: string }[] = [
  { status: 'failing', title: 'Needs attention' },
  { status: 'ok', title: 'Working' },
  { status: 'idle', title: 'No traffic today' },
  { status: 'disabled', title: 'Disabled' },
]

const byKey = computed(() => {
  const m = new Map<number, Map<string, DailyStat>>()
  for (const d of props.daily) {
    if (!m.has(d.api_key_id)) m.set(d.api_key_id, new Map())
    m.get(d.api_key_id)!.set(d.date, d)
  }
  return m
})

const rows = computed(() => {
  const days = lastDays(SPARK_DAYS)
  const today = todayUTC()
  return props.keys.map((key) => {
    const series = byKey.value.get(key.id)
    const todayRow = series?.get(today)
    const counts = days.map((d) => series?.get(d)?.challenges_issued ?? 0)
    const max = Math.max(1, ...counts)
    return {
      key,
      status: keyStatus(key.enabled, todayRow),
      domains: domainList(key.domain),
      challengesToday: todayRow?.challenges_issued ?? 0,
      failedToday: todayRow?.verifications_fail ?? 0,
      points: counts
        .map((c, i) => `${((i / (SPARK_DAYS - 1)) * W).toFixed(1)},${(H - 2 - (c / max) * (H - 4)).toFixed(1)}`)
        .join(' '),
    }
  })
})

const groups = computed(() =>
  GROUPS.map((g) => ({
    ...g,
    rows: rows.value
      .filter((r) => r.status === g.status)
      .sort((a, b) => b.challengesToday - a.challengesToday || (a.key.name || a.key.key_id).localeCompare(b.key.name || b.key.key_id)),
  })).filter((g) => g.rows.length),
)
</script>

<template>
  <section class="rounded-md border border-rule bg-panel" aria-labelledby="keys-today-title">
    <div class="flex items-baseline justify-between gap-4 border-b border-rule px-5 py-3">
      <h2 id="keys-today-title" class="text-[17px] font-semibold text-ink">Keys</h2>
      <router-link to="/keys" class="inline-flex min-h-6 items-center text-sm font-semibold text-link hover:text-link-hover">Manage keys →</router-link>
    </div>

    <table v-if="rows.length" class="w-full border-collapse">
      <thead>
        <tr class="text-left text-[11px] font-semibold uppercase tracking-[0.09em] text-muted">
          <th scope="col" class="py-2 pl-5 pr-3 font-semibold">Key</th>
          <th scope="col" class="hidden w-28 py-2 px-3 font-semibold sm:table-cell">7 days</th>
          <th scope="col" class="w-20 py-2 px-3 text-right font-semibold">Today</th>
          <th scope="col" class="w-20 py-2 pl-3 pr-5 text-right font-semibold">Failed</th>
        </tr>
      </thead>
      <tbody v-for="g in groups" :key="g.status">
        <tr>
          <th
            scope="rowgroup"
            colspan="4"
            class="border-t border-rule bg-sunken py-2 pl-5 pr-5 text-left text-[11px] font-semibold uppercase tracking-[0.09em] text-muted"
          >
            {{ g.title }}<span class="font-mono"> · {{ g.rows.length }}</span>
          </th>
        </tr>
        <tr
          v-for="(row, i) in g.rows"
          :key="row.key.id"
          :class="[i ? 'border-t border-rule-soft' : '', g.status === 'failing' ? 'bg-warn-soft' : '']"
        >
          <td class="max-w-0 w-full py-2.5 pl-5 pr-3">
            <router-link
              :to="`/keys/${row.key.id}`"
              class="block truncate text-sm font-semibold text-ink hover:text-link"
              :class="{ 'text-muted': g.status === 'disabled' }"
              :title="row.key.name || row.key.key_id"
            >{{ row.key.name || row.key.key_id }}</router-link>
            <div class="truncate text-xs leading-4 text-muted" :title="row.domains.join('\n') || undefined">
              <template v-if="row.domains.length">{{ row.domains[0] }}<span v-if="row.domains.length > 1"> +{{ row.domains.length - 1 }} more</span></template>
              <template v-else>Any domain</template>
            </div>
          </td>
          <td class="hidden py-2.5 px-3 sm:table-cell">
            <svg :width="W" :height="H" :viewBox="`0 0 ${W} ${H}`" role="img" :aria-label="`7-day challenges for ${row.key.name || row.key.key_id}`">
              <polyline
                fill="none"
                :stroke="g.status === 'failing' ? 'var(--color-fail)' : g.status === 'disabled' ? 'var(--color-rule-strong)' : 'var(--color-link)'"
                stroke-width="1.5"
                stroke-linejoin="round"
                :stroke-dasharray="g.status === 'disabled' ? '3 3' : undefined"
                :points="row.points"
              />
            </svg>
          </td>
          <td class="py-2.5 px-3 text-right font-mono text-sm tabular-nums" :class="row.challengesToday ? 'text-ink' : 'text-muted'">
            {{ row.challengesToday.toLocaleString('en-US') }}
          </td>
          <td
            class="py-2.5 pl-3 pr-5 text-right font-mono text-sm tabular-nums"
            :class="g.status === 'failing' ? 'font-semibold text-fail' : 'text-muted'"
          >
            {{ g.status === 'disabled' ? '—' : row.failedToday.toLocaleString('en-US') }}
          </td>
        </tr>
      </tbody>
    </table>

    <div v-else class="px-5 py-10 text-center">
      <p class="text-sm text-muted">No keys yet. Create one for each site that shows the widget.</p>
      <router-link to="/keys/new" class="mt-3 inline-flex min-h-8 items-center text-sm font-semibold text-link hover:text-link-hover">Create your first key →</router-link>
    </div>
  </section>
</template>
