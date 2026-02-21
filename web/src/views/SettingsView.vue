<script setup lang="ts">
import { ref } from 'vue'
import api from '../lib/api'

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
</script>

<template>
  <div class="max-w-xl">
    <h1 class="text-2xl font-bold text-gray-900 mb-6">Settings</h1>

    <form @submit.prevent="handleSubmit" class="bg-white shadow rounded-lg p-6 space-y-4">
      <h2 class="text-lg font-medium text-gray-900">Change Password</h2>

      <div v-if="error" class="bg-red-50 text-red-700 px-4 py-3 rounded text-sm">{{ error }}</div>
      <div v-if="success" class="bg-green-50 text-green-700 px-4 py-3 rounded text-sm">{{ success }}</div>

      <div>
        <label class="block text-sm font-medium text-gray-700 mb-1">Current Password</label>
        <input
          v-model="currentPassword"
          type="password"
          required
          class="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
        />
      </div>

      <div>
        <label class="block text-sm font-medium text-gray-700 mb-1">New Password</label>
        <input
          v-model="newPassword"
          type="password"
          required
          class="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
        />
      </div>

      <div>
        <label class="block text-sm font-medium text-gray-700 mb-1">Confirm New Password</label>
        <input
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
  </div>
</template>
