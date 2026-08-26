<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { useApiKeysStore, type APIKey } from '../stores/apikeys'
import { useStatsStore } from '../stores/stats'
import CopyButton from '../components/CopyButton.vue'

const store = useApiKeysStore()
const statsStore = useStatsStore()

type SortColumn = 'name' | 'domain' | 'enabled' | 'challenges' | 'verified' | 'failed' | 'lastUsed'
const sortColumn = ref<SortColumn>('name')
const sortDirection = ref<'asc' | 'desc'>('asc')

const PAGE_SIZE_OPTIONS = [25, 50, 100]
const search = ref('')
const page = ref(1)
const pageSize = ref(PAGE_SIZE_OPTIONS[0])

function getKeyStat(keyId: number, field: 'challenges_issued' | 'verifications_ok' | 'verifications_fail'): number {
  const summary = statsStore.keysSummary[String(keyId)]
  return summary ? summary[field] : 0
}

function getLastUsed(keyId: number): string {
  const summary = statsStore.keysSummary[String(keyId)]
  return summary ? summary.last_used_at : ''
}

function formatLastUsed(keyId: number): string {
  const date = getLastUsed(keyId)
  if (!date) return 'Never'
  return new Date(date + 'T00:00:00Z').toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: 'numeric' })
}

function toggleSort(column: SortColumn) {
  if (sortColumn.value === column) {
    sortDirection.value = sortDirection.value === 'asc' ? 'desc' : 'asc'
  } else {
    sortColumn.value = column
    sortDirection.value = 'asc'
  }
  page.value = 1
}

// Filter first, then sort, then slice: sorting has to see the whole filtered set
// so the order is global rather than shuffled within each page.
const filteredKeys = computed(() => {
  const q = search.value.trim().toLowerCase()
  if (!q) return store.keys
  return store.keys.filter((k: APIKey) =>
    (k.name || '').toLowerCase().includes(q) ||
    (k.domain || '').toLowerCase().includes(q) ||
    (k.key_id || '').toLowerCase().includes(q)
  )
})

const sortedKeys = computed(() => {
  const keys = [...filteredKeys.value]
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
      case 'lastUsed':
        return dir * (getLastUsed(a.id) || '').localeCompare(getLastUsed(b.id) || '')
      default:
        return 0
    }
  })
})

const totalPages = computed(() => Math.max(1, Math.ceil(sortedKeys.value.length / pageSize.value)))

// Clamping here rather than in a watcher keeps the slice valid on the very same
// render as the change (a deleted key can shrink the list under the current page).
const currentPage = computed(() => Math.min(page.value, totalPages.value))

const pagedKeys = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value
  return sortedKeys.value.slice(start, start + pageSize.value)
})

const rangeStart = computed(() => (sortedKeys.value.length ? (currentPage.value - 1) * pageSize.value + 1 : 0))
const rangeEnd = computed(() => Math.min(currentPage.value * pageSize.value, sortedKeys.value.length))

// Only worth showing the page-size control once the list is long enough to page.
const paginationRelevant = computed(() => store.keys.length > PAGE_SIZE_OPTIONS[0])

watch([search, pageSize], () => {
  page.value = 1
})

function goToPage(n: number) {
  page.value = Math.min(Math.max(n, 1), totalPages.value)
}

