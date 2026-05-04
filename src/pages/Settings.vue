<template>
  <div>
    <h2 class="text-xl font-bold mb-6">系统设置</h2>

    <div class="card mb-6">
      <h3 class="text-sm font-semibold mb-4">交易所 API 密钥</h3>
      <div class="mb-4 flex gap-3">
        <button @click="showForm = !showForm" class="btn btn-primary text-sm">{{ showForm ? '取消' : '添加密钥' }}</button>
      </div>

      <div v-if="showForm" class="card bg-bg mb-4">
        <form @submit.prevent="handleAdd" class="space-y-3">
          <div class="grid grid-cols-2 gap-3">
            <div>
              <label class="text-xs text-text-secondary">交易所</label>
              <select v-model="form.exchange" class="input">
                <option value="binance">Binance</option>
                <option value="okx">OKX</option>
              </select>
            </div>
            <div>
              <label class="text-xs text-text-secondary">名称</label>
              <input v-model="form.name" class="input" placeholder="如: 主账户" />
            </div>
            <div>
              <label class="text-xs text-text-secondary">API Key</label>
              <input v-model="form.api_key" class="input" type="password" />
            </div>
            <div>
              <label class="text-xs text-text-secondary">Secret Key</label>
              <input v-model="form.secret_key" class="input" type="password" />
            </div>
            <div v-if="form.exchange === 'okx'">
              <label class="text-xs text-text-secondary">Passphrase (OKX)</label>
              <input v-model="form.passphrase" class="input" type="password" />
            </div>
            <div class="flex items-center gap-4 pt-5">
              <label class="flex items-center gap-2 text-sm">
                <input type="checkbox" v-model="form.testnet" />
                测试网
              </label>
            </div>
          </div>
          <button type="submit" class="btn btn-primary text-sm">保存</button>
        </form>
      </div>

      <div v-if="keys.length === 0" class="text-text-secondary text-sm py-4 text-center">暂无 API 密钥</div>
      <table v-else class="w-full text-sm">
        <thead>
          <tr class="text-text-secondary text-left">
            <th class="pb-2">名称</th>
            <th class="pb-2">交易所</th>
            <th class="pb-2">模式</th>
            <th class="pb-2">状态</th>
            <th class="pb-2">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="key in keys" :key="key.id" class="border-t border-border">
            <td class="py-2">{{ key.name }}</td>
            <td class="py-2">
              <span class="badge" :class="key.exchange === 'binance' ? 'badge-blue' : 'badge-green'">{{ key.exchange }}</span>
            </td>
            <td class="py-2">
              <span class="badge" :class="key.testnet ? 'badge-blue' : 'badge-green'">{{ key.testnet ? '测试网' : '主网' }}</span>
            </td>
            <td class="py-2">
              <span class="badge" :class="key.enabled ? 'badge-green' : 'badge-red'">{{ key.enabled ? '已启用' : '已禁用' }}</span>
            </td>
            <td class="py-2">
              <button @click="handleDelete(key.id)" class="text-red text-xs hover:underline">删除</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getExchangeKeys, createExchangeKey, deleteExchangeKey } from '../api'
import type { ExchangeKey } from '../types'

const keys = ref<ExchangeKey[]>([])
const showForm = ref(false)
const form = ref({ exchange: 'binance', name: '', api_key: '', secret_key: '', passphrase: '', testnet: true })

onMounted(async () => {
  const { data } = await getExchangeKeys()
  keys.value = data
})

async function handleAdd() {
  const { data } = await createExchangeKey(form.value)
  keys.value.push(data)
  showForm.value = false
  form.value = { exchange: 'binance', name: '', api_key: '', secret_key: '', passphrase: '', testnet: true }
}

async function handleDelete(id: number) {
  await deleteExchangeKey(id)
  keys.value = keys.value.filter(k => k.id !== id)
}
</script>
