<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useSettingsStore } from '../stores/settings'
import { useMCPTokensStore } from '../stores/mcptokens'
import CopyButton from './CopyButton.vue'

const settingsStore = useSettingsStore()
const store = useMCPTokensStore()

const error = ref('')
const name = ref('')
const readOnly = ref(false)
const creating = ref(false)
// Held only until the operator dismisses it: the secret cannot be read back.
const newSecret = ref('')

onMounted(() => {
  store.fetchTokens()
})

async function toggleMCP(event: Event) {
  const checked = (event.target as HTMLInputElement).checked
  error.value = ''
  try {
    await settingsStore.updateSettings({ mcp_enabled: checked })
  } catch {
    error.value = 'Failed to update setting.'
    await settingsStore.fetchSettings()
  }
}

async function createToken() {
  if (!name.value.trim()) return
  error.value = ''
  creating.value = true
  try {
    newSecret.value = await store.createToken({ name: name.value.trim(), read_only: readOnly.value })
    name.value = ''
    readOnly.value = false
  } catch {
    error.value = 'Failed to create token.'
  } finally {
    creating.value = false
  }
}

async function revokeToken(id: number, tokenName: string) {
  if (!globalThis.confirm(`Revoke "${tokenName}"? Any client using it will stop working.`)) return
  error.value = ''
  try {
    await store.revokeToken(id)
  } catch {
    error.value = 'Failed to revoke token.'
  }
}

function formatDate(value: string | null): string {
  if (!value) return 'Never'
  return new Date(value).toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: 'numeric' })
}
</script>

<template>
  <div class="rounded-xl border border-slate-200 bg-white shadow-sm p-6 space-y-4">
    <h2 class="text-lg font-medium text-slate-900">MCP Access</h2>

    <div v-if="error" class="bg-red-50 text-red-700 px-4 py-3 rounded text-sm">{{ error }}</div>

    <div class="flex items-center justify-between">
      <div>
        <label for="mcpToggle" class="text-sm font-medium text-slate-700">MCP endpoint</label>
        <p class="text-xs text-slate-500 mt-0.5">
          Serve the API key management tools over MCP at <code class="font-mono">/mcp</code>.
          This is a second way in to full admin access, so it stays off until you need it.
        </p>
      </div>
      <div class="relative inline-flex items-center cursor-pointer shrink-0 ml-4">
        <input
          id="mcpToggle"
          type="checkbox"
          class="sr-only peer"
          :checked="settingsStore.settings.mcp_enabled"
          :disabled="settingsStore.loading"
          @change="toggleMCP"
        />
        <div class="w-11 h-6 bg-slate-200 peer-focus:ring-2 peer-focus:ring-teal-500 rounded-full peer
                    peer-checked:after:translate-x-full peer-checked:after:border-white
                    after:content-[''] after:absolute after:top-0.5 after:left-[2px]
                    after:bg-white after:border-slate-300 after:border after:rounded-full
                    after:h-5 after:w-5 after:transition-all peer-checked:bg-teal-600
                    peer-disabled:opacity-50"></div>
      </div>
    </div>

    <div v-if="newSecret" class="rounded-md border border-amber-300 bg-amber-50 p-4 space-y-2">
      <p class="text-sm font-medium text-amber-900">Copy this token now. It will not be shown again.</p>
      <div class="flex items-center gap-2">
        <code class="text-sm bg-white px-2 py-1 rounded font-mono break-all flex-1">{{ newSecret }}</code>
        <CopyButton :value="newSecret" label="Copy MCP token" />
      </div>
      <button type="button" class="text-sm text-amber-900 underline" @click="newSecret = ''">Done</button>
    </div>

    <form class="flex flex-wrap items-end gap-3 border-t border-slate-200 pt-4" @submit.prevent="createToken">
      <div class="flex-1 min-w-48">
        <label for="mcpTokenName" class="block text-sm font-medium text-slate-700 mb-1">New token</label>
        <input
          id="mcpTokenName"
          v-model="name"
          type="text"
          placeholder="Who is this token for?"
          class="w-full px-3 py-2 border border-slate-300 rounded-md shadow-sm text-sm focus:outline-none focus:ring-2 focus:ring-teal-500"
        />
      </div>
      <label class="flex items-center gap-2 text-sm text-slate-700 pb-2">
        <input v-model="readOnly" type="checkbox" class="rounded border-slate-300 text-teal-600 focus:ring-teal-500" />
        Read only
      </label>
      <button
        type="submit"
        :disabled="creating || !name.trim()"
        class="py-2 px-4 bg-teal-600 text-white text-sm font-medium rounded-md hover:bg-teal-700 disabled:opacity-50"
      >
        {{ creating ? 'Creating...' : 'Create token' }}
      </button>
    </form>

    <p v-if="!store.tokens.length && !store.loading" class="text-sm text-slate-500">
      No tokens yet. Create one per person rather than sharing a single token, so you can revoke
      access for one machine without disturbing the others.
    </p>

    <table v-else-if="store.tokens.length" class="min-w-full divide-y divide-slate-200 text-sm">
      <thead>
        <tr class="text-left text-xs font-medium text-slate-500 uppercase">
          <th class="py-2">Name</th>
          <th class="py-2">Token</th>
          <th class="py-2">Access</th>
          <th class="py-2">Last used</th>
          <th class="py-2"></th>
        </tr>
      </thead>
      <tbody class="divide-y divide-slate-200">
        <tr v-for="token in store.tokens" :key="token.id">
          <td class="py-2 font-medium text-slate-900">{{ token.name }}</td>
          <td class="py-2 font-mono text-xs text-slate-400">{{ token.display }}…</td>
          <td class="py-2">
            <span
              :class="token.read_only ? 'bg-slate-100 text-slate-700' : 'bg-amber-100 text-amber-800'"
              class="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium"
            >
              {{ token.read_only ? 'Read only' : 'Full access' }}
            </span>
          </td>
          <td class="py-2 text-slate-500">{{ formatDate(token.last_used_at) }}</td>
          <td class="py-2 text-right">
            <button
              type="button"
              class="text-red-600 hover:text-red-800"
              :aria-label="`Revoke ${token.name}`"
              @click="revokeToken(token.id, token.name)"
            >
              Revoke
            </button>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>
