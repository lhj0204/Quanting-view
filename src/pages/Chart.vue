<template>
  <div>
    <div class="flex items-center gap-4 mb-4">
      <router-link to="/market" class="text-text-secondary hover:text-text">&larr; 行情</router-link>
      <h2 class="text-xl font-bold">{{ symbol }}</h2>
      <select v-model="activeExchange" class="input w-28 text-sm">
        <option value="binance">Binance</option>
        <option value="okx">OKX</option>
      </select>
      <select v-model="activeInterval" class="input w-24 text-sm" @change="loadKlines">
        <option v-for="i in intervals" :key="i" :value="i">{{ i }}</option>
      </select>
      <div class="flex gap-1">
        <button v-for="ind in activeIndicators" :key="ind" class="btn text-xs px-2 py-1 border border-border" :class="{ 'bg-blue/10 border-blue': true }">{{ ind }}</button>
        <button @click="showIndicators = !showIndicators" class="btn text-xs px-2 py-1 border border-border">+ 指标</button>
      </div>
    </div>

    <div class="flex gap-4">
      <div class="flex-1">
        <div ref="chartContainer" class="w-full bg-surface rounded border border-border" style="height: 500px"></div>
        <div ref="indicatorContainer" class="w-full" style="min-height: 150px"></div>
        <div ref="volumeContainer" class="w-full" style="height: 150px"></div>
      </div>
      <div class="w-72 space-y-4">
        <div class="card">
          <h3 class="text-sm font-medium mb-2">下单</h3>
          <form @submit.prevent="placeOrder" class="space-y-2">
            <div class="flex gap-2">
              <button type="button" @click="orderSide = 'buy'" class="flex-1 py-1.5 rounded text-sm font-medium" :class="orderSide === 'buy' ? 'bg-green text-black' : 'bg-surface border border-border text-text-secondary'">买入</button>
              <button type="button" @click="orderSide = 'sell'" class="flex-1 py-1.5 rounded text-sm font-medium" :class="orderSide === 'sell' ? 'bg-red text-white' : 'bg-surface border border-border text-text-secondary'">卖出</button>
            </div>
            <div>
              <label class="text-xs text-text-secondary">类型</label>
              <select v-model="orderType" class="input text-sm">
                <option value="market">市价单</option>
                <option value="limit">限价单</option>
              </select>
            </div>
            <div v-if="orderType === 'limit'">
              <label class="text-xs text-text-secondary">限价</label>
              <input v-model.number="orderPrice" type="number" step="0.01" class="input text-sm" />
            </div>
            <div>
              <label class="text-xs text-text-secondary">数量</label>
              <input v-model.number="orderQty" type="number" step="0.001" class="input text-sm" placeholder="0.001" />
            </div>
            <button type="submit" class="btn w-full text-sm" :class="orderSide === 'buy' ? 'bg-green text-black' : 'bg-red text-white'">
              {{ orderSide === 'buy' ? '买入' : '卖出' }} {{ symbol }}
            </button>
          </form>
        </div>

        <div class="card">
          <h3 class="text-sm font-medium mb-2">盘口 ({{ activeExchange }})</h3>
          <div class="text-xs">
            <div v-for="(a, i) in depthAsks" :key="'a'+i" class="flex justify-between text-red mb-0.5">
              <span>{{ a[0] }}</span><span>{{ a[1] }}</span>
            </div>
            <div class="text-center font-bold text-blue my-1 py-1 border-y border-border">
              {{ fmt(ticker?.price) }}
            </div>
            <div v-for="(b, i) in depthBids" :key="'b'+i" class="flex justify-between text-green mb-0.5">
              <span>{{ b[0] }}</span><span>{{ b[1] }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, nextTick, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useMarketStore } from '../stores/market'
import { useAccountStore } from '../stores/account'
import { createOrder } from '../api'
import { createChart, ColorType } from 'lightweight-charts'
import type { Kline, Ticker } from '../types'

const route = useRoute()
const marketStore = useMarketStore()
const accountStore = useAccountStore()

const symbol = computed(() => route.params.symbol as string || 'BTCUSDT')
const activeExchange = ref('binance')
const activeInterval = ref('1h')
const activeIndicators = ref(['MA', 'RSI'])
const showIndicators = ref(false)
const intervals = ['1m', '5m', '15m', '30m', '1h', '4h', '1d']

const chartContainer = ref<HTMLElement>()
const indicatorContainer = ref<HTMLElement>()
const volumeContainer = ref<HTMLElement>()

const ticker = ref<Ticker | null>(null)
const depthAsks = ref<[string, string][]>([])
const depthBids = ref<[string, string][]>([])

const orderSide = ref<'buy' | 'sell'>('buy')
const orderType = ref<'market' | 'limit'>('market')
const orderPrice = ref(0)
const orderQty = ref(0.001)

let mainChart: any = null
let volumeSeries: any = null
let candleSeries: any = null
let maSeriesList: any[] = []
let ws: WebSocket | null = null

