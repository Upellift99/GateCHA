<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useApiKeysStore, type APIKey } from '../stores/apikeys'
import { useStatsStore } from '../stores/stats'
import StatsChart from '../components/StatsChart.vue'
import CountryTraffic from '../components/CountryTraffic.vue'

const route = useRoute()
const router = useRouter()
const keysStore = useApiKeysStore()
const statsStore = useStatsStore()

const key = ref<APIKey | null>(null)
const showSecret = ref(false)
const copied = ref('')
const showDeleteConfirm = ref(false)

const keyId = computed(() => Number(route.params.id))

onMounted(async () => {
  key.value = await keysStore.getKey(keyId.value)
  statsStore.fetchKeyStats(keyId.value)
})

function copyToClipboard(text: string, label: string) {
  navigator.clipboard.writeText(text)
  copied.value = label
  setTimeout(() => { copied.value = '' }, 2000)
}

async function handleDelete() {
  await keysStore.deleteKey(keyId.value)
  router.push('/keys')
}

async function handleRotateSecret() {
  if (!confirm('Are you sure? This will invalidate all existing challenges for this key.')) return
  const newSecret = await keysStore.rotateSecret(keyId.value)
  if (key.value) {
    key.value.hmac_secret = newSecret
  }
}

async function toggleEnabled() {
  if (!key.value) return
  await keysStore.updateKey(keyId.value, { enabled: !key.value.enabled })
  key.value = await keysStore.getKey(keyId.value)
}

const challengeUrl = computed(() => {
  if (!key.value) return ''
  return `${globalThis.location.origin}/api/v1/challenge?apiKey=${key.value.key_id}`
})

const widgetSnippet = computed(() => {
  if (!key.value) return ''
  return `<altcha-widget
  challenge="${challengeUrl.value}"
></altcha-widget>`
})

const hisTotals = computed(() => ({
  observations: statsStore.keyStats.reduce((s, d) => s + (d.his_observations || 0), 0),
  suspected: statsStore.keyStats.reduce((s, d) => s + (d.his_bot_suspected || 0), 0),
}))
</script>

