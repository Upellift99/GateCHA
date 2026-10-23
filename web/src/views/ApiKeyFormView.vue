<script setup lang="ts">
import { ref, computed, onMounted, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useApiKeysStore } from '../stores/apikeys'
import CopyButton from '../components/CopyButton.vue'

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
  rate_limit_per_min: 0,
  adaptive_difficulty: false,
  his_sampling: false,
  his_enforce: false,
  his_threshold: 0.8,
})

const error = ref('')
const loading = ref(false)
const createdKey = ref<{ id: number; key_id: string; hmac_secret: string } | null>(null)
const createdHeading = ref<HTMLElement | null>(null)

onMounted(async () => {
  if (isEdit.value) {
    const key = await store.getKey(keyId.value)
    form.value = {
      name: key.name,
      domain: key.domain,
      max_number: key.max_number,
      expire_seconds: key.expire_seconds,
      algorithm: key.algorithm,
      rate_limit_per_min: key.rate_limit_per_min,
      adaptive_difficulty: key.adaptive_difficulty,
      his_sampling: key.his_sampling,
      his_enforce: key.his_enforce,
      his_threshold: key.his_threshold,
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
      createdKey.value = { id: key.id, key_id: key.key_id, hmac_secret: key.hmac_secret || '' }
      // The form is replaced by the result; move focus there so keyboard and
      // screen-reader users land on it instead of on a node that disappeared.
      await nextTick()
      createdHeading.value?.focus()
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
    <router-link to="/keys" class="inline-flex min-h-6 items-center text-sm text-muted hover:text-ink">&larr; Back to keys</router-link>
    <h1 class="text-2xl font-extrabold tracking-tight text-ink mt-1 mb-6">
      {{ isEdit ? 'Edit API Key' : 'Create API Key' }}
    </h1>

    <!-- Created key: the API key is the one value to copy. The HMAC secret
         used to sit beside it as an equal, which read as "copy both"; GateCHA
         signs and checks challenges itself, so it is tucked away. -->
    <div v-if="createdKey" class="rounded-md border border-ok-edge bg-ok-soft p-6 mb-6">
      <h2 ref="createdHeading" tabindex="-1" class="text-[17px] font-extrabold text-ok focus:outline-none">API key created</h2>
      <p class="mt-1 max-w-prose text-sm text-ok">
        Copy this key into your widget's challenge URL and your backend's
        <code class="font-mono">/api/v1/verify</code> call. It is the only value your site needs.
      </p>

      <div class="mt-4 flex flex-wrap items-center gap-3">
        <code class="min-w-0 break-all rounded-md border border-ok-edge bg-panel px-3 py-2 font-mono text-base text-ink">{{
          createdKey.key_id
        }}</code>
        <CopyButton :value="createdKey.key_id" label="Copy API key" variant="button" />
      </div>

      <div class="mt-5 flex flex-wrap gap-x-6 gap-y-2 text-sm font-medium">
        <router-link :to="`/keys/${createdKey.id}`" class="inline-flex min-h-6 items-center text-link hover:text-link-hover">
          See the integration snippets &rarr;
        </router-link>
        <router-link to="/keys" class="inline-flex min-h-6 items-center text-ink-2 hover:text-ink">
          Back to API keys
        </router-link>
      </div>

      <details class="mt-5 border-t border-ok-edge pt-4 text-sm">
        <summary class="min-h-6 cursor-pointer py-0.5 font-medium text-ok">
          Advanced: HMAC secret
        </summary>
        <p class="mt-2 max-w-prose text-ok">
          You don't need this for the widget or <code class="font-mono">/api/v1/verify</code>: GateCHA uses it
          to sign challenges on its side. It stays on the key's page if you ever need it.
        </p>
        <code class="mt-2 block break-all rounded-md bg-panel px-3 py-2 font-mono text-ink">{{ createdKey.hmac_secret }}</code>
      </details>
    </div>

    <!-- Form -->
    <form v-else @submit.prevent="handleSubmit" class="rounded-md border border-rule bg-panel p-6 space-y-4">
      <div v-if="error" role="alert" class="bg-fail-soft text-fail px-4 py-3 rounded text-sm">{{ error }}</div>

      <div>
        <label for="key-name" class="block text-sm font-medium text-ink-2 mb-1">Name</label>
        <input
          id="key-name"
          v-model="form.name"
          type="text"
          placeholder="My Website"
          class="w-full px-3 py-2 border border-rule-strong rounded-md focus:outline-none focus:ring-2 focus:ring-focus"
        />
      </div>

      <div>
        <label for="key-domain" class="block text-sm font-medium text-ink-2 mb-1">Domains</label>
        <textarea
          id="key-domain"
          v-model="form.domain"
          rows="3"
          placeholder="example.com&#10;app.example.com&#10;*.staging.example.com"
          class="w-full px-3 py-2 border border-rule-strong rounded-md font-mono text-sm focus:outline-none focus:ring-2 focus:ring-focus"
        ></textarea>
        <p class="mt-1 max-w-[52ch] text-xs text-muted">
          Optional. One domain per line restricts the key to those origins. Use
          <code class="font-mono">*.example.com</code> to allow every subdomain (and the bare domain).
          Leave empty to allow any.
        </p>
      </div>

      <div class="grid grid-cols-2 gap-4">
        <div>
          <label for="key-max-number" class="block text-sm font-medium text-ink-2 mb-1">Difficulty (maxNumber)</label>
          <input
            id="key-max-number"
            v-model.number="form.max_number"
            type="number"
            min="1000"
            max="10000000"
            class="w-full px-3 py-2 border border-rule-strong rounded-md focus:outline-none focus:ring-2 focus:ring-focus"
          />
          <p class="mt-1 max-w-[52ch] text-xs text-muted">Higher = harder. 100,000 ~ 0.5s</p>
        </div>
        <div>
          <label for="key-expire-seconds" class="block text-sm font-medium text-ink-2 mb-1">Challenge TTL (seconds)</label>
          <input
            id="key-expire-seconds"
            v-model.number="form.expire_seconds"
            type="number"
            min="60"
            max="3600"
            class="w-full px-3 py-2 border border-rule-strong rounded-md focus:outline-none focus:ring-2 focus:ring-focus"
          />
        </div>
      </div>

      <div>
        <label for="key-algorithm" class="block text-sm font-medium text-ink-2 mb-1">Algorithm</label>
        <select
          id="key-algorithm"
          v-model="form.algorithm"
          class="w-full px-3 py-2 border border-rule-strong rounded-md focus:outline-none focus:ring-2 focus:ring-focus"
        >
          <option value="SHA-256">SHA-256 (recommended)</option>
          <option value="SHA-512">SHA-512</option>
        </select>
      </div>

      <div>
        <label for="key-rate-limit" class="block text-sm font-medium text-ink-2 mb-1">Rate limit (requests / minute)</label>
        <input
          id="key-rate-limit"
          v-model.number="form.rate_limit_per_min"
          type="number"
          min="0"
          max="100000"
          class="w-full px-3 py-2 border border-rule-strong rounded-md focus:outline-none focus:ring-2 focus:ring-focus"
        />
        <p class="mt-1 max-w-[52ch] text-xs text-muted">Caps challenge + verify calls for this key, across all clients. 0 = unlimited.</p>
      </div>

      <div class="flex items-start gap-3">
        <input
          id="key-adaptive"
          v-model="form.adaptive_difficulty"
          type="checkbox"
          class="mt-0.5 h-5 w-5 shrink-0 rounded border-rule-strong text-link focus:ring-focus"
        />
        <label for="key-adaptive" class="text-sm">
          <span class="font-medium text-ink-2">Adaptive difficulty</span>
          <span class="block max-w-[52ch] text-xs text-muted">Raises proof-of-work difficulty above the base for clients (by IP) requesting challenges at an abusive rate. The difficulty above is the floor.</span>
        </label>
      </div>

      <div class="flex items-start gap-3">
        <input
          id="key-his-sampling"
          v-model="form.his_sampling"
          type="checkbox"
          class="mt-0.5 h-5 w-5 shrink-0 rounded border-rule-strong text-link focus:ring-focus"
        />
        <label for="key-his-sampling" class="text-sm">
          <span class="font-medium text-ink-2">HIS sampling</span>
          <span class="block max-w-[52ch] text-xs text-muted">Stores each scored interaction sample (privacy-preserving aggregates only — no IP, coordinates or key contents) so HIS enforcement thresholds can be calibrated on real traffic. Samples are pruned after the retention window.</span>
        </label>
      </div>

      <div class="flex items-start gap-3">
        <input
          id="key-his-enforce"
          v-model="form.his_enforce"
          type="checkbox"
          class="mt-0.5 h-5 w-5 shrink-0 rounded border-rule-strong text-link focus:ring-focus"
        />
        <label for="key-his-enforce" class="text-sm">
          <span class="font-medium text-ink-2">Block suspected bots</span>
          <span class="block max-w-[52ch] text-xs text-muted">
            Rejects a verification whose interaction score reaches the threshold below, instead of
            only reporting it. Only ever acts on requests that actually carried
            <code class="font-mono">his_signals</code>: a site whose collector never runs is
            unaffected. Read this key's calibration histogram before switching it on.
          </span>
        </label>
      </div>

      <div>
        <label for="key-his-threshold" class="block text-sm font-medium text-ink-2 mb-1">
          Suspect threshold
        </label>
        <input
          id="key-his-threshold"
          v-model.number="form.his_threshold"
          type="number"
          min="0.05"
          max="1"
          step="0.05"
          class="w-full px-3 py-2 border border-rule-strong rounded-md focus:ring-2 focus:ring-focus focus:border-focus"
        />
        <p class="mt-1 max-w-[52ch] text-xs text-muted">
          Score from 0 to 1 at or above which this key calls a sample automation. Higher means more
          bot-like, the reverse of reCAPTCHA. Drives the reported
          <code class="font-mono">his_bot_suspected</code>, the Monitor counters and the blocking
          above, so lowering it raises the bot-suspected figure on the dashboard by design.
          <span class="block mt-1">
            0.70 is the score of a sample whose collector observed nothing at all, which a
            keyboard-only or screen-reader visitor also produces. Going at or below it without
            reading your own histogram will reject those visitors.
          </span>
        </p>
      </div>

      <button
        type="submit"
        :disabled="loading"
        class="w-full py-2 px-4 bg-action text-action-ink text-sm font-semibold rounded-md hover:bg-action-hover focus:outline-none focus-visible:ring-2 focus-visible:ring-focus disabled:opacity-50"
      >
        {{ loading ? 'Saving...' : (isEdit ? 'Update key' : 'Create key') }}
      </button>
    </form>
  </div>
</template>
