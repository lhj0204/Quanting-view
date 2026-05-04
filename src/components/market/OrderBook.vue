<template>
  <div class="card">
    <div class="flex items-center justify-between mb-3">
      <h3 class="text-sm font-medium">订单簿</h3>
      <select v-model="localExchange" class="bg-bg border border-border rounded px-2 py-1 text-xs text-text" @change="$emit('update:exchange', localExchange)">
        <option value="binance">Binance</option>
        <option value="okx">OKX</option>
      </select>
    </div>

    <div class="text-xs">
      <div class="grid grid-cols-3 text-text-secondary mb-1 px-1">
        <span>价格</span>
        <span class="text-right">数量</span>
        <span class="text-right">累计</span>
      </div>

      <div class="space-y-0.5 mb-2">
        <div v-for="(row, i) in askRows" :key="'a'+i" class="grid grid-cols-3 px-1 rounded-sm hover:bg-bg relative">
          <div class="absolute inset-y-0 right-0 bg-red/10" :style="{ width: row.pct + '%' }"></div>
          <span class="relative z-10 text-red">{{ fmtPrice(row.price) }}</span>
          <span class="relative z-10 text-right text-text-secondary">{{ fmtQty(row.qty) }}</span>
          <span class="relative z-10 text-right text-text-secondary">{{ fmtQty(row.cumQty) }}</span>
        </div>
      </div>

      <div class="text-center font-bold py-2 my-2 border-y border-border text-lg" :class="lastPrice >= 0 ? 'text-green' : 'text-red'">
        {{ fmtPrice(lastPrice) }}
      </div>

      <div class="space-y-0.5">
        <div v-for="(row, i) in bidRows" :key="'b'+i" class="grid grid-cols-3 px-1 rounded-sm hover:bg-bg relative">
          <div class="absolute inset-y-0 right-0 bg-green/10" :style="{ width: row.pct + '%' }"></div>
          <span class="relative z-10 text-green">{{ fmtPrice(row.price) }}</span>
          <span class="relative z-10 text-right text-text-secondary">{{ fmtQty(row.qty) }}</span>
          <span class="relative z-10 text-right text-text-secondary">{{ fmtQty(row.cumQty) }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted, computed } from 'vue'
import { getDepth } from '../../api'

const props = defineProps<{ symbol: string; exchange?: string }>()
defineEmits<{ 'update:exchange': [value: string] }>()

const localExchange = ref(props.exchange || 'binance')
const asks = ref<[string, string][]>([])
const bids = ref<[string, string][]>([])
const lastPrice = ref(0)

watch(() => props.symbol, fetchDepth)
watch(localExchange, fetchDepth)
onMounted(fetchDepth)

async function fetchDepth() {
  if (!props.symbol) return
  try {
    const { data } = await getDepth({ exchange: localExchange.value, symbol: props.symbol, limit: 20 })
    asks.value = data.asks.slice(0, 10).reverse()
    bids.value = data.bids.slice(0, 10)
    if (bids.value.length > 0) {
      lastPrice.value = parseFloat(bids.value[0][0])
    }
  } catch {}
}

function maxQty(rows: [string, string][]) {
  return Math.max(...rows.map(r => parseFloat(r[1])), 1)
}

const askRows = computed(() => {
  let cum = 0
  const maxQ = maxQty(asks.value)
  return asks.value.map(r => {
    cum += parseFloat(r[1])
    return { price: parseFloat(r[0]), qty: parseFloat(r[1]), cumQty: cum, pct: (parseFloat(r[1]) / maxQ * 100) }
  })
})

const bidRows = computed(() => {
  let cum = 0
  const maxQ = maxQty(bids.value)
  return bids.value.map(r => {
    cum += parseFloat(r[1])
    return { price: parseFloat(r[0]), qty: parseFloat(r[1]), cumQty: cum, pct: (parseFloat(r[1]) / maxQ * 100) }
  })
})

function fmtPrice(v: number) { return v?.toFixed(v < 1 ? 6 : 2) ?? '-' }
function fmtQty(v: number) { return v >= 1000 ? (v/1000).toFixed(1)+'K' : v.toFixed(v < 1 ? 4 : 2) }
</script>
