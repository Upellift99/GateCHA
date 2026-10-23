<script setup lang="ts">
import { ref, onBeforeUnmount } from 'vue'

const props = withDefaults(
  defineProps<{
    value: string
    label?: string
    variant?: 'inline' | 'overlay' | 'button'
  }>(),
  { label: 'Copy', variant: 'inline' },
)

const copied = ref(false)
const failed = ref(false)
let timer: ReturnType<typeof setTimeout> | undefined

// Shows "Copied!" straight away, then takes it back if the browser refuses
// (no clipboard API outside secure contexts, or a denied permission), so the
// button never claims a copy that did not happen.
function copy() {
  failed.value = false
  copied.value = true
  clearTimeout(timer)
  timer = setTimeout(() => { copied.value = false; failed.value = false }, 2000)
  let pending: Promise<void> | undefined
  try {
    pending = navigator.clipboard.writeText(props.value)
  } catch {
    pending = Promise.reject(new Error('Clipboard API unavailable'))
  }
  Promise.resolve(pending).catch(() => {
    copied.value = false
    failed.value = true
  })
}

onBeforeUnmount(() => clearTimeout(timer))
</script>

<template>
  <!-- The overlay is placed by its parent (absolute, in a code block corner);
       every other variant anchors its own "Copied!" tooltip. -->
  <button
    type="button"
    @click="copy"
    :aria-label="copied ? 'Copied!' : failed ? 'Copy failed' : label"
    :title="label"
    class="inline-flex shrink-0 items-center justify-center rounded transition-colors"
    :class="{
      relative: variant !== 'overlay',
      'h-7 w-7 bg-code text-code-ink ring-1 ring-white/15 hover:bg-band': variant === 'overlay',
      'h-6 w-6 text-muted hover:text-link-hover': variant === 'inline',
      'gap-1.5 rounded-md bg-action px-3 py-2 text-sm font-medium text-action-ink hover:bg-action-hover focus:outline-none focus:ring-2 focus:ring-focus focus:ring-offset-2':
        variant === 'button',
    }"
  >
    <!-- clipboard icon (idle) -->
    <svg v-if="!copied" class="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" aria-hidden="true">
      <path stroke-linecap="round" stroke-linejoin="round" d="M15.666 3.888A2.25 2.25 0 0 0 13.5 2.25h-3c-1.03 0-1.9.693-2.166 1.638m7.332 0c.055.194.084.4.084.612a.75.75 0 0 1-.75.75H9a.75.75 0 0 1-.75-.75c0-.212.03-.418.084-.612m7.332 0c.646.049 1.288.11 1.927.184 1.1.128 1.907 1.077 1.907 2.185V19.5a2.25 2.25 0 0 1-2.25 2.25H6.75A2.25 2.25 0 0 1 4.5 19.5V6.257c0-1.108.806-2.057 1.907-2.185a48.208 48.208 0 0 1 1.927-.184" />
    </svg>
    <!-- check icon (copied) -->
    <svg v-else class="h-4 w-4" :class="variant === 'button' ? 'text-action-ink' : 'text-ok'" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.4" aria-hidden="true">
      <path stroke-linecap="round" stroke-linejoin="round" d="m4.5 12.75 6 6 9-13.5" />
    </svg>
    <!-- The button variant says what it does in visible text -->
    <span v-if="variant === 'button'">{{ copied ? 'Copied!' : failed ? 'Copy failed' : 'Copy' }}</span>
    <!-- "Copied!" tooltip — <output> carries an implicit status role + aria-live;
         absolutely positioned so it never shifts layout. The button variant
         already shows it, so there it is only announced. -->
    <output
      v-if="copied || failed"
      :class="variant === 'button' ? 'sr-only' : ''"
      class="pointer-events-none absolute -top-7 left-1/2 z-10 -translate-x-1/2 whitespace-nowrap rounded bg-code px-1.5 py-0.5 text-xs font-medium text-action-ink"
      >{{ failed ? 'Copy failed, select the text and copy it' : 'Copied!' }}</output
    >
  </button>
</template>
