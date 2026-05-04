package services

import (
	"encoding/json"
	"fmt"
	"time"

	"quant/internal/indicators"
	"quant/internal/models"

	"github.com/uptrace/bun"
)

type BacktestService struct {
	DB *bun.DB
}

type IndicatorService struct{}

func NewIndicatorService() *IndicatorService {
	return &IndicatorService{}
}

func (is *IndicatorService) ComputeAll(klines []models.Kline, config models.StrategyConfig) map[string]interface{} {
	result := make(map[string]interface{})

	for _, ind := range config.Indicators {
		switch ind.Name {
		case "EMA":
			period := 20
			if p, ok := ind.Params["period"]; ok {
				period = p
			}
			result["EMA"] = indicators.ComputeEMA(klines, period)
		case "SMA":
			period := 20
			if p, ok := ind.Params["period"]; ok {
				period = p
			}
			result["SMA"] = indicators.ComputeSMA(klines, period)
		case "RSI":
			period := 14
			if p, ok := ind.Params["period"]; ok {
				period = p
			}
			result["RSI"] = indicators.ComputeRSI(klines, period)
		case "MACD":
			fast := 12
			slow := 26
			signal := 9
			if p, ok := ind.Params["fast"]; ok {
				fast = p
			}
			if p, ok := ind.Params["slow"]; ok {
				slow = p
			}
			if p, ok := ind.Params["signal"]; ok {
				signal = p
			}
			result["MACD"] = indicators.ComputeMACD(klines, fast, slow, signal)
		case "BB":
			period := 20
			mult := 2.0
			if p, ok := ind.Params["period"]; ok {
				period = p
			}
			if p, ok := ind.Params["multiplier"]; ok {
				mult = float64(p)
			}
			result["BB"] = indicators.ComputeBollingerBands(klines, period, mult)
		case "ATR":
			period := 14
			if p, ok := ind.Params["period"]; ok {
				period = p
			}
			result["ATR"] = indicators.ComputeATR(klines, period)
		case "OBV":
			result["OBV"] = indicators.ComputeOBV(klines)
		case "KDJ":
			period := 9
			s1 := 3
			s2 := 3
			if p, ok := ind.Params["period"]; ok {
				period = p
			}
			if p, ok := ind.Params["smooth1"]; ok {
				s1 = p
			}
			if p, ok := ind.Params["smooth2"]; ok {
				s2 = p
			}
			k, d, j := indicators.ComputeKDJ(klines, period, s1, s2)
			result["KDJ_K"] = k
			result["KDJ_D"] = d
			result["KDJ_J"] = j
		}
	}
	return result
}

