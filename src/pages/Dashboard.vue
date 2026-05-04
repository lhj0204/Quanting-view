<template>
  <div>
    <h2 class="text-xl font-bold mb-6">仪表盘</h2>

    <!-- Account Summary Cards -->
    <div class="grid grid-cols-4 gap-4 mb-6">
      <div class="card">
        <div class="text-xs text-text-secondary mb-1">账户余额</div>
        <div class="text-2xl font-bold">{{ fmt(account?.balance) }}</div>
        <div class="text-xs text-text-secondary mt-1">{{ account?.currency }}</div>
      </div>
      <div class="card">
        <div class="text-xs text-text-secondary mb-1">总盈亏</div>
        <div class="text-2xl font-bold" :class="pnlColor">{{ fmt(account?.total_pnl) }}</div>
        <div class="text-xs mt-1" :class="pnlColor">{{ fmtPct(account?.total_pnl_pct) }}%</div>
      </div>
      <div class="card">
        <div class="text-xs text-text-secondary mb-1">持仓数量</div>
        <div class="text-2xl font-bold">{{ positions.length }}</div>
      </div>
      <div class="card">
        <div class="text-xs text-text-secondary mb-1">活跃策略</div>
        <div class="text-2xl font-bold">{{ activeStrategies }}</div>
      </div>
    </div>

    <!-- Recent Trades + Positions -->
    <div class="grid grid-cols-2 gap-6">
      <div class="card">
        <h3 class="text-sm font-semibold mb-3">最近交易</h3>
        <div v-if="recentTrades.length === 0" class="text-text-secondary text-sm py-4 text-center">暂无交易记录</div>
        <table v-else class="w-full text-sm">
          <thead>
            <tr class="text-text-secondary text-left">
              <th class="pb-2">时间</th>
              <th class="pb-2">交易对</th>
              <th class="pb-2">方向</th>
              <th class="pb-2 text-right">价格</th>
              <th class="pb-2 text-right">盈亏</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="t in recentTrades" :key="t.id" class="border-t border-border">
              <td class="py-2 text-text-secondary text-xs">{{ formatTime(t.created_at) }}</td>
              <td class="py-2">{{ t.symbol }}</td>
              <td class="py-2">
                <span :class="t.side === 'buy' ? 'text-green' : 'text-red'">{{ t.side === 'buy' ? '买入' : '卖出' }}</span>
              </td>
              <td class="py-2 text-right">{{ fmt(t.price) }}</td>
              <td class="py-2 text-right" :class="t.realized_pnl >= 0 ? 'text-green' : 'text-red'">{{ fmt(t.realized_pnl) }}</td>
            </tr>
          </tbody>
        </table>
      </div>

      <div class="card">
        <h3 class="text-sm font-semibold mb-3">当前持仓</h3>
        <div v-if="positions.length === 0" class="text-text-secondary text-sm py-4 text-center">暂无持仓</div>
        <table v-else class="w-full text-sm">
          <thead>
            <tr class="text-text-secondary text-left">
              <th class="pb-2">币种</th>
              <th class="pb-2 text-right">数量</th>
              <th class="pb-2 text-right">均价</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="p in positions" :key="p.id" class="border-t border-border">
              <td class="py-2">{{ p.symbol }}</td>
              <td class="py-2 text-right">{{ p.quantity }}</td>
              <td class="py-2 text-right">{{ fmt(p.avg_entry_price) }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useAccountStore } from '../stores/account'
import { useStrategyStore } from '../stores/strategy'

const accountStore = useAccountStore()
const strategyStore = useStrategyStore()

const account = computed(() => accountStore.summary)
const positions = computed(() => accountStore.positions)
const recentTrades = computed(() => accountStore.trades.slice(0, 10))
const activeStrategies = computed(() => strategyStore.strategies.filter(s => s.status === 'active').length)

const pnlColor = computed(() => {
  const pnl = account.value?.total_pnl ?? 0
  return pnl >= 0 ? 'text-green' : 'text-red'
})

function fmt(v?: number) {
  if (v == null) return '-'
  return v.toFixed(2)
}
function fmtPct(v?: number) {
  if (v == null) return '-'
  return (v >= 0 ? '+' : '') + v.toFixed(2)
}
function formatTime(t: string) {
  return new Date(t).toLocaleString('zh-CN')
}

onMounted(async () => {
  await Promise.all([
    accountStore.fetchSummary(),
    accountStore.fetchTrades(),
    accountStore.fetchPositions(),
    strategyStore.fetchAll(),
  ])
})
</script>
