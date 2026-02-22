<script setup lang="ts">
import 'altcha'
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import axios from 'axios'

const router = useRouter()
const authStore = useAuthStore()

const username = ref('')
const password = ref('')
const error = ref('')
const loading = ref(false)

const captchaRequired = ref(false)
const challengeUrl = ref('')
const altchaPayload = ref('')
const altchaVerified = ref(false)

onMounted(async () => {
  try {
    const { data } = await axios.get('/api/public/login-config')
    captchaRequired.value = data.captcha_required
    if (data.challenge_url) {
      challengeUrl.value = data.challenge_url
    }
  } catch {
    captchaRequired.value = false
  }
})

function onAltchaStateChange(event: Event) {
  const detail = (event as CustomEvent).detail
  if (detail?.state === 'verified') {
    altchaPayload.value = detail.payload ?? ''
    altchaVerified.value = true
  } else {
    altchaPayload.value = ''
    altchaVerified.value = false
  }
}

const canSubmit = computed(() => {
  if (captchaRequired.value && !altchaVerified.value) return false
  return true
})

async function handleLogin() {
  error.value = ''
  loading.value = true
  try {
    await authStore.login(username.value, password.value, altchaPayload.value || undefined)
    router.push('/')
  } catch {
    error.value = 'Invalid credentials'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="min-h-screen flex items-center justify-center bg-gray-50">
    <div class="w-full max-w-sm">
      <div class="bg-white shadow rounded-lg p-8">
        <h1 class="text-2xl font-bold text-center text-gray-900 mb-6">GateCHA</h1>
        <p class="text-sm text-center text-gray-500 mb-6">Sign in to your dashboard</p>

        <form @submit.prevent="handleLogin" class="space-y-4">
          <div v-if="error" class="bg-red-50 text-red-700 px-4 py-3 rounded text-sm">
            {{ error }}
          </div>

          <div>
            <label for="username" class="block text-sm font-medium text-gray-700 mb-1">Username</label>
            <input
              id="username"
              v-model="username"
              type="text"
              required
              class="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500"
            />
          </div>

          <div>
            <label for="password" class="block text-sm font-medium text-gray-700 mb-1">Password</label>
            <input
              id="password"
              v-model="password"
              type="password"
              required
              class="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500"
            />
          </div>

          <div v-if="captchaRequired && challengeUrl">
            <altcha-widget
              :challengeurl="challengeUrl"
              @statechange="onAltchaStateChange"
              style="--altcha-max-width: 100%;"
            ></altcha-widget>
          </div>

          <button
            type="submit"
            :disabled="loading || !canSubmit"
            class="w-full py-2 px-4 bg-indigo-600 text-white font-medium rounded-md hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-indigo-500 disabled:opacity-50"
          >
            {{ loading ? 'Signing in...' : 'Sign in' }}
          </button>
        </form>
      </div>
    </div>
  </div>
</template>
