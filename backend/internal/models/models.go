package models

import "time"

type ExchangeKey struct {
	ID         int64     `bun:"id,pk,autoincrement" json:"id"`
	Exchange   string    `bun:"exchange,notnull" json:"exchange"`
	Name       string    `bun:"name,notnull" json:"name"`
	APIKey     string    `bun:"api_key,notnull" json:"api_key"`
	SecretKey  string    `bun:"secret_key,notnull" json:"secret_key"`
	Passphrase string    `bun:"passphrase" json:"passphrase,omitempty"`
	Testnet    bool      `bun:"testnet,default:true" json:"testnet"`
	Enabled    bool      `bun:"enabled,default:true" json:"enabled"`
	CreatedAt  time.Time `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt  time.Time `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`
}

type Strategy struct {
	ID         int64     `bun:"id,pk,autoincrement" json:"id"`
	Name       string    `bun:"name,notnull" json:"name"`
	Exchange   string    `bun:"exchange,notnull" json:"exchange"`
	ConfigJSON string    `bun:"config_json,type:text" json:"config_json"`
	TradeMode  string    `bun:"trade_mode,notnull,default:'paper'" json:"trade_mode"`
	Status     string    `bun:"status,notnull,default:'draft'" json:"status"`
	CreatedAt  time.Time `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt  time.Time `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`
}

type StrategyConfig struct {
	Symbol    string                   `json:"symbol"`
	Interval  string                   `json:"interval"`
	Indicators []IndicatorConfig       `json:"indicators"`
	EntryRule EntryExitRule            `json:"entry_rule"`
	ExitRule  EntryExitRule            `json:"exit_rule"`
}

type IndicatorConfig struct {
	Name   string            `json:"name"`
	Params map[string]int    `json:"params"`
}

type EntryExitRule struct {
	Logic    string            `json:"logic"`    // "and" or "or"
	Conditions []RuleCondition `json:"conditions"`
}

type RuleCondition struct {
	Indicator string  `json:"indicator"`
	Field     string  `json:"field"`
	Operator  string  `json:"operator"` // ">", "<", ">=", "<=", "==", "cross_above", "cross_below"
	Value     float64 `json:"value"`
}

type Backtest struct {
	ID             int64     `bun:"id,pk,autoincrement" json:"id"`
	StrategyID     int64     `bun:"strategy_id,notnull" json:"strategy_id"`
	Symbol         string    `bun:"symbol,notnull" json:"symbol"`
	Interval       string    `bun:"interval,notnull" json:"interval"`
	StartTime      time.Time `bun:"start_time,notnull" json:"start_time"`
	EndTime        time.Time `bun:"end_time,notnull" json:"end_time"`
	InitialCapital float64   `bun:"initial_capital,notnull" json:"initial_capital"`
	FinalCapital   float64   `bun:"final_capital" json:"final_capital"`
	TotalTrades    int       `bun:"total_trades" json:"total_trades"`
	WinRate        float64   `bun:"win_rate" json:"win_rate"`
	Sharpe         float64   `bun:"sharpe" json:"sharpe"`
	MaxDrawdown    float64   `bun:"max_drawdown" json:"max_drawdown"`
	EquityCurveJSON string   `bun:"equity_curve_json,type:text" json:"equity_curve_json"`
	CreatedAt      time.Time `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
}

type Order struct {
	ID         int64     `bun:"id,pk,autoincrement" json:"id"`
	StrategyID int64     `bun:"strategy_id" json:"strategy_id"`
	Exchange   string    `bun:"exchange,notnull" json:"exchange"`
	Symbol     string    `bun:"symbol,notnull" json:"symbol"`
	Side       string    `bun:"side,notnull" json:"side"`
	Type       string    `bun:"type,notnull" json:"type"`
	Price      float64   `bun:"price" json:"price"`
	Quantity   float64   `bun:"quantity,notnull" json:"quantity"`
	FilledQty  float64   `bun:"filled_qty" json:"filled_qty"`
	Status     string    `bun:"status,notnull,default:'open'" json:"status"`
	Mode       string    `bun:"mode,notnull,default:'paper'" json:"mode"`
	ExtOrderID string    `bun:"ext_order_id" json:"ext_order_id"`
	CreatedAt  time.Time `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt  time.Time `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`
}

