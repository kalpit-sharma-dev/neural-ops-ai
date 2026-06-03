package connectors_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/neuralops/platform/internal/finops/connectors"
)

func TestBillingSourcesStatusLiveAWS(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cur.ndjson")
	if err := os.WriteFile(path, []byte(`{"provider":"aws","amortizedCost":1}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("FINOPS_BILLING_MODE", "live")
	t.Setenv("FINOPS_AWS_CUR_FILE", path)
	t.Setenv("FINOPS_AWS_INVOICE_USD", "100")

	r := connectors.BillingSourcesStatus()
	if !r.Ready {
		t.Fatalf("expected ready live billing: %+v", r)
	}
	found := false
	for _, s := range r.Sources {
		if s.Provider == "aws" && s.FileExists {
			found = true
		}
	}
	if !found {
		t.Fatal("expected aws file exists")
	}
}

func TestIngestorMultiProviderLiveFiles(t *testing.T) {
	aws := filepath.Join("testdata", "aws_cur_sample.ndjson")
	gcp := filepath.Join("testdata", "gcp_billing_sample.ndjson")
	azure := filepath.Join("testdata", "azure_billing_sample.ndjson")
	t.Setenv("FINOPS_BILLING_MODE", "live")
	t.Setenv("FINOPS_AWS_CUR_FILE", aws)
	t.Setenv("FINOPS_GCP_BILLING_FILE", gcp)
	t.Setenv("FINOPS_AZURE_BILLING_FILE", azure)
	t.Setenv("FINOPS_AWS_INVOICE_USD", "730.75")

	ing := connectors.NewIngestor(nil)
	snaps, items, err := ing.IngestAll(t.Context(), "default")
	if err != nil {
		t.Fatal(err)
	}
	if len(snaps) < 3 {
		t.Fatalf("expected 3 provider snapshots, got %d", len(snaps))
	}
	if len(items) < 4 {
		t.Fatalf("expected line items from all providers, got %d", len(items))
	}
}
