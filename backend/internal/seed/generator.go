package seed

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

const chunkSize = 2_000

// GenerateAllLogs returns baseline logs plus UPI outage simulation logs.
func GenerateAllLogs(now time.Time, normalCount int) []LogRecord {
	if normalCount <= 0 {
		normalCount = DefaultLogCount
	}
	timeline := NewScenarioTimeline(now)
	start := now.Add(-7 * 24 * time.Hour)
	end := timeline.OutageStart.Add(-time.Minute)

	logs := make([]LogRecord, 0, normalCount+4_000)
	logs = append(logs, generateNormalLogs(normalCount, start, end, timeline.OutageStart)...)
	logs = append(logs, generateUPIOutageLogs(timeline)...)
	return logs
}

func generateNormalLogs(count int, start, end, outageStart time.Time) []LogRecord {
	rng := rand.New(rand.NewSource(42))
	logs := make([]LogRecord, 0, count)
	window := end.Sub(start)
	if window <= 0 {
		window = time.Hour
	}

	messages := []string{
		"Request processed successfully",
		"Health check passed",
		"Cache hit for customer profile",
		"Downstream call completed",
		"Batch job finished",
		"Connection pool stats within limits",
	}

	for i := 0; i < count; i++ {
		service := DemoServices[i%len(DemoServices)]
		at := start.Add(time.Duration(rng.Int63n(int64(window))))
		if at.After(outageStart) {
			at = outageStart.Add(-time.Duration(rng.Int63n(int64(30 * time.Minute))))
		}

		severity := "INFO"
		message := messages[rng.Intn(len(messages))]
		classification := ""

		switch roll := rng.Float64(); {
		case roll < 0.005:
			severity = "ERROR"
			message = "Transient downstream timeout recovered"
			classification = "TIMEOUT"
		case roll < 0.02:
			severity = "WARN"
			message = "Elevated latency observed on dependency call"
		}

		logs = append(logs, LogRecord{
			ID:             uuid.NewString(),
			Timestamp:      at.UTC(),
			TenantID:       DemoTenantID,
			Service:        service,
			Environment:    "production",
			Severity:       severity,
			Message:        fmt.Sprintf("%s [%d]", message, i),
			TraceID:        fmt.Sprintf("trace-normal-%d", i),
			TxnID:          fmt.Sprintf("TXN-NORM-%08d", i),
			Host:           fmt.Sprintf("%s-%02d", service, i%3),
			Pod:            fmt.Sprintf("%s-pod-%d", service, i%5),
			Classification: classification,
		})
	}
	return logs
}

func generateUPIOutageLogs(t ScenarioTimeline) []LogRecord {
	rng := rand.New(rand.NewSource(99))
	logs := make([]LogRecord, 0, 4_000)

	appendWindow := func(start, end time.Time, service, severity, message string, count int) {
		if !end.After(start) {
			return
		}
		step := end.Sub(start) / time.Duration(max(count, 1))
		for i := 0; i < count; i++ {
			at := start.Add(step * time.Duration(i))
			logs = append(logs, LogRecord{
				ID:             uuid.NewString(),
				Timestamp:      at.UTC(),
				TenantID:       DemoTenantID,
				Service:        service,
				Environment:    "production",
				Severity:       severity,
				Message:        message,
				TraceID:        fmt.Sprintf("trace-upi-%d", len(logs)),
				TxnID:          fmt.Sprintf("UPI-FAIL-%06d", len(logs)),
				Host:           service + "-01",
				Pod:            service + "-pod-0",
				Classification: "TIMEOUT",
			})
		}
	}

	appendWindow(t.DeployAt, t.DeployAt.Add(time.Minute), "upi-service", "INFO",
		"Deployment v2.3.1 completed successfully", 20)
	appendWindow(t.DeployAt, t.ErrorsStart, "upi-service", "INFO",
		"UPI routing request completed", 400)
	appendWindow(t.ErrorsStart, t.ThreadPoolAt, "upi-service", "ERROR",
		"java.net.SocketTimeoutException: Read timed out", 900)
	appendWindow(t.ThreadPoolAt, t.CascadeAt, "upi-service", "ERROR",
		"java.util.concurrent.RejectedExecutionException: Thread pool exhausted", 700)
	appendWindow(t.CascadeAt, t.ResolvedAt, "payment-api", "ERROR",
		"Cascade failure: upstream upi-service unavailable", 600)
	appendWindow(t.CascadeAt, t.ResolvedAt, "notification-service", "ERROR",
		"Failed to publish payment notification", 300)
	appendWindow(t.RollbackAt, t.RollbackAt.Add(2*time.Minute), "upi-service", "INFO",
		"Rollback to v2.3.0 completed", 15)

	// sprinkle random ERROR on ledger during cascade
	for i := 0; i < 200; i++ {
		at := t.CascadeAt.Add(time.Duration(rng.Int63n(int64(t.ResolvedAt.Sub(t.CascadeAt)))))
		logs = append(logs, LogRecord{
			ID: uuid.NewString(), Timestamp: at.UTC(), TenantID: DemoTenantID,
			Service: "ledger-service", Environment: "production", Severity: "ERROR",
			Message: "Ledger write delayed due to upstream payment-api failures",
			TraceID: fmt.Sprintf("trace-ledger-%d", i), TxnID: fmt.Sprintf("UPI-FAIL-L-%06d", i),
			Host: "ledger-service-01", Pod: "ledger-service-pod-1", Classification: "DEPENDENCY_FAILURE",
		})
	}

	return logs
}

