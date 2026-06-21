<script setup lang="ts">
import { ref, onBeforeUnmount } from 'vue'

const props = withDefaults(
  defineProps<{
    value: string
    label?: string
    variant?: 'inline' | 'overlay'
  }>(),
  { label: 'Copy', variant: 'inline' },
)

const copied = ref(false)
let timer: ReturnType<typeof setTimeout> | undefined

function copy() {
  navigator.clipboard.writeText(props.value)
  copied.value = true
  clearTimeout(timer)
  timer = setTimeout(() => { copied.value = false }, 2000)
}

onBeforeUnmount(() => clearTimeout(timer))
</script>

<template>
  <button
    type="button"
    @click="copy"
    :aria-label="copied ? 'Copied!' : label"
    :title="label"
    class="relative inline-flex shrink-0 items-center justify-center rounded transition-colors"
    :class="variant === 'overlay'
      ? 'h-7 w-7 bg-slate-700/80 text-slate-200 shadow-sm hover:bg-slate-600 hover:text-white'
      : 'h-6 w-6 text-slate-400 hover:text-teal-600'"
  >
    <!-- clipboard icon (idle) -->
    <svg v-if="!copied" class="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" aria-hidden="true">
      <path stroke-linecap="round" stroke-linejoin="round" d="M15.666 3.888A2.25 2.25 0 0 0 13.5 2.25h-3c-1.03 0-1.9.693-2.166 1.638m7.332 0c.055.194.084.4.084.612a.75.75 0 0 1-.75.75H9a.75.75 0 0 1-.75-.75c0-.212.03-.418.084-.612m7.332 0c.646.049 1.288.11 1.927.184 1.1.128 1.907 1.077 1.907 2.185V19.5a2.25 2.25 0 0 1-2.25 2.25H6.75A2.25 2.25 0 0 1 4.5 19.5V6.257c0-1.108.806-2.057 1.907-2.185a48.208 48.208 0 0 1 1.927-.184" />
    </svg>
    <!-- check icon (copied) -->
    <svg v-else class="h-4 w-4 text-green-500" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.4" aria-hidden="true">
      <path stroke-linecap="round" stroke-linejoin="round" d="m4.5 12.75 6 6 9-13.5" />
    </svg>
    <!-- "Copied!" tooltip — absolutely positioned so it never shifts layout -->
    <span
      v-if="copied"
      role="status"
      class="pointer-events-none absolute -top-7 left-1/2 z-10 -translate-x-1/2 whitespace-nowrap rounded bg-slate-900 px-1.5 py-0.5 text-xs font-medium text-white shadow"
      >Copied!</span
    >
  </button>
</template>
