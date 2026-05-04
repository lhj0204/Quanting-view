package services

import (
	"context"
	"encoding/json"
	"log"

	"quant/internal/exchange"
	"quant/internal/models"

	"github.com/robfig/cron/v3"
	"github.com/uptrace/bun"
)

type StrategyExecutor struct {
	DB        *bun.DB
	Exchanges map[string]exchange.Client
	Cron      *cron.Cron
}

func NewStrategyExecutor(db *bun.DB, exchanges map[string]exchange.Client) *StrategyExecutor {
	return &StrategyExecutor{
		DB:        db,
		Exchanges: exchanges,
		Cron:      cron.New(cron.WithSeconds()),
	}
}

func (se *StrategyExecutor) Start() {
	se.Cron.Start()
	log.Println("[Executor] Strategy executor started")
	se.Cron.AddFunc("@every 30s", func() {
		se.evaluateAllActive()
	})
}

func (se *StrategyExecutor) Stop() {
	se.Cron.Stop()
}

func (se *StrategyExecutor) evaluateAllActive() {
	ctx := context.Background()
	var strategies []models.Strategy
	if err := se.DB.NewSelect().Model(&strategies).Where("status = ?", "active").Scan(ctx); err != nil {
		log.Printf("[Executor] Failed to fetch active strategies: %v", err)
		return
	}
	for _, s := range strategies {
		se.evaluateStrategy(s)
	}
}

func (se *StrategyExecutor) evaluateStrategy(s models.Strategy) {
	ctx := context.Background()
	var config models.StrategyConfig
	if err := json.Unmarshal([]byte(s.ConfigJSON), &config); err != nil {
		log.Printf("[Executor] Invalid config for strategy %d: %v", s.ID, err)
		return
	}

	client, ok := se.Exchanges[s.Exchange]
	if !ok {
		return
	}

	klines, err := client.GetKlines(ctx, config.Symbol, config.Interval, 100, 0, 0)
	if err != nil {
		log.Printf("[Executor] Failed to fetch klines: %v", err)
		return
	}
	if len(klines) < 50 {
		return
	}

	is := NewIndicatorService()
	indValues := is.ComputeAll(klines, config)
	lastIdx := len(klines) - 1

	entrySignal := evaluateRules(config.EntryRule, indValues, klines, lastIdx)
	exitSignal := evaluateRules(config.ExitRule, indValues, klines, lastIdx)

	var position models.Position
	posErr := se.DB.NewSelect().Model(&position).Where("strategy_id = ? AND symbol = ?", s.ID, config.Symbol).Scan(ctx)
	hasPosition := posErr == nil && position.Quantity > 0

	ticker, err := client.GetTicker(ctx, config.Symbol)
	if err != nil {
		return
	}
	price := ticker.Price

	var risk models.RiskRule
	if err := se.DB.NewSelect().Model(&risk).Where("strategy_id = ?", s.ID).Scan(ctx); err != nil {
		risk.MaxPositionPct = 20
		risk.StopLossPct = 10
		risk.TakeProfitPct = 20
	}

	if hasPosition && risk.StopLossPct > 0 {
		lossPct := (price - position.AvgEntryPrice) / position.AvgEntryPrice * 100
		if lossPct <= -risk.StopLossPct {
			log.Printf("[Executor] Strategy %d: Stop-loss triggered (%.2f%%)", s.ID, lossPct)
			se.executeOrder(s, config.Symbol, "sell", "market", price, position.Quantity)
			return
		}
	}
	if hasPosition && risk.TakeProfitPct > 0 {
		profitPct := (price - position.AvgEntryPrice) / position.AvgEntryPrice * 100
		if profitPct >= risk.TakeProfitPct {
			log.Printf("[Executor] Strategy %d: Take-profit triggered (%.2f%%)", s.ID, profitPct)
			se.executeOrder(s, config.Symbol, "sell", "market", price, position.Quantity)
			return
		}
	}

	if exitSignal && hasPosition {
		log.Printf("[Executor] Strategy %d: Exit signal for %s at %.4f", s.ID, config.Symbol, price)
		se.executeOrder(s, config.Symbol, "sell", "market", price, position.Quantity)
		return
	}

	if entrySignal && !hasPosition {
		log.Printf("[Executor] Strategy %d: Entry signal for %s at %.4f", s.ID, config.Symbol, price)
		var account models.Account
		if err := se.DB.NewSelect().Model(&account).Order("id ASC").Limit(1).Scan(ctx); err == nil {
			maxPositionPct := risk.MaxPositionPct
			if maxPositionPct <= 0 {
				maxPositionPct = 20
			}
			cash := account.Balance * maxPositionPct / 100.0
			qty := cash / price
			if qty > 0 {
				se.executeOrder(s, config.Symbol, "buy", "market", price, qty)
			}
		}
	}
}

func (se *StrategyExecutor) executeOrder(s models.Strategy, symbol, side, orderType string, price, qty float64) {
	ctx := context.Background()

	order := &models.Order{
		StrategyID: s.ID,
		Exchange:   s.Exchange,
		Symbol:     symbol,
		Side:       side,
		Type:       orderType,
		Price:      price,
		Quantity:   qty,
		FilledQty:  qty,
		Status:     "filled",
		Mode:       s.TradeMode,
	}

	if _, err := se.DB.NewInsert().Model(order).Exec(ctx); err != nil {
		log.Printf("[Executor] Failed to insert order: %v", err)
		return
	}

	trade := &models.Trade{
		OrderID:  order.ID,
		Symbol:   symbol,
		Side:     side,
		Price:    price,
		Quantity: qty,
		Fee:      price * qty * 0.001,
	}

	if side == "sell" {
		var pos models.Position
		if err := se.DB.NewSelect().Model(&pos).Where("strategy_id = ? AND symbol = ?", s.ID, symbol).Scan(ctx); err == nil {
			trade.RealizedPnl = (price - pos.AvgEntryPrice) * qty
		}
	}

	if _, err := se.DB.NewInsert().Model(trade).Exec(ctx); err != nil {
		log.Printf("[Executor] Failed to insert trade: %v", err)
	}

	// Update position
	var pos models.Position
	posErr := se.DB.NewSelect().Model(&pos).Where("strategy_id = ? AND symbol = ?", s.ID, symbol).Scan(ctx)

	if side == "buy" {
		if posErr != nil {
			newPos := &models.Position{StrategyID: s.ID, Symbol: symbol, Quantity: qty, AvgEntryPrice: price}
			se.DB.NewInsert().Model(newPos).Exec(ctx)
		} else {
			newQty := pos.Quantity + qty
			pos.AvgEntryPrice = (pos.AvgEntryPrice*pos.Quantity + price*qty) / newQty
			pos.Quantity = newQty
			se.DB.NewUpdate().Model(&pos).Where("id = ?", pos.ID).Exec(ctx)
		}
	} else {
		if posErr == nil {
			pos.Quantity -= qty
			if pos.Quantity <= 0.00001 {
				se.DB.NewDelete().Model(&pos).Where("id = ?", pos.ID).Exec(ctx)
			} else {
				se.DB.NewUpdate().Model(&pos).Where("id = ?", pos.ID).Exec(ctx)
			}
		}
	}

	// Update account
	var account models.Account
	if err := se.DB.NewSelect().Model(&account).Order("id ASC").Limit(1).Scan(ctx); err == nil {
		if side == "buy" {
			account.Balance -= price * qty
		} else {
			account.Balance += price * qty
		}
		account.Balance -= trade.Fee
		se.DB.NewUpdate().Model(&account).Where("id = ?", account.ID).Exec(ctx)
	}
}
