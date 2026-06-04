package observability

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// LogPatternCluster groups similar log lines (REQ-LOG-008).
type LogPatternCluster struct {
	ID            string    `json:"id"`
	Template      string    `json:"template"`
	Count         int64     `json:"count"`
	Service       string    `json:"service,omitempty"`
	Severity      string    `json:"severity,omitempty"`
	SampleMessage string    `json:"sampleMessage,omitempty"`
	FirstSeen     time.Time `json:"firstSeen"`
	LastSeen      time.Time `json:"lastSeen"`
}

func (h *Handler) GetLogPatterns(c *gin.Context) {
	limit := 50
	if raw := c.Query("limit"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 && n <= 200 {
			limit = n
		}
	}
	service := c.Query("service")
	writeSuccess(c, h.deps.Mem.LogPatternClusters(tenantID(c), service, limit))
}

// LogPatternClusters returns seeded pattern groups for demo / in-memory mode.
func (s *Store) LogPatternClusters(tenantID, service string, limit int) []LogPatternCluster {
	_ = tenantID
	now := time.Now().UTC()
	clusters := []LogPatternCluster{
		{
			ID: "pat-5xx", Template: "POST /payments failed status=<code> traceId=<traceId>",
			Count: 1842, Service: "payment-service", Severity: "ERROR",
			SampleMessage: "POST /payments failed status=500 traceId=trace-demo-04",
			FirstSeen:     now.Add(-72 * time.Hour), LastSeen: now.Add(-8 * time.Minute),
		},
		{
			ID: "pat-timeout", Template: "downstream <service> timeout after <ms>ms",
			Count: 412, Service: "ledger-service", Severity: "WARN",
			SampleMessage: "downstream ledger-service timeout after 3000ms",
			FirstSeen:     now.Add(-48 * time.Hour), LastSeen: now.Add(-22 * time.Minute),
		},
		{
			ID: "pat-auth", Template: "authentication failed for user=<user> reason=<reason>",
			Count: 96, Service: "auth-service", Severity: "INFO",
			SampleMessage: "authentication failed for user=demo reason=invalid_token",
			FirstSeen:     now.Add(-24 * time.Hour), LastSeen: now.Add(-3 * time.Hour),
		},
		{
			ID: "pat-ingest", Template: "ingest lag elevated partition=<partition> lagMs=<lag>",
			Count: 58, Service: "api-gateway", Severity: "WARN",
			SampleMessage: "ingest lag elevated partition=logs-3 lagMs=4200",
			FirstSeen:     now.Add(-12 * time.Hour), LastSeen: now.Add(-45 * time.Minute),
		},
	}
	out := make([]LogPatternCluster, 0, len(clusters))
	for _, c := range clusters {
		if service != "" && c.Service != service {
			continue
		}
		out = append(out, c)
		if len(out) >= limit {
			break
		}
	}
	return out
}
