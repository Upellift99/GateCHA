<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useApiKeysStore, type APIKey } from '../stores/apikeys'
import { useStatsStore } from '../stores/stats'
import StatsChart from '../components/StatsChart.vue'

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
  return `${window.location.origin}/api/v1/challenge?apiKey=${key.value.key_id}`
})

const widgetSnippet = computed(() => {
  if (!key.value) return ''
  return `<altcha-widget
  challengeurl="${challengeUrl.value}"
></altcha-widget>`
})
</script>

<template>
  <div v-if="key">
    <div class="flex items-center justify-between mb-6">
      <div>
        <router-link to="/keys" class="text-sm text-gray-500 hover:text-gray-700">&larr; Back to keys</router-link>
        <h1 class="text-2xl font-bold text-gray-900 mt-1">{{ key.name || key.key_id }}</h1>
      </div>
      <div class="flex gap-2">
        <button @click="toggleEnabled" :class="key.enabled ? 'bg-yellow-500 hover:bg-yellow-600' : 'bg-green-500 hover:bg-green-600'" class="px-4 py-2 text-white text-sm font-medium rounded-md">
          {{ key.enabled ? 'Disable' : 'Enable' }}
        </button>
        <router-link :to="`/keys/${key.id}/edit`" class="px-4 py-2 bg-gray-600 text-white text-sm font-medium rounded-md hover:bg-gray-700">
          Edit
        </router-link>
        <button @click="showDeleteConfirm = true" class="px-4 py-2 bg-red-600 text-white text-sm font-medium rounded-md hover:bg-red-700">
          Delete
        </button>
      </div>
    </div>

    <!-- Key Info -->
    <div class="bg-white shadow rounded-lg p-6 mb-6">
      <h2 class="text-lg font-medium text-gray-900 mb-4">Key Details</h2>
      <dl class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div>
          <dt class="text-sm font-medium text-gray-500">Key ID</dt>
          <dd class="mt-1 flex items-center gap-2">
            <code class="text-sm bg-gray-100 px-2 py-1 rounded font-mono">{{ key.key_id }}</code>
            <button @click="copyToClipboard(key.key_id, 'key')" class="text-xs text-indigo-600 hover:text-indigo-800">
              {{ copied === 'key' ? 'Copied!' : 'Copy' }}
            </button>
          </dd>
        </div>
        <div>
          <dt class="text-sm font-medium text-gray-500">HMAC Secret</dt>
          <dd class="mt-1 flex items-center gap-2">
            <code v-if="showSecret" class="text-sm bg-gray-100 px-2 py-1 rounded font-mono break-all">{{ key.hmac_secret }}</code>
            <code v-else class="text-sm bg-gray-100 px-2 py-1 rounded font-mono">••••••••••••••••</code>
            <button @click="showSecret = !showSecret" class="text-xs text-indigo-600 hover:text-indigo-800">
              {{ showSecret ? 'Hide' : 'Show' }}
            </button>
            <button @click="handleRotateSecret" class="text-xs text-orange-600 hover:text-orange-800">Rotate</button>
          </dd>
        </div>
        <div>
          <dt class="text-sm font-medium text-gray-500">Domain</dt>
          <dd class="mt-1 text-sm text-gray-900">{{ key.domain || 'Any (*)' }}</dd>
        </div>
        <div>
          <dt class="text-sm font-medium text-gray-500">Difficulty (maxNumber)</dt>
          <dd class="mt-1 text-sm text-gray-900">{{ key.max_number.toLocaleString() }}</dd>
        </div>
        <div>
          <dt class="text-sm font-medium text-gray-500">Challenge TTL</dt>
          <dd class="mt-1 text-sm text-gray-900">{{ key.expire_seconds }}s</dd>
        </div>
        <div>
          <dt class="text-sm font-medium text-gray-500">Algorithm</dt>
          <dd class="mt-1 text-sm text-gray-900">{{ key.algorithm }}</dd>
        </div>
      </dl>
    </div>

    <!-- Integration Snippet -->
    <div class="bg-white shadow rounded-lg p-6 mb-6">
      <h2 class="text-lg font-medium text-gray-900 mb-4">Integration</h2>
      <div class="relative">
        <pre class="bg-gray-900 text-green-400 text-sm p-4 rounded-lg overflow-x-auto"><code>{{ widgetSnippet }}</code></pre>
        <button
          @click="copyToClipboard(widgetSnippet, 'snippet')"
          class="absolute top-2 right-2 text-xs bg-gray-700 text-gray-300 px-2 py-1 rounded hover:bg-gray-600"
        >
          {{ copied === 'snippet' ? 'Copied!' : 'Copy' }}
        </button>
      </div>
    </div>

    <!-- Stats -->
    <div class="bg-white shadow rounded-lg p-6">
      <h2 class="text-lg font-medium text-gray-900 mb-4">Statistics (30 days)</h2>
      <StatsChart v-if="statsStore.keyStats.length" :data="statsStore.keyStats" />
      <p v-else class="text-gray-500 text-center py-12">No data yet</p>
    </div>

    <!-- Delete Confirmation -->
    <div v-if="showDeleteConfirm" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div class="bg-white rounded-lg p-6 max-w-sm mx-4">
        <h3 class="text-lg font-medium text-gray-900 mb-2">Delete API Key?</h3>
        <p class="text-sm text-gray-500 mb-4">This will permanently delete <strong>{{ key.name || key.key_id }}</strong> and all its statistics. This action cannot be undone.</p>
        <div class="flex justify-end gap-2">
          <button @click="showDeleteConfirm = false" class="px-4 py-2 text-sm font-medium text-gray-700 bg-gray-100 rounded-md hover:bg-gray-200">Cancel</button>
          <button @click="handleDelete" class="px-4 py-2 text-sm font-medium text-white bg-red-600 rounded-md hover:bg-red-700">Delete</button>
        </div>
      </div>
    </div>
  </div>
  <div v-else class="text-center py-12 text-gray-500">Loading...</div>
</template>
