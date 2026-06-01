package classifier

import (
	"testing"

	"github.com/neuralops/platform/internal/domain"
)

func TestClassifyRuleBasedTimeout(t *testing.T) {
	result := ClassifyRuleBased("java.net.SocketTimeoutException: Read timed out")
	if result == nil {
		t.Fatal("expected classification")
	}
	if result.Category != domain.ErrorCategoryTimeout {
		t.Fatalf("expected TIMEOUT, got %s", result.Category)
	}
	if result.Confidence < 0.8 {
		t.Fatalf("expected high confidence, got %f", result.Confidence)
	}
}

func TestClassifyRuleBasedDeadlock(t *testing.T) {
	result := ClassifyRuleBased("ERROR: deadlock detected while waiting for resource")
	if result == nil || result.Category != domain.ErrorCategoryDBDeadlock {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestClassifyRuleBasedMemoryLeak(t *testing.T) {
	result := ClassifyRuleBased("Container OOMKilled due to memory limit")
	if result == nil || result.Category != domain.ErrorCategoryMemoryLeak {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestClassifyRuleBasedConnectionExhaustion(t *testing.T) {
	result := ClassifyRuleBased("dial tcp 10.0.0.5:5432: connection refused")
	if result == nil || result.Category != domain.ErrorCategoryConnectionExhaustion {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestClassifyRuleBasedUnknown(t *testing.T) {
	result := ClassifyRuleBased("service started successfully")
	if result != nil {
		t.Fatalf("expected nil for non-error message, got %+v", result)
	}
}
