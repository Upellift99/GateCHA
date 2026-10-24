<script setup lang="ts">
import { computed } from 'vue'
import type { CountryStat } from '../stores/stats'

const props = withDefaults(
  defineProps<{
    countries: CountryStat[]
    title?: string
    limit?: number
    /** Match the p-6 cards of the key detail page instead of the dashboard panels. */
    roomy?: boolean
  }>(),
  { title: 'Traffic by country', limit: 10 },
)

const inset = computed(() => (props.roomy ? 'px-6' : 'px-5'))

const regionNames = (() => {
  try {
    return new Intl.DisplayNames(['en'], { type: 'region' })
  } catch {
    return null
  }
})()

// An ISO code tag instead of a flag emoji: regional-indicator flags render as
// bare letter pairs on Windows, and a code reads the same everywhere.
function countryCode(code: string): string {
  return code && /^[A-Za-z]{2}$/.test(code) ? code.toUpperCase() : '··'
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
  <section class="rounded-md border border-rule bg-panel" aria-labelledby="country-title">
    <div class="border-b border-rule py-3" :class="inset">
      <h2 id="country-title" class="text-[17px] font-semibold text-ink">{{ title }}</h2>
      <p class="text-xs text-muted">Country only, resolved offline; no IP stored</p>
    </div>

    <ul v-if="rows.length" class="divide-y divide-rule-soft">
      <li v-for="c in rows" :key="c.country || 'unknown'" class="grid grid-cols-[auto_1fr] items-center gap-x-3 py-2.5" :class="inset">
        <span
          class="row-span-2 inline-flex h-6 w-8 items-center justify-center rounded-[3px] border border-rule font-mono text-[11px] font-semibold text-ink-2"
          :title="countryName(c.country)"
        >{{ countryCode(c.country) }}</span>
        <div class="flex items-baseline justify-between gap-3 text-sm">
          <span class="min-w-0 truncate font-semibold" :class="c.country ? 'text-ink' : 'text-muted'">{{ countryName(c.country) }}</span>
          <span class="shrink-0 font-mono text-xs tabular-nums text-muted">
            <span class="text-sm text-ink">{{ c.total.toLocaleString('en-US') }}</span>
            &ensp;<span :class="failPct(c) >= 50 ? 'font-semibold text-fail' : ''">{{ failPct(c) }}% failed</span>
          </span>
        </div>
        <div class="mt-1 flex h-1.5 overflow-hidden rounded-full bg-sunken" aria-hidden="true">
          <div class="h-full bg-ok" :style="{ width: (c.verifications_ok / maxTotal) * 100 + '%' }"></div>
          <div class="h-full bg-fail" :style="{ width: (c.verifications_fail / maxTotal) * 100 + '%' }"></div>
        </div>
      </li>
    </ul>

    <p v-else class="py-8 text-center text-sm text-muted" :class="inset">No verifications recorded yet</p>
  </section>
</template>
