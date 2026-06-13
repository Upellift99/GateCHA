<script setup lang="ts">
withDefaults(
  defineProps<{
    label: string
    value: number
    color?: 'brand' | 'blue' | 'green' | 'red' | 'purple' | 'amber'
    hint?: string
    suffix?: string
  }>(),
  { color: 'brand' },
)

const accent: Record<string, string> = {
  brand: 'bg-brand-50 text-brand-600',
  blue: 'bg-blue-50 text-blue-600',
  green: 'bg-green-50 text-green-600',
  red: 'bg-red-50 text-red-600',
  purple: 'bg-purple-50 text-purple-600',
  amber: 'bg-amber-50 text-amber-600',
}
</script>

<template>
  <div
    class="rounded-xl border border-slate-200 bg-white p-5 shadow-sm transition-shadow hover:shadow-md"
  >
    <div class="flex items-start justify-between">
      <p class="text-xs font-semibold uppercase tracking-wide text-slate-500">{{ label }}</p>
      <span
        class="inline-flex h-8 w-8 items-center justify-center rounded-lg"
        :class="accent[color]"
      >
        <slot name="icon">
          <span class="h-2 w-2 rounded-full bg-current"></span>
        </slot>
      </span>
    </div>
    <p class="mt-3 text-3xl font-bold tracking-tight text-slate-900 tabular-nums">
      {{ value.toLocaleString() }}<span v-if="suffix" class="text-xl text-slate-400">{{ suffix }}</span>
    </p>
    <p v-if="hint" class="mt-1 text-xs text-slate-400">{{ hint }}</p>
  </div>
</template>
