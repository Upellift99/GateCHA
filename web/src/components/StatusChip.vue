<script setup lang="ts">
import { STATUS_LABEL, type LineStatus } from '../lib/lines'

// Transit-board vocabulary. Colour backs the words, never replaces them.
// compact keeps only the count, for tables that have no room for the words.
defineProps<{ status: LineStatus; failedToday?: number; compact?: boolean }>()
</script>

<template>
  <span
    class="inline-flex items-center gap-1.5 whitespace-nowrap rounded-[3px] px-2 py-0.5 text-xs font-bold"
    :class="{
      'text-ok': status === 'normal',
      'bg-warn text-warn-ink': status === 'disturbed',
      'text-muted': status === 'quiet',
      'bg-closed text-closed-ink': status === 'closed',
    }"
  >
    <span
      v-if="status === 'normal' || status === 'disturbed'"
      class="h-[7px] w-[7px] rounded-full bg-current"
      aria-hidden="true"
    ></span>
    {{ STATUS_LABEL[status] }}
    <template v-if="status === 'disturbed' && failedToday">
      <span aria-hidden="true"> — </span>
      <span class="rounded-sm bg-panel px-1 font-mono text-fail tabular-nums">{{ failedToday.toLocaleString('en-US') }}</span>
      <span v-if="!compact" class="hidden sm:inline"> failed today</span>
    </template>
  </span>
</template>
