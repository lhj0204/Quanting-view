<template>
  <div class="card">
    <div class="text-xs text-text-secondary mb-1">{{ label }}</div>
    <div class="text-lg font-bold" :class="trendColor">{{ formatted }}</div>
    <div v-if="subtitle" class="text-xs text-text-secondary mt-0.5">{{ subtitle }}</div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  label: string
  value?: number | string
  trend?: 'up' | 'down' | 'neutral'
  subtitle?: string
  precision?: number
}>()

const formatted = computed(() => {
  const v = props.value
  if (v == null) return '-'
  if (typeof v === 'string') return v
  return v.toFixed(props.precision ?? 2)
})

const trendColor = computed(() => {
  if (props.trend === 'up') return 'text-green'
  if (props.trend === 'down') return 'text-red'
  return 'text-text'
})
</script>
