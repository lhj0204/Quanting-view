<template>
  <div>
    <div class="flex justify-between items-center mb-6">
      <h2 class="text-xl font-bold">策略管理</h2>
      <router-link to="/strategies/new" class="btn btn-primary text-sm">创建策略</router-link>
    </div>

    <div v-if="strategies.length === 0" class="card text-center py-12">
      <p class="text-text-secondary mb-4">暂无策略</p>
      <router-link to="/strategies/new" class="btn btn-primary text-sm">创建第一个策略</router-link>
    </div>

    <div v-else class="grid gap-4">
      <div v-for="s in strategies" :key="s.id" class="card flex items-center justify-between">
        <div class="flex items-center gap-4">
          <div>
            <h3 class="font-medium">{{ s.name }}</h3>
            <div class="flex gap-2 mt-1">
              <span class="badge" :class="s.exchange === 'binance' ? 'badge-blue' : 'badge-green'">{{ s.exchange }}</span>
              <span class="badge" :class="s.trade_mode === 'paper' ? 'badge-blue' : 'badge-green'">{{ s.trade_mode === 'paper' ? '模拟' : '实盘' }}</span>
              <span class="badge" :class="statusBadge(s.status)">{{ statusLabel(s.status) }}</span>
            </div>
          </div>
        </div>
        <div class="flex gap-2">
          <router-link :to="`/strategies/${s.id}`" class="btn btn-primary text-xs px-3 py-1.5">编辑</router-link>
          <button v-if="s.status !== 'active'" @click="activate(s.id)" class="btn btn-success text-xs px-3 py-1.5">启动</button>
          <button @click="backtest(s.id)" class="btn text-xs px-3 py-1.5 border border-border hover:bg-bg">回测</button>
          <button @click="remove(s.id)" class="btn btn-danger text-xs px-3 py-1.5">删除</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useStrategyStore } from '../stores/strategy'

const store = useStrategyStore()
const strategies = computed(() => store.strategies)

onMounted(() => store.fetchAll())

function statusBadge(s: string) {
  return s === 'active' ? 'badge-green' : s === 'draft' ? 'badge-blue' : 'badge-red'
}
function statusLabel(s: string) {
  const m: Record<string, string> = { draft: '草稿', active: '运行中', paused: '已暂停', stopped: '已停止' }
  return m[s] || s
}
async function activate(id: number) { await store.activate(id); store.fetchAll() }
async function backtest(id: number) {
  await store.backtest(id)
  alert('回测完成，请在回测记录查看结果')
}
async function remove(id: number) {
  if (confirm('确定删除此策略？')) await store.remove(id)
}
</script>
