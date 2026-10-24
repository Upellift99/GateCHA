<script setup lang="ts">
import { STATUS_LABEL, type KeyStatus } from '../lib/status'

// Colour backs the words, never replaces them.
// compact keeps only the count, for tables that have no room for the words.
defineProps<{ status: KeyStatus; failedToday?: number; compact?: boolean }>()
</script>

<template>
  <span
    class="inline-flex items-center gap-1.5 whitespace-nowrap rounded-[3px] px-2 py-0.5 text-xs font-semibold"
    :class="{
      'text-ok': status === 'ok',
      'bg-warn text-warn-ink': status === 'failing',
      'text-muted': status === 'idle',
      'bg-disabled text-disabled-ink': status === 'disabled',
    }"
  >
    <span
      v-if="status === 'ok' || status === 'failing'"
      class="h-[7px] w-[7px] rounded-full bg-current"
      aria-hidden="true"
    ></span>
    {{ STATUS_LABEL[status] }}
    <template v-if="status === 'failing' && failedToday">
      <span aria-hidden="true"> — </span>
      <span class="rounded-sm bg-panel px-1 font-mono text-fail tabular-nums">{{ failedToday.toLocaleString('en-US') }}</span>
      <span v-if="!compact" class="hidden sm:inline"> failures today</span>
    </template>
  </span>
</template>
