<script setup lang="ts">
import { onMounted } from 'vue'
import { storeToRefs } from 'pinia'
import { useAuthStore } from '../stores/auth'

const authStore = useAuthStore()
const { isAuthenticated, version } = storeToRefs(authStore)

onMounted(() => {
  if (isAuthenticated.value) authStore.fetchVersion()
})
</script>

<template>
  <footer class="border-t border-slate-200 bg-white">
    <div
      class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4 flex flex-col sm:flex-row items-center justify-between gap-2 text-sm text-slate-400"
    >
      <p>
        <span class="font-semibold text-slate-500">Gate<span class="text-brand-600">CHA</span></span>
        <span v-if="isAuthenticated && version" class="ml-2 tabular-nums">{{ version }}</span>
      </p>
      <p class="inline-flex items-center gap-1">
        Made with
        <svg class="h-4 w-4 text-red-500" viewBox="0 0 20 20" fill="currentColor" aria-label="love" role="img">
          <path d="M9.653 16.915l-.005-.003-.019-.01a20.759 20.759 0 0 1-1.162-.682 22.045 22.045 0 0 1-2.582-1.9C4.045 12.733 2 10.352 2 7.5a4.5 4.5 0 0 1 8-2.828A4.5 4.5 0 0 1 18 7.5c0 2.852-2.044 5.233-3.885 6.82a22.049 22.049 0 0 1-3.744 2.582l-.019.01-.005.003h-.002a.739.739 0 0 1-.69.001l-.002-.001Z" />
        </svg>
        in Sceaux, FR
      </p>
      <a
        href="https://gatecha.org"
        target="_blank"
        rel="noopener noreferrer"
        class="inline-flex items-center gap-1 hover:text-brand-600 transition-colors"
      >
        gatecha.org
        <svg class="h-3.5 w-3.5" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
          <path d="M11 3a1 1 0 1 0 0 2h2.586l-6.293 6.293a1 1 0 1 0 1.414 1.414L15 6.414V9a1 1 0 1 0 2 0V4a1 1 0 0 0-1-1h-5Z" />
          <path d="M5 5a2 2 0 0 0-2 2v8a2 2 0 0 0 2 2h8a2 2 0 0 0 2-2v-3a1 1 0 1 0-2 0v3H5V7h3a1 1 0 0 0 0-2H5Z" />
        </svg>
      </a>
    </div>
  </footer>
</template>
