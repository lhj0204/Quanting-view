package indicators

import (
	"math"
	"quant/internal/models"
)

func ComputeEMA(data []models.Kline, period int) []float64 {
	if len(data) < period {
		return make([]float64, len(data))
	}
	result := make([]float64, len(data))
	k := 2.0 / float64(period+1)

	// SMA for first EMA value
	sum := 0.0
	for i := 0; i < period; i++ {
		sum += data[i].Close
	}
	result[period-1] = sum / float64(period)

	for i := period; i < len(data); i++ {
		result[i] = data[i].Close*k + result[i-1]*(1-k)
	}
	return result
}

func ComputeSMA(data []models.Kline, period int) []float64 {
	result := make([]float64, len(data))
	for i := period - 1; i < len(data); i++ {
		sum := 0.0
		for j := i - period + 1; j <= i; j++ {
			sum += data[j].Close
		}
		result[i] = sum / float64(period)
	}
	return result
}

func ComputeRSI(data []models.Kline, period int) []float64 {
	if len(data) < period+1 {
		return make([]float64, len(data))
	}
	result := make([]float64, len(data))

	gains := 0.0
	losses := 0.0

	for i := 1; i <= period; i++ {
		diff := data[i].Close - data[i-1].Close
		if diff > 0 {
			gains += diff
		} else {
			losses -= diff
		}
	}

	avgGain := gains / float64(period)
	avgLoss := losses / float64(period)

	if avgLoss == 0 {
		result[period] = 100
	} else {
		rs := avgGain / avgLoss
		result[period] = 100 - (100 / (1 + rs))
	}

	for i := period + 1; i < len(data); i++ {
		diff := data[i].Close - data[i-1].Close
		gain := 0.0
		loss := 0.0
		if diff > 0 {
			gain = diff
		} else {
			loss = -diff
		}
		avgGain = (avgGain*float64(period-1) + gain) / float64(period)
		avgLoss = (avgLoss*float64(period-1) + loss) / float64(period)

		if avgLoss == 0 {
			result[i] = 100
		} else {
			rs := avgGain / avgLoss
			result[i] = 100 - (100 / (1 + rs))
		}
	}
	return result
}

type MACDResult struct {
	MACD  []float64
	Signal []float64
	Hist  []float64
}

func ComputeMACD(data []models.Kline, fast, slow, signal int) *MACDResult {
	n := len(data)
	emaFast := ComputeEMA(data, fast)
	emaSlow := ComputeEMA(data, slow)

	macd := make([]float64, n)
	for i := 0; i < n; i++ {
		macd[i] = emaFast[i] - emaSlow[i]
	}

	// Compute signal line from MACD values using a dummy Kline series
	dummy := make([]models.Kline, n)
	for i := 0; i < n; i++ {
		dummy[i] = models.Kline{Close: macd[i]}
	}
	signalLine := ComputeEMA(dummy, signal)

	hist := make([]float64, n)
	for i := 0; i < n; i++ {
		hist[i] = macd[i] - signalLine[i]
	}

	return &MACDResult{
		MACD:  macd,
		Signal: signalLine,
		Hist:  hist,
	}
}

type BBResult struct {
	Upper  []float64
	Middle []float64
	Lower  []float64
}

func ComputeBollingerBands(data []models.Kline, period int, multiplier float64) *BBResult {
	n := len(data)
	sma := ComputeSMA(data, period)
	upper := make([]float64, n)
	lower := make([]float64, n)

	for i := period - 1; i < n; i++ {
		sumSq := 0.0
		for j := i - period + 1; j <= i; j++ {
			diff := data[j].Close - sma[i]
			sumSq += diff * diff
		}
		stdDev := math.Sqrt(sumSq / float64(period))
		upper[i] = sma[i] + multiplier*stdDev
		lower[i] = sma[i] - multiplier*stdDev
	}

	return &BBResult{
		Upper:  upper,
		Middle: sma,
		Lower:  lower,
	}
}

func ComputeATR(data []models.Kline, period int) []float64 {
	n := len(data)
	tr := make([]float64, n)
	for i := 1; i < n; i++ {
		h := data[i].High
		l := data[i].Low
		pc := data[i-1].Close
		tr[i] = math.Max(h-l, math.Max(math.Abs(h-pc), math.Abs(l-pc)))
	}

	atr := make([]float64, n)
	if n > period {
		sum := 0.0
		for i := 1; i <= period; i++ {
			sum += tr[i]
		}
		atr[period] = sum / float64(period)
		for i := period + 1; i < n; i++ {
			atr[i] = (atr[i-1]*float64(period-1) + tr[i]) / float64(period)
		}
	}
	return atr
}

func ComputeOBV(data []models.Kline) []float64 {
	n := len(data)
	obv := make([]float64, n)
	for i := 1; i < n; i++ {
		if data[i].Close > data[i-1].Close {
			obv[i] = obv[i-1] + data[i].Volume
		} else if data[i].Close < data[i-1].Close {
			obv[i] = obv[i-1] - data[i].Volume
		} else {
			obv[i] = obv[i-1]
		}
	}
	return obv
}

func ComputeKDJ(data []models.Kline, period, smoothPeriod1, smoothPeriod2 int) ([]float64, []float64, []float64) {
	n := len(data)
	k := make([]float64, n)
	d := make([]float64, n)
	j := make([]float64, n)

	for i := period - 1; i < n; i++ {
		high := data[i].High
		low := data[i].Low
		for j2 := i - period + 1; j2 <= i; j2++ {
			if data[j2].High > high {
				high = data[j2].High
			}
			if data[j2].Low < low {
				low = data[j2].Low
			}
		}
		if high != low {
			rsv := (data[i].Close - low) / (high - low) * 100
			if i == period-1 {
				k[i] = 50
				d[i] = 50
			}
			k[i] = (rsv + float64(smoothPeriod1-1)*k[i-1]) / float64(smoothPeriod1)
			d[i] = (k[i] + float64(smoothPeriod2-1)*d[i-1]) / float64(smoothPeriod2)
		} else {
			if i > 0 {
				k[i] = k[i-1]
				d[i] = d[i-1]
			}
		}
		j[i] = 3*k[i] - 2*d[i]
	}
	return k, d, j
}
