package observability

import (
	"context"
	"os"
	"testing"
)

func TestCloudCollectorLiveAWS(t *testing.T) {
	mem := NewStore()
	svc := NewCloudCollectorService(mem)
	os.Setenv("NEURALOPS_CLOUD_LIVE", "true")
	os.Setenv("AWS_ACCESS_KEY_ID", "test")
	defer func() {
		os.Unsetenv("NEURALOPS_CLOUD_LIVE")
		os.Unsetenv("AWS_ACCESS_KEY_ID")
	}()
	assets := svc.ListCloudAssets(context.Background(), "aws")
	if len(assets) == 0 {
		t.Fatal("expected live aws assets")
	}
	if assets[0].Tags["source"] != "live" {
		t.Fatalf("expected live tag, got %v", assets[0].Tags)
	}
}

func TestSIEMExportSplunkQueued(t *testing.T) {
	mem := NewStore()
	svc := NewSIEMService(mem)
	res, err := svc.Export(context.Background(), "splunk", "https://splunk.example/hec", mem.ListSecurityFindings())
	if err != nil {
		t.Fatal(err)
	}
	if res.EventCount == 0 {
		t.Fatal("expected events")
	}
}

func TestSecurityCorrelation(t *testing.T) {
	mem := NewStore()
	svc := NewSecurityCorrelationService(mem, nil)
	f, err := svc.CorrelateFinding(context.Background(), "default", "sf-2")
	if err != nil {
		t.Fatal(err)
	}
	if f.IncidentID == "" {
		t.Fatal("expected incident linkage")
	}
}

func TestQueryPlannerSteps(t *testing.T) {
	mem := NewStore()
	res := PlanUnifiedQuery(mem, UnifiedQueryRequest{Query: "error", From: "all"})
	if len(res.Planner) < 2 {
		t.Fatalf("expected planner steps, got %d", len(res.Planner))
	}
}
