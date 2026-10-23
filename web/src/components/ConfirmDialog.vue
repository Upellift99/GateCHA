<script setup lang="ts">
import { nextTick, ref, useId, watch } from 'vue'

const props = withDefaults(
  defineProps<{
    open: boolean
    title: string
    confirmLabel?: string
    cancelLabel?: string
    tone?: 'danger' | 'primary'
    busy?: boolean
  }>(),
  { confirmLabel: 'Confirm', cancelLabel: 'Cancel', tone: 'primary', busy: false },
)

const emit = defineEmits<{ confirm: []; cancel: [] }>()

const titleId = useId()
const dialog = ref<HTMLDialogElement | null>(null)
const cancelButton = ref<HTMLButtonElement | null>(null)

// showModal() is what makes this modal: it lifts the dialog into the top layer,
// keeps Tab inside it, makes the rest of the page inert to assistive technology,
// routes Escape to the cancel event and hands focus back on close. None of that
// is reimplemented here.
watch(
  () => props.open,
  async (open) => {
    // Waits for the panel below to render, so there is something to focus, and
    // covers a caller that mounts this already open.
    await nextTick()
    if (open) {
      if (!dialog.value?.open) dialog.value?.showModal()
      // Browsers focus the first focusable child on their own. Doing it here is
      // free and makes the choice explicit rather than positional.
      cancelButton.value?.focus()
    } else if (dialog.value?.open) {
      dialog.value.close()
    }
  },
  { immediate: true },
)

// Escape raises cancel. Suppressing the default close leaves the caller's `open`
// prop the single source of truth: it closes us back through the watch above.
function onCancel(event: Event) {
  event.preventDefault()
  emit('cancel')
}

// A click on the backdrop is dispatched to the dialog element itself. Anything
// inside the panel targets the panel's own markup, so it is left alone.
function onBackdropClick(event: MouseEvent) {
  if (event.target === dialog.value) emit('cancel')
}
</script>

<template>
  <dialog
    ref="dialog"
    :aria-labelledby="titleId"
    class="m-auto w-[calc(100%-2rem)] max-w-md rounded-md bg-panel text-ink shadow-xl
           backdrop:bg-backdrop backdrop:backdrop-blur-sm"
    @cancel="onCancel"
    @click="onBackdropClick"
  >
    <div v-if="open" class="p-6">
      <div class="flex gap-4">
        <span
          :class="tone === 'danger' ? 'bg-fail-soft text-fail' : 'bg-sunken text-link'"
          class="flex h-10 w-10 shrink-0 items-center justify-center rounded-full"
          aria-hidden="true"
        >
          <svg class="h-5 w-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path
              v-if="tone === 'danger'"
              stroke-linecap="round"
              stroke-linejoin="round"
              d="M12 9v4m0 4h.01M10.29 3.86 1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0Z"
            />
            <path v-else stroke-linecap="round" stroke-linejoin="round" d="M12 8v5m0 3h.01M12 21a9 9 0 1 0 0-18 9 9 0 0 0 0 18Z" />
          </svg>
        </span>
        <div class="min-w-0 flex-1">
          <h3 :id="titleId" class="text-[17px] font-extrabold text-ink">{{ title }}</h3>
          <div class="mt-1 text-sm text-muted">
            <slot />
          </div>
        </div>
      </div>

      <div class="mt-6 flex justify-end gap-2">
        <button
          ref="cancelButton"
          type="button"
          class="rounded-md bg-sunken px-4 py-2 text-sm font-medium text-ink-2 hover:bg-sunken focus:outline-none focus:ring-2 focus:ring-focus"
          @click="emit('cancel')"
        >
          {{ cancelLabel }}
        </button>
        <button
          type="button"
          :disabled="busy"
          :class="
            tone === 'danger'
              ? 'bg-danger hover:bg-danger-hover focus:ring-fail'
              : 'bg-action hover:bg-action-hover focus:ring-focus'
          "
          class="rounded-md px-4 py-2 text-sm font-medium text-action-ink focus:outline-none focus:ring-2 disabled:opacity-50"
          @click="emit('confirm')"
        >
          {{ confirmLabel }}
        </button>
      </div>
    </div>
  </dialog>
</template>
