//go:build !linux

package validate

import "runtime"

// ProbeHost returns a failed report on non-Linux hosts.
func ProbeHost() FleetReport {
	return Evaluate(runtime.GOOS, "n/a", runtime.GOARCH, false)
}
