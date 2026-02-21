<script setup lang="ts">
import { useRoute } from 'vue-router'
import Navbar from './components/Navbar.vue'

const route = useRoute()
</script>

<template>
  <div class="min-h-screen bg-gray-50">
    <Navbar v-if="route.name !== 'login'" />
    <main class="max-w-7xl mx-auto px-4 py-8 sm:px-6 lg:px-8">
      <router-view />
    </main>
  </div>
</template>
