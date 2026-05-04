package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
	"quant/internal/exchange"
	"quant/internal/models"
	"quant/internal/services"
)

type Handler struct {
	DB        *bun.DB
	Exchanges map[string]exchange.Client
}

func New(db *bun.DB, exchanges map[string]exchange.Client) *Handler {
	return &Handler{DB: db, Exchanges: exchanges}
}

// ---- Account ----

func (h *Handler) GetAccountSummary(c *gin.Context) {
	var acct models.Account
	if err := h.DB.NewSelect().Model(&acct).Order("id ASC").Limit(1).Scan(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var positions []models.Position
	h.DB.NewSelect().Model(&positions).Where("quantity != 0").Scan(c.Request.Context())
	if positions == nil {
		positions = []models.Position{}
	}

	var trades []models.Trade
	h.DB.NewSelect().Model(&trades).Scan(c.Request.Context())

	var realizedPnl float64
	for _, t := range trades {
		realizedPnl += t.RealizedPnl
	}

	var unrealizedPnl float64
	// For paper mode, unrealized PnL requires current prices. Simplified for now.

	summary := models.AccountSummary{
		Balance:        acct.Balance,
		InitialBalance: acct.InitialBalance,
		Currency:       acct.Currency,
		UnrealizedPnl:  unrealizedPnl,
		RealizedPnl:    realizedPnl,
		TotalPnl:       acct.Balance - acct.InitialBalance + realizedPnl,
		Positions:      positions,
	}
	if acct.InitialBalance > 0 {
		summary.TotalPnlPct = (acct.Balance - acct.InitialBalance + realizedPnl) / acct.InitialBalance * 100
	}

	c.JSON(http.StatusOK, summary)
}

// ---- Exchange Keys ----

func (h *Handler) ListExchangeKeys(c *gin.Context) {
	var keys []models.ExchangeKey
	if err := h.DB.NewSelect().Model(&keys).Order("id ASC").Scan(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if keys == nil {
		keys = []models.ExchangeKey{}
	}
	c.JSON(http.StatusOK, keys)
}

func (h *Handler) CreateExchangeKey(c *gin.Context) {
	var key models.ExchangeKey
	if err := c.ShouldBindJSON(&key); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if _, err := h.DB.NewInsert().Model(&key).Exec(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, key)
}

func (h *Handler) DeleteExchangeKey(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	_, err := h.DB.NewDelete().Model((*models.ExchangeKey)(nil)).Where("id = ?", id).Exec(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// ---- Market Data ----

func (h *Handler) GetKlines(c *gin.Context) {
	exch := c.DefaultQuery("exchange", "binance")
	symbol := c.DefaultQuery("symbol", "BTCUSDT")
	interval := c.DefaultQuery("interval", "1h")
	limitStr := c.DefaultQuery("limit", "500")
	limit, _ := strconv.Atoi(limitStr)

	var startTime, endTime int64
	if st := c.Query("start"); st != "" {
		startTime, _ = strconv.ParseInt(st, 10, 64)
	}
	if et := c.Query("end"); et != "" {
		endTime, _ = strconv.ParseInt(et, 10, 64)
	}

	client, ok := h.Exchanges[exch]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported exchange: " + exch})
		return
	}

	klines, err := client.GetKlines(c.Request.Context(), symbol, interval, limit, startTime, endTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, klines)
}

func (h *Handler) GetTicker(c *gin.Context) {
	exch := c.DefaultQuery("exchange", "binance")
	symbol := c.DefaultQuery("symbol", "BTCUSDT")

	client, ok := h.Exchanges[exch]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported exchange: " + exch})
		return
	}

	ticker, err := client.GetTicker(c.Request.Context(), symbol)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ticker)
}

func (h *Handler) GetDepth(c *gin.Context) {
	exch := c.DefaultQuery("exchange", "binance")
	symbol := c.DefaultQuery("symbol", "BTCUSDT")
	limitStr := c.DefaultQuery("limit", "20")
	limit, _ := strconv.Atoi(limitStr)

	client, ok := h.Exchanges[exch]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported exchange: " + exch})
		return
	}

	depth, err := client.GetDepth(c.Request.Context(), symbol, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, depth)
}

// ---- Strategies ----

func (h *Handler) ListStrategies(c *gin.Context) {
	var strategies []models.Strategy
	if err := h.DB.NewSelect().Model(&strategies).Order("updated_at DESC").Scan(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if strategies == nil {
		strategies = []models.Strategy{}
	}
	c.JSON(http.StatusOK, strategies)
}

func (h *Handler) CreateStrategy(c *gin.Context) {
	var s models.Strategy
	if err := c.ShouldBindJSON(&s); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	s.Status = "draft"
	if s.TradeMode == "" {
		s.TradeMode = "paper"
	}
	if _, err := h.DB.NewInsert().Model(&s).Exec(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Create default risk rules
	risk := &models.RiskRule{
		StrategyID:     s.ID,
		MaxPositionPct: 20,
		StopLossPct:    10,
		TakeProfitPct:  20,
		MaxDrawdownPct: 30,
	}
	h.DB.NewInsert().Model(risk).Exec(c.Request.Context())

	c.JSON(http.StatusCreated, s)
}

func (h *Handler) GetStrategy(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var s models.Strategy
	if err := h.DB.NewSelect().Model(&s).Where("id = ?", id).Scan(c.Request.Context()); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "strategy not found"})
		return
	}
	c.JSON(http.StatusOK, s)
}

