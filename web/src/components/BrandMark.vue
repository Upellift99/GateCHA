<script setup lang="ts">
// The GateCHA mark: a "G" whose bar turns into the check of a passed
// verification. Same drawing as web/public/favicon.svg; keep them in step.
withDefaults(defineProps<{ size?: number }>(), { size: 30 })
</script>

<template>
  <svg :width="size" :height="size" viewBox="0 0 32 32" aria-hidden="true" class="block shrink-0">
    <rect width="32" height="32" rx="7" fill="#10285A" />
    <rect x="0.75" y="0.75" width="30.5" height="30.5" rx="6.25" fill="none" stroke="#2A4479" stroke-width="1.5" />
    <path d="M23.4 10.6A8.6 8.6 0 1 0 24.6 16" fill="none" stroke="#fff" stroke-width="3.4" stroke-linecap="round" />
    <path d="M14.2 16.2l2.6 2.6 6.4-7.2" fill="none" stroke="#FFCD00" stroke-width="3" stroke-linecap="round" stroke-linejoin="round" />
  </svg>
</template>
