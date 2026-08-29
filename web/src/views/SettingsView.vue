<script setup lang="ts">
import { ref, onMounted } from 'vue'
import api from '../lib/api'
import { useSettingsStore } from '../stores/settings'
import MCPTokensPanel from '../components/MCPTokensPanel.vue'
import ToggleSwitch from '../components/ToggleSwitch.vue'

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

async function toggleCaptcha(checked: boolean) {
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
  <div class="space-y-6">
    <h1 class="text-2xl font-bold text-slate-900">Settings</h1>

    <!-- Two columns from lg up: the account cards are narrow by nature, the MCP
         panel carries a table and wants the room. Below lg they stack. -->
    <div class="grid items-start gap-6 lg:grid-cols-5">
      <div class="space-y-6 lg:col-span-2">
        <form
          @submit.prevent="handleSubmit"
          class="rounded-xl border border-slate-200 bg-white shadow-sm p-6 space-y-4"
        >
          <h2 class="text-lg font-medium text-slate-900">Change Password</h2>

          <div v-if="error" class="bg-red-50 text-red-700 px-4 py-3 rounded text-sm">{{ error }}</div>
          <div v-if="success" class="bg-green-50 text-green-700 px-4 py-3 rounded text-sm">{{ success }}</div>

          <div>
            <label for="currentPassword" class="block text-sm font-medium text-slate-700 mb-1">Current Password</label>
            <input
              id="currentPassword"
              v-model="currentPassword"
              type="password"
              required
              class="w-full px-3 py-2 border border-slate-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-teal-500"
            />
          </div>

          <div>
            <label for="newPassword" class="block text-sm font-medium text-slate-700 mb-1">New Password</label>
            <input
              id="newPassword"
              v-model="newPassword"
              type="password"
              required
              class="w-full px-3 py-2 border border-slate-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-teal-500"
            />
          </div>

          <div>
            <label for="confirmPassword" class="block text-sm font-medium text-slate-700 mb-1">Confirm New Password</label>
            <input
              id="confirmPassword"
              v-model="confirmPassword"
              type="password"
              required
              class="w-full px-3 py-2 border border-slate-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-teal-500"
            />
          </div>

          <button
            type="submit"
            :disabled="loading"
            class="py-2 px-4 bg-teal-600 text-white font-medium rounded-md hover:bg-teal-700 disabled:opacity-50"
          >
            {{ loading ? 'Saving...' : 'Change Password' }}
          </button>
        </form>

        <div class="rounded-xl border border-slate-200 bg-white shadow-sm p-6 space-y-4">
          <h2 class="text-lg font-medium text-slate-900">Security</h2>

          <div v-if="settingsError" class="bg-red-50 text-red-700 px-4 py-3 rounded text-sm">
            {{ settingsError }}
          </div>

          <!-- One label around the whole row: the checkbox is sr-only, so the pill
               beside it is the only thing a user can click and it has to sit inside
               a label to activate anything. See #146. -->
          <label for="loginCaptchaToggle" class="flex items-center justify-between cursor-pointer">
            <span>
              <span class="block text-sm font-medium text-slate-700">Login CAPTCHA</span>
              <span class="block text-xs text-slate-500 mt-0.5">
                Require an ALTCHA proof-of-work challenge before signing in.
              </span>
            </span>
            <ToggleSwitch
              id="loginCaptchaToggle"
              class="ml-4"
              :model-value="settingsStore.settings.login_captcha_enabled"
              :disabled="settingsStore.loading"
              @update:model-value="toggleCaptcha"
            />
          </label>
        </div>
      </div>

      <div class="lg:col-span-3">
        <MCPTokensPanel />
      </div>
    </div>
  </div>
</template>
