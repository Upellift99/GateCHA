<script setup lang="ts">
import { ref, onMounted } from 'vue'
import api from '../lib/api'
import { useSettingsStore } from '../stores/settings'

const currentPassword = ref('')
const newPassword = ref('')
const confirmPassword = ref('')
const error = ref('')
const success = ref('')
const loading = ref(false)

async function handleSubmit() {
  error.value = ''
  success.value = ''

  if (newPassword.value !== confirmPassword.value) {
    error.value = 'Passwords do not match'
    return
  }
  if (newPassword.value.length < 8) {
    error.value = 'Password must be at least 8 characters'
    return
  }

  loading.value = true
  try {
    await api.post('/change-password', {
      current_password: currentPassword.value,
      new_password: newPassword.value,
    })
    success.value = 'Password changed successfully'
    currentPassword.value = ''
    newPassword.value = ''
    confirmPassword.value = ''
  } catch {
    error.value = 'Failed to change password. Check your current password.'
  } finally {
    loading.value = false
  }
}

const settingsStore = useSettingsStore()
const settingsError = ref('')

onMounted(() => {
  settingsStore.fetchSettings()
})

async function toggleCaptcha(event: Event) {
  const checked = (event.target as HTMLInputElement).checked
  settingsError.value = ''
  try {
    await settingsStore.updateSettings({ login_captcha_enabled: checked })
  } catch {
    settingsError.value = 'Failed to update setting.'
    await settingsStore.fetchSettings()
  }
}
</script>

<template>
  <div class="max-w-xl space-y-8">
    <h1 class="text-2xl font-bold text-gray-900 mb-6">Settings</h1>

    <form @submit.prevent="handleSubmit" class="bg-white shadow rounded-lg p-6 space-y-4">
      <h2 class="text-lg font-medium text-gray-900">Change Password</h2>

      <div v-if="error" class="bg-red-50 text-red-700 px-4 py-3 rounded text-sm">{{ error }}</div>
      <div v-if="success" class="bg-green-50 text-green-700 px-4 py-3 rounded text-sm">{{ success }}</div>

      <div>
        <label for="currentPassword" class="block text-sm font-medium text-gray-700 mb-1">Current Password</label>
        <input
          id="currentPassword"
          v-model="currentPassword"
          type="password"
          required
          class="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
        />
      </div>

      <div>
        <label for="newPassword" class="block text-sm font-medium text-gray-700 mb-1">New Password</label>
        <input
          id="newPassword"
          v-model="newPassword"
          type="password"
          required
          class="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
        />
      </div>

      <div>
        <label for="confirmPassword" class="block text-sm font-medium text-gray-700 mb-1">Confirm New Password</label>
        <input
          id="confirmPassword"
          v-model="confirmPassword"
          type="password"
          required
          class="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
        />
      </div>

      <button
        type="submit"
        :disabled="loading"
        class="py-2 px-4 bg-indigo-600 text-white font-medium rounded-md hover:bg-indigo-700 disabled:opacity-50"
      >
        {{ loading ? 'Saving...' : 'Change Password' }}
      </button>
    </form>

    <div class="bg-white shadow rounded-lg p-6 space-y-4">
      <h2 class="text-lg font-medium text-gray-900">Security</h2>

      <div v-if="settingsError" class="bg-red-50 text-red-700 px-4 py-3 rounded text-sm">
        {{ settingsError }}
      </div>

      <div class="flex items-center justify-between">
        <div>
          <p class="text-sm font-medium text-gray-700">Login CAPTCHA</p>
          <p class="text-xs text-gray-500 mt-0.5">
            Require an ALTCHA proof-of-work challenge before signing in.
          </p>
        </div>
        <label for="loginCaptchaToggle" class="relative inline-flex items-center cursor-pointer">
          <input
            id="loginCaptchaToggle"
            type="checkbox"
            class="sr-only peer"
            :checked="settingsStore.settings.login_captcha_enabled"
            :disabled="settingsStore.loading"
            @change="toggleCaptcha"
          />
          <div class="w-11 h-6 bg-gray-200 peer-focus:ring-2 peer-focus:ring-indigo-500 rounded-full peer
                      peer-checked:after:translate-x-full peer-checked:after:border-white
                      after:content-[''] after:absolute after:top-0.5 after:left-[2px]
                      after:bg-white after:border-gray-300 after:border after:rounded-full
                      after:h-5 after:w-5 after:transition-all peer-checked:bg-indigo-600
                      peer-disabled:opacity-50"></div>
        </label>
      </div>
    </div>
  </div>
</template>
