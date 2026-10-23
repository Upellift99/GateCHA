<script setup lang="ts">
// One figure of a hairline-divided KPI strip: label, the number in tabular
// mono, a line of context. No icon tile, no shadow: the number leads.
withDefaults(
  defineProps<{
    label: string
    value: number
    tone?: 'default' | 'fail'
    hint?: string
    suffix?: string
  }>(),
  { tone: 'default' },
)
</script>

<template>
  <div class="min-w-0 bg-panel px-5 py-4">
    <p class="text-[11px] font-bold uppercase leading-4 tracking-[0.09em] text-muted">{{ label }}</p>
    <p
      class="mt-1.5 font-mono text-[32px] font-semibold leading-10 tracking-tight tabular-nums"
      :class="tone === 'fail' && value > 0 ? 'text-fail' : 'text-ink'"
    >
      {{ value.toLocaleString('en-US') }}<span v-if="suffix" class="text-xl text-muted">{{ suffix }}</span>
    </p>
    <p v-if="hint || $slots.default" class="mt-1 flex min-h-5 items-center gap-2 text-xs text-muted">
      <slot>{{ hint }}</slot>
    </p>
  </div>
</template>
