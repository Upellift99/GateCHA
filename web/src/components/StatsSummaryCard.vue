<script setup lang="ts">
defineProps<{
  label: string
  value: number
  color: 'blue' | 'green' | 'red' | 'purple'
}>()

const colorClasses: Record<string, string> = {
  blue: 'bg-blue-50 text-blue-700',
  green: 'bg-green-50 text-green-700',
  red: 'bg-red-50 text-red-700',
  purple: 'bg-purple-50 text-purple-700',
}
</script>

<template>
  <div class="bg-white shadow rounded-lg p-6">
    <dl>
      <dt class="text-sm font-medium text-gray-500">{{ label }}</dt>
      <dd class="mt-2">
        <span :class="colorClasses[color]" class="text-3xl font-bold">
          {{ value.toLocaleString() }}
        </span>
      </dd>
    </dl>
  </div>
</template>
