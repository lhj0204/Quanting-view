package exchange

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"quant/internal/models"

	"github.com/gorilla/websocket"
)

type okxClient struct {
	restURL    string
	wsURL      string
	apiKey     string
	secretKey  string
	passphrase string
}

func newOKXClient(restURL, wsURL, apiKey, secretKey, passphrase string) *okxClient {
	return &okxClient{
		restURL:    restURL,
		wsURL:      wsURL,
		apiKey:     apiKey,
		secretKey:  secretKey,
		passphrase: passphrase,
	}
}

func (c *okxClient) Name() string { return "okx" }

func (c *okxClient) GetKlines(ctx context.Context, symbol, interval string, limit int, startTime, endTime int64) ([]models.Kline, error) {
	instID := formatOKXSymbol(symbol)
	bar := formatOKXInterval(interval)
	url := fmt.Sprintf("%s/api/v5/market/candles?instId=%s&bar=%s&limit=%d",
		c.restURL, instID, bar, limit)

	resp, err := httpGET(url)
	if err != nil {
		return nil, err
	}

	var result struct {
		Code string     `json:"code"`
		Data [][]string `json:"data"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	if result.Code != "0" {
		return nil, fmt.Errorf("okx error: %s", result.Code)
	}

	klines := make([]models.Kline, 0, len(result.Data))
	for i := len(result.Data) - 1; i >= 0; i-- {
		r := result.Data[i]
		k := models.Kline{
			Time:   parseStrI64(r[0]) / 1000,
			Open:   parseStrFloat(r[1]),
			High:   parseStrFloat(r[2]),
			Low:    parseStrFloat(r[3]),
			Close:  parseStrFloat(r[4]),
			Volume: parseStrFloat(r[5]),
		}
		klines = append(klines, k)
	}
	return klines, nil
}

func (c *okxClient) GetTicker(ctx context.Context, symbol string) (*models.Ticker, error) {
	instID := formatOKXSymbol(symbol)
	url := fmt.Sprintf("%s/api/v5/market/ticker?instId=%s", c.restURL, instID)

	resp, err := httpGET(url)
	if err != nil {
		return nil, err
	}

	var result struct {
		Code string `json:"code"`
		Data []struct {
			InstID  string `json:"instId"`
			Last    string `json:"last"`
			High24h string `json:"high24h"`
			Low24h  string `json:"low24h"`
			Vol24h  string `json:"vol24h"`
		} `json:"data"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	if len(result.Data) == 0 {
		return nil, fmt.Errorf("okx: no ticker data for %s", symbol)
	}

	d := result.Data[0]
	return &models.Ticker{
		Symbol:    symbol,
		Price:     parseStrFloat(d.Last),
		High24h:   parseStrFloat(d.High24h),
		Low24h:    parseStrFloat(d.Low24h),
		Volume24h: parseStrFloat(d.Vol24h),
		Exchange:  "okx",
	}, nil
}

func (c *okxClient) GetDepth(ctx context.Context, symbol string, limit int) (*models.DepthData, error) {
	instID := formatOKXSymbol(symbol)
	url := fmt.Sprintf("%s/api/v5/market/books?instId=%s&sz=%d", c.restURL, instID, limit)

	resp, err := httpGET(url)
	if err != nil {
		return nil, err
	}

	var result struct {
		Code string `json:"code"`
		Data []struct {
			Bids [][]string `json:"bids"`
			Asks [][]string `json:"asks"`
		} `json:"data"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	if len(result.Data) == 0 {
		return nil, fmt.Errorf("okx: no depth data for %s", symbol)
	}

	d := result.Data[0]
	depth := &models.DepthData{
		Bids: make([][2]string, len(d.Bids)),
		Asks: make([][2]string, len(d.Asks)),
	}
	for i, b := range d.Bids {
		if len(b) >= 2 {
			depth.Bids[i] = [2]string{b[0], b[1]}
		}
	}
	for i, a := range d.Asks {
		if len(a) >= 2 {
			depth.Asks[i] = [2]string{a[0], a[1]}
		}
	}
	return depth, nil
}

func (c *okxClient) ConnectWebSocket(symbols []string, onKline func(kline *models.Kline)) (chan struct{}, error) {
	conn, _, err := wsDialer.Dial(c.wsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("okx ws dial: %w", err)
	}

	stopCh := make(chan struct{})

	// Subscribe to kline channels
	for _, sym := range symbols {
		instID := formatOKXSymbol(sym)
		subMsg := map[string]interface{}{
			"op": "subscribe",
			"args": []map[string]string{
				{"channel": "candle1H", "instId": instID},
			},
		}
		if err := conn.WriteJSON(subMsg); err != nil {
			conn.Close()
			return nil, err
		}
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
					Event string `json:"event"`
					Arg   struct {
						Channel string `json:"channel"`
						InstID  string `json:"instId"`
					} `json:"arg"`
					Data [][]string `json:"data"`
				}
				if err := json.Unmarshal(msg, &event); err != nil {
					continue
				}
				if event.Event != "" || len(event.Data) == 0 {
					continue
				}
				// OKX candle format: [ts,o,h,l,c,vol,volCcy,...]
				for _, d := range event.Data {
					if len(d) < 6 {
						continue
					}
					onKline(&models.Kline{
						Time:   parseStrI64(d[0]) / 1000,
						Open:   parseStrFloat(d[1]),
						High:   parseStrFloat(d[2]),
						Low:    parseStrFloat(d[3]),
						Close:  parseStrFloat(d[4]),
						Volume: parseStrFloat(d[5]),
					})
				}
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

func (c *okxClient) Ping() error {
	_, err := httpGET(fmt.Sprintf("%s/api/v5/public/time", c.restURL))
	return err
}

func (c *okxClient) sign(method, path, body string, timestamp string) (string, string) {
	preSign := timestamp + method + path + body
	mac := hmac.New(sha256.New, []byte(c.secretKey))
	mac.Write([]byte(preSign))
	sign := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	return sign, timestamp
}

func formatOKXSymbol(symbol string) string {
	return strings.Replace(symbol, "USDT", "-USDT-SWAP", 1)
}

func formatOKXInterval(interval string) string {
	m := map[string]string{
		"1m": "1m", "5m": "5m", "15m": "15m", "30m": "30m",
		"1h": "1H", "4h": "4H", "1d": "1D", "1w": "1W",
	}
	if v, ok := m[interval]; ok {
		return v
	}
	return "1H"
}

func parseStrI64(s string) int64 {
	var i int64
	fmt.Sscanf(s, "%d", &i)
	return i
}

