package websocket

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	gorilla "github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// Event is a realtime dashboard event.
type Event struct {
	Type      string    `json:"type"`
	Service   string    `json:"service,omitempty"`
	Severity  string    `json:"severity,omitempty"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

// Subscription filters realtime events.
type Subscription struct {
	Services   []string `json:"services"`
	Severities []string `json:"severities"`
}

// Hub manages websocket clients and broadcasts.
type Hub struct {
	log         *zap.Logger
	upgrader    gorilla.Upgrader
	clients     map[*Client]struct{}
	subscribers map[*Client]Subscription
	mu          sync.RWMutex
}

// Client represents a websocket connection.
type Client struct {
	conn *gorilla.Conn
	hub  *Hub
}

// NewHub creates a websocket hub.
func NewHub(log *zap.Logger, allowedOrigins []string) *Hub {
	return &Hub{
		log:         log,
		clients:     make(map[*Client]struct{}),
		subscribers: make(map[*Client]Subscription),
		upgrader: gorilla.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				origin := r.Header.Get("Origin")
				if origin == "" {
					return true
				}
				for _, allowed := range allowedOrigins {
					if allowed == "*" || allowed == origin {
						return true
					}
				}
				return false
			},
		},
	}
}

// Handle upgrades HTTP to websocket and manages the connection lifecycle.
func (h *Hub) Handle(w http.ResponseWriter, r *http.Request, pingInterval time.Duration) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.log.Warn("websocket upgrade failed", zap.Error(err))
		return
	}

	client := &Client{conn: conn, hub: h}
	h.register(client)
	defer h.unregister(client)

	if pingInterval <= 0 {
		pingInterval = 30 * time.Second
	}
	conn.SetReadLimit(4096)
	_ = conn.SetReadDeadline(time.Now().Add(pingInterval * 2))
	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(pingInterval * 2))
	})

	go h.pingLoop(client, pingInterval)

	for {
		_, payload, err := conn.ReadMessage()
		if err != nil {
			break
		}
		var sub Subscription
		if err := json.Unmarshal(payload, &sub); err == nil {
			h.mu.Lock()
			h.subscribers[client] = sub
			h.mu.Unlock()
		}
	}
}

func (h *Hub) register(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[client] = struct{}{}
	h.subscribers[client] = Subscription{}
}

func (h *Hub) unregister(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.clients, client)
	delete(h.subscribers, client)
	_ = client.conn.Close()
}

func (h *Hub) pingLoop(client *Client, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range ticker.C {
		if err := client.conn.WriteControl(gorilla.PingMessage, []byte("ping"), time.Now().Add(time.Second)); err != nil {
			return
		}
	}
}

// Publish broadcasts an event to subscribed clients.
func (h *Hub) Publish(event Event) {
	payload, err := json.Marshal(event)
	if err != nil {
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()
	for client, sub := range h.subscribers {
		if !matchesSubscription(sub, event) {
			continue
		}
		if err := client.conn.WriteMessage(gorilla.TextMessage, payload); err != nil {
			h.log.Debug("websocket write failed", zap.Error(err))
		}
	}
}

func matchesSubscription(sub Subscription, event Event) bool {
	if len(sub.Services) > 0 {
		found := false
		for _, service := range sub.Services {
			if service == event.Service {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	if len(sub.Severities) > 0 {
		found := false
		for _, severity := range sub.Severities {
			if severity == event.Severity {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// StartHeartbeat emits demo heartbeat events for local development.
func (h *Hub) StartHeartbeat(interval time.Duration) {
	if interval <= 0 {
		interval = 60 * time.Second
	}
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for now := range ticker.C {
			h.Publish(Event{
				Type:      "heartbeat",
				Message:   "gateway realtime channel connected",
				Timestamp: now.UTC(),
			})
		}
	}()
}
