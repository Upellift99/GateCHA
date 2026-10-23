<script setup lang="ts">
import { computed } from 'vue'
import { lineVars } from '../lib/lines'

const props = withDefaults(defineProps<{ id: number; size?: 'sm' | 'md' | 'lg' }>(), { size: 'md' })

const style = computed(() => {
  const v = lineVars(props.id)
  return { backgroundColor: v.fill, color: v.ink, boxShadow: 'inset 0 0 0 1px var(--color-line-ring)' }
})

// Long ids shrink the type rather than the circle, so every bullet keeps its size.
const digits = computed(() => String(props.id).length)
</script>

<template>
  <span
    class="inline-flex shrink-0 items-center justify-center rounded-full font-extrabold leading-none tabular-nums select-none"
    :class="{
      'h-[18px] w-[18px] text-[10px]': size === 'sm',
      'h-[26px] w-[26px]': size === 'md',
      'h-9 w-9': size === 'lg',
      'text-[14px]': size === 'md' && digits < 3,
      'text-[11px]': size === 'md' && digits >= 3,
      'text-lg': size === 'lg' && digits < 3,
      'text-sm': size === 'lg' && digits >= 3,
    }"
    :style="style"
    aria-hidden="true"
  >{{ id }}</span>
</template>
