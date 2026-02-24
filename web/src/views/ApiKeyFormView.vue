<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useApiKeysStore } from '../stores/apikeys'

const route = useRoute()
const router = useRouter()
const store = useApiKeysStore()

const isEdit = computed(() => route.name === 'keys-edit')
const keyId = computed(() => Number(route.params.id))

const form = ref({
  name: '',
  domain: '',
  max_number: 100000,
  expire_seconds: 300,
  algorithm: 'SHA-256',
})

const error = ref('')
const loading = ref(false)
const createdKey = ref<{ key_id: string; hmac_secret: string } | null>(null)

onMounted(async () => {
  if (isEdit.value) {
    const key = await store.getKey(keyId.value)
    form.value = {
      name: key.name,
      domain: key.domain,
      max_number: key.max_number,
      expire_seconds: key.expire_seconds,
      algorithm: key.algorithm,
    }
  }
})

async function handleSubmit() {
  error.value = ''
  loading.value = true
  try {
    if (isEdit.value) {
      await store.updateKey(keyId.value, form.value)
      router.push(`/keys/${keyId.value}`)
    } else {
      const key = await store.createKey(form.value)
      createdKey.value = { key_id: key.key_id, hmac_secret: key.hmac_secret || '' }
    }
  } catch {
    error.value = 'Failed to save key'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="max-w-xl">
    <router-link to="/keys" class="text-sm text-gray-500 hover:text-gray-700">&larr; Back to keys</router-link>
    <h1 class="text-2xl font-bold text-gray-900 mt-1 mb-6">
      {{ isEdit ? 'Edit API Key' : 'Create API Key' }}
    </h1>

    <!-- Created key info -->
    <div v-if="createdKey" class="bg-green-50 border border-green-200 rounded-lg p-6 mb-6">
      <h2 class="text-lg font-medium text-green-900 mb-2">Key Created Successfully!</h2>
      <dl class="space-y-2 text-sm">
        <div>
          <dt class="font-medium text-green-800">Key ID</dt>
          <dd class="font-mono bg-white px-2 py-1 rounded mt-1">{{ createdKey.key_id }}</dd>
        </div>
        <div>
          <dt class="font-medium text-green-800">HMAC Secret</dt>
          <dd class="font-mono bg-white px-2 py-1 rounded mt-1 break-all">{{ createdKey.hmac_secret }}</dd>
        </div>
      </dl>
      <router-link to="/keys" class="inline-block mt-4 text-sm text-indigo-600 hover:text-indigo-800 font-medium">
        Go to API Keys &rarr;
      </router-link>
    </div>

    <!-- Form -->
    <form v-else @submit.prevent="handleSubmit" class="bg-white shadow rounded-lg p-6 space-y-4">
      <div v-if="error" class="bg-red-50 text-red-700 px-4 py-3 rounded text-sm">{{ error }}</div>

      <div>
        <label for="key-name" class="block text-sm font-medium text-gray-700 mb-1">Name</label>
        <input
          id="key-name"
          v-model="form.name"
          type="text"
          placeholder="My Website"
          class="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
        />
      </div>

      <div>
        <label for="key-domain" class="block text-sm font-medium text-gray-700 mb-1">Domain</label>
        <input
          id="key-domain"
          v-model="form.domain"
          type="text"
          placeholder="example.com (leave empty for any)"
          class="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
        />
        <p class="mt-1 text-xs text-gray-500">Optional. Restricts the API key to this domain.</p>
      </div>

      <div class="grid grid-cols-2 gap-4">
        <div>
          <label for="key-max-number" class="block text-sm font-medium text-gray-700 mb-1">Difficulty (maxNumber)</label>
          <input
            id="key-max-number"
            v-model.number="form.max_number"
            type="number"
            min="1000"
            max="10000000"
            class="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
          />
          <p class="mt-1 text-xs text-gray-500">Higher = harder. 100,000 ~ 0.5s</p>
        </div>
        <div>
          <label for="key-expire-seconds" class="block text-sm font-medium text-gray-700 mb-1">Challenge TTL (seconds)</label>
          <input
            id="key-expire-seconds"
            v-model.number="form.expire_seconds"
            type="number"
            min="60"
            max="3600"
            class="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
          />
        </div>
      </div>

      <div>
        <label for="key-algorithm" class="block text-sm font-medium text-gray-700 mb-1">Algorithm</label>
        <select
          id="key-algorithm"
          v-model="form.algorithm"
          class="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
        >
          <option value="SHA-256">SHA-256 (recommended)</option>
          <option value="SHA-512">SHA-512</option>
        </select>
      </div>

      <button
        type="submit"
        :disabled="loading"
        class="w-full py-2 px-4 bg-indigo-600 text-white font-medium rounded-md hover:bg-indigo-700 disabled:opacity-50"
      >
        {{ loading ? 'Saving...' : (isEdit ? 'Update Key' : 'Create Key') }}
      </button>
    </form>
  </div>
</template>
