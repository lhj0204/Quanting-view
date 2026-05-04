<template>
  <div>
    <h2 class="text-xl font-bold mb-6">订单管理</h2>

    <div class="flex gap-4 mb-4">
      <button @click="filterStatus = ''" class="btn text-sm" :class="filterStatus === '' ? 'btn-primary' : 'border border-border'">全部</button>
      <button @click="filterStatus = 'open'" class="btn text-sm" :class="filterStatus === 'open' ? 'btn-primary' : 'border border-border'">未成交</button>
      <button @click="filterStatus = 'filled'" class="btn text-sm" :class="filterStatus === 'filled' ? 'btn-primary' : 'border border-border'">已成交</button>
      <button @click="filterStatus = 'cancelled'" class="btn text-sm" :class="filterStatus === 'cancelled' ? 'btn-primary' : 'border border-border'">已取消</button>
      <button @click="refresh" class="btn text-sm border border-border">刷新</button>
    </div>

    <div class="card">
      <div v-if="filteredOrders.length === 0" class="text-text-secondary text-sm py-8 text-center">暂无订单</div>
      <table v-else class="w-full text-sm">
        <thead>
          <tr class="text-text-secondary text-left">
            <th class="pb-3">时间</th>
            <th class="pb-3">交易所</th>
            <th class="pb-3">交易对</th>
            <th class="pb-3">方向</th>
            <th class="pb-3">类型</th>
            <th class="pb-3 text-right">价格</th>
            <th class="pb-3 text-right">数量</th>
            <th class="pb-3 text-right">已成交</th>
            <th class="pb-3">模式</th>
            <th class="pb-3">状态</th>
            <th class="pb-3">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="o in filteredOrders" :key="o.id" class="border-t border-border">
            <td class="py-2 text-xs text-text-secondary">{{ formatTime(o.created_at) }}</td>
            <td class="py-2"><span class="badge" :class="o.exchange === 'binance' ? 'badge-blue' : 'badge-green'">{{ o.exchange }}</span></td>
            <td class="py-2">{{ o.symbol }}</td>
            <td class="py-2"><span :class="o.side === 'buy' ? 'text-green' : 'text-red'">{{ o.side === 'buy' ? '买入' : '卖出' }}</span></td>
            <td class="py-2">{{ o.type === 'market' ? '市价' : '限价' }}</td>
            <td class="py-2 text-right">{{ o.price || '-' }}</td>
            <td class="py-2 text-right">{{ o.quantity }}</td>
            <td class="py-2 text-right">{{ o.filled_qty }}</td>
            <td class="py-2"><span class="badge" :class="o.mode === 'paper' ? 'badge-blue' : 'badge-green'">{{ o.mode === 'paper' ? '模拟' : '实盘' }}</span></td>
            <td class="py-2"><span class="badge" :class="statusBadge(o.status)">{{ statusLabel(o.status) }}</span></td>
            <td class="py-2">
              <button v-if="o.status === 'open'" @click="cancel(o.id)" class="text-red text-xs hover:underline">取消</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useAccountStore } from '../stores/account'
import { cancelOrder as cancelOrderApi } from '../api'

const store = useAccountStore()
const filterStatus = ref('')

const filteredOrders = computed(() => {
  const orders = store.orders
  if (!filterStatus.value) return orders
  return orders.filter(o => o.status === filterStatus.value)
})

onMounted(() => store.fetchOrders())

async function refresh() { await store.fetchOrders(filterStatus.value || undefined) }

async function cancel(id: number) {
  await cancelOrderApi(id)
  await store.fetchOrders()
}

function statusBadge(s: string) {
  return s === 'filled' ? 'badge-green' : s === 'open' ? 'badge-blue' : 'badge-red'
}
function statusLabel(s: string) {
  const m: Record<string, string> = { open: '未成交', filled: '已成交', cancelled: '已取消' }
  return m[s] || s
}
function formatTime(t: string) { return new Date(t).toLocaleString('zh-CN') }
</script>
