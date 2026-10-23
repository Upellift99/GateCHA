<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useApiKeysStore, type APIKey } from '../stores/apikeys'
import { useStatsStore } from '../stores/stats'
import StatsChart from '../components/StatsChart.vue'
import CountryTraffic from '../components/CountryTraffic.vue'
import HISCalibrationPanel from '../components/HISCalibrationPanel.vue'
import CopyButton from '../components/CopyButton.vue'
import LineBullet from '../components/LineBullet.vue'
import StatusChip from '../components/StatusChip.vue'
import { lineStatus, todayUTC } from '../lib/lines'
import ConfirmDialog from '../components/ConfirmDialog.vue'

const route = useRoute()
const router = useRouter()
const keysStore = useApiKeysStore()
const statsStore = useStatsStore()

// Same daily status as the dashboard board and the keys list.
const todayRow = computed(() => statsStore.keyStats.find((d) => d.date === todayUTC()))
const status = computed(() => (key.value ? lineStatus(key.value.enabled, todayRow.value) : 'quiet'))

const key = ref<APIKey | null>(null)
const showSecret = ref(false)
const showDeleteConfirm = ref(false)
const deleting = ref(false)

const keyId = computed(() => Number(route.params.id))

/**
 * Mirrors the server's own fallback (models.APIKey.SuspectThreshold): an
 * out-of-range or absent value means the default applies, not "reject
 * everything". Reading it through one computed also keeps the badge, the
 * caption and the settings row from ever disagreeing.
 */
const suspectThreshold = computed(() => {
  const t = key.value?.his_threshold
  return typeof t === 'number' && t > 0 && t <= 1 ? t : 0.8
})

const domainList = computed(() =>
  (key.value?.domain ?? '')
    .split(/[\n,]/)
    .map((d) => d.trim())
    .filter(Boolean),
)

onMounted(async () => {
  key.value = await keysStore.getKey(keyId.value)
  statsStore.fetchKeyStats(keyId.value)
})

