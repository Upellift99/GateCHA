<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useStatsStore } from '../stores/stats'
import { useApiKeysStore } from '../stores/apikeys'
import AttentionBanner from '../components/AttentionBanner.vue'
import KeysToday from '../components/KeysToday.vue'
import VerificationsChart from '../components/VerificationsChart.vue'
import CountryTraffic from '../components/CountryTraffic.vue'
import { keyStatus, todayUTC, type FailingKey } from '../lib/status'

const statsStore = useStatsStore()
const keysStore = useApiKeysStore()

const today = computed(() => todayUTC())
const todayLabel = computed(() =>
  new Date(today.value + 'T00:00:00Z').toLocaleDateString('en-US', { weekday: 'long', month: 'short', day: 'numeric', timeZone: 'UTC' }),
)
const todayRows = computed(() => statsStore.keysDaily.filter((d) => d.date === today.value))
const sum = (f: (d: (typeof todayRows.value)[number]) => number) => todayRows.value.reduce((n, d) => n + f(d), 0)
const issuedToday = computed(() => sum((d) => d.challenges_issued))
const verifiedToday = computed(() => sum((d) => d.verifications_ok))
const failedToday = computed(() => sum((d) => d.verifications_fail))
const checkedToday = computed(() => verifiedToday.value + failedToday.value)
const passPct = computed(() => (checkedToday.value ? Math.round((verifiedToday.value / checkedToday.value) * 100) : 0))

const failing = computed<FailingKey[]>(() =>
  keysStore.keys
    .map((key) => ({ key, row: todayRows.value.find((d) => d.api_key_id === key.id) }))
    .filter(({ key, row }) => keyStatus(key.enabled, row) === 'failing')
    .map(({ key, row }) => ({ key, failed: row!.verifications_fail, verified: row!.verifications_ok + row!.verifications_fail }))
    .sort((a, b) => b.failed - a.failed),
)
const failedOnFailing = computed(() => failing.value.reduce((n, f) => n + f.failed, 0))

const thirtyDay = computed(() => {
  const o = statsStore.overview
  if (!o) return { issued: 0, pct: 0 }
  const days = o.daily ?? []
  const issued = days.reduce((n, d) => n + d.challenges_issued, 0)
  const ok = days.reduce((n, d) => n + d.verifications_ok, 0)
  const fail = days.reduce((n, d) => n + d.verifications_fail, 0)
  return { issued, pct: ok + fail ? Math.round((ok / (ok + fail)) * 100) : 0 }
})

const hisSuspectedPct = computed(() => {
  const o = statsStore.overview
  if (!o || o.total_his_observations === 0) return 0
  return Math.round((o.total_his_bot_suspected / o.total_his_observations) * 1000) / 10
})

onMounted(() => {
  statsStore.fetchOverview(30)
  statsStore.fetchKeysDaily(7)
  keysStore.fetchKeys()
})
</script>

