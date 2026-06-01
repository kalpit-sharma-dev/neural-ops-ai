package security

import "testing"

func TestHashAPIKeyDeterministic(t *testing.T) {
	first := HashAPIKey("demo-api-key")
	second := HashAPIKey("demo-api-key")
	if first != second {
		t.Fatalf("expected deterministic hash, got %q and %q", first, second)
	}
	if len(first) != 64 {
		t.Fatalf("expected sha256 hex length 64, got %d", len(first))
	}
}