func (bs *BacktestService) Run(strategyID int64, klines []models.Kline, config models.StrategyConfig) (*models.Backtest, error) {
	is := NewIndicatorService()
	indValues := is.ComputeAll(klines, config)

	initialCapital := 10000.0
	cash := initialCapital
	position := 0.0
	trades := []struct {
		Time   int64
		Side   string
		Price  float64
		Qty    float64
		Reason string
	}{}
	equityCurve := []models.EquityPoint{}

	for i := 1; i < len(klines); i++ {
		currentPrice := klines[i].Close
		signal := evaluateRules(config.EntryRule, indValues, klines, i)
		exitSignal := evaluateRules(config.ExitRule, indValues, klines, i)

		// Check exit signal first
		if exitSignal && position > 0 {
			trades = append(trades, struct {
				Time   int64
				Side   string
				Price  float64
				Qty    float64
				Reason string
			}{klines[i].Time, "sell", currentPrice, position, "exit_rule"})
			cash += position * currentPrice
			position = 0
		}

		// Check entry signal
		if signal && position == 0 {
			position = cash / currentPrice * 0.95
			cash -= position * currentPrice
			trades = append(trades, struct {
				Time   int64
				Side   string
				Price  float64
				Qty    float64
				Reason string
			}{klines[i].Time, "buy", currentPrice, position, "entry_rule"})
		}

		equity := cash + position*currentPrice
		equityCurve = append(equityCurve, models.EquityPoint{
			Time:   klines[i].Time,
			Equity: equity,
		})
	}

	finalEquity := cash + position*klines[len(klines)-1].Close

	totalTrades := len(trades)
	wins := 0
	for i := 0; i < len(trades)-1; i += 2 {
		if i+1 < len(trades) && trades[i+1].Price > trades[i].Price {
			wins++
		}
	}
	winRate := 0.0
	if totalTrades/2 > 0 {
		winRate = float64(wins) / float64(totalTrades/2) * 100
	}

	// Calculate max drawdown
	maxEquity := 0.0
	maxDrawdown := 0.0
	for _, ep := range equityCurve {
		if ep.Equity > maxEquity {
			maxEquity = ep.Equity
		}
		dd := (maxEquity - ep.Equity) / maxEquity * 100
		if dd > maxDrawdown {
			maxDrawdown = dd
		}
	}

	// Simple Sharpe ratio
	sharpe := 0.0
	if len(equityCurve) > 1 {
		returns := make([]float64, len(equityCurve)-1)
		sum := 0.0
		for i := 1; i < len(equityCurve); i++ {
			r := (equityCurve[i].Equity - equityCurve[i-1].Equity) / equityCurve[i-1].Equity
			returns[i-1] = r
			sum += r
		}
		mean := sum / float64(len(returns))
		variance := 0.0
		for _, r := range returns {
			variance += (r - mean) * (r - mean)
		}
		if variance > 0 && len(returns) > 1 {
			stdDev := 0.0
			variance /= float64(len(returns) - 1)
			stdDev = float64(int(variance*1e8)) / 1e8
			if stdDev > 0 {
				sharpe = mean / stdDev * float64(252)
			}
		}
	}

	equityJSON, _ := json.Marshal(equityCurve)

	bt := &models.Backtest{
		StrategyID:     strategyID,
		Symbol:         config.Symbol,
		Interval:       config.Interval,
		StartTime:      time.Unix(klines[0].Time, 0),
		EndTime:        time.Unix(klines[len(klines)-1].Time, 0),
		InitialCapital: initialCapital,
		FinalCapital:   finalEquity,
		TotalTrades:    totalTrades,
		WinRate:        winRate,
		Sharpe:         sharpe,
		MaxDrawdown:    maxDrawdown,
		EquityCurveJSON: string(equityJSON),
	}

	fmt.Println("Backtest complete:", bt.Symbol, bt.Interval, "trades:", totalTrades, "win%:", winRate)

	return bt, nil
}

func evaluateRules(rule models.EntryExitRule, indValues map[string]interface{}, klines []models.Kline, index int) bool {
	if len(rule.Conditions) == 0 {
		return false
	}

	results := make([]bool, len(rule.Conditions))
	for ci, cond := range rule.Conditions {
		val := getIndicatorValue(indValues, cond.Indicator, cond.Field, index)
		results[ci] = compareValue(val, cond.Operator, cond.Value)
	}

	if rule.Logic == "or" {
		for _, r := range results {
			if r {
				return true
			}
		}
		return false
	}

	// Default: "and"
	for _, r := range results {
		if !r {
			return false
		}
	}
	return true
}

func getIndicatorValue(indValues map[string]interface{}, name, field string, index int) float64 {
	raw, ok := indValues[name]
	if !ok {
		// Try direct field name (e.g., KDJ_K)
		raw, ok = indValues[name+"_"+field]
		if !ok {
			return 0
		}
	}

	switch v := raw.(type) {
	case []float64:
		if index < len(v) {
			return v[index]
		}
	case *indicators.MACDResult:
		switch field {
		case "macd":
			if index < len(v.MACD) {
				return v.MACD[index]
			}
		case "signal":
			if index < len(v.Signal) {
				return v.Signal[index]
			}
		case "hist":
			if index < len(v.Hist) {
				return v.Hist[index]
			}
		}
	case *indicators.BBResult:
		switch field {
		case "upper":
			if index < len(v.Upper) {
				return v.Upper[index]
			}
		case "middle":
			if index < len(v.Middle) {
				return v.Middle[index]
			}
		case "lower":
			if index < len(v.Lower) {
				return v.Lower[index]
			}
		}
	}
	return 0
}

func compareValue(a float64, op string, b float64) bool {
	switch op {
	case ">":
		return a > b
	case "<":
		return a < b
	case ">=":
		return a >= b
	case "<=":
		return a <= b
	case "==":
		return a == b
	case "cross_above":
		return a > b
	case "cross_below":
		return a < b
	}
	return false
}

