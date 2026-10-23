<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import Navbar from './components/Navbar.vue'
import Footer from './components/Footer.vue'

const route = useRoute()
// Before the first navigation resolves, route.name is undefined: wait for it
// rather than flashing the signed-in band on the login page.
const signedInLayout = computed(() => route.matched.length > 0 && route.name !== 'login')
</script>

<template>
  <div class="min-h-screen flex flex-col bg-ground text-ink">
    <Navbar v-if="signedInLayout" />
    <!-- Signed out: the same signage band, carrying only the mark -->
    <header v-else-if="route.matched.length" class="bg-band text-band-ink border-b border-band-rule">
      <div class="max-w-7xl mx-auto flex h-14 items-center gap-2.5 px-4 text-lg font-extrabold tracking-tight sm:px-6 lg:px-8">
        <span class="inline-flex h-7 w-7 items-center justify-center rounded-full bg-band-ink text-band text-base leading-none" aria-hidden="true">G</span>
        GateCHA
      </div>
    </header>
    <main class="flex-1 w-full max-w-7xl mx-auto px-4 py-8 sm:px-6 lg:px-8">
      <router-view />
    </main>
    <Footer :show-version="route.name !== 'login'" />
  </div>
</template>
