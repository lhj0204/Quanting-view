<template>
  <div class="bg-surface border-b border-border overflow-hidden">
    <div class="flex gap-6 py-2 px-4 animate-scroll" v-if="tickerList.length > 0">
      <div v-for="t in tickerList" :key="t.symbol" class="flex items-center gap-2 text-sm whitespace-nowrap shrink-0 cursor-pointer hover:text-blue" @click="$router.push(`/market/${t.symbol}`)">
        <span class="font-medium">{{ t.symbol }}</span>
        <span class="font-mono">{{ fmt(t.price) }}</span>
        <span class="text-xs" :class="t.change_pct >= 0 ? 'text-green' : 'text-red'">
          {{ t.change_pct >= 0 ? '+' : '' }}{{ t.change_pct.toFixed(2) }}%
        </span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted } from 'vue'
import { useMarketStore } from '../../stores/market'
import type { Ticker } from '../../types'

const marketStore = useMarketStore()
const tickerList = computed<Ticker[]>(() => Object.values(marketStore.tickers))

const symbols = ['BTCUSDT', 'ETHUSDT', 'BNBUSDT', 'SOLUSDT', 'XRPUSDT', 'ADAUSDT', 'DOGEUSDT', 'AVAXUSDT']

let timer: number
onMounted(() => {
  marketStore.fetchTickers('binance', symbols)
  timer = window.setInterval(() => marketStore.fetchTickers('binance', symbols), 30000)
})
onUnmounted(() => clearInterval(timer))

function fmt(v: number) { return v?.toFixed(2) ?? '-' }
</script>

<style scoped>
@keyframes scroll {
  0% { transform: translateX(0); }
  100% { transform: translateX(-50%); }
}
.animate-scroll {
  animation: scroll 30s linear infinite;
}
.animate-scroll:hover {
  animation-play-state: paused;
}
</style>
