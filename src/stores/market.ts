import { defineStore } from 'pinia'
import { ref } from 'vue'
import { getKlines, getTicker, getDepth } from '../api'
import type { Kline, Ticker, DepthData } from '../types'

export const useMarketStore = defineStore('market', () => {
  const klines = ref<Kline[]>([])
  const tickers = ref<Record<string, Ticker>>({})
  const depth = ref<Record<string, DepthData>>({})
  const loading = ref(false)

  async function fetchKlines(exchange: string, symbol: string, interval: string, limit = 500) {
    const { data } = await getKlines({ exchange, symbol, interval, limit })
    klines.value = data
    return data
  }

  async function fetchTicker(exchange: string, symbol: string) {
    const { data } = await getTicker({ exchange, symbol })
    tickers.value[symbol] = data
    return data
  }

  async function fetchTickers(exchange: string, symbols: string[]) {
    loading.value = true
    const results = await Promise.allSettled(
      symbols.map(s => getTicker({ exchange, symbol: s }))
    )
    results.forEach((r, i) => {
      if (r.status === 'fulfilled') {
        tickers.value[symbols[i]] = r.value.data
      }
    })
    loading.value = false
  }

  async function fetchDepth(exchange: string, symbol: string, limit = 20) {
    const { data } = await getDepth({ exchange, symbol, limit })
    depth.value[symbol] = data
    return data
  }

  return { klines, tickers, depth, loading, fetchKlines, fetchTicker, fetchTickers, fetchDepth }
})
