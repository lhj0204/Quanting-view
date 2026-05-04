# Quanting-view
基于Goland和Vue量化交易系统
# Quanting-view

基于 Golang 和 Vue 量化交易系统

---

## 项目简介

Quanting-view 是一套全栈加密货币量化交易平台，后端采用 Go (Gin) 构建高性能 API 服务，前端采用 Vue 3 + TypeScript 构建现代化交易界面。系统集成 TradingView 开源图表库 lightweight-charts，支持实时 K 线展示、8 种技术指标叠加、策略回测引擎、模拟盘/实盘双模式自动交易。

## 技术栈

| 层级 | 技术 | 版本 | 说明 |
|------|------|------|------|
| 前端框架 | Vue 3 + TypeScript | ^3.4 | Composition API + `<script setup>` |
| 构建工具 | Vite | ^5.4 | 极速 HMR |
| UI 样式 | Tailwind CSS | ^3.4 | 深色交易主题 |
| 状态管理 | Pinia | ^2.1 | Vue 3 官方推荐 |
| 路由 | Vue Router | ^4.3 | 懒加载路由 |
| 图表 | lightweight-charts | ^4.2 | TradingView 开源版 (Apache 2.0) |
| 后端框架 | Gin | ^1.12 | 高性能 HTTP |
| ORM | Bun + SQLite | ^1.2 | 支持无缝升级 PostgreSQL |
| WebSocket | gorilla/websocket | ^1.5 | 实时行情推送 |
| 交易所 | go-binance/v2 + okx-go | — | Binance + OKX 双交易所 |
| 任务调度 | robfig/cron/v3 | ^3.0 | 策略定时执行 |

## 项目结构

```
quant/
├── backend/                               # Go 后端
│   ├── cmd/server/main.go                 # 入口：路由注册 + 策略执行器
│   ├── internal/
│   │   ├── config/config.go               # YAML 配置解析
│   │   ├── database/database.go           # SQLite 初始化 + 自动建表 + 种子数据
│   │   ├── models/models.go               # 9 张表 ORM 模型 + API 响应结构体
│   │   ├── handlers/handlers.go           # 全部 REST API (账户/行情/策略/订单/风控)
│   │   ├── exchange/
│   │   │   ├── client.go                  # 统一交易接口 + HTTP/SOCKS5 代理配置
│   │   │   ├── binance.go                 # Binance REST + WebSocket 客户端
│   │   │   └── okx.go                     # OKX REST + WebSocket 客户端
│   │   ├── indicators/indicators.go       # 8 种技术指标纯 Go 实现（无 CGO 依赖）
│   │   ├── ws/hub.go                      # WebSocket Hub：连接管理/订阅/广播
│   │   └── services/
│   │       ├── services.go                # 回测引擎 + 指标计算服务
│   │       └── executor.go                # 策略执行器：cron 调度 + 止损止盈
│   ├── config.yaml                        # 系统配置（代理/交易所/风控参数）
│   ├── go.mod
│   └── go.sum
├── frontend/                              # Vue 3 前端
│   ├── index.html
│   ├── vite.config.ts                     # Vite 配置 + API/WS 代理
│   ├── tailwind.config.js                 # 深色主题定制
│   ├── src/
│   │   ├── main.ts                        # 入口：挂载 Vue + Pinia + Router
│   │   ├── App.vue
│   │   ├── style.css                      # Tailwind + 全局样式 + 自定义组件类
│   │   ├── api/index.ts                   # Axios 封装：全部 20+ 接口
│   │   ├── router/index.ts                # 10 条路由（全部懒加载）
│   │   ├── types/index.ts                 # 20+ TypeScript 接口定义
│   │   ├── stores/                        # Pinia 状态管理
│   │   │   ├── market.ts                  # 行情数据（K 线/Ticker/盘口）
│   │   │   ├── strategy.ts                # 策略 CRUD + 启动/回测
│   │   │   └── account.ts                 # 账户/订单/持仓/成交
│   │   ├── pages/                         # 10 个页面组件
│   │   │   ├── Dashboard.vue              # 仪表盘
│   │   │   ├── Market.vue                 # 行情列表
│   │   │   ├── Chart.vue                  # K 线图 + 下单面板
│   │   │   ├── StrategyList.vue           # 策略列表
│   │   │   ├── StrategyEdit.vue           # 策略编辑器
│   │   │   ├── BacktestList.vue           # 回测记录列表
│   │   │   ├── BacktestDetail.vue         # 回测详情 + 权益曲线
│   │   │   ├── Orders.vue                 # 订单管理
│   │   │   ├── RiskManager.vue            # 风控规则编辑
│   │   │   └── Settings.vue               # 交易所 API 密钥管理
│   │   └── components/                    # 可复用组件
│   │       ├── layout/AppLayout.vue       # 主布局：侧边栏 + TickerBar + 内容区
│   │       ├── market/TickerBar.vue        # 顶部滚动行情条
│   │       ├── market/OrderBook.vue        # 盘口深度组件
│   │       └── common/MetricsCard.vue      # 统计指标卡片
│   ├── package.json
│   └── tsconfig.json
└── README.md
```

