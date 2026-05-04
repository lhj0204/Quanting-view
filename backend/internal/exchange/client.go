package exchange

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"quant/internal/models"

	"golang.org/x/net/proxy"
	"github.com/gorilla/websocket"
)

type Client interface {
	Name() string
	GetKlines(ctx context.Context, symbol, interval string, limit int, startTime, endTime int64) ([]models.Kline, error)
	GetTicker(ctx context.Context, symbol string) (*models.Ticker, error)
	GetDepth(ctx context.Context, symbol string, limit int) (*models.DepthData, error)
	ConnectWebSocket(symbols []string, onKline func(kline *models.Kline)) (chan struct{}, error)
	Ping() error
}

var httpClient = &http.Client{Timeout: 10 * time.Second}
var wsDialer = &websocket.Dialer{HandshakeTimeout: 10 * time.Second}

func SetProxy(proxyURL string) error {
	if proxyURL == "" {
		return nil
	}

	u, err := url.Parse(proxyURL)
	if err != nil {
		return fmt.Errorf("invalid proxy URL: %w", err)
	}

	switch u.Scheme {
	case "http", "https":
		httpClient.Transport = &http.Transport{
			Proxy: http.ProxyURL(u),
		}
	case "socks5":
		dialer, err := proxy.SOCKS5("tcp", u.Host, nil, proxy.Direct)
		if err != nil {
			return fmt.Errorf("socks5 proxy error: %w", err)
		}
		httpClient.Transport = &http.Transport{
			Dial: dialer.Dial,
		}
		wsDialer.NetDial = dialer.Dial
	}

	return nil
}

func NewClient(exchange string, restURL, wsURL string, apiKey, secretKey, passphrase string) (Client, error) {
	switch exchange {
	case "binance":
		return newBinanceClient(restURL, wsURL, apiKey, secretKey), nil
	case "okx":
		return newOKXClient(restURL, wsURL, apiKey, secretKey, passphrase), nil
	default:
		return nil, fmt.Errorf("unsupported exchange: %s", exchange)
	}
}