<template>
  <div v-if="key">
    <div class="flex items-center justify-between mb-6">
      <div>
        <router-link to="/keys" class="text-sm text-slate-500 hover:text-slate-700">&larr; Back to keys</router-link>
        <h1 class="text-2xl font-bold text-slate-900 mt-1">{{ key.name || key.key_id }}</h1>
      </div>
      <div class="flex gap-2">
        <button @click="toggleEnabled" :class="key.enabled ? 'bg-yellow-500 hover:bg-yellow-600' : 'bg-green-500 hover:bg-green-600'" class="px-4 py-2 text-white text-sm font-medium rounded-md">
          {{ key.enabled ? 'Disable' : 'Enable' }}
        </button>
        <router-link :to="`/keys/${key.id}/edit`" class="px-4 py-2 bg-slate-600 text-white text-sm font-medium rounded-md hover:bg-slate-700">
          Edit
        </router-link>
        <button @click="showDeleteConfirm = true" class="px-4 py-2 bg-red-600 text-white text-sm font-medium rounded-md hover:bg-red-700">
          Delete
        </button>
      </div>
    </div>

    <!-- Key Info -->
    <div class="rounded-xl border border-slate-200 bg-white shadow-sm p-6 mb-6">
      <h2 class="text-lg font-medium text-slate-900 mb-4">Key Details</h2>
      <dl class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div>
          <dt class="text-sm font-medium text-slate-500">Key ID</dt>
          <dd class="mt-1 flex items-center gap-2">
            <code class="text-sm bg-slate-100 px-2 py-1 rounded font-mono">{{ key.key_id }}</code>
            <button @click="copyToClipboard(key.key_id, 'key')" class="text-xs text-teal-600 hover:text-teal-800">
              {{ copied === 'key' ? 'Copied!' : 'Copy' }}
            </button>
          </dd>
        </div>
        <div>
          <dt class="text-sm font-medium text-slate-500">HMAC Secret</dt>
          <dd class="mt-1 flex items-center gap-2">
            <code v-if="showSecret" class="text-sm bg-slate-100 px-2 py-1 rounded font-mono break-all">{{ key.hmac_secret }}</code>
            <code v-else class="text-sm bg-slate-100 px-2 py-1 rounded font-mono">••••••••••••••••</code>
            <button @click="showSecret = !showSecret" class="text-xs text-teal-600 hover:text-teal-800">
              {{ showSecret ? 'Hide' : 'Show' }}
            </button>
            <button @click="handleRotateSecret" class="text-xs text-orange-600 hover:text-orange-800">Rotate</button>
          </dd>
        </div>
        <div>
          <dt class="text-sm font-medium text-slate-500">Domain</dt>
          <dd class="mt-1 text-sm text-slate-900">{{ key.domain || 'Any (*)' }}</dd>
        </div>
        <div>
          <dt class="text-sm font-medium text-slate-500">Difficulty (maxNumber)</dt>
          <dd class="mt-1 text-sm text-slate-900">{{ key.max_number.toLocaleString() }}</dd>
        </div>
        <div>
          <dt class="text-sm font-medium text-slate-500">Challenge TTL</dt>
          <dd class="mt-1 text-sm text-slate-900">{{ key.expire_seconds }}s</dd>
        </div>
        <div>
          <dt class="text-sm font-medium text-slate-500">Algorithm</dt>
          <dd class="mt-1 text-sm text-slate-900">{{ key.algorithm }}</dd>
        </div>
        <div>
          <dt class="text-sm font-medium text-slate-500">Rate limit</dt>
          <dd class="mt-1 text-sm text-slate-900">
            {{ key.rate_limit_per_min > 0 ? `${key.rate_limit_per_min.toLocaleString()} req/min` : 'Unlimited' }}
          </dd>
        </div>
        <div>
          <dt class="text-sm font-medium text-slate-500">Adaptive difficulty</dt>
          <dd class="mt-1 text-sm text-slate-900">{{ key.adaptive_difficulty ? 'On' : 'Off' }}</dd>
        </div>
      </dl>
    </div>

    <!-- Integration Snippet -->
    <div class="rounded-xl border border-slate-200 bg-white shadow-sm p-6 mb-6">
      <h2 class="text-lg font-medium text-slate-900 mb-4">Integration</h2>
      <div class="relative">
        <pre class="bg-slate-900 text-green-400 text-sm p-4 rounded-lg overflow-x-auto"><code>{{ widgetSnippet }}</code></pre>
        <button
          @click="copyToClipboard(widgetSnippet, 'snippet')"
          class="absolute top-2 right-2 text-xs bg-slate-700 text-slate-300 px-2 py-1 rounded hover:bg-slate-600"
        >
          {{ copied === 'snippet' ? 'Copied!' : 'Copy' }}
        </button>
      </div>
    </div>

    <!-- Stats -->
    <div class="rounded-xl border border-slate-200 bg-white shadow-sm p-6">
      <h2 class="text-lg font-medium text-slate-900 mb-4">Statistics (30 days)</h2>
      <div class="mb-4 flex flex-wrap gap-x-6 gap-y-1 text-sm">
        <span class="text-slate-500">HIS <span class="text-xs px-1.5 py-0.5 rounded bg-amber-50 text-amber-700">Monitor</span>:</span>
        <span class="text-slate-900">{{ hisTotals.observations.toLocaleString() }} observed</span>
        <span class="text-slate-900">{{ hisTotals.suspected.toLocaleString() }} bot-suspected</span>
      </div>
      <StatsChart v-if="statsStore.keyStats.length" :data="statsStore.keyStats" />
      <p v-else class="text-slate-500 text-center py-12">No data yet</p>
    </div>

    <CountryTraffic :countries="statsStore.keyCountries" />


    <!-- Delete Confirmation -->
    <div v-if="showDeleteConfirm" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div class="bg-white rounded-lg p-6 max-w-sm mx-4">
        <h3 class="text-lg font-medium text-slate-900 mb-2">Delete API Key?</h3>
        <p class="text-sm text-slate-500 mb-4">This will permanently delete <strong>{{ key.name || key.key_id }}</strong> and all its statistics. This action cannot be undone.</p>
        <div class="flex justify-end gap-2">
          <button @click="showDeleteConfirm = false" class="px-4 py-2 text-sm font-medium text-slate-700 bg-slate-100 rounded-md hover:bg-slate-200">Cancel</button>
          <button @click="handleDelete" class="px-4 py-2 text-sm font-medium text-white bg-red-600 rounded-md hover:bg-red-700">Delete</button>
        </div>
      </div>
    </div>
  </div>
  <div v-else class="text-center py-12 text-slate-500">Loading...</div>
</template>
