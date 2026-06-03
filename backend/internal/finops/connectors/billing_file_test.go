package connectors_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/neuralops/platform/internal/finops/connectors"
)

func TestLoadBillingFileNDJSON(t *testing.T) {
	path := filepath.Join("testdata", "aws_cur_sample.ndjson")
	items, err := connectors.LoadBillingFile(path, "t-1", "aws")
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items got %d", len(items))
	}
	if items[0].Team != "payments" {
		t.Fatalf("expected team tag applied")
	}
}

func TestIngestorLiveCURFile(t *testing.T) {
	path, _ := filepath.Abs(filepath.Join("testdata", "aws_cur_sample.ndjson"))
	t.Setenv("FINOPS_BILLING_MODE", "live")
	t.Setenv("FINOPS_AWS_CUR_FILE", path)
	t.Setenv("FINOPS_AWS_INVOICE_USD", "730.75")

	ing := connectors.NewIngestor(nil)
	snaps, items, err := ing.IngestAll(context.Background(), "live-tenant")
	if err != nil {
		t.Fatalf("ingest: %v", err)
	}
	if len(items) < 2 {
		t.Fatalf("expected live items")
	}
	var awsSnap bool
	for _, s := range snaps {
		if s.Provider == "aws" {
			awsSnap = true
			if s.DriftPct > 1.0 {
				t.Fatalf("drift %.2f exceeds 1%% with matched invoice", s.DriftPct)
			}
		}
	}
	if !awsSnap {
		t.Fatal("expected aws snapshot")
	}
}
