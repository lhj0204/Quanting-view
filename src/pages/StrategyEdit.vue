<template>
  <div>
    <div class="flex items-center gap-4 mb-6">
      <router-link to="/strategies" class="text-text-secondary hover:text-text">&larr; 返回</router-link>
      <h2 class="text-xl font-bold">{{ isNew ? '创建策略' : '编辑策略' }}</h2>
    </div>

    <div class="card max-w-2xl">
      <form @submit.prevent="handleSave" class="space-y-4">
        <div>
          <label class="text-xs text-text-secondary">策略名称</label>
          <input v-model="form.name" class="input" placeholder="如: 双均线策略" required />
        </div>
        <div class="grid grid-cols-2 gap-4">
          <div>
            <label class="text-xs text-text-secondary">交易所</label>
            <select v-model="form.exchange" class="input">
              <option value="binance">Binance</option>
              <option value="okx">OKX</option>
            </select>
          </div>
          <div>
            <label class="text-xs text-text-secondary">模式</label>
            <select v-model="form.trade_mode" class="input">
              <option value="paper">模拟盘</option>
              <option value="live">实盘</option>
            </select>
          </div>
        </div>

        <div>
          <label class="text-xs text-text-secondary">策略配置 (JSON)</label>
          <textarea v-model="form.config_json" class="input h-48 font-mono text-xs" placeholder='{"symbol":"BTCUSDT","interval":"1h","indicators":[...],"entry_rule":{...},"exit_rule":{...}}'></textarea>
          <p class="text-xs text-text-secondary mt-1">参考示例格式配置交易对、时间周期、指标和入场/出场条件</p>
        </div>

        <div class="bg-bg border border-border rounded p-3 text-xs text-text-secondary">
          <p class="font-medium mb-1">配置示例:</p>
          <pre class="whitespace-pre-wrap">{{ exampleConfig }}</pre>
        </div>

        <div class="flex gap-3">
          <button type="submit" class="btn btn-primary">保存</button>
          <router-link to="/strategies" class="btn border border-border">取消</router-link>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useStrategyStore } from '../stores/strategy'

const route = useRoute()
const router = useRouter()
const store = useStrategyStore()

const id = computed(() => route.params.id as string)
const isNew = computed(() => id.value === 'new')

const form = ref<{
  name: string
  exchange: string
  trade_mode: 'paper' | 'live'
  config_json: string
}>({
  name: '',
  exchange: 'binance',
  trade_mode: 'paper',
  config_json: '',
})

const exampleConfig = JSON.stringify({
  symbol: 'BTCUSDT',
  interval: '1h',
  indicators: [
    { name: 'EMA', params: { period: 12 } },
    { name: 'EMA', params: { period: 26 } },
    { name: 'RSI', params: { period: 14 } }
  ],
  entry_rule: {
    logic: 'and',
    conditions: [
      { indicator: 'EMA', field: 'fast', operator: 'cross_above', value: 0 }
    ]
  },
  exit_rule: {
    logic: 'or',
    conditions: [
      { indicator: 'RSI', field: 'value', operator: '>=', value: 70 }
    ]
  }
}, null, 2)

onMounted(async () => {
  if (!isNew.value) {
    const s = await store.fetchOne(Number(id.value))
    form.value = {
      name: s.name,
      exchange: s.exchange,
      trade_mode: s.trade_mode,
      config_json: s.config_json,
    }
  }
})

async function handleSave() {
  if (isNew.value) {
    await store.create(form.value)
  } else {
    await store.update(Number(id.value), form.value)
  }
  router.push('/strategies')
}
</script>