// GenerateTransactions creates demo banking transactions including UPI outage failures.
func GenerateTransactions(now time.Time) []TransactionRecord {
	timeline := NewScenarioTimeline(now)
	rng := rand.New(rand.NewSource(77))
	txns := make([]TransactionRecord, 0, DefaultTransactionCount+UPIFailedTxnCount)

	for i := 0; i < DefaultTransactionCount; i++ {
		status := "SUCCESS"
		failedAt := ""
		if i >= 400 {
			status = "FAILED"
			failedAt = DemoServices[(i-400)%len(DemoServices)]
		}
		started := now.Add(-time.Duration(rng.Int63n(int64(7*24*time.Hour))))
		completed := started.Add(time.Duration(100+rng.Intn(400)) * time.Millisecond)
		txns = append(txns, buildTransaction(
			fmt.Sprintf("TXN-DEMO-%06d", i),
			TxnTypes[i%len(TxnTypes)],
			status,
			failedAt,
			int64(100+rng.Intn(900)),
			started,
			&completed,
		))
	}

	outageWindow := timeline.ResolvedAt.Sub(timeline.ErrorsStart)
	for i := 0; i < UPIFailedTxnCount; i++ {
		started := timeline.ErrorsStart.Add(time.Duration(rng.Int63n(int64(outageWindow))))
		failedAt := "upi-service"
		if i > 800 {
			failedAt = "payment-api"
		}
		txns = append(txns, buildTransaction(
			fmt.Sprintf("UPI-FAIL-%06d", i),
			"UPI",
			"FAILED",
			failedAt,
			int64(500+rng.Intn(1500)),
			started,
			nil,
		))
	}

	return txns
}

func buildTransaction(id, txnType, status, failedAt string, latency int64, started time.Time, completed *time.Time) TransactionRecord {
	hops := []map[string]any{
		{"serviceName": "api-gateway", "latencyMs": 12, "status": "SUCCESS"},
		{"serviceName": "auth-service", "latencyMs": 8, "status": "SUCCESS"},
		{"serviceName": "upi-service", "latencyMs": 145, "status": status},
		{"serviceName": "payment-api", "latencyMs": 234, "status": status},
		{"serviceName": "ledger-service", "latencyMs": 180, "status": status},
	}
	if status == "FAILED" && failedAt != "" {
		for i := range hops {
			if hops[i]["serviceName"] == failedAt {
				hops[i]["status"] = "FAILED"
				hops[i]["errorMessage"] = "Processing failed"
			}
		}
	}
	raw, _ := json.Marshal(hops)
	return TransactionRecord{
		TxnID:          id,
		TxnType:        txnType,
		Status:         status,
		FailedAt:       failedAt,
		TotalLatencyMs: latency,
		RetryCount:     0,
		HopsJSON:       string(raw),
		StartedAt:      started.UTC(),
		CompletedAt:    completed,
	}
}

// GenerateDeployments creates 100 deployment records per spec distribution.
func GenerateDeployments(now time.Time) []DeploymentRecord {
	rng := rand.New(rand.NewSource(55))
	deployments := make([]DeploymentRecord, 0, DefaultDeploymentCount)

	for i := 0; i < DefaultDeploymentCount; i++ {
		category := "clean"
		switch {
		case i >= 90:
			category = "caused_incident"
		case i >= 60:
			category = "minor_issue"
		}
		service := DemoServices[i%len(DemoServices)]
		version := fmt.Sprintf("v%d.%d.%d", 1+(i/20), i%10, i%5)
		at := now.Add(-time.Duration(rng.Int63n(int64(30*24*time.Hour))))

		id := uuid.NewString()
		if i == 90 {
			id = UPIOutageDeploymentID
			service = "upi-service"
			version = "v2.3.1"
			at = NewScenarioTimeline(now).DeployAt
			category = "caused_incident"
		}

		deployments = append(deployments, DeploymentRecord{
			ID:          id,
			Service:     service,
			Version:     version,
			DeployedAt:  at.UTC(),
			DeployedBy:  "ci-bot",
			ChangeType:  "deploy",
			Environment: "production",
			Category:    category,
		})
	}
	return deployments
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
