package elasticsearch

import (
	"testing"
	"time"

	"github.com/neuralops/platform/internal/search/dto"
)

func TestBuildQueryIncludesFiltersAndAggs(t *testing.T) {
	start := time.Date(2026, 5, 30, 22, 0, 0, 0, time.UTC)
	end := start.Add(time.Hour)

	req := dto.LogSearchRequest{
		Query:    "payment timeout",
		Service:  "upi-service",
		Severity: "ERROR",
		TraceID:  "trace-1",
		StartTime: &start,
		EndTime:   &end,
		Size:     25,
	}

	body := BuildQuery(req, true)
	if body["size"] != 25 {
		t.Fatalf("expected size 25, got %v", body["size"])
	}

	query, ok := body["query"].(map[string]any)
	if !ok {
		t.Fatal("expected query map")
	}
	boolQuery, ok := query["bool"].(map[string]any)
	if !ok {
		t.Fatal("expected bool query")
	}
	if _, ok := boolQuery["must"]; !ok {
		t.Fatal("expected must clause")
	}
	if _, ok := boolQuery["filter"]; !ok {
		t.Fatal("expected filter clause")
	}
	if _, ok := body["aggs"]; !ok {
		t.Fatal("expected aggregations")
	}
}

func TestBuildQueryAllowedServicesScope(t *testing.T) {
	req := dto.LogSearchRequest{
		AllowedServices: []string{"upi-service", "payment-api"},
		Size:            10,
	}
	body := BuildQuery(req, false)
	query := body["query"].(map[string]any)["bool"].(map[string]any)
	filters := query["filter"].([]map[string]any)
	found := false
	for _, filter := range filters {
		if terms, ok := filter["terms"].(map[string]any); ok {
			if services, ok := terms["service"].([]string); ok && len(services) == 2 {
				found = true
			}
		}
	}
	if !found {
		t.Fatal("expected terms filter on allowed services")
	}
}

func TestBuildTraceQuerySortsAscending(t *testing.T) {
	body := BuildTraceQuery("trace-abc", "tenant-demo", 100)
	sort, ok := body["sort"].([]map[string]any)
	if !ok || len(sort) == 0 {
		t.Fatal("expected sort clause")
	}
	order, ok := sort[0]["timestamp"].(map[string]string)
	if !ok || order["order"] != "asc" {
		t.Fatalf("expected ascending timestamp sort, got %#v", sort[0])
	}
}
