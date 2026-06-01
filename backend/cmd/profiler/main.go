package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/grafana/pyroscope-go"
	"github.com/neuralops/platform/pkg/config"
	"go.uber.org/zap"
)

type profileHotspot struct {
	FunctionName string  `json:"functionName"`
	FilePath     string  `json:"filePath,omitempty"`
	LineNo       int     `json:"lineNo"`
	SelfTimeMs   float64 `json:"selfTimeMs"`
	SampleCount  int64   `json:"sampleCount"`
	Service      string  `json:"service"`
}

func main() {
	log, _ := zap.NewProduction()
	defer func() { _ = log.Sync() }()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	apiBase := config.Getenv("API_BASE", "http://localhost:8080/api/v1")
	service := config.Getenv("SERVICE_NAME", "payment-api")
	tenantID := config.Getenv("TENANT_ID", "default")
	interval, _ := time.ParseDuration(config.Getenv("PROFILER_INTERVAL", "30s"))
	pyroscopeURL := config.Getenv("PYROSCOPE_SERVER", "")

	if pyroscopeURL != "" {
		_, err := pyroscope.Start(pyroscope.Config{
			ApplicationName: service,
			ServerAddress:   pyroscopeURL,
			Logger:          pyroscope.StandardLogger,
			ProfileTypes: []pyroscope.ProfileType{
				pyroscope.ProfileCPU,
				pyroscope.ProfileAllocObjects,
				pyroscope.ProfileInuseObjects,
			},
		})
		if err != nil {
			log.Warn("pyroscope start failed", zap.Error(err))
		} else {
			log.Info("pyroscope continuous profiling enabled", zap.String("server", pyroscopeURL))
		}
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	sample := func() { sampleStack(ctx, apiBase, service, tenantID, interval) }
	sample()
	log.Info("profiler agent started", zap.String("service", service), zap.Duration("interval", interval))
	for {
		select {
		case <-ctx.Done():
			log.Info("profiler agent stopped")
			return
		case <-ticker.C:
			sample()
		}
	}
}

func sampleStack(ctx context.Context, apiBase, service, tenantID string, interval time.Duration) {
	pc := make([]uintptr, 128)
	n := runtime.Callers(3, pc)
	if n == 0 {
		return
	}
	frames := runtime.CallersFrames(pc[:n])
	count := 0
	for {
		frame, more := frames.Next()
		if frame.Function == "" {
			if !more {
				break
			}
			continue
		}
		body := profileHotspot{
			FunctionName: frame.Function,
			FilePath:     frame.File,
			LineNo:       frame.Line,
			SelfTimeMs:   float64(interval.Milliseconds()) / float64(n),
			SampleCount:  1,
			Service:      service,
		}
		raw, _ := json.Marshal(body)
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiBase+"/apm/profiles", bytes.NewReader(raw))
		if err != nil {
			return
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Tenant-ID", tenantID)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return
		}
		resp.Body.Close()
		count++
		if !more || count > 8 {
			break
		}
	}
}