<template>
  <div>
    <div class="mb-5 flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
      <div>
        <h1 class="text-[32px] font-bold leading-9 tracking-tight text-ink">Dashboard</h1>
        <p class="mt-1 text-sm text-muted">Today, {{ todayLabel }} (UTC)</p>
      </div>
      <router-link
        to="/keys/new"
        class="inline-flex min-h-10 items-center gap-2 self-start rounded-md px-4 text-sm font-semibold bg-action text-action-ink hover:bg-action-hover focus:outline-none focus-visible:ring-2 focus-visible:ring-focus focus-visible:ring-offset-2 sm:self-auto"
      >
        <svg class="h-3.5 w-3.5" viewBox="0 0 14 14" aria-hidden="true"><path d="M7 1v12M1 7h12" stroke="currentColor" stroke-width="2" /></svg>
        Create key
      </router-link>
    </div>

    <div v-if="statsStore.overview" class="space-y-6">
      <AttentionBanner :failing="failing" :active-keys="statsStore.overview.active_keys" :checked-today="checkedToday" />

      <div class="grid grid-cols-1 gap-6 lg:grid-cols-[minmax(0,1fr)_380px]">
        <div class="min-w-0 space-y-6">
          <!-- Today's funnel: puzzles handed out, then what the sites' backends checked -->
          <section class="rounded-md border border-rule bg-panel" aria-labelledby="today-title">
            <div class="flex flex-wrap items-baseline justify-between gap-x-4 gap-y-1 border-b border-rule px-5 py-3">
              <h2 id="today-title" class="text-[17px] font-semibold text-ink">Today's verifications</h2>
              <span class="text-[13px] text-muted">challenge issued → solved → checked by your backend</span>
            </div>
            <div class="p-5">
              <dl class="grid grid-cols-3 gap-4 sm:gap-5">
                <div>
                  <dt class="text-[11px] font-semibold uppercase tracking-[0.09em] text-muted">Challenges</dt>
                  <dd class="mt-1.5 font-mono text-[26px] font-semibold leading-8 sm:text-[32px] sm:leading-10 tracking-tight tabular-nums text-ink">{{ issuedToday.toLocaleString('en-US') }}</dd>
                  <dd class="mt-1 hidden text-xs text-muted sm:block">widgets that asked for a puzzle</dd>
                </div>
                <div>
                  <dt class="text-[11px] font-semibold uppercase tracking-[0.09em] text-muted">Verified</dt>
                  <dd class="mt-1.5 font-mono text-[26px] font-semibold leading-8 sm:text-[32px] sm:leading-10 tracking-tight tabular-nums text-ok">{{ verifiedToday.toLocaleString('en-US') }}</dd>
                  <dd class="mt-1 hidden text-xs text-muted sm:block">
                    <template v-if="checkedToday">{{ passPct }}% of verifications passed</template>
                    <template v-else>no verification yet today</template>
                  </dd>
                </div>
                <div>
                  <dt class="text-[11px] font-semibold uppercase tracking-[0.09em] text-muted">Failed</dt>
                  <dd class="mt-1.5 font-mono text-[26px] font-semibold leading-8 sm:text-[32px] sm:leading-10 tracking-tight tabular-nums" :class="failedToday ? 'text-fail' : 'text-ink'">{{ failedToday.toLocaleString('en-US') }}</dd>
                  <dd class="mt-1 hidden text-xs text-muted sm:block">
                    <template v-if="failing.length === 1 && failedOnFailing">{{ failedOnFailing.toLocaleString('en-US') }} on one key</template>
                    <template v-else-if="failing.length > 1">{{ failedOnFailing.toLocaleString('en-US') }} on {{ failing.length }} keys</template>
                    <template v-else>spread across keys</template>
                  </dd>
                </div>
              </dl>
              <div v-if="checkedToday" class="mt-5">
                <div class="flex h-3.5 overflow-hidden rounded-[3px] bg-sunken" aria-hidden="true">
                  <span class="block h-full bg-ok" :style="{ width: (verifiedToday / checkedToday) * 100 + '%' }"></span>
                  <span class="block h-full bg-fail" :style="{ width: (failedToday / checkedToday) * 100 + '%' }"></span>
                </div>
                <div class="mt-2 flex justify-between text-xs text-muted">
                  <span>Verified <span class="font-mono tabular-nums">{{ verifiedToday.toLocaleString('en-US') }}</span></span>
                  <span>Failed <span class="font-mono tabular-nums">{{ failedToday.toLocaleString('en-US') }}</span></span>
                </div>
              </div>
            </div>
          </section>

          <KeysToday :keys="keysStore.keys" :daily="statsStore.keysDaily" />
        </div>

        <aside class="min-w-0 space-y-6">
          <section class="rounded-md border border-rule bg-panel" aria-labelledby="month-title">
            <div class="flex flex-wrap items-baseline justify-between gap-x-3 gap-y-1 border-b border-rule px-5 py-3">
              <h2 id="month-title" class="text-[17px] font-semibold text-ink">Last 30 days</h2>
              <span class="text-[13px] text-muted"><span class="font-mono tabular-nums">{{ thirtyDay.issued.toLocaleString('en-US') }}</span> challenges · {{ thirtyDay.pct }}% verified</span>
            </div>
            <div class="px-5 pt-4 pb-4">
              <VerificationsChart :data="statsStore.overview.daily ?? []" />
            </div>
          </section>

          <CountryTraffic :countries="statsStore.overview.countries ?? []" title="Countries" />

          <!-- HIS: instance-wide totals. Not labelled "Monitor mode": blocking is
               a per-key switch, so this panel cannot claim the whole instance. -->
          <section class="rounded-md border border-rule bg-panel" aria-labelledby="his-title">
            <div class="flex items-baseline justify-between gap-3 border-b border-rule px-5 py-3">
              <h2 id="his-title" class="text-[17px] font-semibold text-ink">Bot-suspected</h2>
              <span class="font-mono font-semibold tabular-nums text-ink">{{ hisSuspectedPct }}%</span>
            </div>
            <p class="px-5 py-3 text-[13px] leading-5 text-muted">
              <span class="font-mono tabular-nums text-ink">{{ statsStore.overview.total_his_bot_suspected.toLocaleString('en-US') }}</span>
              of
              <span class="font-mono tabular-nums text-ink">{{ statsStore.overview.total_his_observations.toLocaleString('en-US') }}</span>
              scored forms looked automated (Human Interaction Signature). Blocking is off by default and switched on per key.
            </p>
          </section>
        </aside>
      </div>
    </div>

    <div v-else class="py-12 text-center text-muted">Loading…</div>
  </div>
</template>
