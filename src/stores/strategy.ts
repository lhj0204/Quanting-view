import { defineStore } from 'pinia'
import { ref } from 'vue'
import { getStrategies, getStrategy, createStrategy, updateStrategy, deleteStrategy, activateStrategy, runBacktest } from '../api'
import type { Strategy } from '../types'

export const useStrategyStore = defineStore('strategy', () => {
  const strategies = ref<Strategy[]>([])
  const current = ref<Strategy | null>(null)
  const loading = ref(false)

  async function fetchAll() {
    loading.value = true
    const { data } = await getStrategies()
    strategies.value = data
    loading.value = false
  }

  async function fetchOne(id: number) {
    const { data } = await getStrategy(id)
    current.value = data
    return data
  }

  async function create(s: Partial<Strategy>) {
    const { data } = await createStrategy(s)
    strategies.value.unshift(data)
    return data
  }

  async function update(id: number, s: Partial<Strategy>) {
    const { data } = await updateStrategy(id, s)
    const idx = strategies.value.findIndex(x => x.id === id)
    if (idx >= 0) strategies.value[idx] = data
    return data
  }

  async function remove(id: number) {
    await deleteStrategy(id)
    strategies.value = strategies.value.filter(x => x.id !== id)
  }

  async function activate(id: number) {
    const { data } = await activateStrategy(id)
    const idx = strategies.value.findIndex(x => x.id === id)
    if (idx >= 0) strategies.value[idx] = data
    return data
  }

  async function backtest(id: number) {
    return await runBacktest(id)
  }

  return { strategies, current, loading, fetchAll, fetchOne, create, update, remove, activate, backtest }
})