async function handleDelete() {
  // The dialog stays up until the navigation happens, so without this a second
  // click sends a second DELETE.
  if (deleting.value) return
  deleting.value = true
  try {
    await keysStore.deleteKey(keyId.value)
    router.push('/keys')
  } finally {
    deleting.value = false
  }
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

const instanceOrigin = computed(() => globalThis.location.origin)

const challengeUrl = computed(() => {
  if (!key.value) return ''
  return `${instanceOrigin.value}/api/v1/challenge?apiKey=${key.value.key_id}`
})

const widgetSnippet = computed(() => {
  if (!key.value) return ''
  return `<altcha-widget
  challenge="${challengeUrl.value}"
></altcha-widget>`
})

// Kept separate from the widget snippet on purpose: collecting interaction
// signals is an opt-in an integrator makes deliberately, not something that
// arrives by copying the standard snippet.
// The snippet ends in a closing script tag, which cannot appear literally
// anywhere in this block (comments included): the SFC parser would end the
// block there. Escaping the slash works but reads as a stray escape to linters
// (S6535), so the tag is assembled from two pieces instead.
const SCRIPT_CLOSE = '</' + 'script>'

const hisSnippet = computed(
  () => `<script src="${instanceOrigin.value}/api/public/his.js" defer>${SCRIPT_CLOSE}`,
)

const hisTotals = computed(() => ({
  observations: statsStore.keyStats.reduce((s, d) => s + (d.his_observations || 0), 0),
  suspected: statsStore.keyStats.reduce((s, d) => s + (d.his_bot_suspected || 0), 0),
}))
</script>

<template>
  <div v-if="key">
    <!-- Stacks below sm: a long key name next to three buttons squeezed the
         title to one word per line and pushed Delete off a 390px screen. -->
    <div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between mb-6">
      <div class="min-w-0">
        <router-link to="/keys" class="inline-flex min-h-6 items-center text-sm text-muted hover:text-ink">&larr; Back to keys</router-link>
        <div class="mt-1 flex items-start gap-3">
          <LineBullet :id="key.id" size="lg" class="mt-0.5" />
          <div class="min-w-0">
            <h1 class="text-2xl font-extrabold tracking-tight text-ink break-words">{{ key.name || key.key_id }}</h1>
            <p class="mt-0.5 flex flex-wrap items-center gap-x-3 gap-y-1 text-sm text-muted">
              <span>Line {{ key.id }}</span>
              <span>{{ key.enabled ? 'Key enabled' : 'Key disabled' }}</span>
              <StatusChip :status="status" :failed-today="todayRow?.verifications_fail ?? 0" />
            </p>
          </div>
        </div>
      </div>
      <div class="flex flex-wrap gap-2 sm:shrink-0">
        <button @click="toggleEnabled" class="px-4 py-2 border border-rule-strong bg-panel text-ink text-sm font-semibold rounded-md hover:bg-sunken focus:outline-none focus-visible:ring-2 focus-visible:ring-focus">
          {{ key.enabled ? 'Disable' : 'Enable' }}
        </button>
        <router-link :to="`/keys/${key.id}/edit`" class="px-4 py-2 border border-rule-strong bg-panel text-ink text-sm font-semibold rounded-md hover:bg-sunken focus:outline-none focus-visible:ring-2 focus-visible:ring-focus">
          Edit
        </router-link>
        <button @click="showDeleteConfirm = true" class="px-4 py-2 bg-danger text-action-ink text-sm font-semibold rounded-md hover:bg-danger-hover focus:outline-none focus-visible:ring-2 focus-visible:ring-focus">
          Delete
        </button>
      </div>
    </div>

    <!-- Key Info -->
    <div class="rounded-md border border-rule bg-panel p-6 mb-6">
      <h2 class="text-[17px] font-extrabold text-ink mb-4">Key Details</h2>
      <dl class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div>
          <dt class="text-sm font-medium text-muted">Key ID</dt>
          <dd class="mt-1 flex items-center gap-2">
            <code class="text-sm bg-sunken px-2 py-1 rounded font-mono">{{ key.key_id }}</code>
            <CopyButton :value="key.key_id" label="Copy key ID" />
          </dd>
        </div>
        <div>
          <dt class="text-sm font-medium text-muted">HMAC Secret</dt>
          <dd class="mt-1 flex items-center gap-2">
            <code v-if="showSecret" class="text-sm bg-sunken px-2 py-1 rounded font-mono break-all">{{ key.hmac_secret }}</code>
            <code v-else class="text-sm bg-sunken px-2 py-1 rounded font-mono">••••••••••••••••</code>
            <button @click="showSecret = !showSecret" class="inline-flex min-h-8 items-center px-2 text-xs text-link hover:text-link-hover">
              {{ showSecret ? 'Hide' : 'Show' }}
            </button>
            <button @click="handleRotateSecret" class="inline-flex min-h-8 items-center px-2 text-xs font-semibold text-fail hover:text-ink">Rotate</button>
          </dd>
        </div>
        <div class="md:col-span-2">
          <dt class="text-sm font-medium text-muted">Instance URL</dt>
          <dd class="mt-1 flex items-center gap-2">
            <code class="text-sm bg-sunken px-2 py-1 rounded font-mono break-all">{{ instanceOrigin }}</code>
            <CopyButton :value="instanceOrigin" label="Copy instance URL" />
          </dd>
        </div>
        <div>
          <dt class="text-sm font-medium text-muted">Domains</dt>
          <dd class="mt-1 text-sm text-ink">
            <span v-if="domainList.length === 0">Any (*)</span>
            <span
              v-for="d in domainList"
              :key="d"
              class="inline-block mr-1 mb-1 px-2 py-0.5 rounded bg-sunken font-mono text-xs text-ink-2"
              >{{ d }}</span
            >
          </dd>
        </div>
        <div>
          <dt class="text-sm font-medium text-muted">Difficulty (maxNumber)</dt>
          <dd class="mt-1 font-mono text-sm text-ink tabular-nums">{{ key.max_number.toLocaleString('en-US') }}</dd>
        </div>
        <div>
          <dt class="text-sm font-medium text-muted">Challenge TTL</dt>
          <dd class="mt-1 font-mono text-sm text-ink tabular-nums">{{ key.expire_seconds }}s</dd>
        </div>
        <div>
          <dt class="text-sm font-medium text-muted">Algorithm</dt>
          <dd class="mt-1 text-sm text-ink">{{ key.algorithm }}</dd>
        </div>
        <div>
          <dt class="text-sm font-medium text-muted">Rate limit</dt>
          <dd class="mt-1 font-mono text-sm text-ink tabular-nums">
            {{ key.rate_limit_per_min > 0 ? `${key.rate_limit_per_min.toLocaleString('en-US')} req/min` : 'Unlimited' }}
          </dd>
        </div>
        <div>
          <dt class="text-sm font-medium text-muted">Adaptive difficulty</dt>
          <dd class="mt-1">
            <span
              :class="key.adaptive_difficulty ? 'bg-ok-soft text-ok' : 'bg-sunken text-ink-2'"
              class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium"
            >{{ key.adaptive_difficulty ? 'On' : 'Off' }}</span>
          </dd>
        </div>
        <div>
          <dt class="text-sm font-medium text-muted">HIS sampling</dt>
          <dd class="mt-1">
            <span
              :class="key.his_sampling ? 'bg-ok-soft text-ok' : 'bg-sunken text-ink-2'"
              class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium"
            >{{ key.his_sampling ? 'On' : 'Off' }}</span>
          </dd>
        </div>
        <div>
          <dt class="text-sm font-medium text-muted">Block suspected bots</dt>
          <dd class="mt-1">
            <span
              :class="key.his_enforce ? 'bg-fail-soft text-fail' : 'bg-sunken text-ink-2'"
              class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium"
            >{{ key.his_enforce ? 'Blocking' : 'Monitor only' }}</span>
          </dd>
        </div>
        <div>
          <dt class="text-sm font-medium text-muted">Suspect threshold</dt>
          <dd class="mt-1 font-mono text-sm text-ink tabular-nums">{{ suspectThreshold.toFixed(2) }}</dd>
        </div>
      </dl>
    </div>

    <!-- Integration Snippet -->
    <div class="rounded-md border border-rule bg-panel p-6 mb-6">
      <h2 class="text-[17px] font-extrabold text-ink mb-4">Integration</h2>
      <div class="relative">
        <pre class="bg-code text-code-ink text-sm p-4 pr-12 rounded-md overflow-x-auto"><code>{{ widgetSnippet }}</code></pre>
        <CopyButton
          :value="widgetSnippet"
          label="Copy snippet"
          variant="overlay"
          class="absolute top-2 right-2"
        />
      </div>

      <h3 class="text-sm font-medium text-ink-2 mt-6 mb-1">Interaction signals (optional)</h3>
      <p class="text-xs text-muted mb-2">
        Add this alongside the widget to score how visitors interact, on top of the proof of work.
        It fills a hidden <code class="font-mono">gatecha_his_signals</code> field that your backend
        forwards to <code class="font-mono">/verify</code>. Aggregates only, and it never blocks.
      </p>
      <div class="relative">
        <pre class="bg-code text-code-ink text-sm p-4 pr-12 rounded-md overflow-x-auto"><code>{{ hisSnippet }}</code></pre>
        <CopyButton
          :value="hisSnippet"
          label="Copy HIS snippet"
          variant="overlay"
          class="absolute top-2 right-2"
        />
      </div>
    </div>

    <!-- Stats -->
    <div class="mb-6 rounded-md border border-rule bg-panel p-6">
      <h2 class="text-[17px] font-extrabold text-ink mb-4">Statistics (30 days)</h2>
      <div class="mb-1 flex flex-wrap gap-x-6 gap-y-1 text-sm">
        <span class="text-muted">HIS <span class="rounded-[3px] border border-rule px-1.5 py-0.5 text-[11px] font-bold uppercase tracking-[0.08em] text-muted">Monitor</span>:</span>
        <span class="text-ink"><span class="font-mono tabular-nums">{{ hisTotals.observations.toLocaleString('en-US') }}</span> observed</span>
        <span class="text-ink"><span class="font-mono tabular-nums">{{ hisTotals.suspected.toLocaleString('en-US') }}</span> bot-suspected</span>
      </div>
      <p v-if="key.his_enforce" class="text-xs text-muted mb-4">
        Blocking is on for this key: a verification scoring
        {{ suspectThreshold.toFixed(2) }} or above is rejected with
        <code class="font-mono">bot_suspected</code> and counted as a failed verification.
      </p>
      <p v-else class="text-xs text-muted mb-4">
        Monitor mode: suspected requests are counted and reported, never blocked. Turn on
        "Block suspected bots" in this key's settings to act on the score.
      </p>
      <StatsChart v-if="statsStore.keyStats.length" :data="statsStore.keyStats" :line-id="key.id" />
      <p v-else class="text-muted text-center py-12">No data yet</p>
    </div>

    <CountryTraffic :countries="statsStore.keyCountries" roomy class="mb-6" />

    <HISCalibrationPanel
      v-if="key.his_sampling"
      :key-id="key.id"
      :observations="hisTotals.observations"
    />


    <ConfirmDialog
      :open="showDeleteConfirm"
      title="Delete API Key?"
      confirm-label="Delete"
      tone="danger"
      :busy="deleting"
      @cancel="showDeleteConfirm = false"
      @confirm="handleDelete"
    >
      This will permanently delete <strong class="font-medium text-ink-2">{{ key.name || key.key_id }}</strong>
      and all its statistics. This action cannot be undone.
    </ConfirmDialog>
  </div>
  <div v-else class="text-center py-12 text-muted">Loading...</div>
</template>
