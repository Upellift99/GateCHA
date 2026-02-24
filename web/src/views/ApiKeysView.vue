<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useApiKeysStore, type APIKey } from '../stores/apikeys'
import { useStatsStore } from '../stores/stats'

const store = useApiKeysStore()
const statsStore = useStatsStore()

type SortColumn = 'name' | 'domain' | 'enabled' | 'challenges' | 'verified' | 'failed'
const sortColumn = ref<SortColumn>('name')
const sortDirection = ref<'asc' | 'desc'>('asc')

function getKeyStat(keyId: number, field: 'challenges_issued' | 'verifications_ok' | 'verifications_fail'): number {
  const summary = statsStore.keysSummary[String(keyId)]
  return summary ? summary[field] : 0
}

function toggleSort(column: SortColumn) {
  if (sortColumn.value === column) {
    sortDirection.value = sortDirection.value === 'asc' ? 'desc' : 'asc'
  } else {
    sortColumn.value = column
    sortDirection.value = 'asc'
  }
}

const sortedKeys = computed(() => {
  const keys = [...store.keys]
  const dir = sortDirection.value === 'asc' ? 1 : -1

  return keys.sort((a: APIKey, b: APIKey) => {
    switch (sortColumn.value) {
      case 'name':
        return dir * (a.name || '').localeCompare(b.name || '')
      case 'domain':
        return dir * (a.domain || '').localeCompare(b.domain || '')
      case 'enabled':
        return dir * (Number(a.enabled) - Number(b.enabled))
      case 'challenges':
        return dir * (getKeyStat(a.id, 'challenges_issued') - getKeyStat(b.id, 'challenges_issued'))
      case 'verified':
        return dir * (getKeyStat(a.id, 'verifications_ok') - getKeyStat(b.id, 'verifications_ok'))
      case 'failed':
        return dir * (getKeyStat(a.id, 'verifications_fail') - getKeyStat(b.id, 'verifications_fail'))
      default:
        return 0
    }
  })
})

onMounted(() => {
  store.fetchKeys()
  statsStore.fetchKeysSummary()
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
            <th
              class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase cursor-pointer hover:text-gray-700 select-none"
              @click="toggleSort('name')"
            >
              Name
              <span v-if="sortColumn === 'name'" class="ml-1">{{ sortDirection === 'asc' ? '\u25B2' : '\u25BC' }}</span>
            </th>
            <th
              class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase cursor-pointer hover:text-gray-700 select-none"
              @click="toggleSort('domain')"
            >
              Domain
              <span v-if="sortColumn === 'domain'" class="ml-1">{{ sortDirection === 'asc' ? '\u25B2' : '\u25BC' }}</span>
            </th>
            <th
              class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase cursor-pointer hover:text-gray-700 select-none"
              @click="toggleSort('enabled')"
            >
              Status
              <span v-if="sortColumn === 'enabled'" class="ml-1">{{ sortDirection === 'asc' ? '\u25B2' : '\u25BC' }}</span>
            </th>
            <th
              class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase cursor-pointer hover:text-gray-700 select-none"
              @click="toggleSort('challenges')"
            >
              Challenges
              <span v-if="sortColumn === 'challenges'" class="ml-1">{{ sortDirection === 'asc' ? '\u25B2' : '\u25BC' }}</span>
            </th>
            <th
              class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase cursor-pointer hover:text-gray-700 select-none"
              @click="toggleSort('verified')"
            >
              Verified
              <span v-if="sortColumn === 'verified'" class="ml-1">{{ sortDirection === 'asc' ? '\u25B2' : '\u25BC' }}</span>
            </th>
            <th
              class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase cursor-pointer hover:text-gray-700 select-none"
              @click="toggleSort('failed')"
            >
              Failed
              <span v-if="sortColumn === 'failed'" class="ml-1">{{ sortDirection === 'asc' ? '\u25B2' : '\u25BC' }}</span>
            </th>
            <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase">Actions</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-gray-200">
          <tr v-for="key in sortedKeys" :key="key.id" class="hover:bg-gray-50">
            <td class="px-6 py-4">
              <div class="text-sm font-medium text-gray-900">{{ key.name || '-' }}</div>
              <div class="text-xs text-gray-400 font-mono">{{ key.key_id }}</div>
            </td>
            <td class="px-6 py-4 text-sm text-gray-500">{{ key.domain || '*' }}</td>
            <td class="px-6 py-4">
              <span
                :class="key.enabled ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'"
                class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium"
              >
                {{ key.enabled ? 'Active' : 'Disabled' }}
              </span>
            </td>
            <td class="px-6 py-4 text-sm text-right font-medium text-indigo-600">
              {{ getKeyStat(key.id, 'challenges_issued').toLocaleString() }}
            </td>
            <td class="px-6 py-4 text-sm text-right font-medium text-green-600">
              {{ getKeyStat(key.id, 'verifications_ok').toLocaleString() }}
            </td>
            <td class="px-6 py-4 text-sm text-right font-medium text-red-600">
              {{ getKeyStat(key.id, 'verifications_fail').toLocaleString() }}
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
