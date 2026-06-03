// Package validate provides kernel/distro matrix checks for NEXAGENT eBPF fleet
// rollout. It is pure Go (no cilium dependency) so CI can gate compatibility
// without loading BPF programs.
package validate

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
)

// MinKernelMajor/Minor is the minimum Linux kernel for CO-RE eBPF (5.4+).
const (
	MinKernelMajor = 5
	MinKernelMinor = 4
)

// SupportedArch lists architectures the tcp_rtt program is validated against.
var SupportedArch = map[string]struct{}{
	"amd64": {},
	"arm64": {},
}

// CheckResult is one compatibility finding.
type CheckResult struct {
	Name    string `json:"name"`
	Passed  bool   `json:"passed"`
	Detail  string `json:"detail"`
	Blocked bool   `json:"blocked"` // when true, fleet rollout must not proceed
}

// FleetReport aggregates matrix checks for CI and ops tooling.
type FleetReport struct {
	Passed   bool          `json:"passed"`
	Checks   []CheckResult `json:"checks"`
	Platform string        `json:"platform"`
}

// Evaluate runs the kernel matrix against supplied runtime facts (typically
// from /proc or uname on the target host).
func Evaluate(osName, kernelRelease, goarch string, btfAvailable bool) FleetReport {
	checks := []CheckResult{
		checkOS(osName),
		checkArch(goarch),
		checkKernelVersion(kernelRelease),
		checkBTF(btfAvailable),
	}
	passed := true
	for _, c := range checks {
		if c.Blocked && !c.Passed {
			passed = false
		}
	}
	return FleetReport{
		Passed:   passed,
		Checks:   checks,
		Platform: fmt.Sprintf("%s/%s %s", osName, goarch, kernelRelease),
	}
}

func checkOS(osName string) CheckResult {
	ok := strings.EqualFold(osName, "linux")
	return CheckResult{
		Name: "os_linux", Passed: ok, Blocked: true,
		Detail: fmt.Sprintf("os=%q (eBPF collector requires Linux)", osName),
	}
}

func checkArch(goarch string) CheckResult {
	_, ok := SupportedArch[goarch]
	detail := fmt.Sprintf("arch=%q", goarch)
	if !ok {
		detail += " (supported: amd64, arm64)"
	}
	return CheckResult{Name: "arch_supported", Passed: ok, Blocked: true, Detail: detail}
}

func checkKernelVersion(release string) CheckResult {
	major, minor, err := parseKernelVersion(release)
	if err != nil {
		return CheckResult{
			Name: "kernel_version", Passed: false, Blocked: true,
			Detail: fmt.Sprintf("cannot parse kernel release %q: %v", release, err),
		}
	}
	ok := major > MinKernelMajor || (major == MinKernelMajor && minor >= MinKernelMinor)
	return CheckResult{
		Name: "kernel_version", Passed: ok, Blocked: true,
		Detail: fmt.Sprintf("kernel=%s (minimum %d.%d for CO-RE eBPF)", release, MinKernelMajor, MinKernelMinor),
	}
}

func checkBTF(btfAvailable bool) CheckResult {
	detail := "BTF vmlinux present (/sys/kernel/btf/vmlinux)"
	if !btfAvailable {
		detail = "BTF vmlinux missing — run bpftool btf dump or enable CONFIG_DEBUG_INFO_BTF"
	}
	return CheckResult{Name: "btf_available", Passed: btfAvailable, Blocked: true, Detail: detail}
}

// parseKernelVersion extracts major.minor from strings like "6.8.0-45-generic".
func parseKernelVersion(release string) (major, minor int, err error) {
	base := strings.SplitN(release, "-", 2)[0]
	parts := strings.Split(base, ".")
	if len(parts) < 2 {
		return 0, 0, fmt.Errorf("expected major.minor")
	}
	major, err = strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, err
	}
	minor, err = strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, err
	}
	return major, minor, nil
}

// CurrentGOARCH returns runtime.GOARCH for convenience in host probes.
func CurrentGOARCH() string { return runtime.GOARCH }
