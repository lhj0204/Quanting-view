<template>
  <div>
    <div class="flex items-center gap-4 mb-6">
      <router-link to="/backtests" class="text-text-secondary hover:text-text">&larr; 返回</router-link>
      <h2 class="text-xl font-bold">回测详情 #{{ bt?.id }}</h2>
    </div>

    <div v-if="bt" class="space-y-6">
      <div class="grid grid-cols-4 gap-4">
        <div class="card"><div class="text-xs text-text-secondary">初始资金</div><div class="text-lg font-bold">{{ fmt(bt.initial_capital) }}</div></div>
        <div class="card"><div class="text-xs text-text-secondary">最终资金</div><div class="text-lg font-bold">{{ fmt(bt.final_capital) }}</div></div>
        <div class="card">
          <div class="text-xs text-text-secondary">收益率</div>
          <div class="text-lg font-bold" :class="bt.final_capital >= bt.initial_capital ? 'text-green' : 'text-red'">
            {{ ((bt.final_capital - bt.initial_capital) / bt.initial_capital * 100).toFixed(2) }}%
          </div>
        </div>
        <div class="card"><div class="text-xs text-text-secondary">交易次数</div><div class="text-lg font-bold">{{ bt.total_trades }}</div></div>
        <div class="card"><div class="text-xs text-text-secondary">胜率</div><div class="text-lg font-bold">{{ bt.win_rate.toFixed(2) }}%</div></div>
        <div class="card"><div class="text-xs text-text-secondary">夏普比率</div><div class="text-lg font-bold">{{ bt.sharpe.toFixed(3) }}</div></div>
        <div class="card"><div class="text-xs text-text-secondary">最大回撤</div><div class="text-lg font-bold text-red">{{ bt.max_drawdown.toFixed(2) }}%</div></div>
        <div class="card"><div class="text-xs text-text-secondary">交易对/周期</div><div class="text-lg font-bold">{{ bt.symbol }} {{ bt.interval }}</div></div>
      </div>

      <div class="card">
        <h3 class="text-sm font-medium mb-3">权益曲线</h3>
        <div ref="chartEl" style="width: 100%; height: 350px"></div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, nextTick } from 'vue'
import { useRoute } from 'vue-router'
import { getBacktest } from '../api'
import type { Backtest, EquityPoint } from '../types'
import { createChart, ColorType } from 'lightweight-charts'

const route = useRoute()
const bt = ref<Backtest | null>(null)
const chartEl = ref<HTMLElement>()

onMounted(async () => {
  const id = Number(route.params.id)
  const { data } = await getBacktest(id)
  bt.value = data
  await nextTick()
  renderEquityChart()
})

function renderEquityChart() {
  if (!bt.value || !chartEl.value) return
  const equity: EquityPoint[] = JSON.parse(bt.value.equity_curve_json || '[]')

  const chart = createChart(chartEl.value, {
    layout: {
      background: { type: ColorType.Solid, color: '#1a1d27' },
      textColor: '#787b86',
    },
    grid: { vertLines: { color: '#2a2d37' }, horzLines: { color: '#2a2d37' } },
    width: chartEl.value.clientWidth,
    height: 350,
    timeScale: { timeVisible: true },
    rightPriceScale: { borderColor: '#2a2d37' },
  })

  const lineSeries = chart.addLineSeries({ color: '#00bcd4', lineWidth: 2 })
  const data = equity.map(e => ({ time: e.time as any, value: e.equity }))
  lineSeries.setData(data)
  chart.timeScale().fitContent()
}

function fmt(v: number) { return v?.toFixed(2) ?? '-' }
</script>
