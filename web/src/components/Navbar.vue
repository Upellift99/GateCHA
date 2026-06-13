<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const authStore = useAuthStore()

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
  <nav class="sticky top-0 z-40 bg-white/80 backdrop-blur border-b border-slate-200">
    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
      <div class="flex justify-between h-16">
        <div class="flex items-center gap-8">
          <router-link to="/" class="flex items-center gap-2" aria-label="GateCHA home">
            <span class="inline-flex h-8 w-8 items-center justify-center rounded-lg bg-brand-600 shadow-sm">
              <svg viewBox="0 0 32 32" fill="none" class="h-5 w-5" aria-hidden="true">
                <path
                  d="M16 4L6 9v7c0 7 4.5 11.5 10 14 5.5-2.5 10-7 10-14V9L16 4z"
                  fill="none"
                  stroke="#fff"
                  stroke-width="2.2"
                  stroke-linejoin="round"
                />
                <path
                  d="M11 16l3.5 3.5L21 12"
                  stroke="#fff"
                  stroke-width="2.6"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  fill="none"
                />
              </svg>
            </span>
            <span class="text-xl font-bold tracking-tight">
              <span class="text-slate-900">Gate</span><span class="text-brand-600">CHA</span>
            </span>
          </router-link>
          <div class="hidden sm:flex gap-1">
            <router-link
              v-for="link in links"
              :key="link.to"
              :to="link.to"
              class="rounded-md px-3 py-2 text-sm font-medium text-slate-500 transition-colors hover:bg-slate-100 hover:text-slate-900"
              active-class="!bg-brand-50 !text-brand-700"
            >
              {{ link.label }}
            </router-link>
          </div>
        </div>
        <div class="flex items-center">
          <button
            @click="logout"
            class="inline-flex items-center gap-1.5 rounded-md px-3 py-2 text-sm font-medium text-slate-500 transition-colors hover:bg-slate-100 hover:text-slate-900"
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

      <!-- Mobile nav -->
      <div class="flex sm:hidden gap-1 pb-2 -mt-1">
        <router-link
          v-for="link in links"
          :key="link.to"
          :to="link.to"
          class="rounded-md px-3 py-1.5 text-sm font-medium text-slate-500 hover:bg-slate-100 hover:text-slate-900"
          active-class="!bg-brand-50 !text-brand-700"
        >
          {{ link.label }}
        </router-link>
      </div>
    </div>
  </nav>
</template>
