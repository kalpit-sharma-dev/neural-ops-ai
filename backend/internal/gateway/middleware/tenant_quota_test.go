package middleware

import (
	"testing"

	"github.com/neuralops/platform/internal/gateway/config"
)

func TestEvaluateQuota(t *testing.T) {
	cases := []struct {
		name      string
		reqCount  int64
		byteCount int64
		cfg       config.TenantQuotaConfig
		blocked   bool
		code      string
	}{
		{
			name:     "under request quota",
			reqCount: 5, cfg: config.TenantQuotaConfig{DailyRequestQuota: 10},
			blocked: false,
		},
		{
			name:     "at request quota boundary allowed",
			reqCount: 10, cfg: config.TenantQuotaConfig{DailyRequestQuota: 10},
			blocked: false,
		},
		{
			name:     "over request quota blocked",
			reqCount: 11, cfg: config.TenantQuotaConfig{DailyRequestQuota: 10},
			blocked: true, code: "QUOTA001",
		},
		{
			name:      "over byte quota blocked",
			reqCount:  1, byteCount: 2048,
			cfg:     config.TenantQuotaConfig{DailyBytesQuota: 1024},
			blocked: true, code: "QUOTA002",
		},
		{
			name:     "zero quota means unlimited",
			reqCount: 1_000_000, byteCount: 1_000_000,
			cfg:     config.TenantQuotaConfig{},
			blocked: false,
		},
		{
			name:     "request quota takes precedence over bytes",
			reqCount: 11, byteCount: 2048,
			cfg:     config.TenantQuotaConfig{DailyRequestQuota: 10, DailyBytesQuota: 1024},
			blocked: true, code: "QUOTA001",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			blocked, code, msg := evaluateQuota(tc.reqCount, tc.byteCount, tc.cfg)
			if blocked != tc.blocked {
				t.Fatalf("blocked=%v want %v (msg=%q)", blocked, tc.blocked, msg)
			}
			if tc.blocked && code != tc.code {
				t.Fatalf("code=%q want %q", code, tc.code)
			}
			if tc.blocked && msg == "" {
				t.Fatal("expected non-empty message when blocked")
			}
		})
	}
}

func TestQuotaPathMatches(t *testing.T) {
	prefixes := defaultQuotaPaths()
	matches := []string{"/api/v1/events", "/api/v1/logs/ingest", "/api/v1/logs/bulk/extra"}
	for _, p := range matches {
		if !quotaPathMatches(p, prefixes) {
			t.Errorf("expected %q to match ingest quota paths", p)
		}
	}
	nonMatches := []string{"/api/v1/incidents", "/health", "/api/v1/dashboard/overview"}
	for _, p := range nonMatches {
		if quotaPathMatches(p, prefixes) {
			t.Errorf("did not expect %q to match ingest quota paths", p)
		}
	}
}
