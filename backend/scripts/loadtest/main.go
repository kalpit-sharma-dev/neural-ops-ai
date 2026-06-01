package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

type logEntry struct {
	Timestamp   time.Time         `json:"timestamp"`
	Service     string            `json:"service"`
	Environment string            `json:"environment"`
	Severity    string            `json:"severity"`
	Message     string            `json:"message"`
	TraceID     string            `json:"traceId"`
	TxnID       string            `json:"txnId"`
	TenantID    string            `json:"tenantId"`
	Labels      map[string]string `json:"labels"`
}

func main() {
	url := flag.String("url", envOr("INGESTION_URL", "http://localhost:8081/api/v1/logs"), "Ingestion logs endpoint")
	rate := flag.Int("rate", 100000, "Target logs per second")
	duration := flag.Duration("duration", 60*time.Second, "Test duration")
	batch := flag.Int("batch", 200, "Logs per HTTP request")
	workers := flag.Int("workers", 64, "Concurrent workers")
	flag.Parse()

	services := []string{
		"api-gateway", "auth-service", "upi-service", "payment-api",
		"ledger-service", "notification-service", "cbs-adapter", "fraud-service",
	}

	client := &http.Client{Timeout: 15 * time.Second}
	ctx, cancel := context.WithTimeout(context.Background(), *duration+30*time.Second)
	defer cancel()

	var sent, failed atomic.Int64
	latencyMu := sync.Mutex{}
	latencySamples := make([]time.Duration, 0, 4096)
	stop := time.After(*duration)
	var wg sync.WaitGroup

	interval := time.Second / time.Duration(max(*rate / *workers, 1))
	for w := 0; w < *workers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			ticker := time.NewTicker(interval)
			defer ticker.Stop()
			for {
				select {
				case <-stop:
					return
				case <-ctx.Done():
					return
				case <-ticker.C:
					payload := make([]logEntry, 0, *batch)
					now := time.Now().UTC()
					for i := 0; i < *batch; i++ {
						service := services[(workerID+i)%len(services)]
						payload = append(payload, logEntry{
							Timestamp:   now,
							Service:     service,
							Environment: "production",
							Severity:    "INFO",
							Message:     fmt.Sprintf("loadtest worker=%d seq=%d", workerID, sent.Load()),
							TraceID:     fmt.Sprintf("load-%d-%d", workerID, time.Now().UnixNano()),
							TenantID:    "00000000-0000-0000-0000-000000000002",
							Labels:      map[string]string{"source": "loadtest"},
						})
					}
					start := time.Now()
					if err := postLogs(ctx, client, *url, payload); err != nil {
						failed.Add(int64(len(payload)))
						continue
					}
					elapsedBatch := time.Since(start)
					latencyMu.Lock()
					latencySamples = append(latencySamples, elapsedBatch)
					latencyMu.Unlock()
					sent.Add(int64(len(payload)))
				}
			}
		}(w)
	}

	wg.Wait()

	elapsed := *duration
	actualRate := float64(sent.Load()) / elapsed.Seconds()
	p99 := percentile(latencySamples, 0.99)
	avg := time.Duration(0)
	if len(latencySamples) > 0 {
		var total time.Duration
		for _, sample := range latencySamples {
			total += sample
		}
		avg = total / time.Duration(len(latencySamples))
	}

	fmt.Println("")
	fmt.Println("NeuralOps ingestion load test report")
	fmt.Println("====================================")
	fmt.Printf("Target rate:     %d logs/sec\n", *rate)
	fmt.Printf("Duration:        %s\n", duration.String())
	fmt.Printf("Accepted logs:   %d\n", sent.Load())
	fmt.Printf("Failed logs:     %d\n", failed.Load())
	fmt.Printf("Actual rate:     %.0f logs/sec\n", actualRate)
	if len(latencySamples) > 0 {
		fmt.Printf("Batch latency p99: %s\n", p99)
		fmt.Printf("Batch latency avg: %s\n", avg)
	}
	fmt.Println("")
}

func postLogs(ctx context.Context, client *http.Client, url string, logs []logEntry) error {
	body, err := json.Marshal(logs)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	return nil
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func percentile(values []time.Duration, p float64) time.Duration {
	if len(values) == 0 {
		return 0
	}
	sorted := append([]time.Duration(nil), values...)
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j] < sorted[i] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	idx := int(float64(len(sorted)-1) * p)
	if idx < 0 {
		idx = 0
	}
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}
