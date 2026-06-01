//go:build integration

package integration_test

import (
	"os/exec"
	"testing"
)

func skipUnlessDocker(t *testing.T) {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	if err := exec.Command("docker", "info").Run(); err != nil {
		t.Skipf("docker not available for integration tests: %v", err)
	}
}
