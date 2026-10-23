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
    <h1 class="text-2xl font-extrabold tracking-tight text-ink">Settings</h1>

    <!-- Two columns from lg up: the account cards are narrow by nature, the MCP
         panel carries a table and wants the room. Below lg they stack. -->
    <div class="grid items-start gap-6 lg:grid-cols-5">
      <div class="space-y-6 lg:col-span-2">
        <form
          @submit.prevent="handleSubmit"
          class="rounded-md border border-rule bg-panel p-6 space-y-4"
        >
          <h2 class="text-[17px] font-extrabold text-ink">Change Password</h2>

          <div v-if="error" role="alert" class="bg-fail-soft text-fail px-4 py-3 rounded text-sm">{{ error }}</div>
          <div v-if="success" class="bg-ok-soft text-ok px-4 py-3 rounded text-sm">{{ success }}</div>

          <div>
            <label for="currentPassword" class="block text-sm font-medium text-ink-2 mb-1">Current Password</label>
            <input
              id="currentPassword"
              v-model="currentPassword"
              type="password"
              required
              class="w-full px-3 py-2 border border-rule-strong rounded-md focus:outline-none focus:ring-2 focus:ring-focus"
            />
          </div>

          <div>
            <label for="newPassword" class="block text-sm font-medium text-ink-2 mb-1">New Password</label>
            <input
              id="newPassword"
              v-model="newPassword"
              type="password"
              required
              class="w-full px-3 py-2 border border-rule-strong rounded-md focus:outline-none focus:ring-2 focus:ring-focus"
            />
          </div>

          <div>
            <label for="confirmPassword" class="block text-sm font-medium text-ink-2 mb-1">Confirm New Password</label>
            <input
              id="confirmPassword"
              v-model="confirmPassword"
              type="password"
              required
              class="w-full px-3 py-2 border border-rule-strong rounded-md focus:outline-none focus:ring-2 focus:ring-focus"
            />
          </div>

          <button
            type="submit"
            :disabled="loading"
            class="py-2 px-4 bg-action text-action-ink font-medium rounded-md hover:bg-action-hover disabled:opacity-50"
          >
            {{ loading ? 'Saving...' : 'Change Password' }}
          </button>
        </form>

        <div class="rounded-md border border-rule bg-panel p-6 space-y-4">
          <h2 class="text-[17px] font-extrabold text-ink">Security</h2>

          <div v-if="settingsError" role="alert" class="bg-fail-soft text-fail px-4 py-3 rounded text-sm">
            {{ settingsError }}
          </div>

          <!-- One label around the whole row: the checkbox is sr-only, so the pill
               beside it is the only thing a user can click and it has to sit inside
               a label to activate anything. See #146. -->
          <label for="loginCaptchaToggle" class="flex items-center justify-between cursor-pointer">
            <span>
              <span class="block text-sm font-medium text-ink-2">Login CAPTCHA</span>
              <span class="block text-xs text-muted mt-0.5">
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
