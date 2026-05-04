import { defineStore } from 'pinia'
import { ref } from 'vue'
import { getAccountSummary, getOrders, getTrades, getPositions } from '../api'
import type { AccountSummary, Order, Trade, Position } from '../types'

export const useAccountStore = defineStore('account', () => {
  const summary = ref<AccountSummary | null>(null)
  const orders = ref<Order[]>([])
  const trades = ref<Trade[]>([])
  const positions = ref<Position[]>([])

  async function fetchSummary() {
    const { data } = await getAccountSummary()
    summary.value = data
  }

  async function fetchOrders(status?: string, mode?: string) {
    const { data } = await getOrders({ status, mode })
    orders.value = data
  }

  async function fetchTrades() {
    const { data } = await getTrades()
    trades.value = data
  }

  async function fetchPositions(strategyId?: number) {
    const { data } = await getPositions(strategyId)
    positions.value = data
  }

  return { summary, orders, trades, positions, fetchSummary, fetchOrders, fetchTrades, fetchPositions }
})
