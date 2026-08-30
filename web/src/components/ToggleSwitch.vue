<script setup lang="ts">
withDefaults(
  defineProps<{
    id: string
    modelValue: boolean
    disabled?: boolean
  }>(),
  { disabled: false },
)

const emit = defineEmits<{ 'update:modelValue': [value: boolean] }>()

function onChange(event: Event) {
  emit('update:modelValue', (event.target as HTMLInputElement).checked)
}
</script>

<template>
  <!-- The caller wraps this in a label bound to the same id, carrying the text:
       the checkbox is sr-only, so the switch is the only clickable thing and it
       has to sit inside that label to activate anything. See #146. -->
  <span class="inline-flex shrink-0 items-center">
    <input
      :id="id"
      type="checkbox"
      class="sr-only peer"
      :checked="modelValue"
      :disabled="disabled"
      @change="onChange"
    />
    <!-- Drawn as one SVG rather than a box with an ::after knob, because a box
         and its pseudo-element snap to whole device pixels independently. The
         switch is 44x24 holding a 16px knob, so 8px of track splits either side
         of it, and wherever the page is not rendered at a whole device scale
         that 8px lands on an odd number of physical pixels and cannot split
         evenly. At 87.5%, where this was reported, the box version put 5px of
         track above the knob and 4px below. That is one physical pixel, and it
         is plainly visible.

         Reshaping the boxes only moves which scales are unlucky. Measured over
         14 scales from 0.8 to 2, in both states: the box version is uneven at
         0.85, 0.875, 0.9 and 1.1, and the best alternative geometry found by
         sweeping 63 of them was uneven at 8 of the 14. This SVG is uneven at
         0.8, 0.9, 1.05 and 2 when on, and 0.85, 0.95, 1.1 and 1.2 when off. So
         it is not uniformly better by count, and no fixed geometry is even
         everywhere. What it does is come out even at 87.5%, and at 1, 1.25 and
         1.5, which the box version could not do.

         The numbers still have to sum: a knob of r=8 at cx=12 and cx=32 leaves
         4 units of track on every side, and 20 units of travel. -->
    <svg
      class="block h-6 w-11 rounded-full peer-focus-visible:ring-2
             peer-focus-visible:ring-teal-500 peer-disabled:opacity-50"
      viewBox="0 0 44 24"
      aria-hidden="true"
      focusable="false"
    >
      <rect
        width="44"
        height="24"
        rx="12"
        class="transition-colors"
        :class="modelValue ? 'fill-teal-600' : 'fill-slate-200'"
      />
      <!-- No stroke on the knob. The old box carried a 1px slate border, but a
           stroke is centred on the path, so it pushes the antialiased edge half
           a unit out and reintroduces the uneven rounding this rewrite exists to
           remove: with one the gap measured 5/4 again at 87.5%, without it 4/4.
           White on slate-200 carries the off state on its own. -->
      <circle
        cx="12"
        cy="12"
        r="8"
        class="fill-white transition-transform"
        :class="{ 'translate-x-5': modelValue }"
      />
    </svg>
  </span>
</template>