## 核心功能

### 1. 仪表盘
- **账户概览**：余额、总盈亏、盈亏百分比、持仓数量、活跃策略数
- **最近交易**：最近 10 笔成交记录，含方向/价格/盈亏
- **当前持仓**：币种/数量/均价
- 页面加载时自动从后端拉取所有数据

### 2. 行情市场
- **双交易所切换**：Binance / OKX 一键切换
- **实时价格**：BTC、ETH、BNB、SOL、XRP、ADA、DOGE、AVAX、DOT、LINK、MATIC、UNI
- **24h 数据**：涨跌幅、最高价、最低价、成交量
- **搜索过滤**：按交易对名称实时筛选
- 点击任意行跳转到 K 线图页面

### 3. K 线图表 + 技术指标

基于 lightweight-charts v5，纯前端渲染，`~35KB` gzipped。

- **蜡烛图**：实时 WebSocket 更新，`series.update()` 增量合并
- **成交量**：底部直方图，红涨绿跌
- **MA 均线**：MA7（黄色）+ MA25（红色）叠加在主图
- **时间周期**：1m / 5m / 15m / 30m / 1h / 4h / 1d
- **交易所切换**：图表页面内切换 Binance/OKX 数据源
- **右侧面板**：
  - 买入/卖出下单（市价/限价）
  - 盘口深度实时展示

### 4. 技术指标（8 种，纯 Go 实现）

| 指标 | 默认参数 | 用途 |
|------|---------|------|
| EMA | period=20 | 趋势跟踪（快线） |
| SMA | period=20 | 趋势跟踪（慢线） |
| RSI | period=14 | 超买超卖（>70 / <30） |
| MACD | fast=12, slow=26, signal=9 | 金叉死叉信号 |
| Bollinger Bands | period=20, multiplier=2 | 波动率区间 |
| KDJ | period=9, smooth=3 | 短期超买超卖 |
| ATR | period=14 | 波动率测量（止损参考） |
| OBV | — | 量价关系验证 |

### 5. 策略引擎

**策略配置**（JSON 格式）：
```json
{
  "symbol": "BTCUSDT",
  "interval": "1h",
  "indicators": [
    { "name": "EMA", "params": { "period": 12 } },
    { "name": "EMA", "params": { "period": 26 } },
    { "name": "RSI", "params": { "period": 14 } }
  ],
  "entry_rule": {
    "logic": "and",
    "conditions": [
      { "indicator": "RSI", "field": "value", "operator": "<=", "value": 30 }
    ]
  },
  "exit_rule": {
    "logic": "or",
    "conditions": [
      { "indicator": "RSI", "field": "value", "operator": ">=", "value": 70 }
    ]
  }
}
```

- **条件逻辑**：支持 `and` / `or` 组合多个条件
- **操作符**：`>` `<` `>=` `<=` `==` `cross_above` `cross_below`
- **策略生命周期**：草稿 → 运行中 → 已暂停 → 已停止

### 6. 回测引擎

- 加载历史 K 线（500 根），逐根评估入场/出场条件
- 每次回测初始资金 10,000 USDT
- 输出指标：
  - 初始资金 / 最终资金 / 总收益率
  - 交易次数 / 胜率
  - 夏普比率（年化）
  - 最大回撤（%）
  - 权益曲线（JSON → 前端图表渲染）

### 7. 自动交易

**模拟盘（默认）**：
- 初始资金 10,000 USDT
- 即时模拟成交
- 完整的订单/成交/持仓/账户余额追踪

**实盘**：
- 在「系统设置」页面配置交易所 API Key
- 策略选择 `live` 模式后，提交真实订单到 Binance/OKX

