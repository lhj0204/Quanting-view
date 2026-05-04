<template>
  <div class="flex h-screen flex-col">
    <TickerBar />
    <div class="flex flex-1 overflow-hidden">
      <aside class="w-56 bg-surface border-r border-border flex flex-col shrink-0">
        <div class="p-4 border-b border-border">
          <h1 class="text-lg font-bold text-blue">Quant Trading</h1>
          <p class="text-xs text-text-secondary mt-1">虚拟币量化交易系统</p>
        </div>
        <nav class="flex-1 py-2">
          <router-link
            v-for="item in navItems"
            :key="item.path"
            :to="item.path"
            class="flex items-center gap-3 px-4 py-2.5 mx-2 rounded text-sm transition-colors"
            :class="isActive(item.path) ? 'bg-blue/10 text-blue' : 'text-text-secondary hover:text-text hover:bg-bg'"
          >
            <span class="w-5 text-center">{{ item.icon }}</span>
            <span>{{ item.label }}</span>
          </router-link>
        </nav>
        <div class="p-4 border-t border-border text-xs text-text-secondary">
          v1.0.0
        </div>
      </aside>
      <main class="flex-1 overflow-auto">
        <div class="p-6">
          <slot />
        </div>
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useRoute } from 'vue-router'
import TickerBar from '../market/TickerBar.vue'

const route = useRoute()

const navItems = [
  { path: '/', label: '仪表盘', icon: '📊' },
  { path: '/market', label: '行情', icon: '📈' },
  { path: '/strategies', label: '策略管理', icon: '⚙' },
  { path: '/backtests', label: '回测记录', icon: '⏪' },
  { path: '/orders', label: '订单管理', icon: '📋' },
  { path: '/risk', label: '风险控制', icon: '🛡' },
  { path: '/settings', label: '系统设置', icon: '🔧' },
]

function isActive(path: string) {
  if (path === '/') return route.path === '/'
  return route.path.startsWith(path)
}
</script>
