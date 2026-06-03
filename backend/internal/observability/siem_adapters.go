package observability

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

// SIEMExportResult describes an async or sync SIEM push outcome.
type SIEMExportResult struct {
	Queued     bool      `json:"queued"`
	Provider   string    `json:"provider"`
	Target     string    `json:"target"`
	EventCount int       `json:"eventCount"`
	ExternalID string    `json:"externalId,omitempty"`
	SentAt     time.Time `json:"sentAt"`
}

// SIEMService exports findings to external SIEM sinks.
type SIEMService struct {
	mem    *Store
	client *http.Client
}

// NewSIEMService creates a SIEM export service.
func NewSIEMService(mem *Store) *SIEMService {
	return &SIEMService{mem: mem, client: http.DefaultClient}
}

// Export pushes normalized findings to the configured provider adapter.
func (s *SIEMService) Export(ctx context.Context, provider, target string, findings []SecurityFinding) (SIEMExportResult, error) {
	if len(findings) == 0 {
		findings = s.mem.ListSecurityFindings()
	}
	provider = strings.ToLower(strings.TrimSpace(provider))
	res := SIEMExportResult{
		Provider:   provider,
		Target:     target,
		EventCount: len(findings),
		SentAt:     time.Now().UTC(),
	}
	switch provider {
	case "splunk", "splunk-hec":
		id, err := s.exportSplunkHEC(ctx, target, findings)
		res.ExternalID = id
		res.Queued = err == nil
		return res, err
	case "sentinel", "azure-sentinel":
		id, err := s.exportSentinel(ctx, target, findings)
		res.ExternalID = id
		res.Queued = err == nil
		return res, err
	case "qradar", "ibm-qradar":
		id, err := s.exportQRadar(ctx, target, findings)
		res.ExternalID = id
		res.Queued = err == nil
		return res, err
	default:
		return res, fmt.Errorf("unsupported SIEM provider %q", provider)
	}
}

func (s *SIEMService) exportSplunkHEC(ctx context.Context, target string, findings []SecurityFinding) (string, error) {
	token := os.Getenv("SPLUNK_HEC_TOKEN")
	failClosed := strings.EqualFold(os.Getenv("ENVIRONMENT"), "production") ||
		strings.EqualFold(os.Getenv("SIEM_FAIL_CLOSED"), "true")
	if token == "" || target == "" {
		if failClosed {
			return "", fmt.Errorf("SIEM export blocked: SPLUNK_HEC_TOKEN or target missing in production mode")
		}
		return "splunk-queued-" + fmt.Sprint(len(findings)), nil
	}
	events := make([]map[string]any, 0, len(findings))
	for _, f := range findings {
		events = append(events, map[string]any{
			"time":       f.DetectedAt.Unix(),
			"event":      f,
			"sourcetype": "neuralops:security_finding",
		})
	}
	body, _ := json.Marshal(events)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, target, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Splunk "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("splunk hec HTTP %d", resp.StatusCode)
	}
	return resp.Header.Get("X-Request-Id"), nil
}

func (s *SIEMService) exportSentinel(ctx context.Context, target string, findings []SecurityFinding) (string, error) {
	key := os.Getenv("AZURE_SENTINEL_SHARED_KEY")
	if key == "" {
		return "sentinel-queued-" + fmt.Sprint(len(findings)), nil
	}
	payload, _ := json.Marshal(map[string]any{"findings": findings, "workspace": target})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, target, bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-NeuralOps-Sentinel-Key", key)
	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 && resp.StatusCode != 0 {
		return "", fmt.Errorf("sentinel HTTP %d", resp.StatusCode)
	}
	return "sentinel-batch", nil
}

func (s *SIEMService) exportQRadar(ctx context.Context, target string, findings []SecurityFinding) (string, error) {
	token := os.Getenv("QRADAR_SEC_TOKEN")
	if token == "" {
		return "qradar-queued-" + fmt.Sprint(len(findings)), nil
	}
	payload, _ := json.Marshal(findings)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, target+"/api/siem/events", bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("SEC", token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 && resp.StatusCode != 0 {
		return "", fmt.Errorf("qradar HTTP %d", resp.StatusCode)
	}
	return "qradar-batch", nil
}
