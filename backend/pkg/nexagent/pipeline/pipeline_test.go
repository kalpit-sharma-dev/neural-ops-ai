package pipeline

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/neuralops/platform/pkg/nexagent/buffer"
	"github.com/neuralops/platform/pkg/nexagent/collector"
)

// stubCollector emits a fixed sample so the pipeline test is deterministic.
type stubCollector struct{ n int }

func (s *stubCollector) Name() string { return "stub" }
func (s *stubCollector) Collect(_ context.Context) ([]collector.Sample, error) {
	out := make([]collector.Sample, s.n)
	for i := range out {
		out[i] = collector.Sample{Name: "test.metric", Value: float64(i), Kind: "gauge", Timestamp: time.Now().UTC()}
	}
	return out, nil
}

func TestPipelineScrapeEnqueuesBatch(t *testing.T) {
	dir := t.TempDir()
	spool := buffer.NewDiskSpool(dir)
	p := New(Config{Hostname: "test-host", Gateway: "http://localhost:0", AgentVersion: "test"}, spool, nil, &stubCollector{n: 3})

	p.scrapeOnce(context.Background())

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 spooled batch, got %d", len(entries))
	}

	raw, err := os.ReadFile(filepath.Join(dir, entries[0].Name()))
	if err != nil {
		t.Fatal(err)
	}
	var payload struct {
		Source      string             `json:"source"`
		Host        string             `json:"host"`
		SampleCount int                `json:"sampleCount"`
		Samples     []collector.Sample `json:"samples"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Source != "nexagent" || payload.Host != "test-host" {
		t.Fatalf("unexpected envelope: %+v", payload)
	}
	if payload.SampleCount != 3 || len(payload.Samples) != 3 {
		t.Fatalf("expected 3 samples, got count=%d len=%d", payload.SampleCount, len(payload.Samples))
	}
}

func TestPipelineEmptyScrapeSkipsSpool(t *testing.T) {
	dir := t.TempDir()
	spool := buffer.NewDiskSpool(dir)
	p := New(Config{Hostname: "h"}, spool, nil, &stubCollector{n: 0})

	p.scrapeOnce(context.Background())

	entries, _ := os.ReadDir(dir)
	if len(entries) != 0 {
		t.Fatalf("expected no spooled batch for empty scrape, got %d", len(entries))
	}
}

func TestHostCollectorProducesSamples(t *testing.T) {
	c := collector.NewHostCollector("h")
	samples, err := c.Collect(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(samples) == 0 {
		t.Fatal("expected at least one host sample (cpu/mem)")
	}
	for _, s := range samples {
		if s.Name == "" || s.Timestamp.IsZero() {
			t.Fatalf("invalid sample: %+v", s)
		}
	}
}