onMounted(async () => {
  await loadKlines()
  await loadTicker()
  await loadDepth()
  connectWS()
})

onUnmounted(() => {
  if (ws) ws.close()
  if (mainChart) mainChart.remove()
})

watch([activeExchange], () => {
  loadKlines()
  loadTicker()
  loadDepth()
})

async function loadKlines() {
  const klines = await marketStore.fetchKlines(activeExchange.value, symbol.value, activeInterval.value, 500)
  await nextTick()
  renderChart(klines)
}

async function loadTicker() {
  ticker.value = await marketStore.fetchTicker(activeExchange.value, symbol.value)
}

async function loadDepth() {
  const depth = await marketStore.fetchDepth(activeExchange.value, symbol.value, 10)
  depthAsks.value = depth.asks.slice(0, 8).reverse()
  depthBids.value = depth.bids.slice(0, 8)
}

function renderChart(klines: Kline[]) {
  if (mainChart) mainChart.remove()

  const chartOptions: any = {
    layout: {
      background: { type: ColorType.Solid, color: '#1a1d27' },
      textColor: '#787b86',
    },
    grid: {
      vertLines: { color: '#2a2d37' },
      horzLines: { color: '#2a2d37' },
    },
    crosshair: { mode: 0 },
    rightPriceScale: { borderColor: '#2a2d37' },
    timeScale: {
      borderColor: '#2a2d37',
      timeVisible: true,
    },
    width: chartContainer.value?.clientWidth || 800,
    height: 500,
  }

  mainChart = createChart(chartContainer.value!, chartOptions)
  candleSeries = mainChart.addCandlestickSeries({
    upColor: '#00bcd4', downColor: '#ef5350',
    borderUpColor: '#00bcd4', borderDownColor: '#ef5350',
    wickUpColor: '#00bcd4', wickDownColor: '#ef5350',
  })

  const candleData = klines.map(k => ({
    time: k.time as any,
    open: k.open, high: k.high, low: k.low, close: k.close,
  }))
  candleSeries.setData(candleData)

  // Volume series
  volumeSeries = mainChart.addHistogramSeries({
    color: '#26a69a33',
    priceFormat: { type: 'volume' },
    priceScaleId: '',
  })
  mainChart.priceScale('').applyOptions({
    scaleMargins: { top: 0.8, bottom: 0 },
  })
  const volumeData = klines.map(k => ({
    time: k.time as any,
    value: k.volume,
    color: k.close >= k.open ? '#00bcd433' : '#ef535033',
  }))
  volumeSeries.setData(volumeData)

  // MA indicators
  maSeriesList = []
  const ma7 = calcMA(klines, 7)
  const ma25 = calcMA(klines, 25)
  ;[ma7, ma25].forEach((maData, idx) => {
    const color = idx === 0 ? '#ffca28' : '#ef5350'
    const series = mainChart.addLineSeries({ color, lineWidth: 1 })
    const lineData = maData.map((v, i) => v ? { time: klines[i].time as any, value: v } : null).filter(Boolean)
    series.setData(lineData)
    maSeriesList.push(series)
  })

  // RSI sub-chart (simplified placed in main chart area)
  mainChart.timeScale().fitContent()
}

function calcMA(data: Kline[], period: number) {
  const result: (number | null)[] = new Array(data.length).fill(null)
  for (let i = period - 1; i < data.length; i++) {
    let sum = 0
    for (let j = i - period + 1; j <= i; j++) sum += data[j].close
    result[i] = sum / period
  }
  return result
}

function connectWS() {
  const proto = location.protocol === 'https:' ? 'wss:' : 'ws:'
  ws = new WebSocket(`${proto}//${location.host}/ws`)

  ws.onopen = () => {
    ws!.send(JSON.stringify({
      event: 'subscribe', channel: 'kline',
      symbol: symbol.value, interval: activeInterval.value
    }))
    ws!.send(JSON.stringify({
      event: 'subscribe', channel: 'ticker',
      symbol: symbol.value
    }))
  }

  ws.onmessage = (evt) => {
    const msg = JSON.parse(evt.data)
    if (msg.channel === 'kline' && msg.data && candleSeries) {
      candleSeries.update({
        time: msg.data.time as any,
        open: msg.data.open, high: msg.data.high,
        low: msg.data.low, close: msg.data.close,
      })
    }
    if (msg.channel === 'ticker' && msg.data) {
      ticker.value = { ...ticker.value, ...msg.data }
    }
  }
}

async function placeOrder() {
  try {
    await createOrder({
      symbol: symbol.value,
      exchange: activeExchange.value,
      side: orderSide.value,
      type: orderType.value,
      price: orderType.value === 'limit' ? orderPrice.value : 0,
      quantity: orderQty.value,
      mode: 'paper',
    })
    alert('下单成功')
    accountStore.fetchSummary()
  } catch (e: any) {
    alert('下单失败: ' + (e?.response?.data?.error || e.message))
  }
}

function fmt(v?: number) { return v?.toFixed(2) ?? '-' }
</script>
