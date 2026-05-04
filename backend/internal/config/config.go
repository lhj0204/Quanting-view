package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server       ServerConfig       `yaml:"server"`
	Database     DatabaseConfig     `yaml:"database"`
	Proxy        ProxyConfig        `yaml:"proxy"`
	Exchange     ExchangeConfig     `yaml:"exchange"`
	PaperTrading PaperTradingConfig `yaml:"paper_trading"`
	Risk         RiskConfig         `yaml:"risk"`
}

type ProxyConfig struct {
	URL string `yaml:"url"`
}

type ServerConfig struct {
	Port string `yaml:"port"`
	Host string `yaml:"host"`
}

type DatabaseConfig struct {
	Driver string `yaml:"driver"`
	DSN    string `yaml:"dsn"`
}

type ExchangeConfig struct {
	Binance ExchangeItem `yaml:"binance"`
	OKX     ExchangeItem `yaml:"okx"`
}

type ExchangeItem struct {
	Testnet bool   `yaml:"testnet"`
	WSUrl   string `yaml:"ws_url"`
	RestURL string `yaml:"rest_url"`
}

type PaperTradingConfig struct {
	InitialBalance float64 `yaml:"initial_balance"`
	Currency       string  `yaml:"currency"`
}

type RiskConfig struct {
	DefaultMaxPositionPct  float64 `yaml:"default_max_position_pct"`
	DefaultStopLossPct     float64 `yaml:"default_stop_loss_pct"`
	DefaultTakeProfitPct   float64 `yaml:"default_take_profit_pct"`
	DefaultMaxDrawdownPct  float64 `yaml:"default_max_drawdown_pct"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	if cfg.Server.Port == "" {
		cfg.Server.Port = "8080"
	}
	if cfg.Server.Host == "" {
		cfg.Server.Host = "localhost"
	}
	return cfg, nil
}
