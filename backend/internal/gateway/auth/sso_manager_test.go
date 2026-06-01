package auth_test

import (
	"context"
	"testing"

	"github.com/neuralops/platform/internal/gateway/auth"
	"github.com/neuralops/platform/internal/gateway/config"
)

func TestSSOManagerMultiTenant(t *testing.T) {
	m := auth.NewSSOManager(nil, config.OIDCConfig{Enabled: false}, "tenant-a", nil)
	if m.Flow(context.Background(), "tenant-b") != nil {
		t.Fatal("expected nil flow without config")
	}
}
