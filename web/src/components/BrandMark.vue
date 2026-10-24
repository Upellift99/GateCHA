<script setup lang="ts">
// The GateCHA mark: a shield with a check, white on the GateCHA teal. Same
// drawing as web/public/favicon.svg, gatecha.org and the WordPress.org icon;
// keep them in step.
withDefaults(defineProps<{ size?: number }>(), { size: 28 })
</script>

<template>
  <svg :width="size" :height="size" viewBox="0 0 32 32" aria-hidden="true" class="block shrink-0">
    <rect width="32" height="32" rx="7" fill="#0F766E" />
    <path d="M16 5.5 7.5 9.2v6.3c0 5.6 3.6 9.4 8.5 11.2 4.9-1.8 8.5-5.6 8.5-11.2V9.2L16 5.5z" fill="none" stroke="#fff" stroke-width="2.6" stroke-linejoin="round" />
    <path d="M12 16.2l2.9 2.9 5.3-6" fill="none" stroke="#fff" stroke-width="2.8" stroke-linecap="round" stroke-linejoin="round" />
  </svg>
</template>
