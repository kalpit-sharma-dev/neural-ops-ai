// Package buffer provides air-gapped durable spooling for NEXAGENT telemetry batches.
package buffer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// DiskSpool stores JSON batches on disk until gateway is reachable.
type DiskSpool struct {
	dir string
}

// NewDiskSpool creates a spool directory.
func NewDiskSpool(dir string) *DiskSpool {
	_ = os.MkdirAll(dir, 0o750)
	return &DiskSpool{dir: dir}
}

// Enqueue writes a batch file for later replay.
func (s *DiskSpool) Enqueue(batch map[string]any) error {
	name := filepath.Join(s.dir, fmt.Sprintf("batch-%d.json", time.Now().UnixNano()))
	raw, _ := json.Marshal(batch)
	return os.WriteFile(name, raw, 0o640)
}

// Replay sends all pending batches to the gateway ingest endpoint.
func (s *DiskSpool) Replay(ctx context.Context, gateway string) (int, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return 0, err
	}
	sent := 0
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		path := filepath.Join(s.dir, e.Name())
		raw, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, gateway+"/api/v1/events", bytes.NewReader(raw))
		if err != nil {
			continue
		}
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			continue
		}
		resp.Body.Close()
		if resp.StatusCode < 300 {
			_ = os.Remove(path)
			sent++
		}
	}
	return sent, nil
}

// FlushPending replays pending batches when the gateway is reachable.
func (s *DiskSpool) FlushPending(ctx context.Context, gateway string) error {
	_, err := s.Replay(ctx, gateway)
	return err
}