**风控**：
- 止损：按百分比自动平仓
- 止盈：达到目标利润自动平仓
- 仓位限制：最大持仓占总资金比例
- 最大回撤：超过阈值暂停策略

**调度**：`cron` 每 30 秒遍历所有活跃策略，拉取最新 K 线，评估指标条件，触发生成交易信号。

## 数据库设计（SQLite，9 张表）

| 表名 | 说明 | 关键字段 |
|------|------|---------|
| `exchange_keys` | 交易所 API 密钥 | exchange, api_key, secret_key, testnet |
| `strategies` | 量化策略 | name, config_json, trade_mode, status |
| `backtests` | 回测记录 | strategy_id, initial_capital, final_capital, win_rate, sharpe, max_drawdown, equity_curve_json |
| `orders` | 交易订单 | strategy_id, symbol, side, type, price, qty, status, mode |
| `trades` | 成交记录 | order_id, price, qty, fee, realized_pnl |
| `positions` | 当前持仓 | strategy_id, symbol, qty, avg_entry_price |
| `risk_rules` | 风控规则 | strategy_id, max_position_pct, stop_loss_pct, take_profit_pct, max_drawdown_pct |
| `accounts` | 账户余额 | balance, initial_balance, currency |
| `market_data_cache` | 行情缓存 | symbol, exchange, interval, kline_json |

表由 Bun ORM 在启动时自动创建（`CreateTable().IfNotExists()`），无需手动执行 SQL。

## API 文档

所有接口前缀 `/api`，返回 JSON。

### 账户
| 方法 | 路径 | 响应 |
|------|------|------|
| `GET` | `/api/account/summary` | `{ balance, total_pnl, positions, ... }` |

### 交易所密钥管理
| 方法 | 路径 | 说明 |
|------|------|------|
| `GET` | `/api/exchange/keys` | 列出所有密钥 |
| `POST` | `/api/exchange/keys` | 添加密钥（body: ExchangeKey） |
| `DELETE` | `/api/exchange/keys/:id` | 删除密钥 |

### 市场行情
| 方法 | 路径 | 查询参数 |
|------|------|---------|
| `GET` | `/api/market/klines` | `exchange`, `symbol`, `interval`, `limit`, `start`, `end` |
| `GET` | `/api/market/ticker` | `exchange`, `symbol` |
| `GET` | `/api/market/depth` | `exchange`, `symbol`, `limit` |

### 策略管理
| 方法 | 路径 | 说明 |
|------|------|------|
| `GET` | `/api/strategies` | 策略列表 |
| `POST` | `/api/strategies` | 创建策略 |
| `GET` | `/api/strategies/:id` | 策略详情 |
| `PUT` | `/api/strategies/:id` | 更新策略 |
| `DELETE` | `/api/strategies/:id` | 删除策略 |
| `POST` | `/api/strategies/:id/activate` | 激活策略 |
| `POST` | `/api/strategies/:id/backtest` | 执行回测 |

### 回测记录
| 方法 | 路径 | 说明 |
|------|------|------|
| `GET` | `/api/backtests` | 回测列表（支持 `?strategy_id=` 筛选） |
| `GET` | `/api/backtests/:id` | 回测详情（含 `equity_curve_json`） |

### 订单与成交
| 方法 | 路径 | 说明 |
|------|------|------|
| `GET` | `/api/orders` | 订单列表（`?status=open&mode=paper`） |
| `POST` | `/api/orders` | 手动下单 |
| `DELETE` | `/api/orders/:id` | 取消未成交订单 |
| `GET` | `/api/trades` | 成交记录 |
| `GET` | `/api/positions` | 当前持仓（`?strategy_id=`） |

### 风控
| 方法 | 路径 | 说明 |
|------|------|------|
| `GET` | `/api/risk/:strategy_id` | 获取风控规则 |
| `PUT` | `/api/risk/:strategy_id` | 更新风控规则 |

### WebSocket

端点：`ws://localhost:8080/ws`

**订阅：**
```json
{ "event": "subscribe", "channel": "kline",  "symbol": "BTCUSDT", "interval": "1h" }
{ "event": "subscribe", "channel": "ticker", "symbol": "BTCUSDT" }
{ "event": "subscribe", "channel": "depth",  "symbol": "BTCUSDT" }
{ "event": "subscribe", "channel": "orders" }
```

