<template>
  <div>
    <h2 class="text-xl font-bold mb-6">风险控制</h2>

    <div class="card max-w-xl mb-4">
      <div class="text-sm text-text-secondary mb-3">选择策略查看/编辑风控规则</div>
      <select v-model="selectedId" class="input" @change="loadRules">
        <option :value="0" disabled>选择策略...</option>
        <option v-for="s in strategies" :key="s.id" :value="s.id">{{ s.name }} ({{ s.exchange }})</option>
      </select>
    </div>

    <div v-if="selectedId && rules" class="card max-w-xl">
      <form @submit.prevent="handleSave" class="space-y-4">
        <div>
          <label class="text-xs text-text-secondary">最大持仓比例 (%)</label>
          <input v-model.number="rules.max_position_pct" type="number" step="1" class="input" />
        </div>
        <div>
          <label class="text-xs text-text-secondary">止损比例 (%)</label>
          <input v-model.number="rules.stop_loss_pct" type="number" step="0.1" class="input" />
        </div>
        <div>
          <label class="text-xs text-text-secondary">止盈比例 (%)</label>
          <input v-model.number="rules.take_profit_pct" type="number" step="0.1" class="input" />
        </div>
        <div>
          <label class="text-xs text-text-secondary">最大回撤 (%)</label>
          <input v-model.number="rules.max_drawdown_pct" type="number" step="0.1" class="input" />
        </div>
        <button type="submit" class="btn btn-primary">保存</button>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useStrategyStore } from '../stores/strategy'
import { getRiskRules, updateRiskRules } from '../api'
import type { RiskRule } from '../types'

const store = useStrategyStore()
const strategies = ref(store.strategies)
const selectedId = ref(0)
const rules = ref<RiskRule | null>(null)

onMounted(() => store.fetchAll().then(() => strategies.value = store.strategies))

async function loadRules() {
  if (!selectedId.value) return
  try {
    const { data } = await getRiskRules(selectedId.value)
    rules.value = data
  } catch {
    rules.value = {
      strategy_id: selectedId.value,
      max_position_pct: 20,
      stop_loss_pct: 10,
      take_profit_pct: 20,
      max_drawdown_pct: 30,
    }
  }
}

async function handleSave() {
  if (!rules.value) return
  await updateRiskRules(selectedId.value, rules.value)
  alert('保存成功')
}
</script>
