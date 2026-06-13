<script setup lang="ts">
import { useRoute } from 'vue-router'
import Navbar from './components/Navbar.vue'
import Footer from './components/Footer.vue'

const route = useRoute()
</script>

<template>
  <div class="min-h-screen flex flex-col bg-slate-50 text-slate-900">
    <Navbar v-if="route.name !== 'login'" />
    <main class="flex-1 w-full max-w-7xl mx-auto px-4 py-8 sm:px-6 lg:px-8">
      <router-view />
    </main>
    <Footer v-if="route.name !== 'login'" />
  </div>
</template>
