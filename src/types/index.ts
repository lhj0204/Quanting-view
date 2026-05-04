export interface Kline {
  time: number
  open: number
  high: number
  low: number
  close: number
  volume: number
}

export interface Ticker {
  symbol: string
  price: number
  change_24h: number
  change_pct: number
  high_24h: number
  low_24h: number
  volume_24h: number
  exchange: string
}

export interface DepthData {
  bids: [string, string][]
  asks: [string, string][]
}

export interface ExchangeKey {
  id: number
  exchange: string
  name: string
  api_key: string
  secret_key: string
  passphrase?: string
  testnet: boolean
  enabled: boolean
  created_at: string
  updated_at: string
}

export interface Strategy {
  id: number
  name: string
  exchange: string
  config_json: string
  trade_mode: 'paper' | 'live'
  status: 'draft' | 'active' | 'paused' | 'stopped'
  created_at: string
  updated_at: string
}

export interface StrategyConfig {
  symbol: string
  interval: string
  indicators: IndicatorConfig[]
  entry_rule: Rule
  exit_rule: Rule
}

export interface IndicatorConfig {
  name: string
  params: Record<string, number>
}

export interface Rule {
  logic: 'and' | 'or'
  conditions: RuleCondition[]
}

export interface RuleCondition {
  indicator: string
  field: string
  operator: string
  value: number
}

export interface Backtest {
  id: number
  strategy_id: number
  symbol: string
  interval: string
  start_time: string
  end_time: string
  initial_capital: number
  final_capital: number
  total_trades: number
  win_rate: number
  sharpe: number
  max_drawdown: number
  equity_curve_json: string
  created_at: string
}

export interface EquityPoint {
  time: number
  equity: number
}

export interface Order {
  id: number
  strategy_id: number
  exchange: string
  symbol: string
  side: 'buy' | 'sell'
  type: 'market' | 'limit'
  price: number
  quantity: number
  filled_qty: number
  status: 'open' | 'filled' | 'cancelled'
  mode: 'paper' | 'live'
  ext_order_id: string
  created_at: string
  updated_at: string
}

export interface Trade {
  id: number
  order_id: number
  symbol: string
  side: string
  price: number
  quantity: number
  fee: number
  realized_pnl: number
  created_at: string
}

export interface Position {
  id: number
  strategy_id: number
  symbol: string
  quantity: number
  avg_entry_price: number
  created_at: string
  updated_at: string
}

export interface RiskRule {
  id?: number
  strategy_id: number
  max_position_pct: number
  stop_loss_pct: number
  take_profit_pct: number
  max_drawdown_pct: number
}

export interface AccountSummary {
  balance: number
  initial_balance: number
  currency: string
  unrealized_pnl: number
  realized_pnl: number
  total_pnl: number
  total_pnl_pct: number
  positions: Position[]
}

export interface WSMessage {
  event: string
  channel: string
  symbol?: string
  interval?: string
  data?: any
}
