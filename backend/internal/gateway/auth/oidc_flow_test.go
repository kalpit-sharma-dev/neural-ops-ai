package auth_test

import (
	"testing"

	"github.com/neuralops/platform/internal/gateway/auth"
)

func TestFrontendRedirectUsesExchangeCode(t *testing.T) {
	redirect := auth.FrontendRedirect("http://localhost:5173", "abc123")
	if redirect != "http://localhost:5173/auth/callback?code=abc123" {
		t.Fatalf("unexpected redirect: %s", redirect)
	}
}

func TestValidateSubscriptionStatus(t *testing.T) {
	t.Parallel()

	cases := []struct {
		status string
		valid  bool
	}{
		{"active", true},
		{"ACTIVE", true},
		{"", true},
		{"suspended", false},
		{"cancelled", false},
	}

	for _, tc := range cases {
		err := auth.ValidateSubscriptionStatus(tc.status)
		if tc.valid && err != nil {
			t.Fatalf("status %q should be valid: %v", tc.status, err)
		}
		if !tc.valid && err == nil {
			t.Fatalf("status %q should be invalid", tc.status)
		}
	}
}