onMounted(() => {
  store.fetchKeys()
  statsStore.fetchKeysSummary()
})
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-2xl font-bold text-slate-900">API Keys</h1>
      <router-link
        to="/keys/new"
        class="px-4 py-2 bg-teal-600 text-white text-sm font-medium rounded-md hover:bg-teal-700"
      >
        Create Key
      </router-link>
    </div>

    <div v-if="store.loading" class="text-center py-12 text-slate-500">Loading...</div>

    <div v-else-if="!store.keys.length" class="rounded-xl border border-slate-200 bg-white shadow-sm p-12 text-center">
      <p class="text-slate-500 mb-4">No API keys yet</p>
      <router-link to="/keys/new" class="text-teal-600 hover:text-teal-800 font-medium">
        Create your first key
      </router-link>
    </div>

    <div v-else>
      <div class="mb-4">
        <label for="key-search" class="sr-only">Search API keys</label>
        <div class="relative max-w-md">
          <input
            id="key-search"
            v-model="search"
            type="search"
            placeholder="Search by name, domain or key ID"
            class="w-full rounded-md border border-slate-300 bg-white py-2 pl-3 pr-9 text-sm text-slate-900 placeholder-slate-400 focus:border-teal-500 focus:outline-none focus:ring-1 focus:ring-teal-500"
          />
          <button
            v-if="search"
            type="button"
            aria-label="Clear search"
            class="absolute inset-y-0 right-0 flex items-center px-3 text-slate-400 hover:text-slate-600"
            @click="search = ''"
          >
            &#x2715;
          </button>
        </div>
      </div>

      <div
        v-if="!sortedKeys.length"
        class="rounded-xl border border-slate-200 bg-white shadow-sm p-12 text-center"
      >
        <p class="text-slate-500">No keys match "{{ search.trim() }}"</p>
        <button type="button" class="mt-4 text-teal-600 hover:text-teal-800 font-medium" @click="search = ''">
          Clear search
        </button>
      </div>

      <div v-else class="rounded-xl border border-slate-200 bg-white shadow-sm overflow-hidden">
        <table class="min-w-full divide-y divide-slate-200">
          <thead class="bg-slate-50">
            <tr>
              <th
                class="px-6 py-3 text-left text-xs font-medium text-slate-500 uppercase cursor-pointer hover:text-slate-700 select-none"
                @click="toggleSort('name')"
              >
                Name
                <span v-if="sortColumn === 'name'" class="ml-1">{{ sortDirection === 'asc' ? '\u25B2' : '\u25BC' }}</span>
              </th>
              <th
                class="px-6 py-3 text-left text-xs font-medium text-slate-500 uppercase cursor-pointer hover:text-slate-700 select-none"
                @click="toggleSort('domain')"
              >
                Domain
                <span v-if="sortColumn === 'domain'" class="ml-1">{{ sortDirection === 'asc' ? '\u25B2' : '\u25BC' }}</span>
              </th>
              <th
                class="px-6 py-3 text-left text-xs font-medium text-slate-500 uppercase cursor-pointer hover:text-slate-700 select-none"
                @click="toggleSort('enabled')"
              >
                Status
                <span v-if="sortColumn === 'enabled'" class="ml-1">{{ sortDirection === 'asc' ? '\u25B2' : '\u25BC' }}</span>
              </th>
              <th
                class="px-6 py-3 text-right text-xs font-medium text-slate-500 uppercase cursor-pointer hover:text-slate-700 select-none"
                @click="toggleSort('challenges')"
              >
                Challenges
                <span v-if="sortColumn === 'challenges'" class="ml-1">{{ sortDirection === 'asc' ? '\u25B2' : '\u25BC' }}</span>
              </th>
              <th
                class="px-6 py-3 text-right text-xs font-medium text-slate-500 uppercase cursor-pointer hover:text-slate-700 select-none"
                @click="toggleSort('verified')"
              >
                Verified
                <span v-if="sortColumn === 'verified'" class="ml-1">{{ sortDirection === 'asc' ? '\u25B2' : '\u25BC' }}</span>
              </th>
              <th
                class="px-6 py-3 text-right text-xs font-medium text-slate-500 uppercase cursor-pointer hover:text-slate-700 select-none"
                @click="toggleSort('failed')"
              >
                Failed
                <span v-if="sortColumn === 'failed'" class="ml-1">{{ sortDirection === 'asc' ? '\u25B2' : '\u25BC' }}</span>
              </th>
              <th
                class="px-6 py-3 text-right text-xs font-medium text-slate-500 uppercase cursor-pointer hover:text-slate-700 select-none"
                @click="toggleSort('lastUsed')"
              >
                Last Used
                <span v-if="sortColumn === 'lastUsed'" class="ml-1">{{ sortDirection === 'asc' ? '\u25B2' : '\u25BC' }}</span>
              </th>
              <th class="px-6 py-3 text-right text-xs font-medium text-slate-500 uppercase">Actions</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-200">
            <tr v-for="key in pagedKeys" :key="key.id" class="hover:bg-slate-50">
              <td class="px-6 py-4">
                <div class="text-sm font-medium text-slate-900">{{ key.name || '-' }}</div>
                <div class="flex items-center gap-1.5">
                  <span class="text-xs text-slate-400 font-mono">{{ key.key_id }}</span>
                  <CopyButton :value="key.key_id" :label="`Copy ${key.key_id}`" />
                </div>
              </td>
              <td class="px-6 py-4 text-sm text-slate-500">{{ key.domain || '*' }}</td>
              <td class="px-6 py-4">
                <span
                  :class="key.enabled ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'"
                  class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium"
                >
                  {{ key.enabled ? 'Active' : 'Disabled' }}
                </span>
              </td>
              <td class="px-6 py-4 text-sm text-right font-medium text-teal-600">
                {{ getKeyStat(key.id, 'challenges_issued').toLocaleString() }}
              </td>
              <td class="px-6 py-4 text-sm text-right font-medium text-green-600">
                {{ getKeyStat(key.id, 'verifications_ok').toLocaleString() }}
              </td>
              <td class="px-6 py-4 text-sm text-right font-medium text-red-600">
                {{ getKeyStat(key.id, 'verifications_fail').toLocaleString() }}
              </td>
              <td class="px-6 py-4 text-sm text-right text-slate-500">
                {{ formatLastUsed(key.id) }}
              </td>
              <td class="px-6 py-4 text-right space-x-3">
                <router-link :to="`/keys/${key.id}`" class="text-teal-600 hover:text-teal-800 text-sm">
                  View
                </router-link>
                <router-link :to="`/keys/${key.id}/edit`" class="text-slate-600 hover:text-slate-800 text-sm">
                  Edit
                </router-link>
              </td>
            </tr>
          </tbody>
        </table>

        <div class="flex flex-wrap items-center justify-between gap-3 border-t border-slate-200 bg-slate-50 px-6 py-3">
          <p class="text-sm text-slate-500" aria-live="polite">
            Showing {{ rangeStart }}&ndash;{{ rangeEnd }} of {{ sortedKeys.length }}
            <span v-if="search.trim()"> (filtered from {{ store.keys.length }})</span>
          </p>

          <div class="flex items-center gap-3">
            <div v-if="paginationRelevant" class="flex items-center gap-2">
              <label for="key-page-size" class="text-sm text-slate-500">Per page</label>
              <select
                id="key-page-size"
                v-model.number="pageSize"
                class="rounded-md border border-slate-300 bg-white py-1 pl-2 pr-7 text-sm text-slate-700 focus:border-teal-500 focus:outline-none focus:ring-1 focus:ring-teal-500"
              >
                <option v-for="size in PAGE_SIZE_OPTIONS" :key="size" :value="size">{{ size }}</option>
              </select>
            </div>

            <div v-if="totalPages > 1" class="flex items-center gap-2">
              <button
                type="button"
                aria-label="Previous page"
                :disabled="currentPage <= 1"
                class="rounded-md border border-slate-300 bg-white px-3 py-1 text-sm text-slate-700 hover:bg-slate-100 disabled:cursor-not-allowed disabled:opacity-50"
                @click="goToPage(currentPage - 1)"
              >
                Previous
              </button>
              <span class="text-sm text-slate-500">Page {{ currentPage }} of {{ totalPages }}</span>
              <button
                type="button"
                aria-label="Next page"
                :disabled="currentPage >= totalPages"
                class="rounded-md border border-slate-300 bg-white px-3 py-1 text-sm text-slate-700 hover:bg-slate-100 disabled:cursor-not-allowed disabled:opacity-50"
                @click="goToPage(currentPage + 1)"
              >
                Next
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
