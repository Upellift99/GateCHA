<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { useApiKeysStore, type APIKey } from '../stores/apikeys'
import { useStatsStore } from '../stores/stats'
import CopyButton from '../components/CopyButton.vue'
import { domainList } from '../lib/format'
import StatusChip from '../components/StatusChip.vue'
import { keyStatus, todayUTC } from '../lib/status'

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

// Today's row per key: the list reads the same status as the dashboard board.
const todayByKey = computed(() => {
  const today = todayUTC()
  return new Map(statsStore.keysDaily.filter((d) => d.date === today).map((d) => [d.api_key_id, d]))
})

function todayStatus(key: APIKey) {
  const row = todayByKey.value.get(key.id)
  return { status: keyStatus(key.enabled, row), failed: row?.verifications_fail ?? 0 }
}

function getLastUsed(keyId: number): string {
  const summary = statsStore.keysSummary[String(keyId)]
  return summary ? summary.last_used_at : ''
}

const DAY_MS = 86_400_000

// last_used_at is a UTC calendar day, so compare whole UTC days. A short,
// relative label keeps the column on one line; the full date sits in the title.
function formatLastUsed(keyId: number): string {
  const date = getLastUsed(keyId)
  if (!date) return 'Never'
  const day = Date.parse(date + 'T00:00:00Z')
  const now = new Date()
  const days = Math.round((Date.UTC(now.getUTCFullYear(), now.getUTCMonth(), now.getUTCDate()) - day) / DAY_MS)
  if (days <= 0) return 'Today'
  if (days === 1) return 'Yesterday'
  if (days < 7) return `${days} days ago`
  const sameYear = new Date(day).getUTCFullYear() === now.getUTCFullYear()
  return new Date(day).toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
    ...(sameYear ? {} : { year: 'numeric' }),
    timeZone: 'UTC',
  })
}

function formatLastUsedFull(keyId: number): string | undefined {
  const date = getLastUsed(keyId)
  if (!date) return undefined
  return new Date(date + 'T00:00:00Z').toLocaleDateString('en-US', { dateStyle: 'full', timeZone: 'UTC' })
}


// Counts read as plain data; only failures that actually happened get colour.
function statClass(value: number, kind: 'neutral' | 'failure'): string {
  if (value === 0) return 'text-muted'
  return kind === 'failure' ? 'font-medium text-fail' : 'text-ink'
}

