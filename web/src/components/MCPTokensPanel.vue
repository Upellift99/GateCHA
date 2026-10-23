<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useSettingsStore } from '../stores/settings'
import { useMCPTokensStore } from '../stores/mcptokens'
import CopyButton from './CopyButton.vue'
import ConfirmDialog from './ConfirmDialog.vue'
import ToggleSwitch from './ToggleSwitch.vue'

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

async function toggleMCP(checked: boolean) {
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

// Kept past the close so the name does not blank out mid fade-out; the dialog's
// own flag is what opens and closes it.
const revokeTarget = ref<{ id: number; name: string } | null>(null)
const revokeOpen = ref(false)
const revoking = ref(false)

function askRevoke(id: number, tokenName: string) {
  revokeTarget.value = { id, name: tokenName }
  revokeOpen.value = true
}

async function confirmRevoke() {
  const target = revokeTarget.value
  if (!target) return
  error.value = ''
  revoking.value = true
  try {
    await store.revokeToken(target.id)
  } catch {
    error.value = 'Failed to revoke token.'
  } finally {
    revoking.value = false
    revokeOpen.value = false
  }
}

function formatDate(value: string | null): string {
  if (!value) return 'Never'
  return new Date(value).toLocaleDateString('en-US', { year: 'numeric', month: 'short', day: 'numeric' })
}
</script>

<template>
  <div class="rounded-md border border-rule bg-panel p-6 space-y-4">
    <h2 class="text-[17px] font-extrabold text-ink">MCP Access</h2>

    <div v-if="error" role="alert" class="bg-fail-soft text-fail px-4 py-3 rounded text-sm">{{ error }}</div>

    <!-- One label around the whole row: the checkbox is sr-only, so the pill
         beside it is the only thing a user can click and it has to sit inside
         a label to activate anything. See #146. -->
    <label for="mcpToggle" class="flex items-center justify-between cursor-pointer">
      <span>
        <span class="block text-sm font-medium text-ink-2">MCP endpoint</span>
        <span class="block text-xs text-muted mt-0.5">
          Serve the API key management tools over MCP at <code class="font-mono">/mcp</code>.
          This is a second way in to full admin access, so it stays off until you need it.
        </span>
      </span>
      <ToggleSwitch
        id="mcpToggle"
        class="ml-4"
        :model-value="settingsStore.settings.mcp_enabled"
        :disabled="settingsStore.loading"
        @update:model-value="toggleMCP"
      />
    </label>

    <div v-if="newSecret" class="rounded-md border border-warn-edge bg-warn-soft p-4 space-y-2">
      <p class="text-sm font-medium text-warn-text">Copy this token now. It will not be shown again.</p>
      <div class="flex items-center gap-2">
        <code class="text-sm bg-panel px-2 py-1 rounded font-mono break-all flex-1">{{ newSecret }}</code>
        <CopyButton :value="newSecret" label="Copy MCP token" />
      </div>
      <button type="button" class="text-sm text-warn-text underline" @click="newSecret = ''">Done</button>
    </div>

    <form class="flex flex-wrap items-end gap-3 border-t border-rule pt-4" @submit.prevent="createToken">
      <div class="flex-1 min-w-48">
        <label for="mcpTokenName" class="block text-sm font-medium text-ink-2 mb-1">New token</label>
        <input
          id="mcpTokenName"
          v-model="name"
          type="text"
          placeholder="Who is this token for?"
          class="w-full px-3 py-2 border border-rule-strong rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-focus"
        />
      </div>
      <label class="flex items-center gap-2 text-sm text-ink-2 pb-2">
        <input v-model="readOnly" type="checkbox" class="h-5 w-5 rounded border-rule-strong text-link focus:ring-focus" />
        Read only
      </label>
      <button
        type="submit"
        :disabled="creating || !name.trim()"
        class="py-2 px-4 bg-action text-action-ink text-sm font-medium rounded-md hover:bg-action-hover disabled:opacity-50"
      >
        {{ creating ? 'Creating...' : 'Create token' }}
      </button>
    </form>

    <p v-if="!store.tokens.length && !store.loading" class="text-sm text-muted">
      No tokens yet. Create one per person rather than sharing a single token, so you can revoke
      access for one machine without disturbing the others.
    </p>

    <table v-else-if="store.tokens.length" class="min-w-full divide-y divide-rule text-sm">
      <thead>
        <tr class="text-left text-xs font-medium text-muted uppercase">
          <th class="py-2">Name</th>
          <th class="py-2">Token</th>
          <th class="py-2">Access</th>
          <th class="py-2">Last used</th>
          <th class="py-2"></th>
        </tr>
      </thead>
      <tbody class="divide-y divide-rule">
        <tr v-for="token in store.tokens" :key="token.id">
          <td class="py-2 font-medium text-ink">{{ token.name }}</td>
          <td class="py-2 font-mono text-xs text-muted">{{ token.display }}…</td>
          <td class="py-2">
            <span
              :class="token.read_only ? 'bg-sunken text-ink-2' : 'bg-warn-soft text-warn-text'"
              class="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium"
            >
              {{ token.read_only ? 'Read only' : 'Full access' }}
            </span>
          </td>
          <td class="py-2 text-muted">{{ formatDate(token.last_used_at) }}</td>
          <td class="py-2 text-right">
            <button
              type="button"
              class="text-fail hover:text-fail"
              :aria-label="`Revoke ${token.name}`"
              @click="askRevoke(token.id, token.name)"
            >
              Revoke
            </button>
          </td>
        </tr>
      </tbody>
    </table>

    <ConfirmDialog
      :open="revokeOpen"
      title="Revoke this token?"
      confirm-label="Revoke"
      tone="danger"
      :busy="revoking"
      @cancel="revokeOpen = false"
      @confirm="confirmRevoke"
    >
      <strong class="font-medium text-ink-2">{{ revokeTarget?.name }}</strong> stops working
      the moment you confirm, and any client still using it will start failing. Revoking cannot be
      undone: issue a new token instead.
    </ConfirmDialog>
  </div>
</template>