func (h *Handler) UpdateStrategy(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var s models.Strategy
	if err := h.DB.NewSelect().Model(&s).Where("id = ?", id).Scan(c.Request.Context()); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "strategy not found"})
		return
	}
	if err := c.ShouldBindJSON(&s); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	s.ID = id
	s.UpdatedAt = time.Now()
	if _, err := h.DB.NewUpdate().Model(&s).Where("id = ?", id).Exec(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, s)
}

func (h *Handler) DeleteStrategy(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	_, err := h.DB.NewDelete().Model((*models.Strategy)(nil)).Where("id = ?", id).Exec(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (h *Handler) ActivateStrategy(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var s models.Strategy
	if err := h.DB.NewSelect().Model(&s).Where("id = ?", id).Scan(c.Request.Context()); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "strategy not found"})
		return
	}
	s.Status = "active"
	s.UpdatedAt = time.Now()
	if _, err := h.DB.NewUpdate().Model(&s).Where("id = ?", id).Column("status", "updated_at").Exec(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, s)
}

// ---- Orders ----

func (h *Handler) ListOrders(c *gin.Context) {
	var orders []models.Order
	query := h.DB.NewSelect().Model(&orders).Order("created_at DESC")
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if mode := c.Query("mode"); mode != "" {
		query = query.Where("mode = ?", mode)
	}
	if err := query.Scan(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if orders == nil {
		orders = []models.Order{}
	}
	c.JSON(http.StatusOK, orders)
}

func (h *Handler) CreateOrder(c *gin.Context) {
	var o models.Order
	if err := c.ShouldBindJSON(&o); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	o.Status = "open"
	o.FilledQty = 0
	if o.Mode == "" {
		o.Mode = "paper"
	}
	if _, err := h.DB.NewInsert().Model(&o).Exec(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, o)
}

func (h *Handler) CancelOrder(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var o models.Order
	if err := h.DB.NewSelect().Model(&o).Where("id = ?", id).Scan(c.Request.Context()); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}
	o.Status = "cancelled"
	o.UpdatedAt = time.Now()
	if _, err := h.DB.NewUpdate().Model(&o).Where("id = ?", id).Column("status", "updated_at").Exec(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, o)
}

// ---- Trades ----

func (h *Handler) ListTrades(c *gin.Context) {
	var trades []models.Trade
	if err := h.DB.NewSelect().Model(&trades).Order("created_at DESC").Scan(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if trades == nil {
		trades = []models.Trade{}
	}
	c.JSON(http.StatusOK, trades)
}

// ---- Positions ----

func (h *Handler) ListPositions(c *gin.Context) {
	var positions []models.Position
	query := h.DB.NewSelect().Model(&positions).Where("quantity != 0")
	if sID := c.Query("strategy_id"); sID != "" {
		query = query.Where("strategy_id = ?", sID)
	}
	if err := query.Scan(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if positions == nil {
		positions = []models.Position{}
	}
	c.JSON(http.StatusOK, positions)
}

// ---- Risk Rules ----

func (h *Handler) GetRiskRules(c *gin.Context) {
	sID, _ := strconv.ParseInt(c.Param("strategy_id"), 10, 64)
	var r models.RiskRule
	if err := h.DB.NewSelect().Model(&r).Where("strategy_id = ?", sID).Scan(c.Request.Context()); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "risk rules not found"})
		return
	}
	c.JSON(http.StatusOK, r)
}

func (h *Handler) UpdateRiskRules(c *gin.Context) {
	sID, _ := strconv.ParseInt(c.Param("strategy_id"), 10, 64)
	var r models.RiskRule
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	existing := new(models.RiskRule)
	if err := h.DB.NewSelect().Model(existing).Where("strategy_id = ?", sID).Scan(c.Request.Context()); err != nil {
		// Insert new
		r.StrategyID = sID
		if _, err := h.DB.NewInsert().Model(&r).Exec(c.Request.Context()); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	} else {
		r.StrategyID = sID
		if _, err := h.DB.NewUpdate().Model(&r).Where("strategy_id = ?", sID).Exec(c.Request.Context()); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	c.JSON(http.StatusOK, r)
}

// ---- Backtests ----

func (h *Handler) RunBacktest(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var s models.Strategy
	if err := h.DB.NewSelect().Model(&s).Where("id = ?", id).Scan(c.Request.Context()); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "strategy not found"})
		return
	}

	var config models.StrategyConfig
	if err := json.Unmarshal([]byte(s.ConfigJSON), &config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid strategy config: " + err.Error()})
		return
	}

	client, ok := h.Exchanges[s.Exchange]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "exchange not available: " + s.Exchange})
		return
	}

	klines, err := client.GetKlines(c.Request.Context(), config.Symbol, config.Interval, 500, 0, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch klines: " + err.Error()})
		return
	}

	bs := &services.BacktestService{DB: h.DB}
	bt, err := bs.Run(id, klines, config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if _, err := h.DB.NewInsert().Model(bt).Exec(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, bt)
}

func (h *Handler) ListBacktests(c *gin.Context) {
	var backtests []models.Backtest
	query := h.DB.NewSelect().Model(&backtests).Order("created_at DESC")
	if sID := c.Query("strategy_id"); sID != "" {
		query = query.Where("strategy_id = ?", sID)
	}
	if err := query.Scan(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if backtests == nil {
		backtests = []models.Backtest{}
	}
	c.JSON(http.StatusOK, backtests)
}

func (h *Handler) GetBacktest(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var bt models.Backtest
	if err := h.DB.NewSelect().Model(&bt).Where("id = ?", id).Scan(c.Request.Context()); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "backtest not found"})
		return
	}
	c.JSON(http.StatusOK, bt)
}
