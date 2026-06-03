package websocket

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	gorilla "github.com/gorilla/websocket"
	"go.uber.org/zap"
)

const (
	defaultLogTailQueueDepth = 256
	maxLogTailClients        = 500
)

// LogTailLine is one log event streamed to clients (LOG-04).
type LogTailLine struct {
	Timestamp time.Time `json:"timestamp"`
	Service   string    `json:"service"`
	Severity  string    `json:"severity"`
	Message   string    `json:"message"`
	TraceID   string    `json:"traceId,omitempty"`
	Host      string    `json:"host,omitempty"`
	Pod       string    `json:"pod,omitempty"`
	TenantID  string    `json:"tenantId,omitempty"`
}

// LogTailSubscription filters tail output.
type LogTailSubscription struct {
	Services   []string `json:"services"`
	Severities []string `json:"severities"`
	Query      string   `json:"query"`
}

type logTailClient struct {
	conn *gorilla.Conn
	hub  *LogTailHub
	sub  LogTailSubscription
	send chan LogTailLine
}

// LogTailHub streams ingested logs to WebSocket subscribers with backpressure.
type LogTailHub struct {
	log      *zap.Logger
	upgrader gorilla.Upgrader
	mu       sync.RWMutex
	clients  map[*logTailClient]struct{}
}

// NewLogTailHub creates a log tail broadcaster.
func NewLogTailHub(log *zap.Logger, allowedOrigins []string) *LogTailHub {
	return &LogTailHub{
		log:     log,
		clients: make(map[*logTailClient]struct{}),
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

// Handle upgrades HTTP to a log tail WebSocket.
func (h *LogTailHub) Handle(w http.ResponseWriter, r *http.Request, pingInterval time.Duration) {
	h.mu.RLock()
	if len(h.clients) >= maxLogTailClients {
		h.mu.RUnlock()
		http.Error(w, "too many concurrent log tail sessions", http.StatusTooManyRequests)
		return
	}
	h.mu.RUnlock()

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.log.Warn("log tail websocket upgrade failed", zap.Error(err))
		return
	}

	client := &logTailClient{
		conn: conn,
		hub:  h,
		send: make(chan LogTailLine, defaultLogTailQueueDepth),
	}
	h.register(client)
	defer h.unregister(client)

	if pingInterval <= 0 {
		pingInterval = 30 * time.Second
	}
	conn.SetReadLimit(8192)
	_ = conn.SetReadDeadline(time.Now().Add(pingInterval * 2))
	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(pingInterval * 2))
	})

	go h.writePump(client, pingInterval)

	for {
		_, payload, err := conn.ReadMessage()
		if err != nil {
			break
		}
		var sub LogTailSubscription
		if err := json.Unmarshal(payload, &sub); err == nil {
			client.sub = sub
		}
	}
}

func (h *LogTailHub) register(c *logTailClient) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[c] = struct{}{}
}

func (h *LogTailHub) unregister(c *logTailClient) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.clients, c)
	close(c.send)
	_ = c.conn.Close()
}

func (h *LogTailHub) writePump(c *logTailClient, pingInterval time.Duration) {
	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()
	for {
		select {
		case line, ok := <-c.send:
			if !ok {
				return
			}
			raw, err := json.Marshal(line)
			if err != nil {
				continue
			}
			if err := c.conn.WriteMessage(gorilla.TextMessage, raw); err != nil {
				return
			}
		case <-ticker.C:
			if err := c.conn.WriteControl(gorilla.PingMessage, []byte("ping"), time.Now().Add(time.Second)); err != nil {
				return
			}
		}
	}
}

// Publish enqueues a log line for matching subscribers (drops oldest on backpressure).
func (h *LogTailHub) Publish(line LogTailLine) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for client := range h.clients {
		if !matchesLogTailSub(client.sub, line) {
			continue
		}
		select {
		case client.send <- line:
		default:
			// Drop oldest to preserve real-time tail semantics under backpressure.
			select {
			case <-client.send:
			default:
			}
			select {
			case client.send <- line:
			default:
			}
		}
	}
}

func matchesLogTailSub(sub LogTailSubscription, line LogTailLine) bool {
	if len(sub.Services) > 0 {
		found := false
		for _, svc := range sub.Services {
			if svc == line.Service {
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
		upper := strings.ToUpper(line.Severity)
		for _, sev := range sub.Severities {
			if strings.ToUpper(sev) == upper {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	if q := strings.TrimSpace(sub.Query); q != "" {
		needle := strings.ToLower(q)
		hay := strings.ToLower(line.Message + " " + line.Service + " " + line.TraceID)
		if !strings.Contains(hay, needle) {
			return false
		}
	}
	return true
}

// StartDemoFeed emits sample payment logs when Kafka is unavailable (local dev).
func (h *LogTailHub) StartDemoFeed(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		interval = 3 * time.Second
	}
	services := []string{"upi-gateway", "payment-service", "ledger-service", "notification-service"}
	severities := []string{"INFO", "INFO", "WARN", "ERROR"}
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		i := 0
		for {
			select {
			case <-ctx.Done():
				return
			case now := <-ticker.C:
				svc := services[i%len(services)]
				sev := severities[i%len(severities)]
				h.Publish(LogTailLine{
					Timestamp: now.UTC(),
					Service:   svc,
					Severity:  sev,
					Message:   "demo tail event txnId=txn-demo-" + now.Format("150405"),
					TraceID:   "trace-demo-" + now.Format("150405"),
					TenantID:  "default",
				})
				i++
			}
		}
	}()
}
