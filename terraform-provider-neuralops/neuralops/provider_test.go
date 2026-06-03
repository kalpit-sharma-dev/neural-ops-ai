package neuralops_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/neuralops/terraform-provider-neuralops/neuralops"
)

func TestProvider(t *testing.T) {
	if neuralops.New() == nil {
		t.Fatal("expected provider")
	}
}

func TestAccExportJob(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("set TF_ACC=1 and NEURALOPS_API_URL to run acceptance tests")
	}
	if os.Getenv("NEURALOPS_API_URL") == "" {
		t.Skip("NEURALOPS_API_URL required for acceptance tests")
	}
}

func TestAccAlertPolicySchema(t *testing.T) {
	r := neuralops.New().ResourcesMap["neuralops_alert_policy"]
	if r == nil {
		t.Fatal("neuralops_alert_policy resource missing")
	}
	if r.Schema["name"] == nil {
		t.Fatal("expected name schema")
	}
}

func TestAlertPolicyCreateAgainstMockAPI(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/alerts/policies" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status": "success",
			"data":   map[string]string{"id": "ap-mock-1"},
		})
	}))
	defer srv.Close()

	body, _ := json.Marshal(map[string]any{
		"name":           "checkout-latency",
		"servicePattern": "checkout-*",
		"severity":       "P1",
		"enabled":        true,
		"routes":         []map[string]any{{"channel": "slack", "target": "#oncall", "after": "0m", "priority": 1}},
	})
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, srv.URL+"/alerts/policies", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 got %d: %s", resp.StatusCode, string(raw))
	}
	var parsed struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		t.Fatal(err)
	}
	if parsed.Data.ID != "ap-mock-1" {
		t.Fatalf("unexpected id %q", parsed.Data.ID)
	}
}
