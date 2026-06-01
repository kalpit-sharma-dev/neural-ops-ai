// Package oneagent provides lightweight HTTP auto-instrumentation for Go services.
package oneagent

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"net/http"
	"os"
	"time"
)

// Config controls span emission.
type Config struct {
	ServiceName string
	IngestURL   string
	TenantID    string
	SampleRate  float64
}

// InitFromEnv configures the agent from NEURALOPS_* environment variables.
func InitFromEnv() *Config {
	cfg := &Config{
		ServiceName: envOr("NEURALOPS_SERVICE", "go-service"),
		IngestURL:   envOr("NEURALOPS_INGEST_URL", "http://localhost:8080/api/v1/apm/profiles"),
		TenantID:    envOr("NEURALOPS_TENANT_ID", "default"),
		SampleRate:  1.0,
	}
	Init(cfg)
	return cfg
}

// Init wraps http.DefaultTransport with span emission.
func Init(cfg *Config) {
	if cfg == nil {
		cfg = &Config{ServiceName: "go-service"}
	}
	http.DefaultTransport = &transport{base: http.DefaultTransport, cfg: cfg}
}

type transport struct {
	base http.RoundTripper
	cfg  *Config
}

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()
	resp, err := t.base.RoundTrip(req)
	if shouldSample(t.cfg.SampleRate) {
		status := 0
		if resp != nil {
			status = resp.StatusCode
		}
		emitSpan(t.cfg, req.Method+" "+req.URL.Path, start, status)
	}
	return resp, err
}

func emitSpan(cfg *Config, op string, start time.Time, status int) {
	body, _ := json.Marshal(map[string]any{
		"functionName": op,
		"filePath":     "net/http",
		"lineNo":       status,
		"selfTimeMs":   float64(time.Since(start).Milliseconds()),
		"sampleCount":  1,
		"service":      cfg.ServiceName,
	})
	req, err := http.NewRequest(http.MethodPost, cfg.IngestURL, bytes.NewReader(body))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", cfg.TenantID)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	resp.Body.Close()
}

func shouldSample(rate float64) bool {
	if rate >= 1 {
		return true
	}
	var b [1]byte
	_, _ = rand.Read(b[:])
	return float64(b[0])/255.0 < rate
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