**取消订阅：**
```json
{ "event": "unsubscribe", "channel": "kline", "symbol": "BTCUSDT" }
```

**服务端推送格式：**
```json
{ "channel": "kline",  "data": { "symbol": "BTCUSDT", "time": 1714000000, "open": 80000, "high": 81000, "low": 79000, "close": 80500, "volume": 1234.5 } }
{ "channel": "ticker", "data": { "symbol": "BTCUSDT", "price": 80500, "change_24h": 500 } }
{ "channel": "order",  "data": { "id": 1, "status": "filled", "filled_qty": 0.01 } }
```

## 快速开始

### 环境要求

- **Go** 1.21+
- **Node.js** 18+
- **代理软件**（中国大陆用户必需）：Clash / v2ray / SSR 等

### 第一步：配置代理

编辑 `backend/config.yaml`：

```yaml
proxy:
  url: "http://127.0.0.1:7890"       # Clash 等 HTTP 代理
  # url: "socks5://127.0.0.1:1080"   # SOCKS5 代理
```

> 如果代理软件在本机运行，默认地址通常是 `http://127.0.0.1:7890`（Clash）或 `http://127.0.0.1:10809`（v2ray）。代理地址为空则跳过交易所连接。

### 第二步：安装依赖

```bash
# 后端
cd backend
go mod tidy

# 前端
cd frontend
npm install
```

### 第三步：启动后端

```bash
cd backend
go run cmd/server/main.go
```

看到以下日志表示启动成功：

```
Database initialized successfully
Proxy configured: http://127.0.0.1:7890
Binance client connected
OKX client connected
[Executor] Strategy executor started
Server starting on localhost:8080
Exchanges available: [binance okx]
```

### 第四步：启动前端

```bash
cd frontend
npm run dev
```

浏览器打开 Vite 输出的地址（默认 `http://localhost:5173`）。

### 第五步：使用

1. 打开仪表盘 — 查看账户状态（初始 10,000 USDT 模拟资金）
2. 打开行情页 — 查看实时币种价格
3. 点击任意币种进入 K 线图 — 切换周期/交易所/指标
4. 策略管理 — 创建策略 → 填写 JSON 配置 → 保存 → 点击「回测」
5. 回测记录 — 查看收益率/胜率/夏普比率/权益曲线
6. 启动策略 — 策略自动每 30 秒评估并执行交易
7. 订单管理 — 查看自动成交记录
8. 系统设置 — 添加交易所真实 API Key（如需实盘）

## 代理配置说明

中国大陆用户无法直接访问 Binance / OKX API，需通过代理：

| 代理软件 | 典型地址 | config.yaml 配置 |
|---------|---------|-----------------|
| Clash Verge | `http://127.0.0.1:7890` | `url: "http://127.0.0.1:7890"` |
| Clash for Windows | `http://127.0.0.1:7890` | `url: "http://127.0.0.1:7890"` |
| v2rayN | `http://127.0.0.1:10809` | `url: "http://127.0.0.1:10809"` |
| SSR | `socks5://127.0.0.1:1080` | `url: "socks5://127.0.0.1:1080"` |

支持 HTTP 和 SOCKS5 两种代理协议。

## 前端路由

| 路径 | 页面 | 说明 |
|------|------|------|
| `/` | Dashboard | 仪表盘 |
| `/market` | Market | 行情列表 |
| `/market/:symbol` | Chart | K 线图（如 `/market/BTCUSDT`） |
| `/strategies` | StrategyList | 策略列表 |
| `/strategies/:id` | StrategyEdit | 策略编辑（`:id` 为 `new` 时创建） |
| `/backtests` | BacktestList | 回测记录 |
| `/backtests/:id` | BacktestDetail | 回测详情 |
| `/orders` | Orders | 订单管理 |
| `/risk` | RiskManager | 风控规则 |
| `/settings` | Settings | 系统设置 |

## 架构数据流

```
浏览器 ──HTTP──▶ Vite Dev Server (5173) ──proxy──▶ Go Gin Server (8080)
   │                                                    │
   │                                                    ├── SQLite (quant.db)
   │                                                    ├── Binance API (via proxy)
   │                                                    ├── OKX API (via proxy)
   │                                                    └── cron 策略执行器
   │
   └── WebSocket ◀── /ws ──▶ WebSocket Hub ◀── Binance/OKX WS Stream
       实时 K 线 / Ticker / 盘口
```
