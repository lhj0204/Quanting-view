<template>
  <div>
    <h2 class="text-xl font-bold mb-6">回测记录</h2>

    <div v-if="backtests.length === 0" class="card text-center py-12">
      <p class="text-text-secondary">暂无回测记录，请先在策略管理页面执行回测</p>
    </div>

    <div class="card" v-else>
      <table class="w-full text-sm">
        <thead>
          <tr class="text-text-secondary text-left">
            <th class="pb-3">策略ID</th>
            <th class="pb-3">交易对</th>
            <th class="pb-3">周期</th>
            <th class="pb-3 text-right">初始资金</th>
            <th class="pb-3 text-right">最终资金</th>
            <th class="pb-3 text-right">收益率</th>
            <th class="pb-3 text-right">交易次数</th>
            <th class="pb-3 text-right">胜率</th>
            <th class="pb-3 text-right">夏普比率</th>
            <th class="pb-3 text-right">最大回撤</th>
            <th class="pb-3">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="bt in backtests" :key="bt.id" class="border-t border-border hover:bg-bg/50">
            <td class="py-3">{{ bt.strategy_id }}</td>
            <td class="py-3">{{ bt.symbol }}</td>
            <td class="py-3">{{ bt.interval }}</td>
            <td class="py-3 text-right">{{ fmt(bt.initial_capital) }}</td>
            <td class="py-3 text-right">{{ fmt(bt.final_capital) }}</td>
            <td class="py-3 text-right" :class="bt.final_capital >= bt.initial_capital ? 'text-green' : 'text-red'">
              {{ ((bt.final_capital - bt.initial_capital) / bt.initial_capital * 100).toFixed(2) }}%
            </td>
            <td class="py-3 text-right">{{ bt.total_trades }}</td>
            <td class="py-3 text-right">{{ fmtPct(bt.win_rate) }}%</td>
            <td class="py-3 text-right">{{ bt.sharpe.toFixed(3) }}</td>
            <td class="py-3 text-right text-red">{{ fmtPct(bt.max_drawdown) }}%</td>
            <td class="py-3">
              <router-link :to="`/backtests/${bt.id}`" class="btn btn-primary text-xs px-3 py-1">详情</router-link>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getBacktests } from '../api'
import type { Backtest } from '../types'

const backtests = ref<Backtest[]>([])

onMounted(async () => {
  const { data } = await getBacktests()
  backtests.value = data
})

function fmt(v: number) { return v?.toFixed(2) ?? '-' }
function fmtPct(v: number) { return v?.toFixed(2) ?? '-' }
</script>
