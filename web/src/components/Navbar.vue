<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { storeToRefs } from 'pinia'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const authStore = useAuthStore()
const { version } = storeToRefs(authStore)

// /version needs a token: calling it signed out answers 401, and the API client
// then reloads /login, so the call is guarded. fetchVersion() is idempotent.
onMounted(() => {
  if (authStore.isAuthenticated) authStore.fetchVersion()
})

const links = [
  { to: '/', label: 'Dashboard' },
  { to: '/keys', label: 'API Keys' },
  { to: '/settings', label: 'Settings' },
]

function logout() {
  authStore.logout()
  router.push('/login')
}
</script>

<template>
  <!-- The signage band: navy in both themes, the navigation lives on it. -->
  <nav class="sticky top-0 z-40 bg-band text-band-ink border-b border-band-rule" aria-label="Main">
    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
      <div class="flex h-14 items-stretch gap-6 sm:gap-10">
        <router-link to="/" class="flex items-center gap-2.5 font-extrabold text-lg tracking-tight" aria-label="GateCHA home">
          <span class="inline-flex h-7 w-7 items-center justify-center rounded-full bg-band-ink text-band font-extrabold text-base leading-none" aria-hidden="true">G</span>
          GateCHA
        </router-link>
        <div class="hidden sm:flex items-stretch gap-1">
          <router-link
            v-for="link in links"
            :key="link.to"
            :to="link.to"
            class="flex items-center border-b-[3px] border-transparent px-4 pt-[3px] text-sm font-semibold text-band-muted transition-colors hover:text-band-ink"
            active-class="!border-band-ink !text-band-ink"
          >
            {{ link.label }}
          </router-link>
        </div>
        <div class="ml-auto flex items-center gap-4">
          <span v-if="version" class="hidden sm:inline font-mono text-xs text-band-muted">{{ version }}</span>
          <button
            @click="logout"
            class="inline-flex min-h-8 items-center gap-1.5 rounded border border-band-rule px-3 text-sm font-semibold text-band-ink transition-colors hover:bg-white/10 focus:outline-none focus-visible:ring-2 focus-visible:ring-band-ink"
          >
            <svg class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
              <path
                fill-rule="evenodd"
                d="M3 4.25A2.25 2.25 0 0 1 5.25 2h5.5A2.25 2.25 0 0 1 13 4.25v2a.75.75 0 0 1-1.5 0v-2a.75.75 0 0 0-.75-.75h-5.5a.75.75 0 0 0-.75.75v11.5c0 .414.336.75.75.75h5.5a.75.75 0 0 0 .75-.75v-2a.75.75 0 0 1 1.5 0v2A2.25 2.25 0 0 1 10.75 18h-5.5A2.25 2.25 0 0 1 3 15.75V4.25Z"
                clip-rule="evenodd"
              />
              <path
                fill-rule="evenodd"
                d="M19 10a.75.75 0 0 0-.75-.75H8.704l1.048-.943a.75.75 0 1 0-1.004-1.114l-2.5 2.25a.75.75 0 0 0 0 1.114l2.5 2.25a.75.75 0 1 0 1.004-1.114l-1.048-.943h9.546A.75.75 0 0 0 19 10Z"
                clip-rule="evenodd"
              />
            </svg>
            Logout
          </button>
        </div>
      </div>

      <!-- Mobile nav: same tabs, second row of the band -->
      <div class="flex sm:hidden items-stretch gap-1 -mt-1">
        <router-link
          v-for="link in links"
          :key="link.to"
          :to="link.to"
          class="flex min-h-11 items-center border-b-[3px] border-transparent px-3 text-sm font-semibold text-band-muted hover:text-band-ink"
          active-class="!border-band-ink !text-band-ink"
        >
          {{ link.label }}
        </router-link>
      </div>
    </div>
  </nav>
</template>
