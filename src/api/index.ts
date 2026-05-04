import axios from 'axios'
import type { AccountSummary, ExchangeKey, Kline, Ticker, DepthData, Strategy, Backtest, Order, Trade, Position, RiskRule } from '../types'

const api = axios.create({ baseURL: '/api' })

// Account
export const getAccountSummary = () => api.get<AccountSummary>('/account/summary')

// Exchange Keys
export const getExchangeKeys = () => api.get<ExchangeKey[]>('/exchange/keys')
export const createExchangeKey = (key: Partial<ExchangeKey>) => api.post<ExchangeKey>('/exchange/keys', key)
export const deleteExchangeKey = (id: number) => api.delete(`/exchange/keys/${id}`)

// Market
export const getKlines = (params: { exchange?: string; symbol?: string; interval?: string; limit?: number }) =>
  api.get<Kline[]>('/market/klines', { params })
export const getTicker = (params: { exchange?: string; symbol: string }) =>
  api.get<Ticker>('/market/ticker', { params })
export const getDepth = (params: { exchange?: string; symbol: string; limit?: number }) =>
  api.get<DepthData>('/market/depth', { params })

// Strategies
export const getStrategies = () => api.get<Strategy[]>('/strategies')
export const getStrategy = (id: number) => api.get<Strategy>(`/strategies/${id}`)
export const createStrategy = (s: Partial<Strategy>) => api.post<Strategy>('/strategies', s)
export const updateStrategy = (id: number, s: Partial<Strategy>) => api.put<Strategy>(`/strategies/${id}`, s)
export const deleteStrategy = (id: number) => api.delete(`/strategies/${id}`)
export const activateStrategy = (id: number) => api.post(`/strategies/${id}/activate`)
export const runBacktest = (id: number) => api.post<Backtest>(`/strategies/${id}/backtest`)

// Backtests
export const getBacktests = (strategyId?: number) =>
  api.get<Backtest[]>('/backtests', { params: strategyId ? { strategy_id: strategyId } : {} })
export const getBacktest = (id: number) => api.get<Backtest>(`/backtests/${id}`)

// Orders
export const getOrders = (params?: { status?: string; mode?: string }) =>
  api.get<Order[]>('/orders', { params })
export const createOrder = (o: Partial<Order>) => api.post<Order>('/orders', o)
export const cancelOrder = (id: number) => api.delete(`/orders/${id}`)

// Trades & Positions
export const getTrades = () => api.get<Trade[]>('/trades')
export const getPositions = (strategyId?: number) =>
  api.get<Position[]>('/positions', { params: strategyId ? { strategy_id: strategyId } : {} })

// Risk
export const getRiskRules = (strategyId: number) => api.get<RiskRule>(`/risk/${strategyId}`)
export const updateRiskRules = (strategyId: number, r: Partial<RiskRule>) =>
  api.put(`/risk/${strategyId}`, r)
