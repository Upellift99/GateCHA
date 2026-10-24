<script setup lang="ts">
import 'altcha'
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { startHISCollector, type HISCollector } from '../lib/his'
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

// Collect privacy-preserving interaction signals (HIS) for the duration of the
// login page. Sent with the login request; the server scores them in Monitor
// mode and never blocks on them.
let hisCollector: HISCollector | null = null

onMounted(async () => {
  hisCollector = startHISCollector()
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

onUnmounted(() => hisCollector?.stop())

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
    await authStore.login(
      username.value,
      password.value,
      altchaPayload.value || undefined,
      hisCollector?.signals(),
    )
    router.push('/')
  } catch {
    error.value = 'Invalid credentials'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="flex justify-center pt-[10dvh] pb-16">
    <div class="w-full max-w-sm">
      <div class="mb-5">
        <h1 class="text-2xl font-bold tracking-tight text-ink">Sign in</h1>
        <p class="text-sm text-muted">to your GateCHA dashboard</p>
      </div>
      <div class="rounded-md border border-rule bg-panel p-6">

        <form @submit.prevent="handleLogin" class="space-y-3">
          <div v-if="error" role="alert" class="bg-fail-soft text-fail px-4 py-3 rounded text-sm">
            {{ error }}
          </div>

          <div>
            <label for="username" class="block text-sm font-medium text-ink-2 mb-1">Username</label>
            <input
              id="username"
              v-model="username"
              type="text"
              required
              class="w-full px-3 py-2 border border-rule-strong rounded-md focus:outline-none focus:ring-2 focus:ring-focus focus:border-focus"
            />
          </div>

          <div>
            <label for="password" class="block text-sm font-medium text-ink-2 mb-1">Password</label>
            <input
              id="password"
              v-model="password"
              type="password"
              required
              class="w-full px-3 py-2 border border-rule-strong rounded-md focus:outline-none focus:ring-2 focus:ring-focus focus:border-focus"
            />
          </div>

          <div v-if="captchaRequired && challengeUrl">
            <altcha-widget
              :challenge="challengeUrl"
              @statechange="onAltchaStateChange"
              style="--altcha-max-width: 100%;"
            ></altcha-widget>
          </div>

          <button
            type="submit"
            :disabled="loading || !canSubmit"
            class="w-full py-2 px-4 bg-action text-action-ink text-sm font-semibold rounded-md hover:bg-action-hover focus:outline-none focus-visible:ring-2 focus-visible:ring-focus disabled:opacity-50"
          >
            {{ loading ? 'Signing in...' : 'Sign in' }}
          </button>
        </form>
      </div>
    </div>
  </div>
</template>