function ariaSort(column: SortColumn): 'ascending' | 'descending' | 'none' {
  if (sortColumn.value !== column) return 'none'
  return sortDirection.value === 'asc' ? 'ascending' : 'descending'
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
  statsStore.fetchKeysDaily(1)
})
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-2xl font-bold tracking-tight text-ink">API Keys</h1>
      <router-link
        to="/keys/new"
        class="inline-flex min-h-10 items-center gap-2 rounded-md px-4 text-sm font-semibold bg-action text-action-ink hover:bg-action-hover focus:outline-none focus-visible:ring-2 focus-visible:ring-focus focus-visible:ring-offset-2"
      >
        <svg class="h-3.5 w-3.5" viewBox="0 0 14 14" aria-hidden="true"><path d="M7 1v12M1 7h12" stroke="currentColor" stroke-width="2" /></svg>
        Create key
      </router-link>
    </div>

    <div v-if="store.loading" class="text-center py-12 text-muted">Loading...</div>

    <div v-else-if="!store.keys.length" class="rounded-md border border-rule bg-panel p-12 text-center">
      <p class="text-muted mb-4">No API keys yet</p>
      <router-link to="/keys/new" class="text-link hover:text-link-hover font-medium">
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
            class="w-full rounded-md border border-rule-strong bg-panel py-2 pl-3 pr-9 text-sm text-ink placeholder-muted focus:border-focus focus:outline-none focus:ring-1 focus:ring-focus"
          />
          <button
            v-if="search"
            type="button"
            aria-label="Clear search"
            class="absolute inset-y-0 right-0 flex items-center px-3 text-muted hover:text-ink"
            @click="search = ''"
          >
            &#x2715;
          </button>
        </div>
      </div>

      <div
        v-if="!sortedKeys.length"
        class="rounded-md border border-rule bg-panel p-12 text-center"
      >
        <p class="text-muted">No keys match "{{ search.trim() }}"</p>
        <button type="button" class="mt-4 text-link hover:text-link-hover font-medium" @click="search = ''">
          Clear search
        </button>
      </div>

      <div v-else class="rounded-md border border-rule bg-panel overflow-hidden">
        <!-- overflow-hidden on the card used to clip the right-hand columns,
             Actions included, on phones. Secondary columns now drop out below
             sm/md, and whatever is left can scroll. -->
        <div class="overflow-x-auto">
          <table class="min-w-full divide-y divide-rule">
            <thead class="bg-sunken">
              <tr>
                <th
                  class="px-4 py-3 text-left text-[11px] font-semibold uppercase tracking-[0.09em] text-muted hover:text-ink select-none"
                  :aria-sort="ariaSort('name')"
                  @click="toggleSort('name')"
                >
                  <button type="button" class="inline-flex min-h-6 items-center uppercase cursor-pointer hover:text-ink">
                    Name
                    <span v-if="sortColumn === 'name'" class="ml-1">{{ sortDirection === 'asc' ? '\u25B2' : '\u25BC' }}</span>
                  </button>
                </th>
                <th
                  class="hidden sm:table-cell px-4 py-3 text-left text-[11px] font-semibold uppercase tracking-[0.09em] text-muted hover:text-ink select-none"
                  :aria-sort="ariaSort('domain')"
                  @click="toggleSort('domain')"
                >
                  <button type="button" class="inline-flex min-h-6 items-center uppercase cursor-pointer hover:text-ink">
                    Domain
                    <span v-if="sortColumn === 'domain'" class="ml-1">{{ sortDirection === 'asc' ? '\u25B2' : '\u25BC' }}</span>
                  </button>
                </th>
                <th
                  class="hidden sm:table-cell px-4 py-3 text-left text-[11px] font-semibold uppercase tracking-[0.09em] text-muted hover:text-ink select-none"
                  :aria-sort="ariaSort('enabled')"
                  @click="toggleSort('enabled')"
                >
                  <button type="button" class="inline-flex min-h-6 items-center uppercase cursor-pointer hover:text-ink">
                    Status
                    <span v-if="sortColumn === 'enabled'" class="ml-1">{{ sortDirection === 'asc' ? '\u25B2' : '\u25BC' }}</span>
                  </button>
                </th>
                <th
                  class="hidden md:table-cell px-4 py-3 text-right text-[11px] font-semibold uppercase tracking-[0.09em] text-muted hover:text-ink select-none"
                  :aria-sort="ariaSort('challenges')"
                  @click="toggleSort('challenges')"
                >
                  <button type="button" class="inline-flex min-h-6 items-center uppercase cursor-pointer hover:text-ink">
                    Challenges
                    <span v-if="sortColumn === 'challenges'" class="ml-1">{{ sortDirection === 'asc' ? '\u25B2' : '\u25BC' }}</span>
                  </button>
                </th>
                <th
                  class="hidden md:table-cell px-4 py-3 text-right text-[11px] font-semibold uppercase tracking-[0.09em] text-muted hover:text-ink select-none"
                  :aria-sort="ariaSort('verified')"
                  @click="toggleSort('verified')"
                >
                  <button type="button" class="inline-flex min-h-6 items-center uppercase cursor-pointer hover:text-ink">
                    Verified
                    <span v-if="sortColumn === 'verified'" class="ml-1">{{ sortDirection === 'asc' ? '\u25B2' : '\u25BC' }}</span>
                  </button>
                </th>
                <th
                  class="hidden md:table-cell px-4 py-3 text-right text-[11px] font-semibold uppercase tracking-[0.09em] text-muted hover:text-ink select-none"
                  :aria-sort="ariaSort('failed')"
                  @click="toggleSort('failed')"
                >
                  <button type="button" class="inline-flex min-h-6 items-center uppercase cursor-pointer hover:text-ink">
                    Failed
                    <span v-if="sortColumn === 'failed'" class="ml-1">{{ sortDirection === 'asc' ? '\u25B2' : '\u25BC' }}</span>
                  </button>
                </th>
                <th
                  class="hidden md:table-cell px-4 py-3 text-right text-[11px] font-semibold uppercase tracking-[0.09em] text-muted hover:text-ink select-none whitespace-nowrap"
                  :aria-sort="ariaSort('lastUsed')"
                  @click="toggleSort('lastUsed')"
                >
                  <button type="button" class="inline-flex min-h-6 items-center uppercase cursor-pointer hover:text-ink">
                    Last Used
                    <span v-if="sortColumn === 'lastUsed'" class="ml-1">{{ sortDirection === 'asc' ? '\u25B2' : '\u25BC' }}</span>
                  </button>
                </th>
                <th class="hidden sm:table-cell px-4 py-3 text-right text-[11px] font-semibold uppercase tracking-[0.09em] text-muted">Actions</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-rule">
              <tr
                v-for="key in pagedKeys"
                :key="key.id"
                :class="todayStatus(key).status === 'failing' ? 'bg-warn-soft' : 'hover:bg-sunken'"
              >
                <td class="px-4 py-3">
                  <div>
                    <div class="min-w-0">
                      <router-link :to="`/keys/${key.id}`" class="inline-flex min-h-6 items-center text-sm font-semibold text-ink hover:text-link">{{ key.name || '-' }}</router-link>
                      <div class="flex items-center gap-1.5">
                        <span class="text-xs text-muted font-mono">{{ key.key_id }}</span>
                        <CopyButton :value="key.key_id" :label="`Copy ${key.key_id}`" />
                      </div>
                      <!-- Phones: the status moves under the name instead of its own column -->
                      <StatusChip :status="todayStatus(key).status" :failed-today="todayStatus(key).failed" class="mt-1 sm:hidden" />
                    </div>
                  </div>
                </td>
                <td class="hidden sm:table-cell px-4 py-3 text-sm text-ink-2" :title="domainList(key.domain).join('\n') || undefined">
                  <template v-if="domainList(key.domain).length">
                    <span class="inline-block max-w-[200px] truncate align-bottom">{{ domainList(key.domain)[0] }}</span>
                    <span
                      v-if="domainList(key.domain).length > 1"
                      class="ml-1.5 inline-flex rounded bg-sunken px-1.5 text-xs font-medium text-ink-2 tabular-nums"
                    >+{{ domainList(key.domain).length - 1 }}</span>
                  </template>
                  <span v-else class="text-muted">Any domain</span>
                </td>
                <td class="hidden sm:table-cell px-4 py-3">
                  <StatusChip :status="todayStatus(key).status" :failed-today="todayStatus(key).failed" compact />
                </td>
                <td
                  class="hidden md:table-cell px-4 py-3 text-right font-mono text-sm tabular-nums"
                  :class="statClass(getKeyStat(key.id, 'challenges_issued'), 'neutral')"
                >
                  {{ getKeyStat(key.id, 'challenges_issued').toLocaleString('en-US') }}
                </td>
                <td
                  class="hidden md:table-cell px-4 py-3 text-right font-mono text-sm tabular-nums"
                  :class="statClass(getKeyStat(key.id, 'verifications_ok'), 'neutral')"
                >
                  {{ getKeyStat(key.id, 'verifications_ok').toLocaleString('en-US') }}
                </td>
                <td
                  class="hidden md:table-cell px-4 py-3 text-right font-mono text-sm tabular-nums"
                  :class="statClass(getKeyStat(key.id, 'verifications_fail'), 'failure')"
                >
                  {{ getKeyStat(key.id, 'verifications_fail').toLocaleString('en-US') }}
                </td>
                <td
                  class="hidden md:table-cell px-4 py-3 text-sm text-right text-ink-2 whitespace-nowrap"
                  :title="formatLastUsedFull(key.id)"
                >
                  {{ formatLastUsed(key.id) }}
                </td>
                <td class="hidden sm:table-cell px-4 py-3 text-right whitespace-nowrap">
                  <router-link :to="`/keys/${key.id}/edit`" :aria-label="`Edit ${key.name || key.key_id}`" class="inline-flex min-h-8 items-center px-2 text-ink-2 hover:text-ink text-sm">
                    Edit
                  </router-link>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <div class="flex flex-wrap items-center justify-between gap-3 border-t border-rule bg-sunken px-6 py-3">
          <p class="text-sm text-muted" aria-live="polite">
            Showing {{ rangeStart }}&ndash;{{ rangeEnd }} of {{ sortedKeys.length }}
            <span v-if="search.trim()"> (filtered from {{ store.keys.length }})</span>
          </p>

          <div class="flex items-center gap-3">
            <div v-if="paginationRelevant" class="flex items-center gap-2">
              <label for="key-page-size" class="text-sm text-muted">Per page</label>
              <select
                id="key-page-size"
                v-model.number="pageSize"
                class="rounded-md border border-rule-strong bg-panel py-1 pl-2 pr-7 text-sm text-ink-2 focus:border-focus focus:outline-none focus:ring-1 focus:ring-focus"
              >
                <option v-for="size in PAGE_SIZE_OPTIONS" :key="size" :value="size">{{ size }}</option>
              </select>
            </div>

            <div v-if="totalPages > 1" class="flex items-center gap-2">
              <button
                type="button"
                aria-label="Previous page"
                :disabled="currentPage <= 1"
                class="rounded-md border border-rule-strong bg-panel px-3 py-1 text-sm text-ink-2 hover:bg-sunken disabled:cursor-not-allowed disabled:opacity-50"
                @click="goToPage(currentPage - 1)"
              >
                Previous
              </button>
              <span class="text-sm text-muted">Page {{ currentPage }} of {{ totalPages }}</span>
              <button
                type="button"
                aria-label="Next page"
                :disabled="currentPage >= totalPages"
                class="rounded-md border border-rule-strong bg-panel px-3 py-1 text-sm text-ink-2 hover:bg-sunken disabled:cursor-not-allowed disabled:opacity-50"
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
