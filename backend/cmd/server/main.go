package main

import (
	"fmt"
	"log"
	"net/http"

	"quant/internal/config"
	"quant/internal/database"
	"quant/internal/exchange"
	"quant/internal/handlers"
	"quant/internal/services"
	"quant/internal/ws"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if err := database.Init(cfg.Database.DSN); err != nil {
		log.Fatalf("Failed to init database: %v", err)
	}

	if cfg.Proxy.URL != "" {
		if err := exchange.SetProxy(cfg.Proxy.URL); err != nil {
			log.Printf("Warning: proxy setup failed: %v", err)
		} else {
			log.Printf("Proxy configured: %s", cfg.Proxy.URL)
		}
	}

	exchanges := map[string]exchange.Client{}

	// Initialize Binance
	binanceClient, err := exchange.NewClient("binance", cfg.Exchange.Binance.RestURL, cfg.Exchange.Binance.WSUrl, "", "", "")
	if err != nil {
		log.Printf("Warning: failed to init Binance client: %v", err)
	} else {
		if err := binanceClient.Ping(); err != nil {
			log.Printf("Warning: Binance ping failed: %v", err)
		} else {
			exchanges["binance"] = binanceClient
			log.Println("Binance client connected")
		}
	}

	// Initialize OKX
	okxClient, err := exchange.NewClient("okx", cfg.Exchange.OKX.RestURL, cfg.Exchange.OKX.WSUrl, "", "", "")
	if err != nil {
		log.Printf("Warning: failed to init OKX client: %v", err)
	} else {
		if err := okxClient.Ping(); err != nil {
			log.Printf("Warning: OKX ping failed: %v", err)
		} else {
			exchanges["okx"] = okxClient
			log.Println("OKX client connected")
		}
	}

	h := handlers.New(database.DB, exchanges)

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:3000", "http://127.0.0.1:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	hub := ws.NewHub()
	go hub.Run()

	executor := services.NewStrategyExecutor(database.DB, exchanges)
	executor.Start()
	defer executor.Stop()

	api := r.Group("/api")
	{
		api.GET("/account/summary", h.GetAccountSummary)

		exchangeGroup := api.Group("/exchange")
		{
			exchangeGroup.GET("/keys", h.ListExchangeKeys)
			exchangeGroup.POST("/keys", h.CreateExchangeKey)
			exchangeGroup.DELETE("/keys/:id", h.DeleteExchangeKey)
		}

		marketGroup := api.Group("/market")
		{
			marketGroup.GET("/klines", h.GetKlines)
			marketGroup.GET("/ticker", h.GetTicker)
			marketGroup.GET("/depth", h.GetDepth)
		}

		strategyGroup := api.Group("/strategies")
		{
			strategyGroup.GET("", h.ListStrategies)
			strategyGroup.POST("", h.CreateStrategy)
			strategyGroup.GET("/:id", h.GetStrategy)
			strategyGroup.PUT("/:id", h.UpdateStrategy)
			strategyGroup.DELETE("/:id", h.DeleteStrategy)
			strategyGroup.POST("/:id/activate", h.ActivateStrategy)
			strategyGroup.POST("/:id/backtest", h.RunBacktest)
		}

		api.GET("/backtests", h.ListBacktests)
		api.GET("/backtests/:id", h.GetBacktest)

		api.GET("/orders", h.ListOrders)
		api.POST("/orders", h.CreateOrder)
		api.DELETE("/orders/:id", h.CancelOrder)

		api.GET("/trades", h.ListTrades)
		api.GET("/positions", h.ListPositions)

		api.GET("/risk/:strategy_id", h.GetRiskRules)
		api.PUT("/risk/:strategy_id", h.UpdateRiskRules)
	}

	r.GET("/ws", func(c *gin.Context) {
		ws.HandleWebSocket(hub, c.Writer, c.Request)
	})

	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Server starting on %s", addr)
	log.Printf("Exchanges available: %v", getExchangeNames(exchanges))

	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func getExchangeNames(m map[string]exchange.Client) []string {
	names := make([]string, 0, len(m))
	for k := range m {
		names = append(names, k)
	}
	return names
}
