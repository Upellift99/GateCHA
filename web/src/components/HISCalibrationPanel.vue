<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useStatsStore, type HISCalibration } from '../stores/stats'

const props = defineProps<{ keyId?: number }>()

const stats = useStatsStore()
const cal = ref<HISCalibration | null>(null)
const loading = ref(true)

onMounted(async () => {
  try {
    cal.value = await stats.fetchHISCalibration(props.keyId)
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
    <p class="text-sm text-slate-500 mb-4">
      Where stored samples fall on the automation-probability scale. Bars at or past the
      suspect threshold are counted as bot-suspected. Use this to pick an enforcement threshold.
    </p>

    <p v-if="loading" class="text-slate-400 text-center py-8 text-sm">Loading…</p>

    <template v-else-if="cal && cal.samples > 0">
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

    <p v-else class="text-slate-400 text-center py-8 text-sm">
      No samples yet. Enable <span class="font-medium text-slate-500">HIS sampling</span> on this key to collect calibration data.
    </p>
  </div>
</template>
