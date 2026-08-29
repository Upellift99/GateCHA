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
         ends up being. -->
    <span
      class="relative block h-6 w-11 rounded-full bg-slate-200 transition-colors
             peer-checked:bg-teal-600 peer-focus:ring-2 peer-focus:ring-teal-500
             peer-focus:ring-offset-2 peer-disabled:opacity-50
             after:absolute after:top-1/2 after:left-0.5 after:h-5 after:w-5
             after:-translate-y-1/2 after:rounded-full after:border
             after:border-slate-300 after:bg-white after:shadow-sm
             after:transition-all after:content-['']
             peer-checked:after:translate-x-5 peer-checked:after:border-white"
    ></span>
  </span>
</template>
