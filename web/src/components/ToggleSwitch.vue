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
       the checkbox is sr-only, so the pill is the only clickable thing and it has
       to sit inside that label to activate anything. See #146. -->
  <span class="inline-flex shrink-0 items-center">
    <input
      :id="id"
      type="checkbox"
      class="sr-only peer"
      :checked="modelValue"
      :disabled="disabled"
      @change="onChange"
    />
    <!-- The knob is positioned against this pill (hence `relative` here) and
         centred on its midline, so it stays centred whatever the pill's box
         ends up being.

         A 16px knob inset by 4px. The inset has to clear two antialiased curves
         at once, the pill's cap and the knob, and each eats about 1.5 physical
         pixels: measured against the rendered pixels, a 3px inset leaves 1px of
         solid teal at the resting end at 100% and none at 87.5%, so the knob
         reads as bursting through the cap. 4px leaves 2px and 1px. The gap above
         the knob is bounded by the pill's straight edge instead, which is why it
         looked fine there all along and the fault seemed to be a centring one.
         The travel is unchanged: 44 - 16 - 4 - 4 is still 20px. See #160,
         which fixed the centring and left this. -->
    <span
      class="relative block h-6 w-11 rounded-full bg-slate-200 transition-colors
             peer-checked:bg-teal-600 peer-focus-visible:ring-2
             peer-focus-visible:ring-teal-500 peer-disabled:opacity-50
             after:absolute after:top-1/2 after:left-1 after:h-4 after:w-4
             after:-translate-y-1/2 after:rounded-full after:border
             after:border-slate-300 after:bg-white after:shadow-sm
             after:transition-all after:content-['']
             peer-checked:after:translate-x-5 peer-checked:after:border-white"
    ></span>
  </span>
</template>
