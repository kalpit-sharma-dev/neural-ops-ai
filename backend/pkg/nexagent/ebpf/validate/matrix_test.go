package validate

import "testing"

func TestEvaluateFleetMatrix(t *testing.T) {
	cases := []struct {
		name     string
		os       string
		kernel   string
		arch     string
		btf      bool
		wantPass bool
	}{
		{"ubuntu 22.04 lts", "linux", "5.15.0-91-generic", "amd64", true, true},
		{"ubuntu 24.04", "linux", "6.8.0-45-generic", "amd64", true, true},
		{"arm64 eks", "linux", "5.10.200-eks", "arm64", true, true},
		{"kernel too old", "linux", "4.19.0-16-amd64", "amd64", true, false},
		{"missing btf", "linux", "6.8.0-45-generic", "amd64", false, false},
		{"windows host", "windows", "10.0.19045", "amd64", false, false},
		{"unsupported arch", "linux", "6.8.0", "386", true, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			report := Evaluate(tc.os, tc.kernel, tc.arch, tc.btf)
			if report.Passed != tc.wantPass {
				t.Fatalf("passed=%v want %v checks=%+v", report.Passed, tc.wantPass, report.Checks)
			}
		})
	}
}

func TestParseKernelVersion(t *testing.T) {
	major, minor, err := parseKernelVersion("6.8.0-45-generic")
	if err != nil || major != 6 || minor != 8 {
		t.Fatalf("got %d.%d err=%v", major, minor, err)
	}
}
