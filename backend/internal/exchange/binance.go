package exchange

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"

	"quant/internal/models"

	"github.com/gorilla/websocket"
)

type binanceClient struct {
	restURL   string
	wsURL     string
	apiKey    string
	secretKey string
}

func newBinanceClient(restURL, wsURL, apiKey, secretKey string) *binanceClient {
	return &binanceClient{
		restURL:   restURL,
		wsURL:     wsURL,
		apiKey:    apiKey,
		secretKey: secretKey,
	}
}

func (c *binanceClient) Name() string { return "binance" }

func (c *binanceClient) GetKlines(ctx context.Context, symbol, interval string, limit int, startTime, endTime int64) ([]models.Kline, error) {
	url := fmt.Sprintf("%s/api/v3/klines?symbol=%s&interval=%s&limit=%d",
		c.restURL, symbol, interval, limit)
	if startTime > 0 {
		url += fmt.Sprintf("&startTime=%d", startTime)
	}
	if endTime > 0 {
		url += fmt.Sprintf("&endTime=%d", endTime)
	}

	resp, err := httpGET(url)
	if err != nil {
		return nil, err
	}

	var raw [][]interface{}
	if err := json.Unmarshal(resp, &raw); err != nil {
		return nil, err
	}

	klines := make([]models.Kline, 0, len(raw))
	for _, r := range raw {
		k := models.Kline{
			Time:   int64(r[0].(float64)) / 1000,
			Open:   parseFloat(r[1]),
			High:   parseFloat(r[2]),
			Low:    parseFloat(r[3]),
			Close:  parseFloat(r[4]),
			Volume: parseFloat(r[5]),
		}
		klines = append(klines, k)
	}
	return klines, nil
}

func (c *binanceClient) GetTicker(ctx context.Context, symbol string) (*models.Ticker, error) {
	priceURL := fmt.Sprintf("%s/api/v3/ticker/price?symbol=%s", c.restURL, symbol)
	priceResp, err := httpGET(priceURL)
	if err != nil {
		return nil, err
	}
	var priceData struct {
		Symbol string `json:"symbol"`
		Price  string `json:"price"`
	}
	if err := json.Unmarshal(priceResp, &priceData); err != nil {
		return nil, err
	}

	statURL := fmt.Sprintf("%s/api/v3/ticker/24hr?symbol=%s", c.restURL, symbol)
	statResp, err := httpGET(statURL)
	if err != nil {
		return nil, err
	}
	var statData struct {
		PriceChange string `json:"priceChange"`
		PriceChangePercent string `json:"priceChangePercent"`
		HighPrice   string `json:"highPrice"`
		LowPrice    string `json:"lowPrice"`
		Volume      string `json:"volume"`
	}
	if err := json.Unmarshal(statResp, &statData); err != nil {
		return nil, err
	}

	return &models.Ticker{
		Symbol:    symbol,
		Price:     parseStrFloat(priceData.Price),
		Change24h: parseStrFloat(statData.PriceChange),
		ChangePct: parseStrFloat(statData.PriceChangePercent),
		High24h:   parseStrFloat(statData.HighPrice),
		Low24h:    parseStrFloat(statData.LowPrice),
		Volume24h: parseStrFloat(statData.Volume),
		Exchange:  "binance",
	}, nil
}

func (c *binanceClient) GetDepth(ctx context.Context, symbol string, limit int) (*models.DepthData, error) {
	url := fmt.Sprintf("%s/api/v3/depth?symbol=%s&limit=%d", c.restURL, symbol, limit)
	resp, err := httpGET(url)
	if err != nil {
		return nil, err
	}
	var data struct {
		Bids [][]string `json:"bids"`
		Asks [][]string `json:"asks"`
	}
	if err := json.Unmarshal(resp, &data); err != nil {
		return nil, err
	}

	depth := &models.DepthData{
		Bids: make([][2]string, len(data.Bids)),
		Asks: make([][2]string, len(data.Asks)),
	}
	for i, b := range data.Bids {
		if len(b) >= 2 {
			depth.Bids[i] = [2]string{b[0], b[1]}
		}
	}
	for i, a := range data.Asks {
		if len(a) >= 2 {
			depth.Asks[i] = [2]string{a[0], a[1]}
		}
	}
	return depth, nil
}

func (c *binanceClient) ConnectWebSocket(symbols []string, onKline func(kline *models.Kline)) (chan struct{}, error) {
	stopCh := make(chan struct{})

	var streams []string
	for _, sym := range symbols {
		s := strings.ToLower(sym)
		streams = append(streams, fmt.Sprintf("%s@kline_1h", s))
	}
	url := fmt.Sprintf("%s/%s", c.wsURL, strings.Join(streams, "/"))

	conn, _, err := wsDialer.Dial(url, nil)
	if err != nil {
		return nil, fmt.Errorf("binance ws dial: %w", err)
	}

	go func() {
		defer conn.Close()
		done := make(chan struct{})
		go func() {
			defer close(done)
			for {
				_, msg, err := conn.ReadMessage()
				if err != nil {
					return
				}
				var event struct {
					Stream string `json:"stream"`
					Data   struct {
						Kline struct {
							StartTime int64  `json:"t"`
							Open      string `json:"o"`
							High      string `json:"h"`
							Low       string `json:"l"`
							Close     string `json:"c"`
							Volume    string `json:"v"`
							Closed    bool   `json:"x"`
						} `json:"k"`
					} `json:"data"`
				}
				if err := json.Unmarshal(msg, &event); err != nil {
					continue
				}
				onKline(&models.Kline{
					Time:   event.Data.Kline.StartTime / 1000,
					Open:   parseStrFloat(event.Data.Kline.Open),
					High:   parseStrFloat(event.Data.Kline.High),
					Low:    parseStrFloat(event.Data.Kline.Low),
					Close:  parseStrFloat(event.Data.Kline.Close),
					Volume: parseStrFloat(event.Data.Kline.Volume),
				})
			}
		}()

		select {
		case <-stopCh:
			conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		case <-done:
		}
	}()

	return stopCh, nil
}

func (c *binanceClient) Ping() error {
	_, err := httpGET(fmt.Sprintf("%s/api/v3/ping", c.restURL))
	return err
}

var wsConnections sync.Map

func httpGET(url string) ([]byte, error) {
	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func parseFloat(v interface{}) float64 {
	s, ok := v.(string)
	if ok {
		return parseStrFloat(s)
	}
	f, _ := v.(float64)
	return f
}

func parseStrFloat(s string) float64 {
	var f float64
	fmt.Sscanf(s, "%f", &f)
	return f
}
