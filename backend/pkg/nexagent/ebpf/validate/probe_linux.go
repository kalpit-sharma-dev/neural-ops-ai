//go:build linux

package validate

import (
	"os"
	"runtime"
)

// ProbeHost gathers runtime facts from the local host for fleet validation.
func ProbeHost() FleetReport {
	release := readKernelRelease()
	_, btfErr := os.Stat("/sys/kernel/btf/vmlinux")
	return Evaluate("linux", release, runtime.GOARCH, btfErr == nil)
}

func readKernelRelease() string {
	raw, err := os.ReadFile("/proc/sys/kernel/osrelease")
	if err != nil {
		return "unknown"
	}
	return string(bytesTrimSpace(raw))
}

func bytesTrimSpace(b []byte) string {
	for len(b) > 0 && (b[len(b)-1] == '\n' || b[len(b)-1] == ' ') {
		b = b[:len(b)-1]
	}
	return string(b)
}
