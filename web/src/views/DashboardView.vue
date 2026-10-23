<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useStatsStore } from '../stores/stats'
import { useApiKeysStore } from '../stores/apikeys'
import StatsSummaryCard from '../components/StatsSummaryCard.vue'
import CountryTraffic from '../components/CountryTraffic.vue'
import NetworkStatus from '../components/NetworkStatus.vue'
import LinesChart from '../components/LinesChart.vue'
import LineBullet from '../components/LineBullet.vue'
import { lineStatus, todayUTC } from '../lib/lines'

const statsStore = useStatsStore()
const keysStore = useApiKeysStore()

// The board hovers a line; the chart legend pins one. Hover wins while it lasts.
const hovered = ref<number | null>(null)
const pinned = ref<number | null>(null)

const today = computed(() => todayUTC())
const todayRows = computed(() => statsStore.keysDaily.filter((d) => d.date === today.value))
const challengesToday = computed(() => todayRows.value.reduce((n, d) => n + d.challenges_issued, 0))
const failedToday = computed(() => todayRows.value.reduce((n, d) => n + d.verifications_fail, 0))

const statuses = computed(() =>
  keysStore.keys.map((k) => ({ id: k.id, status: lineStatus(k.enabled, todayRows.value.find((d) => d.api_key_id === k.id)) })),
)
const disturbed = computed(() => statuses.value.filter((s) => s.status === 'disturbed'))
const closedCount = computed(() => statuses.value.filter((s) => s.status === 'closed').length)

const successRate = computed(() => {
  const o = statsStore.overview
  if (!o) return 0
  const total = o.total_verifications_ok + o.total_verifications_fail
  return total === 0 ? 0 : Math.round((o.total_verifications_ok / total) * 100)
})

// The chart opens on the problem: the first disturbed line is pinned until the
// admin picks another line (or clears it) from the legend.
let pinTouched = false
watch(disturbed, (lines) => {
  if (!pinTouched && pinned.value === null && lines.length) pinned.value = lines[0]!.id
})
function pin(id: number | null) {
  pinTouched = true
  pinned.value = id
}

// The chart and its legend read in line order.
const linesById = computed(() => [...keysStore.keys].sort((a, b) => a.id - b.id))

const hisSuspectedPct = computed(() => {
  const o = statsStore.overview
  if (!o || o.total_his_observations === 0) return 0
  return Math.round((o.total_his_bot_suspected / o.total_his_observations) * 1000) / 10
})

onMounted(() => {
  statsStore.fetchOverview(30)
  statsStore.fetchKeysDaily(30)
  keysStore.fetchKeys()
})
</script>

<template>
  <div>
    <div class="mb-5 flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
      <div>
        <h1 class="text-[32px] font-extrabold leading-9 tracking-tight text-ink">Dashboard</h1>
        <p class="mt-1 text-sm text-muted">Every API key is a line of your network. Totals cover all time; today is the current UTC day.</p>
      </div>
      <router-link
        to="/keys/new"
        class="inline-flex min-h-10 items-center gap-2 self-start rounded px-4 text-sm font-bold bg-action text-action-ink hover:bg-action-hover focus:outline-none focus-visible:ring-2 focus-visible:ring-focus focus-visible:ring-offset-2 sm:self-auto"
      >
        <svg class="h-3.5 w-3.5" viewBox="0 0 14 14" aria-hidden="true"><path d="M7 1v12M1 7h12" stroke="currentColor" stroke-width="2" /></svg>
        Create key
      </router-link>
    </div>

    <div v-if="statsStore.overview" class="space-y-6">
      <!-- KPI strip: four figures divided by hairlines -->
      <section class="grid grid-cols-2 gap-px overflow-hidden rounded-md border border-rule bg-rule lg:grid-cols-4" aria-label="Totals">
        <StatsSummaryCard label="Challenges · all time" :value="statsStore.overview.total_challenges" :hint="`${challengesToday.toLocaleString('en-US')} today`" />
        <StatsSummaryCard label="Verified · all time" :value="statsStore.overview.total_verifications_ok">
          <span class="font-bold text-ok">{{ successRate }}% success</span>
        </StatsSummaryCard>
        <StatsSummaryCard label="Failed · all time" :value="statsStore.overview.total_verifications_fail" tone="fail">
          <template v-if="failedToday">
            <span class="font-bold text-fail">{{ failedToday.toLocaleString('en-US') }} today</span>
            <template v-if="disturbed.length">
              on
              <span class="inline-flex gap-0.5"><LineBullet v-for="d in disturbed" :key="d.id" :id="d.id" size="sm" /></span>
            </template>
          </template>
          <template v-else>None today</template>
        </StatsSummaryCard>
        <StatsSummaryCard
          label="Active lines"
          :value="statsStore.overview.active_keys"
          :hint="`${disturbed.length} disturbed · ${closedCount} closed`"
        />
      </section>

      <div class="grid grid-cols-1 gap-6 lg:grid-cols-3">
        <div class="space-y-6 lg:col-span-2 min-w-0">
          <NetworkStatus :keys="keysStore.keys" :daily="statsStore.keysDaily" :selected="hovered" @select="hovered = $event" />
          <LinesChart
            v-if="keysStore.keys.length"
            :keys="linesById"
            :daily="statsStore.keysDaily"
            :selected="hovered"
            :pinned="pinned"
            @pin="pin"
          />
        </div>

        <aside class="space-y-6 min-w-0">
          <CountryTraffic :countries="statsStore.overview.countries ?? []" />

          <!-- HIS: instance-wide totals. Not labelled "Monitor mode": blocking is
               a per-key switch, so this panel cannot claim the whole instance. -->
          <section class="rounded-md border border-rule bg-panel" aria-labelledby="his-title">
            <div class="flex items-center justify-between gap-3 border-b border-rule px-5 py-3">
              <h2 id="his-title" class="text-[17px] font-extrabold text-ink">Human Interaction Signature</h2>
              <span class="whitespace-nowrap rounded-[3px] border border-rule px-1.5 pt-0.5 text-[11px] font-bold uppercase tracking-[0.08em] text-muted">All keys</span>
            </div>
            <div class="grid grid-cols-3 gap-px bg-rule">
              <StatsSummaryCard label="Observed" :value="statsStore.overview.total_his_observations" />
              <StatsSummaryCard label="Suspected" :value="statsStore.overview.total_his_bot_suspected" tone="fail" />
              <StatsSummaryCard label="Share" :value="hisSuspectedPct" suffix="%" />
            </div>
            <p class="border-t border-rule-soft px-5 py-3 text-xs leading-[18px] text-muted">
              Scores how forms are filled, from aggregates only. Blocking is off by default and switched on per key, once its calibration histogram shows where to set the threshold.
            </p>
          </section>
        </aside>
      </div>
    </div>

    <div v-else class="py-12 text-center text-muted">Loading…</div>
  </div>
</template>
