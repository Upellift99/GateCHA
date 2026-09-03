<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useStatsStore, type HISCalibration } from '../stores/stats'

const props = defineProps<{
  keyId?: number
  /**
   * Scored observations for this key over the same window. Counted for every
   * request carrying `his_signals` whether or not sampling is on, so it is what
   * separates "the collector never reaches /verify" from "sampling was switched
   * on after those requests came in".
   */
  observations?: number
}>()

/** Shared by the query and the copy below, so the two cannot drift apart. */
const DAYS = 30

const stats = useStatsStore()
const cal = ref<HISCalibration | null>(null)
const loading = ref(true)
const failed = ref(false)

onMounted(async () => {
  try {
    cal.value = await stats.fetchHISCalibration(props.keyId, DAYS)
  } catch {
    // Kept distinct from "no samples". A failed request used to fall through to
    // the empty state, which told the reader to enable sampling on a key that
    // already had it enabled: a diagnosis that was both wrong and unactionable.
    failed.value = true
  } finally {
    loading.value = false
  }
})

const maxCount = computed(() => (cal.value?.score_histogram ?? []).reduce((m, b) => Math.max(m, b.count), 0) || 1)

const suspectedPct = computed(() => {
  if (!cal.value || cal.value.samples === 0) return 0
  return Math.round((cal.value.suspected / cal.value.samples) * 100)
})

function barClass(loValue: number, threshold: number): string {
  return loValue + 1e-9 >= threshold ? 'bg-red-400' : 'bg-brand-500'
}
</script>

<template>
  <div class="rounded-xl border border-slate-200 bg-white p-6 shadow-sm">
    <div class="flex items-center justify-between mb-1 gap-2">
      <h2 class="text-lg font-semibold text-slate-900">HIS calibration</h2>
      <span class="text-xs font-medium px-2.5 py-1 rounded-full bg-slate-100 text-slate-600">score distribution</span>
    </div>
    <p v-if="loading" class="text-slate-400 text-center py-8 text-sm">Loading…</p>

    <p v-else-if="failed" class="text-slate-500 text-sm py-6">
      Could not load the calibration data. This says nothing about how many samples exist.
      Reload the page; if it keeps failing, your session may have expired, or the request
      failed server side, in which case the instance log records the reason.
    </p>

    <template v-else-if="cal && cal.samples > 0">
      <!-- Describes the histogram, so it belongs with the histogram: printed
           above the empty states it explained bars that were not there. -->
      <p class="text-sm text-slate-500 mb-4">
        Where stored samples fall on the automation-probability scale. Bars at or past the
        suspect threshold are counted as bot-suspected. Use this to pick an enforcement threshold.
      </p>

      <div class="grid grid-cols-3 gap-4 mb-5">
        <div>
          <p class="text-xs font-semibold uppercase tracking-wide text-slate-500">Samples</p>
          <p class="mt-1 text-2xl font-bold text-slate-900 tabular-nums">{{ cal.samples.toLocaleString() }}</p>
        </div>
        <div>
          <p class="text-xs font-semibold uppercase tracking-wide text-slate-500">Bot-suspected</p>
          <p class="mt-1 text-2xl font-bold text-slate-900 tabular-nums">{{ suspectedPct }}<span class="text-base text-slate-400">%</span></p>
        </div>
        <div>
          <p class="text-xs font-semibold uppercase tracking-wide text-slate-500">No motion</p>
          <p class="mt-1 text-2xl font-bold text-slate-900 tabular-nums">{{ Math.round(cal.no_motion_pct) }}<span class="text-base text-slate-400">%</span></p>
        </div>
      </div>

      <!-- Histogram -->
      <div class="flex items-end gap-1 h-32">
        <div
          v-for="b in cal.score_histogram"
          :key="b.lo"
          class="flex-1 flex flex-col justify-end h-full"
          :title="`[${b.lo.toFixed(1)}, ${b.hi.toFixed(1)}): ${b.count}`"
        >
          <span class="text-[10px] text-center text-slate-400 tabular-nums">{{ b.count || '' }}</span>
          <div
            class="rounded-t"
            :class="barClass(b.lo, cal.threshold)"
            :style="{ height: (b.count / maxCount) * 100 + '%' }"
          ></div>
        </div>
      </div>
      <div class="flex justify-between mt-1 text-[10px] text-slate-400 tabular-nums">
        <span>0.0</span>
        <span class="text-red-500">▲ threshold {{ cal.threshold.toFixed(2) }}</span>
        <span>1.0</span>
      </div>
    </template>

    <div v-else class="text-sm text-slate-500 py-6 space-y-3">
      <p>
        <span class="font-medium text-slate-700">HIS sampling is on for this key</span>, and no
        sample has been stored in the last {{ DAYS }} days.
      </p>

      <template v-if="observations && observations > 0">
        <p>
          Signals are reaching <code class="text-xs bg-slate-100 rounded px-1 py-0.5">/verify</code>:
          {{ observations.toLocaleString() }} scored in the same window. Samples are only stored from
          the moment sampling is switched on, so the histogram fills from now on.
        </p>
      </template>

      <template v-else>
        <p>
          No <code class="text-xs bg-slate-100 rounded px-1 py-0.5">his_signals</code> reached
          <code class="text-xs bg-slate-100 rounded px-1 py-0.5">/verify</code> at all in that window,
          so the collector is not getting through. The usual causes:
        </p>
        <ul class="list-disc pl-5 space-y-1">
          <li>the ALTCHA widget sits outside the <code class="text-xs bg-slate-100 rounded px-1 py-0.5">&lt;form&gt;</code>, so the collector never recognises the form;</li>
          <li>the form is posted with <code class="text-xs bg-slate-100 rounded px-1 py-0.5">fetch</code> and fires no native submit event. Read <code class="text-xs bg-slate-100 rounded px-1 py-0.5">globalThis.gatechaHIS.signals()</code> yourself instead;</li>
          <li>your backend does not forward the hidden <code class="text-xs bg-slate-100 rounded px-1 py-0.5">gatecha_his_signals</code> field into the <code class="text-xs bg-slate-100 rounded px-1 py-0.5">/verify</code> body. It does not travel on its own.</li>
        </ul>
        <p>
          The browser console also reports it when the collector finds no form to attach to. Full
          integration steps are in
          <a
            class="text-brand-600 hover:text-brand-700 underline"
            href="https://github.com/Upellift99/GateCHA#4-optional-collect-interaction-signals-his"
            target="_blank"
            rel="noopener"
            >README step 4</a
          >.
        </p>
      </template>
    </div>
  </div>
</template>
