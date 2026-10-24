<script setup lang="ts">
import type { APIKey } from '../stores/apikeys'
import type { FailingKey } from '../lib/status'

// The first thing on the dashboard: which keys are failing today, or a quiet
// line saying everything works. "Working" needs traffic to stand on: with no
// verification yet today the line says so instead.
defineProps<{ failing: FailingKey[]; activeKeys: number; checkedToday: number }>()

const pct = (f: FailingKey) => (f.verified ? Math.round((f.failed / f.verified) * 100) : 0)
const name = (k: APIKey) => k.name || k.key_id
</script>

<template>
  <section
    v-if="failing.length"
    class="grid grid-cols-[auto_1fr] items-start gap-x-4 gap-y-3 rounded-md border border-warn-edge bg-warn-soft px-5 py-4 sm:grid-cols-[auto_1fr_auto] sm:items-center"
    aria-labelledby="attention-title"
  >
    <span class="grid h-9 w-9 place-items-center rounded-full bg-warn text-warn-ink" aria-hidden="true">
      <svg class="h-5 w-5" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round">
        <path d="M10 5v6" />
        <circle cx="10" cy="14.75" r="0.6" fill="currentColor" />
      </svg>
    </span>
    <div class="min-w-0">
      <h2 id="attention-title" class="text-[17px] font-semibold text-ink">
        {{ failing.length === 1 ? '1 key needs attention' : `${failing.length} keys need attention` }}
      </h2>
      <p v-if="failing.length === 1" class="mt-0.5 text-sm text-ink-2">
        <strong class="text-ink">{{ name(failing[0]!.key) }}</strong> failed
        <strong class="font-mono tabular-nums text-ink">{{ failing[0]!.failed.toLocaleString('en-US') }}</strong> of
        <strong class="font-mono tabular-nums text-ink">{{ failing[0]!.verified.toLocaleString('en-US') }}</strong>
        verifications today ({{ pct(failing[0]!) }}%). Check the widget on that site and the domains allowed for the key.
      </p>
      <ul v-else class="mt-1 space-y-0.5 text-sm text-ink-2">
        <li v-for="f in failing" :key="f.key.id">
          <router-link :to="`/keys/${f.key.id}`" class="font-semibold text-ink hover:text-link">{{ name(f.key) }}</router-link>
          — <span class="font-mono tabular-nums">{{ f.failed.toLocaleString('en-US') }}</span> of
          <span class="font-mono tabular-nums">{{ f.verified.toLocaleString('en-US') }}</span> failed ({{ pct(f) }}%)
        </li>
      </ul>
    </div>
    <router-link
      v-if="failing.length === 1"
      :to="`/keys/${failing[0]!.key.id}`"
      class="col-start-2 inline-flex min-h-10 items-center justify-self-start rounded-md border border-rule-strong bg-panel px-4 text-sm font-semibold text-ink hover:bg-sunken focus:outline-none focus-visible:ring-2 focus-visible:ring-focus sm:col-start-3"
    >Open key</router-link>
  </section>

  <p v-else-if="activeKeys && checkedToday" class="flex items-center gap-2 rounded-md border border-ok-edge bg-ok-soft px-5 py-3 text-sm font-semibold text-ok">
    <span class="h-2 w-2 rounded-full bg-current" aria-hidden="true"></span>
    All keys are working today.
  </p>

  <p v-else-if="activeKeys" class="flex items-center gap-2 rounded-md border border-rule bg-panel px-5 py-3 text-sm font-semibold text-muted">
    No verification yet today.
  </p>
</template>
