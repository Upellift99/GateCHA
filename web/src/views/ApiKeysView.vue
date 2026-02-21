<script setup lang="ts">
import { onMounted } from 'vue'
import { useApiKeysStore } from '../stores/apikeys'

const store = useApiKeysStore()

onMounted(() => {
  store.fetchKeys()
})
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-2xl font-bold text-gray-900">API Keys</h1>
      <router-link
        to="/keys/new"
        class="px-4 py-2 bg-indigo-600 text-white text-sm font-medium rounded-md hover:bg-indigo-700"
      >
        Create Key
      </router-link>
    </div>

    <div v-if="store.loading" class="text-center py-12 text-gray-500">Loading...</div>

    <div v-else-if="!store.keys.length" class="bg-white shadow rounded-lg p-12 text-center">
      <p class="text-gray-500 mb-4">No API keys yet</p>
      <router-link to="/keys/new" class="text-indigo-600 hover:text-indigo-800 font-medium">
        Create your first key
      </router-link>
    </div>

    <div v-else class="bg-white shadow rounded-lg overflow-hidden">
      <table class="min-w-full divide-y divide-gray-200">
        <thead class="bg-gray-50">
          <tr>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Name</th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Key ID</th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Domain</th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Status</th>
            <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase">Actions</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-gray-200">
          <tr v-for="key in store.keys" :key="key.id" class="hover:bg-gray-50">
            <td class="px-6 py-4 text-sm font-medium text-gray-900">{{ key.name || '-' }}</td>
            <td class="px-6 py-4 text-sm text-gray-500 font-mono">{{ key.key_id }}</td>
            <td class="px-6 py-4 text-sm text-gray-500">{{ key.domain || '*' }}</td>
            <td class="px-6 py-4">
              <span
                :class="key.enabled ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'"
                class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium"
              >
                {{ key.enabled ? 'Active' : 'Disabled' }}
              </span>
            </td>
            <td class="px-6 py-4 text-right space-x-3">
              <router-link :to="`/keys/${key.id}`" class="text-indigo-600 hover:text-indigo-800 text-sm">
                View
              </router-link>
              <router-link :to="`/keys/${key.id}/edit`" class="text-gray-600 hover:text-gray-800 text-sm">
                Edit
              </router-link>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
