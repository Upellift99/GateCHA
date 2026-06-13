<script setup lang="ts">
import { computed } from 'vue'
import type { CountryStat } from '../stores/stats'

const props = withDefaults(
  defineProps<{
    countries: CountryStat[]
    title?: string
    limit?: number
  }>(),
  { title: 'Traffic by country', limit: 10 },
)

const regionNames = (() => {
  try {
    return new Intl.DisplayNames(['en'], { type: 'region' })
  } catch {
    return null
  }
})()

/** Convert an ISO 3166-1 alpha-2 code to its flag emoji via regional indicators. */
function flagEmoji(code: string): string {
  if (!code || code.length !== 2 || !/^[A-Za-z]{2}$/.test(code)) return '🌐'
  const base = 0x1f1e6
  return String.fromCodePoint(
    ...[...code.toUpperCase()].map((c) => base + (c.codePointAt(0) ?? 0) - 65),
  )
}

function countryName(code: string): string {
  if (!code) return 'Unknown'
  try {
    return regionNames?.of(code.toUpperCase()) ?? code.toUpperCase()
  } catch {
    return code.toUpperCase()
  }
}

const rows = computed(() => props.countries.slice(0, props.limit))

const maxTotal = computed(() => rows.value.reduce((m, c) => Math.max(m, c.total), 0) || 1)

function failPct(c: CountryStat): number {
  return c.total === 0 ? 0 : Math.round((c.verifications_fail / c.total) * 100)
}
</script>

<template>
  <div class="bg-white shadow rounded-lg p-6">
    <div class="flex items-center justify-between mb-4">
      <h2 class="text-lg font-medium text-gray-900">{{ title }}</h2>
      <span class="text-xs text-gray-400">privacy-first · country only, no IP stored</span>
    </div>

    <ul v-if="rows.length" class="space-y-3">
      <li v-for="c in rows" :key="c.country || 'unknown'" class="flex items-center gap-3">
        <span class="text-xl leading-none w-7 text-center" :title="countryName(c.country)">{{
          flagEmoji(c.country)
        }}</span>
        <div class="flex-1 min-w-0">
          <div class="flex items-center justify-between text-sm">
            <span class="font-medium text-gray-700 truncate">{{ countryName(c.country) }}</span>
            <span class="text-gray-500 tabular-nums">{{ c.total.toLocaleString() }}</span>
          </div>
          <div class="mt-1 h-2 rounded-full bg-gray-100 overflow-hidden flex">
            <div
              class="h-full bg-green-500"
              :style="{ width: (c.verifications_ok / maxTotal) * 100 + '%' }"
              :title="`${c.verifications_ok.toLocaleString()} OK`"
            ></div>
            <div
              class="h-full bg-red-400"
              :style="{ width: (c.verifications_fail / maxTotal) * 100 + '%' }"
              :title="`${c.verifications_fail.toLocaleString()} failed`"
            ></div>
          </div>
        </div>
        <span
          class="text-xs tabular-nums w-12 text-right"
          :class="failPct(c) >= 50 ? 'text-red-600' : 'text-gray-400'"
          :title="'failure rate'"
          >{{ failPct(c) }}%</span
        >
      </li>
    </ul>

    <p v-else class="text-gray-500 text-center py-8 text-sm">No verifications recorded yet</p>
  </div>
</template>
