package connector

import (
	"bufio"
	"context"
	"encoding/json"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/neuralops/platform/internal/ingestion/config"
	"github.com/neuralops/platform/internal/ingestion/dto"
	"github.com/neuralops/platform/internal/ingestion/service"
	"go.uber.org/zap"
)

// Manager runs optional ingestion connectors.
type Manager struct {
	cfg     config.ConnectorsConfig
	log     *zap.Logger
	service *service.Service
}

// NewManager creates a connector manager.
func NewManager(cfg config.ConnectorsConfig, log *zap.Logger, svc *service.Service) *Manager {
	return &Manager{cfg: cfg, log: log, service: svc}
}

// Start launches enabled connectors until context cancellation.
func (m *Manager) Start(ctx context.Context) {
	if m.cfg.FileTail.Enabled {
		go m.runFileTail(ctx)
	}
	if m.cfg.Syslog.Enabled {
		go m.runSyslogUDP(ctx)
		go m.runSyslogTCP(ctx)
	}
	if m.cfg.Fluent.Enabled {
		go m.runFluentForward(ctx)
	}
}

func (m *Manager) runFileTail(ctx context.Context) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		m.log.Error("file tail watcher init failed", zap.Error(err))
		return
	}
	defer watcher.Close()

	for _, path := range m.cfg.FileTail.Paths {
		if err := watcher.Add(path); err != nil {
			m.log.Warn("unable to watch file", zap.String("path", path), zap.Error(err))
		}
	}

	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				m.tailFile(ctx, event.Name)
			}
		}
	}
}

func (m *Manager) tailFile(ctx context.Context, path string) {
	file, err := os.Open(path)
	if err != nil {
		m.log.Warn("open tailed file failed", zap.String("path", path), zap.Error(err))
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}
		req := dto.LogIngestRequest{
			Timestamp: time.Now().UTC(),
			Service:   inferServiceFromPath(path),
			Message:   line,
			Environment: "dev",
		}
		if _, _, err := m.service.IngestLogs(ctx, []dto.LogIngestRequest{req}, "file-tail"); err != nil {
			m.log.Warn("file tail ingest failed", zap.Error(err))
		}
	}
}

func (m *Manager) runSyslogUDP(ctx context.Context) {
	addr, err := net.ResolveUDPAddr("udp", net.JoinHostPort("0.0.0.0", itoa(m.cfg.Syslog.UDPPort)))
	if err != nil {
		m.log.Error("syslog udp resolve failed", zap.Error(err))
		return
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		m.log.Error("syslog udp listen failed", zap.Error(err))
		return
	}
	defer conn.Close()

	buffer := make([]byte, 65536)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			_ = conn.SetReadDeadline(time.Now().Add(time.Second))
			n, _, readErr := conn.ReadFromUDP(buffer)
			if readErr != nil {
				if netErr, ok := readErr.(net.Error); ok && netErr.Timeout() {
					continue
				}
				m.log.Warn("syslog udp read failed", zap.Error(readErr))
				continue
			}
			m.ingestSyslogLine(ctx, string(buffer[:n]))
		}
	}
}

func (m *Manager) runSyslogTCP(ctx context.Context) {
	listener, err := net.Listen("tcp", net.JoinHostPort("0.0.0.0", itoa(m.cfg.Syslog.TCPPort)))
	if err != nil {
		m.log.Error("syslog tcp listen failed", zap.Error(err))
		return
	}
	defer listener.Close()

	go func() {
		<-ctx.Done()
		_ = listener.Close()
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return
			default:
				m.log.Warn("syslog tcp accept failed", zap.Error(err))
				continue
			}
		}
		go m.handleSyslogConn(ctx, conn)
	}
}

func (m *Manager) handleSyslogConn(ctx context.Context, conn net.Conn) {
	defer conn.Close()
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		m.ingestSyslogLine(ctx, scanner.Text())
	}
}

func (m *Manager) ingestSyslogLine(ctx context.Context, line string) {
	req := dto.LogIngestRequest{
		Timestamp:   time.Now().UTC(),
		Service:     "syslog",
		Message:     line,
		Environment: "dev",
	}
	if _, _, err := m.service.IngestLogs(ctx, []dto.LogIngestRequest{req}, "syslog"); err != nil {
		m.log.Warn("syslog ingest failed", zap.Error(err))
	}
}

func (m *Manager) runFluentForward(ctx context.Context) {
	listener, err := net.Listen("tcp", net.JoinHostPort("0.0.0.0", itoa(m.cfg.Fluent.Port)))
	if err != nil {
		m.log.Error("fluent forward listen failed", zap.Error(err))
		return
	}
	defer listener.Close()

	go func() {
		<-ctx.Done()
		_ = listener.Close()
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return
			default:
				m.log.Warn("fluent forward accept failed", zap.Error(err))
				continue
			}
		}
		go m.handleFluentConn(ctx, conn)
	}
}

func (m *Manager) handleFluentConn(ctx context.Context, conn net.Conn) {
	defer conn.Close()
	decoder := json.NewDecoder(conn)
	for decoder.More() {
		var payload map[string]any
		if err := decoder.Decode(&payload); err != nil {
			m.log.Warn("fluent decode failed", zap.Error(err))
			return
		}
		message, _ := payload["message"].(string)
		if message == "" {
			raw, _ := json.Marshal(payload)
			message = string(raw)
		}
		serviceName, _ := payload["service"].(string)
		if serviceName == "" {
			serviceName = "fluent"
		}
		req := dto.LogIngestRequest{
			Timestamp:   time.Now().UTC(),
			Service:     serviceName,
			Message:     message,
			Environment: "dev",
		}
		if _, _, err := m.service.IngestLogs(ctx, []dto.LogIngestRequest{req}, "fluent-forward"); err != nil {
			m.log.Warn("fluent ingest failed", zap.Error(err))
		}
	}
}

func inferServiceFromPath(path string) string {
	base := path
	if idx := strings.LastIndex(path, string(os.PathSeparator)); idx >= 0 {
		base = path[idx+1:]
	}
	if idx := strings.Index(base, "."); idx >= 0 {
		base = base[:idx]
	}
	if base == "" {
		return "file-tail"
	}
	return base
}

func itoa(v int) string {
	return strconv.Itoa(v)
}
