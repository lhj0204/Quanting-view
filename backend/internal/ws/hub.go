package ws

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Message struct {
	Event   string `json:"event"`
	Channel string `json:"channel"`
	Symbol  string `json:"symbol"`
	Interval string `json:"interval,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type Client struct {
	ID     string
	Conn   *websocket.Conn
	Hub    *Hub
	Send   chan []byte
	Subs   map[string]bool // "kline:BTCUSDT:1h"
}

type Hub struct {
	mu       sync.RWMutex
	Clients  map[*Client]bool
	Register   chan *Client
	Unregister chan *Client
	Broadcast  chan []byte
}

func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[*Client]bool),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Broadcast:  make(chan []byte, 256),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			h.Clients[client] = true
			h.mu.Unlock()

		case client := <-h.Unregister:
			h.mu.Lock()
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				close(client.Send)
			}
			h.mu.Unlock()

		case msg := <-h.Broadcast:
			h.mu.RLock()
			for client := range h.Clients {
				select {
				case client.Send <- msg:
				default:
					close(client.Send)
					delete(h.Clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

func HandleWebSocket(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("ws upgrade error: %v", err)
		return
	}

	client := &Client{
		ID:   r.RemoteAddr,
		Conn: conn,
		Hub:  hub,
		Send: make(chan []byte, 256),
		Subs: make(map[string]bool),
	}
	hub.Register <- client

	go client.writePump()
	go client.readPump()
}

func (c *Client) readPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	for {
		_, msg, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}
		var req Message
		if err := json.Unmarshal(msg, &req); err != nil {
			continue
		}

		switch req.Event {
		case "subscribe":
			key := req.Channel
			if req.Symbol != "" {
				key += ":" + req.Symbol
			}
			if req.Interval != "" {
				key += ":" + req.Interval
			}
			c.Subs[key] = true
			log.Printf("client %s subscribed to %s", c.ID, key)

		case "unsubscribe":
			key := req.Channel
			if req.Symbol != "" {
				key += ":" + req.Symbol
			}
			if req.Interval != "" {
				key += ":" + req.Interval
			}
			delete(c.Subs, key)
		}
	}
}

func (c *Client) writePump() {
	defer c.Conn.Close()
	for msg := range c.Send {
		if err := c.Conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			return
		}
	}
}
