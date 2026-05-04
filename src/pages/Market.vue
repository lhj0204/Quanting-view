<template>
  <div>
    <h2 class="text-xl font-bold mb-6">行情</h2>

    <div class="flex gap-4 mb-4">
      <select v-model="exchange" class="input w-32">
        <option value="binance">Binance</option>
        <option value="okx">OKX</option>
      </select>
      <input v-model="search" class="input w-64" placeholder="搜索币种..." />
      <button @click="refreshTickers" class="btn btn-primary text-sm" :disabled="loading">刷新</button>
    </div>

    <div class="card">
      <div v-if="loading" class="text-text-secondary text-sm py-8 text-center">加载中...</div>
      <table v-else class="w-full text-sm">
        <thead>
          <tr class="text-text-secondary text-left">
            <th class="pb-3">交易对</th>
            <th class="pb-3 text-right">最新价</th>
            <th class="pb-3 text-right">24h 涨跌</th>
            <th class="pb-3 text-right">24h 最高</th>
            <th class="pb-3 text-right">24h 最低</th>
            <th class="pb-3 text-right">24h 成交量</th>
            <th class="pb-3"></th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="t in filteredTickers" :key="t.symbol" class="border-t border-border hover:bg-bg/50 cursor-pointer" @click="$router.push(`/market/${t.symbol}`)">
            <td class="py-3 font-medium">{{ t.symbol }}</td>
            <td class="py-3 text-right font-mono">{{ fmt(t.price) }}</td>
            <td class="py-3 text-right" :class="t.change_pct >= 0 ? 'text-green' : 'text-red'">
              {{ t.change_pct >= 0 ? '+' : '' }}{{ fmt(t.change_pct) }}%
            </td>
            <td class="py-3 text-right font-mono">{{ fmt(t.high_24h) }}</td>
            <td class="py-3 text-right font-mono">{{ fmt(t.low_24h) }}</td>
            <td class="py-3 text-right font-mono">{{ fmtVol(t.volume_24h) }}</td>
            <td class="py-3 text-right">
              <button class="btn btn-primary text-xs px-3 py-1">图表</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useMarketStore } from '../stores/market'

const marketStore = useMarketStore()
const exchange = ref('binance')
const search = ref('')
const loading = ref(false)

const tickerList = computed(() => Object.values(marketStore.tickers))
const filteredTickers = computed(() =>
  tickerList.value.filter(t => t.symbol.toLowerCase().includes(search.value.toLowerCase()))
)

const defaultSymbols = [
  'BTCUSDT', 'ETHUSDT', 'BNBUSDT', 'SOLUSDT', 'XRPUSDT', 'ADAUSDT',
  'DOGEUSDT', 'AVAXUSDT', 'DOTUSDT', 'LINKUSDT', 'MATICUSDT', 'UNIUSDT'
]

onMounted(() => refreshTickers())

async function refreshTickers() {
  loading.value = true
  await marketStore.fetchTickers(exchange.value, defaultSymbols)
  loading.value = false
}

function fmt(v: number) { return v?.toFixed(2) ?? '-' }
function fmtVol(v: number) {
  if (!v) return '-'
  if (v > 1e9) return (v / 1e9).toFixed(2) + 'B'
  if (v > 1e6) return (v / 1e6).toFixed(2) + 'M'
  return v.toFixed(0)
}
</script>
