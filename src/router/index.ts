import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      name: 'dashboard',
      component: () => import('../pages/Dashboard.vue')
    },
    {
      path: '/market',
      name: 'market',
      component: () => import('../pages/Market.vue')
    },
    {
      path: '/market/:symbol',
      name: 'chart',
      component: () => import('../pages/Chart.vue')
    },
    {
      path: '/strategies',
      name: 'strategies',
      component: () => import('../pages/StrategyList.vue')
    },
    {
      path: '/strategies/:id',
      name: 'strategy-edit',
      component: () => import('../pages/StrategyEdit.vue')
    },
    {
      path: '/backtests',
      name: 'backtests',
      component: () => import('../pages/BacktestList.vue')
    },
    {
      path: '/backtests/:id',
      name: 'backtest-detail',
      component: () => import('../pages/BacktestDetail.vue')
    },
    {
      path: '/orders',
      name: 'orders',
      component: () => import('../pages/Orders.vue')
    },
    {
      path: '/risk',
      name: 'risk',
      component: () => import('../pages/RiskManager.vue')
    },
    {
      path: '/settings',
      name: 'settings',
      component: () => import('../pages/Settings.vue')
    }
  ]
})

export default router