type Trade struct {
	ID          int64     `bun:"id,pk,autoincrement" json:"id"`
	OrderID     int64     `bun:"order_id,notnull" json:"order_id"`
	Symbol      string    `bun:"symbol,notnull" json:"symbol"`
	Side        string    `bun:"side,notnull" json:"side"`
	Price       float64   `bun:"price,notnull" json:"price"`
	Quantity    float64   `bun:"quantity,notnull" json:"quantity"`
	Fee         float64   `bun:"fee" json:"fee"`
	RealizedPnl float64   `bun:"realized_pnl" json:"realized_pnl"`
	CreatedAt   time.Time `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
}

type Position struct {
	ID             int64     `bun:"id,pk,autoincrement" json:"id"`
	StrategyID     int64     `bun:"strategy_id" json:"strategy_id"`
	Symbol         string    `bun:"symbol,notnull" json:"symbol"`
	Quantity       float64   `bun:"quantity,notnull" json:"quantity"`
	AvgEntryPrice  float64   `bun:"avg_entry_price,notnull" json:"avg_entry_price"`
	CreatedAt      time.Time `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt      time.Time `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`
}

type RiskRule struct {
	ID               int64   `bun:"id,pk,autoincrement" json:"id"`
	StrategyID       int64   `bun:"strategy_id,notnull,unique" json:"strategy_id"`
	MaxPositionPct   float64 `bun:"max_position_pct,notnull" json:"max_position_pct"`
	StopLossPct      float64 `bun:"stop_loss_pct" json:"stop_loss_pct"`
	TakeProfitPct    float64 `bun:"take_profit_pct" json:"take_profit_pct"`
	MaxDrawdownPct   float64 `bun:"max_drawdown_pct" json:"max_drawdown_pct"`
}

type Account struct {
	ID             int64     `bun:"id,pk,autoincrement" json:"id"`
	Balance        float64   `bun:"balance,notnull" json:"balance"`
	Currency       string    `bun:"currency,notnull,default:'USDT'" json:"currency"`
	InitialBalance float64   `bun:"initial_balance,notnull" json:"initial_balance"`
	CreatedAt      time.Time `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt      time.Time `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`
}

type MarketDataCache struct {
	ID        int64     `bun:"id,pk,autoincrement" json:"id"`
	Symbol    string    `bun:"symbol,notnull" json:"symbol"`
	Exchange  string    `bun:"exchange,notnull" json:"exchange"`
	Interval  string    `bun:"interval,notnull" json:"interval"`
	KlineJSON string    `bun:"kline_json,type:text" json:"kline_json"`
	UpdatedAt time.Time `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`
}

// API response types

type AccountSummary struct {
	Balance        float64   `json:"balance"`
	InitialBalance float64   `json:"initial_balance"`
	Currency       string    `json:"currency"`
	UnrealizedPnl  float64   `json:"unrealized_pnl"`
	RealizedPnl    float64   `json:"realized_pnl"`
	TotalPnl       float64   `json:"total_pnl"`
	TotalPnlPct    float64   `json:"total_pnl_pct"`
	Positions      []Position `json:"positions"`
}

type Kline struct {
	Time   int64   `json:"time"`
	Open   float64 `json:"open"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	Close  float64 `json:"close"`
	Volume float64 `json:"volume"`
}

type Ticker struct {
	Symbol     string  `json:"symbol"`
	Price      float64 `json:"price"`
	Change24h  float64 `json:"change_24h"`
	ChangePct  float64 `json:"change_pct"`
	High24h    float64 `json:"high_24h"`
	Low24h     float64 `json:"low_24h"`
	Volume24h  float64 `json:"volume_24h"`
	Exchange   string  `json:"exchange"`
}

type DepthData struct {
	Bids [][2]string `json:"bids"`
	Asks [][2]string `json:"asks"`
}

type EquityPoint struct {
	Time   int64   `json:"time"`
	Equity float64 `json:"equity"`
}
