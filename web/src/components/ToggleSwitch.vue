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

         An 18px knob inset by 3px rather than a 20px one inset by 2px: at a
         browser zoom of 87.5% a 2px gap is 1.75 physical pixels, which the
         antialiasing of two curved edges eats entirely, and the knob reads as
         bulging out of the pill. 3px survives it. The travel is unchanged:
         44 - 18 - 3 - 3 is still 20px. -->
    <span
      class="relative block h-6 w-11 rounded-full bg-slate-200 transition-colors
             peer-checked:bg-teal-600 peer-focus-visible:ring-2
             peer-focus-visible:ring-teal-500 peer-disabled:opacity-50
             after:absolute after:top-1/2 after:left-0.75 after:h-4.5 after:w-4.5
             after:-translate-y-1/2 after:rounded-full after:border
             after:border-slate-300 after:bg-white after:shadow-sm
             after:transition-all after:content-['']
             peer-checked:after:translate-x-5 peer-checked:after:border-white"
    ></span>
  </span>
</template>
