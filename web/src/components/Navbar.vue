<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const authStore = useAuthStore()

function logout() {
  authStore.logout()
  router.push('/login')
}
</script>

<template>
  <nav class="bg-white border-b border-gray-200">
    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
      <div class="flex justify-between h-16">
        <div class="flex items-center gap-8">
          <router-link to="/" class="text-xl font-bold text-gray-900">
            GateCHA
          </router-link>
          <div class="flex gap-4">
            <router-link
              to="/"
              class="text-sm font-medium text-gray-500 hover:text-gray-900 py-5 border-b-2 border-transparent"
              active-class="!text-indigo-600 !border-indigo-600"
            >
              Dashboard
            </router-link>
            <router-link
              to="/keys"
              class="text-sm font-medium text-gray-500 hover:text-gray-900 py-5 border-b-2 border-transparent"
              active-class="!text-indigo-600 !border-indigo-600"
            >
              API Keys
            </router-link>
            <router-link
              to="/settings"
              class="text-sm font-medium text-gray-500 hover:text-gray-900 py-5 border-b-2 border-transparent"
              active-class="!text-indigo-600 !border-indigo-600"
            >
              Settings
            </router-link>
          </div>
        </div>
        <div class="flex items-center">
          <button
            @click="logout"
            class="text-sm font-medium text-gray-500 hover:text-gray-900"
          >
            Logout
          </button>
        </div>
      </div>
    </div>
  </nav>
</template>
